package core

import (
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/mendes11/swarm-browser/internal/services/connector"
	"github.com/moby/moby/api/types/filters"
	"github.com/moby/moby/api/types/swarm"
	"github.com/pkg/errors"
)

type ContainerConnection interface {
	ResizeTTY(ctx context.Context, width, height uint) error
	ContainerID() string
	Conn() net.Conn
	Close() error
}

// ClusterBrowser exposes methods to browse a specific Swarm Cluster
type ClusterBrowser interface {
	ListStacks(ctx context.Context) ([]models.Stack, error)
	ListServices(ctx context.Context, stack models.Stack) ([]models.Service, error)
	ListTasks(ctx context.Context, service models.Service) ([]models.Task, error)
	AttachToService(ctx context.Context, service models.Service, cmd []string) (ContainerConnection, error)
	AttachToTask(ctx context.Context, task models.Task, cmd []string) (ContainerConnection, error)

	// Closes all open connections to the cluster nodes / containers
	Close() error
}

type SwarmConnector struct {
	Cluster   models.Cluster
	connector *connector.DockerConnector
}

// Ensure it conforms to the interface
var _ ClusterBrowser = &SwarmConnector{}

func New(cluster models.Cluster) *SwarmConnector {
	return &SwarmConnector{
		Cluster:   cluster,
		connector: connector.NewConnector(),
	}
}

// AttachToService implements ClusterBrowser.
func (s *SwarmConnector) AttachToService(ctx context.Context, service models.Service, cmd []string) (ContainerConnection, error) {
	tasks, err := s.ListTasks(ctx, service)
	if err != nil {
		return nil, errors.Wrap(err, "browswer.SwarmConnector#AttachToService: ListTasks")
	}
	// Iterate through tasks until it is able to connect to one
	for _, task := range tasks {
		if task.Status != swarm.TaskStateRunning {
			continue
		}
		containerConn, err := s.connector.AttachToContainer(ctx, task.Node.Host, task.ContainerID, cmd)
		if err != nil {
			return nil, errors.Wrap(err, "connector.SwarmConnector#AttachToService: AttachToContainer")
		}
		return containerConn, nil
	}
	return nil, fmt.Errorf("connector.SwarmConnector#AttachToService: unable to attach to a running container")
}

// AttachToTask implements Clusterconnector.
func (s *SwarmConnector) AttachToTask(ctx context.Context, task models.Task, cmd []string) (ContainerConnection, error) {
	if task.Status != "running" {
		return nil, fmt.Errorf("connector.SwarmConnector#AttachToTask: task %s is not running (status: %s)", task.TaskID, task.Status)
	}

	containerConn, err := s.connector.AttachToContainer(ctx, task.Node.Host, task.ContainerID, cmd)
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#AttachToTask: AttachToContainer")
	}

	return containerConn, nil
}

// Close implements Clusterconnector.
func (s *SwarmConnector) Close() error {
	err := s.connector.Close()
	if err != nil {
		return errors.Wrap(err, "connector.SwarmConnector#Close: Close")
	}
	return nil
}

// ListServices implements Clusterconnector.
func (s *SwarmConnector) ListServices(ctx context.Context, stack models.Stack) ([]models.Service, error) {
	cli, err := s.connector.ClientForHost(s.Cluster.Host)
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#ListServices: ClientForHost")
	}
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("com.docker.stack.namespace=%s", stack.Name))
	servicesResp, err := cli.ServiceList(ctx, swarm.ServiceListOptions{
		Status:  true,
		Filters: filter,
	})
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#ListServices: ServiceList")
	}
	services := make([]models.Service, len(servicesResp))
	for i, service := range servicesResp {
		services[i] = models.Service{
			ID:           service.ID,
			Name:         service.Spec.Name,
			RunningTasks: service.ServiceStatus.RunningTasks,
			DesiredTasks: *service.Spec.Mode.Replicated.Replicas,
			Stack:        stack,
		}
	}
	return services, nil
}

// ListStacks implements Clusterconnector.
func (s *SwarmConnector) ListStacks(ctx context.Context) ([]models.Stack, error) {
	cli, err := s.connector.ClientForHost(s.Cluster.Host)
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#ListStacks: ClientForHost")
	}
	services, err := cli.ServiceList(ctx, swarm.ServiceListOptions{Status: true})
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#ListStacks: ServiceList")
	}
	stacks := make([]models.Stack, 0)
	for _, service := range services {
		if stackName, exists := service.Spec.Labels["com.docker.stack.namespace"]; exists {
			stacks = append(stacks, models.Stack{Name: stackName})
		}
	}
	return slices.Compact(stacks), nil
}

// ListTasks implements Clusterconnector.
func (s *SwarmConnector) ListTasks(ctx context.Context, service models.Service) ([]models.Task, error) {
	cli, err := s.connector.ClientForHost(s.Cluster.Host)
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#ListTasks: ClientForHost")
	}
	filter := filters.NewArgs()
	filter.Add("service", service.ID)
	filter.Add("desired-state", "running")
	tasksResp, err := cli.TaskList(ctx, swarm.TaskListOptions{Filters: filter})
	if err != nil {
		return nil, errors.Wrap(err, "connector.SwarmConnector#ListTasks: TaskList")
	}
	tasks := make([]models.Task, len(tasksResp))
	nodeIDMap := make(map[string]models.Node)
	for i, task := range tasksResp {
		if _, exists := nodeIDMap[task.NodeID]; !exists {
			node, _, err := cli.NodeInspectWithRaw(ctx, task.NodeID)
			if err != nil {
				return nil, errors.Wrap(err, "connector.SwarmConnector#ListTasks: NodeInspectWithRaw")
			}
			nodeInfo, found := s.Cluster.GetNodeByHostname(node.Description.Hostname)
			if !found {
				return nil, fmt.Errorf("connector.SwarmConnector#ListTasks: node hostname %s is missing in the cluster configurations", node.Description.Hostname)
			}
			nodeIDMap[task.NodeID] = nodeInfo
		}
		tasks[i] = models.Task{
			TaskID:      task.ID,
			ContainerID: task.Status.ContainerStatus.ContainerID,
			Node:        nodeIDMap[task.NodeID],
			Status:      task.Status.State,
		}
	}
	return tasks, nil
}

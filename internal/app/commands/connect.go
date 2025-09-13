package commands

import (
	"context"
	"fmt"
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/types"
	"github.com/mendes11/swarm-browser/internal/config"
	"github.com/mendes11/swarm-browser/internal/services/connector"
	"github.com/moby/moby/api/types/filters"
	"github.com/moby/moby/api/types/swarm"
	"github.com/pkg/errors"
)

type ConnectedToService struct {
	Cluster     config.Cluster
	Node        config.Node
	Service     types.Service
	ContainerID string
	Conn        net.Conn
}

func ConnectToServiceCommand(ctx context.Context, conn *connector.Connector, cluster config.Cluster, service types.Service) tea.Cmd {
	return func() tea.Msg {
		cli, err := conn.ClientForHost(cluster.Host)
		if err != nil {
			return ConnectionErrorMsg{Cluster: cluster, Err: err}
		}
		filter := filters.NewArgs()
		filter.Add("service", service.ID)
		filter.Add("desired-state", "running")
		tasks, err := cli.TaskList(ctx, swarm.TaskListOptions{Filters: filter})
		if err != nil || len(tasks) == 0 {
			return ConnectionErrorMsg{Cluster: cluster, Err: err}
		}
		for _, task := range tasks {
			if task.Status.State != swarm.TaskStateRunning {
				continue
			}
			node, _, err := cli.NodeInspectWithRaw(ctx, task.NodeID)
			if err != nil {
				return ConnectionErrorMsg{Cluster: cluster, Err: err}
			}
			nodeInfo, found := cluster.GetNodeByHostname(node.Description.Hostname)
			if !found {
				return ConnectionErrorMsg{Cluster: cluster, Err: fmt.Errorf("node %s not found in cluster config", node.Description.Hostname)}
			}
			containerConn, err := conn.AttachToContainer(ctx, nodeInfo.Host, task.Status.ContainerStatus.ContainerID, []string{"bash"})
			if err != nil {
				return ConnectionErrorMsg{Cluster: cluster, Err: errors.Wrap(err, "commands.ConnectToService: failed to attach to container")}
			}
			log.Printf("Attached to container %s on host %s\n", task.Status.ContainerStatus.ContainerID, nodeInfo.Host)
			return ConnectedToService{
				Cluster:     cluster,
				Service:     service,
				Node:        *nodeInfo,
				ContainerID: task.Status.ContainerStatus.ContainerID,
				Conn:        containerConn,
			}
		}
		return nil
	}
}

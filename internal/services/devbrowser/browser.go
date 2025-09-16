package devbrowser

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/moby/moby/api/types/swarm"
)

type DevBrowser struct {
	clusterName string
	config      *DevConfig
	conns       []net.Conn
}

// Ensure it conforms to the interface
var _ core.ClusterBrowser = &DevBrowser{}

// New creates a new DevBrowser with the given cluster name and config path
func New(clusterName string, configPath string) (*DevBrowser, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate cluster exists
	if _, exists := config.Clusters[clusterName]; !exists {
		return nil, fmt.Errorf("cluster '%s' not found in config", clusterName)
	}

	return &DevBrowser{
		clusterName: clusterName,
		config:      config,
		conns:       make([]net.Conn, 0),
	}, nil
}

// NewWithConfig creates a new DevBrowser with a provided config
func NewWithConfig(clusterName string, config *DevConfig) (*DevBrowser, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Validate cluster exists
	if _, exists := config.Clusters[clusterName]; !exists {
		return nil, fmt.Errorf("cluster '%s' not found in config", clusterName)
	}

	return &DevBrowser{
		clusterName: clusterName,
		config:      config,
		conns:       make([]net.Conn, 0),
	}, nil
}

// GetCluster returns the cluster configuration for this browser
func (d *DevBrowser) GetCluster() models.Cluster {
	return d.config.Clusters[d.clusterName]
}

// ListStacks implements core.ClusterBrowser using config data
func (d *DevBrowser) ListStacks(ctx context.Context) ([]models.Stack, error) {
	stackConfigs := d.config.GetStacksForCluster(d.clusterName)
	stacks := make([]models.Stack, len(stackConfigs))
	for i, stackConfig := range stackConfigs {
		stacks[i] = models.Stack{
			Name: stackConfig.Name,
		}
	}
	return stacks, nil
}

// ListServices implements core.ClusterBrowser using config data
func (d *DevBrowser) ListServices(ctx context.Context, stack models.Stack) ([]models.Service, error) {
	stackConfigs := d.config.GetStacksForCluster(d.clusterName)
	for _, stackConfig := range stackConfigs {
		if stackConfig.Name == stack.Name {
			services := make([]models.Service, len(stackConfig.Services))
			for i, svcConfig := range stackConfig.Services {
				services[i] = models.Service{
					ID:           svcConfig.ID,
					Name:         fmt.Sprintf("%s_%s", stack.Name, svcConfig.Name),
					RunningTasks: svcConfig.RunningTasks,
					DesiredTasks: svcConfig.DesiredTasks,
					Stack:        stack,
				}
			}
			return services, nil
		}
	}

	// Return empty list if stack not found
	return []models.Service{}, nil
}

// ListTasks implements core.ClusterBrowser using config data
func (d *DevBrowser) ListTasks(ctx context.Context, service models.Service) ([]models.Task, error) {
	// Find the service in config
	var serviceConfig *ServiceConfig
	var stackClusterName string
	stackConfigs := d.config.GetStacksForCluster(d.clusterName)
	for _, stackConfig := range stackConfigs {
		if stackConfig.Name == service.Stack.Name {
			stackClusterName = stackConfig.ClusterName
			for _, svc := range stackConfig.Services {
				if svc.ID == service.ID {
					serviceConfig = &svc
					break
				}
			}
		}
	}

	if serviceConfig == nil {
		return []models.Task{}, nil
	}

	// Get nodes from the cluster
	cluster := d.config.Clusters[stackClusterName]
	nodes := make([]models.Node, 0, len(cluster.Nodes))
	nodesByName := make(map[string]models.Node)
	for nodeName, node := range cluster.Nodes {
		nodes = append(nodes, node)
		nodesByName[nodeName] = node
	}

	// If specific tasks are configured, use them
	if len(serviceConfig.Tasks) > 0 {
		tasks := make([]models.Task, len(serviceConfig.Tasks))
		for i, taskConfig := range serviceConfig.Tasks {
			var node models.Node
			if taskConfig.NodeName != "" {
				node = nodesByName[taskConfig.NodeName]
			} else {
				// Fallback to round-robin if no node specified
				node = nodes[i%len(nodes)]
			}

			taskID := taskConfig.ID
			if taskID == "" {
				taskID = fmt.Sprintf("%s-task-%03d", service.ID, i+1)
			}

			containerID := taskConfig.ContainerID
			if containerID == "" {
				containerID = fmt.Sprintf("container-%s-%03d", service.ID, i+1)
			}

			status := d.parseTaskStatus(taskConfig.Status)

			tasks[i] = models.Task{
				TaskID:      taskID,
				ContainerID: containerID,
				Node:        node,
				Status:      status,
			}
		}
		return tasks, nil
	}

	// Otherwise, generate tasks based on counts
	tasks := []models.Task{}

	// Generate running tasks
	for i := uint64(0); i < serviceConfig.RunningTasks; i++ {
		nodeIndex := int(i) % len(nodes)
		tasks = append(tasks, models.Task{
			TaskID:      fmt.Sprintf("%s-task-%03d", service.ID, i+1),
			ContainerID: fmt.Sprintf("container-%s-%03d", service.ID, i+1),
			Node:        nodes[nodeIndex],
			Status:      swarm.TaskStateRunning,
		})
	}

	// Generate pending/failed tasks for the remaining desired count
	for i := serviceConfig.RunningTasks; i < serviceConfig.DesiredTasks; i++ {
		nodeIndex := int(i) % len(nodes)
		status := swarm.TaskStatePending
		if i%2 == 1 { // Alternate between pending and failed
			status = swarm.TaskStateFailed
		}

		tasks = append(tasks, models.Task{
			TaskID:      fmt.Sprintf("%s-task-%03d", service.ID, i+1),
			ContainerID: fmt.Sprintf("container-%s-%03d", service.ID, i+1),
			Node:        nodes[nodeIndex],
			Status:      status,
		})
	}

	return tasks, nil
}

// AttachToService implements core.ClusterBrowser with local terminal simulation
func (d *DevBrowser) AttachToService(ctx context.Context, service models.Service, cmd []string) (net.Conn, error) {
	// Get tasks to find a running one
	tasks, err := d.ListTasks(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Find a running task
	var runningTask *models.Task
	for _, task := range tasks {
		if task.Status == swarm.TaskStateRunning {
			runningTask = &task
			break
		}
	}

	if runningTask == nil {
		return nil, fmt.Errorf("no running tasks found for service %s", service.Name)
	}

	// Create a pair of connected pipes to simulate network connection
	clientConn, serverConn := net.Pipe()

	// Store connection for cleanup
	d.conns = append(d.conns, clientConn, serverConn)

	// Determine which shell to use
	shellCmd := cmd
	if len(shellCmd) == 0 {
		// Default to bash, fallback to sh if not available
		if _, err := exec.LookPath("bash"); err == nil {
			shellCmd = []string{"bash"}
		} else {
			shellCmd = []string{"sh"}
		}
	}

	// Start a local process to simulate container
	process := exec.Command(shellCmd[0], shellCmd[1:]...)

	// Set up the process environment to simulate being in a container
	process.Env = append(os.Environ(),
		fmt.Sprintf("MOCK_SERVICE_NAME=%s", service.Name),
		fmt.Sprintf("MOCK_TASK_ID=%s", runningTask.TaskID),
		fmt.Sprintf("MOCK_CONTAINER_ID=%s", runningTask.ContainerID),
		fmt.Sprintf("MOCK_NODE_HOST=%s", runningTask.Node.Host),
		fmt.Sprintf("MOCK_STACK_NAME=%s", service.Stack.Name),
		fmt.Sprintf("MOCK_CLUSTER=%s", d.clusterName),
		"MOCK_ENVIRONMENT=development",
		"PS1=[DEV-CONTAINER]$ ",
	)

	// Connect process stdin/stdout/stderr to the server connection
	stdin, err := process.StdinPipe()
	if err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := process.StdoutPipe()
	if err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := process.StderrPipe()
	if err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := process.Start(); err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to start shell process: %w", err)
	}

	// Create goroutines to copy data between the connection and process
	go func() {
		defer serverConn.Close()
		defer stdin.Close()

		// Write initial banner
		banner := fmt.Sprintf(
			"\n========================================\n"+
			"  Development Browser - Mock Container\n"+
			"  Cluster: %s\n"+
			"  Stack: %s\n"+
			"  Service: %s\n"+
			"  Task: %s\n"+
			"  Container: %s\n"+
			"  Node: %s\n"+
			"========================================\n\n",
			d.clusterName, service.Stack.Name, service.Name,
			runningTask.TaskID, runningTask.ContainerID, runningTask.Node.Hostname,
		)
		serverConn.Write([]byte(banner))

		// Copy from connection to process stdin
		io.Copy(stdin, serverConn)
	}()

	go func() {
		// Copy from process stdout to connection
		io.Copy(serverConn, stdout)
	}()

	go func() {
		// Copy from process stderr to connection
		io.Copy(serverConn, stderr)
	}()

	// Monitor process and close connection when it exits
	go func() {
		process.Wait()
		time.Sleep(100 * time.Millisecond) // Give time for final output
		serverConn.Close()
	}()

	return clientConn, nil
}

// AttachToTask implements core.ClusterBrowser by attaching to a specific task
func (d *DevBrowser) AttachToTask(ctx context.Context, task models.Task, cmd []string) (net.Conn, error) {
	// Create a pair of connected pipes to simulate network connection
	clientConn, serverConn := net.Pipe()

	// Store connection for cleanup
	d.conns = append(d.conns, clientConn, serverConn)

	// Determine which shell to use
	shellCmd := cmd
	if len(shellCmd) == 0 {
		// Default to bash, fallback to sh if not available
		if _, err := exec.LookPath("bash"); err == nil {
			shellCmd = []string{"bash"}
		} else {
			shellCmd = []string{"sh"}
		}
	}

	// Start a local process to simulate container
	process := exec.Command(shellCmd[0], shellCmd[1:]...)

	// Set up the process environment to simulate being in a specific task container
	process.Env = append(os.Environ(),
		fmt.Sprintf("MOCK_TASK_ID=%s", task.TaskID),
		fmt.Sprintf("MOCK_CONTAINER_ID=%s", task.ContainerID),
		fmt.Sprintf("MOCK_NODE_HOST=%s", task.Node.Host),
		fmt.Sprintf("MOCK_CLUSTER=%s", d.clusterName),
		"MOCK_ENVIRONMENT=development",
		"PS1=[DEV-TASK]$ ",
	)

	// Connect process stdin/stdout/stderr to the server connection
	stdin, err := process.StdinPipe()
	if err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := process.StdoutPipe()
	if err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := process.StderrPipe()
	if err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := process.Start(); err != nil {
		clientConn.Close()
		serverConn.Close()
		return nil, fmt.Errorf("failed to start shell process: %w", err)
	}

	// Create goroutines to copy data between the connection and process
	go func() {
		defer serverConn.Close()
		defer stdin.Close()

		// Write initial banner (safely truncate IDs)
		taskIDDisplay := task.TaskID
		if len(taskIDDisplay) > 12 {
			taskIDDisplay = taskIDDisplay[:12]
		}
		containerIDDisplay := task.ContainerID
		if len(containerIDDisplay) > 12 {
			containerIDDisplay = containerIDDisplay[:12]
		}

		banner := fmt.Sprintf(
			"\n========================================\n"+
			"  Development Browser - Task Terminal\n"+
			"  Cluster: %s\n"+
			"  Task: %s\n"+
			"  Container: %s\n"+
			"  Node: %s\n"+
			"========================================\n\n",
			d.clusterName,
			taskIDDisplay,
			containerIDDisplay,
			task.Node.Hostname,
		)
		serverConn.Write([]byte(banner))

		// Copy from connection to stdin
		io.Copy(stdin, serverConn)
	}()

	go func() {
		// Copy stdout to connection
		io.Copy(serverConn, stdout)
	}()

	go func() {
		// Copy stderr to connection
		io.Copy(serverConn, stderr)
	}()

	// Monitor process completion
	go func() {
		process.Wait()
		time.Sleep(100 * time.Millisecond) // Give time for final output
		serverConn.Close()
	}()

	return clientConn, nil
}

// Close implements core.ClusterBrowser
func (d *DevBrowser) Close() error {
	// Close all stored connections
	for _, conn := range d.conns {
		if conn != nil {
			conn.Close()
		}
	}
	d.conns = nil
	return nil
}

// Helper methods

func (d *DevBrowser) parseTaskStatus(status string) swarm.TaskState {
	switch status {
	case "running":
		return swarm.TaskStateRunning
	case "pending":
		return swarm.TaskStatePending
	case "failed":
		return swarm.TaskStateFailed
	case "complete":
		return swarm.TaskStateComplete
	case "shutdown":
		return swarm.TaskStateShutdown
	case "rejected":
		return swarm.TaskStateRejected
	case "orphaned":
		return swarm.TaskStateOrphaned
	case "remove":
		return swarm.TaskStateRemove
	default:
		return swarm.TaskStateRunning
	}
}
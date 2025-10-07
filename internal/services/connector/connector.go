package connector

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// Connector manages the connections to remote Docker hosts
//
// When connecting to a host, it first establishes an SSH tunnel with the host,
// forwarding the remote Docker socket to a local temporary socket file.
// So it's important that the user has SSH access to the remote hosts already set.
type DockerConnector struct {
	// Maps hostnames to Docker clients
	clients        map[string]*client.Client
	sshConnections map[string]*exec.Cmd
}

type Options func(*DockerConnector)

func NewConnector(opts ...Options) *DockerConnector {
	conn := &DockerConnector{
		clients:        make(map[string]*client.Client),
		sshConnections: make(map[string]*exec.Cmd),
	}
	for _, opt := range opts {
		opt(conn)
	}
	return conn
}

func (c *DockerConnector) Close() error {
	log.Println("DockerConnector: Closing")
	for host, cli := range c.clients {
		log.Printf("DockerConnector: Closing %s\n", host)
		if err := cli.Close(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to close Docker client for host %s", host))
		}
	}

	for host, sshConn := range c.sshConnections {
		log.Printf("DockerConnector: Closing SSH process for %s\n", host)
		sshConn.Process.Kill()
		if err := sshConn.Process.Kill(); err != nil {
			log.Printf("failed to kill SSH connection for host %s: %v", host, err)
		}
		sshConn.Wait()
		log.Printf("DockerConnector: SSH connection to host %s closed\n", host)
	}
	return nil
}

// ClientForHost returns a Docker client for the given host.
// This connector expects your machine to have an SSH configuration for the host already set
// through the ~/.ssh/config file.
//
// Make sure you call Close() at the end of your program to ensure all connections are properly closed.
func (c *DockerConnector) ClientForHost(host string) (*client.Client, error) {
	if _, exists := c.clients[host]; !exists {
		return c.connectToHost(host)
	}
	return c.clients[host], nil
}

// AttachToContainer executes a command in a running container, returning an open connection to it.
// IMPORTANT: You must make sure to close the connection to avoid any issues.
func (c *DockerConnector) AttachToContainer(ctx context.Context, host string, containerID string, cmd []string) (*ContainerConnection, error) {
	cli, err := c.ClientForHost(host)
	if err != nil {
		return nil, errors.Wrap(err, "Connector#AttachToContainer: failed to retrieve host")
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          cmd,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Connector#AttachToContainer: failed to create exec")
	}
	containerCli, err := cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Connector#AttachToContainer: failed to attach to container")
	}
	return &ContainerConnection{cli: cli, attachID: execResp.ID, containerID: containerID, conn: containerCli.Conn}, nil
}

func (c *DockerConnector) connectToHost(host string) (cli *client.Client, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in connectToHost for host %s: %v", host, r)
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	log.Printf("Establishing SSH connection to host %s\n", host)
	socketPath := fmt.Sprintf("/tmp/swarm-browser-%d.sock", time.Now().UnixNano())
	sshCommand := exec.Command("ssh", "-N", "-L", fmt.Sprintf("%s:/var/run/docker.sock", socketPath), host)
	if err := sshCommand.Start(); err != nil {
		log.Println("Failed to start SSH command:", err)
		return nil, errors.Wrap(err, fmt.Sprintf("failed to start SSH tunnel to host %s", host))
	}
	log.Printf("SSH connection to host %s established with PID %d\n", host, sshCommand.Process.Pid)
	c.sshConnections[host] = sshCommand

	// Wait for the socket to be available
	if err := waitForSocket(socketPath, 5*time.Second); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to connect to Docker socket for host %s", host))
	}

	// Create a Docker client that connects to the forwarded socket
	cli, err = client.NewClientWithOpts(
		client.WithHost(fmt.Sprintf("unix://%s", socketPath)),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create Docker client for host %s", host))
	}
	c.clients[host] = cli
	return cli, nil
}

func waitForSocket(socketPath string, timeout time.Duration) error {
	ticker := time.NewTimer(timeout)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Timeout waiting for the socket to be available
			return errors.New("timeout waiting for Docker socket to be available")
		default:
			// Check if the socket file exists
			log.Println("Checking for socket:", socketPath)
			if _, err := exec.Command("test", "-S", socketPath).Output(); err == nil {
				log.Printf("Docker socket %s is now available\n", socketPath)
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

package core

import (
	"context"
	"net"

	"github.com/mendes11/swarm-browser/internal/core/models"
)

// ClusterBrowser exposes methods to browse a specific Swarm Cluster
type ClusterBrowser interface {
	ListStacks(ctx context.Context) ([]models.Stack, error)
	ListServices(ctx context.Context, stack models.Stack) ([]models.Service, error)
	ListTasks(ctx context.Context, service models.Service) ([]models.Task, error)
	AttachToService(ctx context.Context, service models.Service, cmd []string) (net.Conn, error)
	AttachToTask(ctx context.Context, task models.Task, cmd []string) (net.Conn, error)

	// Closes all open connections to the cluster nodes / containers
	Close() error
}

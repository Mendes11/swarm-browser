package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/types"
	"github.com/mendes11/swarm-browser/internal/config"
	"github.com/mendes11/swarm-browser/internal/services/connector"
	"github.com/moby/moby/api/types/swarm"
)

type ConnectedToClusterMsg struct {
	Cluster config.Cluster
	Stacks  []types.Stack
}

type ConnectionErrorMsg struct {
	Cluster config.Cluster
	Node    *config.Node
	Err     error
}

func ConnectToClusterCommand(ctx context.Context, conn *connector.Connector, cluster config.Cluster) tea.Cmd {
	return func() tea.Msg {
		cli, err := conn.ClientForHost(cluster.Host)
		if err != nil {
			return ConnectionErrorMsg{Cluster: cluster, Err: err}
		}
		services, err := cli.ServiceList(ctx, swarm.ServiceListOptions{
			Status: true,
		})
		if err != nil {
			return ConnectionErrorMsg{Cluster: cluster, Err: err}
		}

		stacks := make(map[string]*types.Stack)
		stacksArr := make([]types.Stack, 0)
		for _, service := range services {
			if stackName, exists := service.Spec.Labels["com.docker.stack.namespace"]; exists {
				if _, exists := stacks[stackName]; !exists {
					stacks[stackName] = &types.Stack{
						Name:     stackName,
						Services: make([]types.Service, 0),
					}
				}
				stacks[stackName].Services = append(stacks[stackName].Services, types.Service(service))
			}
		}

		for _, stack := range stacks {
			stacksArr = append(stacksArr, *stack)
		}
		return ConnectedToClusterMsg{Cluster: cluster, Stacks: stacksArr}
	}
}

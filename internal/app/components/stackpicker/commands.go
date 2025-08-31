package stackpicker

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/types"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
)

// StacksUpdated is sent when the list of stacks is updated
type StacksUpdated struct {
	Stacks      []types.Stack
	LastUpdated time.Time
}

// StackSelected is sent when a user selects a stack
type StackSelected struct {
	Stack types.Stack
}

func ListStacksCommandFactory(ctx context.Context, cli *client.Client) func() tea.Msg {
	return func() tea.Msg {
		services, err := cli.ServiceList(ctx, swarm.ServiceListOptions{
			Status: true,
		})
		if err != nil {
			panic(err)
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
		return StacksUpdated{
			Stacks:      stacksArr,
			LastUpdated: time.Now(),
		}
	}
}

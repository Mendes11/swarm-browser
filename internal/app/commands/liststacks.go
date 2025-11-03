package commands

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type StacksUpdated struct {
	Stacks []models.Stack
}

type ListStacksError struct {
	Err error
}

func ListStacks(browser core.ClusterBrowser) tea.Cmd {
	return func() tea.Msg {
		log.Println("Listing Stacks...")
		stacks, err := browser.ListStacks(context.Background())
		if err != nil {
			return ListStacksError{
				Err: err,
			}
		}
		log.Printf("Stacks List Received: %v\n", stacks)
		return StacksUpdated{
			Stacks: stacks,
		}
	}
}

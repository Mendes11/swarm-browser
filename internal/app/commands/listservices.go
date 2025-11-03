package commands

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type ServicesUpdated struct {
	Services []models.Service
	Stack    models.Stack
}

type ListServicesError struct {
	Err error
}

func ListServices(browser core.ClusterBrowser, stack models.Stack) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Listing Services for stack: %s...\n", stack.Name)
		services, err := browser.ListServices(context.Background(), stack)
		if err != nil {
			return ListServicesError{
				Err: err,
			}
		}
		log.Printf("Services List Received: %v\n", services)
		return ServicesUpdated{
			Services: services,
			Stack:    stack,
		}
	}
}
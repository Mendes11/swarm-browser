package commands

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type TasksUpdated struct {
	Service models.Service
	Tasks   []models.Task
}

type ListTasksError struct {
	Err error
}

func ListTasks(browser core.ClusterBrowser, service models.Service) tea.Cmd {
	return func() tea.Msg {
		log.Printf("commands.ListTasks: Listing tasks for service %s\n", service.Name)
		tasks, err := browser.ListTasks(context.Background(), service)
		if err != nil {
			return ListTasksError{Err: err}
		}
		log.Printf("Found %d tasks\n", len(tasks))
		return TasksUpdated{Service: service, Tasks: tasks}
	}
}

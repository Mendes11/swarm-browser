package commands

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

// ContainerAttachedMsg is sent when starting a container attachment
type ContainerAttachedMsg struct {
	Service *models.Service
	Task    *models.Task
	Conn    core.ContainerConnection
}

// ContainerDetachedMsg is sent when the container session ends
type ContainerDetachedMsg struct {
	Err error
}

// AttachToService creates a command to attach to a service's container
func AttachToService(browser core.ClusterBrowser, service models.Service) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Attaching to service %s\n", service.Name)
		conn, err := browser.AttachToService(context.Background(), service, []string{"/bin/bash"})
		if err != nil {
			// Try with bash if sh fails
			conn, err = browser.AttachToService(context.Background(), service, []string{"/bin/sh"})
			if err != nil {
				log.Println(fmt.Errorf("failed to attach to service: %w", err))
				return ContainerDetachedMsg{Err: fmt.Errorf("failed to attach to service: %w", err)}
			}
		}
		log.Printf("Attached to service %s\n", service.Name)
		return ContainerAttachedMsg{
			Service: &service,
			Conn:    conn,
		}
	}
}

func AttachToTask(browser core.ClusterBrowser, task models.Task) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Attaching to task %s\n", task.TaskID)
		conn, err := browser.AttachToTask(context.Background(), task, []string{"/bin/bash"})
		if err != nil {
			conn, err = browser.AttachToTask(context.Background(), task, []string{"/bin/sh"})
			if err != nil {
				log.Println(fmt.Errorf("failed to attach to service: %w", err))
				return ContainerDetachedMsg{Err: err}
			}
		}
		log.Printf("Attached to Task %s\n", task.TaskID)
		return ContainerAttachedMsg{
			Task: &task,
			Conn: conn,
		}
	}
}

// AttachToServiceWithCmd attaches to a service's container with a specific command (no fallback).
func AttachToServiceWithCmd(browser core.ClusterBrowser, service models.Service, cmd []string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Attaching to service %s with cmd %v\n", service.Name, cmd)
		conn, err := browser.AttachToService(context.Background(), service, cmd)
		if err != nil {
			log.Println(fmt.Errorf("failed to attach to service: %w", err))
			return ContainerDetachedMsg{Err: fmt.Errorf("failed to attach to service: %w", err)}
		}
		log.Printf("Attached to service %s\n", service.Name)
		return ContainerAttachedMsg{
			Service: &service,
			Conn:    conn,
		}
	}
}

// AttachToTaskWithCmd attaches to a task's container with a specific command (no fallback).
func AttachToTaskWithCmd(browser core.ClusterBrowser, task models.Task, cmd []string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Attaching to task %s with cmd %v\n", task.TaskID, cmd)
		conn, err := browser.AttachToTask(context.Background(), task, cmd)
		if err != nil {
			log.Println(fmt.Errorf("failed to attach to task: %w", err))
			return ContainerDetachedMsg{Err: err}
		}
		log.Printf("Attached to task %s\n", task.TaskID)
		return ContainerAttachedMsg{
			Task: &task,
			Conn: conn,
		}
	}
}

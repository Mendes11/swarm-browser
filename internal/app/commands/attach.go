package commands

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
	"golang.org/x/term"
)

// ContainerAttachedMsg is sent when starting a container attachment
type ContainerAttachedMsg struct {
	Service models.Service
	Conn    core.ContainerConnection
}

// ContainerOutputMsg sends container output to the UI
type ContainerOutputMsg struct {
	Data []byte
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
			Service: service,
			Conn:    conn,
		}
	}
}

// ReadContainerOutput continuously reads from the container connection
func ReadContainerOutput(conn core.ContainerConnection) tea.Cmd {
	return func() tea.Msg {
		buffer := make([]byte, 4096)
		n, err := conn.Conn().Read(buffer)
		if err != nil {
			if err == io.EOF {
				return ContainerDetachedMsg{Err: nil}
			}
			return ContainerDetachedMsg{Err: err}
		}
		return ContainerOutputMsg{Data: buffer[:n]}
	}
}

// SendToContainer sends input to the container
func SendToContainer(conn core.ContainerConnection, data []byte) tea.Cmd {
	return func() tea.Msg {
		_, err := conn.Conn().Write(data)
		if err != nil {
			return ContainerDetachedMsg{Err: err}
		}
		return nil
	}
}

// ResizeContainerTTY resizes the container's TTY
func ResizeContainerTTY(conn core.ContainerConnection) tea.Cmd {
	return func() tea.Msg {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			// Default to 80x24 if we can't get the size
			width, height = 80, 24
		}

		err = conn.ResizeTTY(context.Background(), uint(width), uint(height))
		if err != nil {
			// Non-fatal, just log it
			return nil
		}
		return nil
	}
}

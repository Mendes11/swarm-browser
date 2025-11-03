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

// ContainerDetachedMsg is sent when the container session ends
type ContainerDetachedMsg struct {
	Err error
}

// ListenToAttachmentFinishedMsg is sent when the locking command that proxies
// stdin and stdout to the connection is closed.
type ListenToAttachmentFinishedMsg struct {
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

		return ContainerAttachedMsg{
			Service: service,
			Conn:    conn,
		}
	}
}

// ListenToAttachment will proxy Stdin to the attachment's Stdin, and its Stdout to the program's stdout
// this should be called as soon as the program receives ContainerAttachedMsg.
// Once the pipe is closed with an io.EOF message, this command returns ListenToAttachmentFinished.
func ListenToAttachment(conn core.ContainerConnection, stdin, stdout *os.File) tea.Cmd {
	return func() tea.Msg {
		log.Println("Listening to container Connection...")
		go io.Copy(conn.Conn(), stdin)
		_, err := io.Copy(stdout, conn.Conn())
		if err != nil && err != io.EOF {
			log.Printf("commands.ListenToAttachment: Copy to stdout failed: %v\n", err)
			return ListenToAttachmentFinishedMsg{}
		}

		log.Println("Finished listening to container connection.")
		return ListenToAttachmentFinishedMsg{}
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

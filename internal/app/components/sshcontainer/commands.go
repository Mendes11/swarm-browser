package sshcontainer

import (
	"context"
	"io"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
)

type SshMessage struct {
	Content []byte
}

type SshDisconnected struct {
	Err error
}

type ContainerResized struct {
	Err    error
	Height uint
	Width  uint
}

type KeyCommandSent struct {
	Err error
	Key []byte
}

func ConnectCommand(conn core.ContainerConnection, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		buffer := make([]byte, 4096)
		for {
			n, err := conn.Conn().Read(buffer)
			if err == io.EOF {
				log.Println("SSH Container: Connection closed")
				return SshDisconnected{Err: nil}
			} else if err != nil {
				return SshDisconnected{Err: err}
			}
			if n > 0 {
				log.Printf("SSH Container: Received: %v\n", string(buffer[:n]))
				p.Send(SshMessage{Content: buffer[:n]})
			}
		}
	}
}

func SendResizeCommand(conn core.ContainerConnection, width, height uint) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SSH Container: Sending Resize Command: %d, %d\n", width, height)
		err := conn.ResizeTTY(context.Background(), width, height)
		log.Printf("SSH Container: Resize Command completed. Error: %v\n", err)
		return ContainerResized{
			Err: err, Width: width, Height: height,
		}
	}
}

func SendKeyCommand(conn core.ContainerConnection, k []byte) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SSH Container: Sending key command: %v\n", k)
		_, err := conn.Conn().Write(k)
		log.Printf("SSH Container: Key command sent. Error: %v\n", err)
		return KeyCommandSent{Err: err, Key: k}
	}
}

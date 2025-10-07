package sshcontainer

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/mendes11/swarm-browser/internal/core"
)

type SSHContainer struct {
	conn      core.ContainerConnection
	Program   *tea.Program
	content   string
	connected bool
}

func New(conn core.ContainerConnection, program *tea.Program) *SSHContainer {
	return &SSHContainer{
		conn:      conn,
		Program:   program,
		connected: true,
	}
}

func (c *SSHContainer) Init() tea.Cmd {
	return ConnectCommand(c.conn, c.Program)
}

func (c *SSHContainer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// return nil, SendResizeCommand(c.conn, uint(msg.Width), uint(msg.Height))
	case SshMessage:
		c.content += string(msg.Content)
		log.Println("Container content: ", c.content)
	case SshDisconnected:
		c.connected = false
		if msg.Err != nil {
			c.content += fmt.Sprintf("container disconnected unexpectedly: %v", msg.Err)
		} else {
			c.content += "\nContainer exited..."
		}
	case tea.KeyMsg:
		if !c.connected {
			return c, nil
		}

		var key []byte
		switch msg.String() {
		default:
			log.Printf("SSH Container Received msg: %v\n", msg)
			switch msg.Type {
			case tea.KeyCtrlC:
				key = []byte{0x03}
			case tea.KeyCtrlD:
				return c, tea.Quit
			case tea.KeyBackspace:
				key = []byte{0x7f}
			case tea.KeyEnter:
				key = []byte{'\r'}
			case tea.KeyTab:
				key = []byte{'\t'}
			case tea.KeyUp:
				key = []byte{0x1b, '[', 'A'}
			case tea.KeyDown:
				key = []byte{0x1b, '[', 'B'}
			case tea.KeyRight:
				key = []byte{0x1b, '[', 'C'}
			case tea.KeyLeft:
				key = []byte{0x1b, '[', 'D'}
			case tea.KeyRunes:
				key = []byte(string(msg.Runes))
			default:
				key = []byte(msg.String())
			}
		}
		if len(key) > 0 {
			return c, SendKeyCommand(c.conn, key)
		}
	}
	return c, nil
}

func (c *SSHContainer) View() string {
	return cleanTerminalOutput(c.content)
}

// cleanTerminalOutput removes problematic ANSI escape sequences while preserving visible content
func cleanTerminalOutput(content string) string {
	// Remove OSC (Operating System Command) sequences like ']0;...' (window title)
	// These are in format: ESC ] ... ESC \ or ESC ] ... BEL
	oscPattern := regexp.MustCompile(`\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)
	cleaned := oscPattern.ReplaceAllString(content, "")

	// Remove bracketed paste mode sequences
	cleaned = strings.ReplaceAll(cleaned, "\x1b[?2004h", "")
	cleaned = strings.ReplaceAll(cleaned, "\x1b[?2004l", "")

	// Strip remaining ANSI escape sequences (colors, etc.) using charmbracelet's ansi package
	cleaned = ansi.Strip(cleaned)

	return cleaned
}

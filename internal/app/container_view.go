package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mendes11/swarm-browser/internal/app/commands"
	"github.com/mendes11/swarm-browser/internal/core"
	"golang.org/x/term"
)

// ContainerView handles the interactive container session
type ContainerView struct {
	conn         core.ContainerConnection
	rawMode      bool
	oldTermState *term.State
}

// NewContainerView creates a new container view
func NewContainerView(conn core.ContainerConnection) ContainerView {
	return ContainerView{
		conn: conn,
	}
}

// Init starts the container session
func (v ContainerView) Init() tea.Cmd {
	return tea.Batch(
		commands.EnterRawMode(),
		commands.ResizeContainerTTY(v.conn),
	)
}

// RestoreTerminal restores the terminal to its original state
func (v *ContainerView) RestoreTerminal() {
	if v.rawMode && v.oldTermState != nil {
		term.Restore(int(os.Stdin.Fd()), v.oldTermState)
		v.rawMode = false
	}
}

// Update handles messages for the container view
func (v ContainerView) Update(msg tea.Msg) (ContainerView, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.EnteredRawModeMsg:
		v.rawMode = true
		v.oldTermState = msg.TerminalState
		return v, commands.ListenToAttachment(v.conn, os.Stdin, os.Stdout)

	case commands.EnterRawModeErrMsg:
		return v, func() tea.Msg {
			return ExitContainerMsg{}
		}
	case commands.ListenToAttachmentFinishedMsg:
		v.RestoreTerminal()
		return v, func() tea.Msg {
			return ExitContainerMsg{}
		}

	case tea.WindowSizeMsg:
		return v, commands.ResizeContainerTTY(v.conn)

	case commands.ContainerDetachedMsg:
		v.RestoreTerminal()
		if v.conn != nil {
			v.conn.Close()
		}
		// Return a command to go back to the previous view
		return v, func() tea.Msg {
			return ExitContainerMsg{}
		}
	}

	return v, nil
}

// View renders the container view
func (v ContainerView) View() string {
	// In raw mode, we don't render anything through bubbletea
	// The terminal output is handled directly
	if v.rawMode {
		return "Raw Mode: "
	}

	// Show a message before entering raw mode
	return lipgloss.NewStyle().
		Padding(1).
		Render(fmt.Sprintf("Attaching to container %s", v.conn.ContainerID()))
}

// ExitContainerMsg is sent when exiting the container view
type ExitContainerMsg struct{}

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

// ExitContainerMsg is sent when exiting the container view
type ExitContainerViewMsg struct {
	Err error
}

// ExitContainerView returns a command that signals to exit the container view
func ExitContainerView(err error) tea.Cmd {
	return func() tea.Msg {
		return ExitContainerViewMsg{Err: err}
	}
}

// ContainerView handles the interactive container session
type ContainerView struct {
	conn         core.ContainerConnection
	rawMode      bool
	oldTermState *term.State
	err          error
}

// NewContainerView creates a new container view
func NewContainerView(conn core.ContainerConnection) ContainerView {
	return ContainerView{
		conn:    conn,
		rawMode: false,
	}
}

// Init starts the container session
func (v ContainerView) Init() tea.Cmd {
	return tea.Batch(
		commands.EnterRawMode(),
		commands.ReadContainerOutput(v.conn),
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

		return v, tea.Batch(tea.ExitAltScreen, tea.ShowCursor)

	case commands.EnterRawModeErrMsg:
		return v, ExitContainerView(nil)

	case commands.ContainerOutputMsg:
		// Display container output directly
		os.Stdout.Write(msg.Data)
		// Continue reading
		return v, commands.ReadContainerOutput(v.conn)

	case commands.ContainerDetachedMsg:
		v.RestoreTerminal()
		v.err = msg.Err
		if v.conn != nil {
			v.conn.Close()
		}
		// Return a command to go back to the previous view
		return v, tea.Batch(tea.EnterAltScreen, tea.HideCursor, ExitContainerView(msg.Err))

	case tea.KeyMsg:
		// return v, nil
		// Convert the Bubbletea key message to actual terminal bytes
		data := KeyToBytes(msg)

		// Check if it's the detach key (Ctrl+\) - returns nil to signal detachment
		if data == nil && (msg.Type == tea.KeyCtrlBackslash || msg.String() == "ctrl+\\") {
			v.RestoreTerminal()
			if v.conn != nil {
				v.conn.Close()
			}
			return v, ExitContainerView(nil)
		}

		// Send the converted bytes to the container
		if len(data) > 0 {
			_, err := v.conn.Conn().Write(data)
			if err != nil {
				// Connection error - detach and report error
				v.RestoreTerminal()
				return v, func() tea.Msg {
					return commands.ContainerDetachedMsg{Err: err}
				}
			}
		}
		return v, nil
	}

	return v, nil
}

// View renders the container view
func (v ContainerView) View() string {
	if v.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("Error: %v\nPress any key to return...", v.err))
	}

	// In raw mode, we don't render anything through bubbletea
	// The terminal output is handled directly
	if v.rawMode {
		return "Press ctrl+d or exit the container to return to the view\n"
	}

	// Show a message before entering raw mode
	return lipgloss.NewStyle().
		Padding(1).
		Render(fmt.Sprintf("Attaching to %s...", v.conn.ContainerID()))
}

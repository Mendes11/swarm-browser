package commands

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type EnteredRawModeMsg struct {
	TerminalState *term.State
}

type EnterRawModeErrMsg struct {
	Err error
}

func EnterRawMode() tea.Cmd {
	return func() tea.Msg {
		// Save the current terminal state
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return EnterRawModeErrMsg{Err: fmt.Errorf("failed to enter raw mode: %w", err)}
		}
		return EnteredRawModeMsg{
			TerminalState: oldState,
		}
	}
}

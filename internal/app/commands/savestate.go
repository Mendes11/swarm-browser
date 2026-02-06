package commands

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/config"
)

// StateSavedMsg is sent after state is persisted.
type StateSavedMsg struct {
	Err error
}

// SaveLastCluster returns a tea.Cmd that persists the last-used cluster name.
func SaveLastCluster(clusterName string) tea.Cmd {
	return func() tea.Msg {
		state, _ := config.LoadState()
		state.LastCluster = clusterName
		if err := config.SaveState(state); err != nil {
			log.Printf("Failed to save state: %v", err)
			return StateSavedMsg{Err: err}
		}
		return StateSavedMsg{}
	}
}

// SaveCustomCommand persists a custom command for a given context key (cluster:stack:service).
func SaveCustomCommand(key string, command string) tea.Cmd {
	return func() tea.Msg {
		state, _ := config.LoadState()
		if state.CustomCommands == nil {
			state.CustomCommands = make(map[string][]string)
		}

		// Remove existing occurrence to avoid duplicates, then prepend (MRU order)
		existing := state.CustomCommands[key]
		filtered := make([]string, 0, len(existing))
		for _, cmd := range existing {
			if cmd != command {
				filtered = append(filtered, cmd)
			}
		}
		state.CustomCommands[key] = append([]string{command}, filtered...)

		// Cap at 10 custom commands per key
		if len(state.CustomCommands[key]) > 10 {
			state.CustomCommands[key] = state.CustomCommands[key][:10]
		}

		if err := config.SaveState(state); err != nil {
			log.Printf("Failed to save custom command: %v", err)
			return StateSavedMsg{Err: err}
		}
		return StateSavedMsg{}
	}
}

// SaveLastCommand persists the last-selected command for a given context key.
func SaveLastCommand(key string, commandName string) tea.Cmd {
	return func() tea.Msg {
		state, _ := config.LoadState()
		if state.LastCommands == nil {
			state.LastCommands = make(map[string]string)
		}
		state.LastCommands[key] = commandName
		if err := config.SaveState(state); err != nil {
			log.Printf("Failed to save last command: %v", err)
			return StateSavedMsg{Err: err}
		}
		return StateSavedMsg{}
	}
}

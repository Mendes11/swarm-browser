package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppState holds persistent application state across sessions.
type AppState struct {
	LastCluster    string              `json:"last_cluster"`
	CustomCommands map[string][]string `json:"custom_commands,omitempty"`
	LastCommands   map[string]string   `json:"last_commands,omitempty"`
}

// stateDir returns the path to the swarm-browser state directory.
func stateDir() (string, error) {
	return ConfigDir()
}

// LoadState reads the application state from disk.
// Returns a zero-value AppState if the file does not exist or is corrupted.
func LoadState() (AppState, error) {
	dir, err := stateDir()
	if err != nil {
		return AppState{}, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return AppState{}, nil
		}
		return AppState{}, fmt.Errorf("failed to read state file: %w", err)
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return AppState{}, nil
	}
	return state, nil
}

// SaveState writes the application state to disk.
func SaveState(state AppState) error {
	dir, err := stateDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(filepath.Join(dir, "state.json"), data, 0644)
}

package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the swarm-browser configuration directory.
// On macOS, os.UserConfigDir() returns ~/Library/Application Support, so we
// override it to ~/.config for consistency. On other platforms (Linux, Windows)
// we use the standard os.UserConfigDir().
func ConfigDir() (string, error) {
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "swarm-browser"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "swarm-browser"), nil
}

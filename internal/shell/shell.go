package shell

import "os"

// UserShell returns the user's default shell from $SHELL,
// falling back to "sh" if not set.
func UserShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "sh"
}

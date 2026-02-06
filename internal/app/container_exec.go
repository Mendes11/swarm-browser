package app

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/mendes11/swarm-browser/internal/core"
	"golang.org/x/term"
)

const detachKey = 0x1c // Ctrl+\

// ContainerExecCmd implements tea.ExecCommand to provide raw byte passthrough
// between the terminal and a Docker container connection. This bypasses
// Bubbletea's key parsing entirely, ensuring all key combinations are forwarded.
type ContainerExecCmd struct {
	conn   core.ContainerConnection
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// NewContainerExecCmd creates a new ContainerExecCmd for the given connection.
func NewContainerExecCmd(conn core.ContainerConnection) *ContainerExecCmd {
	return &ContainerExecCmd{conn: conn}
}

func (c *ContainerExecCmd) SetStdin(r io.Reader)  { c.stdin = r }
func (c *ContainerExecCmd) SetStdout(w io.Writer)  { c.stdout = w }
func (c *ContainerExecCmd) SetStderr(w io.Writer)  { c.stderr = w }

// Run executes the bidirectional I/O between terminal and container.
// It blocks until the user detaches (Ctrl+\) or the container connection closes.
func (c *ContainerExecCmd) Run() error {
	// Enter alt screen buffer for a clean terminal
	os.Stdout.WriteString("\033[?1049h")
	defer os.Stdout.WriteString("\033[?1049l")

	// Enter raw mode so individual keystrokes are captured
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Resize container TTY to match current terminal size
	c.resizeTTY()

	// Listen for terminal resize signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	defer signal.Stop(sigCh)

	// Channel to signal when I/O completes
	done := make(chan error, 1)

	// Copy container output → terminal stdout
	go func() {
		_, err := io.Copy(c.stdout, c.conn.Conn())
		done <- err
	}()

	// Handle resize signals
	go func() {
		for range sigCh {
			c.resizeTTY()
		}
	}()

	// Read stdin and forward to container, watching for detach key
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := c.stdin.Read(buf)
			if err != nil {
				done <- err
				return
			}

			// Scan for detach key (Ctrl+\)
			for i := 0; i < n; i++ {
				if buf[i] == detachKey {
					// Send any bytes before the detach key
					if i > 0 {
						c.conn.Conn().Write(buf[:i])
					}
					done <- nil
					return
				}
			}

			// Forward all bytes to the container
			if _, err := c.conn.Conn().Write(buf[:n]); err != nil {
				done <- err
				return
			}
		}
	}()

	// Wait for either direction to finish
	return <-done
}

func (c *ContainerExecCmd) resizeTTY() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}
	// Best-effort resize, ignore errors
	c.conn.ResizeTTY(context.Background(), uint(width), uint(height))
}

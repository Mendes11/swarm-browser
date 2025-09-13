package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/creack/pty"
)

// Terminal view states
const (
	stateCommandList = iota
	stateTerminal
)

// Constants for layout
const (
	sidebarWidth = 30
)

// Message types
type (
	// Terminal output update
	terminalOutputMsg []byte

	// Terminal exited
	terminalExitMsg struct{}

	// Error message
	errMsg error
)

// Command represents a command we can run
type Command struct {
	Name        string
	Cmd         string
	Args        []string
	Description string
}

// Model represents the TUI application state
type Model struct {
	state           int
	commands        []Command
	selectedIndex   int
	terminalView    viewport.Model
	terminalOutput  string      // Store terminal output
	currentInput    string
	cursorPos       int
	ptmx            *os.File // pseudo-terminal master
	cmd             *exec.Cmd
	width           int
	height          int
	mu              sync.Mutex
	program         *tea.Program // reference to the program for sending messages
	lastClickIndex  int          // Track last clicked item for double-click
	lastClickTime   int64        // Track click timing
}

// Styling
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	terminalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Bold(true)
)

// Initialize the model
func initialModel() Model {
	// Create viewport for terminal output
	vp := viewport.New(80, 20)
	vp.Style = terminalStyle

	// Available commands to run
	commands := []Command{
		{
			Name:        "Bash Shell",
			Cmd:         "/bin/bash",
			Args:        []string{},
			Description: "Interactive bash shell",
		},
		{
			Name:        "Python REPL",
			Cmd:         "python3",
			Args:        []string{},
			Description: "Python interactive interpreter",
		},
		{
			Name:        "Top",
			Cmd:         "top",
			Args:        []string{},
			Description: "System process viewer",
		},
		{
			Name:        "Node REPL",
			Cmd:         "node",
			Args:        []string{},
			Description: "Node.js interactive interpreter",
		},
	}

	return Model{
		state:        stateCommandList,
		commands:     commands,
		terminalView: vp,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// SetProgram sets the program reference
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateCommandList:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			case "down", "j":
				if m.selectedIndex < len(m.commands)-1 {
					m.selectedIndex++
				}
			case "enter":
				// Run selected command
				m.state = stateTerminal
				cmd := m.runCommand(m.commands[m.selectedIndex])
				cmds = append(cmds, cmd)
			}

		case stateTerminal:
			switch msg.String() {
			case "ctrl+d":
				// Exit terminal
				m.stopCommand()
				m.state = stateCommandList
				m.terminalOutput = ""
				m.terminalView.SetContent("")

			case "ctrl+c":
				// Send Ctrl+C to the process
				if m.ptmx != nil {
					m.ptmx.Write([]byte{0x03})
				}

			default:
				// Send all other input to the PTY
				if m.ptmx != nil {
					// Handle special keys
					switch msg.Type {
					case tea.KeyBackspace:
						m.ptmx.Write([]byte{0x7f})
					case tea.KeyEnter:
						m.ptmx.Write([]byte{'\r'})
					case tea.KeyTab:
						m.ptmx.Write([]byte{'\t'})
					case tea.KeyUp:
						m.ptmx.Write([]byte{0x1b, '[', 'A'})
					case tea.KeyDown:
						m.ptmx.Write([]byte{0x1b, '[', 'B'})
					case tea.KeyRight:
						m.ptmx.Write([]byte{0x1b, '[', 'C'})
					case tea.KeyLeft:
						m.ptmx.Write([]byte{0x1b, '[', 'D'})
					case tea.KeyRunes:
						m.ptmx.Write([]byte(string(msg.Runes)))
					}
				}
			}
		}

	case tea.MouseMsg:
		if m.state == stateCommandList && msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Calculate which command was clicked based on mouse position
			// Account for title and padding (starts at line 3 roughly)
			clickedLine := msg.Y - 3
			if clickedLine >= 0 && clickedLine < len(m.commands) {
				currentTime := time.Now().UnixMilli()
				
				// Check for double-click (same item within 500ms)
				if m.lastClickIndex == clickedLine && (currentTime - m.lastClickTime) < 500 {
					// Double-click: run the command
					m.selectedIndex = clickedLine
					m.state = stateTerminal
					cmd := m.runCommand(m.commands[m.selectedIndex])
					cmds = append(cmds, cmd)
				} else {
					// Single click: just select
					m.selectedIndex = clickedLine
				}
				
				m.lastClickIndex = clickedLine
				m.lastClickTime = currentTime
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport size
		// Account for sidebar in terminal state
		terminalWidth := msg.Width - 6
		if m.state == stateTerminal {
			terminalWidth = msg.Width - sidebarWidth - 10
		}
		terminalHeight := msg.Height - 10
		m.terminalView.Width = terminalWidth
		m.terminalView.Height = terminalHeight

		// Resize PTY if active
		if m.ptmx != nil {
			pty.Setsize(m.ptmx, &pty.Winsize{
				Rows: uint16(terminalHeight),
				Cols: uint16(terminalWidth),
			})
		}

	case terminalOutputMsg:
		// Update terminal content
		m.terminalOutput += string(msg)
		m.terminalView.SetContent(m.terminalOutput)
		m.terminalView.GotoBottom()

	case terminalExitMsg:
		// Process exited
		m.state = stateCommandList
		m.terminalOutput = ""
		m.terminalView.SetContent("")

	case errMsg:
		// Handle errors
		m.terminalOutput += fmt.Sprintf("\r\n[ERROR] %v\r\n", msg)
		m.terminalView.SetContent(m.terminalOutput)
		m.terminalView.GotoBottom()
	}

	// Update viewport
	var cmd tea.Cmd
	m.terminalView, cmd = m.terminalView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *Model) View() string {
	switch m.state {
	case stateCommandList:
		return m.renderCommandList()
	case stateTerminal:
		return m.renderTerminal()
	default:
		return ""
	}
}

// Render command list view
func (m *Model) renderCommandList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Terminal Emulator Demo") + "\n\n")

	for i, cmd := range m.commands {
		cursor := " "
		if i == m.selectedIndex {
			cursor = ">"
		}

		line := fmt.Sprintf("%s %s - %s", cursor, cmd.Name, cmd.Description)
		if i == m.selectedIndex {
			line = selectedStyle.Render(line)
		}
		s.WriteString(line + "\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/k: up • ↓/j: down • enter: run • click: select • double-click: run • q: quit"))

	return appStyle.Render(s.String())
}

// Render terminal view with sidebar
func (m *Model) renderTerminal() string {
	// Calculate dimensions
	terminalWidth := m.width - sidebarWidth - 10
	contentHeight := m.height - 8

	// Build sidebar
	sidebar := m.renderSidebar(contentHeight)

	// Build terminal content
	var terminalContent strings.Builder
	title := fmt.Sprintf("Running: %s", m.commands[m.selectedIndex].Name)
	terminalContent.WriteString(titleStyle.Render(title) + "\n\n")
	terminalContent.WriteString(m.terminalView.View() + "\n\n")
	terminalContent.WriteString(helpStyle.Render("ctrl+d: exit • ctrl+c: interrupt"))

	// Combine sidebar and terminal side by side
	terminal := lipgloss.NewStyle().Width(terminalWidth).Render(terminalContent.String())
	combined := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, "  ", terminal)

	return appStyle.Render(combined)
}

// Render sidebar with system information
func (m *Model) renderSidebar(height int) string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("System Info") + "\n\n")

	// Command info
	s.WriteString(labelStyle.Render("Command:") + "\n")
	if m.selectedIndex < len(m.commands) {
		cmd := m.commands[m.selectedIndex]
		s.WriteString(infoStyle.Render(cmd.Name) + "\n")
		s.WriteString(infoStyle.Render(cmd.Cmd) + "\n\n")
	}

	// Terminal dimensions
	s.WriteString(labelStyle.Render("Terminal Size:") + "\n")
	s.WriteString(infoStyle.Render(fmt.Sprintf("%dx%d", m.terminalView.Width, m.terminalView.Height)) + "\n\n")

	// Process status
	s.WriteString(labelStyle.Render("Status:") + "\n")
	if m.cmd != nil && m.cmd.Process != nil {
		s.WriteString(infoStyle.Render(fmt.Sprintf("PID: %d", m.cmd.Process.Pid)) + "\n")
		s.WriteString(infoStyle.Render("Running") + "\n")
	} else {
		s.WriteString(infoStyle.Render("Not running") + "\n")
	}

	s.WriteString("\n")
	s.WriteString(labelStyle.Render("Tips:") + "\n")
	s.WriteString(infoStyle.Render("• Type normally\n"))
	s.WriteString(infoStyle.Render("• Arrow keys work\n"))
	s.WriteString(infoStyle.Render("• Tab completion\n"))
	s.WriteString(infoStyle.Render("• Ctrl+D to exit\n"))

	// Wrap in sidebar style with fixed width and height
	return sidebarStyle.
		Width(sidebarWidth).
		Height(height).
		Render(s.String())
}

// Run a command with PTY
func (m *Model) runCommand(command Command) tea.Cmd {
	return func() tea.Msg {
		// Create command
		m.cmd = exec.Command(command.Cmd, command.Args...)

		// Start the command with a PTY
		var err error
		m.ptmx, err = pty.Start(m.cmd)
		if err != nil {
			return errMsg(err)
		}

		// Set initial size
		pty.Setsize(m.ptmx, &pty.Winsize{
			Rows: uint16(m.terminalView.Height),
			Cols: uint16(m.terminalView.Width),
		})

		// Start reading output in a goroutine
		go m.readPTYOutput(m.program)

		return nil
	}
}

// Read PTY output
func (m *Model) readPTYOutput(program *tea.Program) {
	buf := make([]byte, 1024)
	for {
		n, err := m.ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				program.Send(errMsg(err))
			}
			break
		}
		if n > 0 {
			// Send output to the program as a message
			output := make([]byte, n)
			copy(output, buf[:n])
			program.Send(terminalOutputMsg(output))
		}
	}

	// Command exited
	m.ptmx.Close()
	m.cmd.Wait()
	program.Send(terminalExitMsg{})
}

// Stop the running command
func (m *Model) stopCommand() {
	if m.ptmx != nil {
		m.ptmx.Close()
	}
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
		m.cmd.Wait()
	}
}

func main() {
	m := initialModel()

	// Use alt screen and mouse support
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	// Set the program reference so we can send messages from goroutines
	m.SetProgram(p)
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

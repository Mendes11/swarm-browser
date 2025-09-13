package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/moby/moby/api/types"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// DockerModel represents the TUI application state for Docker
type DockerModel struct {
	state          int
	containers     []types.Container
	selectedIndex  int
	terminalView   viewport.Model
	terminalOutput string
	dockerClient   *client.Client
	attachResp     types.HijackedResponse
	width          int
	height         int
	program        *tea.Program
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

// Message types for Docker
type (
	dockerOutputMsg []byte
	dockerExitMsg   struct{}
	dockerErrorMsg  error
	containersMsg   []types.Container
)

// Initialize Docker model
func initialDockerModel() (*DockerModel, error) {
	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	// Create viewport
	vp := viewport.New(80, 20)
	vp.Style = terminalStyle

	ctx := context.Background()

	m := &DockerModel{
		state:        stateCommandList,
		terminalView: vp,
		dockerClient: cli,
		ctx:          ctx,
	}

	// List containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		return nil, err
	}
	m.containers = containers

	return m, nil
}

func (m *DockerModel) Init() tea.Cmd {
	return nil
}

func (m *DockerModel) SetProgram(p *tea.Program) {
	m.program = p
}

func (m *DockerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if m.selectedIndex < len(m.containers)-1 {
					m.selectedIndex++
				}
			case "enter":
				// Attach to selected container
				m.state = stateTerminal
				cmd := m.attachToContainer()
				cmds = append(cmds, cmd)
			case "r":
				// Refresh container list
				cmd := m.refreshContainers()
				cmds = append(cmds, cmd)
			}

		case stateTerminal:
			switch msg.String() {
			case "ctrl+d":
				// Detach from container
				m.detachFromContainer()
				m.state = stateCommandList
				m.terminalOutput = ""
				m.terminalView.SetContent("")

			case "ctrl+c":
				// Send Ctrl+C to the container
				if m.attachResp.Conn != nil {
					m.attachResp.Conn.Write([]byte{0x03})
				}

			default:
				// Send input to container
				if m.attachResp.Conn != nil {
					switch msg.Type {
					case tea.KeyBackspace:
						m.attachResp.Conn.Write([]byte{0x7f})
					case tea.KeyEnter:
						m.attachResp.Conn.Write([]byte{'\r'})
					case tea.KeyTab:
						m.attachResp.Conn.Write([]byte{'\t'})
					case tea.KeyUp:
						m.attachResp.Conn.Write([]byte{0x1b, '[', 'A'})
					case tea.KeyDown:
						m.attachResp.Conn.Write([]byte{0x1b, '[', 'B'})
					case tea.KeyRight:
						m.attachResp.Conn.Write([]byte{0x1b, '[', 'C'})
					case tea.KeyLeft:
						m.attachResp.Conn.Write([]byte{0x1b, '[', 'D'})
					case tea.KeyRunes:
						m.attachResp.Conn.Write([]byte(string(msg.Runes)))
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport size
		terminalWidth := msg.Width - 6
		if m.state == stateTerminal {
			terminalWidth = msg.Width - sidebarWidth - 10
		}
		terminalHeight := msg.Height - 10
		m.terminalView.Width = terminalWidth
		m.terminalView.Height = terminalHeight

		// Resize container TTY if attached
		if m.state == stateTerminal && m.selectedIndex < len(m.containers) {
			m.resizeContainerTTY(terminalHeight, terminalWidth)
		}

	case dockerOutputMsg:
		// Update terminal content
		m.terminalOutput += string(msg)
		m.terminalView.SetContent(m.terminalOutput)
		m.terminalView.GotoBottom()

	case dockerExitMsg:
		// Container detached
		m.state = stateCommandList
		m.terminalOutput = ""
		m.terminalView.SetContent("")

	case dockerErrorMsg:
		// Handle errors
		m.terminalOutput += fmt.Sprintf("\r\n[ERROR] %v\r\n", msg)
		m.terminalView.SetContent(m.terminalOutput)
		m.terminalView.GotoBottom()

	case containersMsg:
		m.containers = msg
	}

	// Update viewport
	var cmd tea.Cmd
	m.terminalView, cmd = m.terminalView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *DockerModel) View() string {
	switch m.state {
	case stateCommandList:
		return m.renderContainerList()
	case stateTerminal:
		return m.renderDockerTerminal()
	default:
		return ""
	}
}

func (m *DockerModel) renderContainerList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Docker Container Attach") + "\n\n")

	if len(m.containers) == 0 {
		s.WriteString("No running containers found.\n")
		s.WriteString("Start a container with: docker run -it ubuntu bash\n")
	} else {
		for i, c := range m.containers {
			cursor := " "
			if i == m.selectedIndex {
				cursor = ">"
			}

			// Show container name and image
			name := c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
			line := fmt.Sprintf("%s %s (%s)", cursor, name, c.Image)
			if i == m.selectedIndex {
				line = selectedStyle.Render(line)
			}
			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/k: up • ↓/j: down • enter: attach • r: refresh • q: quit"))

	return appStyle.Render(s.String())
}

func (m *DockerModel) renderDockerTerminal() string {
	// Calculate dimensions
	terminalWidth := m.width - sidebarWidth - 10
	contentHeight := m.height - 8

	// Build sidebar
	sidebar := m.renderDockerSidebar(contentHeight)

	// Build terminal content
	var terminalContent strings.Builder
	containerName := "Unknown"
	if m.selectedIndex < len(m.containers) && len(m.containers[m.selectedIndex].Names) > 0 {
		containerName = m.containers[m.selectedIndex].Names[0]
		if containerName[0] == '/' {
			containerName = containerName[1:]
		}
	}
	title := fmt.Sprintf("Attached to: %s", containerName)
	terminalContent.WriteString(titleStyle.Render(title) + "\n\n")
	terminalContent.WriteString(m.terminalView.View() + "\n\n")
	terminalContent.WriteString(helpStyle.Render("ctrl+d: detach • ctrl+c: interrupt"))

	// Combine sidebar and terminal
	terminal := lipgloss.NewStyle().Width(terminalWidth).Render(terminalContent.String())
	combined := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, "  ", terminal)

	return appStyle.Render(combined)
}

func (m *DockerModel) renderDockerSidebar(height int) string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Container Info") + "\n\n")

	if m.selectedIndex < len(m.containers) {
		c := m.containers[m.selectedIndex]

		// Container details
		s.WriteString(labelStyle.Render("Name:") + "\n")
		name := c.Names[0]
		if name[0] == '/' {
			name = name[1:]
		}
		s.WriteString(infoStyle.Render(name) + "\n\n")

		s.WriteString(labelStyle.Render("Image:") + "\n")
		s.WriteString(infoStyle.Render(c.Image) + "\n\n")

		s.WriteString(labelStyle.Render("ID:") + "\n")
		s.WriteString(infoStyle.Render(c.ID[:12]) + "\n\n")

		s.WriteString(labelStyle.Render("Status:") + "\n")
		s.WriteString(infoStyle.Render(c.Status) + "\n\n")

		s.WriteString(labelStyle.Render("State:") + "\n")
		s.WriteString(infoStyle.Render(c.State) + "\n")
	}

	s.WriteString("\n")
	s.WriteString(labelStyle.Render("Tips:") + "\n")
	s.WriteString(infoStyle.Render("• Type normally\n"))
	s.WriteString(infoStyle.Render("• Ctrl+D to detach\n"))
	s.WriteString(infoStyle.Render("• Container stays running\n"))

	return sidebarStyle.
		Width(sidebarWidth).
		Height(height).
		Render(s.String())
}

func (m *DockerModel) attachToContainer() tea.Cmd {
	return func() tea.Msg {
		if m.selectedIndex >= len(m.containers) {
			return dockerErrorMsg(fmt.Errorf("invalid container selection"))
		}

		containerID := m.containers[m.selectedIndex].ID

		// Create a new context for this attach session
		ctx, cancel := context.WithCancel(m.ctx)
		m.cancelFunc = cancel

		// Attach to container with TTY
		attachOptions := types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		}

		resp, err := m.dockerClient.ContainerAttach(ctx, containerID, attachOptions)
		if err != nil {
			return dockerErrorMsg(err)
		}

		m.attachResp = resp

		// Start reading output
		go m.readDockerOutput()

		return nil
	}
}

func (m *DockerModel) readDockerOutput() {
	defer m.attachResp.Close()

	// For raw TTY mode, we can read directly
	// For multiplexed mode, we'd need to use stdcopy.StdCopy
	buf := make([]byte, 1024)
	for {
		n, err := m.attachResp.Reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				m.program.Send(dockerErrorMsg(err))
			}
			break
		}
		if n > 0 {
			output := make([]byte, n)
			copy(output, buf[:n])
			m.program.Send(dockerOutputMsg(output))
		}
	}

	m.program.Send(dockerExitMsg{})
}

func (m *DockerModel) detachFromContainer() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	if m.attachResp.Conn != nil {
		m.attachResp.Close()
	}
}

func (m *DockerModel) resizeContainerTTY(height, width int) {
	if m.selectedIndex >= len(m.containers) {
		return
	}

	containerID := m.containers[m.selectedIndex].ID

	resizeOptions := types.ResizeOptions{
		Height: uint(height),
		Width:  uint(width),
	}

	// Fire and forget resize
	go m.dockerClient.ContainerResize(m.ctx, containerID, resizeOptions)
}

func (m *DockerModel) refreshContainers() tea.Cmd {
	return func() tea.Msg {
		containers, err := m.dockerClient.ContainerList(m.ctx, types.ContainerListOptions{All: false})
		if err != nil {
			return dockerErrorMsg(err)
		}
		return containersMsg(containers)
	}
}

// Alternative main function for Docker
func dockerMain() {
	m, err := initialDockerModel()
	if err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		return
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

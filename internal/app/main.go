package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mendes11/swarm-browser/internal/app/commands"
	"github.com/mendes11/swarm-browser/internal/config"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type Model struct {
	conf        config.Config
	state       ViewState
	browser     core.ClusterBrowser
	clusterInfo ClusterInfo
	table       table.Model
	keys        AppKeyMap
	help        help.Model
	width       int
	height      int

	// Navigation state
	stacks        []models.Stack
	selectedStack models.Stack
	services      []models.Service

	// Container session
	containerView *ContainerView
	containerConn core.ContainerConnection
}

var _ tea.Model = Model{}

func New(conf config.Config) Model {
	clusterInfo := ClusterInfo{
		Cluster:  models.Cluster{},
		Status:   Disconnected,
		NodeInfo: models.NodeInfo{},
	}
	if conf.InitialCluster != "" {
		initialCluster := conf.Clusters[conf.InitialCluster]
		clusterInfo.Cluster = initialCluster
	}

	// Initialize keys first so we can pass the table keymap
	keys := DefaultAppKeyMap()

	return Model{
		conf:        conf,
		state:       Initializing,
		clusterInfo: clusterInfo,
		table:       newTable(keys.Table),
		keys:        keys,
		help:        help.New(),
	}
}

func (m Model) Close() error {
	// Clean up container connection if active
	if m.containerView != nil {
		m.containerView.RestoreTerminal()
	}
	if m.containerConn != nil {
		m.containerConn.Close()
	}
	if m.browser != nil {
		return m.browser.Close()
	}
	return nil
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	m.clusterInfo.Status = Connecting
	return commands.ConnectToCluster(m.clusterInfo.Cluster)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// If we're in container mode, pass resize to container view
		if m.state == ContainerAttached && m.containerView != nil {
			*m.containerView, cmd = m.containerView.Update(msg)
			return m, cmd
		}

		m.table.SetWidth(m.tableWidth())
		m.table.SetHeight(m.tableHeight())

	case commands.ClusterConnected:
		m.browser = msg.Browser
		m.clusterInfo = ClusterInfo{
			Cluster:  msg.Cluster,
			NodeInfo: msg.Info,
			Status:   Connected,
		}
		return m, commands.ListStacks(m.browser)

	case commands.StacksUpdated:
		m.state = StacksList
		m.stacks = msg.Stacks
		rows := make([]table.Row, len(msg.Stacks))
		for i, stack := range msg.Stacks {
			rows[i] = []string{stack.Name}
		}
		m.table.SetColumns([]table.Column{{Title: "Name", Width: m.table.Width()}})
		m.table.SetRows(rows)
		m.table.SetCursor(0) // Reset cursor to first item
		return m, nil

	case commands.ServicesUpdated:
		m.state = ServicesList
		m.services = msg.Services
		m.selectedStack = msg.Stack
		rows := make([]table.Row, len(msg.Services))
		for i, service := range msg.Services {
			rows[i] = []string{
				service.ID,
				service.Name,
				fmt.Sprintf("%d/%d", service.RunningTasks, service.DesiredTasks),
			}
		}
		// Calculate column widths based on table width
		tableWidth := m.table.Width()
		idWidth := 20
		replicasWidth := 12
		nameWidth := tableWidth - idWidth - replicasWidth - 4 // Account for borders

		m.table.SetColumns([]table.Column{
			{Title: "ID", Width: idWidth},
			{Title: "Name", Width: nameWidth},
			{Title: "Replicas", Width: replicasWidth},
		})
		m.table.SetRows(rows)
		m.table.SetCursor(0) // Reset cursor to first item
		return m, nil
	case commands.ClusterConnectionFailed:
		m.clusterInfo.Err = msg.Err
		m.clusterInfo.Status = Disconnected
		return m, nil

	case commands.ContainerAttachedMsg:
		// Successfully attached to container
		m.state = ContainerAttached
		m.containerConn = msg.Conn
		view := NewContainerView(msg.Conn)
		m.containerView = &view
		return m, m.containerView.Init()

	case ExitContainerMsg:
		// Clean exit from container
		m.state = ServicesList
		m.containerConn.Close()
		m.containerView = nil
		m.containerConn = nil
		return m, nil

	case tea.KeyMsg:
		// If we're in container mode, pass all keys to the container view
		if m.state == ContainerAttached && m.containerView != nil {
			*m.containerView, cmd = m.containerView.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.table.SetHeight(m.tableHeight())
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Refresh):
			switch m.state {
			case StacksList:
				if m.browser != nil {
					return m, commands.ListStacks(m.browser)
				}
			case ServicesList:
				if m.browser != nil && m.selectedStack.Name != "" {
					return m, commands.ListServices(m.browser, m.selectedStack)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			switch m.state {
			case StacksList:
				// Get selected stack and navigate to services
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.stacks) && m.browser != nil {
					selectedStack := m.stacks[cursor]
					return m, commands.ListServices(m.browser, selectedStack)
				}
			case ServicesList:
				// TODO: Handle service selection to show tasks
				// This will be implemented when you add task list functionality
			}
			return m, nil

		case key.Matches(msg, m.keys.Back):
			switch m.state {
			case ServicesList:
				// Go back to stacks list
				m.state = StacksList
				rows := make([]table.Row, len(m.stacks))
				for i, stack := range m.stacks {
					rows[i] = []string{stack.Name}
				}
				m.table = newTable(m.keys.Table)
				m.table.SetHeight(m.tableHeight())
				m.table.SetColumns([]table.Column{{Title: "Name", Width: m.table.Width()}})
				m.table.SetRows(rows)
				// Try to restore cursor position to the selected stack
				for i, stack := range m.stacks {
					if stack.Name == m.selectedStack.Name {
						m.table.SetCursor(i)
						break
					}
				}
				return m, nil
			case TaskList:
				// TODO: Go back to services list
				// This will be implemented when you add task list functionality
			}
			return m, nil

		case key.Matches(msg, m.keys.Connect):
			// Connect to container - only available in ServicesList
			if m.state == ServicesList {
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.services) && m.browser != nil {
					selectedService := m.services[cursor]
					return m, commands.AttachToService(m.browser, selectedService)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Filter):
			// TODO: Implement filtering
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	// If we're in container mode, show the container view
	if m.state == ContainerAttached && m.containerView != nil {
		return m.containerView.View()
	}

	header := ClusterInfoView(m.clusterInfo)

	// Create contextual keymap for help display
	contextualKeys := NewContextualKeyMap(&m.keys, m.state)
	helpView := m.help.View(contextualKeys)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		TableStyle.Render(m.table.View()),
		helpView,
	)
}

func (m Model) renderAppHeader() string {
	return AppHeaderStyle.Render(ClusterInfoView(m.clusterInfo))
}

func (m Model) tableHeight() int {
	// Calculate available height by subtracting header and help text
	headerHeight := lipgloss.Height(m.renderAppHeader())
	contextualKeys := NewContextualKeyMap(&m.keys, m.state)
	helpHeight := lipgloss.Height(m.help.View(contextualKeys))
	padding := 4 // Some padding for borders and spacing

	availableHeight := m.height - headerHeight - helpHeight - padding

	// Ensure we don't return negative height
	if availableHeight < 1 {
		return 1
	}
	return availableHeight
}

func (m Model) tableWidth() int {
	return m.width - 4
}

func newTable(keyMap table.KeyMap) table.Model {
	t := table.New(
		table.WithFocused(true),
		table.WithKeyMap(keyMap),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

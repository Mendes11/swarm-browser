package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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

	// Filter state
	filterActive bool
	filterInput  textinput.Model

	// Navigation state
	stacks          []models.Stack
	selectedStack   *models.Stack
	services        []models.Service
	selectedService *models.Service
	tasks           []models.Task

	// Cluster selection state
	clustersForDisplay []commands.ClusterTableRow
	previousState      ViewState
	currentClusterName string

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

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter..."
	filterInput.Prompt = "/ "
	filterInput.CharLimit = 100
	filterInput.Width = 50

	return Model{
		conf:               conf,
		state:              Initializing,
		clusterInfo:        clusterInfo,
		table:              newTable(keys.Table),
		keys:               keys,
		help:               help.New(),
		filterInput:        filterInput,
		currentClusterName: conf.InitialCluster,
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

		// Update filter input width to match table width
		m.filterInput.Width = m.tableWidth()

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

		var selectedStack *models.Stack
		if m.state == StacksList && m.table.Cursor() < len(m.stacks) {
			selectedStack = &m.stacks[m.table.Cursor()]
		}
		m.stacks = msg.Stacks
		m.showStacksTable(msg.Stacks, selectedStack)
		return m, nil

	case commands.ServicesUpdated:
		m.state = ServicesList
		m.services = msg.Services
		m.selectedStack = &msg.Stack
		m.showServicesTable(msg.Services, nil)
		return m, nil
	case commands.TasksUpdated:
		m.state = TaskList
		m.tasks = msg.Tasks
		m.selectedService = &msg.Service
		m.showTasksTable(msg.Tasks, nil)
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

	case ExitContainerViewMsg:
		// Clean exit from container
		m.state = ServicesList
		m.containerConn.Close()
		m.containerView = nil
		m.containerConn = nil
		return m, nil

	case commands.ClustersListed:
		m.clustersForDisplay = msg.Clusters
		m.showClustersTable(msg.Clusters, msg.CurrentCluster)
		return m, nil

	case tea.KeyMsg:
		// If we're in container mode, pass all keys to the container view
		if m.state == ContainerAttached && m.containerView != nil {
			*m.containerView, cmd = m.containerView.Update(msg)
			return m, cmd
		}

		// Handle filter mode
		if m.filterActive {
			switch {
			case key.Matches(msg, m.keys.Enter):
				// Exit filter mode and keep filter applied
				m.filterActive = false
				m.filterInput.Blur()
				m.table.Focus()
				m.table.SetHeight(m.tableHeight())
				return m, nil

			case key.Matches(msg, m.keys.Cancel):
				// Exit filter mode and clear the filter
				m.clearFilter()
				m.table.Focus()
				m.table.SetHeight(m.tableHeight())
				// Refresh to show unfiltered data
				m.refreshCurrentView()
				return m, nil

			default:
				// Pass key to text input and apply filter as user types
				m.filterInput, cmd = m.filterInput.Update(msg)
				m.refreshCurrentView()
				return m, cmd
			}
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
				if m.browser != nil && m.selectedStack != nil {
					return m, commands.ListServices(m.browser, *m.selectedStack)
				}
			case TaskList:
				if m.browser != nil && m.selectedService != nil {
					return m, commands.ListTasks(m.browser, *m.selectedService)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			switch m.state {
			case ClusterSelection:
				// Get selected cluster and connect
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.clustersForDisplay) {
					selectedCluster := m.clustersForDisplay[cursor]
					// If it's the same cluster, just go back
					if selectedCluster.Name == m.currentClusterName {
						m.state = m.previousState
						// Restore the appropriate table
						switch m.previousState {
						case StacksList:
							m.showStacksTable(m.stacks, m.selectedStack)
						case ServicesList:
							m.showServicesTable(m.services, m.selectedService)
						case TaskList:
							m.showTasksTable(m.tasks, nil)
						}
						return m, nil
					}

					// Different cluster - disconnect and reconnect
					if m.browser != nil {
						m.browser.Close()
						m.browser = nil
					}
					// Clear navigation state
					m.stacks = nil
					m.selectedStack = nil
					m.services = nil
					m.selectedService = nil
					m.tasks = nil
					// Update current cluster
					m.currentClusterName = selectedCluster.Name
					m.clusterInfo.Cluster = m.conf.Clusters[selectedCluster.Name]
					m.clusterInfo.Status = Connecting
					m.state = Initializing
					return m, commands.ConnectToCluster(m.conf.Clusters[selectedCluster.Name])
				}

			case StacksList:
				// Get selected stack and navigate to services
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.stacks) && m.browser != nil {
					selectedStack := m.stacks[cursor]
					m.clearFilter()
					return m, commands.ListServices(m.browser, selectedStack)
				}
			case ServicesList:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.services) && m.browser != nil {
					selectedService := m.services[cursor]
					m.clearFilter()
					return m, commands.ListTasks(m.browser, selectedService)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Back):
			switch m.state {
			case ClusterSelection:
				// Go back to previous view without changing cluster
				m.state = m.previousState
				m.clearFilter()
				// Restore the appropriate table
				switch m.previousState {
				case StacksList:
					m.showStacksTable(m.stacks, m.selectedStack)
				case ServicesList:
					m.showServicesTable(m.services, m.selectedService)
				case TaskList:
					m.showTasksTable(m.tasks, nil)
				}
				return m, nil
			case ServicesList:
				// Go back to stacks list
				m.state = StacksList
				m.clearFilter()
				m.showStacksTable(m.stacks, m.selectedStack)
				m.selectedStack = nil
				return m, commands.ListStacks(m.browser)
			case TaskList:
				m.state = ServicesList
				m.clearFilter()
				m.showServicesTable(m.services, m.selectedService)
				m.selectedService = nil
			}
			return m, commands.ListServices(m.browser, *m.selectedStack)

		case key.Matches(msg, m.keys.Connect):
			// Connect to container
			switch m.state {
			case ServicesList:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.services) && m.browser != nil {
					selectedService := m.services[cursor]
					return m, commands.AttachToService(m.browser, selectedService)
				}
			case TaskList:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.services) && m.browser != nil {
					selectedTask := m.tasks[cursor]
					return m, commands.AttachToTask(m.browser, selectedTask)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Cluster):
			// Switch cluster - not allowed in container view
			if m.state != ContainerAttached {
				m.previousState = m.state
				m.state = ClusterSelection
				m.clearFilter()
				return m, commands.ListClusters(m.conf.Clusters, m.currentClusterName)
			}
			return m, nil

		case key.Matches(msg, m.keys.Filter):
			// Enter filter mode
			m.filterActive = true
			m.filterInput.Focus()
			m.table.Blur()
			m.table.SetHeight(m.tableHeight())
			return m, textinput.Blink
		}
	}
	if m.containerView != nil {
		*m.containerView, cmd = m.containerView.Update(msg)
		return m, cmd
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

	// Build filter view
	filterView := ""
	if m.filterActive || m.filterInput.Value() != "" {
		filterView = m.filterInput.View()
	}

	// Join all sections vertically
	sections := []string{
		header,
		TableStyle.Render(m.table.View()),
	}

	if filterView != "" {
		sections = append(sections, filterView)
	}

	sections = append(sections, helpView)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderAppHeader() string {
	return AppHeaderStyle.Render(ClusterInfoView(m.clusterInfo))
}

func (m Model) tableHeight() int {
	// Calculate available height by subtracting header and help text
	headerHeight := lipgloss.Height(m.renderAppHeader())
	contextualKeys := NewContextualKeyMap(&m.keys, m.state)
	helpHeight := lipgloss.Height(m.help.View(contextualKeys))

	// Account for filter input if active or has text
	filterHeight := 0
	if m.filterActive || m.filterInput.Value() != "" {
		filterHeight = lipgloss.Height(m.filterInput.View())
	}

	padding := 4 // Some padding for borders and spacing

	availableHeight := m.height - headerHeight - helpHeight - filterHeight - padding

	// Ensure we don't return negative height
	if availableHeight < 1 {
		return 1
	}
	return availableHeight
}

func (m Model) tableWidth() int {
	return m.width - 4
}

// refreshCurrentView refreshes the current view with the filter applied
func (m *Model) refreshCurrentView() {
	filterText := m.filterInput.Value()

	switch m.state {
	case StacksList:
		stacks := m.stacks
		if filterText != "" {
			stacks = m.filterStacks(filterText)
		}
		m.showStacksTable(stacks, m.selectedStack)
	case ServicesList:
		services := m.services
		if filterText != "" {
			services = m.filterServices(filterText)
		}
		m.showServicesTable(services, m.selectedService)
	case TaskList:
		tasks := m.tasks
		if filterText != "" {
			tasks = m.filterTasks(filterText)
		}
		m.showTasksTable(tasks, nil)
	case ClusterSelection:
		clusters := m.clustersForDisplay
		if filterText != "" {
			clusters = m.filterClusters(filterText)
		}
		m.showClustersTable(clusters, m.currentClusterName)
	}
}

// filterStacks filters stacks by name (case-insensitive)
func (m *Model) filterStacks(filterText string) []models.Stack {
	filterLower := strings.ToLower(filterText)
	filtered := make([]models.Stack, 0)

	for _, stack := range m.stacks {
		if strings.Contains(strings.ToLower(stack.Name), filterLower) {
			filtered = append(filtered, stack)
		}
	}

	return filtered
}

// filterServices filters services by name or ID (case-insensitive)
func (m *Model) filterServices(filterText string) []models.Service {
	filterLower := strings.ToLower(filterText)
	filtered := make([]models.Service, 0)

	for _, service := range m.services {
		if strings.Contains(strings.ToLower(service.Name), filterLower) ||
			strings.Contains(strings.ToLower(service.ID), filterLower) {
			filtered = append(filtered, service)
		}
	}

	return filtered
}

// filterTasks filters tasks by ID, container ID, status, or node (case-insensitive)
func (m *Model) filterTasks(filterText string) []models.Task {
	filterLower := strings.ToLower(filterText)
	filtered := make([]models.Task, 0)

	for _, task := range m.tasks {
		if strings.Contains(strings.ToLower(task.TaskID), filterLower) ||
			strings.Contains(strings.ToLower(task.ContainerID), filterLower) ||
			strings.Contains(strings.ToLower(string(task.Status)), filterLower) ||
			strings.Contains(strings.ToLower(task.Node.Host), filterLower) {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// filterClusters filters clusters by name or host (case-insensitive)
func (m *Model) filterClusters(filterText string) []commands.ClusterTableRow {
	filterLower := strings.ToLower(filterText)
	filtered := make([]commands.ClusterTableRow, 0)

	for _, cluster := range m.clustersForDisplay {
		if strings.Contains(strings.ToLower(cluster.Name), filterLower) ||
			strings.Contains(strings.ToLower(cluster.Host), filterLower) {
			filtered = append(filtered, cluster)
		}
	}

	return filtered
}

// clearFilter clears the filter input and resets filter state
func (m *Model) clearFilter() {
	m.filterActive = false
	m.filterInput.SetValue("")
	m.filterInput.Blur()
}

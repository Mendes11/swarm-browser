package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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

	// Filtered lists (what's currently displayed in the table)
	filteredStacks   []models.Stack
	filteredServices []models.Service
	filteredTasks    []models.Task
	filteredClusters []commands.ClusterTableRow

	// Cluster selection state
	clustersForDisplay []commands.ClusterTableRow
	previousState      ViewState
	currentClusterName string

	// Container session
	containerConn     core.ContainerConnection
	spinner           spinner.Model
	attachingTarget   string
	attachError       error
	stateBeforeAttach ViewState

	// Command picker state
	commandPickerItems []commandPickerItem
	customCmdInput     textinput.Model
	commandTarget      commandTarget
	appState           config.AppState
	clustersConfig     *config.ClustersConfig
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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorStatusPending)

	// Initialize custom command text input
	customCmdInput := textinput.New()
	customCmdInput.Placeholder = "e.g., rails console, psql -U postgres"
	customCmdInput.Prompt = "> "
	customCmdInput.CharLimit = 200
	customCmdInput.Width = 60

	// Load app state for custom commands and last-selected commands
	appState, _ := config.LoadState()

	return Model{
		conf:               conf,
		state:              Initializing,
		clusterInfo:        clusterInfo,
		table:              newTable(keys.Table),
		keys:               keys,
		help:               help.New(),
		filterInput:        filterInput,
		currentClusterName: conf.InitialCluster,
		spinner:            s,
		customCmdInput:     customCmdInput,
		appState:           appState,
		clustersConfig:     newClustersConfig(conf),
	}
}

func (m Model) Close() error {
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
		return m, tea.Batch(
			commands.ListStacks(m.browser),
			commands.SaveLastCluster(m.currentClusterName),
		)

	case commands.StacksUpdated:
		m.state = StacksList

		var selectedStack *models.Stack
		if m.state == StacksList && m.table.Cursor() < len(m.stacks) {
			selectedStack = &m.stacks[m.table.Cursor()]
		}
		m.stacks = msg.Stacks
		m.filteredStacks = msg.Stacks
		m.showStacksTable(msg.Stacks, selectedStack)
		return m, nil

	case commands.ServicesUpdated:
		m.state = ServicesList
		m.services = msg.Services
		m.filteredServices = msg.Services
		m.selectedStack = &msg.Stack
		m.showServicesTable(msg.Services, nil)
		return m, nil
	case commands.TasksUpdated:
		m.state = TaskList
		m.tasks = msg.Tasks
		m.filteredTasks = msg.Tasks
		m.selectedService = &msg.Service
		m.showTasksTable(msg.Tasks, nil)
		return m, nil

	case commands.ClusterConnectionFailed:
		m.clusterInfo.Err = msg.Err
		m.clusterInfo.Status = Disconnected
		return m, nil

	case commands.ContainerAttachedMsg:
		// Successfully attached to container — hand off terminal via tea.Exec
		m.state = ContainerAttached
		m.containerConn = msg.Conn
		execCmd := NewContainerExecCmd(msg.Conn)
		return m, tea.Exec(execCmd, func(err error) tea.Msg {
			return ExitContainerViewMsg{Err: err}
		})

	case ExitContainerViewMsg:
		// Container session ended — return to the view we came from
		m.state = m.stateBeforeAttach
		if m.containerConn != nil {
			m.containerConn.Close()
		}
		m.containerConn = nil
		m.commandTarget = commandTarget{}
		// Restore the table for the state we're returning to
		switch m.state {
		case TaskList:
			m.showTasksTable(m.filteredTasks, nil)
		default:
			m.state = ServicesList
			m.showServicesTable(m.filteredServices, m.selectedService)
		}
		return m, nil

	case commands.ClustersListed:
		m.clustersForDisplay = msg.Clusters
		m.filteredClusters = msg.Clusters
		m.showClustersTable(msg.Clusters, msg.CurrentCluster)
		return m, nil

	case commands.StateSavedMsg:
		if msg.Err != nil {
			log.Printf("Warning: failed to save state: %v", msg.Err)
		} else {
			// Reload state to pick up saved custom commands / last commands
			m.appState, _ = config.LoadState()
		}
		return m, nil

	case commands.ContainerDetachedMsg:
		m.state = m.stateBeforeAttach
		m.attachError = msg.Err
		m.attachingTarget = ""
		m.commandTarget = commandTarget{}
		log.Printf("Container attachment failed: %v", msg.Err)
		// Restore the table for the state we're returning to
		switch m.state {
		case TaskList:
			m.showTasksTable(m.filteredTasks, nil)
		default:
			m.state = ServicesList
			m.showServicesTable(m.filteredServices, m.selectedService)
		}
		return m, nil

	case spinner.TickMsg:
		if m.state == ContainerAttaching {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Clear attachment error on any key press
		if m.attachError != nil {
			m.attachError = nil
		}

		// During container attachment, only allow quit
		if m.state == ContainerAttaching {
			if key.Matches(msg, m.keys.Quit) {
				return m, tea.Quit
			}
			return m, nil
		}

		// Handle custom command input mode
		if m.state == CustomCommandInput {
			switch {
			case key.Matches(msg, m.keys.Enter):
				cmdText := m.customCmdInput.Value()
				if cmdText == "" {
					return m, nil
				}
				item := commandPickerItem{
					Name:   cmdText,
					Cmd:    []string{"sh", "-c", cmdText},
					Source: "history",
				}
				m.customCmdInput.SetValue("")
				m.customCmdInput.Blur()
				stateKey := m.customCommandKey()
				return m, tea.Batch(
					m.executeAttach(item),
					commands.SaveCustomCommand(stateKey, cmdText),
				)
			case key.Matches(msg, m.keys.Cancel):
				// Go back to command selection
				m.state = CommandSelection
				m.customCmdInput.SetValue("")
				m.customCmdInput.Blur()
				m.table.Focus()
				m.showCommandPickerTable(m.commandPickerItems, m.lastSelectedCommand())
				return m, nil
			default:
				m.customCmdInput, cmd = m.customCmdInput.Update(msg)
				return m, cmd
			}
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
			case CommandSelection:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.commandPickerItems) {
					selected := m.commandPickerItems[cursor]
					if selected.IsCustom {
						// Transition to custom command text input
						m.state = CustomCommandInput
						m.customCmdInput.Focus()
						m.table.Blur()
						return m, textinput.Blink
					}
					// Execute the selected command
					return m, m.executeAttach(selected)
				}
				return m, nil

			case ClusterSelection:
				// Get selected cluster and connect
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.filteredClusters) {
					selectedCluster := m.filteredClusters[cursor]
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
					return m, tea.Batch(
						commands.ConnectToCluster(m.conf.Clusters[selectedCluster.Name]),
						commands.SaveLastCluster(selectedCluster.Name),
					)
				}

			case StacksList:
				// Get selected stack and navigate to services
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.filteredStacks) && m.browser != nil {
					selectedStack := m.filteredStacks[cursor]
					m.clearFilter()
					return m, commands.ListServices(m.browser, selectedStack)
				}
			case ServicesList:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.filteredServices) && m.browser != nil {
					selectedService := m.filteredServices[cursor]
					m.clearFilter()
					return m, commands.ListTasks(m.browser, selectedService)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Cancel):
			if m.state == CommandSelection {
				m.state = m.commandTarget.Source
				switch m.commandTarget.Source {
				case ServicesList:
					m.showServicesTable(m.filteredServices, m.selectedService)
				case TaskList:
					m.showTasksTable(m.filteredTasks, nil)
				}
				m.commandTarget = commandTarget{}
				return m, nil
			}

		case key.Matches(msg, m.keys.Back):
			switch m.state {
			case CommandSelection:
				// Go back to the view we came from
				m.state = m.commandTarget.Source
				switch m.commandTarget.Source {
				case ServicesList:
					m.showServicesTable(m.filteredServices, m.selectedService)
				case TaskList:
					m.showTasksTable(m.filteredTasks, nil)
				}
				m.commandTarget = commandTarget{}
				return m, nil
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
			// Open command picker before connecting
			switch m.state {
			case ServicesList:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.filteredServices) && m.browser != nil {
					selectedService := m.filteredServices[cursor]
					m.commandTarget = commandTarget{
						Service: &selectedService,
						Source:  ServicesList,
					}
					m.commandPickerItems = m.buildCommandPickerItems(selectedService.Name)
					m.state = CommandSelection
					m.showCommandPickerTable(m.commandPickerItems, m.lastSelectedCommand())
				}
			case TaskList:
				cursor := m.table.Cursor()
				if cursor >= 0 && cursor < len(m.filteredTasks) && m.browser != nil {
					selectedTask := m.filteredTasks[cursor]
					serviceName := ""
					if m.selectedService != nil {
						serviceName = m.selectedService.Name
					}
					m.commandTarget = commandTarget{
						Task:   &selectedTask,
						Source: TaskList,
					}
					m.commandPickerItems = m.buildCommandPickerItems(serviceName)
					m.state = CommandSelection
					m.showCommandPickerTable(m.commandPickerItems, m.lastSelectedCommand())
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Cluster):
			// Switch cluster - not allowed in container view or while attaching
			if m.state != ContainerAttached && m.state != ContainerAttaching {
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
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	header := ClusterInfoView(m.clusterInfo)

	// Create contextual keymap for help display
	contextualKeys := NewContextualKeyMap(&m.keys, m.state)
	helpView := m.help.View(contextualKeys)

	// Build main content
	var mainContent string
	switch m.state {
	case ContainerAttaching:
		msg := fmt.Sprintf("%s Connecting to %s...", m.spinner.View(), m.attachingTarget)
		mainContent = lipgloss.NewStyle().
			Foreground(ColorStatusPending).
			Padding(2, 4).
			Render(msg)
	case CustomCommandInput:
		prompt := lipgloss.NewStyle().
			Foreground(ColorLabel).
			Padding(1, 2).
			Render("Enter command to execute in container:")
		mainContent = lipgloss.JoinVertical(lipgloss.Left,
			TableStyle.Render(m.table.View()),
			prompt,
			m.customCmdInput.View(),
		)
	default:
		mainContent = TableStyle.Render(m.table.View())
	}

	// Build filter view
	filterView := ""
	if m.filterActive || m.filterInput.Value() != "" {
		filterView = m.filterInput.View()
	}

	// Join all sections vertically
	sections := []string{
		header,
		mainContent,
	}

	if filterView != "" {
		sections = append(sections, filterView)
	}

	if m.attachError != nil {
		errorView := lipgloss.NewStyle().
			Foreground(ColorError).
			Padding(0, 2).
			Render(fmt.Sprintf("Error: %v", m.attachError))
		sections = append(sections, errorView)
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
		m.filteredStacks = m.stacks
		if filterText != "" {
			m.filteredStacks = m.filterStacks(filterText)
		}
		m.showStacksTable(m.filteredStacks, m.selectedStack)
	case ServicesList:
		m.filteredServices = m.services
		if filterText != "" {
			m.filteredServices = m.filterServices(filterText)
		}
		m.showServicesTable(m.filteredServices, m.selectedService)
	case TaskList:
		m.filteredTasks = m.tasks
		if filterText != "" {
			m.filteredTasks = m.filterTasks(filterText)
		}
		m.showTasksTable(m.filteredTasks, nil)
	case ClusterSelection:
		m.filteredClusters = m.clustersForDisplay
		if filterText != "" {
			m.filteredClusters = m.filterClusters(filterText)
		}
		m.showClustersTable(m.filteredClusters, m.currentClusterName)
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

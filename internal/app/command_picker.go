package app

import (
	"fmt"
	"path"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/commands"
	"github.com/mendes11/swarm-browser/internal/config"
	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/mendes11/swarm-browser/internal/shell"
)

// commandPickerItem represents a row in the command picker table.
type commandPickerItem struct {
	Name     string
	Cmd      []string
	IsShell  bool   // uses bash/sh fallback
	IsCustom bool   // "Custom command..." entry
	Source   string // "config", "history", "builtin"
}

// commandTarget stores what we're connecting to, so we can resume after command selection.
type commandTarget struct {
	Service *models.Service
	Task    *models.Task
	Source  ViewState // where to return on cancel
}

// buildCommandPickerItems builds the list of commands available for a given service.
func (m *Model) buildCommandPickerItems(serviceName string) []commandPickerItem {
	items := []commandPickerItem{}

	// 1. Resolved commands from config (global + cluster merge, filtered by service name)
	resolved := m.clustersConfig.ResolveCommandsForCluster(m.currentClusterName)
	for _, cmd := range resolved {
		if matchesService(cmd.MatchServices, serviceName) {
			items = append(items, commandPickerItem{
				Name:   cmd.Name,
				Cmd:    []string{shell.UserShell(), "-c", cmd.Cmd},
				Source: "config",
			})
		}
	}

	// 2. Custom commands from history for this context
	key := m.customCommandKey()
	if history, ok := m.appState.CustomCommands[key]; ok {
		for _, cmd := range history {
			items = append(items, commandPickerItem{
				Name:   cmd,
				Cmd:    []string{shell.UserShell(), "-c", cmd},
				Source: "history",
			})
		}
	}

	// 3. Shell (always present)
	items = append(items, commandPickerItem{
		Name:    "Shell",
		Cmd:     []string{shell.UserShell()},
		IsShell: true,
		Source:  "builtin",
	})

	// 4. Custom command... (always last)
	items = append(items, commandPickerItem{
		Name:     "Custom command...",
		IsCustom: true,
		Source:   "builtin",
	})

	return items
}

// customCommandKey returns the state key for the current context.
func (m *Model) customCommandKey() string {
	cluster := m.currentClusterName
	stack := ""
	service := ""
	if m.selectedStack != nil {
		stack = m.selectedStack.Name
	}
	if m.commandTarget.Service != nil {
		service = m.commandTarget.Service.Name
	} else if m.commandTarget.Task != nil && m.selectedService != nil {
		service = m.selectedService.Name
	}
	return fmt.Sprintf("%s:%s:%s", cluster, stack, service)
}

// matchesService checks if a service name matches any of the given glob patterns.
func matchesService(patterns []string, serviceName string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		matched, err := path.Match(pattern, serviceName)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// lastSelectedCommand returns the last-selected command name for the current context.
func (m *Model) lastSelectedCommand() string {
	key := m.customCommandKey()
	if cmd, ok := m.appState.LastCommands[key]; ok {
		return cmd
	}
	return ""
}

// executeAttach starts the container attachment with the selected command.
func (m *Model) executeAttach(item commandPickerItem) tea.Cmd {
	m.stateBeforeAttach = m.commandTarget.Source
	m.state = ContainerAttaching
	m.attachError = nil

	stateKey := m.customCommandKey()

	if m.commandTarget.Service != nil {
		service := *m.commandTarget.Service
		m.attachingTarget = service.Name
		if item.IsShell {
			return tea.Batch(
				commands.AttachToService(m.browser, service),
				commands.SaveLastCommand(stateKey, item.Name),
				m.spinner.Tick,
			)
		}
		return tea.Batch(
			commands.AttachToServiceWithCmd(m.browser, service, item.Cmd),
			commands.SaveLastCommand(stateKey, item.Name),
			m.spinner.Tick,
		)
	}
	if m.commandTarget.Task != nil {
		task := *m.commandTarget.Task
		m.attachingTarget = task.TaskID
		if item.IsShell {
			return tea.Batch(
				commands.AttachToTask(m.browser, task),
				commands.SaveLastCommand(stateKey, item.Name),
				m.spinner.Tick,
			)
		}
		return tea.Batch(
			commands.AttachToTaskWithCmd(m.browser, task, item.Cmd),
			commands.SaveLastCommand(stateKey, item.Name),
			m.spinner.Tick,
		)
	}
	return nil
}

// newClustersConfig creates a ClustersConfig needed for command resolution.
func newClustersConfig(conf config.Config) *config.ClustersConfig {
	return &config.ClustersConfig{
		Clusters: conf.Clusters,
		Commands: conf.Commands,
	}
}

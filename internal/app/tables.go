package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

func (m *Model) showStacksTable(stacks []models.Stack, selectedStack *models.Stack) {
	rows := make([]table.Row, len(stacks))
	cursor := 0
	for i, stack := range stacks {
		rows[i] = []string{stack.Name}
		if selectedStack != nil && stack == *selectedStack {
			cursor = i
		}
	}

	m.table = newTable(m.keys.Table)
	m.table.SetHeight(m.tableHeight())
	m.table.SetWidth(m.tableWidth())

	m.table.SetColumns([]table.Column{{Title: "Name", Width: m.table.Width()}})
	m.table.SetRows(rows)
	m.table.SetCursor(cursor)
}

func (m *Model) showServicesTable(services []models.Service, selectedService *models.Service) {
	rows := make([]table.Row, len(services))
	cursor := 0
	for i, service := range services {
		rows[i] = []string{
			service.ID,
			service.Name,
			fmt.Sprintf("%d/%d", service.RunningTasks, service.DesiredTasks),
		}
		if selectedService != nil && *selectedService == service {
			cursor = i
		}
	}
	m.table = newTable(m.keys.Table)
	m.table.SetWidth(m.tableWidth())
	m.table.SetHeight(m.tableHeight())
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
	m.table.SetCursor(cursor)
}

func (m *Model) showTasksTable(tasks []models.Task, selectedTask *models.Task) {
	rows := make([]table.Row, len(tasks))
	cursor := 0
	for i, task := range tasks {
		rows[i] = []string{
			task.TaskID,
			task.ContainerID,
			fmt.Sprintf("%s", task.Status),
			task.Node.Host,
		}
		if selectedTask != nil && *selectedTask == task {
			cursor = i
		}
	}

	m.table = newTable(m.keys.Table)
	m.table.SetWidth(m.tableWidth())
	m.table.SetHeight(m.tableHeight())
	// Calculate column widths based on table width
	tableWidth := m.table.Width()
	idWidth := 20
	containerIdWidth := 20
	statusWidth := 12
	nodeWidth := tableWidth - idWidth - containerIdWidth - statusWidth - 4 // Account for borders

	m.table.SetColumns([]table.Column{
		{Title: "ID", Width: idWidth},
		{Title: "ContainerID", Width: containerIdWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Node", Width: nodeWidth},
	})
	m.table.SetRows(rows)
	m.table.SetCursor(cursor)
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

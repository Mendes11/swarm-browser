package tabs

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// var emptySpaceStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true)

type Tab struct {
	Name  string
	Model tea.Model
}

type Tabs struct {
	Tabs      []Tab
	ActiveIdx int
	Width     int
	keys      keyMap
}

type keyMap struct {
	PrevTab key.Binding
	NextTab key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		PrevTab: key.NewBinding(
			key.WithKeys("ctrl+["),
			key.WithHelp("ctrl+[", "previous tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("ctrl+]"),
			key.WithHelp("ctrl+]", "next tab"),
		),
	}
}

func NewTab(tabs ...Tab) Tabs {
	return Tabs{
		Tabs:      tabs,
		ActiveIdx: -1,
		keys:      defaultKeyMap(),
	}
}

func (t Tabs) Init() tea.Cmd {
	return nil
}

func (t Tabs) Update(msg tea.Msg) (Tabs, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle tab switching with Ctrl+[ and Ctrl+]
		switch {
		case key.Matches(msg, t.keys.PrevTab):
			// Move left
			if t.ActiveIdx > 0 {
				t.ActiveIdx--
			}
			return t, nil
		case key.Matches(msg, t.keys.NextTab):
			// Move right
			if t.ActiveIdx < len(t.Tabs)-1 {
				t.ActiveIdx++
			}
			return t, nil
		}
	}

	// Delegate to active tab's model
	if len(t.Tabs) > 0 && t.ActiveIdx >= 0 {
		model, cmd := t.Tabs[t.ActiveIdx].Model.Update(msg)
		t.Tabs[t.ActiveIdx].Model = model
		return t, cmd
	}
	return t, nil
}

// ShortHelp returns the keybindings for the short help view
func (t Tabs) ShortHelp() []key.Binding {
	return []key.Binding{t.keys.PrevTab, t.keys.NextTab}
}

func (t Tabs) View() string {
	header := t.renderHeader()
	content := t.renderTab()

	return lipgloss.JoinVertical(lipgloss.Top, header, content)
}

func (t Tabs) renderTab() string {
	content := ""
	if len(t.Tabs) > 0 && t.ActiveIdx >= 0 {
		content = activeTabStyle.Render(t.Tabs[t.ActiveIdx].Model.View())
	}
	return content
}

func (t Tabs) renderHeader() string {
	totalWidth := t.Width
	tabContent := make([]string, len(t.Tabs))
	for idx, tab := range t.Tabs {
		if idx == t.ActiveIdx {
			tabContent[idx] = activeTabStyle.Render(tab.Name)
		} else {
			tabContent[idx] = tabStyle.Render(tab.Name)
		}
		totalWidth -= lipgloss.Width(tabContent[idx])
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, tabContent...)

	// Fill remaining width with bottom border
	if totalWidth > 0 {
		borderFill := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true).
			Width(totalWidth - 1).
			Render("")
		tabs = lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, borderFill)
	}

	return tabs + "╮"
}

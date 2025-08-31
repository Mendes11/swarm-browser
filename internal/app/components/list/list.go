package list

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mendes11/swarm-browser/internal/app/styles"
)

var (
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	CursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	NormalStyle   = lipgloss.NewStyle()
)

type List[T fmt.Stringer] struct {
	items      []T
	cursor     int
	selected   bool
	selectedId int
}

func New[T fmt.Stringer](items []T) List[T] {
	return List[T]{
		items: items,
	}
}

func (s List[T]) Init() tea.Cmd {
	return nil
}

func (s List[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.items)-1 {
				s.cursor++
			}
		case "enter":
			if len(s.items) > 0 {
				s.selected = true
				return s, selectedCommand(s.items[s.cursor])
			}
		}
	}
	return s, nil
}

func (s List[T]) View() string {
	var sb strings.Builder

	for i, item := range s.items {
		style := styles.NormalStyle
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
			style = styles.CursorStyle
			if s.selected {
				style = styles.SelectedStyle
			}
		}
		sb.WriteString(style.Render(fmt.Sprintf("%s%d. %s", cursor, i+1, item.String())))
		sb.WriteString("\n")
	}

	return sb.String()
}

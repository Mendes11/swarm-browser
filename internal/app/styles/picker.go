package styles

import "github.com/charmbracelet/lipgloss"

var SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
var CursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var NormalStyle = lipgloss.NewStyle()

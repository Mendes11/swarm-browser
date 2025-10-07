package tabs

import "github.com/charmbracelet/lipgloss"

var highlight = lipgloss.AdaptiveColor{Light: "#347aeb", Dark: "#347aeb"}

// Define styles for active and inactive tabs
var tabStyle = lipgloss.NewStyle().
	Border(lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}, true).
	Padding(0, 1)

var activeTabStyle = tabStyle.
	Border(lipgloss.Border{
		Top:         "─",
		Bottom:      " ", // No bottom border for active tab
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "│",
		BottomRight: "└",
	}, true).
	BorderForeground(highlight)

// Style for the active tab content
var activeTabContentStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), false, true, true, true).
	BorderForeground(highlight).
	Padding(1, 2)


package app

import "github.com/charmbracelet/lipgloss"

var LabelStyle = lipgloss.NewStyle().Foreground(ColorLabel).Bold(true)
var TextStyle = lipgloss.NewStyle().Foreground(ColorTextBody)
var DisconnectedStyle = lipgloss.NewStyle().Foreground(ColorStatusStopped)
var ConnectedStyle = lipgloss.NewStyle().Foreground(ColorStatusRunning)
var ConnectingStyle = lipgloss.NewStyle().Foreground(ColorStatusPending)

var TableStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var AppHeaderStyle = lipgloss.NewStyle().
	PaddingLeft(2)

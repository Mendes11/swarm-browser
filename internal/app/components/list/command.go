package list

import (
	tea "github.com/charmbracelet/bubbletea"
)

type SelectedMsg[T any] struct {
	Item T
}

func selectedCommand[T any](item T) tea.Cmd {
	return func() tea.Msg {
		return SelectedMsg[T]{Item: item}
	}
}

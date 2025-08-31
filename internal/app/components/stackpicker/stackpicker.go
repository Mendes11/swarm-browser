package stackpicker

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/components/list"
	"github.com/mendes11/swarm-browser/internal/app/types"
)

type StackPicker struct {
	stacks []types.Stack
	list   list.List[types.Stack]
}

func NewStackPicker(stacks []types.Stack) StackPicker {
	return StackPicker{
		stacks: stacks,
		list:   list.New(stacks),
	}
}

func (s StackPicker) Init() tea.Cmd {
	return nil
}

func (s StackPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("Stack Picker Update: %v", msg)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}
	input, cmd := s.list.Update(msg)
	s.list = input.(list.List[types.Stack])
	return s, cmd
}

func (s StackPicker) View() string {
	if len(s.stacks) == 0 {
		return "No stacks available\n\n"
	}

	var sb strings.Builder
	sb.WriteString("Select a Docker Swarm Stack:\n\n")
	sb.WriteString(s.list.View() + "\n")
	return sb.String()
}

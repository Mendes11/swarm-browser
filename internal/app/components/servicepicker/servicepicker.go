package servicepicker

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/components/list"
	"github.com/mendes11/swarm-browser/internal/app/types"
)

type ServicePicker struct {
	services []types.Service
	list     list.List[types.Service]
}

func New(services []types.Service) ServicePicker {
	return ServicePicker{
		services: services,
		list:     list.New(services),
	}
}

func (s ServicePicker) Init() tea.Cmd {
	return nil
}

func (s ServicePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("Service Picker Update: %v", msg)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}
	input, cmd := s.list.Update(msg)
	s.list = input.(list.List[types.Service])
	return s, cmd
}

func (s ServicePicker) View() string {
	var sb strings.Builder

	sb.WriteString("Select a Service:\n\n")
	sb.WriteString(s.list.View() + "\n")
	return sb.String()
}

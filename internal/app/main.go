package app

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/components/list"
	"github.com/mendes11/swarm-browser/internal/app/components/servicepicker"
	"github.com/mendes11/swarm-browser/internal/app/components/stackpicker"
	"github.com/mendes11/swarm-browser/internal/app/styles"
	"github.com/mendes11/swarm-browser/internal/app/types"
	"github.com/moby/moby/client"
)

type ViewState int

const (
	ViewStackPicker ViewState = iota
	ViewServices
)

type App struct {
	cli           *client.Client
	viewState     ViewState
	stackPicker   stackpicker.StackPicker
	servicePicker servicepicker.ServicePicker

	selectedStack   *types.Stack
	selectedService *types.Service
}

func NewApp(cli *client.Client) *App {
	return &App{
		cli:         cli,
		viewState:   ViewStackPicker,
		stackPicker: stackpicker.NewStackPicker([]types.Stack{}),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		stackpicker.ListStacksCommandFactory(context.Background(), a.cli),
		a.stackPicker.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch a.viewState {
		case ViewServices:
			switch msg.String() {
			case "esc", "backspace":
				a.viewState = ViewStackPicker
				a.selectedStack = nil
				return a, nil
			case "ctrl+c", "q":
				return a, tea.Quit
			}
		}
	case list.SelectedMsg[types.Stack]:
		a.selectedStack = &msg.Item
		a.viewState = ViewServices
		a.servicePicker = servicepicker.New(a.selectedStack.Services)
		return a, nil
	case list.SelectedMsg[types.Service]:
		a.selectedService = &msg.Item
		// TODO: Generate a command to connect to the one running
		// container of the service
		// The flow is the command should connect, and the msg must contain the
		// I/O elements to communicate with the container (stdin, stdout, stderr)
		return a, nil
	case stackpicker.StacksUpdated:
		// Update the stackpicker with new stacks
		a.stackPicker = stackpicker.NewStackPicker(msg.Stacks)
		return a, nil
	}

	// Delegate to the appropriate model based on view state
	switch a.viewState {
	case ViewStackPicker:
		updatedPicker, cmd := a.stackPicker.Update(msg)
		a.stackPicker = updatedPicker.(stackpicker.StackPicker)
		return a, cmd
	case ViewServices:
		updatedPicker, cmd := a.servicePicker.Update(msg)
		a.servicePicker = updatedPicker.(servicepicker.ServicePicker)
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	var sb strings.Builder
	switch a.viewState {
	case ViewStackPicker:
		sb.WriteString(a.stackPicker.View() + "\n")
		sb.WriteString(styles.HelpTextStyle.Render("\n← Backspace/ESC: Back to stack selection • ↑/↓ or j/k: Navigate • Enter: Select • q: Quit"))
	case ViewServices:
		sb.WriteString(a.servicePicker.View() + "\n")
		sb.WriteString(styles.HelpTextStyle.Render("\n← Backspace/ESC: Back to stack selection • ↑/↓ or j/k: Navigate • Enter: Select • q: Quit"))
	default:
		sb.WriteString("Unknown view state")
	}
	return sb.String()
}

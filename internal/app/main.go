package app

import (
	"context"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app/commands"
	"github.com/mendes11/swarm-browser/internal/app/components/list"
	"github.com/mendes11/swarm-browser/internal/app/components/servicepicker"
	"github.com/mendes11/swarm-browser/internal/app/components/stackpicker"
	"github.com/mendes11/swarm-browser/internal/app/styles"
	"github.com/mendes11/swarm-browser/internal/app/types"
	"github.com/mendes11/swarm-browser/internal/config"
	"github.com/mendes11/swarm-browser/internal/services/connector"
)

type ViewState int

const (
	ViewStackPicker ViewState = iota
	ViewServices
)

type App struct {
	conn          *connector.Connector
	conf          config.Config
	viewState     ViewState
	stackPicker   stackpicker.StackPicker
	servicePicker servicepicker.ServicePicker

	cluster         *config.Cluster
	selectedStack   *types.Stack
	selectedService *types.Service
}

func NewApp(conf config.Config, conn *connector.Connector) *App {
	return &App{
		conn:        conn,
		conf:        conf,
		viewState:   ViewStackPicker,
		stackPicker: stackpicker.NewStackPicker([]types.Stack{}),
	}
}

func (a App) Init() tea.Cmd {
	cluster := func() config.Cluster {
		// Pick first cluster
		for _, v := range a.conf.Clusters {
			return v
		}
		return config.Cluster{}
	}()
	return commands.ConnectToClusterCommand(context.Background(), a.conn, cluster)
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
		return a, commands.ConnectToServiceCommand(context.Background(), a.conn, *a.cluster, *a.selectedService)
	case commands.ConnectedToClusterMsg:
		// Update the stackpicker with new stacks
		a.stackPicker = stackpicker.NewStackPicker(msg.Stacks)
		a.cluster = &msg.Cluster
		return a, nil
	case commands.ConnectionErrorMsg:
		log.Printf("Error connecting to cluster %+v: %v\n", msg.Cluster, msg.Err)
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

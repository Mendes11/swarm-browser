package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
)

// AppKeyMap defines both table and app-specific keybindings
type AppKeyMap struct {
	// Table navigation (from bubbles table)
	Table table.KeyMap

	// App-specific actions
	Back    key.Binding
	Refresh key.Binding
	Cluster key.Binding
	Connect key.Binding
	Filter  key.Binding
	Enter   key.Binding
	Cancel  key.Binding

	// Application
	Help key.Binding
	Quit key.Binding
}

// DefaultAppKeyMap returns the default keybindings combining table and app keys
func DefaultAppKeyMap() AppKeyMap {
	return AppKeyMap{
		// Use the default table keymap
		Table: table.DefaultKeyMap(),

		// App-specific actions
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "back"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Cluster: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "switch cluster"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Connect: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "connect"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),

		// Application
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k AppKeyMap) ShortHelp() []key.Binding {
	// Show the most important keys in compact view
	return []key.Binding{
		k.Table.LineUp,
		k.Table.LineDown,
		k.Enter,
		k.Back,
		k.Help,
		k.Quit,
	}
}

// ShortHelpForView returns context-aware keybindings based on the current view
func (k AppKeyMap) ShortHelpForView(viewState ViewState) []key.Binding {
	switch viewState {
	case StacksList:
		return []key.Binding{
			k.Table.LineUp,
			k.Table.LineDown,
			k.Enter,
			k.Refresh,
			k.Help,
			k.Quit,
		}
	case ServicesList:
		return []key.Binding{
			k.Table.LineUp,
			k.Table.LineDown,
			k.Enter,
			k.Back,
			k.Refresh,
			k.Help,
			k.Quit,
		}
	case TaskList:
		return []key.Binding{
			k.Table.LineUp,
			k.Table.LineDown,
			k.Enter,
			k.Back,
			k.Connect,
			k.Help,
			k.Quit,
		}
	default:
		return k.ShortHelp()
	}
}

// FullHelp returns keybindings for the expanded help view
func (k AppKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Table navigation
		{
			k.Table.LineUp,
			k.Table.LineDown,
			k.Table.PageUp,
			k.Table.PageDown,
		},
		// More table navigation
		{
			k.Table.GotoTop,
			k.Table.GotoBottom,
		},
		// App actions
		{k.Enter, k.Back, k.Refresh, k.Connect, k.Filter},
		// App controls
		{k.Help, k.Quit},
	}
}

// FullHelpForView returns context-aware full help based on the current view
func (k AppKeyMap) FullHelpForView(viewState ViewState) [][]key.Binding {
	switch viewState {
	case StacksList:
		// In stacks list, don't show back, but show connect
		return [][]key.Binding{
			// Table navigation
			{
				k.Table.LineUp,
				k.Table.LineDown,
				k.Table.PageUp,
				k.Table.PageDown,
			},
			// More table navigation
			{
				k.Table.GotoTop,
				k.Table.GotoBottom,
			},
			// App actions - no back in stacks list
			{k.Enter, k.Refresh, k.Connect, k.Filter},
			// App controls
			{k.Help, k.Quit},
		}
	case ServicesList:
		// In services list, show back but not connect
		return [][]key.Binding{
			// Table navigation
			{
				k.Table.LineUp,
				k.Table.LineDown,
				k.Table.PageUp,
				k.Table.PageDown,
			},
			// More table navigation
			{
				k.Table.GotoTop,
				k.Table.GotoBottom,
			},
			// App actions - show back but not connect
			{k.Enter, k.Back, k.Refresh, k.Filter},
			// App controls
			{k.Help, k.Quit},
		}
	case TaskList:
		// In task list, show both back and connect
		return [][]key.Binding{
			// Table navigation
			{
				k.Table.LineUp,
				k.Table.LineDown,
				k.Table.PageUp,
				k.Table.PageDown,
			},
			// More table navigation
			{
				k.Table.GotoTop,
				k.Table.GotoBottom,
			},
			// App actions
			{k.Enter, k.Back, k.Refresh, k.Connect, k.Filter},
			// App controls
			{k.Help, k.Quit},
		}
	default:
		return k.FullHelp()
	}
}

// SetEnterHelpText updates the Enter key's help text based on the current view
func (k *AppKeyMap) SetEnterHelpText(viewState ViewState) {
	switch viewState {
	case StacksList:
		k.Enter.SetHelp("enter", "view services")
	case ServicesList:
		k.Enter.SetHelp("enter", "view tasks")
	case TaskList:
		k.Enter.SetHelp("enter", "attach to container")
	default:
		k.Enter.SetHelp("enter", "select")
	}
}

// ContextualKeyMap wraps AppKeyMap to provide context-aware help
type ContextualKeyMap struct {
	*AppKeyMap
	ViewState ViewState
}

// NewContextualKeyMap creates a new contextual keymap
func NewContextualKeyMap(keys *AppKeyMap, viewState ViewState) ContextualKeyMap {
	keys.SetEnterHelpText(viewState)
	return ContextualKeyMap{
		AppKeyMap: keys,
		ViewState: viewState,
	}
}

// ShortHelp returns context-aware short help
func (k ContextualKeyMap) ShortHelp() []key.Binding {
	return k.AppKeyMap.ShortHelpForView(k.ViewState)
}

// FullHelp returns context-aware full help
func (k ContextualKeyMap) FullHelp() [][]key.Binding {
	return k.AppKeyMap.FullHelpForView(k.ViewState)
}

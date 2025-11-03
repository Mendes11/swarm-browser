package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// KeyToBytes converts a Bubbletea KeyMsg to the actual bytes that should be sent to a terminal
func KeyToBytes(msg tea.KeyMsg) []byte {
	// Special handling for detach key
	if msg.String() == "ctrl+\\" || msg.String() == "ctrl+backslash" {
		return nil // Signal to detach, don't send to terminal
	}

	// Map tea.KeyType to the corresponding terminal sequence
	switch msg.Type {
	case tea.KeyEnter: // Also handles tea.KeyCtrlM (same value)
		return []byte{'\r'} // Carriage return
	case tea.KeyTab: // Also handles tea.KeyCtrlI (same value)
		return []byte{'\t'}
	case tea.KeyBackspace:
		return []byte{0x7f} // DEL character (127)
	case tea.KeyEscape: // Also handles tea.KeyCtrlOpenBracket (same value)
		return []byte{0x1b} // ESC
	case tea.KeySpace:
		return []byte{' '}

	// Arrow keys
	case tea.KeyUp:
		return []byte{0x1b, '[', 'A'}
	case tea.KeyDown:
		return []byte{0x1b, '[', 'B'}
	case tea.KeyRight:
		return []byte{0x1b, '[', 'C'}
	case tea.KeyLeft:
		return []byte{0x1b, '[', 'D'}

	// Control sequences
	case tea.KeyCtrlA:
		return []byte{0x01}
	case tea.KeyCtrlB:
		return []byte{0x02}
	case tea.KeyCtrlC:
		return []byte{0x03}
	case tea.KeyCtrlD:
		return []byte{0x04}
	case tea.KeyCtrlE:
		return []byte{0x05}
	case tea.KeyCtrlF:
		return []byte{0x06}
	case tea.KeyCtrlG:
		return []byte{0x07}
	case tea.KeyCtrlH:
		return []byte{0x08} // Same as backspace
	case tea.KeyCtrlJ:
		return []byte{0x0a} // Line feed
	case tea.KeyCtrlK:
		return []byte{0x0b}
	case tea.KeyCtrlL:
		return []byte{0x0c}
	case tea.KeyCtrlN:
		return []byte{0x0e}
	case tea.KeyCtrlO:
		return []byte{0x0f}
	case tea.KeyCtrlP:
		return []byte{0x10}
	case tea.KeyCtrlQ:
		return []byte{0x11}
	case tea.KeyCtrlR:
		return []byte{0x12}
	case tea.KeyCtrlS:
		return []byte{0x13}
	case tea.KeyCtrlT:
		return []byte{0x14}
	case tea.KeyCtrlU:
		return []byte{0x15}
	case tea.KeyCtrlV:
		return []byte{0x16}
	case tea.KeyCtrlW:
		return []byte{0x17}
	case tea.KeyCtrlX:
		return []byte{0x18}
	case tea.KeyCtrlY:
		return []byte{0x19}
	case tea.KeyCtrlZ:
		return []byte{0x1a}
	case tea.KeyCtrlBackslash: // Ctrl+\
		return nil // Signal to detach
	case tea.KeyCtrlCloseBracket: // Ctrl+]
		return []byte{0x1d}
	case tea.KeyCtrlCaret: // Ctrl+^
		return []byte{0x1e}
	case tea.KeyCtrlUnderscore: // Ctrl+_
		return []byte{0x1f}

	// Page navigation
	case tea.KeyPgUp:
		return []byte{0x1b, '[', '5', '~'}
	case tea.KeyPgDown:
		return []byte{0x1b, '[', '6', '~'}
	case tea.KeyHome:
		return []byte{0x1b, '[', 'H'}
	case tea.KeyEnd:
		return []byte{0x1b, '[', 'F'}
	case tea.KeyInsert:
		return []byte{0x1b, '[', '2', '~'}
	case tea.KeyDelete:
		return []byte{0x1b, '[', '3', '~'}

	// Function keys
	case tea.KeyF1:
		return []byte{0x1b, 'O', 'P'}
	case tea.KeyF2:
		return []byte{0x1b, 'O', 'Q'}
	case tea.KeyF3:
		return []byte{0x1b, 'O', 'R'}
	case tea.KeyF4:
		return []byte{0x1b, 'O', 'S'}
	case tea.KeyF5:
		return []byte{0x1b, '[', '1', '5', '~'}
	case tea.KeyF6:
		return []byte{0x1b, '[', '1', '7', '~'}
	case tea.KeyF7:
		return []byte{0x1b, '[', '1', '8', '~'}
	case tea.KeyF8:
		return []byte{0x1b, '[', '1', '9', '~'}
	case tea.KeyF9:
		return []byte{0x1b, '[', '2', '0', '~'}
	case tea.KeyF10:
		return []byte{0x1b, '[', '2', '1', '~'}
	case tea.KeyF11:
		return []byte{0x1b, '[', '2', '3', '~'}
	case tea.KeyF12:
		return []byte{0x1b, '[', '2', '4', '~'}

	// Runes (regular characters)
	case tea.KeyRunes:
		// Check if Alt is pressed
		if msg.Alt {
			// Send ESC followed by the character for Alt combinations
			result := []byte{0x1b}
			result = append(result, []byte(string(msg.Runes))...)
			return result
		}
		// Regular character input
		return []byte(string(msg.Runes))

	default:
		// For any unhandled key type, return empty
		// This shouldn't happen in normal operation
		return []byte{}
	}
}

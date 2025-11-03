package app

import "github.com/charmbracelet/lipgloss"

// Color palette for the Docker Swarm Browser TUI
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00A7E1") // Docker blue
	ColorSecondary = lipgloss.Color("#00D4FF") // Bright cyan
	ColorAccent    = lipgloss.Color("#FF6B6B") // Coral red for warnings/alerts

	// Text colors
	ColorTextBody      = lipgloss.Color("#FFFFFF") // White
	ColorTextSecondary = lipgloss.Color("#A0A0A0") // Gray
	ColorTextMuted     = lipgloss.Color("#666666") // Dim gray
	ColorTextHighlight = lipgloss.Color("#00FFA3") // Bright green

	// UI element colors
	ColorTitle       = lipgloss.Color("#FFD700") // Gold
	ColorLabel       = lipgloss.Color("#87CEEB") // Sky blue
	ColorBorder      = lipgloss.Color("#4A5568") // Slate gray
	ColorBorderFocus = lipgloss.Color("#00D4FF") // Bright cyan (when focused)

	// Interactive elements
	ColorSelected   = lipgloss.Color("#00FFA3") // Bright green
	ColorSelectedBg = lipgloss.Color("#1A4D2E") // Dark green background
	ColorHover      = lipgloss.Color("#FFB84D") // Orange
	ColorDisabled   = lipgloss.Color("#4A4A4A") // Dark gray

	// Help and info
	ColorHelpKey  = lipgloss.Color("#7C3AED") // Purple
	ColorHelpText = lipgloss.Color("#9CA3AF") // Light gray
	ColorInfo     = lipgloss.Color("#60A5FA") // Blue
	ColorSuccess  = lipgloss.Color("#34D399") // Green
	ColorWarning  = lipgloss.Color("#FBBF24") // Amber
	ColorError    = lipgloss.Color("#EF4444") // Red

	// Status colors
	ColorStatusRunning = lipgloss.Color("#10B981") // Green
	ColorStatusStopped = lipgloss.Color("#6B7280") // Gray
	ColorStatusError   = lipgloss.Color("#DC2626") // Red
	ColorStatusPending = lipgloss.Color("#F59E0B") // Orange

	// Background colors (for subtle highlights)
	ColorBgSubtle = lipgloss.Color("#1F2937") // Dark blue-gray
	ColorBgPanel  = lipgloss.Color("#111827") // Darker blue-gray
)

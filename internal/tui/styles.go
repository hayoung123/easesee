package tui

import "github.com/charmbracelet/lipgloss"

// Cell content uses plain strings instead of lipgloss SetString styles because
// the bubbles/table component truncates cells without stripping the embedded
// ANSI escape sequences first — that mangles multibyte runes and the cell
// renders as the replacement character (替).
const (
	stateOn  = "● on"
	stateOff = "○ off"
	dirty    = "*"
)

var (
	// appHeader is the top-left app name.
	appHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#403e3c"))

	// sectionHeader is used for secondary headings (log pane, etc.).
	header = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6f6e69"))

	// help is for the key-binding bar and status messages.
	help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#b7b5ac"))

	// status is slightly more prominent than help (used for action feedback).
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6f6e69"))
)

package tui

import "github.com/charmbracelet/lipgloss"

// Cell content uses plain strings instead of lipgloss SetString styles because
// the bubbles/table component truncates cells without stripping the embedded
// ANSI escape sequences first — that mangles multibyte runes and the cell
// renders as the replacement character (�).
const (
	stateOn  = "● ON"
	stateOff = "○ OFF"
	dirty    = "★"
)

var (
	header = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3b82f6"))
	help   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
)

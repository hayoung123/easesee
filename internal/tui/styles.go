package tui

import "github.com/charmbracelet/lipgloss"

var (
	stateOn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#10b981")).SetString("● ON ")
	stateOff = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).SetString("○ OFF")
	dirty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b")).SetString("★")
	header   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3b82f6"))
	help     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
)

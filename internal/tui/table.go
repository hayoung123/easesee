package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/hayoung123/easesee/internal/discovery"
	"github.com/hayoung123/easesee/internal/git"
	"github.com/hayoung123/easesee/internal/registry"
)

func newTable() table.Model {
	cols := []table.Column{
		{Title: "NAME", Width: 18},
		{Title: "STATE", Width: 6},
		{Title: "PORT", Width: 6},
		{Title: "BRANCH", Width: 16},
		{Title: "CMD", Width: 40},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(false).
		Foreground(lipgloss.Color("#6f6e69")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#e6e4d9")).
		BorderBottom(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#100f0f")).
		Background(lipgloss.Color("#e6e4d9")).
		Bold(false)
	t.SetStyles(s)
	return t
}

// rowFor builds a table row for a project, given a match (empty MatchResult means OFF).
func rowFor(p registry.Project, m discovery.MatchResult) table.Row {
	state := stateOff
	port := "—"
	if m.PID != 0 {
		state = stateOn
		if m.Port > 0 {
			port = fmt.Sprintf("%d", m.Port)
		} else if m.StartedAt.IsZero() || time.Since(m.StartedAt) < 8*time.Second {
			port = "…" // starting up, port not yet bound
		} else {
			port = "—" // running but no port (e.g. CLI tools)
		}
	}
	branch := git.Branch(p.Cwd)
	if branch == "" {
		branch = "—"
	} else if git.IsDirty(p.Cwd) {
		branch += dirty
	}
	cmd := p.Cmd
	if len(cmd) > 40 {
		cmd = cmd[:37] + "…"
	}
	return table.Row{p.Name, state, port, branch, cmd}
}

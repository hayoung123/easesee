package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
	"github.com/proshy/devs/internal/discovery"
	"github.com/proshy/devs/internal/registry"
)

type model struct {
	paths   config.Paths
	keys    keymap
	reg     *registry.Registry
	tbl     table.Model
	matches map[string]discovery.MatchResult
	err     error
	status  string
}

func New() *model {
	m := &model{
		paths:   config.New(),
		keys:    defaultKeys(),
		tbl:     newTable(),
		matches: map[string]discovery.MatchResult{},
	}
	m.reloadRegistry()
	m.refresh()
	return m
}

func (m *model) reloadRegistry() {
	r, err := registry.Load(m.paths.RegistryFile)
	if err != nil {
		m.err = err
		return
	}
	m.reg = r
}

func (m *model) refresh() {
	if m.reg == nil {
		return
	}
	listeners, err := discovery.ListListeners()
	if err != nil {
		m.status = "discovery error: " + err.Error()
	} else {
		m.matches = discovery.Match(m.reg.Projects, listeners, discovery.GetProcInfo)
	}
	var rows []table.Row
	for _, p := range m.reg.Projects {
		rows = append(rows, rowFor(p, m.matches[p.Name]))
	}
	m.tbl.SetRows(rows)
}

func (m *model) Init() tea.Cmd { return tea.Batch(tickCmd()) }

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return tickMsg{} })
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.QuitAll):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			m.refresh()
		}
	case tickMsg:
		m.refresh()
		return m, tickCmd()
	}
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	var b strings.Builder
	b.WriteString(header.Render(" devs "))
	b.WriteString("\n\n")
	b.WriteString(m.tbl.View())
	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(help.Render(m.status))
		b.WriteString("\n")
	}
	b.WriteString(help.Render(" enter:toggle  s:start  x:stop  r:restart  l:log  a:add  e:edit  R:refresh  q:quit "))
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("ERROR: %v", m.err))
	}
	return b.String()
}

func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
	"github.com/proshy/devs/internal/discovery"
	"github.com/proshy/devs/internal/process"
	"github.com/proshy/devs/internal/registry"
	"github.com/proshy/devs/internal/state"
)

type model struct {
	paths   config.Paths
	keys    keymap
	reg     *registry.Registry
	tbl     table.Model
	matches map[string]discovery.MatchResult
	st      *state.State
	release func() error
	err     error
	status  string
	log     logPane
}

func New() *model {
	m := &model{
		paths:   config.New(),
		keys:    defaultKeys(),
		tbl:     newTable(),
		matches: map[string]discovery.MatchResult{},
	}
	if err := m.paths.EnsureDirs(); err != nil {
		m.err = err
		return m
	}
	rel, err := state.AcquireLock(m.paths.LockFile)
	if err != nil {
		m.err = err
		return m
	}
	m.release = rel
	s, _ := state.Load(m.paths.StateFile)
	m.st = s
	m.reloadRegistry()
	m.refresh()
	m.log = newLog()
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

func (m *model) selectedProject() *registry.Project {
	if m.reg == nil {
		return nil
	}
	i := m.tbl.Cursor()
	if i < 0 || i >= len(m.reg.Projects) {
		return nil
	}
	return &m.reg.Projects[i]
}

func (m *model) startSelected() {
	p := m.selectedProject()
	if p == nil {
		return
	}
	if _, ok := m.matches[p.Name]; ok {
		m.status = p.Name + " already running"
		return
	}
	logPath := filepath.Join(m.paths.LogsDir, p.Name+".log")
	pid, err := process.Start(*p, logPath)
	if err != nil {
		m.status = "start failed: " + err.Error()
		return
	}
	if m.st == nil {
		m.st = &state.State{Managed: map[string]state.Managed{}}
	}
	if m.st.Managed == nil {
		m.st.Managed = map[string]state.Managed{}
	}
	m.st.Managed[p.Name] = state.Managed{PID: pid, StartedAt: time.Now(), LogPath: logPath}
	_ = m.st.Save(m.paths.StateFile)
	m.status = fmt.Sprintf("started %s (pid=%d)", p.Name, pid)
	m.refresh()
}

func (m *model) stopSelected() {
	p := m.selectedProject()
	if p == nil {
		return
	}
	match, ok := m.matches[p.Name]
	if !ok {
		m.status = p.Name + " is not running"
		return
	}
	if err := process.Stop(match.PID, 3*time.Second); err != nil {
		m.status = "stop failed: " + err.Error()
		return
	}
	if m.st != nil {
		delete(m.st.Managed, p.Name)
		_ = m.st.Save(m.paths.StateFile)
	}
	m.status = "stopped " + p.Name
	m.refresh()
}

func (m *model) restartSelected() {
	m.stopSelected()
	time.Sleep(200 * time.Millisecond)
	m.startSelected()
}

func (m *model) toggleSelected() {
	p := m.selectedProject()
	if p == nil {
		return
	}
	if _, on := m.matches[p.Name]; on {
		m.stopSelected()
	} else {
		m.startSelected()
	}
}

func (m *model) cleanup(killAll bool) {
	if killAll && m.st != nil {
		for _, mg := range m.st.Managed {
			_ = process.Stop(mg.PID, 1*time.Second)
		}
	}
	if m.release != nil {
		_ = m.release()
		m.release = nil
	}
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
		case key.Matches(msg, m.keys.QuitAll):
			m.cleanup(true)
			return m, tea.Quit
		case key.Matches(msg, m.keys.Quit):
			m.cleanup(false)
			return m, tea.Quit
		case key.Matches(msg, m.keys.Toggle):
			m.toggleSelected()
			return m, nil
		case key.Matches(msg, m.keys.Start):
			m.startSelected()
			return m, nil
		case key.Matches(msg, m.keys.Stop):
			m.stopSelected()
			return m, nil
		case key.Matches(msg, m.keys.Restart):
			m.restartSelected()
			return m, nil
		case key.Matches(msg, m.keys.Log):
			m.log.visible = !m.log.visible
			if m.log.visible {
				if p := m.selectedProject(); p != nil {
					if m.st != nil {
						if mg, ok := m.st.Managed[p.Name]; ok {
							m.log.load(mg.LogPath)
						} else {
							m.log.load("")
						}
					} else {
						m.log.load("")
					}
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			m.refresh()
			return m, nil
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
	if m.log.visible {
		b.WriteString("\n")
		b.WriteString(header.Render(" log "))
		b.WriteString("\n")
		b.WriteString(m.log.vp.View())
		b.WriteString("\n")
	}
	if m.status != "" {
		b.WriteString(help.Render(m.status))
		b.WriteString("\n")
	}
	b.WriteString(help.Render(" enter:toggle  s:start  x:stop  r:restart  l:log  a:add  e:edit  R:refresh  q:quit  Q:quit+kill "))
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

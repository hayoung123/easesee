package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hayoung123/easesee/internal/config"
	"github.com/hayoung123/easesee/internal/discovery"
	"github.com/hayoung123/easesee/internal/process"
	"github.com/hayoung123/easesee/internal/registry"
	"github.com/hayoung123/easesee/internal/state"
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
	status   string
	log      logPane
	form     addForm
	killPort killPortForm
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
	m.form = newAddForm()
	m.killPort = newKillPortForm()
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
	// For dashboard-spawned servers, bridge two gaps the name-based matcher
	// can't close on its own:
	//   - Right after `s` the listener isn't bound yet, so lsof can't see it.
	//   - Even once bound, pnpm/yarn launch the actual listener (vite, etc.)
	//     as a child whose cmdline ("node …/vite") doesn't contain the
	//     project name, so cwd+cmd matching fails.
	// Both cases are covered by attributing any listener whose PGID equals
	// the PID we recorded in state.Managed — setsid gives the whole spawned
	// tree a shared PGID.
	if m.st != nil {
		for name, mg := range m.st.Managed {
			if _, ok := m.matches[name]; ok {
				continue
			}
			if !process.IsAlive(mg.PID) {
				delete(m.st.Managed, name)
				_ = m.st.Save(m.paths.StateFile)
				continue
			}
			var port int
			for _, l := range listeners {
				if discovery.GetPgid(l.PID) == mg.PID {
					if port == 0 || l.Port < port {
						port = l.Port
					}
				}
			}
			m.matches[name] = discovery.MatchResult{PID: mg.PID, Port: port, StartedAt: mg.StartedAt}
		}
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

func (m *model) lookupPort() {
	port, err := strconv.Atoi(strings.TrimSpace(m.killPort.input.Value()))
	if err != nil || port <= 0 {
		m.status = "invalid port: " + m.killPort.input.Value()
		m.killPort.reset()
		return
	}
	listeners, err := discovery.ListListeners()
	if err != nil {
		m.status = "discovery error: " + err.Error()
		m.killPort.reset()
		return
	}
	var targets []killTarget
	for _, l := range listeners {
		if l.Port == port {
			info := discovery.GetProcInfo(l.PID)
			targets = append(targets, killTarget{PID: l.PID, Cmd: info.Cmdline})
		}
	}
	if len(targets) == 0 {
		m.status = fmt.Sprintf("port %d is not in use", port)
		m.killPort.reset()
		return
	}
	m.killPort.targets = targets
	m.killPort.confirming = true
}

func (m *model) confirmKillPort() {
	for _, t := range m.killPort.targets {
		_ = process.Stop(t.PID, 3*time.Second)
	}
	m.status = fmt.Sprintf("killed port (pid=%d)", m.killPort.targets[0].PID)
	m.killPort.reset()
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
	// If killPort form is visible, route keys to it first.
	if m.killPort.visible {
		if k, ok := msg.(tea.KeyMsg); ok {
			if m.killPort.confirming {
				switch k.String() {
				case "y", "enter":
					m.confirmKillPort()
				case "n", "esc":
					m.killPort.reset()
				}
				return m, nil
			}
			switch k.String() {
			case "esc":
				m.killPort.reset()
				return m, nil
			case "enter":
				m.lookupPort()
				return m, nil
			}
		}
		cmd := m.killPort.Update(msg)
		return m, cmd
	}

	// If form is visible, route keys to form first.
	if m.form.visible {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "esc":
				m.form.reset()
				return m, nil
			case "tab":
				m.form.next()
				return m, nil
			case "enter":
				name, cwd, cmd := m.form.values()
				if err := registerInline(m.paths.RegistryFile, name, cwd, cmd); err != nil {
					m.status = "add failed: " + err.Error()
				} else {
					m.status = "added " + name
					m.form.reset()
					m.reloadRegistry()
					m.refresh()
				}
				return m, nil
			}
		}
		cmd := m.form.Update(msg)
		return m, cmd
	}
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
		case key.Matches(msg, m.keys.Add):
			m.form.visible = true
			return m, nil
		case key.Matches(msg, m.keys.KillPort):
			m.killPort.visible = true
			m.killPort.input.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Edit):
			return m, openEditor(m.paths.RegistryFile)
		}
	case tickMsg:
		m.refresh()
		return m, tickCmd()
	case reloadMsg:
		if msg.err != nil {
			m.status = "editor error: " + msg.err.Error()
		}
		m.reloadRegistry()
		m.refresh()
		return m, nil
	}
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	var b strings.Builder
	b.WriteString(header.Render(" easesee "))
	b.WriteString("\n\n")
	if m.form.visible {
		b.WriteString(m.form.View())
		b.WriteString("\n\n")
	}
	if m.killPort.visible {
		b.WriteString(m.killPort.View())
		b.WriteString("\n\n")
	}
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
	b.WriteString(help.Render(" enter:toggle  s:start  x:stop  r:restart  l:log  a:add  e:edit  K:kill-port  R:refresh  q:quit  Q:quit+kill "))
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

func registerInline(regPath, name, cwd, cmd string) error {
	r, err := registry.Load(regPath)
	if err != nil {
		return err
	}
	if err := r.Add(registry.Project{Name: name, Cwd: cwd, Cmd: cmd}); err != nil {
		return err
	}
	return r.Save(regPath)
}

type reloadMsg struct{ err error }

func openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return reloadMsg{err: err}
	})
}

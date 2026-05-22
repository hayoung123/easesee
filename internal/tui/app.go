package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
)

type model struct {
	paths  config.Paths
	keys   keymap
	status string
	quit   bool
}

func New() *model {
	return &model{
		paths: config.New(),
		keys:  defaultKeys(),
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) || key.Matches(msg, m.keys.QuitAll) {
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s",
		header.Render("devs (skeleton)"),
		"press q to quit",
		help.Render("q: quit  Q: quit+kill all"),
	)
}

// Run starts the program.
func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

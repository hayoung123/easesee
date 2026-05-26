package tui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	Start    key.Binding
	Stop     key.Binding
	Restart  key.Binding
	Log      key.Binding
	Add      key.Binding
	Edit     key.Binding
	Refresh  key.Binding
	KillPort key.Binding
	Quit     key.Binding
	QuitAll  key.Binding
}

func defaultKeys() keymap {
	return keymap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Toggle:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle")),
		Start:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
		Stop:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "stop")),
		Restart:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restart")),
		Log:      key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "log")),
		Add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit registry")),
		Refresh:  key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "refresh")),
		KillPort: key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "kill port")),
		Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		QuitAll:  key.NewBinding(key.WithKeys("Q"), key.WithHelp("Q", "quit+kill all")),
	}
}

package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type addForm struct {
	visible bool
	inputs  []textinput.Model
	idx     int
}

func newAddForm() addForm {
	mk := func(placeholder string) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Width = 40
		return ti
	}
	inputs := []textinput.Model{mk("name"), mk("cwd"), mk("cmd")}
	inputs[0].Focus()
	return addForm{inputs: inputs}
}

func (f *addForm) next() {
	f.inputs[f.idx].Blur()
	f.idx = (f.idx + 1) % len(f.inputs)
	f.inputs[f.idx].Focus()
}

func (f *addForm) reset() {
	for i := range f.inputs {
		f.inputs[i].SetValue("")
		f.inputs[i].Blur()
	}
	f.idx = 0
	f.inputs[0].Focus()
	f.visible = false
}

func (f *addForm) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.inputs[f.idx], cmd = f.inputs[f.idx].Update(msg)
	return cmd
}

func (f *addForm) values() (name, cwd, cmd string) {
	return f.inputs[0].Value(), f.inputs[1].Value(), f.inputs[2].Value()
}

func (f addForm) View() string {
	return "─ Add project ─\n" +
		" name: " + f.inputs[0].View() + "\n" +
		" cwd:  " + f.inputs[1].View() + "\n" +
		" cmd:  " + f.inputs[2].View() + "\n" +
		" tab:next  enter:save  esc:cancel"
}

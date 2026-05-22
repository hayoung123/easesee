package tui

import (
	"bufio"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
)

type logPane struct {
	vp      viewport.Model
	path    string
	visible bool
}

func newLog() logPane {
	vp := viewport.New(80, 10)
	return logPane{vp: vp}
}

// load tails up to last N lines from path.
func (l *logPane) load(path string) {
	l.path = path
	if path == "" {
		l.vp.SetContent("(no log)")
		return
	}
	f, err := os.Open(path)
	if err != nil {
		l.vp.SetContent("(log not available: " + err.Error() + ")")
		return
	}
	defer f.Close()
	const max = 200
	lines := make([]string, 0, max)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) > max {
			lines = lines[1:]
		}
	}
	var content string
	for _, ln := range lines {
		content += ln + "\n"
	}
	l.vp.SetContent(content)
	l.vp.GotoBottom()
}

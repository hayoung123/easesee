package discovery

import (
	"os/exec"
	"strconv"
	"strings"
)

type Listener struct {
	PID  int
	Cmd  string
	Port int
}

// ListListeners shells out to lsof and returns TCP LISTEN entries (deduped per PID+port).
func ListListeners() ([]Listener, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn").Output()
	if err != nil {
		return nil, err
	}
	return parseLsof(string(out)), nil
}

func parseLsof(s string) []Listener {
	var pid int
	var cmd string
	seen := map[string]bool{}
	var out []Listener
	for _, line := range strings.Split(s, "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
			cmd = ""
		case 'c':
			cmd = line[1:]
		case 'n':
			n := line[1:]
			if i := strings.LastIndex(n, ":"); i >= 0 {
				n = n[i+1:]
			}
			port, err := strconv.Atoi(n)
			if err != nil {
				continue
			}
			key := strconv.Itoa(pid) + ":" + strconv.Itoa(port)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, Listener{PID: pid, Cmd: cmd, Port: port})
		}
	}
	return out
}

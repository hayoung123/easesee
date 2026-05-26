package discovery

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type ProcInfo struct {
	PID     int
	Cwd     string
	Cmdline string
}

// GetCwd returns the working directory of a process via lsof.
func GetCwd(pid int) string {
	out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
	if err != nil {
		return ""
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		if len(line) > 1 && line[0] == 'n' {
			return line[1:]
		}
	}
	return ""
}

// GetCmdline returns the full command line via ps.
func GetCmdline(pid int) string {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetProcInfo(pid int) ProcInfo {
	return ProcInfo{PID: pid, Cwd: GetCwd(pid), Cmdline: GetCmdline(pid)}
}

// GetPgid returns the process group ID of pid via syscall. Returns 0 on failure.
// Used to attribute pnpm/vite child listeners back to their setsid'd parent.
func GetPgid(pid int) int {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return 0
	}
	return pgid
}

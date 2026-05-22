package git

import (
	"os/exec"
	"strings"
)

// Branch returns the current branch in cwd, or "" if not a git repo.
func Branch(cwd string) string {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// IsDirty returns true if the working tree has uncommitted changes.
func IsDirty(cwd string) bool {
	out, err := exec.Command("git", "-C", cwd, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

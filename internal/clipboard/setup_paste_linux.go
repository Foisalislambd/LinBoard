//go:build linux

package clipboard

import (
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// UinputWritable reports whether /dev/uinput can be opened for writing.
func UinputWritable() bool {
	for _, path := range []string{"/dev/uinput", "/dev/input/uinput"} {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err == nil {
			_ = f.Close()
			return true
		}
	}
	return false
}

// InInputGroupSession reports whether the current process has the input group.
func InInputGroupSession() bool {
	out, err := exec.Command("id", "-nG").Output()
	if err != nil {
		return false
	}
	for _, g := range strings.Fields(string(out)) {
		if g == "input" {
			return true
		}
	}
	return false
}

// InInputGroupConfigured reports whether the user account is in the input group.
func InInputGroupConfigured() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	out, err := exec.Command("getent", "group", "input").Output()
	if err != nil {
		return false
	}
	// input:x:995:foisal,other
	parts := strings.Split(strings.TrimSpace(string(out)), ":")
	if len(parts) < 4 {
		return false
	}
	for _, member := range strings.Split(parts[3], ",") {
		if strings.TrimSpace(member) == u.Username {
			return true
		}
	}
	return false
}

// SessionNeedsRelogin is true when input group was added but this session is stale.
func SessionNeedsRelogin() bool {
	return InInputGroupConfigured() && !InInputGroupSession()
}

// PasteSessionHint returns how to activate paste in the current situation.
func PasteSessionHint() string {
	if UinputWritable() || uinputReady() {
		return ""
	}
	if SessionNeedsRelogin() {
		return `input group is set but this session is old — log out and back in, or run:
  sg input -c "linboard"
  # autostart uses ~/.local/bin/linboard-start after install`
	}
	return PasteSetupHint()
}

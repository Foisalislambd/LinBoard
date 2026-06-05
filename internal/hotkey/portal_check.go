package hotkey

import (
	"os/exec"
	"strings"
)

func portalHasGlobalShortcuts() bool {
	if !hasBin("gdbus") {
		return false
	}
	out, err := exec.Command("gdbus", "introspect", "--session",
		"--dest", portalBusName,
		"--object-path", portalObjectPath,
	).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), portalIface)
}

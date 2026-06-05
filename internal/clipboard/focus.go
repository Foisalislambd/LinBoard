package clipboard

import (
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/foisal/linboard/internal/platform"
)

var (
	pasteTargetMu sync.Mutex
	pasteTargetID string // X11 window id (xdotool)
)

// RememberPasteTarget saves the window that had focus before LinBoard history opens.
// Call at the start of Show(), before the popup takes keyboard focus.
func RememberPasteTarget() {
	pasteTargetMu.Lock()
	defer pasteTargetMu.Unlock()
	pasteTargetID = ""

	if _, err := exec.LookPath("xdotool"); err != nil {
		return
	}

	out, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return
	}
	id := strings.TrimSpace(string(out))
	if id == "" || id == "0" {
		return
	}
	pasteTargetID = id
	log.Printf("paste target window: %s", id)
}

func restorePasteTarget() bool {
	pasteTargetMu.Lock()
	id := pasteTargetID
	pasteTargetMu.Unlock()

	if id == "" {
		return false
	}
	if _, err := exec.LookPath("xdotool"); err != nil {
		return false
	}
	err := exec.Command("xdotool", "windowactivate", "--sync", id).Run()
	if err != nil {
		log.Printf("restore focus (xdotool): %v", err)
		return false
	}
	return true
}

func pasteDelays() []int {
	if platform.UsePortalHotkey() {
		// Wayland compositor needs more time to return focus after popup closes.
		return []int{120, 250, 400, 600}
	}
	return []int{80, 160, 280}
}

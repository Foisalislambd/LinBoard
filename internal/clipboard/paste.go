package clipboard

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/foisal/linboard/internal/platform"
)

// simulatePaste sends Ctrl+V to the focused window using the best tool available.
func simulatePaste() error {
	tools := pasteTools()
	for _, t := range tools {
		if _, err := exec.LookPath(t.bin); err != nil {
			continue
		}
		cmd := exec.Command(t.bin, t.args...)
		if err := cmd.Run(); err != nil {
			log.Printf("auto-paste (%s) failed: %v", t.bin, err)
			continue
		}
		log.Printf("auto-paste: sent via %s", t.bin)
		return nil
	}
	log.Printf("auto-paste: no tool found (install wtype, ydotool, or xdotool)")
	return fmt.Errorf("no paste tool found (install wtype, ydotool, or xdotool)")
}

type pasteTool struct {
	bin  string
	args []string
}

func pasteTools() []pasteTool {
	if platform.UsePortalHotkey() {
		return []pasteTool{
			{bin: "wtype", args: []string{"-M", "ctrl", "-k", "v"}},
			{bin: "ydotool", args: []string{"key", "29:1", "47:1", "47:0", "29:0"}},
			{bin: "xdotool", args: []string{"key", "--clearmodifiers", "ctrl+v"}},
		}
	}
	return []pasteTool{
		{bin: "xdotool", args: []string{"key", "--clearmodifiers", "ctrl+v"}},
		{bin: "wtype", args: []string{"-M", "ctrl", "-k", "v"}},
		{bin: "ydotool", args: []string{"key", "29:1", "47:1", "47:0", "29:0"}},
	}
}

// PasteToolName returns the first available paste tool for diagnostics.
func PasteToolName() string {
	for _, t := range pasteTools() {
		if _, err := exec.LookPath(t.bin); err == nil {
			return t.bin
		}
	}
	return "none"
}

// simulatePasteWithRetry restores the previous window (X11) and retries Ctrl+V.
func simulatePasteWithRetry() error {
	var lastErr error
	for i, ms := range pasteDelays() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		restorePasteTarget()
		if err := simulatePaste(); err == nil {
			return nil
		} else {
			lastErr = err
			if i < len(pasteDelays())-1 {
				log.Printf("auto-paste retry %d/%d", i+2, len(pasteDelays()))
			}
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("auto-paste failed")
}

package clipboard

import (
	"fmt"
	"log"
	"os/exec"

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
		return nil
	}
	log.Printf("auto-paste: no tool found (install wtype, ydotool, or xdotool) — content is on clipboard")
	return fmt.Errorf("no paste tool found (install wtype, ydotool, or xdotool)")
}

type pasteTool struct {
	bin  string
	args []string
}

func pasteTools() []pasteTool {
	if platform.UsePortalHotkey() {
		// Wayland: wtype and ydotool work; xdotool usually does not.
		return []pasteTool{
			{bin: "wtype", args: []string{"-M", "ctrl", "-k", "v"}},
			{bin: "ydotool", args: []string{"key", "29:1", "47:1", "47:0", "29:0"}}, // ctrl+v
			{bin: "xdotool", args: []string{"key", "ctrl+v"}},
		}
	}
	return []pasteTool{
		{bin: "xdotool", args: []string{"key", "ctrl+v"}},
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

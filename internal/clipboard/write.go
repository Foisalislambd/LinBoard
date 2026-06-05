package clipboard

import (
	"fmt"
	"sync"

	"github.com/foisal/linboard/internal/platform"
)

var (
	fyneMu      sync.RWMutex
	fyneSetText func(string)
)

// SetFyneWriter registers a callback that writes text via the Fyne app clipboard.
// CopyQ/Qt use the toolkit clipboard on Wayland; this hook does the same for LinBoard.
func SetFyneWriter(fn func(string)) {
	fyneMu.Lock()
	fyneSetText = fn
	fyneMu.Unlock()
}

// WriteText puts UTF-8 text on the system clipboard using the best available backend.
func WriteText(text string) error {
	for _, w := range textWriters() {
		if err := w(text); err == nil {
			return nil
		}
	}
	return fmt.Errorf("clipboard write failed (install wl-clipboard on Wayland or xclip on X11)")
}

// WriteImage puts PNG bytes on the system clipboard.
func WriteImage(data []byte) error {
	for _, w := range imageWriters() {
		if err := w(data); err == nil {
			return nil
		}
	}
	return fmt.Errorf("clipboard image write failed (install wl-clipboard on Wayland)")
}

// CopyToolName returns the first available clipboard write backend for diagnostics.
func CopyToolName() string {
	for _, name := range []string{"wl-copy", "fyne", "xclip", "xsel"} {
		if writerAvailable(name) {
			return name
		}
	}
	return "none"
}

func textWriters() []func(string) error {
	if platform.UsePortalHotkey() {
		return []func(string) error{
			writeTextWLCopy,
			writeTextFyne,
			writeTextXClip,
			writeTextXSel,
		}
	}
	return []func(string) error{
		writeTextFyne,
		writeTextWLCopy,
		writeTextXClip,
		writeTextXSel,
	}
}

func imageWriters() []func([]byte) error {
	if platform.UsePortalHotkey() {
		return []func([]byte) error{
			writeImageWLCopy,
			writeImageXClip,
		}
	}
	return []func([]byte) error{
		writeImageWLCopy,
		writeImageXClip,
	}
}

func writerAvailable(name string) bool {
	switch name {
	case "wl-copy", "xclip", "xsel":
		return haveCommand(name)
	case "fyne":
		fyneMu.RLock()
		ok := fyneSetText != nil
		fyneMu.RUnlock()
		return ok
	default:
		return false
	}
}

func writeTextWLCopy(text string) error {
	return pipeToCommand("wl-copy", nil, text)
}

func writeTextFyne(text string) error {
	fyneMu.RLock()
	fn := fyneSetText
	fyneMu.RUnlock()
	if fn == nil {
		return fmt.Errorf("fyne clipboard not configured")
	}
	fn(text)
	return nil
}

func writeTextXClip(text string) error {
	return pipeToCommand("xclip", []string{"-selection", "clipboard"}, text)
}

func writeTextXSel(text string) error {
	return pipeToCommand("xsel", []string{"--clipboard", "--input"}, text)
}

func writeImageWLCopy(data []byte) error {
	return pipeToCommand("wl-copy", []string{"-t", "image/png"}, data)
}

func writeImageXClip(data []byte) error {
	return pipeToCommand("xclip", []string{"-selection", "clipboard", "-t", "image/png"}, data)
}

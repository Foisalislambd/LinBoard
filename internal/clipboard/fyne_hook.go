package clipboard

import "sync"

// Fyne hooks use the running app's GLFW clipboard (Wayland + X11). No wl-copy/wl-paste needed.
var (
	fyneMu      sync.RWMutex
	fyneGetText func() string
	fyneSetText func(string)
)

// SetFyneReader registers a callback that reads text from the Fyne app clipboard.
func SetFyneReader(fn func() string) {
	fyneMu.Lock()
	fyneGetText = fn
	fyneMu.Unlock()
}

// SetFyneWriter registers a callback that writes text via the Fyne app clipboard.
func SetFyneWriter(fn func(string)) {
	fyneMu.Lock()
	fyneSetText = fn
	fyneMu.Unlock()
}

func fyneAvailable() bool {
	fyneMu.RLock()
	ok := fyneGetText != nil && fyneSetText != nil
	fyneMu.RUnlock()
	return ok
}

func readTextFyne() (string, error) {
	fyneMu.RLock()
	fn := fyneGetText
	fyneMu.RUnlock()
	if fn == nil {
		return "", errFyneNotReady
	}
	return fn(), nil
}

func writeTextFyne(text string) error {
	fyneMu.RLock()
	fn := fyneSetText
	fyneMu.RUnlock()
	if fn == nil {
		return errFyneNotReady
	}
	fn(text)
	return nil
}

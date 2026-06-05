package hotkey

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/foisal/linboard/internal/config"
)

// RegisterSystemShortcut binds Super+V to `linboard toggle` using the desktop environment API.
func RegisterSystemShortcut() error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	return registerSystemShortcut(exe)
}

func executablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func registerSystemShortcut(exe string) error {
	backends := []struct {
		name string
		fn   func(string) error
	}{
		{"GNOME", setupGNOMEHotkey},
		{"KDE", setupKDEHotkey},
		{"XFCE", setupXFCEHotkey},
		{"Cinnamon", setupCinnamonHotkey},
	}
	var errs []string
	for _, b := range backends {
		if err := b.fn(exe); err == nil {
			return nil
		} else if !isSkipErr(err) {
			errs = append(errs, fmt.Sprintf("%s: %v", b.name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("shortcut registration failed: %s", strings.Join(errs, "; "))
	}
	return fmt.Errorf("no supported desktop environment for automatic %s binding", config.HotkeyLabel)
}

func isSkipErr(err error) bool {
	return strings.Contains(err.Error(), "skip:")
}

func skip(format string, args ...any) error {
	return fmt.Errorf("skip: "+format, args...)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v: %w (%s)", name, args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func hasBin(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

package hotkey

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foisal/linboard/internal/config"
)

const toggleWrapperName = "linboard-toggle"

// ToggleWrapperPath is ~/.config/linboard/linboard-toggle
func ToggleWrapperPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, toggleWrapperName), nil
}

// EnsureToggleWrapper writes a launcher script GNOME can run (no shell args in gsettings).
func EnsureToggleWrapper(exe string) (string, error) {
	path, err := ToggleWrapperPath()
	if err != nil {
		return "", err
	}
	script := "#!/bin/sh\nexec " + fmt.Sprintf("%q", exe) + " toggle\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		return "", err
	}
	return path, nil
}

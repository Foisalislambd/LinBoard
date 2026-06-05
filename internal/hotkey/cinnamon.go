package hotkey

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

type cinnamonBackend struct{}

func (b *cinnamonBackend) start(_ func()) error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	if err := setupCinnamonHotkey(exe); err != nil {
		return err
	}
	log.Printf("hotkey registered (Cinnamon): %s → linboard toggle", config.HotkeyLabel)
	return nil
}

func (b *cinnamonBackend) stop() {}

func setupCinnamonHotkey(exe string) error {
	if !platform.IsCinnamon() {
		return skip("not Cinnamon")
	}
	if !hasBin("gsettings") {
		return fmt.Errorf("gsettings not found")
	}

	listSchema := "org.cinnamon.desktop.keybindings"
	path := "/org/cinnamon/desktop/keybindings/custom-keybindings/custom-linboard/"
	fullSchema := listSchema + ".custom-keybinding:" + path

	// Cinnamon uses custom-keybindings (same idea as GNOME).
	paths, err := gsettingsGetArray(listSchema, "custom-keybindings")
	if err != nil {
		paths, err = gsettingsGetArray(listSchema, "custom-list")
		if err != nil {
			return err
		}
	}
	if !containsPath(paths, path) {
		paths = append(paths, path)
		if err := gsettingsSetArray(listSchema, "custom-keybindings", paths); err != nil {
			if err2 := gsettingsSetArray(listSchema, "custom-list", paths); err2 != nil {
				return err
			}
		}
	}

	if err := gsettingsSet(fullSchema, "name", "LinBoard"); err != nil {
		return err
	}
	if err := gsettingsSet(fullSchema, "command", exe+" toggle"); err != nil {
		return err
	}
	return gsettingsSet(fullSchema, "binding", "<Super>v")
}

func gsettingsGetArray(schema, key string) ([]string, error) {
	out, err := execOutput("gsettings", "get", schema, key)
	if err != nil {
		return nil, err
	}
	return parseGSettingsStringArray(out), nil
}

func gsettingsSetArray(schema, key string, paths []string) error {
	if len(paths) == 0 {
		return run("gsettings", "set", schema, key, "[]")
	}
	var quoted []string
	for _, p := range paths {
		quoted = append(quoted, "'"+p+"'")
	}
	val := "[" + stringsJoin(quoted, ", ") + "]"
	return run("gsettings", "set", schema, key, val)
}

func stringsJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	s := parts[0]
	for _, p := range parts[1:] {
		s += sep + p
	}
	return s
}

func execOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return string(out), nil
}

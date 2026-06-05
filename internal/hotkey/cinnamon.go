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
	customKey := "custom0"
	schema := "org.cinnamon.desktop.keybindings.wm"
	nameKey := "name"
	commandKey := "command"
	bindingKey := "binding"

	// Cinnamon custom keybindings use a different schema; use custom-list approach.
	listSchema := "org.cinnamon.desktop.keybindings"
	path := "/org/cinnamon/desktop/keybindings/custom-keybindings/custom-linboard/"
	fullSchema := listSchema + ".custom-keybinding:" + path

	paths, err := gsettingsListPathsCinnamon()
	if err != nil {
		return err
	}
	if !containsPath(paths, path) {
		paths = append(paths, path)
		if err := gsettingsSetCinnamon(paths); err != nil {
			return err
		}
	}

	_ = nameKey
	_ = commandKey
	_ = bindingKey
	_ = schema
	_ = customKey

	if err := gsettingsSet(fullSchema, "name", "LinBoard"); err != nil {
		return err
	}
	if err := gsettingsSet(fullSchema, "command", exe+" toggle"); err != nil {
		return err
	}
	return gsettingsSet(fullSchema, "binding", "<Super>v")
}

func gsettingsListPathsCinnamon() ([]string, error) {
	out, err := execOutput("gsettings", "get", "org.cinnamon.desktop.keybindings", "custom-list")
	if err != nil {
		return nil, err
	}
	return parseGSettingsStringArray(out), nil
}

func gsettingsSetCinnamon(paths []string) error {
	if len(paths) == 0 {
		return run("gsettings", "set", "org.cinnamon.desktop.keybindings", "custom-list", "[]")
	}
	var quoted []string
	for _, p := range paths {
		quoted = append(quoted, "'"+p+"'")
	}
	val := "[" + joinQuoted(quoted) + "]"
	return run("gsettings", "set", "org.cinnamon.desktop.keybindings", "custom-list", val)
}

func joinQuoted(parts []string) string {
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += ", "
		}
		s += p
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

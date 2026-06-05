package hotkey

import (
	"log"
	"strings"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

const gnomeBindingName = "custom-linboard"

// gnomeBackend registers Super+V via GNOME Settings (gsettings).
// GNOME runs `linboard toggle`; IPC opens the history window.
type gnomeBackend struct{}

func (b *gnomeBackend) start(_ func()) error {
	if gnomeHotkeyConfigured() {
		log.Printf("hotkey using GNOME shortcut: %s → linboard toggle", config.HotkeyLabel)
		return nil
	}
	exe, err := executablePath()
	if err != nil {
		return err
	}
	if err := setupGNOMEHotkey(exe); err != nil {
		return err
	}
	log.Printf("hotkey registered (GNOME): %s → linboard toggle", config.HotkeyLabel)
	return nil
}

func gnomeBindingPath() string {
	return "/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/" + gnomeBindingName + "/"
}

func gnomeBindingSchema() string {
	return gnomeMediaKeys + ".custom-keybinding:" + gnomeBindingPath()
}

func gnomeHotkeyConfigured() bool {
	if !platform.IsGNOME() {
		return false
	}
	cur, err := gsettingsCommand(gnomeBindingSchema(), "command")
	if err != nil {
		return false
	}
	return strings.Contains(cur, "linboard") && strings.Contains(cur, "toggle")
}

func (b *gnomeBackend) stop() {}

func setupGNOMEHotkey(exe string) error {
	if !platform.IsGNOME() {
		return skip("not GNOME")
	}

	schema := gnomeBindingSchema()
	bindingPath := gnomeBindingPath()

	if cur, err := gsettingsCommand(schema, "command"); err == nil {
		if strings.Contains(cur, exe) && strings.Contains(cur, "toggle") {
			return nil
		}
	}

	paths, err := gsettingsListPaths()
	if err != nil {
		return err
	}
	if !containsPath(paths, bindingPath) {
		paths = append(paths, bindingPath)
		if err := gsettingsSetPaths(paths); err != nil {
			return err
		}
	}

	if err := gsettingsSet(schema, "name", "LinBoard"); err != nil {
		return err
	}
	if err := gsettingsSet(schema, "command", exe+" toggle"); err != nil {
		return err
	}
	return gsettingsSet(schema, "binding", "<Super>v")
}

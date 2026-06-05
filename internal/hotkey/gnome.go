package hotkey

import (
	"log"
	"strings"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

const (
	gnomeShellKeybindings = "org.gnome.shell.keybindings"
	gnomeMessageTrayKey   = "toggle-message-tray"
)

const gnomeBindingName = "custom-linboard"

// gnomeBackend registers Super+V via GNOME Settings (gsettings).
// GNOME runs `linboard toggle`; IPC opens the history window.
type gnomeBackend struct{}

func (b *gnomeBackend) start(_ func()) error {
	if err := releaseGNOMESuperVConflict(); err != nil {
		log.Printf("hotkey: GNOME Super+V conflict: %v", err)
	}
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
	if err := gsettingsSet(schema, "binding", "<Super>v"); err != nil {
		return err
	}
	return releaseGNOMESuperVConflict()
}

// releaseGNOMESuperVConflict removes Super+V from GNOME Shell's message tray binding.
// Shell keybindings take priority over custom media-keys shortcuts.
func releaseGNOMESuperVConflict() error {
	bindings, err := gsettingsGetArray(gnomeShellKeybindings, gnomeMessageTrayKey)
	if err != nil {
		return nil
	}
	var kept []string
	removed := false
	for _, b := range bindings {
		if strings.EqualFold(b, "<Super>v") {
			removed = true
			continue
		}
		kept = append(kept, b)
	}
	if !removed {
		return nil
	}
	if err := gsettingsSetArray(gnomeShellKeybindings, gnomeMessageTrayKey, kept); err != nil {
		return err
	}
	log.Printf("hotkey: released %s from GNOME message tray (now free for LinBoard)", config.HotkeyLabel)
	return nil
}

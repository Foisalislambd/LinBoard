package hotkey

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

const (
	gnomeMediaKeys   = "org.gnome.settings-daemon.plugins.media-keys"
	gnomeBindingName = "custom-linboard"
)

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

func gnomeHotkeyConfigured() bool {
	if !platform.IsGNOME() {
		return false
	}
	bindingPath := "/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/" + gnomeBindingName + "/"
	schema := gnomeMediaKeys + ".custom-keybinding:" + bindingPath
	cur, err := gsettingsGet(schema, "command")
	if err != nil {
		return false
	}
	cur = strings.Trim(cur, "'")
	return strings.Contains(cur, "linboard") && strings.Contains(cur, "toggle")
}

func (b *gnomeBackend) stop() {}

func setupGNOMEHotkey(exe string) error {
	if !platform.IsGNOME() {
		return skip("not GNOME")
	}

	bindingPath := "/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/" + gnomeBindingName + "/"
	schema := gnomeMediaKeys + ".custom-keybinding:" + bindingPath

	// Already configured for this binary — do not touch gsettings (re-setting binding can re-fire it on GNOME).
	if cur, err := gsettingsGet(schema, "command"); err == nil {
		cur = strings.Trim(cur, "'")
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
	return nil
}

func gsettingsGet(schema, key string) (string, error) {
	out, err := exec.Command("gsettings", "get", schema, key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gsettingsSet(schema, key, value string) error {
	cmd := exec.Command("gsettings", "set", schema, key, value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gsettings set %s %s: %w (%s)", schema, key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func gsettingsListPaths() ([]string, error) {
	out, err := exec.Command("gsettings", "get", gnomeMediaKeys, "custom-keybindings").Output()
	if err != nil {
		return nil, err
	}
	return parseGSettingsStringArray(strings.TrimSpace(string(out))), nil
}

func gsettingsSetPaths(paths []string) error {
	if len(paths) == 0 {
		return gsettingsSet(gnomeMediaKeys, "custom-keybindings", "[]")
	}
	var quoted []string
	for _, p := range paths {
		quoted = append(quoted, "'"+p+"'")
	}
	val := "[" + strings.Join(quoted, ", ") + "]"
	return gsettingsSet(gnomeMediaKeys, "custom-keybindings", val)
}

func parseGSettingsStringArray(s string) []string {
	if s == "@as []" || s == "[]" {
		return nil
	}
	s = strings.TrimPrefix(s, "@as ")
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}
	var paths []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "'")
		if part != "" {
			paths = append(paths, part)
		}
	}
	return paths
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

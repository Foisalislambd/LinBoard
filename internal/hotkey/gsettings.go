package hotkey

import (
	"fmt"
	"os/exec"
	"strings"
)

const gnomeMediaKeys = "org.gnome.settings-daemon.plugins.media-keys"

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
	return gsettingsSet(gnomeMediaKeys, "custom-keybindings", "["+strings.Join(quoted, ", ")+"]")
}

func gsettingsGetArray(schema, key string) ([]string, error) {
	out, err := gsettingsGet(schema, key)
	if err != nil {
		return nil, err
	}
	return parseGSettingsStringArray(out), nil
}

func gsettingsSetArray(schema, key string, paths []string) error {
	if len(paths) == 0 {
		return gsettingsSet(schema, key, "[]")
	}
	var quoted []string
	for _, p := range paths {
		quoted = append(quoted, "'"+p+"'")
	}
	return gsettingsSet(schema, key, "["+strings.Join(quoted, ", ")+"]")
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

func gsettingsCommand(schema, key string) (string, error) {
	cur, err := gsettingsGet(schema, key)
	if err != nil {
		return "", err
	}
	return strings.Trim(cur, "'"), nil
}

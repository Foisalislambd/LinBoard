package hotkey

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

type kdeBackend struct{}

func (b *kdeBackend) start(_ func()) error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	if err := setupKDEHotkey(exe); err != nil {
		return err
	}
	log.Printf("hotkey registered (KDE): %s → linboard toggle", config.HotkeyLabel)
	return nil
}

func (b *kdeBackend) stop() {}

func setupKDEHotkey(exe string) error {
	if !platform.IsKDE() {
		return skip("not KDE")
	}

	// KDE Plasma: custom command shortcut via khotkeys (khotkeysrc)
	kwrite := "kwriteconfig6"
	if !hasBin(kwrite) {
		kwrite = "kwriteconfig5"
	}
	if !hasBin(kwrite) {
		return fmt.Errorf("kwriteconfig not found")
	}

	uuid := "{linboard-toggle-0001}"
	dataGroup := "Data_1 20 LinBoard"
	path := filepath.Join(os.Getenv("HOME"), ".config", "khotkeysrc")

	if strings.Contains(readFile(path), "linboard-toggle") {
		_ = kwriteSet(kwrite, "khotkeysrc", dataGroup, "Command", exe+" toggle")
		_ = kwriteSet(kwrite, "khotkeysrc", dataGroup, "Name", "LinBoard")
		reloadKHotkeys()
		return nil
	}

	lines := []string{
		"",
		"[" + dataGroup + "]",
		"Comment=LinBoard clipboard history",
		"Enabled=true",
		"Name=LinBoard",
		"Type=SIMPLE_COMMAND_DATA",
		"Command=" + exe + " toggle",
		"Uuid=" + uuid,
		"",
		"[" + dataGroup + "|0 Trigger]",
		"Comment=LinBoard trigger",
		"Enabled=true",
		"Type=SHORTCUT_TRIGGER_DATA",
		"Uuid=" + uuid + "-trigger",
		"Key=Meta+V",
		"",
		"[" + dataGroup + "|0 Action/0]",
		"Type=COMMAND_URLS",
		"Uuid=" + uuid + "-action",
		"Command=" + exe + " toggle",
	}
	if err := appendFile(path, strings.Join(lines, "\n")+"\n"); err != nil {
		return err
	}
	reloadKHotkeys()
	return nil
}

func reloadKHotkeys() {
	for _, dbus := range []string{"qdbus6", "qdbus"} {
		if !hasBin(dbus) {
			continue
		}
		_ = run(dbus, "org.kde.kded6", "/kded", "org.kde.kded6.reloadModule", "khotkeys")
		_ = run(dbus, "org.kde.kded5", "/kded", "org.kde.kded5.reloadModule", "khotkeys")
	}
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func appendFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func kwriteSet(kwrite, file, group, key, value string) error {
	return run(kwrite, "--file", file, "--group", group, "--key", key, value)
}

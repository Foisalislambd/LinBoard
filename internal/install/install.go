package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/foisal/linboard/internal/hotkey"
)

// Run installs LinBoard binary, desktop file, autostart entry, and Super+V shortcut.
func Run() error {
	exe, err := hotkeyExecutable()
	if err != nil {
		return err
	}

	binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(binDir, "linboard")
	if exe != dest {
		if err := copyFile(exe, dest); err != nil {
			return fmt.Errorf("install binary: %w", err)
		}
		_ = os.Chmod(dest, 0o755)
	}

	share := filepath.Join(os.Getenv("HOME"), ".local", "share")
	desktopDir := filepath.Join(share, "applications")
	if err := os.MkdirAll(desktopDir, 0o755); err != nil {
		return err
	}
	desktop := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=LinBoard
GenericName=Clipboard Manager
Comment=Windows-style clipboard history for Linux
Exec=%s
Icon=clipboard
Terminal=false
Categories=Utility;
StartupNotify=false
`, dest)
	if err := os.WriteFile(filepath.Join(desktopDir, "linboard.desktop"), []byte(desktop), 0o644); err != nil {
		return err
	}

	autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
	if err := os.MkdirAll(autostartDir, 0o755); err != nil {
		return err
	}
	autostart := desktop + "X-GNOME-Autostart-enabled=true\n"
	if err := os.WriteFile(filepath.Join(autostartDir, "linboard.desktop"), []byte(autostart), 0o644); err != nil {
		return err
	}

	if err := hotkey.RegisterSystemShortcutAt(dest); err != nil {
		return fmt.Errorf("installed to %s but shortcut failed: %w", dest, err)
	}

	fmt.Printf("Installed: %s\n", dest)
	fmt.Println("Super+V shortcut registered. Log out/in or restart LinBoard.")
	fmt.Println("Add ~/.local/bin to PATH if needed.")
	return nil
}

func hotkeyExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o755)
}

func HasLocalBin() bool {
	_, err := exec.LookPath("linboard")
	return err == nil
}

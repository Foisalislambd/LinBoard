package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed icons/hicolor
var hicolorFS embed.FS

// InstallThemeIcons copies the LinBoard icon set to ~/.local/share/icons.
func InstallThemeIcons() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dest := filepath.Join(home, ".local", "share", "icons", "hicolor")
	return fs.WalkDir(hicolorFS, "icons/hicolor", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("icons/hicolor", path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := hicolorFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icons/linboard-512.png
var icon512 []byte

//go:embed icons/linboard-32.png
var icon32 []byte

// Fyne returns the application window icon.
func Fyne() fyne.Resource {
	return fyne.NewStaticResource("linboard.png", icon512)
}

// TrayPNG returns a small PNG for the system tray.
func TrayPNG() []byte {
	return icon32
}

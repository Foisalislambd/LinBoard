package platform

import "os"

// UsePortalHotkey reports whether global shortcuts must use the
// xdg-desktop-portal API (GNOME/KDE Wayland). X11 grabs do not work there.
func UsePortalHotkey() bool {
	if os.Getenv("LINBOARD_FORCE_X11_HOTKEY") == "1" {
		return false
	}
	switch os.Getenv("XDG_SESSION_TYPE") {
	case "wayland":
		return true
	case "x11":
		return false
	}
	// Fallback: Wayland display without explicit x11 session.
	return os.Getenv("WAYLAND_DISPLAY") != "" && os.Getenv("DISPLAY") == ""
}

func SessionDescription() string {
	if UsePortalHotkey() {
		return "wayland"
	}
	return "x11"
}

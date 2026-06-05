package platform

import (
	"os"
	"strings"
)

// IsGNOME reports whether the current desktop is GNOME-based.
func IsGNOME() bool {
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	return strings.Contains(desktop, "gnome") || strings.Contains(desktop, "ubuntu")
}

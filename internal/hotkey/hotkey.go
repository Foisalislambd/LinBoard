package hotkey

import (
	"fmt"
	"log"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

type backend interface {
	start(func()) error
	stop()
}

// Manager registers Super+V using the best backend for the session and desktop.
type Manager struct {
	backend backend
	onPress func()
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) OnPress(fn func()) {
	m.onPress = fn
}

func (m *Manager) Start() error {
	if m.onPress == nil {
		return fmt.Errorf("hotkey: no handler")
	}

	log.Printf("hotkey: %s on %s, target %s",
		platform.SessionDescription(), platform.DesktopName(), config.HotkeyLabel)

	// 1) xdg-desktop-portal (KDE Plasma 6+, future GNOME)
	if platform.UsePortalHotkey() && portalHasGlobalShortcuts() {
		m.backend = &portalBackend{}
		if err := m.backend.start(m.onPress); err != nil {
			log.Printf("hotkey: portal failed: %v", err)
		} else {
			return nil
		}
	} else if platform.UsePortalHotkey() {
		log.Printf("hotkey: portal GlobalShortcuts not available")
	}

	// 2) Desktop environment system shortcut → linboard toggle
	if b := desktopBackend(); b != nil {
		m.backend = b
		if err := m.backend.start(m.onPress); err != nil {
			log.Printf("hotkey: %s shortcut failed: %v", platform.DesktopName(), err)
		} else {
			return nil
		}
	}

	// 3) Register Super+V via gsettings/khotkeys/etc. → `linboard toggle` (IPC)
	if regErr := RegisterSystemShortcut(); regErr == nil {
		log.Printf("hotkey registered (system): %s → linboard toggle", config.HotkeyLabel)
		m.backend = &noopBackend{}
		return nil
	} else {
		log.Printf("hotkey: system registration: %v", regErr)
	}

	return fmt.Errorf("could not bind %s — run: linboard install-shortcut", config.HotkeyLabel)
}

func desktopBackend() backend {
	switch platform.CurrentDesktop() {
	case platform.DesktopGNOME:
		return &gnomeBackend{}
	case platform.DesktopKDE:
		return &kdeBackend{}
	case platform.DesktopXFCE:
		return &xfceBackend{}
	case platform.DesktopCinnamon:
		return &cinnamonBackend{}
	default:
		return nil
	}
}

func (m *Manager) Stop() {
	if m.backend != nil {
		m.backend.stop()
	}
}

type noopBackend struct{}

func (noopBackend) start(func()) error { return nil }
func (noopBackend) stop()              {}

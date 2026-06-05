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

// Manager registers Super+V using the best backend for the session.
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

	log.Printf("hotkey: session %s, target %s", platform.SessionDescription(), config.HotkeyLabel)

	if platform.UsePortalHotkey() {
		if portalHasGlobalShortcuts() {
			m.backend = &portalBackend{}
			if err := m.backend.start(m.onPress); err == nil {
				return nil
			} else {
				log.Printf("hotkey: portal failed: %v", err)
			}
		} else {
			log.Printf("hotkey: portal GlobalShortcuts not available on this system")
		}

		if platform.IsGNOME() {
			m.backend = &gnomeBackend{}
			if err := m.backend.start(m.onPress); err != nil {
				log.Printf("hotkey: GNOME gsettings failed: %v", err)
			} else {
				return nil
			}
		}

		return fmt.Errorf("no working hotkey backend on Wayland — use tray menu or: linboard toggle")
	}

	m.backend = &x11Backend{}
	if err := m.backend.start(m.onPress); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Stop() {
	if m.backend != nil {
		m.backend.stop()
	}
}

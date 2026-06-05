package hotkey

import (
	"log"

	"golang.design/x/hotkey"

	"github.com/foisal/linboard/internal/config"
)

type Manager struct {
	hk      *hotkey.Hotkey
	onPress func()
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) OnPress(fn func()) {
	m.onPress = fn
}

func (m *Manager) Start() error {
	// Windows-style clipboard history: Super+V (Win key + V)
	m.hk = hotkey.New([]hotkey.Modifier{hotkey.Mod4}, hotkey.KeyV)
	if err := m.hk.Register(); err != nil {
		return err
	}

	go func() {
		for range m.hk.Keydown() {
			if m.onPress != nil {
				m.onPress()
			}
		}
	}()

	log.Printf("hotkey registered: %s", config.HotkeyLabel)
	return nil
}

func (m *Manager) Stop() {
	if m.hk != nil {
		_ = m.hk.Unregister()
	}
}

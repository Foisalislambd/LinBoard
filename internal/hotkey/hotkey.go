package hotkey

import (
	"log"
	"strings"

	"golang.design/x/hotkey"

	"github.com/foisal/linboard/internal/config"
)

type Manager struct {
	cfg    *config.Config
	hk     *hotkey.Hotkey
	onPress func()
}

func New(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) OnPress(fn func()) {
	m.onPress = fn
}

func (m *Manager) Start() error {
	mods, key, err := parseHotkey(m.cfg.HotkeyMod, m.cfg.HotkeyKey)
	if err != nil {
		return err
	}

	m.hk = hotkey.New(mods, key)
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

	log.Printf("hotkey registered: %s", m.cfg.HotkeyLabel())
	return nil
}

func (m *Manager) Stop() {
	if m.hk != nil {
		_ = m.hk.Unregister()
	}
}

func parseHotkey(modStr, keyStr string) ([]hotkey.Modifier, hotkey.Key, error) {
	var mods []hotkey.Modifier
	parts := strings.Split(strings.ToLower(modStr), "+")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch p {
		case "ctrl", "control":
			mods = append(mods, hotkey.ModCtrl)
		case "shift":
			mods = append(mods, hotkey.ModShift)
		case "alt":
			mods = append(mods, hotkey.Mod1) // X11 Alt
		case "super", "win", "meta":
			mods = append(mods, hotkey.Mod4) // X11 Super/Win
		case "":
			continue
		default:
			log.Printf("unknown modifier: %s", p)
		}
	}

	keyStr = strings.ToLower(strings.TrimSpace(keyStr))
	var key hotkey.Key
	switch keyStr {
	case "v":
		key = hotkey.KeyV
	case "c":
		key = hotkey.KeyC
	case "x":
		key = hotkey.KeyX
	case "b":
		key = hotkey.KeyB
	case "space":
		key = hotkey.KeySpace
	default:
		key = hotkey.KeyV
	}

	return mods, key, nil
}

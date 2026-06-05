//go:build !headless

package hotkey

import (
	"log"

	"golang.design/x/hotkey"

	"github.com/foisal/linboard/internal/config"
)

type x11Backend struct {
	hk      *hotkey.Hotkey
	onPress func()
}

func newX11Backend() backend {
	return &x11Backend{}
}

func (b *x11Backend) start(onPress func()) error {
	b.onPress = onPress
	// Super/Win = Mod4 on X11
	b.hk = hotkey.New([]hotkey.Modifier{hotkey.Mod4}, hotkey.KeyV)
	if err := b.hk.Register(); err != nil {
		return err
	}
	go func() {
		for range b.hk.Keydown() {
			if b.onPress != nil {
				b.onPress()
			}
		}
	}()
	log.Printf("hotkey registered (X11): %s", config.HotkeyLabel)
	return nil
}

func (b *x11Backend) stop() {
	if b.hk != nil {
		_ = b.hk.Unregister()
		b.hk = nil
	}
}

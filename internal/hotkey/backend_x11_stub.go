//go:build headless

package hotkey

import "fmt"

type x11Backend struct{}

func newX11Backend() backend {
	return &x11Backend{}
}

func (b *x11Backend) start(func()) error {
	return fmt.Errorf("X11 hotkey unavailable in headless build")
}

func (b *x11Backend) stop() {}

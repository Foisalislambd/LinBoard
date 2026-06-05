package hotkey

import (
	"fmt"
	"log"

	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

type xfceBackend struct{}

func (b *xfceBackend) start(_ func()) error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	if err := setupXFCEHotkey(exe); err != nil {
		return err
	}
	log.Printf("hotkey registered (XFCE): %s → linboard toggle", config.HotkeyLabel)
	return nil
}

func (b *xfceBackend) stop() {}

func setupXFCEHotkey(exe string) error {
	if !platform.IsXFCE() {
		return skip("not XFCE")
	}
	if !hasBin("xfconf-query") {
		return fmt.Errorf("xfconf-query not found")
	}
	prop := "/commands/custom/LinBoard"
	if err := run("xfconf-query", "-c", "xfce4-keyboard-shortcuts", "-p", prop, "-s", exe+" toggle", "-n"); err != nil {
		return err
	}
	return run("xfconf-query", "-c", "xfce4-keyboard-shortcuts", "-p", prop+"/default", "-t", "string", "-s", "Super+v", "--create")
}

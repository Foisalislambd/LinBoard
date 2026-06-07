package hotkey

import (
	"fmt"
	"os"
	"strings"

	"github.com/foisal/linboard/internal/clipboard"
	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/platform"
)

// SetupAt registers Super+V for exe (idempotent). Call from install and app start.
func SetupAt(exe string) error {
	if platform.IsGNOME() {
		return setupGNOME(exe)
	}
	return registerSystemShortcut(exe)
}

// VerifyReport is the result of post-install shortcut checks.
type VerifyReport struct {
	OK     []string
	Warn   []string
	Fail   []string
	Binary string
}

func (r VerifyReport) Healthy() bool {
	return len(r.Fail) == 0
}

// Verify checks shortcut prerequisites after SetupAt.
func Verify(exe string) VerifyReport {
	r := VerifyReport{Binary: exe}
	defer appendPasteVerify(&r)

	if !platform.IsGNOME() {
		r.OK = append(r.OK, "desktop: "+platform.DesktopName())
		return r
	}

	if !hasBin("gsettings") {
		r.Fail = append(r.Fail, "gsettings not found")
		return r
	}

	wrapper, err := ToggleWrapperPath()
	if err != nil {
		r.Fail = append(r.Fail, err.Error())
		return r
	}

	if st, err := os.Stat(wrapper); err != nil || st.Mode()&0o111 == 0 {
		r.Fail = append(r.Fail, "toggle launcher missing: "+wrapper)
	} else {
		r.OK = append(r.OK, "toggle launcher: "+wrapper)
	}

	schema := gnomeBindingSchema()
	cmd, err := gsettingsCommand(schema, "command")
	if err != nil {
		r.Fail = append(r.Fail, "gsettings shortcut not registered")
	} else if cmd != wrapper {
		r.Warn = append(r.Warn, fmt.Sprintf("gsettings command is %q (want %q)", cmd, wrapper))
	} else {
		r.OK = append(r.OK, "gsettings command OK")
	}

	bind, err := gsettingsGet(schema, "binding")
	if err != nil {
		r.Fail = append(r.Fail, "shortcut binding not set")
	} else if strings.Trim(bind, "'") != "<Super>v" {
		r.Warn = append(r.Warn, "binding is "+bind+" (want <Super>v)")
	} else {
		r.OK = append(r.OK, "binding: Super+V")
	}

	paths, err := gsettingsListPaths()
	if err != nil || !containsPath(paths, gnomeBindingPath()) {
		r.Fail = append(r.Fail, "custom-linboard not in media-keys list")
	} else {
		r.OK = append(r.OK, "registered in GNOME media-keys")
	}

	if mediaKeysRunning() {
		r.OK = append(r.OK, "gsd-media-keys running")
	} else {
		r.Fail = append(r.Fail, "gsd-media-keys not running (Super+V will not work)")
	}

	tray, err := gsettingsGetArray(gnomeShellKeybindings, gnomeMessageTrayKey)
	if err == nil {
		for _, b := range tray {
			if strings.EqualFold(b, "<Super>v") {
				r.Warn = append(r.Warn, "GNOME message tray still uses Super+V — run install-shortcut again")
				break
			}
		}
	}

	return r
}

func appendPasteVerify(r *VerifyReport) {
	if clipboard.PasteReady() {
		r.OK = append(r.OK, "auto-paste ready ("+clipboard.PasteToolName()+")")
		return
	}
	if clipboard.SessionNeedsRelogin() {
		r.Warn = append(r.Warn, "auto-paste: log out/in to activate input group (or: linboard-start)")
		return
	}
	r.Warn = append(r.Warn, "auto-paste not ready — run: linboard setup-paste")
	if !platform.UsePortalHotkey() && !hasBin("xdotool") {
		r.Warn = append(r.Warn, "X11 fallback: install xdotool if uinput setup is unavailable")
	}
}

// PrintVerify writes a human-readable report to stdout.
func PrintVerify(r VerifyReport) {
	fmt.Println()
	fmt.Println("Shortcut check (" + config.HotkeyLabel + "):")
	for _, s := range r.OK {
		fmt.Println("  ✓", s)
	}
	for _, s := range r.Warn {
		fmt.Println("  !", s)
	}
	for _, s := range r.Fail {
		fmt.Println("  ✗", s)
	}
	if r.Healthy() {
		fmt.Println()
		fmt.Println("Super+V is ready. Press Win+V to open clipboard history.")
	} else {
		fmt.Println()
		fmt.Println("Fix issues above, then run: linboard install-shortcut")
	}
	fmt.Println()
}

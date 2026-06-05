package hotkey

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/foisal/linboard/internal/config"
)

const (
	portalBusName    = "org.freedesktop.portal.Desktop"
	portalObjectPath = "/org/freedesktop/portal/desktop"
	portalIface      = "org.freedesktop.portal.GlobalShortcuts"
	requestIface     = "org.freedesktop.portal.Request"
	shortcutID       = "linboard_toggle"
)

var objectPathRE = regexp.MustCompile(`objectpath\s+'([^']+)'`)

type portalBackend struct {
	onPress  func()
	stopCh   chan struct{}
	stopOnce sync.Once
	monitor  *exec.Cmd
}

func (b *portalBackend) start(onPress func()) error {
	b.onPress = onPress
	b.stopCh = make(chan struct{})

	if err := b.setup(); err != nil {
		return err
	}

	if err := b.startActivatedMonitor(); err != nil {
		return err
	}

	log.Printf("hotkey registered (portal): %s — configure in Settings → Keyboard if prompted", config.HotkeyLabel)
	return nil
}

func (b *portalBackend) setup() error {
	session, err := b.createSession()
	if err != nil {
		return err
	}
	return b.bindShortcuts(session)
}

func (b *portalBackend) createSession() (string, error) {
	token := randomToken()
	opts := fmt.Sprintf("{'handle_token': <%q>, 'session_handle_token': <'linboard'>}", token)

	out, err := gdbusCall(portalIface+".CreateSession", opts)
	if err != nil {
		return "", fmt.Errorf("CreateSession: %w", err)
	}
	requestPath := parseObjectPath(out)
	if requestPath == "" {
		return "", fmt.Errorf("CreateSession: missing request path")
	}

	status, results, err := waitPortalResponse(requestPath, 30*time.Second)
	if err != nil {
		return "", err
	}
	if status != 0 {
		return "", fmt.Errorf("CreateSession failed (status %d)", status)
	}
	session, ok := results["session_handle"]
	if !ok || session == "" {
		return "", fmt.Errorf("CreateSession: missing session_handle")
	}
	return session, nil
}

func (b *portalBackend) bindShortcuts(session string) error {
	token := randomToken()
	shortcuts := fmt.Sprintf(
		"[('%s', {'description': <%q>, 'preferred_trigger': <'LOGO+V'>})]",
		shortcutID, config.AppName,
	)
	opts := fmt.Sprintf("{'handle_token': <%q>}", token)

	out, err := gdbusCall(
		portalIface+".BindShortcuts",
		fmt.Sprintf("objectpath '%s'", session),
		shortcuts,
		"''",
		opts,
	)
	if err != nil {
		return fmt.Errorf("BindShortcuts: %w", err)
	}
	requestPath := parseObjectPath(out)
	if requestPath == "" {
		return fmt.Errorf("BindShortcuts: missing request path")
	}

	status, _, err := waitPortalResponse(requestPath, 60*time.Second)
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("BindShortcuts failed (status %d)", status)
	}
	return nil
}

func (b *portalBackend) startActivatedMonitor() error {
	if !hasBin("dbus-monitor") {
		return fmt.Errorf("dbus-monitor not found (install dbus-user-session)")
	}

	match := fmt.Sprintf(
		"type='signal',interface='%s',member='Activated',path='%s'",
		portalIface, portalObjectPath,
	)
	cmd := exec.Command("dbus-monitor", "--session", match)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	b.monitor = cmd

	go b.readActivated(stdout)
	return nil
}

func (b *portalBackend) readActivated(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-b.stopCh:
			return
		default:
		}
		line := scanner.Text()
		if !strings.Contains(line, "member=Activated") {
			continue
		}
		// Next lines contain signal body; shortcut id is the second argument.
		if !scanner.Scan() {
			continue
		}
		for scanner.Scan() {
			body := scanner.Text()
			if strings.TrimSpace(body) == "" {
				break
			}
			if strings.Contains(body, `string "`+shortcutID+`"`) && b.onPress != nil {
				b.onPress()
				break
			}
		}
	}
}

func (b *portalBackend) stop() {
	b.stopOnce.Do(func() {
		close(b.stopCh)
		if b.monitor != nil && b.monitor.Process != nil {
			_ = b.monitor.Process.Kill()
		}
	})
}

func gdbusCall(method string, args ...string) (string, error) {
	callArgs := []string{
		"call", "--session",
		"--dest", portalBusName,
		"--object-path", portalObjectPath,
		"--method", method,
	}
	callArgs = append(callArgs, args...)
	out, err := exec.Command("gdbus", callArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w (%s)", method, err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func waitPortalResponse(requestPath string, timeout time.Duration) (uint32, map[string]string, error) {
	if !hasBin("dbus-monitor") {
		return 2, nil, fmt.Errorf("dbus-monitor not found")
	}

	match := fmt.Sprintf(
		"type='signal',interface='%s',member='Response',path='%s'",
		requestIface, requestPath,
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "dbus-monitor", "--session", match)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 2, nil, err
	}
	if err := cmd.Start(); err != nil {
		return 2, nil, err
	}

	status, results := parsePortalResponse(stdout)
	_ = cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return 2, nil, fmt.Errorf("portal request timed out")
	}
	return status, results, nil
}

func parsePortalResponse(r io.Reader) (uint32, map[string]string) {
	results := make(map[string]string)
	scanner := bufio.NewScanner(r)
	var status uint32
	gotStatus := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "uint32 ") && !gotStatus {
			fmt.Sscanf(line, "uint32 %d", &status)
			gotStatus = true
			continue
		}
		if strings.Contains(line, `string "session_handle"`) {
			continue
		}
		if strings.Contains(line, "objectpath") {
			if m := objectPathRE.FindStringSubmatch(line); len(m) == 2 {
				results["session_handle"] = m[1]
			}
		}
	}
	return status, results
}

func parseObjectPath(out string) string {
	if m := objectPathRE.FindStringSubmatch(out); len(m) == 2 {
		return m[1]
	}
	// gdbus sometimes prints bare quoted paths
	out = strings.TrimSpace(out)
	if strings.HasPrefix(out, "(['") {
		out = strings.TrimPrefix(out, "(['")
		out = strings.TrimSuffix(out, "',])")
		out = strings.TrimSuffix(out, "'])")
		if strings.HasPrefix(out, "/") {
			return out
		}
	}
	return ""
}

func randomToken() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "linboard_" + hex.EncodeToString(b)
}

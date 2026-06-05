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

var (
	objectPathRE  = regexp.MustCompile(`objectpath\s+(?:'([^']+)'|(/[^\s;]+))`)
	signalPathRE  = regexp.MustCompile(`\bpath=([^;\s]+)`)
)

type portalBackend struct {
	onPress  func()
	stopCh   chan struct{}
	stopOnce sync.Once
	monitor  *exec.Cmd
}

func (b *portalBackend) start(onPress func()) error {
	if !hasBin("gdbus") {
		return fmt.Errorf("gdbus not found (install libglib2.0-bin)")
	}
	if !hasBin("dbus-monitor") {
		return fmt.Errorf("dbus-monitor not found (install dbus-user-session)")
	}

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
	opts := fmt.Sprintf("{'handle_token': <'%s'>, 'session_handle_token': <'linboard'>}", token)

	status, results, err := portalCall(30*time.Second, portalIface+".CreateSession", opts)
	if err != nil {
		return "", fmt.Errorf("CreateSession: %w", err)
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
		"[('%s', {'description': <'%s'>, 'preferred_trigger': <'LOGO+V'>})]",
		shortcutID, config.AppName,
	)
	opts := fmt.Sprintf("{'handle_token': <'%s'>}", token)

	status, _, err := portalCall(
		60*time.Second,
		portalIface+".BindShortcuts",
		fmt.Sprintf("objectpath '%s'", session),
		shortcuts,
		"''",
		opts,
	)
	if err != nil {
		return fmt.Errorf("BindShortcuts: %w", err)
	}
	if status != 0 {
		return fmt.Errorf("BindShortcuts failed (status %d)", status)
	}
	return nil
}

// portalCall starts dbus-monitor before the gdbus call to avoid missing fast Response signals.
func portalCall(timeout time.Duration, method string, args ...string) (uint32, map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	match := fmt.Sprintf("type='signal',interface='%s',member='Response'", requestIface)
	cmd := exec.CommandContext(ctx, "dbus-monitor", "--session", match)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 2, nil, err
	}
	if err := cmd.Start(); err != nil {
		return 2, nil, err
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}()

	type response struct {
		path    string
		status  uint32
		results map[string]string
	}
	respCh := make(chan response, 1)

	var expectPath string
	var expectMu sync.Mutex

	go func() {
		parsePortalResponses(stdout, func(path string, status uint32, results map[string]string) bool {
			expectMu.Lock()
			want := expectPath
			expectMu.Unlock()
			if want == "" || path != want {
				return false
			}
			select {
			case respCh <- response{path: path, status: status, results: results}:
			default:
			}
			return true
		})
	}()

	out, err := gdbusCall(method, args...)
	if err != nil {
		return 2, nil, err
	}
	path := parseObjectPath(out)
	if path == "" {
		return 2, nil, fmt.Errorf("%s: missing request path in %q", method, strings.TrimSpace(out))
	}

	expectMu.Lock()
	expectPath = path
	expectMu.Unlock()

	select {
	case r := <-respCh:
		return r.status, r.results, nil
	case <-ctx.Done():
		return 2, nil, fmt.Errorf("portal request timed out")
	}
}

func parsePortalResponses(r io.Reader, onResponse func(path string, status uint32, results map[string]string) bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "member=Response") {
			continue
		}
		path := parseSignalPath(line)
		if path == "" {
			continue
		}
		status, results := readResponseBody(scanner)
		if onResponse(path, status, results) {
			return
		}
	}
}

func readResponseBody(scanner *bufio.Scanner) (uint32, map[string]string) {
	results := make(map[string]string)
	var status uint32
	gotStatus := false

	for scanner.Scan() {
		body := strings.TrimSpace(scanner.Text())
		if body == "" {
			if gotStatus {
				return status, results
			}
			break
		}
		if strings.HasPrefix(body, "uint32 ") && !gotStatus {
			fmt.Sscanf(body, "uint32 %d", &status)
			gotStatus = true
			continue
		}
		if strings.Contains(body, `string "session_handle"`) {
			continue
		}
		if strings.Contains(body, "objectpath") {
			if m := objectPathRE.FindStringSubmatch(body); len(m) >= 2 {
				p := m[1]
				if p == "" {
					p = m[2]
				}
				results["session_handle"] = p
			}
		}
	}
	return status, results
}

func (b *portalBackend) startActivatedMonitor() error {
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
		if !strings.Contains(scanner.Text(), "member=Activated") {
			continue
		}
		for scanner.Scan() {
			body := strings.TrimSpace(scanner.Text())
			if body == "" {
				break
			}
			if strings.Contains(body, `string "`+shortcutID+`"`) {
				if b.onPress != nil {
					b.onPress()
				}
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

func parseObjectPath(out string) string {
	if m := objectPathRE.FindStringSubmatch(out); len(m) >= 2 {
		if m[1] != "" {
			return m[1]
		}
		return m[2]
	}
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

func parseSignalPath(line string) string {
	if m := signalPathRE.FindStringSubmatch(line); len(m) == 2 {
		return m[1]
	}
	return ""
}

func randomToken() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "linboard_" + hex.EncodeToString(b)
}

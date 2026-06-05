package hotkey

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/foisal/linboard/internal/config"
)

const (
	portalBusName    = "org.freedesktop.portal.Desktop"
	portalObjectPath = "/org/freedesktop/portal/desktop"
	portalIface      = "org.freedesktop.portal.GlobalShortcuts"
	requestIface     = "org.freedesktop.portal.Request"
	shortcutID       = "linboard_toggle"
)

type portalBackend struct {
	conn     *dbus.Conn
	onPress  func()
	stopCh   chan struct{}
	stopOnce sync.Once
	session  dbus.ObjectPath
}

func (b *portalBackend) start(onPress func()) error {
	b.onPress = onPress
	b.stopCh = make(chan struct{})

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("session bus: %w", err)
	}
	b.conn = conn

	if err := b.setup(); err != nil {
		conn.Close()
		return err
	}

	go b.listenActivated()
	log.Printf("hotkey registered (portal): %s — configure in Settings → Keyboard if prompted", config.HotkeyLabel)
	return nil
}

func (b *portalBackend) setup() error {
	session, err := b.createSession()
	if err != nil {
		return err
	}
	b.session = session

	if err := b.bindShortcuts(session); err != nil {
		return err
	}
	return nil
}

func (b *portalBackend) createSession() (dbus.ObjectPath, error) {
	obj := b.conn.Object(portalBusName, dbus.ObjectPath(portalObjectPath))
	options := map[string]dbus.Variant{
		"handle_token":         dbus.MakeVariant(randomToken()),
		"session_handle_token": dbus.MakeVariant("linboard"),
	}

	var requestPath dbus.ObjectPath
	if err := obj.Call(portalIface+".CreateSession", 0, options).Store(&requestPath); err != nil {
		return "", fmt.Errorf("CreateSession: %w", err)
	}

	status, results, err := waitPortalResponse(b.conn, requestPath, 30*time.Second)
	if err != nil {
		return "", err
	}
	if status != 0 {
		return "", fmt.Errorf("CreateSession failed (status %d)", status)
	}

	session, ok := results["session_handle"].Value().(string)
	if !ok || session == "" {
		return "", fmt.Errorf("CreateSession: missing session_handle")
	}
	return dbus.ObjectPath(session), nil
}

func (b *portalBackend) bindShortcuts(session dbus.ObjectPath) error {
	shortcuts := []interface{}{
		[]interface{}{
			shortcutID,
			map[string]dbus.Variant{
				"description":       dbus.MakeVariant("LinBoard — Show clipboard history"),
				"preferred_trigger": dbus.MakeVariant("LOGO+V"),
			},
		},
	}

	obj := b.conn.Object(portalBusName, dbus.ObjectPath(portalObjectPath))
	options := map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(randomToken()),
	}

	var requestPath dbus.ObjectPath
	if err := obj.Call(portalIface+".BindShortcuts", 0, session, shortcuts, "", options).Store(&requestPath); err != nil {
		return fmt.Errorf("BindShortcuts: %w", err)
	}

	status, _, err := waitPortalResponse(b.conn, requestPath, 60*time.Second)
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("BindShortcuts failed (status %d)", status)
	}
	return nil
}

func (b *portalBackend) listenActivated() {
	if err := b.conn.AddMatchSignal(
		dbus.WithMatchInterface(portalIface),
		dbus.WithMatchMember("Activated"),
		dbus.WithMatchObjectPath(dbus.ObjectPath(portalObjectPath)),
	); err != nil {
		log.Printf("portal: listen match: %v", err)
		return
	}

	ch := make(chan *dbus.Signal, 8)
	b.conn.Signal(ch)

	for {
		select {
		case <-b.stopCh:
			b.conn.RemoveSignal(ch)
			return
		case sig := <-ch:
			if sig == nil {
				continue
			}
			if sig.Name != portalIface+".Activated" {
				continue
			}
			if len(sig.Body) < 2 {
				continue
			}
			id, _ := sig.Body[1].(string)
			if id == shortcutID && b.onPress != nil {
				b.onPress()
			}
		}
	}
}

func (b *portalBackend) stop() {
	b.stopOnce.Do(func() {
		close(b.stopCh)
		if b.conn != nil {
			_ = b.conn.Close()
		}
	})
}

func waitPortalResponse(conn *dbus.Conn, requestPath dbus.ObjectPath, timeout time.Duration) (uint32, map[string]dbus.Variant, error) {
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(requestIface),
		dbus.WithMatchMember("Response"),
		dbus.WithMatchObjectPath(requestPath),
	); err != nil {
		return 2, nil, err
	}

	ch := make(chan *dbus.Signal, 1)
	conn.Signal(ch)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			conn.RemoveSignal(ch)
			return 2, nil, fmt.Errorf("portal request timed out")
		case sig := <-ch:
			if sig == nil {
				continue
			}
			if sig.Path != requestPath || sig.Name != requestIface+".Response" {
				continue
			}
			conn.RemoveSignal(ch)
			if len(sig.Body) < 2 {
				return 2, nil, fmt.Errorf("invalid portal response")
			}
			status, _ := sig.Body[0].(uint32)
			results, _ := sig.Body[1].(map[string]dbus.Variant)
			return status, results, nil
		}
	}
}

func randomToken() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "linboard_" + hex.EncodeToString(b)
}

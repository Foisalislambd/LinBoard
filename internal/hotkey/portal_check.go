package hotkey

import (
	"strings"

	"github.com/godbus/dbus/v5"
)

func portalHasGlobalShortcuts() bool {
	conn, err := dbus.SessionBus()
	if err != nil {
		return false
	}
	defer conn.Close()

	obj := conn.Object(portalBusName, dbus.ObjectPath(portalObjectPath))
	var xml string
	if err := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&xml); err != nil {
		return false
	}
	return strings.Contains(xml, portalIface)
}

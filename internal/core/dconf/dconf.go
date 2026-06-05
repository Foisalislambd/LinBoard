// Package dconf reads and writes user settings via the session D-Bus (no gsettings CLI).
package dconf

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	busName    = "ca.desrt.dconf"
	readerPath = "/ca/desrt/dconf/Reader/user"
	writerPath = "/ca/desrt/dconf/Writer/user"
	readerIF   = "ca.desrt.dconf.Reader"
	writerIF   = "ca.desrt.dconf.Writer"
)

func withSession(fn func(*dbus.Conn) error) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}

// ReadEntries returns all keys under a dconf path.
func ReadEntries(path string) (map[string]dbus.Variant, error) {
	var out map[string]dbus.Variant
	err := withSession(func(conn *dbus.Conn) error {
		obj := conn.Object(busName, dbus.ObjectPath(readerPath))
		return obj.Call(readerIF+".ReadEntries", 0, path).Store(&out)
	})
	return out, err
}

// ReadString reads a string key from a dconf path.
func ReadString(path, key string) (string, error) {
	entries, err := ReadEntries(path)
	if err != nil {
		return "", err
	}
	v, ok := entries[key]
	if !ok {
		return "", fmt.Errorf("dconf: key %q not found at %s", key, path)
	}
	s, ok := v.Value().(string)
	if !ok {
		return "", fmt.Errorf("dconf: key %q is not a string", key)
	}
	return s, nil
}

// ReadStringArray reads a string array key from a dconf path.
func ReadStringArray(path, key string) ([]string, error) {
	entries, err := ReadEntries(path)
	if err != nil {
		return nil, err
	}
	v, ok := entries[key]
	if !ok {
		return nil, nil
	}
	switch arr := v.Value().(type) {
	case []string:
		return arr, nil
	case []interface{}:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("dconf: key %q is not a string array", key)
	}
}

// SetString sets a string key at a dconf path.
func SetString(path, key, value string) error {
	return change([]dbusChange{{typ: "s", path: path, key: key, value: dbus.MakeVariant(value)}}, nil)
}

// SetStringArray sets a string array key at a dconf path.
func SetStringArray(path, key string, values []string) error {
	// dconf change type "s" = set; value variant carries the "as" payload.
	return change([]dbusChange{{typ: "s", path: path, key: key, value: dbus.MakeVariant(values)}}, nil)
}

type dbusChange struct {
	typ   string
	path  string
	key   string
	value dbus.Variant
}

func change(changes []dbusChange, redirects [][3]string) error {
	return withSession(func(conn *dbus.Conn) error {
		payload := make([]struct {
			Type  string
			Path  string
			Key   string
			Value dbus.Variant
		}, len(changes))
		for i, c := range changes {
			payload[i] = struct {
				Type  string
				Path  string
				Key   string
				Value dbus.Variant
			}{Type: c.typ, Path: c.path, Key: c.key, Value: c.value}
		}
		obj := conn.Object(busName, dbus.ObjectPath(writerPath))
		return obj.Call(writerIF+".Change", 0, payload, redirects).Err
	})
}

//go:build !linux

package clipboard

import "errors"

var errUinputUnavailable = errors.New("uinput unavailable")

func uinputReady() bool { return false }

func pasteViaUinput() error { return errUinputUnavailable }

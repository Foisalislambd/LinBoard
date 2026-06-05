package clipboard

import (
	"fmt"

	"github.com/foisal/linboard/internal/platform"
)

// ReadToolName returns the first available clipboard read backend for diagnostics.
func ReadToolName() string {
	for _, name := range []string{"wl-paste", "xclip", "xsel"} {
		if readerAvailable(name) {
			return name
		}
	}
	return "none"
}

func readerAvailable(name string) bool {
	return haveCommand(name)
}

func textReaders() []func() (string, error) {
	if platform.UsePortalHotkey() {
		return []func() (string, error){
			readTextWLPaste,
			readTextXClip,
			readTextXSel,
		}
	}
	return []func() (string, error){
		readTextXClip,
		readTextXSel,
		readTextWLPaste,
	}
}

func imageReaders() []func() ([]byte, error) {
	if platform.UsePortalHotkey() {
		return []func() ([]byte, error){
			readImageWLPaste,
			readImageXClip,
		}
	}
	return []func() ([]byte, error){
		readImageXClip,
		readImageWLPaste,
	}
}

func readText() (string, error) {
	for _, r := range textReaders() {
		if text, err := r(); err == nil {
			return text, nil
		}
	}
	return "", fmt.Errorf("clipboard read failed")
}

func readImage() ([]byte, error) {
	for _, r := range imageReaders() {
		if data, err := r(); err == nil && len(data) > 0 {
			return data, nil
		}
	}
	return nil, fmt.Errorf("clipboard image read failed")
}

func readTextWLPaste() (string, error) {
	out, err := outputFromCommand("wl-paste", []string{"-n"})
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func readTextXClip() (string, error) {
	out, err := outputFromCommand("xclip", []string{"-selection", "clipboard", "-o"})
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func readTextXSel() (string, error) {
	out, err := outputFromCommand("xsel", []string{"--clipboard", "--output"})
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func readImageWLPaste() ([]byte, error) {
	return outputFromCommand("wl-paste", []string{"-t", "image/png"})
}

func readImageXClip() ([]byte, error) {
	return outputFromCommand("xclip", []string{"-selection", "clipboard", "-t", "image/png", "-o"})
}

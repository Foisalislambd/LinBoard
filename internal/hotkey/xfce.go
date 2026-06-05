package hotkey

import (
	"encoding/xml"
	"log"
	"os"
	"path/filepath"

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

type xfceChannel struct {
	XMLName  xml.Name     `xml:"channel"`
	Name     string       `xml:"name,attr"`
	Version  string       `xml:"version,attr"`
	Property []xfceProperty `xml:"property"`
}

type xfceProperty struct {
	Name     string         `xml:"name,attr"`
	Type     string         `xml:"type,attr"`
	Value    string         `xml:"value,attr,omitempty"`
	Property []xfceProperty `xml:"property"`
}

func setupXFCEHotkey(exe string) error {
	if !platform.IsXFCE() {
		return skip("not XFCE")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", "xfce4", "xfconf", "xfce-perchannel-xml", "xfce4-keyboard-shortcuts.xml")

	ch := xfceChannel{Name: "xfce4-keyboard-shortcuts", Version: "1.0"}
	if data, err := os.ReadFile(path); err == nil {
		_ = xml.Unmarshal(data, &ch)
	}
	if ch.Name == "" {
		ch.Name = "xfce4-keyboard-shortcuts"
		ch.Version = "1.0"
	}

	commands := findProperty(&ch.Property, "commands")
	if commands == nil {
		ch.Property = append(ch.Property, xfceProperty{Name: "commands", Type: "empty"})
		commands = &ch.Property[len(ch.Property)-1]
	}
	custom := findProperty(&commands.Property, "custom")
	if custom == nil {
		commands.Property = append(commands.Property, xfceProperty{Name: "custom", Type: "empty"})
		custom = &commands.Property[len(commands.Property)-1]
	}

	linboard := findProperty(&custom.Property, "LinBoard")
	if linboard == nil {
		custom.Property = append(custom.Property, xfceProperty{
			Name: "LinBoard", Type: "string", Value: exe + " toggle",
			Property: []xfceProperty{
				{Name: "default", Type: "string", Value: "Super+v"},
			},
		})
	} else {
		linboard.Type = "string"
		linboard.Value = exe + " toggle"
		def := findProperty(&linboard.Property, "default")
		if def == nil {
			linboard.Property = append(linboard.Property, xfceProperty{Name: "default", Type: "string", Value: "Super+v"})
		} else {
			def.Type = "string"
			def.Value = "Super+v"
		}
	}

	out, err := xml.MarshalIndent(ch, "", "  ")
	if err != nil {
		return err
	}
	doc := append([]byte(xml.Header), out...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, doc, 0o644)
}

func findProperty(props *[]xfceProperty, name string) *xfceProperty {
	for i := range *props {
		if (*props)[i].Name == name {
			return &(*props)[i]
		}
	}
	return nil
}

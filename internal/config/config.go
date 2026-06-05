package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	AppName    = "LinBoard"
	AppVersion = "1.0.0"
)

type Config struct {
	MaxHistory     int    `json:"max_history"`
	HotkeyMod      string `json:"hotkey_mod"`      // ctrl+shift, super, etc.
	HotkeyKey      string `json:"hotkey_key"`      // v, c, etc.
	StartMinimized bool   `json:"start_minimized"`
	PasteOnSelect  bool   `json:"paste_on_select"` // auto-paste when item selected
	Theme          string `json:"theme"`           // light, dark, system
}

func Default() *Config {
	return &Config{
		MaxHistory:     200,
		HotkeyMod:      "ctrl+shift",
		HotkeyKey:      "v",
		StartMinimized: true,
		PasteOnSelect:  true,
		Theme:          "system",
	}
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "linboard")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func DataDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(data, 0o755); err != nil {
		return "", err
	}
	return data, nil
}

func ImagesDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	images := filepath.Join(dir, "images")
	if err := os.MkdirAll(images, 0o755); err != nil {
		return "", err
	}
	return images, nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if saveErr := cfg.Save(); saveErr != nil {
			return nil, saveErr
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c *Config) HotkeyLabel() string {
	return c.HotkeyMod + "+" + c.HotkeyKey
}

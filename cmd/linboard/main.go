package main

import (
	"fmt"
	"log"
	"os"

	"github.com/foisal/linboard/internal/app"
	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/hotkey"
	"github.com/foisal/linboard/internal/install"
	"github.com/foisal/linboard/internal/ipc"
	"github.com/foisal/linboard/internal/singleinstance"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[LinBoard] ")

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "toggle":
			if err := ipc.Toggle(); err != nil {
				log.Fatal(err)
			}
			return
		case "install":
			if err := install.Run(); err != nil {
				log.Fatal(err)
			}
			return
		case "install-shortcut":
			if err := hotkey.RegisterSystemShortcut(); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Super+V shortcut registered.")
			return
		case "version", "-v", "--version":
			fmt.Printf("LinBoard %s\n", config.AppVersion)
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		log.Fatal("No display found. LinBoard requires X11 or Wayland.")
	}

	release, err := singleinstance.Acquire()
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	defer release()

	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	application.Run()
}

func printHelp() {
	fmt.Printf(`LinBoard %s — open-source clipboard manager for Linux

Usage:
  linboard                  Start LinBoard (system tray, background)
  linboard toggle           Show/hide history (used by Super+V shortcut)
  linboard install          Install to ~/.local/bin + autostart + Super+V
  linboard install-shortcut Register Super+V only
  linboard version          Print version
  linboard help             Show this help

Hotkey: Super+V (Win+V) — like Windows clipboard history.

Supported: GNOME, KDE Plasma, XFCE, Cinnamon, X11 and Wayland.
`, config.AppVersion)
}

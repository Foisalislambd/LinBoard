package main

import (
	"log"
	"os"

	"github.com/foisal/linboard/internal/app"
	"github.com/foisal/linboard/internal/ipc"
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
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		log.Fatal("No display found. LinBoard requires X11 or Wayland.")
	}

	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	application.Run()
}

func printHelp() {
	log.Print(`LinBoard — Linux clipboard manager

Usage:
  linboard          Start LinBoard (system tray)
  linboard toggle   Show/hide history (for GNOME custom shortcut)
  linboard help     Show this help

Hotkey: Super+V (Win+V) via xdg-desktop-portal on Wayland, X11 grab on X11.
If Super+V does not work, add a custom shortcut in Settings → Keyboard:
  Command: /path/to/linboard toggle
`)
}

package main

import (
	"log"
	"os"

	"github.com/foisal/linboard/internal/app"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[LinBoard] ")

	// Fyne needs a display; skip if headless
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		log.Fatal("No display found. LinBoard requires X11 or Wayland.")
	}

	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	application.Run()
}

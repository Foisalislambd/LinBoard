package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"github.com/foisal/linboard/internal/clipboard"
	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/hotkey"
	"github.com/foisal/linboard/internal/ipc"
	"github.com/foisal/linboard/internal/store"
	"github.com/foisal/linboard/internal/tray"
	"github.com/foisal/linboard/internal/ui"
)

type App struct {
	cfg          *config.Config
	fyneApp      fyne.App
	store        *store.Store
	history      *ui.HistoryWindow
	monitor      *clipboard.Monitor
	hotkey       *hotkey.Manager
	tray         *tray.Tray
	ipc          *ipc.Server
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	s, err := store.Open(cfg.MaxHistory)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	a := &App{
		cfg:    cfg,
		store:  s,
		ctx:    ctx,
		cancel: cancel,
	}

	a.fyneApp = fyneapp.NewWithID("com.linboard.app")
	a.fyneApp.SetIcon(nil)

	switch cfg.Theme {
	case "dark":
		a.fyneApp.Settings().SetTheme(theme.DarkTheme())
	case "light":
		a.fyneApp.Settings().SetTheme(theme.LightTheme())
	}

	a.history = ui.NewHistoryWindow(a.fyneApp, s, cfg)
	a.monitor = clipboard.NewMonitor(s)
	a.hotkey = hotkey.New()
	a.tray = tray.New(cfg)

	return a, nil
}

func (a *App) showHistory() {
	a.fyneApp.Driver().DoFromGoroutine(func() {
		a.history.Toggle()
	}, false)
}

func (a *App) Run() {
	a.monitor.OnChange(a.updateTrayCount)
	a.monitor.Start(a.ctx)

	a.hotkey.OnPress(a.showHistory)

	ipcSrv, err := ipc.StartServer(a.showHistory)
	if err != nil {
		log.Printf("ipc server failed: %v", err)
	} else {
		a.ipc = ipcSrv
	}

	a.tray.OnShow(func() {
		a.fyneApp.Driver().DoFromGoroutine(func() {
			a.history.Show()
		}, false)
	}) // tray menu always works as fallback
	a.tray.OnClear(func() {
		if err := a.store.ClearUnpinned(); err != nil {
			log.Printf("clear history: %v", err)
		}
		a.updateTrayCount()
	})
	a.tray.OnQuit(func() {
		a.Shutdown()
	})

	go func() {
		if err := a.hotkey.Start(); err != nil {
			log.Printf("hotkey registration failed: %v", err)
			log.Printf("tray icon → Show History still works")
		}
	}()

	go a.tray.Run(a.updateTrayCount)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		a.tray.Quit()
		a.Shutdown()
	}()

	a.fyneApp.Run()
}

func (a *App) updateTrayCount() {
	n, err := a.store.Count()
	if err != nil {
		return
	}
	a.tray.SetCount(n)
}

func (a *App) Shutdown() {
	a.shutdownOnce.Do(func() {
		a.cancel()
		a.hotkey.Stop()
		if a.ipc != nil {
			a.ipc.Close()
		}
		_ = a.store.Close()
		a.fyneApp.Quit()
	})
}

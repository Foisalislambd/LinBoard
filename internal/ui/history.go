package ui

import (
	"image/color"
	"os"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/foisal/linboard/internal/clipboard"
	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/store"
)

type HistoryWindow struct {
	app      fyne.App
	win      fyne.Window
	store    *store.Store
	cfg      *config.Config
	list     *widget.List
	clips    []store.Clip
	selected int
	search   *widget.Entry
	mu       sync.Mutex
	visible  bool

	onClose func()
}

func NewHistoryWindow(app fyne.App, s *store.Store, cfg *config.Config) *HistoryWindow {
	h := &HistoryWindow{
		app:   app,
		store: s,
		cfg:   cfg,
	}
	h.build()
	return h
}

func (h *HistoryWindow) build() {
	h.win = h.app.NewWindow("LinBoard — Clipboard History")
	h.win.SetFixedSize(true)
	h.win.Resize(fyne.NewSize(480, 420))
	h.win.SetCloseIntercept(func() {
		h.Hide()
	})

	h.search = widget.NewEntry()
	h.search.SetPlaceHolder("Search clipboard history…")
	h.search.OnChanged = func(_ string) {
		h.refreshList()
	}

	h.list = widget.NewList(
		func() int {
			h.mu.Lock()
			defer h.mu.Unlock()
			return len(h.clips)
		},
		func() fyne.CanvasObject {
			pinIcon := widget.NewIcon(theme.MediaRecordIcon())
			preview := widget.NewLabel("")
			preview.Wrapping = fyne.TextTruncate
			timeLabel := widget.NewLabel("")
			timeLabel.TextStyle = fyne.TextStyle{Italic: true}
			typeBadge := widget.NewLabel("")
			typeBadge.TextStyle = fyne.TextStyle{Bold: true}
			return container.NewBorder(
				nil, nil,
				container.NewHBox(pinIcon, typeBadge),
				timeLabel,
				preview,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			h.mu.Lock()
			if id < 0 || id >= len(h.clips) {
				h.mu.Unlock()
				return
			}
			clip := h.clips[id]
			h.mu.Unlock()
			// NewBorder order: center, left, right (see fyne container.NewBorder)
			border := obj.(*fyne.Container)
			preview := border.Objects[0].(*widget.Label)
			left := border.Objects[1].(*fyne.Container)
			timeLabel := border.Objects[2].(*widget.Label)
			pinIcon := left.Objects[0].(*widget.Icon)
			typeBadge := left.Objects[1].(*widget.Label)

			if clip.Pinned {
				pinIcon.SetResource(theme.MediaRecordIcon())
				pinIcon.Show()
			} else {
				pinIcon.Hide()
			}

			switch clip.ContentType {
			case store.TypeURL:
				typeBadge.SetText("URL")
			case store.TypeImage:
				typeBadge.SetText("IMG")
			default:
				typeBadge.SetText("TXT")
			}

			preview.SetText(clip.Preview)
			timeLabel.SetText(store.FormatTime(clip.CreatedAt))
		},
	)

	h.list.OnSelected = func(id widget.ListItemID) {
		h.mu.Lock()
		h.selected = int(id)
		h.mu.Unlock()
	}
	h.list.OnUnselected = func(_ widget.ListItemID) {}

	help := widget.NewLabel("↑↓ Navigate  •  Enter Paste  •  Del Remove  •  P Pin  •  Esc Close")
	help.TextStyle = fyne.TextStyle{Italic: true}
	help.Alignment = fyne.TextAlignCenter

	header := widget.NewLabelWithStyle("Clipboard History", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewBorder(
		container.NewVBox(
			header,
			h.search,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			help,
		),
		nil, nil,
		h.list,
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 35, A: 250})
	bg.CornerRadius = 8
	h.win.SetContent(container.NewStack(bg, container.NewPadded(content)))

	h.win.Canvas().SetOnTypedKey(h.handleKey)
	h.win.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyEscape,
		Modifier: 0,
	}, func(_ fyne.Shortcut) {
		h.Hide()
	})
}

func (h *HistoryWindow) handleKey(ev *fyne.KeyEvent) {
	// Let the search field handle printable text input.
	if h.win.Canvas().Focused() == h.search && ev.Name == fyne.KeyP {
		return
	}

	switch ev.Name {
	case fyne.KeyEscape:
		h.Hide()
	case fyne.KeyReturn, fyne.KeyEnter:
		h.pasteSelected()
	case fyne.KeyUp:
		h.mu.Lock()
		sel := h.selected
		h.mu.Unlock()
		if sel > 0 {
			h.list.Select(sel - 1)
		}
	case fyne.KeyDown:
		h.mu.Lock()
		sel := h.selected
		count := len(h.clips)
		h.mu.Unlock()
		if sel < count-1 {
			h.list.Select(sel + 1)
		}
	case fyne.KeyDelete:
		h.deleteSelected()
	case fyne.KeyP:
		h.pinSelected()
	}
}

func (h *HistoryWindow) RefreshIfVisible() {
	h.mu.Lock()
	visible := h.visible
	h.mu.Unlock()
	if visible {
		h.refreshList()
	}
}

func (h *HistoryWindow) refreshList() {
	search := strings.TrimSpace(h.search.Text)
	clips, err := h.store.List(search, 100)
	if err != nil {
		return
	}
	h.mu.Lock()
	h.clips = clips
	if h.selected >= len(h.clips) {
		h.selected = 0
	}
	h.mu.Unlock()
	h.list.Refresh()
	if len(h.clips) > 0 {
		h.list.Select(h.selected)
	}
}

func (h *HistoryWindow) pasteSelected() {
	h.mu.Lock()
	if h.selected < 0 || h.selected >= len(h.clips) {
		h.mu.Unlock()
		return
	}
	clip := h.clips[h.selected]
	h.mu.Unlock()

	h.Hide()
	if h.cfg.PasteOnSelect {
		_ = clipboard.PasteClip(&clip)
	} else {
		switch clip.ContentType {
		case store.TypeImage:
			data, err := os.ReadFile(clip.ImagePath)
			if err == nil {
				_ = clipboard.WriteImage(data)
			}
		default:
			_ = clipboard.WriteText(clip.Content)
		}
	}
}

func (h *HistoryWindow) deleteSelected() {
	h.mu.Lock()
	if h.selected < 0 || h.selected >= len(h.clips) {
		h.mu.Unlock()
		return
	}
	id := h.clips[h.selected].ID
	h.mu.Unlock()
	_ = h.store.Delete(id)
	h.refreshList()
}

func (h *HistoryWindow) pinSelected() {
	h.mu.Lock()
	if h.selected < 0 || h.selected >= len(h.clips) {
		h.mu.Unlock()
		return
	}
	id := h.clips[h.selected].ID
	h.mu.Unlock()
	_ = h.store.TogglePin(id)
	h.refreshList()
}

func (h *HistoryWindow) Toggle() {
	h.mu.Lock()
	visible := h.visible
	h.mu.Unlock()
	if visible {
		h.Hide()
	} else {
		h.Show()
	}
}

func (h *HistoryWindow) Show() {
	h.search.SetText("")
	h.refreshList()
	h.win.CenterOnScreen()
	h.win.Show()
	h.win.RequestFocus()
	h.mu.Lock()
	h.visible = true
	count := len(h.clips)
	h.mu.Unlock()
	if count > 0 {
		h.list.Select(0)
	}
}

func (h *HistoryWindow) Hide() {
	h.win.Hide()
	h.mu.Lock()
	h.visible = false
	h.mu.Unlock()
	if h.onClose != nil {
		h.onClose()
	}
}

func (h *HistoryWindow) OnClose(fn func()) {
	h.onClose = fn
}

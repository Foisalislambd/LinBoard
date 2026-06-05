package ui

import (
	"image/color"
	"log"
	"os"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/foisal/linboard/internal/assets"
	"github.com/foisal/linboard/internal/clipboard"
	"github.com/foisal/linboard/internal/config"
	"github.com/foisal/linboard/internal/store"
)

const (
	imageThumbW = 220
	imageThumbH = 72
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

// tapListRow makes list rows clickable (Windows-style: click item to paste).
type tapListRow struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onTap   func()
}

func newTapListRow(content fyne.CanvasObject) *tapListRow {
	r := &tapListRow{content: content}
	r.ExtendBaseWidget(r)
	return r
}

func (r *tapListRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

func (r *tapListRow) Tapped(*fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
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
	h.win.SetIcon(assets.Fyne())
	h.win.SetFixedSize(true)
	h.win.Resize(fyne.NewSize(520, 460))
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
			pinBtn := widget.NewButtonWithIcon("", theme.MediaRecordIcon(), nil)
			pinBtn.Importance = widget.LowImportance
			textPreview := widget.NewLabel("")
			textPreview.Wrapping = fyne.TextTruncate
			imgPreview := canvas.NewImageFromFile("")
			imgPreview.FillMode = canvas.ImageFillContain
			imgPreview.SetMinSize(fyne.NewSize(imageThumbW, imageThumbH))
			center := container.NewStack(textPreview, imgPreview)
			timeLabel := widget.NewLabel("")
			timeLabel.TextStyle = fyne.TextStyle{Italic: true}
			typeBadge := widget.NewLabel("")
			typeBadge.TextStyle = fyne.TextStyle{Bold: true}
			body := newTapListRow(container.NewBorder(
				nil, nil,
				typeBadge,
				timeLabel,
				center,
			))
			return container.NewHBox(pinBtn, body)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			rowBox := obj.(*fyne.Container)
			pinBtn := rowBox.Objects[0].(*widget.Button)
			row := rowBox.Objects[1].(*tapListRow)

			itemID := int(id)

			h.mu.Lock()
			if id < 0 || id >= len(h.clips) {
				h.mu.Unlock()
				return
			}
			clip := h.clips[id]
			clipID := clip.ID
			h.mu.Unlock()

			pinBtn.OnTapped = func() {
				h.togglePinByID(clipID)
			}
			row.onTap = func() {
				h.mu.Lock()
				h.selected = itemID
				h.mu.Unlock()
				h.pasteClipByID(clipID)
			}

			border := row.content.(*fyne.Container)
			center := border.Objects[0].(*fyne.Container)
			textPreview := center.Objects[0].(*widget.Label)
			imgPreview := center.Objects[1].(*canvas.Image)
			typeBadge := border.Objects[1].(*widget.Label)
			timeLabel := border.Objects[2].(*widget.Label)

			if clip.Pinned {
				pinBtn.SetIcon(theme.MediaRecordIcon())
				pinBtn.Importance = widget.HighImportance
			} else {
				pinBtn.SetIcon(theme.RadioButtonIcon())
				pinBtn.Importance = widget.LowImportance
			}

			switch clip.ContentType {
			case store.TypeURL:
				typeBadge.SetText("URL")
			case store.TypeImage:
				typeBadge.SetText("IMG")
			default:
				typeBadge.SetText("TXT")
			}

			showImage := clip.ContentType == store.TypeImage && clip.ImagePath != ""
			if showImage {
				if _, err := os.Stat(clip.ImagePath); err != nil {
					showImage = false
				}
			}
			if showImage {
				textPreview.Hide()
				imgPreview.File = clip.ImagePath
				imgPreview.Show()
				imgPreview.Refresh()
			} else {
				imgPreview.Hide()
				imgPreview.File = ""
				textPreview.SetText(clip.Preview)
				textPreview.Show()
			}
			timeLabel.SetText(store.FormatTime(clip.CreatedAt))
		},
	)

	h.list.OnSelected = func(id widget.ListItemID) {
		h.mu.Lock()
		h.selected = int(id)
		h.mu.Unlock()
	}
	h.list.OnUnselected = func(_ widget.ListItemID) {}

	help := widget.NewLabel("Click Paste  •  📌 Pin icon  •  ↑↓ Navigate  •  Enter  •  Del  •  Esc Close")
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
	if h.win.Canvas().Focused() == h.search {
		if ev.Name == fyne.KeyEscape {
			h.Hide()
		}
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
	clipID := h.clips[h.selected].ID
	h.mu.Unlock()
	h.pasteClipByID(clipID)
}

func (h *HistoryWindow) pasteClipByID(id int64) {
	clip, err := h.store.GetByID(id)
	if err != nil || clip == nil {
		return
	}
	h.pasteClip(clip)
}

func (h *HistoryWindow) pasteClip(clip *store.Clip) {
	pasteOnSelect := h.cfg.PasteOnSelect
	h.Hide()

	if pasteOnSelect {
		clipCopy := *clip
		go func() {
			if err := clipboard.PasteClip(&clipCopy); err != nil {
				log.Printf("paste failed: %v", err)
			}
		}()
		return
	}

	if err := clipboard.CopyClipToClipboard(clip); err != nil {
		log.Printf("copy to clipboard failed: %v", err)
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
	sel := h.selected
	h.mu.Unlock()
	h.togglePinAt(sel)
}

func (h *HistoryWindow) togglePinAt(index int) {
	h.mu.Lock()
	if index < 0 || index >= len(h.clips) {
		h.mu.Unlock()
		return
	}
	id := h.clips[index].ID
	h.mu.Unlock()
	h.togglePinByID(id)
}

func (h *HistoryWindow) togglePinByID(id int64) {
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
	// Remember where the user was typing before we take focus (for auto-paste).
	clipboard.RememberPasteTarget()

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

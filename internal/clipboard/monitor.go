package clipboard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"golang.design/x/clipboard"

	"github.com/foisal/linboard/internal/store"
)

type Monitor struct {
	store    *store.Store
	mu       sync.Mutex
	lastText string
	lastImg  string
	onChange func()
}

// watchSuppress prevents re-recording clipboard writes we perform ourselves.
var watchSuppress atomic.Int32

func NewMonitor(s *store.Store) *Monitor {
	return &Monitor{store: s}
}

func (m *Monitor) OnChange(fn func()) {
	m.onChange = fn
}

func (m *Monitor) Start(ctx context.Context) {
	if err := clipboard.Init(); err != nil {
		log.Printf("clipboard init failed: %v", err)
		return
	}

	go m.watchText(ctx)
	go m.watchImage(ctx)
}

func (m *Monitor) watchText(ctx context.Context) {
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		text := string(data)
		m.mu.Lock()
		if text == m.lastText {
			m.mu.Unlock()
			continue
		}
		m.lastText = text
		m.mu.Unlock()

		if watchSuppress.Load() > 0 {
			continue
		}

		if _, err := m.store.AddText(text); err != nil {
			log.Printf("save text clip: %v", err)
			continue
		}
		m.notify()
	}
}

func (m *Monitor) watchImage(ctx context.Context) {
	ch := clipboard.Watch(ctx, clipboard.FmtImage)
	for data := range ch {
		if len(data) == 0 {
			continue
		}
		key := imageFingerprint(data)
		m.mu.Lock()
		if key == m.lastImg {
			m.mu.Unlock()
			continue
		}
		m.lastImg = key
		m.mu.Unlock()

		if watchSuppress.Load() > 0 {
			continue
		}

		if _, err := m.store.AddImage(data); err != nil {
			log.Printf("save image clip: %v", err)
			continue
		}
		m.notify()
	}
}

func (m *Monitor) notify() {
	if m.onChange != nil {
		m.onChange()
	}
}

// WriteText puts text on the system clipboard and waits until the write completes.
func WriteText(text string) error {
	return writeClipboard(clipboard.FmtText, []byte(text))
}

// WriteImage puts PNG image bytes on the system clipboard and waits until done.
func WriteImage(data []byte) error {
	return writeClipboard(clipboard.FmtImage, data)
}

func writeClipboard(format clipboard.Format, data []byte) error {
	if err := clipboard.Init(); err != nil {
		return err
	}
	watchSuppress.Add(1)
	defer watchSuppress.Add(-1)
	done := clipboard.Write(format, data)
	if done != nil {
		<-done
	}
	return nil
}

// PasteClip copies content to clipboard and simulates Ctrl+V in the previously focused window.
func PasteClip(clip *store.Clip) error {
	if err := CopyClipToClipboard(clip); err != nil {
		return err
	}
	return simulatePasteWithRetry()
}

// CopyClipToClipboard writes a history item to the system clipboard without simulating paste.
func CopyClipToClipboard(clip *store.Clip) error {
	switch clip.ContentType {
	case store.TypeImage:
		data, err := os.ReadFile(clip.ImagePath)
		if err != nil {
			return err
		}
		return WriteImage(data)
	default:
		return WriteText(clip.Content)
	}
}

func imageFingerprint(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

package clipboard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/foisal/linboard/internal/platform"
	"github.com/foisal/linboard/internal/store"
)

const (
	textPollInterval  = 400 * time.Millisecond
	imagePollInterval = 1500 * time.Millisecond
)

type Monitor struct {
	store    *store.Store
	mu       sync.Mutex
	lastText string
	lastImg  string
	onChange func()
}

var activeMonitor *Monitor
var watchSuppress atomic.Int32

func NewMonitor(s *store.Store) *Monitor {
	return &Monitor{store: s}
}

func (m *Monitor) OnChange(fn func()) {
	m.onChange = fn
}

func (m *Monitor) Start(ctx context.Context) {
	if ReadToolName() == "none" {
		log.Printf("clipboard monitor: no read backend available")
		return
	}

	activeMonitor = m
	go m.pollText(ctx)
	if haveCommand("xclip") {
		go m.pollImage(ctx)
	} else if platform.UsePortalHotkey() {
		log.Printf("clipboard monitor: image history needs xclip (XWayland); text via Fyne core")
	}

}

func (m *Monitor) pollText(ctx context.Context) {
	ticker := time.NewTicker(textPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if watchSuppress.Load() > 0 {
				continue
			}

			text, err := readText()
			if err != nil || text == "" {
				continue
			}

			m.mu.Lock()
			if text == m.lastText {
				m.mu.Unlock()
				continue
			}
			m.lastText = text
			m.mu.Unlock()

			if _, err := m.store.AddText(text); err != nil {
				log.Printf("save text clip: %v", err)
				continue
			}
			m.notify()
		}
	}
}

func (m *Monitor) pollImage(ctx context.Context) {
	ticker := time.NewTicker(imagePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if watchSuppress.Load() > 0 {
				continue
			}

			data, err := readImage()
			if err != nil || len(data) == 0 {
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

			if _, err := m.store.AddImage(data); err != nil {
				log.Printf("save image clip: %v", err)
				continue
			}
			m.notify()
		}
	}
}

func (m *Monitor) notify() {
	if m.onChange != nil {
		m.onChange()
	}
}

// CopyClip puts clip content on the system clipboard without pasting.
func CopyClip(clip *store.Clip) error {
	watchSuppress.Add(1)
	defer watchSuppress.Add(-1)

	switch clip.ContentType {
	case store.TypeImage:
		data, err := os.ReadFile(clip.ImagePath)
		if err != nil {
			return err
		}
		if err := WriteImage(data); err != nil {
			return err
		}
		noteImageSeen(data)
		return nil
	default:
		if err := WriteText(clip.Content); err != nil {
			return err
		}
		noteTextSeen(clip.Content)
		return nil
	}
}

func noteTextSeen(text string) {
	if activeMonitor == nil {
		return
	}
	activeMonitor.mu.Lock()
	activeMonitor.lastText = text
	activeMonitor.mu.Unlock()
}

func noteImageSeen(data []byte) {
	if activeMonitor == nil {
		return
	}
	activeMonitor.mu.Lock()
	activeMonitor.lastImg = imageFingerprint(data)
	activeMonitor.mu.Unlock()
}

func imageFingerprint(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

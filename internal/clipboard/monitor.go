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

	"github.com/foisal/linboard/internal/store"
)

const pollInterval = 400 * time.Millisecond

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
	if ReadToolName() == "none" {
		log.Printf("clipboard monitor: no read tool (install wl-clipboard or xclip)")
		return
	}

	go m.pollText(ctx)
	go m.pollImage(ctx)
}

func (m *Monitor) pollText(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
}

func (m *Monitor) pollImage(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
		return WriteImage(data)
	default:
		return WriteText(clip.Content)
	}
}

// PasteClip copies content to clipboard and simulates Ctrl+V in the previously focused window.
func PasteClip(clip *store.Clip) error {
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
	default:
		if err := WriteText(clip.Content); err != nil {
			return err
		}
	}

	// Brief pause so focus returns to the previous window after our popup closes.
	time.Sleep(120 * time.Millisecond)
	return simulatePaste()
}

func imageFingerprint(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

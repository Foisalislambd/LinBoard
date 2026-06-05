package clipboard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

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
	if err := clipboard.Init(); err != nil {
		return err
	}
	done := clipboard.Write(clipboard.FmtText, []byte(text))
	if done != nil {
		<-done
	}
	return nil
}

// WriteImage puts PNG image bytes on the system clipboard and waits until done.
func WriteImage(data []byte) error {
	if err := clipboard.Init(); err != nil {
		return err
	}
	done := clipboard.Write(clipboard.FmtImage, data)
	if done != nil {
		<-done
	}
	return nil
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

func simulatePaste() error {
	// xdotool works on X11; on Wayland clipboard is still updated even if paste fails.
	cmd := exec.Command("xdotool", "key", "ctrl+v")
	if err := cmd.Run(); err != nil {
		log.Printf("auto-paste (xdotool) failed: %v — content is on clipboard", err)
	}
	return nil
}

func imageFingerprint(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:16])
}

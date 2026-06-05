package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/foisal/linboard/internal/config"
	_ "modernc.org/sqlite"
)

type ContentType string

const (
	TypeText  ContentType = "text"
	TypeImage ContentType = "image"
	TypeURL   ContentType = "url"
)

type Clip struct {
	ID          int64       `json:"id"`
	Content     string      `json:"content"`
	ContentType ContentType `json:"content_type"`
	ImagePath   string      `json:"image_path,omitempty"`
	Preview     string      `json:"preview"`
	Pinned      bool        `json:"pinned"`
	CreatedAt   time.Time   `json:"created_at"`
	Hash        string      `json:"hash"`
}

type Store struct {
	db        *sql.DB
	maxItems  int
	imagesDir string
}

func Open(maxItems int) (*Store, error) {
	dataDir, err := config.DataDir()
	if err != nil {
		return nil, err
	}
	imagesDir, err := config.ImagesDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "history.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, maxItems: maxItems, imagesDir: imagesDir}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS clips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL DEFAULT '',
			content_type TEXT NOT NULL DEFAULT 'text',
			image_path TEXT NOT NULL DEFAULT '',
			preview TEXT NOT NULL DEFAULT '',
			pinned INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			hash TEXT NOT NULL UNIQUE
		);
		CREATE INDEX IF NOT EXISTS idx_clips_created ON clips(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_clips_pinned ON clips(pinned DESC, created_at DESC);
	`)
	return err
}

func hashContent(content string, contentType ContentType) string {
	h := sha256.Sum256([]byte(string(contentType) + ":" + content))
	return hex.EncodeToString(h[:])
}

func hashBytes(data []byte, contentType ContentType) string {
	h := sha256.New()
	h.Write([]byte(string(contentType) + ":"))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func makePreview(content string, contentType ContentType) string {
	switch contentType {
	case TypeImage:
		return "[Image]"
	case TypeURL:
		if len(content) > 120 {
			return content[:120] + "…"
		}
		return content
	default:
		lines := strings.Split(strings.TrimSpace(content), "\n")
		first := lines[0]
		if len(first) > 120 {
			first = first[:120] + "…"
		}
		if len(lines) > 1 {
			first += fmt.Sprintf(" (+%d lines)", len(lines)-1)
		}
		return first
	}
}

func detectType(content string) ContentType {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		if !strings.Contains(trimmed, " ") && len(trimmed) < 2048 {
			return TypeURL
		}
	}
	return TypeText
}

func (s *Store) AddText(content string) (*Clip, error) {
	content = strings.TrimRight(content, "\x00")
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}
	ct := detectType(content)
	hash := hashContent(content, ct)
	now := time.Now().Unix()

	res, err := s.db.Exec(`
		INSERT INTO clips (content, content_type, preview, created_at, hash)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET created_at = excluded.created_at
	`, content, ct, makePreview(content, ct), now, hash)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if id == 0 {
		row := s.db.QueryRow(`SELECT id FROM clips WHERE hash = ?`, hash)
		_ = row.Scan(&id)
	}

	s.prune()
	return s.GetByID(id)
}

func (s *Store) AddImage(data []byte) (*Clip, error) {
	if len(data) == 0 {
		return nil, nil
	}
	hash := hashBytes(data, TypeImage)
	now := time.Now().Unix()

	filename := hash[:16] + ".png"
	path := filepath.Join(s.imagesDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, err
		}
	}

	preview := fmt.Sprintf("[Image · %d KB]", len(data)/1024)
	res, err := s.db.Exec(`
		INSERT INTO clips (content, content_type, image_path, preview, created_at, hash)
		VALUES ('', 'image', ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET created_at = excluded.created_at
	`, path, preview, now, hash)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	if id == 0 {
		row := s.db.QueryRow(`SELECT id FROM clips WHERE hash = ?`, hash)
		_ = row.Scan(&id)
	}

	s.prune()
	return s.GetByID(id)
}

func (s *Store) prune() {
	_, _ = s.db.Exec(`
		DELETE FROM clips WHERE id NOT IN (
			SELECT id FROM clips
			ORDER BY pinned DESC, created_at DESC
			LIMIT ?
		) AND pinned = 0
	`, s.maxItems)
}

func (s *Store) List(search string, limit int) ([]Clip, error) {
	if limit <= 0 {
		limit = 50
	}
	search = strings.TrimSpace(search)
	var rows *sql.Rows
	var err error
	if search == "" {
		rows, err = s.db.Query(`
			SELECT id, content, content_type, image_path, preview, pinned, created_at, hash
			FROM clips ORDER BY pinned DESC, created_at DESC LIMIT ?
		`, limit)
	} else {
		pattern := "%" + search + "%"
		rows, err = s.db.Query(`
			SELECT id, content, content_type, image_path, preview, pinned, created_at, hash
			FROM clips
			WHERE preview LIKE ? OR content LIKE ?
			ORDER BY pinned DESC, created_at DESC LIMIT ?
		`, pattern, pattern, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clips []Clip
	for rows.Next() {
		c, err := scanClip(rows)
		if err != nil {
			return nil, err
		}
		clips = append(clips, c)
	}
	return clips, rows.Err()
}

func scanClip(scanner interface {
	Scan(dest ...any) error
}) (Clip, error) {
	var c Clip
	var ct string
	var pinned int
	var created int64
	if err := scanner.Scan(&c.ID, &c.Content, &ct, &c.ImagePath, &c.Preview, &pinned, &created, &c.Hash); err != nil {
		return Clip{}, err
	}
	c.ContentType = ContentType(ct)
	c.Pinned = pinned == 1
	c.CreatedAt = time.Unix(created, 0)
	return c, nil
}

func (s *Store) GetByID(id int64) (*Clip, error) {
	row := s.db.QueryRow(`
		SELECT id, content, content_type, image_path, preview, pinned, created_at, hash
		FROM clips WHERE id = ?
	`, id)
	c, err := scanClip(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) TogglePin(id int64) error {
	_, err := s.db.Exec(`
		UPDATE clips SET pinned = CASE WHEN pinned = 1 THEN 0 ELSE 1 END WHERE id = ?
	`, id)
	return err
}

func (s *Store) Delete(id int64) error {
	var imagePath string
	_ = s.db.QueryRow(`SELECT image_path FROM clips WHERE id = ?`, id).Scan(&imagePath)
	_, err := s.db.Exec(`DELETE FROM clips WHERE id = ?`, id)
	if err == nil && imagePath != "" {
		_ = os.Remove(imagePath)
	}
	return err
}

func (s *Store) ClearUnpinned() error {
	rows, err := s.db.Query(`SELECT image_path FROM clips WHERE pinned = 0 AND image_path != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil && p != "" {
			_ = os.Remove(p)
		}
	}
	_, err = s.db.Exec(`DELETE FROM clips WHERE pinned = 0`)
	return err
}

func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM clips`).Scan(&n)
	return n, err
}

func FormatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	default:
		return t.Format("Jan 2, 2006")
	}
}

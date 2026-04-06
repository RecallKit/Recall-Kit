// Package session handles persistence of named chat sessions.
// Sessions are stored as JSON files under ~/.recallkit/sessions/.
// Each file is a single Session struct — no database required at this stage.
// The context graph layer will later index these files into Kùzu.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/RecallKit/recallkit/internal/engine"
)

// NewStore returns a Store rooted at the default RecallKit data directory.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("session store: home dir: %w", err)
	}
	dir := filepath.Join(home, ".recallkit", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("session store: mkdir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

// Create initialises a new empty session with the given name and model.
func (s *Store) Create(name, model string) (*Session, error) {
	now := time.Now()
	sess := &Session{
		ID:        makeID(name, now),
		Name:      name,
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []engine.Message{},
	}
	if err := s.save(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// Load reads a session by ID from disk.
func (s *Store) Load(id string) (*Session, error) {
	path := s.path(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %q not found", id)
		}
		return nil, fmt.Errorf("load session: %w", err)
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	return &sess, nil
}

// Save persists the current state of a session to disk.
func (s *Store) Save(sess *Session) error {
	sess.UpdatedAt = time.Now()
	return s.save(sess)
}

// Delete removes a session file from disk.
func (s *Store) Delete(id string) error {
	if err := os.Remove(s.path(id)); err != nil {
		return fmt.Errorf("delete session %q: %w", id, err)
	}
	return nil
}

// List returns all sessions sorted by UpdatedAt descending (most recent first).
func (s *Store) List() ([]*Session, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	var sessions []*Session
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		sess, err := s.Load(id)
		if err != nil {
			continue // skip corrupt files
		}
		sessions = append(sessions, sess)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})
	return sessions, nil
}

// ── Message helpers ───────────────────────────────────────────────────────────

// AppendMessage adds a message to the session and saves it.
func (s *Store) AppendMessage(sess *Session, role, content string) error {
	sess.Messages = append(sess.Messages, engine.Message{
		Role:    role,
		Content: content,
	})
	return s.Save(sess)
}

// ── Internal ──────────────────────────────────────────────────────────────────

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

func (s *Store) save(sess *Session) error {
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	if err := os.WriteFile(s.path(sess.ID), data, 0o644); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}

// makeID produces a filesystem-safe session ID from a name and timestamp.
func makeID(name string, t time.Time) string {
	slug := strings.ToLower(name)
	slug = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, slug)
	// Collapse multiple dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "session"
	}
	return fmt.Sprintf("%s-%s", slug, t.Format("20060102-150405"))
}

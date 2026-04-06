// Package session handles persistence of named chat sessions.
// Sessions are stored as JSON files under ~/.recallkit/sessions/.
// Each file is a single Session struct — no database required at this stage.
// The context graph layer will later index these files into Kùzu.
package session

import (
	"time"

	"github.com/RecallKit/recallkit/internal/engine"
)

// Session is one complete chat session — a named sequence of turns.
type Session struct {
	ID        string           `json:"id"` // slugified name + timestamp
	Name      string           `json:"name"`
	Model     string           `json:"model"` // ollama model used
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Messages  []engine.Message `json:"messages"` // full history
}

// Store manages session files on disk.
type Store struct {
	dir string // ~/.recallkit/sessions
}

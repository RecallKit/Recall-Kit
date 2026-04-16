package tui

import (
	"github.com/RecallKit/recallkit/internal/engine"
	"github.com/RecallKit/recallkit/internal/session"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type status int

const (
	statusIdle status = iota
	statusConnecting
	statusThinking
	statusStreaming
	statusError
)

// ChatMessage is one rendered bubble in the history pane.
type ChatMessage struct {
	Role    string
	Content string
}

// Model is the root Bubble Tea state.
type Model struct {
	// session
	sess  *session.Session
	store *session.Store

	// engine
	client *engine.OllamaClient

	// conversation
	messages  []ChatMessage // rendered display history
	streamBuf string

	// ui
	viewport viewport.Model
	input    textarea.Model
	status   status
	err      error
	width    int
	height   int
	ready    bool
}

var (
	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69"))

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78"))

	dimStyle = lipgloss.NewStyle().
			Faint(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255"))

	inputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238")).
				Padding(0, 1)
)

// NewModel builds the TUI model from an existing or new session.
func NewModel(sess *session.Session, store *session.Store, client *engine.OllamaClient) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message and press Enter…"
	ta.Focus()
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.CharLimit = 4000
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

	// Populate rendered messages from session history so resumed
	// sessions show previous turns immediately on load.
	var messages []ChatMessage
	for _, m := range sess.Messages {
		messages = append(messages, ChatMessage{Role: m.Role, Content: m.Content})
	}

	return Model{
		sess:     sess,
		store:    store,
		client:   client,
		messages: messages,
		input:    ta,
		status:   statusConnecting,
	}
}

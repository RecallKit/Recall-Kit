package tui

import (
	"github.com/RecallKit/recallkit/internal/engine"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type status int //tracks what the TUI is currently doing

const (
	statusIdle       status = iota
	statusConnecting        // pinging Ollama on startup
	statusThinking          // waiting for first token
	statusStreaming         // tokens arriving
	statusError             // unrecoverable error shown to user
)

// ChatMessage is one rendered bubble in the history pane.
type ChatMessage struct {
	Role    string // "user" | "assistant"
	Content string
}

// Model is the root Bubble Tea state — everything the view needs lives here.
type Model struct {
	// config
	ollamaModel string
	client      *engine.OllamaClient

	// conversation
	history   []engine.Message
	messages  []ChatMessage
	streamBuf string

	// ui components
	viewport viewport.Model
	input    textarea.Model

	// state
	status status
	err    error
	width  int
	height int
	ready  bool
}

// styles used across update + view
var (
	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")) // soft blue

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")) // soft green

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

// NewModel builds the initial TUI model.
func NewModel(ollamaModel string, client *engine.OllamaClient) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message and press Enter…"
	ta.Focus()
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.CharLimit = 4000

	// Remove the default textarea border — we wrap it ourselves
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

	return Model{
		ollamaModel: ollamaModel,
		client:      client,
		input:       ta,
		status:      statusConnecting,
	}
}

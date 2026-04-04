package tui

import (
	"strings"

	"github.com/RecallKit/recallkit/internal/engine"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Init bootstraps the TUI — blink the cursor and ping Ollama right away.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		CmdPing(m.client),
	)
}

// Update is the central event dispatcher.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd    tea.Cmd
		inputCmd tea.Cmd
	)

	switch msg := msg.(type) {

	// ── Layout ───────────────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(msg.Width - 6)

		headerH := 3
		inputH := m.input.Height() + 2 // +2 for border
		footerH := 1
		vpH := msg.Height - headerH - inputH - footerH
		if vpH < 1 {
			vpH = 1
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width, vpH)
			m.viewport.SetContent(m.renderHistory())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = vpH
		}

	// ── Ollama startup ping ──────────────────────────────────────────────────
	case PingResultMsg:
		if msg.Err != nil {
			m.status = statusError
			m.err = msg.Err
			return m, nil
		}
		m.status = statusIdle
		m.messages = append(m.messages, ChatMessage{
			Role:    "assistant",
			Content: "Connected to Ollama · model: " + m.ollamaModel + "\nHow can I help you today?",
		})
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()

	// ── Keyboard ─────────────────────────────────────────────────────────────
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			// Alt+Enter = newline inside the textarea
			if msg.Alt {
				break
			}
			if m.status != statusIdle {
				return m, nil
			}
			userText := strings.TrimSpace(m.input.Value())
			if userText == "" {
				return m, nil
			}

			m.messages = append(m.messages, ChatMessage{Role: "user", Content: userText})
			m.history = append(m.history, engine.Message{Role: "user", Content: userText})
			m.input.Reset()
			m.status = statusThinking
			m.viewport.SetContent(m.renderHistory())
			m.viewport.GotoBottom()

			return m, CmdStartStream(m.client, m.ollamaModel, m.history)
		}

	// ── Stream: channels ready, start pulling ────────────────────────────────
	case nextTokenFnMsg:
		m.status = statusStreaming
		return m, tea.Cmd(msg) // schedule the first pull

	// ── Stream: one token arrived ────────────────────────────────────────────
	case TokenMsg:
		m.streamBuf += msg.Token
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()
		return m, msg.NextPull // schedule next pull immediately

	// ── Stream: finished cleanly ─────────────────────────────────────────────
	case StreamDoneMsg:
		if m.streamBuf != "" {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: m.streamBuf,
			})
			m.history = append(m.history, engine.Message{
				Role:    "assistant",
				Content: m.streamBuf,
			})
			m.streamBuf = ""
		}
		m.status = statusIdle
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()

	// ── Stream: error ─────────────────────────────────────────────────────────
	case StreamErrMsg:
		m.streamBuf = ""
		m.err = msg.Err
		m.status = statusError
		m.viewport.SetContent(m.renderHistory())
	}

	m.viewport, vpCmd = m.viewport.Update(msg)
	m.input, inputCmd = m.input.Update(msg)
	return m, tea.Batch(vpCmd, inputCmd)
}

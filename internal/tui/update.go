package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Init bootstraps the TUI — blink the cursor and ping Ollama right away.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		cmdPing(m.client),
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

		// For resumed sessions show a different greeting
		welcomeMsg := "Connected to Ollama · model: " + m.sess.Model + "\nHow can I help you today?"
		if len(m.sess.Messages) > 0 {
			welcomeMsg = "Resumed session: " + m.sess.Name + " · model: " + m.sess.Model
		}
		m.messages = append(m.messages, ChatMessage{
			Role:    "assistant",
			Content: welcomeMsg,
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
			_ = m.store.AppendMessage(m.sess, "user", userText)
			m.input.Reset()
			m.status = statusThinking
			m.viewport.SetContent(m.renderHistory())
			m.viewport.GotoBottom()

			return m, cmdStartStream(m.client, m.sess.Model, m.sess.Messages)
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
			_ = m.store.AppendMessage(m.sess, "assistant", m.streamBuf)
			m.streamBuf = ""
		}
		m.status = statusIdle
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()

	// ── Stream: error ────────────────────────────────────────────────────────
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

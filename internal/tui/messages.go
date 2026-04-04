package tui

import tea "github.com/charmbracelet/bubbletea"

// PingResultMsg is sent after the startup Ollama health check.
type PingResultMsg struct{ Err error }

// nextTokenFnMsg wraps a tea.Cmd so it can travel as a Msg through the
// Bubble Tea bus. The update loop schedules it as the next command,
// implementing a non-blocking pull loop for streaming tokens.
type nextTokenFnMsg tea.Cmd

// TokenMsg carries one streamed token plus a Cmd to pull the next one.
type TokenMsg struct {
	Token    string
	NextPull tea.Cmd
}

// StreamDoneMsg signals the stream has ended cleanly.
type StreamDoneMsg struct{}

// StreamErrMsg carries a streaming error.
type StreamErrMsg struct{ Err error }

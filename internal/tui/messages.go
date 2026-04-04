package tui

import tea "github.com/charmbracelet/bubbletea"

// pingResultMsg is sent after the startup Ollama health check.
type pingResultMsg struct{ err error }

// nextTokenFnMsg wraps a tea.Cmd so it can travel as a Msg through the
// Bubble Tea bus. The update loop schedules it as the next command,
// implementing a non-blocking pull loop for streaming tokens.
type nextTokenFnMsg tea.Cmd

// tokenMsg carries one streamed token plus a Cmd to pull the next one.
type tokenMsg struct {
	token    string
	nextPull tea.Cmd
}

// streamDoneMsg signals the stream has ended cleanly.
type streamDoneMsg struct{}

// streamErrMsg carries a streaming error.
type streamErrMsg struct{ err error }

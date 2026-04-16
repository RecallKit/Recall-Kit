package tui

import (
	"github.com/RecallKit/recallkit/internal/engine"
	"github.com/RecallKit/recallkit/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the Bubble Tea program. It blocks until the user quits.
func Start(sess *session.Session, store *session.Store) error {
	client := engine.NewOllamaClient()
	m := NewModel(sess, store, client)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // use the alternate terminal buffer
		tea.WithMouseCellMotion(), // enable mouse scroll in the viewport
	)

	_, err := p.Run()
	return err
}

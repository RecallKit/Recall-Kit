// Package tui_tests — tests for program.go (the Start function).
// Start() blocks until the user quits and relies on a real terminal.
// These tests cover what can be verified without a live terminal:
//   - Start() signature and that it returns an error type
//   - The Model correctly implements tea.Model (compile-time check)
package tui_tests

import (
	"testing"

	"github.com/RecallKit/recallkit/internal/session"
	"github.com/RecallKit/recallkit/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

// TestModel_ImplementsTeaModel verifies at compile time that tui.Model satisfies
// the tea.Model interface required by tea.NewProgram.
func TestModel_ImplementsTeaModel(t *testing.T) {
	m := newTestModel(t)
	var _ tea.Model = m // fails at compile time if the interface is not satisfied
}

// TestModel_Init_ReturnsCmd verifies that Init() returns a non-nil Cmd.
// Init should at minimum return the textarea.Blink command.
func TestModel_Init_ReturnsCmd(t *testing.T) {
	m := newTestModel(t)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() must return a non-nil tea.Cmd (at minimum textarea.Blink)")
	}
}

// ---------------------------------------------------------------------------
// Model construction variants
// ---------------------------------------------------------------------------

func TestNewModel_EmptyModelName(t *testing.T) {
	m := newTestModelWithModel(t, "")
	// Must construct without panic even with empty model name
	var _ tea.Model = m
	_ = m.View()
}

func TestNewModel_LongModelName(t *testing.T) {
	longName := "very-long-model-name-that-exceeds-normal-display-width-xxxxxxxxxx"
	m := newTestModelWithModel(t, longName)
	var _ tea.Model = m
	_ = m.View()
}

// ---------------------------------------------------------------------------
// Start() — non-blocking checks
// ---------------------------------------------------------------------------

// TestStart_Signature verifies that the Start() function in program.go has the
// expected signature: func Start(sess *session.Session, store *session.Store) error.
// This is a compile-time check; if the signature changes, this test file will not compile.
func TestStart_Signature(t *testing.T) {
	// We can't call Start() in a unit test without a real terminal.
	// Instead we verify the function value has the correct type.
	var fn func(*session.Session, *session.Store) error = tui.Start
	if fn == nil {
		t.Error("tui.Start must be a non-nil function")
	}
}

// ---------------------------------------------------------------------------
// Init command execution
// ---------------------------------------------------------------------------

func TestModel_Init_ExecutableWithoutPanic(t *testing.T) {
	// Init() returns a tea.Batch of commands. We invoke it and verify
	// it doesn't panic even without a running Ollama instance.
	m := newTestModelWithModel(t, "phi3")
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil")
	}
	// NOTE: executing cmd() would trigger a network ping. We only verify
	// that the Cmd value is callable (not nil).
}

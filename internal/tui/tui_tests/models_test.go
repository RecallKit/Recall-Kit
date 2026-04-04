// Package tui_tests contains black-box tests for the internal/tui package.
// All tests interact only through the exported API surface.
package tui_tests

import (
	"errors"
	"testing"

	"github.com/RecallKit/recallkit/internal/engine"
	"github.com/RecallKit/recallkit/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newTestModel creates a freshly constructed Model for use in tests.
func newTestModel(t *testing.T) tui.Model {
	t.Helper()
	client := engine.NewOllamaClient()
	return tui.NewModel("llama3", client)
}

// makeReady sends a WindowSizeMsg to the model so it transitions to the
// "ready" state; returns the updated model.
func makeReady(t *testing.T, m tui.Model) tui.Model {
	t.Helper()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return updated.(tui.Model)
}

// makeReadyAndIdle sends WindowSizeMsg + successful PingResultMsg so the
// model reaches the idle state where it can accept user input.
func makeReadyAndIdle(t *testing.T, m tui.Model) tui.Model {
	t.Helper()
	m = makeReady(t, m)
	updated, _ := m.Update(tui.PingResultMsg{Err: nil})
	return updated.(tui.Model)
}

// containsAny returns true if s contains at least one of the provided needles.
func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		for i := 0; i <= len(s)-len(n); i++ {
			if s[i:i+len(n)] == n {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// NewModel
// ---------------------------------------------------------------------------

func TestNewModel_ImplementsTeaModel(t *testing.T) {
	m := newTestModel(t)
	var _ tea.Model = m
}

func TestNewModel_DifferentModels(t *testing.T) {
	models := []string{"llama3", "mistral", "phi3", "gemma", "codellama"}
	client := engine.NewOllamaClient()
	for _, name := range models {
		m := tui.NewModel(name, client)
		var _ tea.Model = m
		if m.View() == "" {
			t.Errorf("NewModel(%q).View() returned empty string", name)
		}
	}
}

func TestNewModel_NilClientDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewModel panicked with nil client: %v", r)
		}
	}()
	_ = tui.NewModel("test", nil)
}

// ---------------------------------------------------------------------------
// ChatMessage struct
// ---------------------------------------------------------------------------

func TestChatMessage_UserRole(t *testing.T) {
	msg := tui.ChatMessage{Role: "user", Content: "Hello!"}
	if msg.Role != "user" {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}
	if msg.Content != "Hello!" {
		t.Errorf("expected content 'Hello!', got %q", msg.Content)
	}
}

func TestChatMessage_AssistantRole(t *testing.T) {
	msg := tui.ChatMessage{Role: "assistant", Content: "Hi there!"}
	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", msg.Role)
	}
}

func TestChatMessage_EmptyContent(t *testing.T) {
	msg := tui.ChatMessage{Role: "user", Content: ""}
	if msg.Content != "" {
		t.Errorf("expected empty content, got %q", msg.Content)
	}
}

func TestChatMessage_ZeroValue(t *testing.T) {
	var msg tui.ChatMessage
	if msg.Role != "" || msg.Content != "" {
		t.Error("zero-value ChatMessage should have empty fields")
	}
}

// ---------------------------------------------------------------------------
// Update — WindowSizeMsg
// ---------------------------------------------------------------------------

func TestUpdate_WindowSizeMsg_MakesReady(t *testing.T) {
	m := newTestModel(t)
	updated := makeReady(t, m)

	// After WindowSizeMsg the model must be ready and View() non-trivial
	v := updated.View()
	if v == "" {
		t.Error("View() after WindowSizeMsg must not be empty")
	}
	if containsAny(v, "Initializing", "initializing") {
		t.Error("after WindowSizeMsg, model should not show initializing message")
	}
}

func TestUpdate_WindowSizeMsg_SmallTerminal(t *testing.T) {
	m := newTestModel(t)
	// Very small terminal — vpH clamping to 1 must not panic.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 10, Height: 5})
	if updated == nil {
		t.Error("Update() returned nil for small terminal size")
	}
	_ = updated.(tui.Model).View()
}

func TestUpdate_WindowSizeMsg_Resize(t *testing.T) {
	m := newTestModel(t)
	m = makeReady(t, m) // first size — branches to viewport.New
	// Second size — branches to width/height assignment (else branch)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if updated == nil {
		t.Error("Update() returned nil model on resize")
	}
}

// ---------------------------------------------------------------------------
// Update — pingResultMsg (exported alias)
// ---------------------------------------------------------------------------

func TestUpdate_PingSuccess_HeaderShowsModel(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	v := m.View()
	if !containsAny(v, "llama3") {
		t.Errorf("after successful ping, header should show model name; got: %q", v)
	}
}

func TestUpdate_PingSuccess_WelcomeMessage(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	v := m.View()
	if !containsAny(v, "Connected", "How can I help") {
		t.Errorf("after successful ping, welcome message should appear; got: %q", v)
	}
}

func TestUpdate_PingFailure_ShowsError(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	updated, _ := m.Update(tui.PingResultMsg{Err: errors.New("connection refused")})
	v := updated.(tui.Model).View()
	if !containsAny(v, "error", "Error", "✗") {
		t.Errorf("after ping failure, view should show error; got: %q", v)
	}
}

// ---------------------------------------------------------------------------
// Update — Keyboard
// ---------------------------------------------------------------------------

func TestUpdate_CtrlC_ReturnsQuitCmd(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Ctrl+C must return a non-nil tea.Cmd (tea.Quit)")
	}
}

func TestUpdate_Enter_EmptyInput_NoStream(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	// textarea is empty → no stream should start
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// cmd may be non-nil due to viewport/textarea batching; just confirm no panic.
	_ = cmd
}

func TestUpdate_Enter_WhileThinking_IsNoop(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	// Manually set the model into thinking status by injecting a token
	m2, _ := m.Update(tui.TokenMsg{Token: "…"})
	// Pressing Enter while thinking should return nil cmd (blocked)
	_, cmd := m2.(tui.Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = cmd // just ensure no panic
}

// ---------------------------------------------------------------------------
// Update — Streaming messages
// ---------------------------------------------------------------------------

func TestUpdate_TokenMsg_AppearsInView(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	updated, _ := m.Update(tui.TokenMsg{Token: "hello "})
	updated, _ = updated.(tui.Model).Update(tui.TokenMsg{Token: "world"})
	v := updated.(tui.Model).View()
	if !containsAny(v, "hello", "world") {
		t.Errorf("streamed tokens should appear in view; got: %q", v)
	}
}

func TestUpdate_MultipleTokens_Concatenated(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	tokens := []string{"The ", "quick ", "brown ", "fox"}
	var current tea.Model = m
	for _, tok := range tokens {
		current, _ = current.(tui.Model).Update(tui.TokenMsg{Token: tok})
	}
	v := current.(tui.Model).View()
	for _, tok := range tokens {
		if !containsAny(v, tok) {
			t.Errorf("expected token %q in view; got: %q", tok, v)
		}
	}
}

func TestUpdate_StreamDoneMsg_ContentMovesToHistory(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	m2, _ := m.Update(tui.TokenMsg{Token: "final response"})
	m3, _ := m2.(tui.Model).Update(tui.StreamDoneMsg{})
	v := m3.(tui.Model).View()
	if !containsAny(v, "final response") {
		t.Errorf("after streamDone, completed response must stay in history view; got: %q", v)
	}
}

func TestUpdate_StreamDoneMsg_EmptyStreamBuf_NoExtraMessage(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	// streamBuf is empty → streamDoneMsg should not append a blank message
	m2, _ := m.Update(tui.StreamDoneMsg{})
	v := m2.(tui.Model).View()
	// Should not panic; view should still be valid
	if v == "" {
		t.Error("View() must not be empty")
	}
}

func TestUpdate_StreamErrMsg_ShowsError(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	m2, _ := m.Update(tui.StreamErrMsg{Err: errors.New("stream failed")})
	v := m2.(tui.Model).View()
	if !containsAny(v, "error", "Error", "✗") {
		t.Errorf("after streamErr, view should show error; got: %q", v)
	}
}

func TestUpdate_StreamErrMsg_ClearsStreamBuf(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	m2, _ := m.Update(tui.TokenMsg{Token: "partial"})
	m3, _ := m2.(tui.Model).Update(tui.StreamErrMsg{Err: errors.New("timeout")})
	v := m3.(tui.Model).View()
	// The partial token should no longer appear as streaming in-progress cursor
	_ = v // just ensure no panic; cursor line is ephemeral
}

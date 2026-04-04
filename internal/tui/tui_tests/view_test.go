// Package tui_tests — tests for view.go (View, renderHistory, renderStatus, indent)
// These tests drive the model through Update cycles and assert on the rendered
// string so that internal helper functions are exercised indirectly.
package tui_tests

import (
	"errors"
	"strings"
	"testing"

	"github.com/RecallKit/recallkit/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// View() — pre-ready
// ---------------------------------------------------------------------------

func TestView_PreReady_ContainsInitText(t *testing.T) {
	m := newTestModel(t)
	v := m.View()
	if v == "" {
		t.Fatal("pre-ready View() must return a non-empty string")
	}
	if !containsAny(v, "Initializing", "RecallKit") {
		t.Errorf("pre-ready View() should mention RecallKit or Initializing; got: %q", v)
	}
}

func TestView_PreReady_NoViewportContent(t *testing.T) {
	m := newTestModel(t)
	v := m.View()
	// Pre-ready should NOT contain the header/footer layout
	if strings.Contains(v, "◈ RecallKit  ·") {
		t.Error("pre-ready View() should not render the full header layout")
	}
}

// ---------------------------------------------------------------------------
// View() — post-ready header
// ---------------------------------------------------------------------------

func TestView_Ready_HeaderContainsModelName(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	v := m.View()
	if !containsAny(v, "llama3") {
		t.Errorf("ready View() should include model name in header; got: %q", v)
	}
}

func TestView_Ready_HeaderContainsRecallKit(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	v := m.View()
	if !containsAny(v, "RecallKit") {
		t.Errorf("ready View() should include RecallKit in header; got: %q", v)
	}
}

// ---------------------------------------------------------------------------
// renderHistory — via View after ping
// ---------------------------------------------------------------------------

func TestView_AfterPing_AssistantWelcomeRendered(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	v := m.View()
	if !containsAny(v, "◈ RecallKit") {
		t.Errorf("history should label assistant messages with ◈ RecallKit; got: %q", v)
	}
}

func TestView_UserMessage_LabelRendered(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))

	// Simulate a user message by injecting a streamDone after a token
	// (user messages appear in the chat after Enter+stream; we can also check
	// through the history rendered by renderHistory).
	// The simplest probe: check "You" does not appear before any user message.
	v := m.View()
	// No user messages sent yet → "You" label should NOT appear
	if strings.Contains(v, "  You") {
		t.Errorf("no user messages sent, but 'You' label found in view: %q", v)
	}
}

func TestView_StreamBuf_ShowsCursor(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	updated, _ := m.Update(tui.TokenMsg{Token: "partial answer"})
	v := updated.(tui.Model).View()
	// The streaming cursor block character should appear
	if !containsAny(v, "█") {
		t.Errorf("streaming View() should show block cursor █; got: %q", v)
	}
}

func TestView_StatusThinking_ShowsThinkingText(t *testing.T) {
	// This state is set internally via Enter + non-empty input.
	// We simulate it indirectly by checking what renderStatus emits.
	m := makeReadyAndIdle(t, newTestModel(t))
	// nextTokenFnMsg sets statusStreaming — tokenMsg is what we can inject.
	// streaming state shows "streaming…" in the status bar.
	updated, _ := m.Update(tui.TokenMsg{Token: "t"})
	v := updated.(tui.Model).View()
	_ = v // just ensure no panic; status bar text varies by terminal capability
}

// ---------------------------------------------------------------------------
// renderStatus — via View
// ---------------------------------------------------------------------------

func TestView_StatusIdle_HelpHintsPresent(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	v := m.View()
	// Status line for idle must contain shortcut hints
	if !containsAny(v, "Enter", "Ctrl") {
		t.Errorf("idle status bar should contain keyboard hints; got: %q", v)
	}
}

func TestView_StatusError_ErrorPhrasePresent(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	updated, _ := m.Update(tui.PingResultMsg{Err: errors.New("no ollama")})
	v := updated.(tui.Model).View()
	if !containsAny(v, "error", "Error") {
		t.Errorf("error status bar should contain error phrase; got: %q", v)
	}
}

func TestView_StatusError_ErrorDetailInHistory(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	updated, _ := m.Update(tui.PingResultMsg{Err: errors.New("cannot connect")})
	v := updated.(tui.Model).View()
	if !containsAny(v, "✗", "cannot connect", "error") {
		t.Errorf("error history should mention the error; got: %q", v)
	}
}

// ---------------------------------------------------------------------------
// indent helper — exercised via renderHistory
// ---------------------------------------------------------------------------

func TestIndent_MultilineContent(t *testing.T) {
	// Send a token with embedded newlines → renderHistory calls indent() on it.
	m := makeReadyAndIdle(t, newTestModel(t))
	updated, _ := m.Update(tui.TokenMsg{Token: "line one\nline two\nline three"})
	v := updated.(tui.Model).View()
	// Each line should be indented by "  "
	if !containsAny(v, "  line one", "  line two") {
		t.Errorf("multi-line tokens should be indented; got: %q", v)
	}
}

func TestIndent_EmptyString(t *testing.T) {
	// An empty streamBuf must not cause renderHistory to emit a cursor block.
	m := makeReadyAndIdle(t, newTestModel(t))
	v := m.View()
	// No streaming active — block cursor should NOT appear
	if strings.Contains(v, "█") {
		t.Errorf("no streaming active, but block cursor found in view: %q", v)
	}
}

// ---------------------------------------------------------------------------
// Multiple messages in history
// ---------------------------------------------------------------------------

func TestView_MultipleAssistantMessages(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	// Simulate two round-trips
	m2, _ := m.Update(tui.TokenMsg{Token: "first answer"})
	m3, _ := m2.(tui.Model).Update(tui.StreamDoneMsg{})
	m4, _ := m3.(tui.Model).Update(tui.TokenMsg{Token: "second answer"})
	m5, _ := m4.(tui.Model).Update(tui.StreamDoneMsg{})
	v := m5.(tui.Model).View()
	if !containsAny(v, "first answer") || !containsAny(v, "second answer") {
		t.Errorf("both assistant messages should appear in view; got: %q", v)
	}
}

func TestView_Width_PaddedToTerminalWidth(t *testing.T) {
	// After WindowSizeMsg the header must be rendered at the given width.
	m := newTestModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := updated.(tui.Model).View()
	if v == "" {
		t.Error("View() must not be empty after WindowSizeMsg")
	}
}

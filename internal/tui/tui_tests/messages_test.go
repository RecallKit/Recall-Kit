// Package tui_tests — tests for messages.go (exported message types).
package tui_tests

import (
	"errors"
	"testing"

	"github.com/RecallKit/recallkit/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// PingResultMsg
// ---------------------------------------------------------------------------

func TestPingResultMsg_NoError(t *testing.T) {
	msg := tui.PingResultMsg{Err: nil}
	if msg.Err != nil {
		t.Errorf("expected nil Err, got %v", msg.Err)
	}
}

func TestPingResultMsg_WithError(t *testing.T) {
	err := errors.New("no ollama")
	msg := tui.PingResultMsg{Err: err}
	if msg.Err == nil {
		t.Error("expected non-nil Err")
	}
	if msg.Err.Error() != "no ollama" {
		t.Errorf("expected error message 'no ollama', got %q", msg.Err.Error())
	}
}

func TestPingResultMsg_UpdateRouting_Success(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	updated, cmd := m.Update(tui.PingResultMsg{Err: nil})
	if updated == nil {
		t.Error("Update with PingResultMsg{Err:nil} returned nil model")
	}
	_ = cmd
}

func TestPingResultMsg_UpdateRouting_Failure(t *testing.T) {
	m := makeReady(t, newTestModel(t))
	updated, _ := m.Update(tui.PingResultMsg{Err: errors.New("down")})
	if updated == nil {
		t.Error("Update with PingResultMsg{Err:err} returned nil model")
	}
}

// ---------------------------------------------------------------------------
// TokenMsg
// ---------------------------------------------------------------------------

func TestTokenMsg_Fields(t *testing.T) {
	msg := tui.TokenMsg{Token: "hello", NextPull: nil}
	if msg.Token != "hello" {
		t.Errorf("expected Token 'hello', got %q", msg.Token)
	}
	if msg.NextPull != nil {
		t.Error("expected nil NextPull")
	}
}

func TestTokenMsg_WithNextPull(t *testing.T) {
	var called bool
	nextPull := tea.Cmd(func() tea.Msg {
		called = true
		return nil
	})
	msg := tui.TokenMsg{Token: "tok", NextPull: nextPull}
	if msg.NextPull == nil {
		t.Error("expected non-nil NextPull")
	}
	// Call it to confirm it's the right func
	msg.NextPull()
	if !called {
		t.Error("NextPull was not called")
	}
}

func TestTokenMsg_UpdateRouting(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	updated, _ := m.Update(tui.TokenMsg{Token: "ping"})
	if updated == nil {
		t.Error("Update with TokenMsg returned nil model")
	}
}

func TestTokenMsg_EmptyToken(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	// An empty token must not panic.
	updated, _ := m.Update(tui.TokenMsg{Token: ""})
	if updated == nil {
		t.Error("Update with empty TokenMsg returned nil model")
	}
}

// ---------------------------------------------------------------------------
// StreamDoneMsg
// ---------------------------------------------------------------------------

func TestStreamDoneMsg_IsZeroValue(t *testing.T) {
	msg := tui.StreamDoneMsg{}
	_ = msg // should be constructible with no fields
}

func TestStreamDoneMsg_UpdateRouting(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	updated, _ := m.Update(tui.StreamDoneMsg{})
	if updated == nil {
		t.Error("Update with StreamDoneMsg returned nil model")
	}
}

func TestStreamDoneMsg_AfterTokens_ClearsStreamBuf(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	m2, _ := m.Update(tui.TokenMsg{Token: "partial"})
	m3, _ := m2.(tui.Model).Update(tui.StreamDoneMsg{})
	v := m3.(tui.Model).View()
	// Block cursor should be gone (stream ended)
	if containsAny(v, "█") {
		t.Errorf("after StreamDoneMsg, block cursor should not appear; got: %q", v)
	}
}

// ---------------------------------------------------------------------------
// StreamErrMsg
// ---------------------------------------------------------------------------

func TestStreamErrMsg_Fields(t *testing.T) {
	err := errors.New("stream timeout")
	msg := tui.StreamErrMsg{Err: err}
	if msg.Err == nil {
		t.Error("expected non-nil Err")
	}
	if msg.Err.Error() != "stream timeout" {
		t.Errorf("expected 'stream timeout', got %q", msg.Err.Error())
	}
}

func TestStreamErrMsg_NilError(t *testing.T) {
	msg := tui.StreamErrMsg{Err: nil}
	if msg.Err != nil {
		t.Errorf("expected nil Err, got: %v", msg.Err)
	}
}

func TestStreamErrMsg_UpdateRouting(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	updated, _ := m.Update(tui.StreamErrMsg{Err: errors.New("rpc error")})
	if updated == nil {
		t.Error("Update with StreamErrMsg returned nil model")
	}
}

func TestStreamErrMsg_AfterTokens_ClearsStreamBuf(t *testing.T) {
	m := makeReadyAndIdle(t, newTestModel(t))
	m2, _ := m.Update(tui.TokenMsg{Token: "incomplete"})
	m3, _ := m2.(tui.Model).Update(tui.StreamErrMsg{Err: errors.New("dropped")})
	v := m3.(tui.Model).View()
	// Block cursor should be gone (error cleared the stream buffer)
	if containsAny(v, "█") {
		t.Errorf("after StreamErrMsg, block cursor should not appear; got: %q", v)
	}
}

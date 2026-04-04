// Package tui_tests — tests for commands.go (MakeTokenPuller, cmdPing, cmdStartStream).
// cmdPing and cmdStartStream are unexported; their behaviour is covered indirectly
// by the model-level Update tests in models_test.go (PingResultMsg round-trips).
// MakeTokenPuller is exported and tested directly below via pre-built channels.
package tui_tests

import (
	"errors"
	"testing"

	"github.com/RecallKit/recallkit/internal/tui"
)


// ---------------------------------------------------------------------------
// MakeTokenPuller
// ---------------------------------------------------------------------------

func TestMakeTokenPuller_ReceivesToken(t *testing.T) {
	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)
	tokenCh <- "hello"

	cmd := tui.MakeTokenPuller(tokenCh, errCh)
	msg := cmd()

	tmsg, ok := msg.(tui.TokenMsg)
	if !ok {
		t.Fatalf("expected TokenMsg, got %T", msg)
	}
	if tmsg.Token != "hello" {
		t.Errorf("expected Token 'hello', got %q", tmsg.Token)
	}
	if tmsg.NextPull == nil {
		t.Error("NextPull must not be nil after receiving a token")
	}
}

func TestMakeTokenPuller_ChannelClosed_NoError_ReturnsStreamDone(t *testing.T) {
	tokenCh := make(chan string)
	errCh := make(chan error, 1)
	close(tokenCh)

	cmd := tui.MakeTokenPuller(tokenCh, errCh)
	msg := cmd()

	if _, ok := msg.(tui.StreamDoneMsg); !ok {
		t.Fatalf("expected StreamDoneMsg when channel is closed with no error, got %T", msg)
	}
}

func TestMakeTokenPuller_ChannelClosed_WithError_ReturnsStreamErr(t *testing.T) {
	tokenCh := make(chan string)
	errCh := make(chan error, 1)
	errCh <- errors.New("EOF")
	close(tokenCh)

	cmd := tui.MakeTokenPuller(tokenCh, errCh)
	msg := cmd()

	serr, ok := msg.(tui.StreamErrMsg)
	if !ok {
		t.Fatalf("expected StreamErrMsg, got %T", msg)
	}
	if serr.Err == nil || serr.Err.Error() != "EOF" {
		t.Errorf("expected error 'EOF', got %v", serr.Err)
	}
}

func TestMakeTokenPuller_ErrorChannel_ReturnsStreamErr(t *testing.T) {
	tokenCh := make(chan string) // never sends
	errCh := make(chan error, 1)
	errCh <- errors.New("network error")

	cmd := tui.MakeTokenPuller(tokenCh, errCh)
	msg := cmd()

	serr, ok := msg.(tui.StreamErrMsg)
	if !ok {
		t.Fatalf("expected StreamErrMsg from errCh, got %T", msg)
	}
	if serr.Err.Error() != "network error" {
		t.Errorf("expected 'network error', got %q", serr.Err.Error())
	}
}

func TestMakeTokenPuller_MultipleTokens_ChainedNextPull(t *testing.T) {
	tokenCh := make(chan string, 3)
	errCh := make(chan error, 1)
	tokenCh <- "a"
	tokenCh <- "b"
	tokenCh <- "c"
	close(tokenCh)

	var collected []string
	cmd := tui.MakeTokenPuller(tokenCh, errCh)
	for {
		msg := cmd()
		if tmsg, ok := msg.(tui.TokenMsg); ok {
			collected = append(collected, tmsg.Token)
			cmd = tmsg.NextPull
		} else {
			break
		}
	}

	if len(collected) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(collected), collected)
	}
	expected := []string{"a", "b", "c"}
	for i, tok := range expected {
		if collected[i] != tok {
			t.Errorf("token[%d]: expected %q, got %q", i, tok, collected[i])
		}
	}
}

// ---------------------------------------------------------------------------
// cmdPing / cmdStartStream — tested indirectly through the model
// ---------------------------------------------------------------------------
// Both functions are unexported. Their message-routing behaviour is already
// covered by the PingResultMsg and streaming tests in models_test.go.
// A direct integration test would require a running Ollama instance and is
// intentionally left out of the unit-test suite.

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient returns an OllamaClient wired to the given test server.
func newTestClient(srv *httptest.Server) *OllamaClient {
	return &OllamaClient{
		BaseURL: srv.URL,
		http:    srv.Client(),
	}
}

// Ping

func TestPing_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Ping creates its own http.Client, so we only need the URL.
	c := &OllamaClient{BaseURL: srv.URL, http: srv.Client()}
	if err := c.Ping(); err != nil {
		t.Fatalf("Ping() expected nil error, got: %v", err)
	}
}

func TestPing_Unreachable(t *testing.T) {
	// Point at a port that is definitely not listening.
	c := &OllamaClient{BaseURL: "http://127.0.0.1:1", http: &http.Client{Timeout: time.Second}}
	err := c.Ping()
	if err == nil {
		t.Fatal("Ping() expected an error for unreachable host, got nil")
	}
	if !strings.Contains(err.Error(), "ollama not reachable") {
		t.Errorf("Ping() error message unexpected: %v", err)
	}
}

// ListModels

func TestListModels_Success(t *testing.T) {
	want := []string{"llama3", "mistral"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		type model struct {
			Name string `json:"name"`
		}
		resp := struct {
			Models []model `json:"models"`
		}{
			Models: []model{{Name: "llama3"}, {Name: "mistral"}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	got, err := c.ListModels()
	if err != nil {
		t.Fatalf("ListModels() unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("ListModels() returned %d models, want %d", len(got), len(want))
	}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("ListModels()[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestListModels_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"models":[]}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	got, err := c.ListModels()
	if err != nil {
		t.Fatalf("ListModels() unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListModels() expected empty slice, got %v", got)
	}
}

func TestListModels_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not-json`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.ListModels()
	if err == nil {
		t.Fatal("ListModels() expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode models") {
		t.Errorf("ListModels() error message unexpected: %v", err)
	}
}

func TestListModels_NetworkError(t *testing.T) {
	// Close the server immediately so the request fails at the transport layer.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := newTestClient(srv)
	_, err := c.ListModels()
	if err == nil {
		t.Fatal("ListModels() expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "list models") {
		t.Errorf("ListModels() error message unexpected: %v", err)
	}
}

// ---------------------------------------------------------------------------
// StreamChat
// ---------------------------------------------------------------------------

// buildNDJSON produces a newline-delimited sequence of chatChunk JSON lines.
func buildNDJSON(chunks []chatChunk) string {
	var sb strings.Builder
	for _, ch := range chunks {
		b, _ := json.Marshal(ch)
		sb.Write(b)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func TestStreamChat_Success(t *testing.T) {
	chunks := []chatChunk{
		{Message: Message{Role: "assistant", Content: "Hello"}, Done: false},
		{Message: Message{Role: "assistant", Content: ", world"}, Done: false},
		{Message: Message{Role: "assistant", Content: "!"}, Done: true},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprint(w, buildNDJSON(chunks))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	tokenCh := make(chan string, 10)
	errCh := make(chan error, 1)

	c.StreamChat(context.Background(), "llama3", []Message{
		{Role: "user", Content: "Say hello"},
	}, tokenCh, errCh)

	var got []string
	for tok := range tokenCh {
		got = append(got, tok)
	}

	select {
	case err := <-errCh:
		t.Fatalf("StreamChat() unexpected error: %v", err)
	default:
	}

	want := []string{"Hello", ", world", "!"}
	if len(got) != len(want) {
		t.Fatalf("StreamChat() got tokens %v, want %v", got, want)
	}
	for i, tok := range want {
		if got[i] != tok {
			t.Errorf("StreamChat() token[%d] = %q, want %q", i, got[i], tok)
		}
	}
}

func TestStreamChat_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	tokenCh := make(chan string, 10)
	errCh := make(chan error, 1)

	c.StreamChat(context.Background(), "llama3", nil, tokenCh, errCh)

	// Drain tokens (channel will be closed by StreamChat goroutine).
	for range tokenCh {
	}

	select {
	case err := <-errCh:
		if !strings.Contains(err.Error(), "status 400") {
			t.Errorf("StreamChat() expected status-code error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamChat() timed out waiting for error")
	}
}

func TestStreamChat_OllamaErrorChunk(t *testing.T) {
	chunks := []chatChunk{
		{Error: "model not found"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, buildNDJSON(chunks))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	tokenCh := make(chan string, 10)
	errCh := make(chan error, 1)

	c.StreamChat(context.Background(), "missing-model", nil, tokenCh, errCh)

	for range tokenCh {
	}

	select {
	case err := <-errCh:
		if !strings.Contains(err.Error(), "model not found") {
			t.Errorf("StreamChat() error = %v, want it to contain 'model not found'", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamChat() timed out waiting for error")
	}
}

func TestStreamChat_ContextCancelled(t *testing.T) {
	// Server that blocks until the request context is cancelled.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		// Send one token then block.
		chunk := chatChunk{Message: Message{Role: "assistant", Content: "start"}, Done: false}
		b, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "%s\n", b)
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	c := newTestClient(srv)
	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	c.StreamChat(ctx, "llama3", []Message{{Role: "user", Content: "hi"}}, tokenCh, errCh)

	// Receive the first token then cancel.
	select {
	case <-tokenCh:
		cancel()
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("StreamChat() timed out waiting for first token")
	}

	// tokenCh must eventually close.
	select {
	case _, open := <-tokenCh:
		if open {
			// drain any buffered tokens
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamChat() channel never closed after context cancel")
	}
}

func TestStreamChat_MalformedChunksSkipped(t *testing.T) {
	// Mix of bad JSON and a valid final chunk.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not-json\n")
		b, _ := json.Marshal(chatChunk{Message: Message{Role: "assistant", Content: "ok"}, Done: true})
		fmt.Fprintf(w, "%s\n", b)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	tokenCh := make(chan string, 10)
	errCh := make(chan error, 1)

	c.StreamChat(context.Background(), "llama3", nil, tokenCh, errCh)

	var tokens []string
	for tok := range tokenCh {
		tokens = append(tokens, tok)
	}

	if len(tokens) != 1 || tokens[0] != "ok" {
		t.Errorf("StreamChat() tokens = %v, want [\"ok\"]", tokens)
	}

	select {
	case err := <-errCh:
		t.Errorf("StreamChat() unexpected error: %v", err)
	default:
	}
}

func TestStreamChat_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close before the request

	c := newTestClient(srv)
	tokenCh := make(chan string, 10)
	errCh := make(chan error, 1)

	c.StreamChat(context.Background(), "llama3", nil, tokenCh, errCh)

	for range tokenCh {
	}

	select {
	case err := <-errCh:
		if !strings.Contains(err.Error(), "ollama request") {
			t.Errorf("StreamChat() error = %v, want it to contain 'ollama request'", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamChat() timed out waiting for network error")
	}
}

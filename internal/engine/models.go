package engine

import "net/http"

// message is a single turn during the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the payload for POST /api/chat.
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// chatChunk is one line of the NDJSON stream from Ollama.
type chatChunk struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
	Error   string  `json:"error,omitempty"`
}

// OllamaClient talks to a local Ollama runtime.
type OllamaClient struct {
	BaseURL string
	http    *http.Client
}

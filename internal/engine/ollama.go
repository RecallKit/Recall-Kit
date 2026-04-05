package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultOllamaURL = "http://localhost:11434"

// NewOllamaClient returns a client pointed at the local Ollama instance.
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		BaseURL: defaultOllamaURL,
		http: &http.Client{
			Timeout: 0, // no timeout — streams can be long
		},
	}
}

// Ping checks that Ollama is reachable. Returns an error if not.
func (c *OllamaClient) Ping() error {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(c.BaseURL)
	if err != nil {
		return fmt.Errorf("ollama not reachable at %s: %w", c.BaseURL, err)
	}
	resp.Body.Close()
	return nil
}

// StreamChat sends messages to Ollama and streams token chunks into tokenCh.
// The channel is closed when streaming completes. Errors are sent to errCh.
// Call this in a goroutine or let it manage its own goroutine internally.
func (c *OllamaClient) StreamChat(
	ctx context.Context,
	model string,
	messages []Message,
	tokenCh chan<- string,
	errCh chan<- error,
) {
	go func() {
		defer close(tokenCh)

		body, err := json.Marshal(chatRequest{
			Model:    model,
			Messages: messages,
			Stream:   true,
		})
		if err != nil {
			errCh <- fmt.Errorf("marshal: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.BaseURL+"/api/chat", bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("build request: %w", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("ollama request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("ollama returned status %d", resp.StatusCode)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var chunk chatChunk
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				continue // skip malformed lines
			}
			if chunk.Error != "" {
				errCh <- fmt.Errorf("ollama: %s", chunk.Error)
				return
			}
			if chunk.Message.Content != "" {
				select {
				case <-ctx.Done():
					return
				case tokenCh <- chunk.Message.Content:
				}
			}
			if chunk.Done {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("stream read: %w", err)
		}
	}()
}

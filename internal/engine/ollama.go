package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// ValidateModel checks that the given model is available locally.
// Returns a clear error with available models listed if not found.
func (c *OllamaClient) ValidateModel(model string) error {
	models, err := c.ListModels()
	if err != nil {
		return fmt.Errorf("could not list models: %w", err)
	}

	for _, m := range models {
		if m.Name == model {
			return nil
		}
	}

	// Build a helpful error listing what IS available
	msg := fmt.Sprintf("model %q not found in Ollama.\n\n  Installed models:\n", model)
	if len(models) == 0 {
		msg += "    (none — run `ollama pull <model>` to install one)\n"
	} else {
		for _, m := range models {
			msg += fmt.Sprintf("    • %s\n", m.Name)
		}
		msg += "\n  Use one of the above with: recallkit start --model <name>"
	}
	return errors.New(msg)
}

// StreamChat sends messages to Ollama and streams token chunks into tokenCh.
// The channel is closed when streaming completes. Errors are sent to errCh.
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

		if resp.StatusCode == http.StatusNotFound {
			errCh <- fmt.Errorf("model not found — run `ollama list` to see installed models, or `ollama pull <model>` to install one")
			return
		}
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

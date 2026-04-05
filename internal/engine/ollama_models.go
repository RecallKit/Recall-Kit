package engine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ModelInfo holds metadata about a locally installed Ollama model.
type ModelInfo struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
}

// ListModels returns metadata for all locally installed models.
func (c *OllamaClient) ListModels() ([]ModelInfo, error) {
	resp, err := c.http.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []ModelInfo `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode models: %w", err)
	}
	return result.Models, nil
}

// PullProgress is one status line from the Ollama pull stream.
type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// PullModel pulls a model by name from the Ollama registry, streaming
// progress updates to progressCh. The channel is closed on completion.
// Errors are sent to errCh.
func (c *OllamaClient) PullModel(name string, progressCh chan<- PullProgress, errCh chan<- error) {
	go func() {
		defer close(progressCh)

		body, _ := json.Marshal(map[string]any{
			"name":   name,
			"stream": true,
		})

		resp, err := c.http.Post(c.BaseURL+"/api/pull", "application/json", bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("pull request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("pull returned status %d", resp.StatusCode)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var p PullProgress
			if err := json.Unmarshal(scanner.Bytes(), &p); err != nil {
				continue
			}
			progressCh <- p
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("pull stream: %w", err)
		}
	}()
}

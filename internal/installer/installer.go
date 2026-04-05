// Package installer handles the recallkit init flow — detecting whether
// Ollama is installed and running, and installing it if not.
package installer

import (
	"fmt"
	"runtime"
)

const ollamaURL = "http://localhost:11434"

// Run is the top-level entry point called by `recallkit init`.

func Run() error {
	fmt.Println("◈ RecallKit — initializing")
	fmt.Println()

	// Step 1: check if ollama binary exists on PATH
	if isOllamaInstalled() {
		fmt.Println("✔  Ollama is already installed:", ollamaPath())
	} else {
		fmt.Println("✗  Ollama not found on PATH")
		fmt.Println("→  Installing Ollama for", runtime.GOOS, "/", runtime.GOARCH)
		fmt.Println()

		if err := installOllama(); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
		fmt.Println("✔  Ollama installed successfully")
	}

	fmt.Println()

	// Step 2: check if the Ollama daemon is reachable
	if isOllamaRunning() {
		fmt.Println("✔  Ollama daemon is running at", ollamaURL)
	} else {
		fmt.Println("→  Ollama daemon not running — starting it…")
		if err := startOllamaDaemon(); err != nil {
			return fmt.Errorf("could not start Ollama daemon: %w", err)
		}
		fmt.Println("✔  Ollama daemon started")
	}

	fmt.Println()

	// Step 3: ensure at least one model is available
	if err := ensureModel(); err != nil {
		return fmt.Errorf("model setup: %w", err)
	}

	fmt.Println()
	fmt.Println("✔  RecallKit is ready. Run `recallkit start` to begin chatting.")
	return nil
}

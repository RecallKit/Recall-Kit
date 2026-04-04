// Package installer handles the recallkit init flow — detecting whether
// Ollama is installed and running, and installing it if not.
package installer

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const ollamaURL = "http://localhost:11434"

// Run is the top-level entry point called by `recallkit init`.
// It walks the user through detection → install → verify.
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
	fmt.Println("✔  RecallKit is ready. Run `recallkit start` to begin chatting.")
	return nil
}

// ── Detection ────────────────────────────────────────────────────────────────

// isOllamaInstalled returns true if the ollama binary is on the PATH.
func isOllamaInstalled() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

// ollamaPath returns the full path of the ollama binary, or "not found".
func ollamaPath() string {
	p, err := exec.LookPath("ollama")
	if err != nil {
		return "not found"
	}
	return p
}

// isOllamaRunning pings the Ollama HTTP API with a short timeout.
func isOllamaRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(ollamaURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// ── Installation ─────────────────────────────────────────────────────────────

// installOllama dispatches to the correct platform installer.
func installOllama() error {
	switch runtime.GOOS {
	case "linux":
		return installLinux()
	case "darwin":
		return installMacOS()
	case "windows":
		return installWindows()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// installLinux runs the official Ollama install script via curl | sh.
// This is the method documented on ollama.com for Linux.
func installLinux() error {
	fmt.Println("  Running: curl -fsSL https://ollama.com/install.sh | sh")
	fmt.Println("  (You may be prompted for your sudo password)")
	fmt.Println()

	// curl downloads the script, sh executes it
	curl := exec.Command("curl", "-fsSL", "https://ollama.com/install.sh")
	sh := exec.Command("sh")

	// Pipe curl stdout → sh stdin
	pipe, err := curl.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe setup: %w", err)
	}
	sh.Stdin = pipe
	sh.Stdout = os.Stdout
	sh.Stderr = os.Stderr

	if err := curl.Start(); err != nil {
		return fmt.Errorf("curl start: %w", err)
	}
	if err := sh.Start(); err != nil {
		return fmt.Errorf("sh start: %w", err)
	}
	if err := curl.Wait(); err != nil {
		return fmt.Errorf("curl: %w", err)
	}
	return sh.Wait()
}

// installMacOS guides the user to the official macOS installer.
// Ollama on macOS is a .app bundle — there is no silent CLI install path,
// so we open the download page and give clear instructions.
func installMacOS() error {
	fmt.Println("  Ollama on macOS is distributed as a .app bundle.")
	fmt.Println("  Opening the download page in your browser…")
	fmt.Println()

	// Try to open the browser; non-fatal if it fails
	_ = exec.Command("open", "https://ollama.com/download/mac").Run()

	fmt.Println("  1. Download and open Ollama.dmg")
	fmt.Println("  2. Drag Ollama.app to your Applications folder")
	fmt.Println("  3. Launch Ollama from Applications")
	fmt.Println("  4. Re-run `recallkit init` once Ollama is running")
	fmt.Println()

	return errors.New("manual installation required on mac	OS — see instructions above")
}

// installWindows guides the user to the official Windows installer.
func installWindows() error {
	fmt.Println("  Ollama on Windows is distributed as an installer (.exe).")
	fmt.Println("  Opening the download page in your browser…")
	fmt.Println()

	_ = exec.Command("cmd", "/c", "start", "https://ollama.com/download/windows").Run()

	fmt.Println("  1. Download and run OllamaSetup.exe")
	fmt.Println("  2. Follow the installer prompts")
	fmt.Println("  3. Re-run `recallkit init` once Ollama is running")
	fmt.Println()

	return errors.New("manual installation required on Windows — see instructions above")
}

// ── Daemon management ─────────────────────────────────────────────────────────

// startOllamaDaemon starts `ollama serve` as a background process and
// waits up to 5 seconds for the HTTP API to become reachable.
func startOllamaDaemon() error {
	cmd := exec.Command("ollama", "serve")
	cmd.Stdout = nil // detach stdout — we don't want it in the terminal
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ollama serve: %w", err)
	}

	// Poll until ready or timeout
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if isOllamaRunning() {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}

	return errors.New("ollama serve started but did not become reachable within 5 seconds")
}

// to ensure model is installed and reccomend models otherwise

package installer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/RecallKit/recallkit/internal/engine"
)

// ── Model management ──────────────────────────────────────────────────────────

// recommendedModels is a curated list shown when no models are installed.
var recommendedModels = []struct {
	name        string
	description string
}{
	{"llama3.2:3b", "Meta Llama 3.2 · 3B · ~2GB  · fast, good for most tasks"},
	{"llama3.2:1b", "Meta Llama 3.2 · 1B · ~1GB  · very fast, lightweight"},
	{"llama3.1:8b", "Meta Llama 3.1 · 8B · ~5GB  · best quality in small tier"},
	{"mistral:7b", "Mistral 7B     · 7B · ~4GB  · strong reasoning"},
	{"gemma2:2b", "Google Gemma 2 · 2B · ~2GB  · efficient, multilingual"},
	{"phi3:mini", "Microsoft Phi3 · 3B · ~2GB  · great for coding tasks"},
}

// ensureModel checks if any models are installed. If none are, it prompts
// the user to pick one from the recommended list and pulls it.
func ensureModel() error {
	client := engine.NewOllamaClient()
	models, err := client.ListModels()
	if err != nil {
		return fmt.Errorf("could not list models: %w", err)
	}

	if len(models) > 0 {
		fmt.Printf("✔  %d model(s) available:\n", len(models))
		for _, m := range models {
			fmt.Printf("     • %s\n", m.Name)
		}
		return nil
	}

	// No models installed — prompt the user
	fmt.Println("✗  No models installed")
	fmt.Println()
	fmt.Println("  Recommended models:")
	fmt.Println()
	for i, m := range recommendedModels {
		fmt.Printf("  [%d] %s\n", i+1, m.description)
	}
	fmt.Println("  [0] Skip — I'll pull a model manually later")
	fmt.Println()

	choice := promptChoice("  Enter number: ", 0, len(recommendedModels))
	if choice == 0 {
		fmt.Println("  Skipping model pull. Run `ollama pull <model>` when ready.")
		return nil
	}

	selected := recommendedModels[choice-1]
	fmt.Printf("\n→  Pulling %s — this may take a few minutes…\n\n", selected.name)
	return pullWithProgress(client, selected.name)
}

// pullWithProgress pulls a model and renders a live progress bar.
func pullWithProgress(client *engine.OllamaClient, name string) error {
	progressCh := make(chan engine.PullProgress, 32)
	errCh := make(chan error, 1)

	client.PullModel(name, progressCh, errCh)

	var lastStatus string
	for p := range progressCh {
		if p.Status != lastStatus {
			if lastStatus != "" {
				fmt.Println() // newline after previous status line
			}
			fmt.Printf("  %s", p.Status)
			lastStatus = p.Status
		}

		// Show download progress if total is known
		if p.Total > 0 {
			pct := float64(p.Completed) / float64(p.Total) * 100
			bar := progressBar(pct, 30)
			fmt.Printf("\r  %s  %s  %.1f%%", p.Status, bar, pct)
		}
	}
	fmt.Println()

	select {
	case err := <-errCh:
		return fmt.Errorf("pull failed: %w", err)
	default:
	}

	fmt.Printf("✔  %s is ready\n", name)
	return nil
}

// progressBar renders a simple ASCII progress bar of the given width.
func progressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

// promptChoice reads a number from stdin in the range [min, max].
func promptChoice(prompt string, min, max int) int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		var n int
		if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < min || n > max {
			fmt.Printf("  Please enter a number between %d and %d\n", min, max)
			continue
		}
		return n
	}
}

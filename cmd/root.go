// cmd/root.go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "recallkit",
	Short: "A local-first context engineering tool for LLMs",
	Long:  `RecallKit allows you to store, structure, and reuse conversational context as a versioned graph.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// You can define global flags here later (e.g., --config, --verbose)
}
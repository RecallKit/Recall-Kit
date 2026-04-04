package cmd

import (
	"github.com/RecallKit/recallkit/internal/installer"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Check dependencies and install Ollama if missing",
	Long: `recallkit init checks your system for required dependencies.
If Ollama is not found, it will install it automatically using the
official installer for your platform (macOS, Linux, Windows).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installer.Run()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

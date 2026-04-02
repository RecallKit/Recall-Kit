// cmd/start.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const asciiArt = `
  _____                 _ _ _  ___ _   
 |  __ \               | | | |/ (_) |  
 | |__) |___  ___  __ _| | | ' / _| |_ 
 |  _  // _ \/ __|/ _` + "`" + ` | | |  < | | __|
 | | \ \  __/ (__| (_| | | | . \| | |_ 
 |_|  \_\___|\___|\__,_|_|_|_|\_\_|\__|
                                       
 Context Graph Engine Initialized.
`

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the RecallKit interactive session",
	Long:  `Initializes the local context graph and boots up the interactive terminal UI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// In the future, this is where you will initialize:
		// 1. The Kùzu/SQLite DB connection (from internal/db)
		// 2. The Ollama health check (from internal/engine)
		// 3. The Bubble Tea TUI (from internal/tui)
		
		fmt.Println(asciiArt)
		fmt.Println("🚀 Ready to build your context graph...")
	},
}

func init() {
	// Register the 'start' command with the root command
	rootCmd.AddCommand(startCmd)
}
package cmd

import (
	"fmt"

	"github.com/RecallKit/recallkit/internal/installer"
	"github.com/spf13/cobra"
)

const asciiArt2 = `
  _____                 _ _ _  ___ _   
 |  __ \               | | | |/ (_) |  
 | |__) |___  ___  __ _| | | ' / _| |_ 
 |  _  // _ \/ __|/ _` + "`" + ` | | |  < | | __|
 | | \ \  __/ (__| (_| | | | . \| | |_ 
 |_|  \_\___|\___|\__,_|_|_|_|\_\_|\__|

 Context Graph Engine Initializing...
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Check dependencies and install Ollama if missing",
	Long: `recallkit init checks your system for required dependencies.
If Ollama is not found, it will install it automatically using the
official installer for your platform (macOS, Linux, Windows).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(asciiArt)

		return installer.Run()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

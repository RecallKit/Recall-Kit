// cmd/ui/ui.go
package ui

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the RecallKit visual web interface",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("launching chat UI on localhost:8001")
	},
}

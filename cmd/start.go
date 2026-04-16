package cmd

import (
	"fmt"

	"github.com/RecallKit/recallkit/internal/engine"
	"github.com/RecallKit/recallkit/internal/session"
	"github.com/RecallKit/recallkit/internal/tui"
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

var (
	ollamaModel string
	sessionID   string
	sessionName string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the RecallKit interactive session",
	Long:  `Initializes the local context graph and boots up the interactive terminal UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(asciiArt)

		// Validate the model exists before doing anything else
		client := engine.NewOllamaClient()
		if err := client.ValidateModel(ollamaModel); err != nil {
			return err
		}

		store, err := session.NewStore()
		if err != nil {
			return err
		}

		var sess *session.Session

		if sessionID != "" {
			// Resume an existing session by ID
			sess, err = store.Load(sessionID)
			if err != nil {
				return err
			}
			fmt.Printf(" Resuming session: %s · model: %s\n\n", sess.Name, sess.Model)
		} else {
			// Start a new session
			name := sessionName
			if name == "" {
				name = "session"
			}
			sess, err = store.Create(name, ollamaModel)
			if err != nil {
				return err
			}
			fmt.Printf(" New session: %s · model: %s\n\n", sess.Name, sess.Model)
		}

		return tui.Start(sess, store)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&ollamaModel, "model", "m", "llama3", "Ollama model to use")
	startCmd.Flags().StringVarP(&sessionID, "session", "s", "", "Resume a session by ID")
	startCmd.Flags().StringVarP(&sessionName, "name", "n", "", "Name for the new session")
}

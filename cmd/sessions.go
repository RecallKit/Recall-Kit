package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/RecallKit/recallkit/internal/session"
	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List all saved chat sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.NewStore()
		if err != nil {
			return err
		}

		sessions, err := store.List()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions yet. Start one with: recallkit start")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tMODEL\tMESSAGES\tLAST ACTIVE")
		fmt.Fprintln(w, "──\t────\t─────\t────────\t───────────")
		for _, s := range sessions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				s.ID,
				s.Name,
				s.Model,
				len(s.Messages),
				humanTime(s.UpdatedAt),
			)
		}
		w.Flush()
		return nil
	},
}

var deleteSessionCmd = &cobra.Command{
	Use:   "delete <session-id>",
	Short: "Delete a saved session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.NewStore()
		if err != nil {
			return err
		}
		if err := store.Delete(args[0]); err != nil {
			return err
		}
		fmt.Printf("✔  Session %q deleted.\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(deleteSessionCmd)
}

// humanTime formats a time as a human-friendly relative string.
func humanTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return t.Format("2006-01-02")
	}
}

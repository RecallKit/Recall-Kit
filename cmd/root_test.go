// cmd/root_test.go
package cmd

import (
	"testing"
)

func TestRootCommandConfiguration(t *testing.T) {
	if rootCmd.Use != "recallkit" {
		t.Errorf("Expected root command use to be 'recallkit', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected root command to have a short description")
	}
}
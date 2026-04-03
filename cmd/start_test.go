// cmd/start_test.go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestStartCmdExecution(t *testing.T) {
	//Create a buffer to capture standard output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"start"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("Unexpected error executing 'start' command: %v", err)
	}

	output := buf.String()
	expectedText := "Context Graph Engine Initialized."

	if !strings.Contains(output, expectedText) {
		t.Errorf("Expected output to contain %q, but got:\n%s", expectedText, output)
	}
}

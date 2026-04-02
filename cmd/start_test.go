// cmd/start_test.go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestStartCmdExecution(t *testing.T) {
	// 1. Create a buffer to capture standard output
	buf := new(bytes.Buffer)
	
	// 2. Redirect the root command's output to our buffer
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	
	// 3. Programmatically pass the "start" argument
	rootCmd.SetArgs([]string{"start"})

	// 4. Execute the command
	err := rootCmd.Execute()
	
	// 5. Assert there were no execution errors
	if err != nil {
		t.Fatalf("Unexpected error executing 'start' command: %v", err)
	}

	// 6. Assert the output contains our expected text
	output := buf.String()
	expectedText := "Context Graph Engine Initialized."
	
	if !strings.Contains(output, expectedText) {
		t.Errorf("Expected output to contain %q, but got:\n%s", expectedText, output)
	}
}
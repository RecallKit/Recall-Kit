// cmd/start_test.go
package cmd

import (
	"testing"
)

func TestStartCmd_UseField(t *testing.T) {
	if startCmd.Use != "start" {
		t.Errorf("expected startCmd.Use to be 'start', got %q", startCmd.Use)
	}
}

func TestStartCmd_HasShortDescription(t *testing.T) {
	if startCmd.Short == "" {
		t.Error("startCmd must have a non-empty Short description")
	}
}

func TestStartCmd_IsRegisteredWithRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "start" {
			return // found
		}
	}
	t.Error("'start' command is not registered as a subcommand of rootCmd")
}

func TestStartCmd_ModelFlag(t *testing.T) {
	f := startCmd.Flags().Lookup("model")
	if f == nil {
		t.Fatal("startCmd is missing the --model flag")
	}
	if f.DefValue != "llama3" {
		t.Errorf("--model default expected 'llama3', got %q", f.DefValue)
	}
}

func TestStartCmd_SessionFlag(t *testing.T) {
	f := startCmd.Flags().Lookup("session")
	if f == nil {
		t.Fatal("startCmd is missing the --session flag")
	}
}

func TestStartCmd_NameFlag(t *testing.T) {
	f := startCmd.Flags().Lookup("name")
	if f == nil {
		t.Fatal("startCmd is missing the --name flag")
	}
}

func TestStartCmd_AsciiArtContainsBanner(t *testing.T) {
	if asciiArt == "" {
		t.Error("asciiArt must not be empty")
	}
}

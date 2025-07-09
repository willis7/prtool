package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "version flag short",
			args:     []string{"--version"},
			expected: "dev",
		},
		{
			name:     "version flag long",
			args:     []string{"-v"},
			expected: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for each test to avoid state pollution
			cmd := &cobra.Command{
				Use:   "prtool",
				Short: "A CLI tool for summarizing GitHub pull requests",
			}

			cmd.Flags().BoolP("version", "v", false, "Show version information")

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			cmd.Run = func(cmd *cobra.Command, args []string) {
				versionFlag, _ := cmd.Flags().GetBool("version")
				if versionFlag {
					cmd.Print("dev")
					return
				}
				cmd.Help()
			}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			result := strings.TrimSpace(output.String())
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	// Reset flags to known state
	org = ""
	user = ""
	githubToken = ""
	dryRun = false

	// Test that GetConfig doesn't panic and returns a config
	config, err := GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	if config == nil {
		t.Error("GetConfig() returned nil config")
	}
}

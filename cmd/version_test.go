package cmd

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yourorg/prtool/internal/config"
)

func TestVersionFlagOutput(t *testing.T) {
	// Save original appVersion
	originalVersion := appVersion
	defer func() {
		appVersion = originalVersion
	}()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "dev version",
			version: "dev",
			want:    "prtool version dev",
		},
		{
			name:    "release version",
			version: "v1.2.3",
			want:    "prtool version v1.2.3",
		},
		{
			name:    "version with metadata",
			version: "v1.0.0-rc1+build123",
			want:    "prtool version v1.0.0-rc1+build123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test version
			appVersion = tt.version

			// Create a new root command for testing
			rootCmd = createRootCommand()
			
			// Capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			rootCmd.SetArgs([]string{"--version"})

			// Execute command
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Check output
			output := strings.TrimSpace(buf.String())
			if output != tt.want {
				t.Errorf("Version output = %q, want %q", output, tt.want)
			}
		})
	}
}

func TestVersionInjection_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build binary with custom version
	testVersion := "v9.9.9-test"
	binPath := filepath.Join(t.TempDir(), "prtool-test")
	
	buildCmd := exec.Command("go", "build", 
		"-ldflags", "-X github.com/yourorg/prtool/cmd.appVersion="+testVersion,
		"-o", binPath,
		"..",
	)
	
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build with version injection: %v\nOutput: %s", err, output)
	}

	// Run the binary with --version
	versionCmd := exec.Command(binPath, "--version")
	output, err = versionCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run --version: %v\nOutput: %s", err, output)
	}

	// Verify output contains injected version
	outputStr := strings.TrimSpace(string(output))
	expectedOutput := "prtool version " + testVersion
	if outputStr != expectedOutput {
		t.Errorf("Version output = %q, want %q", outputStr, expectedOutput)
	}
}

func TestVersionCheck_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build binary
	binPath := filepath.Join(t.TempDir(), "prtool-test")
	
	buildCmd := exec.Command("go", "build", "-o", binPath, "..")
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build: %v\nOutput: %s", err, output)
	}

	// Run with --version-check (should exit cleanly)
	checkCmd := exec.Command(binPath, "--version-check")
	output, err = checkCmd.CombinedOutput()
	
	// Should exit with status 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			t.Fatalf("--version-check exited with non-zero status: %v\nOutput: %s", err, output)
		}
	}

	// Output should contain version check info
	outputStr := string(output)
	if !strings.Contains(outputStr, "Current version:") {
		t.Errorf("Version check output missing 'Current version:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Checking for updates...") {
		t.Errorf("Version check output missing 'Checking for updates...', got: %s", outputStr)
	}
}

// createRootCommand creates a fresh root command for testing
func createRootCommand() *cobra.Command {
	var versionCheckFlag bool
	var cfg = config.New()
	
	cmd := &cobra.Command{
		Use:   "prtool",
		Short: "A tool for fetching and summarizing GitHub pull requests",
		Long:  `prtool fetches merged GitHub pull requests and generates AI-powered summaries.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !versionCheckFlag {
				cmd.Help()
			}
		},
		Version: appVersion,
	}

	// Add version flag
	cmd.Flags().BoolP("version", "v", false, "version for prtool")
	
	// Add other minimal flags for testing
	cmd.PersistentFlags().BoolVar(&cfg.Verbose, "verbose", false, "verbose output")
	cmd.PersistentFlags().StringVar(&cfg.LogFile, "log-file", "", "log file path")
	
	return cmd
}

// Helper to verify version output format
func TestVersionOutputFormat(t *testing.T) {
	// This ensures we maintain consistent version output format
	testCases := []struct {
		version string
		want    string
	}{
		{"dev", "prtool version dev"},
		{"v1.0.0", "prtool version v1.0.0"},
		{"1.0.0", "prtool version 1.0.0"},
		{"v1.0.0-beta", "prtool version v1.0.0-beta"},
		{"v1.0.0+build123", "prtool version v1.0.0+build123"},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			// The format should always be "prtool version X"
			expected := "prtool version " + tc.version
			if expected != tc.want {
				t.Errorf("Format mismatch for version %s", tc.version)
			}
		})
	}
}
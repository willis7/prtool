package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	// Redirect stdout to a buffer
	old := os.Stdout
	defer func() { os.Stdout = old }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the root command with the --version flag
	rootCmd.SetArgs([]string{"--version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("error executing command: %v", err)
	}

	// Close the writer and read the output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check if the output contains the version string
	expected := "dev"
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("expected output to contain %q, but got %q", expected, buf.String())
	}
}

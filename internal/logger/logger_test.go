package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogger_New(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		ci      bool
		logFile string
		wantErr bool
	}{
		{
			name:    "basic logger",
			verbose: true,
			ci:      false,
			logFile: "",
			wantErr: false,
		},
		{
			name:    "ci mode logger",
			verbose: false,
			ci:      true,
			logFile: "",
			wantErr: false,
		},
		{
			name:    "with log file",
			verbose: true,
			ci:      false,
			logFile: "/tmp/test.log",
			wantErr: false,
		},
		{
			name:    "invalid log file",
			verbose: true,
			ci:      false,
			logFile: "/invalid/path/test.log",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up log file if it exists
			if tt.logFile != "" && strings.HasPrefix(tt.logFile, "/tmp/") {
				defer func() { _ = os.Remove(tt.logFile) }() // Ignore error in test cleanup
			}

			logger, err := New(tt.verbose, tt.ci, tt.logFile)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if logger == nil {
				t.Fatal("Expected logger but got nil")
			}

			if logger.verbose != tt.verbose {
				t.Errorf("Expected verbose=%v, got %v", tt.verbose, logger.verbose)
			}

			if logger.ci != tt.ci {
				t.Errorf("Expected ci=%v, got %v", tt.ci, logger.ci)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		message  string
		args     []interface{}
		expected bool // whether output should be produced
	}{
		{
			name:     "verbose enabled",
			verbose:  true,
			message:  "Test message %s",
			args:     []interface{}{"arg"},
			expected: true,
		},
		{
			name:     "verbose disabled",
			verbose:  false,
			message:  "Test message %s",
			args:     []interface{}{"arg"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger, _ := New(tt.verbose, false, "")
			logger.Info(tt.message, tt.args...)

			// Restore stderr
			_ = w.Close()
			os.Stderr = oldStderr

			// Read captured output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			if tt.expected && output == "" {
				t.Error("Expected output but got none")
			}
			if !tt.expected && output != "" {
				t.Errorf("Expected no output but got: %s", output)
			}
		})
	}
}

func TestLogger_Progress(t *testing.T) {
	tests := []struct {
		name     string
		ci       bool
		message  string
		expected bool // whether output should be produced
	}{
		{
			name:     "normal mode",
			ci:       false,
			message:  "Processing...",
			expected: true,
		},
		{
			name:     "ci mode",
			ci:       true,
			message:  "Processing...",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger, _ := New(false, tt.ci, "")
			logger.Progress(tt.message)

			// Restore stderr
			_ = w.Close()
			os.Stderr = oldStderr

			// Read captured output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			if tt.expected && output == "" {
				t.Error("Expected output but got none")
			}
			if !tt.expected && output != "" {
				t.Errorf("Expected no output but got: %s", output)
			}
		})
	}
}

func TestLogger_Output(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := New(false, true, "")
	logger.Output("Test output %s", "message")

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) // Ignore error in test
	output := buf.String()

	expected := "Test output message"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', got: %s", expected, output)
	}
}

package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestLogger_Init(t *testing.T) {
	// Reset instance for testing
	instance = nil
	once = sync.Once{}

	// Test initialization without log file
	err := Init(true, false, "")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if instance == nil {
		t.Fatal("Logger instance not initialized")
	}

	if !instance.verbose {
		t.Error("Verbose mode not set")
	}

	if instance.ci {
		t.Error("CI mode should not be set")
	}

	// Close any open log file
	Close()
}

func TestLogger_InitWithLogFile(t *testing.T) {
	// Reset instance for testing
	instance = nil
	once = sync.Once{}

	// Create temp directory for log file
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	err := Init(true, true, logPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if instance.logFile == nil {
		t.Fatal("Log file not opened")
	}

	if instance.fileLogger == nil {
		t.Fatal("File logger not created")
	}

	// Write a test message
	Info("test message")

	// Close and check file contents
	Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "[INFO] test message") {
		t.Errorf("Log file doesn't contain expected message: %s", content)
	}
}

func TestLogger_Verbose(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		ci      bool
		want    string
	}{
		{
			name:    "verbose disabled",
			verbose: false,
			ci:      false,
			want:    "",
		},
		{
			name:    "verbose enabled",
			verbose: true,
			ci:      false,
			want:    "test verbose\n",
		},
		{
			name:    "verbose with CI",
			verbose: true,
			ci:      true,
			want:    "[VERBOSE] test verbose\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and init
			instance = nil
			once = sync.Once{}
			Init(tt.verbose, tt.ci, "")

			// Capture stderr
			old := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			Verbose("test verbose")

			w.Close()
			os.Stderr = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			if buf.String() != tt.want {
				t.Errorf("Verbose() output = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name string
		ci   bool
		want string
	}{
		{
			name: "normal mode",
			ci:   false,
			want: "test info\n",
		},
		{
			name: "CI mode",
			ci:   true,
			want: "[INFO] test info\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and init
			instance = nil
			once = sync.Once{}
			Init(false, tt.ci, "")

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			Info("test info")

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			if buf.String() != tt.want {
				t.Errorf("Info() output = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestLogger_Progress(t *testing.T) {
	tests := []struct {
		name string
		ci   bool
		want string
	}{
		{
			name: "normal mode",
			ci:   false,
			want: "test progress\n",
		},
		{
			name: "CI mode suppresses progress",
			ci:   true,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and init
			instance = nil
			once = sync.Once{}
			Init(false, tt.ci, "")

			// Capture stderr
			old := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			Progress("test progress")

			w.Close()
			os.Stderr = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			if buf.String() != tt.want {
				t.Errorf("Progress() output = %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestLogger_IsCI(t *testing.T) {
	// Reset and init with CI mode
	instance = nil
	once = sync.Once{}
	Init(false, true, "")

	if !IsCI() {
		t.Error("IsCI() = false, want true")
	}

	// Reset and init without CI mode
	instance = nil
	once = sync.Once{}
	Init(false, false, "")

	if IsCI() {
		t.Error("IsCI() = true, want false")
	}
}

func TestLogger_IsVerbose(t *testing.T) {
	// Reset and init with verbose mode
	instance = nil
	once = sync.Once{}
	Init(true, false, "")

	if !IsVerbose() {
		t.Error("IsVerbose() = false, want true")
	}

	// Reset and init without verbose mode
	instance = nil
	once = sync.Once{}
	Init(false, false, "")

	if IsVerbose() {
		t.Error("IsVerbose() = true, want false")
	}
}
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunInit(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(dir string) error
		wantErr     bool
		errContains string
	}{
		{
			name:    "creates config file successfully",
			wantErr: false,
		},
		{
			name: "fails if config already exists",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".prtool.yaml"), []byte("existing"), 0644)
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir, err := os.MkdirTemp("", "prtool-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			// Change to temp directory
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(oldDir)

			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}

			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(dir); err != nil {
					t.Fatal(err)
				}
			}

			// Run init command
			cmd := &cobra.Command{}
			err = runInit(cmd, []string{})

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runInit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("runInit() error = %v, want error containing %q", err, tt.errContains)
			}

			// If no error, check file was created
			if !tt.wantErr {
				configPath := filepath.Join(dir, ".prtool.yaml")
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Error("config file was not created")
				}

				// Read and verify content
				content, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("failed to read created config: %v", err)
				}

				// Check for key sections
				expectedSections := []string{
					"# prtool configuration file",
					"github:",
					"llm:",
					"output:",
					"token:",
					"provider:",
					"format:",
				}

				for _, section := range expectedSections {
					if !strings.Contains(string(content), section) {
						t.Errorf("config missing expected section: %s", section)
					}
				}
			}
		})
	}
}

func TestGenerateConfigTemplate(t *testing.T) {
	template := generateConfigTemplate()

	// Check that template is not empty
	if template == "" {
		t.Error("generateConfigTemplate() returned empty string")
	}

	// Check for required sections and fields
	requiredContent := []struct {
		content string
		desc    string
	}{
		{"# prtool configuration file", "header comment"},
		{"github:", "github section"},
		{"token:", "github token field"},
		{"organization:", "organization field"},
		{"team:", "team field"},
		{"user:", "user field"},
		{"repositories:", "repositories field"},
		{"llm:", "llm section"},
		{"provider:", "provider field"},
		{"model:", "model field"},
		{"api_key:", "api key field"},
		{"base_url:", "base url field"},
		{"temperature:", "temperature field"},
		{"max_tokens:", "max tokens field"},
		{"output:", "output section"},
		{"format:", "format field"},
		{"verbose:", "verbose field"},
		{"dry_run:", "dry run field"},
		{"time_range:", "time range field"},
		{"https://github.com/settings/tokens", "GitHub token URL"},
		{"https://platform.openai.com/api-keys", "OpenAI API key URL"},
	}

	for _, req := range requiredContent {
		if !strings.Contains(template, req.content) {
			t.Errorf("template missing %s: %q", req.desc, req.content)
		}
	}

	// Check that it's valid YAML by checking indentation
	lines := strings.Split(template, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		// Non-comment, non-empty lines should have proper indentation
		if !strings.HasPrefix(line, " ") && !strings.Contains(line, ":") {
			t.Errorf("line %d appears to have invalid YAML format: %q", i+1, line)
		}
	}
}

func TestWriteConfigFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "prtool-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "test.yaml")
	content := "test content"

	err = WriteConfigFile(filename, content)
	if err != nil {
		t.Fatalf("WriteConfigFile() error = %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}

	// Verify permissions
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0644 {
		t.Errorf("file permissions = %v, want 0644", info.Mode().Perm())
	}
}
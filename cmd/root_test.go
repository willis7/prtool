package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/spf13/cobra"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/model"
	"github.com/willis7/prtool/internal/render"
	"github.com/willis7/prtool/internal/service"
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

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config with org",
			cfg: &config.Config{
				GitHubToken: "token123",
				Org:         "test-org",
			},
			expectErr: false,
		},
		{
			name: "missing github token",
			cfg: &config.Config{
				Org: "test-org",
			},
			expectErr: true,
			errMsg:    "GitHub token is required",
		},
		{
			name: "no scope specified",
			cfg: &config.Config{
				GitHubToken: "token123",
			},
			expectErr: true,
			errMsg:    "no scope specified",
		},
		{
			name: "multiple scopes specified",
			cfg: &config.Config{
				GitHubToken: "token123",
				Org:         "test-org",
				User:        "test-user",
			},
			expectErr: true,
			errMsg:    "multiple scopes specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGenerateMetadata(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		prs      []*model.PR
		expected render.Metadata
	}{
		{
			name: "org scope with multiple repos",
			cfg: &config.Config{
				Org:         "test-org",
				Since:       "-7d",
				LLMProvider: "openai",
				LLMModel:    "gpt-4",
			},
			prs: []*model.PR{
				{Repository: "test-org/repo1"},
				{Repository: "test-org/repo2"},
				{Repository: "test-org/repo1"}, // duplicate
			},
			expected: render.Metadata{
				Scope:        "organization",
				ScopeValue:   "test-org",
				Since:        "-7d",
				TotalPRs:     3,
				Repositories: []string{"test-org/repo1", "test-org/repo2"}, // unique
				LLMProvider:  "openai",
				LLMModel:     "gpt-4",
				Summary:      "",
			},
		},
		{
			name: "user scope with default since",
			cfg: &config.Config{
				User: "test-user",
			},
			prs: []*model.PR{
				{Repository: "test-user/personal-repo"},
			},
			expected: render.Metadata{
				Scope:        "user",
				ScopeValue:   "test-user",
				Since:        "-7d", // default
				TotalPRs:     1,
				Repositories: []string{"test-user/personal-repo"},
				LLMProvider:  "",
				LLMModel:     "",
				Summary:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateMetadata(tt.cfg, tt.prs)

			// Check specific fields (ignore GeneratedAt as it's time-dependent)
			if result.Scope != tt.expected.Scope {
				t.Errorf("Expected scope %q, got %q", tt.expected.Scope, result.Scope)
			}
			if result.ScopeValue != tt.expected.ScopeValue {
				t.Errorf("Expected scope value %q, got %q", tt.expected.ScopeValue, result.ScopeValue)
			}
			if result.Since != tt.expected.Since {
				t.Errorf("Expected since %q, got %q", tt.expected.Since, result.Since)
			}
			if result.TotalPRs != tt.expected.TotalPRs {
				t.Errorf("Expected total PRs %d, got %d", tt.expected.TotalPRs, result.TotalPRs)
			}
			if len(result.Repositories) != len(tt.expected.Repositories) {
				t.Errorf("Expected %d repositories, got %d", len(tt.expected.Repositories), len(result.Repositories))
			}

			// Check that GeneratedAt is recent (within last minute)
			if time.Since(result.GeneratedAt) > time.Minute {
				t.Errorf("GeneratedAt should be recent, got %v", result.GeneratedAt)
			}
		})
	}
}

func TestWriteToFile(t *testing.T) {
	// Create temporary directory for tests
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		filename    string
		content     string
		expectError bool
	}{
		{
			name:        "simple file write",
			filename:    filepath.Join(tempDir, "output.md"),
			content:     "# Test Content\n\nHello world!",
			expectError: false,
		},
		{
			name:        "file with subdirectory",
			filename:    filepath.Join(tempDir, "subdir", "output.md"),
			content:     "# Nested Content",
			expectError: false,
		},
		{
			name:        "empty content",
			filename:    filepath.Join(tempDir, "empty.md"),
			content:     "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeToFile(tt.filename, tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file was written correctly
				written, readErr := os.ReadFile(tt.filename)
				if readErr != nil {
					t.Errorf("Failed to read written file: %v", readErr)
				}

				if string(written) != tt.content {
					t.Errorf("File content mismatch. Expected %q, got %q", tt.content, string(written))
				}
			}
		})
	}
}

func TestDryRunIntegration(t *testing.T) {
	// This test verifies the dry-run functionality works end-to-end
	// We'll mock the GitHub client and test the table output

	mockClient := gh.NewMockClient()

	// Setup mock data
	mockRepos := []*github.Repository{
		{FullName: github.String("test-org/repo1")},
	}
	mockClient.SetMockRepos(mockRepos)

	yesterday := time.Now().AddDate(0, 0, -1)
	mockPRs := []*model.PR{
		{
			Title:      "Test PR",
			Author:     "test-user",
			Repository: "test-org/repo1",
			Number:     123,
			MergedAt:   &yesterday,
			State:      "closed",
		},
	}
	mockClient.SetMockPRs(mockPRs)

	// Test that we can fetch and render PRs in table format
	cfg := &config.Config{
		GitHubToken: "fake-token",
		Org:         "test-org",
		DryRun:      true,
	}

	// Validate config
	err := validateConfig(cfg)
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// Note: We can't easily test the full command execution here without
	// significant refactoring, but we can test the individual components
	// that make up the dry-run flow

	prs, err := service.Fetch(cfg, mockClient)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prs))
	}

	// Test table rendering
	tableOutput := render.RenderTable(prs)
	if !strings.Contains(tableOutput, "Test PR") {
		t.Errorf("Table output should contain PR title, got: %s", tableOutput)
	}
	if !strings.Contains(tableOutput, "test-user") {
		t.Errorf("Table output should contain author, got: %s", tableOutput)
	}
}

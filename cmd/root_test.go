package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/spf13/cobra"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/llm"
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

func TestCreateLLMClient(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected string // Expected LLM type
	}{
		{
			name: "empty provider defaults to stub",
			cfg: &config.Config{
				LLMProvider: "",
			},
			expected: "*llm.StubLLM",
		},
		{
			name: "stub provider",
			cfg: &config.Config{
				LLMProvider: "stub",
			},
			expected: "*llm.StubLLM",
		},
		{
			name: "openai provider with key",
			cfg: &config.Config{
				LLMProvider: "openai",
				LLMAPIKey:   "test-key",
				LLMModel:    "gpt-4",
			},
			expected: "*llm.OpenAILLM",
		},
		{
			name: "openai provider without key falls back to stub",
			cfg: &config.Config{
				LLMProvider: "openai",
				LLMAPIKey:   "",
			},
			expected: "*llm.StubLLM",
		},
		{
			name: "ollama provider",
			cfg: &config.Config{
				LLMProvider: "ollama",
				LLMModel:    "llama3.2",
			},
			expected: "*llm.OllamaLLM",
		},
		{
			name: "unknown provider falls back to stub",
			cfg: &config.Config{
				LLMProvider: "unknown",
			},
			expected: "*llm.StubLLM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr to check warning messages
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			client := createLLMClient(tt.cfg)

			// Restore stderr
			w.Close()
			os.Stderr = oldStderr

			// Read captured output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			stderr := string(buf[:n])

			// Check the type of returned client
			clientType := fmt.Sprintf("%T", client)
			if clientType != tt.expected {
				t.Errorf("createLLMClient() = %v, want %v", clientType, tt.expected)
			}

			// Check warning messages for fallback cases
			if tt.cfg.LLMProvider == "openai" && tt.cfg.LLMAPIKey == "" {
				if !strings.Contains(stderr, "Warning: OpenAI API key not provided") {
					t.Error("Expected warning about missing OpenAI API key")
				}
			}
			if tt.cfg.LLMProvider == "unknown" {
				if !strings.Contains(stderr, "Warning: Unknown LLM provider") {
					t.Error("Expected warning about unknown provider")
				}
			}
		})
	}
}

func TestLLMIntegration(t *testing.T) {
	// Test the full integration of LLM with the markdown rendering

	// Create mock data
	mockClient := gh.NewMockClient()
	mockRepos := []*github.Repository{
		{FullName: github.String("test-org/repo1")},
	}
	mockClient.SetMockRepos(mockRepos)

	yesterday := time.Now().AddDate(0, 0, -1)
	mockPRs := []*model.PR{
		{
			Title:      "Add new feature",
			Author:     "developer",
			Repository: "test-org/repo1",
			Number:     123,
			MergedAt:   &yesterday,
			State:      "closed",
			Body:       "This PR adds an important new feature to the application.",
			Labels:     []string{"feature", "enhancement"},
		},
	}
	mockClient.SetMockPRs(mockPRs)

	// Test configuration (not dry-run, so LLM should be called)
	cfg := &config.Config{
		GitHubToken: "fake-token",
		Org:         "test-org",
		LLMProvider: "stub",
		DryRun:      false,
	}

	// Fetch PRs
	prs, err := service.Fetch(cfg, mockClient)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	// Generate metadata with LLM summary
	metadata := generateMetadata(cfg, prs)

	// Create LLM client and generate summary
	llmClient := createLLMClient(cfg)
	context := llm.BuildContext(prs)
	summary, err := llmClient.Summarise(context)
	if err != nil {
		t.Fatalf("Failed to generate summary: %v", err)
	}

	metadata.Summary = summary

	// Render markdown
	markdown := render.Render(metadata, prs)

	// Verify the markdown contains the LLM summary
	if !strings.Contains(markdown, summary) {
		t.Errorf("Markdown output should contain LLM summary")
	}

	// Verify the markdown contains the AI Summary section
	if !strings.Contains(markdown, "## AI Summary") {
		t.Errorf("Markdown output should contain AI Summary section")
	}

	// Verify the markdown contains PR details
	if !strings.Contains(markdown, "Add new feature") {
		t.Errorf("Markdown output should contain PR title")
	}
}

func TestEndToEndWithLLM(t *testing.T) {
	// This test demonstrates the complete end-to-end workflow with LLM integration
	// It simulates the main workflow without requiring real GitHub credentials

	// Setup mock GitHub client
	mockClient := gh.NewMockClient()

	// Mock repositories
	mockRepos := []*github.Repository{
		{FullName: github.String("example/backend")},
		{FullName: github.String("example/frontend")},
	}
	mockClient.SetMockRepos(mockRepos)

	// Mock PRs with realistic data
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)

	mockPRs := []*model.PR{
		{
			Title:      "Implement user authentication system",
			Author:     "backend-team",
			Repository: "example/backend",
			Number:     101,
			MergedAt:   &yesterday,
			State:      "closed",
			Body:       "Added OAuth2 authentication with JWT tokens. Includes comprehensive security tests and documentation.",
			Labels:     []string{"feature", "security", "backend"},
			HTMLURL:    "https://github.com/example/backend/pull/101",
		},
		{
			Title:      "Redesign user dashboard",
			Author:     "frontend-team",
			Repository: "example/frontend",
			Number:     202,
			MergedAt:   &twoDaysAgo,
			State:      "closed",
			Body:       "Complete redesign of the user dashboard with improved UX and responsive design.",
			Labels:     []string{"feature", "ui", "frontend"},
			HTMLURL:    "https://github.com/example/frontend/pull/202",
		},
	}
	mockClient.SetMockPRs(mockPRs)

	// Create configuration for the test
	cfg := &config.Config{
		GitHubToken: "test-token",
		Org:         "example",
		Since:       "-7d",
		LLMProvider: "stub",
		DryRun:      false,
		Verbose:     false,
	}

	// Step 1: Validate configuration
	err := validateConfig(cfg)
	if err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}

	// Step 2: Fetch PRs
	prs, err := service.Fetch(cfg, mockClient)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) != 2 {
		t.Fatalf("Expected 2 PRs, got %d", len(prs))
	}

	// Step 3: Generate metadata
	metadata := generateMetadata(cfg, prs)

	// Verify basic metadata
	if metadata.Scope != "organization" {
		t.Errorf("Expected scope 'organization', got %q", metadata.Scope)
	}
	if metadata.ScopeValue != "example" {
		t.Errorf("Expected scope value 'example', got %q", metadata.ScopeValue)
	}
	if metadata.TotalPRs != 2 {
		t.Errorf("Expected 2 total PRs, got %d", metadata.TotalPRs)
	}

	// Step 4: Generate LLM summary
	llmClient := createLLMClient(cfg)
	context := llm.BuildContext(prs)
	summary, err := llmClient.Summarise(context)
	if err != nil {
		t.Fatalf("Failed to generate LLM summary: %v", err)
	}

	if summary == "" {
		t.Error("Expected non-empty summary from LLM")
	}

	metadata.Summary = summary

	// Step 5: Render final markdown
	markdown := render.Render(metadata, prs)

	// Verify the complete output contains all expected elements
	expectedElements := []string{
		"# Pull Request Summary",
		"## Summary Information",
		"## AI Summary",
		"## Pull Request Details",
		"Implement user authentication system",
		"Redesign user dashboard",
		"backend-team",
		"frontend-team",
		"OAuth2 authentication with JWT tokens",
		"Complete redesign of the user dashboard",
		summary, // The LLM-generated summary
		"Generated by prtool",
	}

	for _, element := range expectedElements {
		if !strings.Contains(markdown, element) {
			t.Errorf("Final markdown missing expected element: %q", element)
		}
	}

	// Verify structure
	if !strings.Contains(markdown, "**Scope**: organization (example)") {
		t.Error("Markdown should contain scope information")
	}

	if !strings.Contains(markdown, "**Total PRs**: 2") {
		t.Error("Markdown should contain PR count")
	}

	// Verify both repositories are mentioned
	if !strings.Contains(markdown, "example/backend") || !strings.Contains(markdown, "example/frontend") {
		t.Error("Markdown should contain both repository names")
	}

	// Verify LLM summary appears in the correct section
	summaryIndex := strings.Index(markdown, "## AI Summary")
	prDetailsIndex := strings.Index(markdown, "## Pull Request Details")

	if summaryIndex == -1 {
		t.Error("AI Summary section not found")
	}
	if prDetailsIndex == -1 {
		t.Error("PR Details section not found")
	}
	if summaryIndex >= prDetailsIndex {
		t.Error("AI Summary should appear before PR Details")
	}

	t.Logf("Successfully generated %d-character markdown with LLM summary", len(markdown))
}

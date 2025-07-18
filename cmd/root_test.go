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
	"github.com/willis7/prtool/internal/logger"
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
				_ = cmd.Help() // Ignore error in test
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
			_ = w.Close() // Ignore error in test cleanup
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

// MockVersionChecker for testing
type MockVersionChecker struct {
	release *GitHubRelease
	err     error
}

func (m *MockVersionChecker) GetLatestRelease() (*GitHubRelease, error) {
	return m.release, m.err
}

func TestVersionCheck(t *testing.T) {
	// Save original version checker
	originalChecker := versionChecker
	defer func() { versionChecker = originalChecker }()

	tests := []struct {
		name           string
		currentVersion string
		mockRelease    *GitHubRelease
		mockError      error
		expectError    bool
		expectedOutput []string
	}{
		{
			name:           "latest version available",
			currentVersion: "v1.0.0",
			mockRelease: &GitHubRelease{
				TagName: "v1.1.0",
				Name:    "Version 1.1.0",
				HTMLURL: "https://github.com/willis7/prtool/releases/tag/v1.1.0",
			},
			expectError: false,
			expectedOutput: []string{
				"Current version: v1.0.0",
				"Latest version: v1.1.0",
				"A newer version is available",
				"https://github.com/willis7/prtool/releases/tag/v1.1.0",
			},
		},
		{
			name:           "up to date version",
			currentVersion: "v1.0.0",
			mockRelease: &GitHubRelease{
				TagName: "v1.0.0",
				Name:    "Version 1.0.0",
				HTMLURL: "https://github.com/willis7/prtool/releases/tag/v1.0.0",
			},
			expectError: false,
			expectedOutput: []string{
				"Current version: v1.0.0",
				"Latest version: v1.0.0",
				"You are running the latest version!",
			},
		},
		{
			name:           "development version",
			currentVersion: "dev",
			mockRelease: &GitHubRelease{
				TagName: "v1.0.0",
				Name:    "Version 1.0.0",
				HTMLURL: "https://github.com/willis7/prtool/releases/tag/v1.0.0",
			},
			expectError: false,
			expectedOutput: []string{
				"Current version: dev",
				"Latest version: v1.0.0",
				"You are running a development version",
			},
		},
		{
			name:           "API error",
			currentVersion: "v1.0.0",
			mockRelease:    nil,
			mockError:      fmt.Errorf("API rate limit exceeded"),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock
			versionChecker = &MockVersionChecker{
				release: tt.mockRelease,
				err:     tt.mockError,
			}

			// Set current version
			originalVersion := version
			version = tt.currentVersion
			defer func() { version = originalVersion }()

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run version check
			err := checkLatestVersion()

			// Restore stdout
			_ = w.Close() // Ignore error in test cleanup
			os.Stdout = oldStdout

			// Read captured output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output contains expected strings
			if !tt.expectError {
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', got: %s", expected, output)
					}
				}
			}
		})
	}
}

func TestVersionCheckCommand(t *testing.T) {
	// Save original version checker
	originalChecker := versionChecker
	defer func() { versionChecker = originalChecker }()

	// Set up mock
	versionChecker = &MockVersionChecker{
		release: &GitHubRelease{
			TagName: "v1.0.0",
			Name:    "Version 1.0.0",
			HTMLURL: "https://github.com/willis7/prtool/releases/tag/v1.0.0",
		},
	}

	// Create command with version-check flag
	cmd := rootCmd
	cmd.SetArgs([]string{"--version-check"})

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	err := cmd.Execute()

	// Restore stdout
	_ = w.Close() // Ignore error in test cleanup
	os.Stdout = oldStdout

	// Check that it executed without error
	if err != nil {
		t.Errorf("version-check command failed: %v", err)
	}

	// Read output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify output contains version check information
	if !strings.Contains(output, "Current version:") {
		t.Error("Expected version check output")
	}
}

func TestRealVersionChecker(t *testing.T) {
	// This is an integration test that can be skipped if GitHub is unreachable
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	checker := NewRealVersionChecker()

	// Test timeout is reasonable
	if checker.client.Timeout != 10*time.Second {
		t.Errorf("Expected 10 second timeout, got %v", checker.client.Timeout)
	}

	// We can't easily test the actual API call without hitting GitHub,
	// but we can test that the method exists and returns the right types
	_, err := checker.GetLatestRelease()
	// Note: This might fail if GitHub is unreachable or rate-limited,
	// but it tests the method signature and basic functionality
	if err != nil {
		t.Logf("GitHub API call failed (this may be expected): %v", err)
	}
}

func TestCIMode(t *testing.T) {
	// Save original version checker
	originalChecker := versionChecker
	defer func() { versionChecker = originalChecker }()

	// Set up mock
	versionChecker = &MockVersionChecker{
		release: &GitHubRelease{
			TagName: "v1.0.0",
			Name:    "Version 1.0.0",
			HTMLURL: "https://github.com/willis7/prtool/releases/tag/v1.0.0",
		},
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "version with CI flag",
			args: []string{"--ci", "--version"},
		},
		{
			name: "version-check with CI flag",
			args: []string{"--ci", "--version-check"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			cmd.SetArgs(tt.args)

			// Capture output
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Execute command
			err := cmd.Execute()

			// Restore streams
			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// Read output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			// CI mode should not fail for these commands
			if err != nil {
				t.Errorf("CI mode command failed: %v", err)
			}

			// Output should be produced
			if output == "" {
				t.Error("Expected output in CI mode")
			}

			t.Logf("CI mode output: %s", output)
		})
	}
}

func TestLoggerIntegration(t *testing.T) {
	tests := []struct {
		name         string
		verbose      bool
		ci           bool
		logFile      string
		expectOutput bool
	}{
		{
			name:         "verbose mode",
			verbose:      true,
			ci:           false,
			expectOutput: true,
		},
		{
			name:         "ci mode",
			verbose:      false,
			ci:           true,
			expectOutput: false, // Progress should be suppressed
		},
		{
			name:         "normal mode",
			verbose:      false,
			ci:           false,
			expectOutput: true, // Progress should be shown
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.New(tt.verbose, tt.ci, tt.logFile)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Capture stderr for progress messages
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			log.Progress("Test progress message")

			// Restore stderr
			w.Close()
			os.Stderr = oldStderr

			// Read captured output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			hasOutput := len(output) > 0
			if hasOutput != tt.expectOutput {
				t.Errorf("Expected output=%v, got output=%v (output: %q)", tt.expectOutput, hasOutput, output)
			}
		})
	}
}

func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectStart string
	}{
		{
			name:        "bash completion",
			args:        []string{"completion", "bash"},
			expectError: false,
			expectStart: "# bash completion for prtool",
		},
		{
			name:        "zsh completion",
			args:        []string{"completion", "zsh"},
			expectError: false,
			expectStart: "#compdef prtool",
		},
		{
			name:        "fish completion",
			args:        []string{"completion", "fish"},
			expectError: false,
			expectStart: "# fish completion for prtool",
		},
		{
			name:        "powershell completion",
			args:        []string{"completion", "powershell"},
			expectError: false,
			expectStart: "# powershell completion for prtool",
		},
		{
			name:        "invalid shell",
			args:        []string{"completion", "invalid"},
			expectError: true,
			expectStart: "",
		},
		{
			name:        "no shell specified",
			args:        []string{"completion"},
			expectError: true,
			expectStart: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			var stdout bytes.Buffer

			// Create a new root command for each test to avoid state pollution
			cmd := &cobra.Command{
				Use:   "prtool",
				Short: "A CLI tool for summarizing GitHub pull requests",
			}
			cmd.AddCommand(completionCmd)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stdout)
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check output if no error expected
			if !tt.expectError && tt.expectStart != "" {
				output := stdout.String()
				if output == "" {
					t.Errorf("Expected output but got empty string")
				} else if !strings.HasPrefix(output, tt.expectStart) {
					t.Errorf("Expected output to start with %q, got first 100 chars: %q", tt.expectStart, output[:min(len(output), 100)])
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

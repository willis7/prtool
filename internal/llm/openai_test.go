package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prtool/internal/model"
)

func TestNewOpenAILLM(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing API key",
			config: Config{
				Provider: "openai",
				APIKey:   "",
			},
			wantErr: true,
			errMsg:  "OpenAI API key is required",
		},
		{
			name: "valid config with defaults",
			config: Config{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
		{
			name: "custom model and settings",
			config: Config{
				Provider:    "openai",
				APIKey:      "test-key",
				Model:       "gpt-4",
				Temperature: 0.5,
				MaxTokens:   1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := NewOpenAILLM(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAILLM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("NewOpenAILLM() error = %v, want error containing %q", err, tt.errMsg)
			}
			if !tt.wantErr && llm == nil {
				t.Error("NewOpenAILLM() returned nil without error")
			}
		})
	}
}

func TestOpenAILLM_Name(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		{"gpt-3.5-turbo", "openai (gpt-3.5-turbo)"},
		{"gpt-4", "openai (gpt-4)"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			llm := &OpenAILLM{model: tt.model}
			if got := llm.Name(); got != tt.expected {
				t.Errorf("Name() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildPRPrompt(t *testing.T) {
	pr := model.PR{
		Repository:  "owner/repo",
		Number:      123,
		Title:       "Add new feature",
		Description: "This PR adds feature X",
		Author:      "alice",
		Labels:      []string{"enhancement", "feature"},
	}

	prompt := buildPRPrompt(pr)

	// Check that prompt contains key information
	expectedParts := []string{
		"Title: Add new feature",
		"Author: @alice",
		"Repository: owner/repo",
		"Description:",
		"This PR adds feature X",
		"Labels: enhancement, feature",
		"concise summary",
	}

	for _, part := range expectedParts {
		if !strings.Contains(prompt, part) {
			t.Errorf("Prompt missing expected part: %q", part)
		}
	}
}

// TestOpenAILLM_Summarize_Smoke is a smoke test that requires a real OpenAI API key
func TestOpenAILLM_Summarize_Smoke(t *testing.T) {
	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI smoke test: OPENAI_KEY environment variable not set")
	}

	config := Config{
		Provider:    "openai",
		APIKey:      apiKey,
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   200,
	}

	llm, err := NewOpenAILLM(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI LLM: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pr := model.PR{
		Repository:  "test/repo",
		Number:      999,
		Title:       "Implement user authentication system",
		Description: "This PR adds a complete user authentication system with JWT tokens, password hashing, and role-based access control.",
		Author:      "testuser",
		Labels:      []string{"security", "feature"},
		URL:         "https://github.com/test/repo/pull/999",
		MergedAt:    time.Now(),
	}

	summary, err := llm.Summarize(ctx, pr)
	if err != nil {
		t.Fatalf("Summarize() error: %v", err)
	}

	// Basic validation of the summary
	if summary == "" {
		t.Error("Summarize() returned empty summary")
	}

	if len(summary) < 50 {
		t.Errorf("Summary seems too short: %q", summary)
	}

	t.Logf("Generated summary: %s", summary)
}

// TestOpenAILLM_SummarizeBatch_Smoke tests batch summarization with real API
func TestOpenAILLM_SummarizeBatch_Smoke(t *testing.T) {
	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI batch smoke test: OPENAI_KEY environment variable not set")
	}

	config := Config{
		Provider:    "openai",
		APIKey:      apiKey,
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   150,
	}

	llm, err := NewOpenAILLM(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI LLM: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prs := []model.PR{
		{
			Repository:  "test/repo",
			Number:      1,
			Title:       "Fix memory leak in cache implementation",
			Description: "This PR fixes a critical memory leak that was causing OOM errors in production.",
			Author:      "developer1",
			URL:         "https://github.com/test/repo/pull/1",
			MergedAt:    time.Now(),
		},
		{
			Repository:  "test/repo",
			Number:      2,
			Title:       "Add comprehensive API documentation",
			Description: "Adds OpenAPI/Swagger documentation for all REST endpoints.",
			Author:      "developer2",
			URL:         "https://github.com/test/repo/pull/2",
			MergedAt:    time.Now(),
		},
	}

	summaries, err := llm.SummarizeBatch(ctx, prs)
	if err != nil {
		t.Fatalf("SummarizeBatch() error: %v", err)
	}

	// Verify we got summaries for all PRs
	if len(summaries) != len(prs) {
		t.Errorf("Expected %d summaries, got %d", len(prs), len(summaries))
	}

	for _, pr := range prs {
		summary, ok := summaries[pr.URL]
		if !ok {
			t.Errorf("Missing summary for PR %s", pr.URL)
			continue
		}

		if summary == "" {
			t.Errorf("Empty summary for PR %s", pr.URL)
		}

		t.Logf("Summary for PR #%d: %s", pr.Number, summary)
	}
}
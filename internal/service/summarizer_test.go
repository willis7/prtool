package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/model"
)

func TestNewSummarizer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "stub provider",
			config: &config.Config{
				LLM: config.LLMConfig{
					Provider: "stub",
				},
			},
			wantErr: false,
		},
		{
			name: "empty provider defaults to stub",
			config: &config.Config{
				LLM: config.LLMConfig{
					Provider: "",
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported provider",
			config: &config.Config{
				LLM: config.LLMConfig{
					Provider: "unsupported",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summarizer, err := NewSummarizer(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSummarizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && summarizer == nil {
				t.Error("NewSummarizer() returned nil without error")
			}
		})
	}
}

func TestSummarizePRs(t *testing.T) {
	// Create summarizer with stub LLM
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "stub",
		},
	}
	summarizer, err := NewSummarizer(cfg)
	if err != nil {
		t.Fatalf("Failed to create summarizer: %v", err)
	}

	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name    string
		prs     []model.PR
		verbose bool
		wantErr bool
	}{
		{
			name: "summarize multiple PRs",
			prs: []model.PR{
				{
					Repository:  "owner/repo",
					Number:      123,
					Title:       "Add new feature",
					Description: "This adds feature X",
					Author:      "alice",
					URL:         "https://github.com/owner/repo/pull/123",
					MergedAt:    now,
				},
				{
					Repository:  "owner/repo",
					Number:      124,
					Title:       "Fix critical bug",
					Description: "This fixes bug Y",
					Author:      "bob",
					URL:         "https://github.com/owner/repo/pull/124",
					MergedAt:    now.Add(-1 * time.Hour),
				},
			},
			verbose: false,
			wantErr: false,
		},
		{
			name:    "empty PR list",
			prs:     []model.PR{},
			verbose: false,
			wantErr: false,
		},
		{
			name: "verbose mode",
			prs: []model.PR{
				{
					Repository: "owner/repo",
					Number:     125,
					Title:      "Refactor code",
					Author:     "charlie",
					URL:        "https://github.com/owner/repo/pull/125",
					MergedAt:   now,
				},
			},
			verbose: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of PRs to avoid modifying test data
			prs := make([]model.PR, len(tt.prs))
			copy(prs, tt.prs)

			result, err := summarizer.SummarizePRs(ctx, prs, tt.verbose)
			if (err != nil) != tt.wantErr {
				t.Errorf("SummarizePRs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that all PRs have summaries
				for i, pr := range result {
					if pr.Summary == "" {
						t.Errorf("PR %d has empty summary", i)
					}
					// Verify summary contains PR title and author
					if !strings.Contains(pr.Summary, pr.Title) {
						t.Errorf("Summary for PR %d should contain title %q", i, pr.Title)
					}
					if !strings.Contains(pr.Summary, "@"+pr.Author) {
						t.Errorf("Summary for PR %d should contain author @%s", i, pr.Author)
					}
				}
			}
		})
	}
}

func TestSummarizePRs_Integration(t *testing.T) {
	// This test verifies that summaries are properly added to PRs
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "stub",
		},
	}
	summarizer, err := NewSummarizer(cfg)
	if err != nil {
		t.Fatalf("Failed to create summarizer: %v", err)
	}

	ctx := context.Background()
	
	originalPR := model.PR{
		Repository:  "owner/repo",
		Number:      999,
		Title:       "Important update",
		Description: "Original description",
		Author:      "developer",
		URL:         "https://github.com/owner/repo/pull/999",
		MergedAt:    time.Now(),
		Summary:     "", // Initially empty
	}

	prs := []model.PR{originalPR}
	result, err := summarizer.SummarizePRs(ctx, prs, false)
	if err != nil {
		t.Fatalf("SummarizePRs() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(result))
	}

	// Verify original fields are preserved
	if result[0].Repository != originalPR.Repository {
		t.Errorf("Repository changed: got %q, want %q", result[0].Repository, originalPR.Repository)
	}
	if result[0].Number != originalPR.Number {
		t.Errorf("Number changed: got %d, want %d", result[0].Number, originalPR.Number)
	}
	if result[0].Description != originalPR.Description {
		t.Errorf("Description changed: got %q, want %q", result[0].Description, originalPR.Description)
	}

	// Verify summary was added
	if result[0].Summary == "" {
		t.Error("Summary should not be empty after summarization")
	}
	if result[0].Summary == originalPR.Summary {
		t.Error("Summary should have been updated")
	}
}
package render

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/prtool/internal/model"
)

func TestRender_WithLLMSummaries(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		meta Metadata
		prs  []model.PR
		want []string // Expected strings to be present in output
	}{
		{
			name: "PR with LLM summary",
			meta: Metadata{
				GeneratedAt:  fixedTime,
				TimeRange:    "-7d",
				TotalPRs:     1,
				Repositories: []string{"owner/repo"},
				LLMProvider:  "stub",
				LLMModel:     "test-model",
			},
			prs: []model.PR{
				{
					Repository:  "owner/repo",
					Number:      123,
					Title:       "Add authentication feature",
					Description: "Original description",
					Author:      "alice",
					URL:         "https://github.com/owner/repo/pull/123",
					MergedAt:    fixedTime.AddDate(0, 0, -1),
					Summary:     "This PR Add authentication feature by @alice introduces significant changes to improve the codebase. Key improvements include enhanced performance, better error handling, and increased test coverage. The changes are well-structured and follow best practices.",
				},
			},
			want: []string{
				"**Summary**:",
				"This PR Add authentication feature by @alice introduces significant changes",
				"enhanced performance",
				"better error handling",
				"increased test coverage",
				"**LLM**: stub (test-model)",
			},
		},
		{
			name: "PR without summary falls back to description",
			meta: Metadata{
				GeneratedAt:  fixedTime,
				TimeRange:    "-7d",
				TotalPRs:     1,
				Repositories: []string{"owner/repo"},
			},
			prs: []model.PR{
				{
					Repository:  "owner/repo",
					Number:      124,
					Title:       "Fix bug",
					Description: "This PR fixes a critical bug in the system",
					Author:      "bob",
					URL:         "https://github.com/owner/repo/pull/124",
					MergedAt:    fixedTime.AddDate(0, 0, -2),
					Summary:     "", // No summary
				},
			},
			want: []string{
				"**Description**:",
				"> This PR fixes a critical bug in the system",
				"Fix bug", // Title should be present
			},
		},
		{
			name: "Mixed PRs with and without summaries",
			meta: Metadata{
				GeneratedAt:  fixedTime,
				TimeRange:    "-7d",
				TotalPRs:     2,
				Repositories: []string{"org/project"},
				LLMProvider:  "openai",
				LLMModel:     "gpt-4",
			},
			prs: []model.PR{
				{
					Repository:  "org/project",
					Number:      200,
					Title:       "Refactor database layer",
					Description: "Refactoring for better performance",
					Author:      "charlie",
					URL:         "https://github.com/org/project/pull/200",
					MergedAt:    fixedTime.AddDate(0, 0, -1),
					Summary:     "This PR Refactor database layer by @charlie introduces significant changes to improve the codebase. The refactoring focuses on query optimization and connection pooling.",
				},
				{
					Repository:  "org/project",
					Number:      201,
					Title:       "Update dependencies",
					Description: "Routine dependency updates",
					Author:      "dave",
					URL:         "https://github.com/org/project/pull/201",
					MergedAt:    fixedTime.AddDate(0, 0, -3),
					Summary:     "", // No summary
				},
			},
			want: []string{
				"Refactor database layer",
				"**Summary**:",
				"query optimization and connection pooling",
				"Update dependencies",
				"**Description**:",
				"> Routine dependency updates",
				"**LLM**: openai (gpt-4)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := Render(tt.meta, tt.prs)

			// Check that all expected strings are present
			for _, expected := range tt.want {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it doesn't.\nFull output:\n%s", expected, output)
				}
			}

			// Verify Summary takes precedence over Description when present
			for _, pr := range tt.prs {
				if pr.Summary != "" {
					// Should show Summary, not Description
					if strings.Contains(output, "**Description**:") && strings.Contains(output, pr.Description) {
						t.Errorf("PR %d: Should not show Description when Summary is present", pr.Number)
					}
				}
			}
		})
	}
}

func TestRenderDryRun_NoSummaries(t *testing.T) {
	// Verify that dry-run mode doesn't show summaries
	prs := []model.PR{
		{
			Repository: "owner/repo",
			Number:     123,
			Title:      "Feature with summary",
			Author:     "alice",
			MergedAt:   time.Now(),
			Summary:    "This is an AI-generated summary that should not appear in dry-run",
		},
	}

	output := RenderDryRun(prs)

	// Should not contain the summary
	if strings.Contains(output, "This is an AI-generated summary") {
		t.Error("Dry-run output should not contain PR summaries")
	}

	// Should still contain basic info
	if !strings.Contains(output, "Feature with summary") {
		t.Error("Dry-run output should contain PR title")
	}
	if !strings.Contains(output, "alice") {
		t.Error("Dry-run output should contain PR author")
	}
}
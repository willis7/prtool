package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/willis7/prtool/internal/model"
)

func TestRender(t *testing.T) {
	// Fixed time for consistent testing
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mergedTime1 := time.Date(2024, 1, 14, 15, 20, 0, 0, time.UTC)
	mergedTime2 := time.Date(2024, 1, 13, 9, 45, 0, 0, time.UTC)

	tests := []struct {
		name       string
		metadata   Metadata
		prs        []*model.PR
		goldenFile string
	}{
		{
			name: "full_report_with_multiple_prs",
			metadata: Metadata{
				GeneratedAt:  fixedTime,
				Scope:        "organization",
				ScopeValue:   "acme-corp",
				Since:        "-7d",
				TotalPRs:     2,
				Repositories: []string{"acme-corp/web-app", "acme-corp/api-service"},
				LLMProvider:  "openai",
				LLMModel:     "gpt-4",
				Summary:      "This week saw significant improvements to the authentication system and API performance optimizations.",
			},
			prs: []*model.PR{
				{
					Title:      "Add OAuth2 authentication support",
					Body:       "This PR implements OAuth2 authentication to replace the legacy token-based system. Includes comprehensive tests and documentation updates.",
					Author:     "alice-dev",
					Repository: "acme-corp/web-app",
					Number:     123,
					MergedAt:   &mergedTime1,
					HTMLURL:    "https://github.com/acme-corp/web-app/pull/123",
					Labels:     []string{"feature", "security", "breaking-change"},
					FilePaths:  []string{"src/auth/oauth.go", "src/auth/oauth_test.go", "docs/auth.md"},
					State:      "closed",
				},
				{
					Title:      "Optimize database queries for better performance",
					Body:       "Refactored several slow database queries by adding proper indexes and optimizing JOIN operations. Performance tests show 40% improvement in response times.",
					Author:     "bob-engineer",
					Repository: "acme-corp/api-service",
					Number:     456,
					MergedAt:   &mergedTime2,
					HTMLURL:    "https://github.com/acme-corp/api-service/pull/456",
					Labels:     []string{"performance", "database"},
					FilePaths:  []string{"internal/db/queries.go", "internal/db/indexes.sql"},
					State:      "closed",
				},
			},
			goldenFile: "full_report.md",
		},
		{
			name: "minimal_report_no_llm",
			metadata: Metadata{
				GeneratedAt:  fixedTime,
				Scope:        "user",
				ScopeValue:   "john-doe",
				Since:        "-1d",
				TotalPRs:     1,
				Repositories: []string{"john-doe/personal-project"},
			},
			prs: []*model.PR{
				{
					Title:      "Fix typo in README",
					Author:     "john-doe",
					Repository: "john-doe/personal-project",
					Number:     1,
					MergedAt:   &mergedTime1,
					State:      "closed",
				},
			},
			goldenFile: "minimal_report.md",
		},
		{
			name: "empty_report",
			metadata: Metadata{
				GeneratedAt:  fixedTime,
				Scope:        "repository",
				ScopeValue:   "example/empty-repo",
				Since:        "-30d",
				TotalPRs:     0,
				Repositories: []string{"example/empty-repo"},
			},
			prs:        []*model.PR{},
			goldenFile: "empty_report.md",
		},
		{
			name: "long_description_truncation",
			metadata: Metadata{
				GeneratedAt:  fixedTime,
				Scope:        "organization",
				ScopeValue:   "test-org",
				Since:        "-7d",
				TotalPRs:     1,
				Repositories: []string{"test-org/repo"},
			},
			prs: []*model.PR{
				{
					Title:      "PR with very long description",
					Body:       strings.Repeat("This is a very long description that should be truncated when it exceeds the maximum length. ", 10) + "This part should not appear in the output.",
					Author:     "verbose-dev",
					Repository: "test-org/repo",
					Number:     999,
					MergedAt:   &mergedTime1,
					State:      "closed",
				},
			},
			goldenFile: "truncated_description.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the markdown
			result := Render(tt.metadata, tt.prs)

			// Compare with golden file
			if shouldUpdateGolden() {
				updateGoldenFile(t, tt.goldenFile, result)
			} else {
				compareWithGolden(t, tt.goldenFile, result)
			}
		})
	}
}

func TestRenderTable(t *testing.T) {
	fixedTime := time.Date(2024, 1, 14, 15, 20, 0, 0, time.UTC)

	tests := []struct {
		name     string
		prs      []*model.PR
		expected string
	}{
		{
			name:     "empty_list",
			prs:      []*model.PR{},
			expected: "No pull requests found for the specified criteria.\n",
		},
		{
			name: "single_pr",
			prs: []*model.PR{
				{
					Title:      "Fix critical bug",
					Author:     "alice",
					Repository: "org/repo",
					MergedAt:   &fixedTime,
				},
			},
			expected: `Found Pull Requests:

| # | Title | Author | Repository | Merged At |
|---|-------|--------|------------|----------|
| 1 | Fix critical bug | alice | org/repo | 2024-01-14 |

Total: 1 pull request(s)
`,
		},
		{
			name: "multiple_prs_with_truncation",
			prs: []*model.PR{
				{
					Title:      "This is a very long title that should be truncated because it exceeds the maximum length",
					Author:     "very-long-username-that-should-be-truncated",
					Repository: "organization-with-long-name/repository-with-very-long-name",
					MergedAt:   &fixedTime,
				},
				{
					Title:      "Short title",
					Author:     "bob",
					Repository: "org/repo",
					MergedAt:   nil, // Test nil MergedAt
				},
			},
			expected: `Found Pull Requests:

| # | Title | Author | Repository | Merged At |
|---|-------|--------|------------|----------|
| 1 | This is a very long title that should... | very-long-us... | organization-with... | 2024-01-14 |
| 2 | Short title | bob | org/repo | N/A |

Total: 2 pull request(s)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTable(tt.prs)
			if result != tt.expected {
				t.Errorf("RenderTable() result mismatch\nExpected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestMetadata_Validation(t *testing.T) {
	// Test that metadata fields are properly formatted
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	meta := Metadata{
		GeneratedAt:  fixedTime,
		Scope:        "test",
		ScopeValue:   "value",
		Since:        "-7d",
		TotalPRs:     5,
		Repositories: []string{"repo1", "repo2"},
		LLMProvider:  "openai",
		LLMModel:     "gpt-4",
		Summary:      "Test summary",
	}

	result := Render(meta, []*model.PR{})

	// Check that all metadata appears in the output
	expectedStrings := []string{
		"2024-01-15 10:30:00 UTC",
		"test (value)",
		"-7d",
		"5",
		"repo1, repo2",
		"openai (gpt-4)",
		"Test summary",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain %q, but it didn't", expected)
		}
	}
}

// Test helper functions

func shouldUpdateGolden() bool {
	return os.Getenv("UPDATE_GOLDEN") == "1"
}

func updateGoldenFile(t *testing.T, filename, content string) {
	goldenDir := "testdata"
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("Failed to create golden dir: %v", err)
	}

	goldenPath := filepath.Join(goldenDir, filename)
	if err := os.WriteFile(goldenPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write golden file %s: %v", goldenPath, err)
	}

	t.Logf("Updated golden file: %s", goldenPath)
}

func compareWithGolden(t *testing.T, filename, actual string) {
	goldenPath := filepath.Join("testdata", filename)

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file %s does not exist. Run with UPDATE_GOLDEN=1 to create it.", goldenPath)
		}
		t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
	}

	if actual != string(expected) {
		t.Errorf("Output doesn't match golden file %s\nExpected:\n%s\nActual:\n%s",
			goldenPath, string(expected), actual)
	}
}

package render

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/prtool/internal/model"
)

func TestRender(t *testing.T) {
	// Fixed time for consistent output
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		meta       Metadata
		prs        []model.PR
		goldenFile string
	}{
		{
			name: "basic report",
			meta: Metadata{
				GeneratedAt:  fixedTime,
				TimeRange:    "-7d",
				TotalPRs:     2,
				Repositories: []string{"owner/repo1", "owner/repo2"},
				LLMProvider:  "openai",
				LLMModel:     "gpt-4",
			},
			prs: []model.PR{
				{
					Repository:  "owner/repo1",
					Number:      123,
					Title:       "Add new feature",
					Description: "This PR adds a new feature to the system.",
					Author:      "alice",
					URL:         "https://github.com/owner/repo1/pull/123",
					MergedAt:    fixedTime.AddDate(0, 0, -1),
					Labels:      []string{"enhancement", "feature"},
				},
				{
					Repository:  "owner/repo2",
					Number:      456,
					Title:       "Fix critical bug",
					Description: "Fixes a critical bug in the authentication system.",
					Author:      "bob",
					URL:         "https://github.com/owner/repo2/pull/456",
					MergedAt:    fixedTime.AddDate(0, 0, -2),
					Labels:      []string{"bug", "critical"},
				},
			},
			goldenFile: "basic_report.golden",
		},
		{
			name: "report with summaries",
			meta: Metadata{
				GeneratedAt:  fixedTime,
				TimeRange:    "-1m",
				TotalPRs:     1,
				Repositories: []string{"org/project"},
			},
			prs: []model.PR{
				{
					Repository: "org/project",
					Number:     789,
					Title:      "Refactor core module",
					Author:     "charlie",
					URL:        "https://github.com/org/project/pull/789",
					MergedAt:   fixedTime.AddDate(0, 0, -5),
					Summary:    "This PR refactors the core module to improve performance and maintainability. Key changes include:\n- Simplified data flow\n- Reduced coupling between components\n- Added comprehensive tests",
				},
			},
			goldenFile: "report_with_summary.golden",
		},
		{
			name: "empty report",
			meta: Metadata{
				GeneratedAt:  fixedTime,
				TimeRange:    "-7d",
				TotalPRs:     0,
				Repositories: []string{"owner/repo"},
			},
			prs:        []model.PR{},
			goldenFile: "empty_report.golden",
		},
		{
			name: "many repos",
			meta: Metadata{
				GeneratedAt: fixedTime,
				TimeRange:   "-7d",
				TotalPRs:    1,
				Repositories: []string{
					"org/repo1", "org/repo2", "org/repo3", "org/repo4", "org/repo5",
					"org/repo6", "org/repo7", "org/repo8", "org/repo9", "org/repo10",
					"org/repo11", // More than 10, should not list individually
				},
			},
			prs: []model.PR{
				{
					Repository: "org/repo1",
					Number:     100,
					Title:      "Update dependencies",
					Author:     "dave",
					URL:        "https://github.com/org/repo1/pull/100",
					MergedAt:   fixedTime.AddDate(0, 0, -3),
				},
			},
			goldenFile: "many_repos.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.meta, tt.prs)

			// Compare with golden file
			goldenPath := filepath.Join("testdata", tt.goldenFile)
			if err := os.MkdirAll("testdata", 0755); err != nil {
				t.Fatal(err)
			}

			if *updateGolden {
				err := os.WriteFile(goldenPath, []byte(got), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}

			if got != string(want) {
				t.Errorf("Render() output mismatch for %s\nGot:\n%s\nWant:\n%s", tt.goldenFile, got, want)
			}
		})
	}
}

func TestRenderDryRun(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		prs  []model.PR
		want string
	}{
		{
			name: "basic table",
			prs: []model.PR{
				{
					Repository: "owner/repo1",
					Number:     123,
					Title:      "Add new feature",
					Author:     "alice",
					MergedAt:   fixedTime,
				},
				{
					Repository: "owner/repo2",
					Number:     456,
					Title:      "Fix critical bug that was causing issues",
					Author:     "bob-the-developer",
					MergedAt:   fixedTime.AddDate(0, 0, -1),
				},
			},
			want: `Found 2 pull requests:

Repository                    | PR # | Title                                           | Author      | Merged
------------------------------|------|-------------------------------------------------|-------------|------------
owner/repo1                   | 123  | Add new feature                                 | alice       | 2024-01-15
owner/repo2                   | 456  | Fix critical bug that was causing issues        | bob-the-... | 2024-01-14
`,
		},
		{
			name: "empty list",
			prs:  []model.PR{},
			want: "No pull requests found.\n",
		},
		{
			name: "long names truncated",
			prs: []model.PR{
				{
					Repository: "very-long-organization-name/very-long-repository-name",
					Number:     999,
					Title:      "This is an extremely long pull request title that should be truncated in the output",
					Author:     "user-with-very-long-username",
					MergedAt:   fixedTime,
				},
			},
			want: `Found 1 pull requests:

Repository                    | PR # | Title                                           | Author      | Merged
------------------------------|------|-------------------------------------------------|-------------|------------
very-long-organization-na...  | 999  | This is an extremely long pull request title... | user-wit... | 2024-01-15
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderDryRun(tt.prs)
			if got != tt.want {
				t.Errorf("RenderDryRun() mismatch\nGot:\n%s\nWant:\n%s", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"test", 3, "tes"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

// Test flag for updating golden files
var updateGolden = flag.Bool("update", false, "update golden files")

var flag = struct {
	Bool func(name string, value bool, usage string) *bool
}{
	Bool: func(name string, value bool, usage string) *bool {
		v := value
		return &v
	},
}
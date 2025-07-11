package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
)

func TestFetch(t *testing.T) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	tenDaysAgo := now.AddDate(0, 0, -10)
	fiveDaysAgo := now.AddDate(0, 0, -5)

	tests := []struct {
		name          string
		cfg           *config.Config
		mockRepos     []*github.Repository
		mockPRs       map[string][]gh.PR // repo -> PRs
		wantPRCount   int
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "fetch merged PRs from single repo",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Repositories: []string{"owner/repo1"},
				},
			},
			mockPRs: map[string][]gh.PR{
				"owner/repo1": {
					{
						Number:      1,
						Title:       "PR 1",
						Description: "Description 1",
						Author:      "user1",
						MergedAt:    fiveDaysAgo,
						URL:         "https://github.com/owner/repo1/pull/1",
						Labels:      []string{"bug"},
					},
					{
						Number:      2,
						Title:       "PR 2",
						Description: "Description 2",
						Author:      "user2",
						MergedAt:    tenDaysAgo, // Should be filtered out
						URL:         "https://github.com/owner/repo1/pull/2",
					},
				},
			},
			wantPRCount: 1, // Only PR 1 should be included
		},
		{
			name: "fetch from multiple repos",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Repositories: []string{"owner/repo1", "owner/repo2"},
				},
			},
			mockPRs: map[string][]gh.PR{
				"owner/repo1": {
					{
						Number:   3,
						Title:    "PR 3",
						Author:   "user1",
						MergedAt: fiveDaysAgo,
						URL:      "https://github.com/owner/repo1/pull/3",
					},
				},
				"owner/repo2": {
					{
						Number:   4,
						Title:    "PR 4",
						Author:   "user2",
						MergedAt: fiveDaysAgo,
						URL:      "https://github.com/owner/repo2/pull/4",
					},
					{
						Number:   5,
						Title:    "PR 5",
						Author:   "user3",
						MergedAt: now.AddDate(0, 0, -3),
						URL:      "https://github.com/owner/repo2/pull/5",
					},
				},
			},
			wantPRCount: 3,
		},
		{
			name: "filter out non-merged PRs",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Repositories: []string{"owner/repo1"},
				},
			},
			mockPRs: map[string][]gh.PR{
				"owner/repo1": {
					{
						Number:   6,
						Title:    "Merged PR",
						MergedAt: fiveDaysAgo,
						URL:      "https://github.com/owner/repo1/pull/6",
					},
					{
						Number:   7,
						Title:    "Open PR",
						MergedAt: time.Time{}, // Zero time = not merged
						URL:      "https://github.com/owner/repo1/pull/7",
					},
				},
			},
			wantPRCount: 1, // Only merged PR
		},
		{
			name: "organization scope",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Organization: "test-org",
				},
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
				{FullName: github.String("test-org/repo2")},
			},
			mockPRs: map[string][]gh.PR{
				"test-org/repo1": {
					{
						Number:   8,
						Title:    "Org PR 1",
						MergedAt: fiveDaysAgo,
						URL:      "https://github.com/test-org/repo1/pull/8",
					},
				},
				"test-org/repo2": {
					{
						Number:   9,
						Title:    "Org PR 2",
						MergedAt: fiveDaysAgo,
						URL:      "https://github.com/test-org/repo2/pull/9",
					},
				},
			},
			wantPRCount: 2,
		},
		{
			name: "no merged PRs in time range",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Repositories: []string{"owner/repo1"},
				},
			},
			mockPRs: map[string][]gh.PR{
				"owner/repo1": {
					{
						Number:   10,
						Title:    "Old PR",
						MergedAt: now.AddDate(0, 0, -30), // Too old
						URL:      "https://github.com/owner/repo1/pull/10",
					},
				},
			},
			wantPRCount: 0,
		},
		{
			name: "API error",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Repositories: []string{"owner/repo1"},
				},
			},
			mockPRs:    map[string][]gh.PR{},
			wantErr:    true,
			wantErrMsg: "failed to fetch PRs",
		},
		{
			name: "no scope specified",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{},
			},
			wantErr:    true,
			wantErrMsg: "no scope specified",
		},
		{
			name: "verbose mode",
			cfg: &config.Config{
				GitHub: config.GitHubConfig{
					Repositories: []string{"owner/repo1"},
				},
				Verbose: true,
			},
			mockPRs: map[string][]gh.PR{
				"owner/repo1": {
					{
						Number:   11,
						Title:    "Verbose test PR",
						MergedAt: fiveDaysAgo,
						URL:      "https://github.com/owner/repo1/pull/11",
					},
				},
			},
			wantPRCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := gh.NewMockClient()
			
			// Configure ListRepos behavior
			mockClient.ListReposFunc = func(scope config.GitHubConfig) ([]*github.Repository, error) {
				return tt.mockRepos, nil
			}
			
			// Configure ListPRs behavior
			mockClient.ListPRsFunc = func(repo string, since time.Time) ([]gh.PR, error) {
				if tt.wantErr && tt.wantErrMsg == "failed to fetch PRs" {
					return nil, fmt.Errorf("API error")
				}
				
				prs, ok := tt.mockPRs[repo]
				if !ok {
					return []gh.PR{}, nil
				}
				return prs, nil
			}

			// Execute fetch
			prs, err := Fetch(tt.cfg, mockClient)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != "" {
				if !contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Fetch() error = %v, want error containing %q", err, tt.wantErrMsg)
				}
				return
			}

			// Check PR count
			if len(prs) != tt.wantPRCount {
				t.Errorf("Fetch() returned %d PRs, want %d", len(prs), tt.wantPRCount)
			}

			// Verify PRs are sorted by merge date (newest first)
			for i := 1; i < len(prs); i++ {
				if prs[i-1].MergedAt.Before(prs[i].MergedAt) {
					t.Errorf("PRs not sorted correctly: PR[%d] merged at %v is before PR[%d] merged at %v",
						i-1, prs[i-1].MergedAt, i, prs[i].MergedAt)
				}
			}

			// Verify all PRs are within the time range
			for _, pr := range prs {
				if pr.MergedAt.Before(sevenDaysAgo) {
					t.Errorf("PR %d merged at %v is before the 7-day cutoff %v",
						pr.Number, pr.MergedAt, sevenDaysAgo)
				}
			}
		})
	}
}

func TestFetch_PRConversion(t *testing.T) {
	// Test that gh.PR is correctly converted to model.PR
	mockClient := gh.NewMockClient()
	
	ghPR := gh.PR{
		Number:      123,
		Title:       "Test PR",
		Description: "Test Description",
		Author:      "testuser",
		MergedAt:    time.Now().AddDate(0, 0, -1),
		URL:         "https://github.com/owner/repo/pull/123",
		Labels:      []string{"enhancement", "documentation"},
	}
	
	mockClient.ListPRsFunc = func(repo string, since time.Time) ([]gh.PR, error) {
		return []gh.PR{ghPR}, nil
	}
	
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Repositories: []string{"owner/repo"},
		},
	}
	
	prs, err := Fetch(cfg, mockClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if len(prs) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(prs))
	}
	
	modelPR := prs[0]
	if modelPR.Repository != "owner/repo" {
		t.Errorf("Repository = %q, want %q", modelPR.Repository, "owner/repo")
	}
	if modelPR.Number != ghPR.Number {
		t.Errorf("Number = %d, want %d", modelPR.Number, ghPR.Number)
	}
	if modelPR.Title != ghPR.Title {
		t.Errorf("Title = %q, want %q", modelPR.Title, ghPR.Title)
	}
	if modelPR.Description != ghPR.Description {
		t.Errorf("Description = %q, want %q", modelPR.Description, ghPR.Description)
	}
	if modelPR.Author != ghPR.Author {
		t.Errorf("Author = %q, want %q", modelPR.Author, ghPR.Author)
	}
	if !modelPR.MergedAt.Equal(ghPR.MergedAt) {
		t.Errorf("MergedAt = %v, want %v", modelPR.MergedAt, ghPR.MergedAt)
	}
	if modelPR.URL != ghPR.URL {
		t.Errorf("URL = %q, want %q", modelPR.URL, ghPR.URL)
	}
	if len(modelPR.Labels) != len(ghPR.Labels) {
		t.Errorf("Labels = %v, want %v", modelPR.Labels, ghPR.Labels)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) >= len(substr) && contains(s[1:], substr)
}
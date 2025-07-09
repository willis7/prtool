package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/model"
)

func TestNewFetcher(t *testing.T) {
	mockClient := gh.NewMockClient()
	fetcher := NewFetcher(mockClient)

	if fetcher == nil {
		t.Error("Expected non-nil fetcher")
	}

	if fetcher.ghClient != mockClient {
		t.Error("Expected fetcher to use provided GitHub client")
	}
}

func TestFetcher_Fetch(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)
	twoWeeksAgo := now.AddDate(0, 0, -13) // 13 days ago to ensure it's within 14 days

	tests := []struct {
		name          string
		cfg           *config.Config
		mockRepos     []*github.Repository
		mockPRs       []*model.PR
		repoError     error
		prError       error
		expectedPRs   int
		expectedRepos []string
		expectError   bool
		errorMsg      string
	}{
		{
			name: "successful fetch with org scope and default since",
			cfg: &config.Config{
				Org: "test-org",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
				{FullName: github.String("test-org/repo2")},
			},
			mockPRs: []*model.PR{
				{
					Title:      "Feature A",
					Author:     "user1",
					MergedAt:   &yesterday,
					State:      "closed",
					Repository: "test-org/repo1",
				},
				{
					Title:      "Feature B",
					Author:     "user2",
					MergedAt:   &yesterday,
					State:      "closed",
					Repository: "test-org/repo2",
				},
				{
					Title:      "Old PR",
					Author:     "user3",
					MergedAt:   &twoWeeksAgo,
					State:      "closed",
					Repository: "test-org/repo1",
				},
			},
			expectedPRs:   2, // Only recent merged PRs
			expectedRepos: []string{"test-org/repo1", "test-org/repo2"},
			expectError:   false,
		},
		{
			name: "successful fetch with custom since filter",
			cfg: &config.Config{
				Org:   "test-org",
				Since: "-14d",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
			},
			mockPRs: []*model.PR{
				{
					Title:      "Recent PR",
					Author:     "user1",
					MergedAt:   &yesterday,
					State:      "closed",
					Repository: "test-org/repo1",
				},
				{
					Title:      "Old PR",
					Author:     "user2",
					MergedAt:   &twoWeeksAgo,
					State:      "closed",
					Repository: "test-org/repo1",
				},
			},
			expectedPRs:   2, // Both PRs within 14 days
			expectedRepos: []string{"test-org/repo1"},
			expectError:   false,
		},
		{
			name: "filter out non-merged PRs",
			cfg: &config.Config{
				User: "test-user",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-user/repo1")},
			},
			mockPRs: []*model.PR{
				{
					Title:      "Merged PR",
					Author:     "user1",
					MergedAt:   &yesterday,
					State:      "closed",
					Repository: "test-user/repo1",
				},
				{
					Title:      "Open PR",
					Author:     "user2",
					MergedAt:   nil,
					State:      "open",
					Repository: "test-user/repo1",
				},
				{
					Title:      "Closed but not merged",
					Author:     "user3",
					MergedAt:   nil,
					State:      "closed",
					Repository: "test-user/repo1",
				},
			},
			expectedPRs:   1, // Only the merged PR
			expectedRepos: []string{"test-user/repo1"},
			expectError:   false,
		},
		{
			name: "filter PRs by since date",
			cfg: &config.Config{
				Repo:  "owner/repo",
				Since: "-3d",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("owner/repo")},
			},
			mockPRs: []*model.PR{
				{
					Title:      "Recent PR",
					Author:     "user1",
					MergedAt:   &yesterday,
					State:      "closed",
					Repository: "owner/repo",
				},
				{
					Title:      "Old PR",
					Author:     "user2",
					MergedAt:   &lastWeek,
					State:      "closed",
					Repository: "owner/repo",
				},
			},
			expectedPRs:   1, // Only recent PR within 3 days
			expectedRepos: []string{"owner/repo"},
			expectError:   false,
		},
		{
			name:        "nil config should return error",
			cfg:         nil,
			expectError: true,
			errorMsg:    "configuration is required",
		},
		{
			name: "invalid since filter should return error",
			cfg: &config.Config{
				Org:   "test-org",
				Since: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid since filter",
		},
		{
			name: "scope resolution error should be propagated",
			cfg: &config.Config{
				Org: "test-org",
			},
			repoError:   fmt.Errorf("API rate limit exceeded"),
			expectError: true,
			errorMsg:    "failed to resolve repositories",
		},
		{
			name: "PR fetching error should be propagated",
			cfg: &config.Config{
				Org: "test-org",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
			},
			prError:     fmt.Errorf("repository not found"),
			expectError: true,
			errorMsg:    "failed to fetch PRs from repository",
		},
		{
			name: "empty results should return empty slice",
			cfg: &config.Config{
				Org: "empty-org",
			},
			mockRepos:     []*github.Repository{},
			expectedPRs:   0,
			expectedRepos: []string{},
			expectError:   true, // Empty repos should cause scope resolution error
			errorMsg:      "failed to resolve repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := gh.NewMockClient()
			mockClient.SetMockRepos(tt.mockRepos)
			mockClient.SetMockPRs(tt.mockPRs)

			if tt.repoError != nil {
				mockClient.SetRepoError(tt.repoError)
			}
			if tt.prError != nil {
				mockClient.SetPRError(tt.prError)
			}

			fetcher := NewFetcher(mockClient)
			prs, err := fetcher.Fetch(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !containsSubstring(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(prs) != tt.expectedPRs {
				t.Errorf("Expected %d PRs, got %d", tt.expectedPRs, len(prs))
			}

			// Verify all returned PRs are merged
			for i, pr := range prs {
				if pr.MergedAt == nil {
					t.Errorf("PR %d should have MergedAt set", i)
				}
				if pr.State != "closed" {
					t.Errorf("PR %d should have state 'closed', got %q", i, pr.State)
				}
			}
		})
	}
}

func TestFetcher_Fetch_NilGitHubClient(t *testing.T) {
	fetcher := &Fetcher{ghClient: nil}
	cfg := &config.Config{Org: "test-org"}

	prs, err := fetcher.Fetch(cfg)

	if err == nil {
		t.Error("Expected error for nil GitHub client")
	}

	if prs != nil {
		t.Error("Expected nil PRs for nil GitHub client")
	}

	expectedMsg := "GitHub client is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestFetch_ConvenienceFunction(t *testing.T) {
	mockClient := gh.NewMockClient()
	mockClient.SetMockRepos([]*github.Repository{
		{FullName: github.String("test-org/repo1")},
	})

	yesterday := time.Now().AddDate(0, 0, -1)
	mockClient.SetMockPRs([]*model.PR{
		{
			Title:      "Test PR",
			Author:     "user1",
			MergedAt:   &yesterday,
			State:      "closed",
			Repository: "test-org/repo1",
		},
	})

	cfg := &config.Config{Org: "test-org"}
	prs, err := Fetch(cfg, mockClient)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prs))
	}

	if prs[0].Title != "Test PR" {
		t.Errorf("Expected PR title 'Test PR', got %q", prs[0].Title)
	}
}

func TestFetcher_Integration(t *testing.T) {
	// Integration test that verifies the full flow with MockClient
	mockClient := gh.NewMockClient()

	// Setup mock repositories
	mockRepos := []*github.Repository{
		{FullName: github.String("org/repo1")},
		{FullName: github.String("org/repo2")},
	}
	mockClient.SetMockRepos(mockRepos)

	// Setup mock PRs with different scenarios
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	twoWeeksAgo := now.AddDate(0, 0, -13) // 13 days ago to ensure it's within filters

	mockPRs := []*model.PR{
		{
			Title:      "Recent merged PR",
			Author:     "user1",
			MergedAt:   &yesterday,
			State:      "closed",
			Repository: "org/repo1",
			Number:     100,
		},
		{
			Title:      "Recent merged PR in repo2",
			Author:     "user2",
			MergedAt:   &yesterday,
			State:      "closed",
			Repository: "org/repo2",
			Number:     200,
		},
		{
			Title:      "Old merged PR",
			Author:     "user3",
			MergedAt:   &twoWeeksAgo,
			State:      "closed",
			Repository: "org/repo1",
			Number:     50,
		},
		{
			Title:      "Open PR",
			Author:     "user4",
			MergedAt:   nil,
			State:      "open",
			Repository: "org/repo1",
			Number:     300,
		},
		{
			Title:      "Closed but not merged",
			Author:     "user5",
			MergedAt:   nil,
			State:      "closed",
			Repository: "org/repo2",
			Number:     400,
		},
	}
	mockClient.SetMockPRs(mockPRs)

	cfg := &config.Config{
		Org:   "org",
		Since: "-10d",
	}

	fetcher := NewFetcher(mockClient)
	prs, err := fetcher.Fetch(cfg)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return only recent merged PRs (within 10 days and merged)
	expectedCount := 2 // yesterday PRs from both repos
	if len(prs) != expectedCount {
		t.Errorf("Expected %d PRs, got %d", expectedCount, len(prs))
	}

	// Verify call log shows scope resolution and PR fetching
	callLog := mockClient.GetCallLog()
	expectedCalls := 3 // 1 ListRepos call + 2 ListPRs calls (one per repo)
	if len(callLog) != expectedCalls {
		t.Errorf("Expected %d calls, got %d: %v", expectedCalls, len(callLog), callLog)
	}

	// Verify the calls were made correctly
	if !containsSubstring(callLog[0], "ListRepos") {
		t.Errorf("Expected first call to be ListRepos, got %q", callLog[0])
	}

	if !containsSubstring(callLog[1], "ListPRs") {
		t.Errorf("Expected second call to be ListPRs, got %q", callLog[1])
	}

	if !containsSubstring(callLog[2], "ListPRs") {
		t.Errorf("Expected third call to be ListPRs, got %q", callLog[2])
	}

	// Verify all returned PRs are properly filtered
	for i, pr := range prs {
		if pr.MergedAt == nil {
			t.Errorf("PR %d (%s) should be merged", i, pr.Title)
		}
		if pr.State != "closed" {
			t.Errorf("PR %d (%s) should be closed, got %q", i, pr.Title, pr.State)
		}
		if pr.MergedAt.Before(now.AddDate(0, 0, -10)) {
			t.Errorf("PR %d (%s) should be within the since filter", i, pr.Title)
		}
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring checks if substr exists in s
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

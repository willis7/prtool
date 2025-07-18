package gh

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/model"
)

func TestNewRestClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "empty token should return error",
			token:       "",
			expectError: true,
		},
		// Note: We skip the invalid token test to avoid making real API calls
		// This would be tested in integration tests with actual tokens
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewRestClient(tt.token)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if client != nil {
					t.Error("Expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected non-nil client")
				}
			}
		})
	}
}

func TestMockClient_ListRepos(t *testing.T) {
	tests := []struct {
		name        string
		scope       *config.Config
		mockRepos   []*github.Repository
		authError   error
		repoError   error
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid org scope",
			scope: &config.Config{
				Org: "test-org",
			},
			mockRepos: []*github.Repository{
				{Name: github.String("repo1")},
				{Name: github.String("repo2")},
			},
			expectError: false,
		},
		{
			name: "valid user scope",
			scope: &config.Config{
				User: "test-user",
			},
			mockRepos: []*github.Repository{
				{Name: github.String("user-repo")},
			},
			expectError: false,
		},
		{
			name: "valid repo scope",
			scope: &config.Config{
				Repo: "owner/repo",
			},
			mockRepos: []*github.Repository{
				{Name: github.String("repo")},
			},
			expectError: false,
		},
		{
			name: "valid team scope",
			scope: &config.Config{
				Team: "org/team",
			},
			mockRepos: []*github.Repository{
				{Name: github.String("team-repo")},
			},
			expectError: false,
		},
		{
			name:        "nil scope should return error",
			scope:       nil,
			expectError: true,
			errorMsg:    "scope configuration is required",
		},
		{
			name:        "empty scope should return error",
			scope:       &config.Config{},
			expectError: true,
			errorMsg:    "no valid scope specified (org, user, repo, or team required)",
		},
		{
			name: "multiple scopes should return error",
			scope: &config.Config{
				Org:  "test-org",
				User: "test-user",
			},
			expectError: true,
			errorMsg:    "multiple scopes specified, only one allowed",
		},
		{
			name: "auth error should be returned",
			scope: &config.Config{
				Org: "test-org",
			},
			authError:   fmt.Errorf("authentication failed"),
			expectError: true,
			errorMsg:    "authentication failed",
		},
		{
			name: "repo error should be returned",
			scope: &config.Config{
				Org: "test-org",
			},
			repoError:   fmt.Errorf("repository access denied"),
			expectError: true,
			errorMsg:    "repository access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.SetMockRepos(tt.mockRepos)

			if tt.authError != nil {
				mock.SetAuthError(tt.authError)
			}
			if tt.repoError != nil {
				mock.SetRepoError(tt.repoError)
			}

			repos, err := mock.ListRepos(tt.scope)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(repos) != len(tt.mockRepos) {
					t.Errorf("Expected %d repos, got %d", len(tt.mockRepos), len(repos))
				}
			}
		})
	}
}

func TestMockClient_ListPRs(t *testing.T) {
	// Test dates
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)

	tests := []struct {
		name        string
		repo        string
		since       time.Time
		mockPRs     []*model.PR
		authError   error
		prError     error
		expectError bool
		expectedPRs int
		errorMsg    string
	}{
		{
			name:  "valid repo with recent PRs",
			repo:  "owner/repo",
			since: lastWeek,
			mockPRs: []*model.PR{
				{
					Title:    "Recent PR",
					Author:   "user1",
					MergedAt: &yesterday,
				},
				{
					Title:    "Old PR",
					Author:   "user2",
					MergedAt: &lastWeek,
				},
			},
			expectError: false,
			expectedPRs: 1, // Only the recent PR should be returned
		},
		{
			name:        "empty repo should return error",
			repo:        "",
			since:       lastWeek,
			expectError: true,
			errorMsg:    "repository name is required",
		},
		{
			name:  "auth error should be returned",
			repo:  "owner/repo",
			since: lastWeek,
			mockPRs: []*model.PR{
				{Title: "Test PR", MergedAt: &yesterday},
			},
			authError:   fmt.Errorf("authentication failed"),
			expectError: true,
			errorMsg:    "authentication failed",
		},
		{
			name:  "PR error should be returned",
			repo:  "owner/repo",
			since: lastWeek,
			mockPRs: []*model.PR{
				{Title: "Test PR", MergedAt: &yesterday},
			},
			prError:     fmt.Errorf("PR access denied"),
			expectError: true,
			errorMsg:    "PR access denied",
		},
		{
			name:  "no PRs since date",
			repo:  "owner/repo",
			since: yesterday,
			mockPRs: []*model.PR{
				{
					Title:    "Old PR",
					Author:   "user1",
					MergedAt: &lastWeek,
				},
			},
			expectError: false,
			expectedPRs: 0,
		},
		{
			name:  "PRs without merged date are filtered out",
			repo:  "owner/repo",
			since: lastWeek,
			mockPRs: []*model.PR{
				{
					Title:    "Unmerged PR",
					Author:   "user1",
					MergedAt: nil,
				},
				{
					Title:    "Merged PR",
					Author:   "user2",
					MergedAt: &yesterday,
				},
			},
			expectError: false,
			expectedPRs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.SetMockPRs(tt.mockPRs)

			if tt.authError != nil {
				mock.SetAuthError(tt.authError)
			}
			if tt.prError != nil {
				mock.SetPRError(tt.prError)
			}

			prs, err := mock.ListPRs(tt.repo, tt.since)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(prs) != tt.expectedPRs {
					t.Errorf("Expected %d PRs, got %d", tt.expectedPRs, len(prs))
				}
			}
		})
	}
}

func TestMockClient_CallLog(t *testing.T) {
	mock := NewMockClient()

	// Setup mock data
	mockRepos := []*github.Repository{
		{Name: github.String("test-repo")},
	}
	mockPRs := []*model.PR{
		{Title: "Test PR", MergedAt: &time.Time{}},
	}

	mock.SetMockRepos(mockRepos)
	mock.SetMockPRs(mockPRs)

	// Make some calls
	scope := &config.Config{Org: "test-org"}
	since := time.Now().AddDate(0, 0, -7)

	_, _ = mock.ListRepos(scope)
	_, _ = mock.ListPRs("owner/repo", since)

	// Check call log
	callLog := mock.GetCallLog()
	if len(callLog) != 2 {
		t.Errorf("Expected 2 calls in log, got %d", len(callLog))
	}

	expectedCalls := []string{
		"ListRepos",
		"ListPRs",
	}

	for i, expectedCall := range expectedCalls {
		if i < len(callLog) {
			if !containsString(callLog[i], expectedCall) {
				t.Errorf("Expected call %d to contain %q, got %q", i, expectedCall, callLog[i])
			}
		}
	}

	// Test clear log
	mock.ClearCallLog()
	if len(mock.GetCallLog()) != 0 {
		t.Error("Expected empty call log after clear")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

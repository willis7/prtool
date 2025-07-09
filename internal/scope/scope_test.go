package scope

import (
	"fmt"
	"testing"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
)

func TestResolveRepos(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		mockRepos    []*github.Repository
		mockError    error
		expectedLen  int
		expectedRepo string
		expectError  bool
		errorMsg     string
	}{
		{
			name: "valid org scope",
			cfg: &config.Config{
				Org: "test-org",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
				{FullName: github.String("test-org/repo2")},
			},
			expectedLen:  2,
			expectedRepo: "test-org/repo1",
			expectError:  false,
		},
		{
			name: "valid user scope",
			cfg: &config.Config{
				User: "test-user",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-user/my-repo")},
			},
			expectedLen:  1,
			expectedRepo: "test-user/my-repo",
			expectError:  false,
		},
		{
			name: "valid repo scope",
			cfg: &config.Config{
				Repo: "owner/specific-repo",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("owner/specific-repo")},
			},
			expectedLen:  1,
			expectedRepo: "owner/specific-repo",
			expectError:  false,
		},
		{
			name: "valid team scope",
			cfg: &config.Config{
				Team: "org/team",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("org/team-repo1")},
				{FullName: github.String("org/team-repo2")},
			},
			expectedLen:  2,
			expectedRepo: "org/team-repo1",
			expectError:  false,
		},
		{
			name: "repo without FullName falls back to owner/name",
			cfg: &config.Config{
				Org: "test-org",
			},
			mockRepos: []*github.Repository{
				{
					Owner: &github.User{Login: github.String("test-org")},
					Name:  github.String("fallback-repo"),
				},
			},
			expectedLen:  1,
			expectedRepo: "test-org/fallback-repo",
			expectError:  false,
		},
		{
			name:        "nil config should return error",
			cfg:         nil,
			expectError: true,
			errorMsg:    "configuration is required",
		},
		{
			name: "empty config should return error",
			cfg:  &config.Config{},
			expectError: true,
			errorMsg:    "no scope specified",
		},
		{
			name: "multiple scopes should return error",
			cfg: &config.Config{
				Org:  "test-org",
				User: "test-user",
			},
			expectError: true,
			errorMsg:    "multiple scopes specified",
		},
		{
			name: "GitHub client error should be propagated",
			cfg: &config.Config{
				Org: "test-org",
			},
			mockError:   fmt.Errorf("API rate limit exceeded"),
			expectError: true,
			errorMsg:    "failed to list repositories",
		},
		{
			name: "no repositories found should return error",
			cfg: &config.Config{
				Org: "empty-org",
			},
			mockRepos:   []*github.Repository{},
			expectError: true,
			errorMsg:    "no repositories found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := gh.NewMockClient()
			mockClient.SetMockRepos(tt.mockRepos)
			
			if tt.mockError != nil {
				mockClient.SetRepoError(tt.mockError)
			}
			
			// Test ResolveRepos
			repos, err := ResolveRepos(tt.cfg, mockClient)
			
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
			
			if len(repos) != tt.expectedLen {
				t.Errorf("Expected %d repositories, got %d", tt.expectedLen, len(repos))
			}
			
			if tt.expectedLen > 0 && repos[0] != tt.expectedRepo {
				t.Errorf("Expected first repo to be %q, got %q", tt.expectedRepo, repos[0])
			}
		})
	}
}

func TestResolveRepos_NilGitHubClient(t *testing.T) {
	cfg := &config.Config{Org: "test-org"}
	
	repos, err := ResolveRepos(cfg, nil)
	
	if err == nil {
		t.Error("Expected error for nil GitHub client")
	}
	
	if repos != nil {
		t.Error("Expected nil repos for nil GitHub client")
	}
	
	expectedMsg := "GitHub client is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestValidateScope(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid org scope",
			cfg: &config.Config{
				Org: "test-org",
			},
			expectError: false,
		},
		{
			name: "valid user scope",
			cfg: &config.Config{
				User: "test-user",
			},
			expectError: false,
		},
		{
			name: "valid repo scope",
			cfg: &config.Config{
				Repo: "owner/repo",
			},
			expectError: false,
		},
		{
			name: "valid team scope",
			cfg: &config.Config{
				Team: "org/team",
			},
			expectError: false,
		},
		{
			name:        "nil config should return error",
			cfg:         nil,
			expectError: true,
			errorMsg:    "configuration is required",
		},
		{
			name:        "empty config should return error",
			cfg:         &config.Config{},
			expectError: true,
			errorMsg:    "no scope specified",
		},
		{
			name: "multiple scopes should return error",
			cfg: &config.Config{
				Org:  "test-org",
				User: "test-user",
			},
			expectError: true,
			errorMsg:    "multiple scopes specified",
		},
		{
			name: "all scopes should return error",
			cfg: &config.Config{
				Org:  "test-org",
				Team: "org/team",
				User: "test-user",
				Repo: "owner/repo",
			},
			expectError: true,
			errorMsg:    "multiple scopes specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScope(tt.cfg)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !containsSubstring(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestResolveRepos_Integration(t *testing.T) {
	// Test integration with MockClient to ensure call logging works
	mockClient := gh.NewMockClient()
	
	// Setup mock repos
	mockRepos := []*github.Repository{
		{FullName: github.String("test-org/repo1")},
		{FullName: github.String("test-org/repo2")},
		{FullName: github.String("test-org/repo3")},
	}
	mockClient.SetMockRepos(mockRepos)
	
	cfg := &config.Config{Org: "test-org"}
	
	repos, err := ResolveRepos(cfg, mockClient)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(repos) != 3 {
		t.Errorf("Expected 3 repositories, got %d", len(repos))
	}
	
	expectedRepos := []string{
		"test-org/repo1",
		"test-org/repo2",
		"test-org/repo3",
	}
	
	for i, expectedRepo := range expectedRepos {
		if i < len(repos) && repos[i] != expectedRepo {
			t.Errorf("Expected repo %d to be %q, got %q", i, expectedRepo, repos[i])
		}
	}
	
	// Verify that ListRepos was called
	callLog := mockClient.GetCallLog()
	if len(callLog) != 1 {
		t.Errorf("Expected 1 call to MockClient, got %d", len(callLog))
	}
	
	if !containsSubstring(callLog[0], "ListRepos") {
		t.Errorf("Expected call to contain 'ListRepos', got %q", callLog[0])
	}
}

func TestResolveRepos_ErrorHandling(t *testing.T) {
	// Test various error scenarios
	testCases := []struct {
		name      string
		cfg       *config.Config
		mockError error
		errorMsg  string
	}{
		{
			name:      "authentication error",
			cfg:       &config.Config{Org: "test-org"},
			mockError: fmt.Errorf("authentication failed"),
			errorMsg:  "failed to list repositories for org",
		},
		{
			name:      "network error",
			cfg:       &config.Config{User: "test-user"},
			mockError: fmt.Errorf("network timeout"),
			errorMsg:  "failed to list repositories for user",
		},
		{
			name:      "permission error",
			cfg:       &config.Config{Repo: "owner/repo"},
			mockError: fmt.Errorf("repository not found"),
			errorMsg:  "failed to list repositories for repo",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := gh.NewMockClient()
			mockClient.SetRepoError(tt.mockError)
			
			repos, err := ResolveRepos(tt.cfg, mockClient)
			
			if err == nil {
				t.Error("Expected error but got none")
			}
			
			if repos != nil {
				t.Error("Expected nil repos on error")
			}
			
			if !containsSubstring(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
			}
		})
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

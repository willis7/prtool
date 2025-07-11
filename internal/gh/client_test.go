package gh

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
)

func TestNewRestClient(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   "ghp_test_token",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errMsg:  "GitHub token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewRestClient(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRestClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("NewRestClient() error = %v, want error containing %q", err, tt.errMsg)
			}
			if !tt.wantErr && client == nil {
				t.Error("NewRestClient() returned nil client without error")
			}
		})
	}
}

func TestMockClient_AuthError(t *testing.T) {
	client := NewMockClient()
	
	// Configure mock to return auth error for ListRepos
	client.ListReposFunc = func(scope config.GitHubConfig) ([]*github.Repository, error) {
		return nil, MockAuthError()
	}

	// Test ListRepos auth error
	scope := config.GitHubConfig{
		Organization: "test-org",
	}
	
	_, err := client.ListRepos(scope)
	if err == nil {
		t.Fatal("expected authentication error, got nil")
	}
	
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected authentication error message, got: %v", err)
	}
	
	if !strings.Contains(err.Error(), "please check your GitHub token") {
		t.Errorf("expected helpful error message about token, got: %v", err)
	}

	// Configure mock to return auth error for ListPRs
	client.ListPRsFunc = func(repo string, since time.Time) ([]PR, error) {
		return nil, MockAuthError()
	}

	// Test ListPRs auth error
	_, err = client.ListPRs("owner/repo", time.Now().AddDate(0, 0, -7))
	if err == nil {
		t.Fatal("expected authentication error, got nil")
	}
	
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected authentication error message, got: %v", err)
	}
}

func TestMockClient_CallTracking(t *testing.T) {
	client := NewMockClient()
	
	// Test ListRepos call tracking
	scope1 := config.GitHubConfig{Organization: "org1"}
	scope2 := config.GitHubConfig{User: "user1"}
	
	client.ListRepos(scope1)
	client.ListRepos(scope2)
	
	if len(client.ListReposCalls) != 2 {
		t.Errorf("expected 2 ListRepos calls, got %d", len(client.ListReposCalls))
	}
	
	if client.ListReposCalls[0].Organization != "org1" {
		t.Errorf("expected first call with org1, got %v", client.ListReposCalls[0])
	}
	
	if client.ListReposCalls[1].User != "user1" {
		t.Errorf("expected second call with user1, got %v", client.ListReposCalls[1])
	}
	
	// Test ListPRs call tracking
	since := time.Now().AddDate(0, 0, -7)
	client.ListPRs("owner/repo1", since)
	client.ListPRs("owner/repo2", since)
	
	if len(client.ListPRsCalls) != 2 {
		t.Errorf("expected 2 ListPRs calls, got %d", len(client.ListPRsCalls))
	}
	
	if client.ListPRsCalls[0].Repo != "owner/repo1" {
		t.Errorf("expected first call with owner/repo1, got %v", client.ListPRsCalls[0].Repo)
	}
}

func TestParseRepoFullName(t *testing.T) {
	tests := []struct {
		name      string
		fullName  string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid repo name",
			fullName:  "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "org with dashes",
			fullName:  "my-org/my-repo",
			wantOwner: "my-org",
			wantRepo:  "my-repo",
			wantErr:   false,
		},
		{
			name:     "missing slash",
			fullName: "invalidrepo",
			wantErr:  true,
		},
		{
			name:     "too many slashes",
			fullName: "owner/repo/extra",
			wantErr:  true,
		},
		{
			name:     "empty string",
			fullName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRepoFullName(tt.fullName)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRepoFullName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("parseRepoFullName() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("parseRepoFullName() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}
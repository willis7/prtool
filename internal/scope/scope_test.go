package scope

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
)

func TestResolveRepos(t *testing.T) {
	tests := []struct {
		name          string
		cfg           config.GitHubConfig
		mockRepos     []*github.Repository
		mockError     error
		want          []string
		wantErr       bool
		wantErrMsg    string
	}{
		{
			name: "direct repositories specified",
			cfg: config.GitHubConfig{
				Repositories: []string{"owner/repo1", "owner/repo2"},
			},
			want:    []string{"owner/repo1", "owner/repo2"},
			wantErr: false,
		},
		{
			name: "organization scope",
			cfg: config.GitHubConfig{
				Organization: "test-org",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
				{FullName: github.String("test-org/repo2")},
			},
			want:    []string{"test-org/repo1", "test-org/repo2"},
			wantErr: false,
		},
		{
			name: "user scope",
			cfg: config.GitHubConfig{
				User: "test-user",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-user/repo1")},
				{FullName: github.String("test-user/repo2")},
			},
			want:    []string{"test-user/repo1", "test-user/repo2"},
			wantErr: false,
		},
		{
			name: "team scope",
			cfg: config.GitHubConfig{
				Team: "test-team",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("org/team-repo1")},
				{FullName: github.String("org/team-repo2")},
			},
			want:    []string{"org/team-repo1", "org/team-repo2"},
			wantErr: false,
		},
		{
			name:       "no scope specified",
			cfg:        config.GitHubConfig{},
			wantErr:    true,
			wantErrMsg: "no scope specified",
		},
		{
			name: "multiple scopes specified - org and user",
			cfg: config.GitHubConfig{
				Organization: "test-org",
				User:         "test-user",
			},
			wantErr:    true,
			wantErrMsg: "multiple scopes specified",
		},
		{
			name: "multiple scopes specified - team and repos",
			cfg: config.GitHubConfig{
				Team:         "test-team",
				Repositories: []string{"owner/repo1"},
			},
			wantErr:    true,
			wantErrMsg: "multiple scopes specified",
		},
		{
			name: "all scopes specified",
			cfg: config.GitHubConfig{
				Organization: "test-org",
				Team:         "test-team",
				User:         "test-user",
				Repositories: []string{"owner/repo1"},
			},
			wantErr:    true,
			wantErrMsg: "multiple scopes specified",
		},
		{
			name: "GitHub API error",
			cfg: config.GitHubConfig{
				Organization: "test-org",
			},
			mockError:  fmt.Errorf("API rate limit exceeded"),
			wantErr:    true,
			wantErrMsg: "failed to list repositories",
		},
		{
			name: "no repositories found",
			cfg: config.GitHubConfig{
				Organization: "empty-org",
			},
			mockRepos:  []*github.Repository{},
			wantErr:    true,
			wantErrMsg: "no repositories found",
		},
		{
			name: "nil repository in response",
			cfg: config.GitHubConfig{
				Organization: "test-org",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
				nil, // nil repository
				{FullName: github.String("test-org/repo2")},
			},
			want:    []string{"test-org/repo1", "test-org/repo2"},
			wantErr: false,
		},
		{
			name: "repository with nil FullName",
			cfg: config.GitHubConfig{
				Organization: "test-org",
			},
			mockRepos: []*github.Repository{
				{FullName: github.String("test-org/repo1")},
				{FullName: nil}, // nil FullName
				{FullName: github.String("test-org/repo2")},
			},
			want:    []string{"test-org/repo1", "test-org/repo2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := gh.NewMockClient()
			mockClient.ListReposFunc = func(scope config.GitHubConfig) ([]*github.Repository, error) {
				if tt.mockError != nil {
					return nil, tt.mockError
				}
				return tt.mockRepos, nil
			}

			// Call ResolveRepos
			got, err := ResolveRepos(tt.cfg, mockClient)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveRepos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("ResolveRepos() error = %v, want error containing %q", err, tt.wantErrMsg)
				return
			}

			// Check result
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ResolveRepos() returned %d repos, want %d", len(got), len(tt.want))
					return
				}
				for i, repo := range got {
					if repo != tt.want[i] {
						t.Errorf("ResolveRepos()[%d] = %q, want %q", i, repo, tt.want[i])
					}
				}
			}

			// Verify mock client calls for non-direct repository cases
			if len(tt.cfg.Repositories) == 0 && !tt.wantErr {
				if len(mockClient.ListReposCalls) != 1 {
					t.Errorf("expected 1 ListRepos call, got %d", len(mockClient.ListReposCalls))
				}
			}
		})
	}
}

func TestResolveRepos_CallsGitHubClient(t *testing.T) {
	tests := []struct {
		name           string
		cfg            config.GitHubConfig
		expectAPICall  bool
	}{
		{
			name: "direct repos - no API call",
			cfg: config.GitHubConfig{
				Repositories: []string{"owner/repo"},
			},
			expectAPICall: false,
		},
		{
			name: "org scope - API call",
			cfg: config.GitHubConfig{
				Organization: "test-org",
			},
			expectAPICall: true,
		},
		{
			name: "user scope - API call",
			cfg: config.GitHubConfig{
				User: "test-user",
			},
			expectAPICall: true,
		},
		{
			name: "team scope - API call",
			cfg: config.GitHubConfig{
				Team: "test-team",
			},
			expectAPICall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := gh.NewMockClient()
			mockClient.ListReposFunc = func(scope config.GitHubConfig) ([]*github.Repository, error) {
				return []*github.Repository{
					{FullName: github.String("test/repo")},
				}, nil
			}

			_, _ = ResolveRepos(tt.cfg, mockClient)

			if tt.expectAPICall {
				if len(mockClient.ListReposCalls) != 1 {
					t.Errorf("expected API call, but got %d calls", len(mockClient.ListReposCalls))
				}
			} else {
				if len(mockClient.ListReposCalls) != 0 {
					t.Errorf("expected no API call, but got %d calls", len(mockClient.ListReposCalls))
				}
			}
		})
	}
}
package gh

import (
	"fmt"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
)

// MockClient implements GitHubClient for testing
type MockClient struct {
	// Configurable behavior
	ListReposFunc func(scope config.GitHubConfig) ([]*github.Repository, error)
	ListPRsFunc   func(repo string, since time.Time) ([]PR, error)

	// Track calls for assertions
	ListReposCalls []config.GitHubConfig
	ListPRsCalls   []struct {
		Repo  string
		Since time.Time
	}
}

// NewMockClient creates a new mock client with default behavior
func NewMockClient() *MockClient {
	return &MockClient{
		ListReposFunc: func(scope config.GitHubConfig) ([]*github.Repository, error) {
			return []*github.Repository{}, nil
		},
		ListPRsFunc: func(repo string, since time.Time) ([]PR, error) {
			return []PR{}, nil
		},
	}
}

// ListRepos implements GitHubClient.ListRepos
func (m *MockClient) ListRepos(scope config.GitHubConfig) ([]*github.Repository, error) {
	m.ListReposCalls = append(m.ListReposCalls, scope)
	if m.ListReposFunc != nil {
		return m.ListReposFunc(scope)
	}
	return nil, fmt.Errorf("ListReposFunc not configured")
}

// ListPRs implements GitHubClient.ListPRs
func (m *MockClient) ListPRs(repo string, since time.Time) ([]PR, error) {
	m.ListPRsCalls = append(m.ListPRsCalls, struct {
		Repo  string
		Since time.Time
	}{Repo: repo, Since: since})
	
	if m.ListPRsFunc != nil {
		return m.ListPRsFunc(repo, since)
	}
	return nil, fmt.Errorf("ListPRsFunc not configured")
}

// MockAuthError returns a mock authentication error
func MockAuthError() error {
	return fmt.Errorf("authentication failed: please check your GitHub token")
}
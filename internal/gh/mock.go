package gh

import (
	"fmt"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/model"
)

// MockClient implements GitHubClient for testing purposes.
type MockClient struct {
	ListReposFunc func(cfg *config.Config) ([]*github.Repository, error)
	ListPRsFunc   func(repo string, since time.Time) ([]model.PR, error)
}

// ListRepos calls the mock function.
func (m *MockClient) ListRepos(cfg *config.Config) ([]*github.Repository, error) {
	if m.ListReposFunc != nil {
		return m.ListReposFunc(cfg)
	}
	return nil, fmt.Errorf("ListRepos not implemented")
}

// ListPRs calls the mock function.
func (m *MockClient) ListPRs(repo string, since time.Time) ([]model.PR, error) {
	if m.ListPRsFunc != nil {
		return m.ListPRsFunc(repo, since)
	}
	return nil, fmt.Errorf("ListPRs not implemented")
}

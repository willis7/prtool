package gh

import (
	"fmt"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/model"
)

// MockClient implements GitHubClient for testing purposes
type MockClient struct {
	// MockRepos can be set to control what ListRepos returns
	MockRepos []*github.Repository

	// MockPRs can be set to control what ListPRs returns
	MockPRs []*model.PR

	// AuthError can be set to simulate authentication failures
	AuthError error

	// RepoError can be set to simulate repository listing failures
	RepoError error

	// PRError can be set to simulate PR listing failures
	PRError error

	// CallLog tracks method calls for verification in tests
	CallLog []string
}

// NewMockClient creates a new mock GitHub client
func NewMockClient() *MockClient {
	return &MockClient{
		CallLog: make([]string, 0),
	}
}

// ListRepos implements GitHubClient.ListRepos for testing
func (m *MockClient) ListRepos(scope *config.Config) ([]*github.Repository, error) {
	m.CallLog = append(m.CallLog, fmt.Sprintf("ListRepos(%+v)", scope))

	if m.AuthError != nil {
		return nil, m.AuthError
	}

	if m.RepoError != nil {
		return nil, m.RepoError
	}

	if scope == nil {
		return nil, fmt.Errorf("scope configuration is required")
	}

	// Validate that exactly one scope is specified
	scopeCount := 0
	if scope.Org != "" {
		scopeCount++
	}
	if scope.User != "" {
		scopeCount++
	}
	if scope.Repo != "" {
		scopeCount++
	}
	if len(scope.Team) > 0 {
		scopeCount++
	}

	if scopeCount == 0 {
		return nil, fmt.Errorf("no valid scope specified (org, user, repo, or team required)")
	}

	if scopeCount > 1 {
		return nil, fmt.Errorf("multiple scopes specified, only one allowed")
	}

	return m.MockRepos, nil
}

// ListPRs implements GitHubClient.ListPRs for testing
func (m *MockClient) ListPRs(repo string, since time.Time) ([]*model.PR, error) {
	m.CallLog = append(m.CallLog, fmt.Sprintf("ListPRs(%s, %s)", repo, since.Format("2006-01-02")))

	if m.AuthError != nil {
		return nil, m.AuthError
	}

	if m.PRError != nil {
		return nil, m.PRError
	}

	if repo == "" {
		return nil, fmt.Errorf("repository name is required")
	}

	// Filter PRs by repository and since date
	var filteredPRs []*model.PR
	for _, pr := range m.MockPRs {
		// Filter by repository name (skip if Repository field is empty - for backward compatibility)
		if pr.Repository != "" && pr.Repository != repo {
			continue
		}
		// Filter by since date (only return PRs merged after since)
		if pr.MergedAt != nil && pr.MergedAt.After(since) {
			filteredPRs = append(filteredPRs, pr)
		}
	}

	return filteredPRs, nil
}

// SetMockRepos sets the mock repositories for testing
func (m *MockClient) SetMockRepos(repos []*github.Repository) {
	m.MockRepos = repos
}

// SetMockPRs sets the mock PRs for testing
func (m *MockClient) SetMockPRs(prs []*model.PR) {
	m.MockPRs = prs
}

// SetAuthError sets an authentication error for testing
func (m *MockClient) SetAuthError(err error) {
	m.AuthError = err
}

// SetRepoError sets a repository listing error for testing
func (m *MockClient) SetRepoError(err error) {
	m.RepoError = err
}

// SetPRError sets a PR listing error for testing
func (m *MockClient) SetPRError(err error) {
	m.PRError = err
}

// GetCallLog returns the log of method calls for verification
func (m *MockClient) GetCallLog() []string {
	return m.CallLog
}

// ClearCallLog clears the call log
func (m *MockClient) ClearCallLog() {
	m.CallLog = make([]string, 0)
}

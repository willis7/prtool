package gh

import (
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/model"
)



func TestMockClient_ListRepos(t *testing.T) {
	expectedRepos := []*github.Repository{{Name: github.String("test-repo")}}
	mockClient := &MockClient{
		ListReposFunc: func(cfg *config.Config) ([]*github.Repository, error) {
			return expectedRepos, nil
		},
	}

	cfg := &config.Config{Org: "test-org"}
	repos, err := mockClient.ListRepos(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 1 || *repos[0].Name != "test-repo" {
		t.Errorf("expected 1 repo named 'test-repo', got %v", repos)
	}
}

func TestMockClient_ListPRs(t *testing.T) {
	expectedPRs := []model.PR{{Title: "Test PR"}}
	mockClient := &MockClient{
		ListPRsFunc: func(repo string, since time.Time) ([]model.PR, error) {
			return expectedPRs, nil
		},
	}

	prs, err := mockClient.ListPRs("owner/repo", time.Now())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prs) != 1 || prs[0].Title != "Test PR" {
		t.Errorf("expected 1 PR named 'Test PR', got %v", prs)
	}
}
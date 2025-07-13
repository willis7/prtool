package service

import (
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
)

type fakeClient struct{}

func (f *fakeClient) ListRepos(cfg config.Config) ([]*github.Repository, error) {
	return nil, nil
}

func (f *fakeClient) ListPRs(repo string, since time.Time) ([]gh.PR, error) {
	return []gh.PR{
		{Title: "Merged PR", Number: 1, MergedAt: time.Now().Add(-1 * time.Hour)},
		{Title: "Old PR", Number: 2, MergedAt: time.Now().Add(-48 * time.Hour)},
	}, nil
}

func TestFetch_FilterMergedPRs(t *testing.T) {
	client := &fakeClient{}
	since := time.Now().Add(-2 * time.Hour)
	prs, err := Fetch(config.Config{}, client, "repo", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 {
		t.Errorf("expected 1 PR, got %d", len(prs))
	}
	if prs[0].Title != "Merged PR" {
		t.Errorf("unexpected PR: %+v", prs[0])
	}
}

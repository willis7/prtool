package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/model"
)

// Mock time.Now() for consistent testing
var fixedTime = time.Date(2024, time.July, 8, 12, 0, 0, 0, time.UTC)

func init() {
	TimeNow = func() time.Time { return fixedTime }
}

func TestFetcher_Fetch_Success(t *testing.T) {
	cfg := &config.Config{Org: "test-org", Since: "-7d"}

	mockClient := &gh.MockClient{
		ListReposFunc: func(cfg *config.Config) ([]*github.Repository, error) {
			return []*github.Repository{{FullName: github.String("test-org/repo1")}}, nil
		},
		ListPRsFunc: func(repo string, since time.Time) ([]model.PR, error) {
			return []model.PR{
				{Title: "PR1", MergedAt: fixedTime.AddDate(0, 0, -1)},
				{Title: "PR2", MergedAt: fixedTime.AddDate(0, 0, -5)},
			},
			nil
		},
	}

	fetcher := NewFetcher(mockClient)

	gotPRs, err := fetcher.Fetch(cfg)

	if err != nil {
		t.Fatalf("Fetch() unexpected error: %v", err)
	}

	expectedPRs := []model.PR{
		{Title: "PR1", MergedAt: fixedTime.AddDate(0, 0, -1)},
		{Title: "PR2", MergedAt: fixedTime.AddDate(0, 0, -5)},
	}

	if len(gotPRs) != len(expectedPRs) {
		t.Errorf("Fetch() got %d PRs, want %d PRs", len(gotPRs), len(expectedPRs))
		return
	}
	for i := range gotPRs {
		if gotPRs[i].Title != expectedPRs[i].Title || !gotPRs[i].MergedAt.Equal(expectedPRs[i].MergedAt) {
			t.Errorf("Fetch() PR at index %d: got %+v, want %+v", i, gotPRs[i], expectedPRs[i])
		}
	}
}

func TestFetcher_Fetch_ErrorResolvingRepos(t *testing.T) {
	cfg := &config.Config{Org: "", User: ""} // Invalid config for scope

	mockClient := &gh.MockClient{
		ListReposFunc: func(cfg *config.Config) ([]*github.Repository, error) {
			return nil, errors.New("error resolving repos")
		},
	}

	fetcher := NewFetcher(mockClient)

	_, err := fetcher.Fetch(cfg)

	if err == nil {
		t.Errorf("Fetch() expected an error, got nil")
	}
	
	if err.Error() != "failed to resolve repositories: exactly one of --org, --team, --user, or --repo must be specified" && err.Error() != "failed to resolve repositories: error resolving repos" {
		t.Errorf("Fetch() unexpected error message: %v", err)
	}
}

func TestFetcher_Fetch_ErrorParsingSince(t *testing.T) {
	cfg := &config.Config{Org: "test-org", Since: "invalid"}

	mockClient := &gh.MockClient{
		ListReposFunc: func(cfg *config.Config) ([]*github.Repository, error) {
			return []*github.Repository{{FullName: github.String("test-org/repo1")}}, nil
		},
	}

	fetcher := NewFetcher(mockClient)

	_, err := fetcher.Fetch(cfg)

	if err == nil {
		t.Errorf("Fetch() expected an error, got nil")
	}

	if err.Error() != "failed to parse 'since' duration: relative duration must be negative (e.g., -7d)" {
		t.Errorf("Fetch() unexpected error message: %v", err)
	}
}

func TestFetcher_Fetch_ErrorListingPRs(t *testing.T) {
	cfg := &config.Config{Org: "test-org", Since: "-7d"}

	mockClient := &gh.MockClient{
		ListReposFunc: func(cfg *config.Config) ([]*github.Repository, error) {
			return []*github.Repository{{FullName: github.String("test-org/repo1")}}, nil
		},
		ListPRsFunc: func(repo string, since time.Time) ([]model.PR, error) {
			return nil, errors.New("failed to list PRs")
		},
	}

	fetcher := NewFetcher(mockClient)

	gotPRs, err := fetcher.Fetch(cfg)

	if err != nil {
		t.Fatalf("Fetch() unexpected error: %v", err)
	}

	if len(gotPRs) != 0 {
		t.Errorf("Fetch() got %d PRs, want 0 PRs", len(gotPRs))
	}
}
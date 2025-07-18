package service

import (
	"fmt"
	"time"

	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/model"
	"github.com/willis7/prtool/internal/scope"
	"github.com/willis7/prtool/internal/timeutil"
)

// Fetcher handles fetching PRs from GitHub
type Fetcher struct {
	ghClient gh.GitHubClient
}

// NewFetcher creates a new PR fetcher
func NewFetcher(ghClient gh.GitHubClient) *Fetcher {
	return &Fetcher{
		ghClient: ghClient,
	}
}

// Fetch retrieves merged PRs from GitHub based on configuration
// It resolves the repository scope, applies the since filter, and returns only merged PRs
func (f *Fetcher) Fetch(cfg *config.Config) ([]*model.PR, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if f.ghClient == nil {
		return nil, fmt.Errorf("GitHub client is required")
	}

	// Parse the since filter
	var sinceTime time.Time
	if cfg.Since != "" {
		parsed, err := timeutil.ParseRelativeDuration(cfg.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid since filter '%s': %w", cfg.Since, err)
		}
		sinceTime = parsed
	} else {
		// Default to 7 days ago if no since filter is specified
		sinceTime = time.Now().AddDate(0, 0, -7)
	}

	// Resolve repositories based on scope
	repoNames, err := scope.ResolveRepos(cfg, f.ghClient)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repositories: %w", err)
	}

	// Fetch PRs from all repositories
	var allPRs []*model.PR
	for _, repoName := range repoNames {
		prs, err := f.ghClient.ListPRs(repoName, sinceTime)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PRs from repository '%s': %w", repoName, err)
		}

		// The GitHub client already filters by since date
		// We only need to filter for merged PRs (MergedAt != nil and State == "closed")
		for _, pr := range prs {
			if pr.MergedAt != nil && pr.State == "closed" {
				allPRs = append(allPRs, pr)
			}
		}
	}

	return allPRs, nil
}

// Fetch is a convenience function that creates a fetcher and fetches PRs
func Fetch(cfg *config.Config, ghClient gh.GitHubClient) ([]*model.PR, error) {
	fetcher := NewFetcher(ghClient)
	return fetcher.Fetch(cfg)
}

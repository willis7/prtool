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

var TimeNow = time.Now // Allow mocking for tests

// Fetcher fetches pull requests from GitHub.
type Fetcher struct {
	GitHubClient gh.GitHubClient
}

// NewFetcher creates a new Fetcher.
func NewFetcher(client gh.GitHubClient) *Fetcher {
	return &Fetcher{GitHubClient: client}
}

// Fetch retrieves merged pull requests based on the configuration.
func (f *Fetcher) Fetch(cfg *config.Config) ([]model.PR, error) {
	// 1. Resolve repositories
	repoNames, err := scope.ResolveRepos(cfg, f.GitHubClient)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repositories: %w", err)
	}

	// 2. Parse 'since' duration
	sinceTime, err := timeutil.ParseRelativeDuration(cfg.Since)
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'since' duration: %w", err)
	}

	// 3. Fetch PRs for each repository
	var allPRs []model.PR
	for _, repoName := range repoNames {
		prs, err := f.GitHubClient.ListPRs(repoName, sinceTime)
		if err != nil {
			// Log error and continue with other repositories
			fmt.Printf("Warning: Failed to fetch PRs for %s: %v\n", repoName, err)
			continue
		}
		allPRs = append(allPRs, prs...)
	}

	return allPRs, nil
}

package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
	"github.com/yourorg/prtool/internal/model"
	"github.com/yourorg/prtool/internal/scope"
	"github.com/yourorg/prtool/internal/timeutil"
)

// Fetcher handles fetching and filtering pull requests
type Fetcher struct {
	client gh.GitHubClient
}

// NewFetcher creates a new PR fetcher
func NewFetcher(client gh.GitHubClient) *Fetcher {
	return &Fetcher{
		client: client,
	}
}

// Fetch retrieves merged PRs from GitHub based on configuration
func (f *Fetcher) Fetch(cfg *config.Config) ([]model.PR, error) {
	// Parse time range
	since, err := parseTimeRange(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	// Resolve repositories based on scope
	repos, err := scope.ResolveRepos(cfg.GitHub, f.client)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repositories: %w", err)
	}

	if cfg.Verbose {
		fmt.Printf("Fetching PRs from %d repositories since %s\n", len(repos), since.Format(time.RFC3339))
	}

	// Fetch PRs from each repository
	var allPRs []model.PR
	for _, repo := range repos {
		if cfg.Verbose {
			fmt.Printf("Processing repository: %s\n", repo)
		}

		prs, err := f.client.ListPRs(repo, since)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PRs from %s: %w", repo, err)
		}

		// Convert gh.PR to model.PR
		for _, pr := range prs {
			// Only include merged PRs
			if pr.MergedAt.IsZero() {
				continue
			}

			// Additional check for since filter
			if pr.MergedAt.Before(since) {
				continue
			}

			modelPR := model.PR{
				Repository:  repo,
				Number:      pr.Number,
				Title:       pr.Title,
				Description: pr.Description,
				Author:      pr.Author,
				URL:         pr.URL,
				MergedAt:    pr.MergedAt,
				Labels:      pr.Labels,
			}

			allPRs = append(allPRs, modelPR)
		}
	}

	// Sort PRs by merge date (newest first)
	sort.Slice(allPRs, func(i, j int) bool {
		return allPRs[i].MergedAt.After(allPRs[j].MergedAt)
	})

	if cfg.Verbose {
		fmt.Printf("Found %d merged PRs\n", len(allPRs))
	}

	return allPRs, nil
}

// parseTimeRange determines the since time based on configuration
func parseTimeRange(cfg *config.Config) (time.Time, error) {
	timeRange := cfg.TimeRange
	if timeRange == "" {
		timeRange = "-7d"
	}
	
	return timeutil.ParseRelativeDuration(timeRange)
}

// Fetch is a convenience function that creates a fetcher and fetches PRs
func Fetch(cfg *config.Config, client gh.GitHubClient) ([]model.PR, error) {
	fetcher := NewFetcher(client)
	return fetcher.Fetch(cfg)
}
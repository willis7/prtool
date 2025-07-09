package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/llm"
	"github.com/willis7/prtool/internal/model"
	"github.com/willis7/prtool/internal/scope"
	"github.com/willis7/prtool/internal/timeutil"
)

var TimeNow = time.Now // Allow mocking for tests

// Fetcher fetches pull requests from GitHub.
type Fetcher struct {
	GitHubClient gh.GitHubClient
	LLMClient    llm.LLM
}

// NewFetcher creates a new Fetcher.
func NewFetcher(ghClient gh.GitHubClient, llmClient llm.LLM) *Fetcher {
	return &Fetcher{GitHubClient: ghClient, LLMClient: llmClient}
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

// GenerateSummary generates a summary of the PRs using the LLM client.
func (f *Fetcher) GenerateSummary(prs []model.PR) (string, error) {
	if f.LLMClient == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	// Prepare context for LLM
	var contextBuilder strings.Builder
	for _, pr := range prs {
		contextBuilder.WriteString(fmt.Sprintf("Title: %s\n", pr.Title))
		contextBuilder.WriteString(fmt.Sprintf("Body: %s\n", pr.Body))
		contextBuilder.WriteString(fmt.Sprintf("URL: %s\n", pr.URL))
		contextBuilder.WriteString(fmt.Sprintf("Author: %s\n", pr.Author))
		contextBuilder.WriteString(fmt.Sprintf("Merged At: %s\n", pr.MergedAt.Format("2006-01-02")))
		contextBuilder.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(pr.Labels, ", ")))
		contextBuilder.WriteString("---\n")
	}

	return f.LLMClient.Summarise(context.Background(), contextBuilder.String())
}

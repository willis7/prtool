package scope

import (
	"fmt"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
)

// ResolveRepos determines which repositories to process based on the configuration scope
// It ensures exactly one scope (org, team, user, or repositories) is specified
func ResolveRepos(cfg config.GitHubConfig, client gh.GitHubClient) ([]string, error) {
	// Validate mutual exclusivity
	scopeCount := 0
	if cfg.Organization != "" {
		scopeCount++
	}
	if cfg.Team != "" {
		scopeCount++
	}
	if cfg.User != "" {
		scopeCount++
	}
	if len(cfg.Repositories) > 0 {
		scopeCount++
	}

	if scopeCount == 0 {
		return nil, fmt.Errorf("no scope specified: must specify one of --github-org, --github-team, --github-user, or --github-repos")
	}
	if scopeCount > 1 {
		return nil, fmt.Errorf("multiple scopes specified: must specify exactly one of --github-org, --github-team, --github-user, or --github-repos")
	}

	// If repositories are directly specified, return them
	if len(cfg.Repositories) > 0 {
		return cfg.Repositories, nil
	}

	// Otherwise, fetch repositories from GitHub based on the scope
	repos, err := client.ListRepos(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Convert repository objects to full names
	var repoNames []string
	for _, repo := range repos {
		if repo == nil || repo.FullName == nil {
			continue
		}
		repoNames = append(repoNames, *repo.FullName)
	}

	if len(repoNames) == 0 {
		return nil, fmt.Errorf("no repositories found for the specified scope")
	}

	return repoNames, nil
}
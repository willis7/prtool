package scope

import (
	"fmt"

	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
)

// ResolveRepos resolves the repository names based on the configuration scope
// It validates that exactly one scope is specified and returns a list of repository names
func ResolveRepos(cfg *config.Config, ghClient gh.GitHubClient) ([]string, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if ghClient == nil {
		return nil, fmt.Errorf("GitHub client is required")
	}

	// Validate that exactly one scope is specified
	scopeCount := 0
	var scopeType string

	if cfg.Org != "" {
		scopeCount++
		scopeType = "org"
	}
	if len(cfg.Team) > 0 {
		scopeCount++
		scopeType = "team"
	}
	if cfg.User != "" {
		scopeCount++
		scopeType = "user"
	}
	if cfg.Repo != "" {
		scopeCount++
		scopeType = "repo"
	}

	if scopeCount == 0 {
		return nil, fmt.Errorf("no scope specified: exactly one of org, team, user, or repo must be provided")
	}

	if scopeCount > 1 {
		return nil, fmt.Errorf("multiple scopes specified: only one of org, team, user, or repo is allowed")
	}

	// Fetch repositories using the GitHub client
	repos, err := ghClient.ListRepos(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for %s: %w", scopeType, err)
	}

	// Extract repository names in "owner/name" format
	var repoNames []string
	for _, repo := range repos {
		if repo.FullName != nil {
			repoNames = append(repoNames, *repo.FullName)
		} else if repo.Owner != nil && repo.Owner.Login != nil && repo.Name != nil {
			// Fallback: construct from owner login and repo name
			repoNames = append(repoNames, fmt.Sprintf("%s/%s", *repo.Owner.Login, *repo.Name))
		}
	}

	if len(repoNames) == 0 {
		return nil, fmt.Errorf("no repositories found for %s scope", scopeType)
	}

	return repoNames, nil
}

// ValidateScope validates that exactly one scope is specified in the configuration
func ValidateScope(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is required")
	}

	scopeCount := 0
	scopes := []string{}

	if cfg.Org != "" {
		scopeCount++
		scopes = append(scopes, "org")
	}
	if len(cfg.Team) > 0 {
		scopeCount++
		scopes = append(scopes, "team")
	}
	if cfg.User != "" {
		scopeCount++
		scopes = append(scopes, "user")
	}
	if cfg.Repo != "" {
		scopeCount++
		scopes = append(scopes, "repo")
	}

	if scopeCount == 0 {
		return fmt.Errorf("no scope specified: exactly one of org, team, user, or repo must be provided")
	}

	if scopeCount > 1 {
		return fmt.Errorf("multiple scopes specified: %v (only one allowed)", scopes)
	}

	return nil
}

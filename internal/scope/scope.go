package scope

import (
	"fmt"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
)

// ResolveRepos resolves the list of repositories based on the provided configuration.
// It ensures that exactly one of org, team, user, or repo is specified.
func ResolveRepos(cfg *config.Config, client gh.GitHubClient) ([]string, error) {
	var repos []*github.Repository
	var err error

	scopeCount := 0
	if cfg.Org != "" {
		scopeCount++
	}
	if cfg.Team != "" {
		scopeCount++
	}
	if cfg.User != "" {
		scopeCount++
	}
	if cfg.Repo != "" {
		scopeCount++
	}

	if scopeCount == 0 {
		return nil, fmt.Errorf("exactly one of --org, --team, --user, or --repo must be specified")
	} else if scopeCount > 1 {
		return nil, fmt.Errorf("only one of --org, --team, --user, or --repo can be specified")
	}

	if cfg.Org != "" {
		repos, err = client.ListRepos(cfg)
	} else if cfg.Team != "" {
		// This case is handled by ListRepos returning an error for now.
		repos, err = client.ListRepos(cfg)
	} else if cfg.User != "" {
		repos, err = client.ListRepos(cfg)
	} else if cfg.Repo != "" {
		repos, err = client.ListRepos(cfg)
	}

	if err != nil {
		return nil, err
	}

	var repoNames []string
	for _, repo := range repos {
		repoNames = append(repoNames, *repo.FullName)
	}

	return repoNames, nil
}

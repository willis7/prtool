package scope

import (
	"errors"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
)

func ResolveRepos(cfg config.Config, client gh.GitHubClient) ([]string, error) {
	fields := []string{cfg.Scope}
	set := 0
	for _, f := range fields {
		if f != "" {
			set++
		}
	}
	if set != 1 {
		return nil, errors.New("exactly one of org/team/user/repo must be set")
	}

	repos, err := client.ListRepos(cfg)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, r := range repos {
		if r.Name != nil {
			names = append(names, *r.Name)
		}
	}
	return names, nil
}

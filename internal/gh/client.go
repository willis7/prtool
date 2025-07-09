package gh

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/model"
	"golang.org/x/oauth2"
)

// GitHubClient defines the interface for interacting with the GitHub API.
type GitHubClient interface {
	ListRepos(cfg *config.Config) ([]*github.Repository, error)
	ListPRs(repo string, since time.Time) ([]model.PR, error)
}

// RestClient implements GitHubClient using the go-github library.
type RestClient struct {
	client *github.Client
}

// NewRestClient creates a new RestClient.
func NewRestClient(token string) *RestClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	c := oauth2.NewClient(ctx, ts)

	return &RestClient{client: github.NewClient(c)}
}

// ListRepos lists repositories based on the provided configuration.
func (c *RestClient) ListRepos(cfg *config.Config) ([]*github.Repository, error) {
	ctx := context.Background()

	if cfg.Org != "" {
		repos, _, err := c.client.Repositories.ListByOrg(ctx, cfg.Org, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for organization %s: %w", cfg.Org, err)
		}
		return repos, nil
	} else if cfg.User != "" {
		repos, _, err := c.client.Repositories.List(ctx, cfg.User, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for user %s: %w", cfg.User, err)
		}
		return repos, nil
	} else if cfg.Repo != "" {
		parts := strings.Split(cfg.Repo, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s. Expected owner/repo", cfg.Repo)
		}
		repo, _, err := c.client.Repositories.Get(ctx, parts[0], parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to get repository %s: %w", cfg.Repo, err)
		}
		return []*github.Repository{repo}, nil
	} else if cfg.Team != "" {
		// go-github does not directly support listing repos by team via a simple call.
		// This would require more complex logic involving team ID lookup.
		return nil, fmt.Errorf("listing repositories by team is not yet supported")
	}

	return nil, fmt.Errorf("no GitHub scope (org, user, repo, team) specified in configuration")
}

// ListPRs lists pull requests for a given repository since a specific time.
func (c *RestClient) ListPRs(repoFullName string, since time.Time) ([]model.PR, error) {
	ctx := context.Background()
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository full name: %s. Expected owner/repo", repoFullName)
	}
	owner, repoName := parts[0], parts[1]

	opts := &github.PullRequestListOptions{
		State:     "closed", // Only interested in merged PRs for now
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []model.PR
	for {
		prs, resp, err := c.client.PullRequests.List(ctx, owner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests for %s/%s: %w", owner, repoName, err)
		}

		for _, pr := range prs {
			if pr.MergedAt != nil && pr.MergedAt.After(since) {
				labels := make([]string, len(pr.Labels))
				for i, label := range pr.Labels {
					labels[i] = github.Stringify(label.Name)
				}
				allPRs = append(allPRs, model.PR{
					Title:     github.Stringify(pr.Title),
					Body:      github.Stringify(pr.Body),
					Author:    github.Stringify(pr.User.Login),
					CreatedAt: pr.CreatedAt.Time,
					MergedAt:  pr.MergedAt.Time,
					Labels:    labels,
					URL:       github.Stringify(pr.HTMLURL),
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

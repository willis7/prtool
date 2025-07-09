package gh

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/model"
)

// GitHubClient defines the interface for interacting with GitHub API
type GitHubClient interface {
	// ListRepos returns repositories based on the scope configuration
	ListRepos(scope *config.Config) ([]*github.Repository, error)

	// ListPRs returns pull requests for a given repository since a specific time
	ListPRs(repo string, since time.Time) ([]*model.PR, error)
}

// RestClient implements GitHubClient using the GitHub REST API
type RestClient struct {
	client *github.Client
	ctx    context.Context
}

// NewRestClient creates a new GitHub REST client with PAT authentication
func NewRestClient(token string) (*RestClient, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	client := github.NewClient(nil).WithAuthToken(token)

	// Test authentication by making a simple API call
	ctx := context.Background()
	_, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("GitHub authentication failed: %w", err)
	}

	return &RestClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// ListRepos returns repositories based on the scope configuration
func (c *RestClient) ListRepos(scope *config.Config) ([]*github.Repository, error) {
	if scope == nil {
		return nil, fmt.Errorf("scope configuration is required")
	}

	// Determine the scope type and fetch accordingly
	if scope.Org != "" {
		return c.listOrgRepos(scope.Org)
	} else if scope.User != "" {
		return c.listUserRepos(scope.User)
	} else if scope.Repo != "" {
		return c.getSingleRepo(scope.Repo)
	} else if scope.Team != "" {
		return c.listTeamRepos(scope.Team)
	}

	return nil, fmt.Errorf("no valid scope specified (org, user, repo, or team required)")
}

// ListPRs returns pull requests for a repository since a specific time
func (c *RestClient) ListPRs(repo string, since time.Time) ([]*model.PR, error) {
	if repo == "" {
		return nil, fmt.Errorf("repository name is required")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("repository must be in format 'owner/repo'")
	}

	owner, repoName := parts[0], parts[1]

	opts := &github.PullRequestListOptions{
		State: "closed", // We want merged PRs which are in closed state
		Sort:  "updated",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []*model.PR

	for {
		prs, resp, err := c.client.PullRequests.List(c.ctx, owner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests for %s: %w", repo, err)
		}

		for _, pr := range prs {
			// Only include merged PRs that were merged after the since time
			if pr.MergedAt != nil && pr.MergedAt.After(since) {
				modelPR := c.convertToModelPR(pr, repo)
				allPRs = append(allPRs, modelPR)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// Helper methods for different scope types
func (c *RestClient) listOrgRepos(org string) ([]*github.Repository, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Repositories.ListByOrg(c.ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for org %s: %w", org, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *RestClient) listUserRepos(user string) ([]*github.Repository, error) {
	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Repositories.List(c.ctx, user, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for user %s: %w", user, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *RestClient) getSingleRepo(repo string) ([]*github.Repository, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("repository must be in format 'owner/repo'")
	}

	owner, repoName := parts[0], parts[1]
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s: %w", repo, err)
	}

	return []*github.Repository{repository}, nil
}

func (c *RestClient) listTeamRepos(team string) ([]*github.Repository, error) {
	parts := strings.Split(team, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("team must be in format 'org/team'")
	}

	org, teamSlug := parts[0], parts[1]

	opts := &github.ListOptions{PerPage: 100}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Teams.ListTeamReposBySlug(c.ctx, org, teamSlug, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for team %s: %w", team, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// convertToModelPR converts a GitHub API PR to our internal model
func (c *RestClient) convertToModelPR(pr *github.PullRequest, repo string) *model.PR {
	modelPR := &model.PR{
		Title:      safeString(pr.Title),
		Body:       safeString(pr.Body),
		Author:     safeString(pr.User.Login),
		CreatedAt:  safeTimestamp(pr.CreatedAt),
		MergedAt:   safeTimestampPtr(pr.MergedAt),
		HTMLURL:    safeString(pr.HTMLURL),
		Number:     safeInt(pr.Number),
		Repository: repo,
		State:      safeString(pr.State),
	}
	
	// Extract labels
	for _, label := range pr.Labels {
		if label.Name != nil {
			modelPR.Labels = append(modelPR.Labels, *label.Name)
		}
	}
	
	// Note: Getting file paths requires additional API calls which we'll implement later if needed
	// For now, we'll leave this empty to keep the implementation simple
	
	return modelPR
}

// Helper functions for safe pointer dereferencing
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func safeTimestamp(t *github.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Time
}

func safeTimestampPtr(t *github.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	timeVal := t.Time
	return &timeVal
}

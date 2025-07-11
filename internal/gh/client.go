package gh

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
)

// PR represents a pull request with relevant information
type PR struct {
	Number      int
	Title       string
	Description string
	Author      string
	MergedAt    time.Time
	URL         string
	Labels      []string
}

// GitHubClient defines the interface for GitHub operations
type GitHubClient interface {
	ListRepos(scope config.GitHubConfig) ([]*github.Repository, error)
	ListPRs(repo string, since time.Time) ([]PR, error)
}

// RestClient implements GitHubClient using GitHub REST API
type RestClient struct {
	client *github.Client
	ctx    context.Context
}

// NewRestClient creates a new REST API client with PAT authentication
func NewRestClient(token string) (*RestClient, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	ctx := context.Background()
	client := github.NewTokenClient(ctx, token)

	// Check for custom API URL (for testing)
	if apiURL := os.Getenv("GITHUB_API_URL"); apiURL != "" {
		parsedURL, err := url.Parse(apiURL + "/")
		if err != nil {
			return nil, fmt.Errorf("invalid GITHUB_API_URL: %w", err)
		}
		client.BaseURL = parsedURL
	}

	return &RestClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// ListRepos returns repositories based on the provided scope configuration
func (c *RestClient) ListRepos(scope config.GitHubConfig) ([]*github.Repository, error) {
	var repos []*github.Repository

	// If specific repositories are listed, fetch them directly
	if len(scope.Repositories) > 0 {
		for _, repoFullName := range scope.Repositories {
			owner, repo, err := parseRepoFullName(repoFullName)
			if err != nil {
				return nil, fmt.Errorf("invalid repository name %s: %w", repoFullName, err)
			}

			repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
			if err != nil {
				if isAuthError(err) {
					return nil, fmt.Errorf("authentication failed: please check your GitHub token")
				}
				return nil, fmt.Errorf("failed to get repository %s: %w", repoFullName, err)
			}
			repos = append(repos, repository)
		}
		return repos, nil
	}

	// Otherwise, list repositories based on org/team/user
	if scope.Organization != "" {
		opts := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		for {
			orgRepos, resp, err := c.client.Repositories.ListByOrg(c.ctx, scope.Organization, opts)
			if err != nil {
				if isAuthError(err) {
					return nil, fmt.Errorf("authentication failed: please check your GitHub token")
				}
				return nil, fmt.Errorf("failed to list organization repositories: %w", err)
			}
			repos = append(repos, orgRepos...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	} else if scope.User != "" {
		opts := &github.RepositoryListOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		for {
			userRepos, resp, err := c.client.Repositories.List(c.ctx, scope.User, opts)
			if err != nil {
				if isAuthError(err) {
					return nil, fmt.Errorf("authentication failed: please check your GitHub token")
				}
				return nil, fmt.Errorf("failed to list user repositories: %w", err)
			}
			repos = append(repos, userRepos...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	}

	return repos, nil
}

// ListPRs returns merged pull requests for a repository since the given time
func (c *RestClient) ListPRs(repo string, since time.Time) ([]PR, error) {
	owner, repoName, err := parseRepoFullName(repo)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name %s: %w", repo, err)
	}

	opts := &github.PullRequestListOptions{
		State:       "closed",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var prs []PR
	for {
		pullRequests, resp, err := c.client.PullRequests.List(c.ctx, owner, repoName, opts)
		if err != nil {
			if isAuthError(err) {
				return nil, fmt.Errorf("authentication failed: please check your GitHub token")
			}
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}

		for _, pr := range pullRequests {
			// Skip if not merged or merged before the since time
			if pr.MergedAt == nil || pr.MergedAt.Before(since) {
				continue
			}

			// Convert GitHub PR to our PR type
			convertedPR := PR{
				Number:      pr.GetNumber(),
				Title:       pr.GetTitle(),
				Description: pr.GetBody(),
				Author:      pr.GetUser().GetLogin(),
				MergedAt:    pr.GetMergedAt().Time,
				URL:         pr.GetHTMLURL(),
			}

			// Extract labels
			for _, label := range pr.Labels {
				convertedPR.Labels = append(convertedPR.Labels, label.GetName())
			}

			prs = append(prs, convertedPR)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return prs, nil
}

// parseRepoFullName splits a repository full name (owner/repo) into owner and repo
func parseRepoFullName(fullName string) (owner, repo string, err error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("repository name must be in format 'owner/repo'")
	}
	return parts[0], parts[1], nil
}

// isAuthError checks if an error is an authentication error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	// Check for 401 Unauthorized or 403 Forbidden errors
	errResp, ok := err.(*github.ErrorResponse)
	if ok && (errResp.Response.StatusCode == 401 || errResp.Response.StatusCode == 403) {
		return true
	}
	return false
}
package gh

import (
	"context"
	"errors"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
	"golang.org/x/oauth2"
)

type PR struct {
	Title    string
	Number   int
	MergedAt time.Time
}

type GitHubClient interface {
	ListRepos(scope config.Config) ([]*github.Repository, error)
	ListPRs(repo string, since time.Time) ([]PR, error)
}

type RestClient struct {
	client *github.Client
}

func NewRestClient(token string) (*RestClient, error) {
	if token == "" {
		return nil, errors.New("missing GitHub token")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	return &RestClient{client: client}, nil
}

func (r *RestClient) ListRepos(scope config.Config) ([]*github.Repository, error) {
	// Minimal stub: real implementation would use scope
	return nil, errors.New("not implemented")
}

func (r *RestClient) ListPRs(repo string, since time.Time) ([]PR, error) {
	// Minimal stub: real implementation would use repo/since
	return nil, errors.New("not implemented")
}

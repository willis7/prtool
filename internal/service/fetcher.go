package service

import (
	"time"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
	"github.com/yourorg/prtool/internal/model"
)

func Fetch(cfg config.Config, client gh.GitHubClient, repo string, since time.Time) ([]model.PR, error) {
	prs, err := client.ListPRs(repo, since)
	if err != nil {
		return nil, err
	}
	var out []model.PR
	for _, pr := range prs {
		if pr.MergedAt.After(since) {
			out = append(out, model.PR{
				Title:    pr.Title,
				Number:   pr.Number,
				MergedAt: pr.MergedAt,
				Author:   "stub-author",
				Repo:     repo,
			})
		}
	}
	return out, nil
}

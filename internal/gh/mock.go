package gh

import (
	"errors"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/yourorg/prtool/internal/config"
)

type MockClient struct {
	FailAuth bool
}

func (m *MockClient) ListRepos(scope config.Config) ([]*github.Repository, error) {
	if m.FailAuth {
		return nil, errors.New("authentication failed: invalid token")
	}
	return []*github.Repository{{Name: github.String("mock-repo")}}, nil
}

func (m *MockClient) ListPRs(repo string, since time.Time) ([]PR, error) {
	if m.FailAuth {
		return nil, errors.New("authentication failed: invalid token")
	}
	return []PR{{Title: "Mock PR", Number: 1, MergedAt: time.Now()}}, nil
}

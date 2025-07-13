package gh

import (
	"testing"
	"time"

	"github.com/yourorg/prtool/internal/config"
)

func TestMockClient_AuthFailure(t *testing.T) {
	mc := &MockClient{FailAuth: true}
	_, err := mc.ListRepos(config.Config{})
	if err == nil || err.Error() != "authentication failed: invalid token" {
		t.Errorf("expected auth failure error, got %v", err)
	}
	_, err = mc.ListPRs("repo", time.Now())
	if err == nil || err.Error() != "authentication failed: invalid token" {
		t.Errorf("expected auth failure error, got %v", err)
	}
}

func TestMockClient_Success(t *testing.T) {
	mc := &MockClient{FailAuth: false}
	repos, err := mc.ListRepos(config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || *repos[0].Name != "mock-repo" {
		t.Errorf("unexpected repo result: %+v", repos)
	}
	prs, err := mc.ListPRs("repo", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 || prs[0].Title != "Mock PR" {
		t.Errorf("unexpected PR result: %+v", prs)
	}
}

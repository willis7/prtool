package scope

import (
	"testing"

	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
)

func TestResolveRepos_Validation(t *testing.T) {
	mc := &gh.MockClient{}
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{"none set", config.Config{}, true},
		{"multiple set", config.Config{Scope: "org"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveRepos(tt.cfg, mc)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveRepos_MockClient(t *testing.T) {
	mc := &gh.MockClient{}
	cfg := config.Config{Scope: "org"}
	repos, err := ResolveRepos(cfg, mc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0] != "mock-repo" {
		t.Errorf("unexpected repos: %+v", repos)
	}
}

func TestResolveRepos_AuthFailure(t *testing.T) {
	mc := &gh.MockClient{FailAuth: true}
	cfg := config.Config{Scope: "org"}
	_, err := ResolveRepos(cfg, mc)
	if err == nil || err.Error() != "authentication failed: invalid token" {
		t.Errorf("expected auth failure error, got %v", err)
	}
}

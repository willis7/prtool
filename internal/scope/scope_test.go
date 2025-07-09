package scope

import (
	"errors"
	"testing"

	"github.com/google/go-github/v55/github"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
)

func TestResolveRepos(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		mockListRepos func(cfg *config.Config) ([]*github.Repository, error)
		want          []string
		wantErr       bool
	}{
		{
			name: "Org scope - success",
			cfg:   &config.Config{Org: "test-org"},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return []*github.Repository{{FullName: github.String("test-org/repo1")}}, nil
			},
			want:    []string{"test-org/repo1"},
			wantErr: false,
		},
		{
			name: "User scope - success",
			cfg:   &config.Config{User: "test-user"},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return []*github.Repository{{FullName: github.String("test-user/repo2")}}, nil
			},
			want:    []string{"test-user/repo2"},
			wantErr: false,
		},
		{
			name: "Repo scope - success",
			cfg:   &config.Config{Repo: "test-owner/test-repo"},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return []*github.Repository{{FullName: github.String("test-owner/test-repo")}}, nil
			},
			want:    []string{"test-owner/test-repo"},
			wantErr: false,
		},
		{
			name: "Team scope - not supported",
			cfg:   &config.Config{Team: "test-org/test-team"},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return nil, errors.New("listing repositories by team is not yet supported")
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No scope specified",
			cfg:   &config.Config{},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return nil, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Multiple scopes specified",
			cfg:   &config.Config{Org: "test-org", User: "test-user"},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return nil, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GitHub client error",
			cfg:   &config.Config{Org: "test-org"},
			mockListRepos: func(cfg *config.Config) ([]*github.Repository, error) {
				return nil, errors.New("github API error")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &gh.MockClient{
				ListReposFunc: tt.mockListRepos,
			}
			got, err := ResolveRepos(tt.cfg, mockClient)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveRepos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ResolveRepos() got = %v, want %v", got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ResolveRepos() got = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

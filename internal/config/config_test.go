package config

import (
	"os"
	"testing"
)

func TestMergeConfigPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		cli, env, yaml *Config
		want           Config
	}{
		{
			name: "CLI overrides all",
			cli:  &Config{GitHubToken: "cli", Scope: "cli", LLMProvider: "cli"},
			env:  &Config{GitHubToken: "env", Scope: "env", LLMProvider: "env"},
			yaml: &Config{GitHubToken: "yaml", Scope: "yaml", LLMProvider: "yaml"},
			want: Config{GitHubToken: "cli", Scope: "cli", LLMProvider: "cli"},
		},
		{
			name: "Env overrides YAML",
			cli:  &Config{},
			env:  &Config{GitHubToken: "env", Scope: "env", LLMProvider: "env"},
			yaml: &Config{GitHubToken: "yaml", Scope: "yaml", LLMProvider: "yaml"},
			want: Config{GitHubToken: "env", Scope: "env", LLMProvider: "env"},
		},
		{
			name: "YAML fallback",
			cli:  &Config{},
			env:  &Config{},
			yaml: &Config{GitHubToken: "yaml", Scope: "yaml", LLMProvider: "yaml"},
			want: Config{GitHubToken: "yaml", Scope: "yaml", LLMProvider: "yaml"},
		},
		{
			name: "Empty config",
			cli:  &Config{},
			env:  &Config{},
			yaml: &Config{},
			want: Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeConfig(tt.cli, tt.env, tt.yaml)
			if got.GitHubToken != tt.want.GitHubToken || got.Scope != tt.want.Scope || got.LLMProvider != tt.want.LLMProvider {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PRTOOL_GITHUB_TOKEN", "envtoken")
	os.Setenv("PRTOOL_SCOPE", "envscope")
	os.Setenv("PRTOOL_LLM_PROVIDER", "envllm")
	cfg := LoadFromEnv()
	if cfg.GitHubToken != "envtoken" || cfg.Scope != "envscope" || cfg.LLMProvider != "envllm" {
		t.Errorf("env load failed: got %+v", cfg)
	}
}

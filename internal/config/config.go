package config

import (
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GitHubToken string `yaml:"github_token"`
	Scope       string `yaml:"scope"`
	LLMProvider string `yaml:"llm_provider"`
}

func LoadFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadFromEnv() *Config {
	return &Config{
		GitHubToken: os.Getenv("PRTOOL_GITHUB_TOKEN"),
		Scope:       os.Getenv("PRTOOL_SCOPE"),
		LLMProvider: os.Getenv("PRTOOL_LLM_PROVIDER"),
	}
}

func BindFlags(cmd *cobra.Command) *Config {
	cfg := &Config{}
	cmd.Flags().StringVar(&cfg.GitHubToken, "github-token", "", "GitHub token")
	cmd.Flags().StringVar(&cfg.Scope, "scope", "", "Scope")
	cmd.Flags().StringVar(&cfg.LLMProvider, "llm-provider", "", "LLM provider")
	return cfg
}

func MergeConfig(cli, env, yaml *Config) *Config {
	return &Config{
		GitHubToken: firstNonEmpty(cli.GitHubToken, env.GitHubToken, yaml.GitHubToken),
		Scope:       firstNonEmpty(cli.Scope, env.Scope, yaml.Scope),
		LLMProvider: firstNonEmpty(cli.LLMProvider, env.LLMProvider, yaml.LLMProvider),
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

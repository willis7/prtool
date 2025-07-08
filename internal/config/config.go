package config

import (
	"os"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// Config holds the application configuration.
// Order of precedence: CLI flags > Environment variables > YAML file.

type Config struct {
	// GitHub API scope
	Org  string `yaml:"org"`
	Team string `yaml:"team"`
	User string `yaml:"user"`
	Repo string `yaml:"repo"`

	// Time range
	Since string `yaml:"since"`

	// LLM provider
	LLMProvider string `yaml:"llm_provider"`
	LLMAPIKey   string `yaml:"llm_api_key"`
	LLMModel    string `yaml:"llm_model"`
	PromptFile  string `yaml:"prompt"`

	// Output options
	Output  string `yaml:"output"`
	CI      bool   `yaml:"ci"`
	DryRun  bool   `yaml:"dry_run"`
	Verbose bool   `yaml:"verbose"`

	// GitHub token
	GitHubToken string `yaml:"github_token"`
}

// LoadFromFile loads configuration from a YAML file.
func LoadFromFile(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *Config {
	return &Config{
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		Org:         os.Getenv("PRTOOL_ORG"),
		Team:        os.Getenv("PRTOOL_TEAM"),
		User:        os.Getenv("PRTOOL_USER"),
		Repo:        os.Getenv("PRTOOL_REPO"),
		Since:       os.Getenv("PRTOOL_SINCE"),
		LLMProvider: os.Getenv("PRTOOL_LLM_PROVIDER"),
		LLMAPIKey:   os.Getenv("PRTOOL_LLM_API_KEY"),
		LLMModel:    os.Getenv("PRTOOL_LLM_MODEL"),
		PromptFile:  os.Getenv("PRTOOL_PROMPT"),
		Output:      os.Getenv("PRTOOL_OUTPUT"),
		CI:          os.Getenv("PRTOOL_CI") != "",
		DryRun:      os.Getenv("PRTOOL_DRY_RUN") != "",
		Verbose:     os.Getenv("PRTOOL_VERBOSE") != "",
	}
}

// BindFlags binds Cobra flags to a Config struct.
func BindFlags(flags *pflag.FlagSet) *Config {
	config := &Config{}
	flags.StringVar(&config.Org, "org", "", "GitHub organization")
	flags.StringVar(&config.Team, "team", "", "GitHub team in org/team format")
	flags.StringVar(&config.User, "user", "", "GitHub user")
	flags.StringVar(&config.Repo, "repo", "", "GitHub repository in org/repo format")
	flags.StringVar(&config.Since, "since", "-7d", "Time range (e.g., -7d, -1m, -1yr)")
	flags.StringVar(&config.LLMProvider, "llm-provider", "openai", "LLM provider (openai or ollama)")
	flags.StringVar(&config.LLMAPIKey, "llm-api-key", "", "LLM API key")
	flags.StringVar(&config.LLMModel, "llm-model", "gpt-3.5-turbo", "LLM model")
	flags.StringVar(&config.PromptFile, "prompt", "", "Path to a custom prompt file")
	flags.StringVar(&config.Output, "output", "", "Output file path")
	flags.BoolVar(&config.CI, "ci", false, "CI mode (non-interactive)")
	flags.BoolVar(&config.DryRun, "dry-run", false, "Dry run (skip LLM and output)")
	flags.BoolVar(&config.Verbose, "verbose", false, "Verbose logging")
	flags.StringVar(&config.GitHubToken, "github-token", "", "GitHub token")
	return config
}

// MergeConfig merges configurations from CLI, environment, and YAML file.
// Precedence: CLI > Environment > YAML.
func MergeConfig(cli, env, yaml *Config) *Config {
	merged := &Config{}

	// Start with YAML config as the base
	if yaml != nil {
		*merged = *yaml
	}

	// Override with environment variables
	if env != nil {
		// A bit repetitive, but ensures correct precedence
		if env.Org != "" { merged.Org = env.Org }
		if env.Team != "" { merged.Team = env.Team }
		if env.User != "" { merged.User = env.User }
		if env.Repo != "" { merged.Repo = env.Repo }
		if env.Since != "" { merged.Since = env.Since }
		if env.LLMProvider != "" { merged.LLMProvider = env.LLMProvider }
		if env.LLMAPIKey != "" { merged.LLMAPIKey = env.LLMAPIKey }
		if env.LLMModel != "" { merged.LLMModel = env.LLMModel }
		if env.PromptFile != "" { merged.PromptFile = env.PromptFile }
		if env.Output != "" { merged.Output = env.Output }
		if env.CI { merged.CI = env.CI }
		if env.DryRun { merged.DryRun = env.DryRun }
		if env.Verbose { merged.Verbose = env.Verbose }
		if env.GitHubToken != "" { merged.GitHubToken = env.GitHubToken }
	}

	// Override with CLI flags
	if cli != nil {
		if cli.Org != "" { merged.Org = cli.Org }
		if cli.Team != "" { merged.Team = cli.Team }
		if cli.User != "" { merged.User = cli.User }
		if cli.Repo != "" { merged.Repo = cli.Repo }
		if cli.Since != "-7d" { merged.Since = cli.Since } // Default value check
		if cli.LLMProvider != "openai" { merged.LLMProvider = cli.LLMProvider } // Default value check
		if cli.LLMAPIKey != "" { merged.LLMAPIKey = cli.LLMAPIKey }
		if cli.LLMModel != "gpt-3.5-turbo" { merged.LLMModel = cli.LLMModel } // Default value check
		if cli.PromptFile != "" { merged.PromptFile = cli.PromptFile }
		if cli.Output != "" { merged.Output = cli.Output }
		if cli.CI { merged.CI = cli.CI }
		if cli.DryRun { merged.DryRun = cli.DryRun }
		if cli.Verbose { merged.Verbose = cli.Verbose }
		if cli.GitHubToken != "" { merged.GitHubToken = cli.GitHubToken }
	}

	return merged
}

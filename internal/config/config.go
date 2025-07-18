package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config represents the complete configuration for prtool
type Config struct {
	// GitHub configuration
	GitHubToken string `yaml:"github_token" env:"PRTOOL_GITHUB_TOKEN"`

	// Scope configuration (mutually exclusive)
	Org  string `yaml:"org" env:"PRTOOL_ORG"`
	Team string `yaml:"team" env:"PRTOOL_TEAM"`
	User string `yaml:"user" env:"PRTOOL_USER"`
	Repo string `yaml:"repo" env:"PRTOOL_REPO"`

	// Time range
	Since string `yaml:"since" env:"PRTOOL_SINCE"`

	// LLM configuration
	LLMProvider string `yaml:"llm_provider" env:"PRTOOL_LLM_PROVIDER"`
	LLMAPIKey   string `yaml:"llm_api_key" env:"PRTOOL_LLM_API_KEY"`
	LLMModel    string `yaml:"llm_model" env:"PRTOOL_LLM_MODEL"`
	Prompt      string `yaml:"prompt" env:"PRTOOL_PROMPT"`

	// Output configuration
	Output  string `yaml:"output" env:"PRTOOL_OUTPUT"`
	DryRun  bool   `yaml:"dry_run" env:"PRTOOL_DRY_RUN"`
	Verbose bool   `yaml:"verbose" env:"PRTOOL_VERBOSE"`
	CI      bool   `yaml:"ci" env:"PRTOOL_CI"`

	// Logging
	LogFile string `yaml:"log_file" env:"PRTOOL_LOG_FILE"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}

	// Expand home directory if needed
	if !filepath.IsAbs(path) && len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, nil // Return empty config if file doesn't exist
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config file %s: %w", path, err)
	}

	return &config, nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := &Config{
		GitHubToken: os.Getenv("PRTOOL_GITHUB_TOKEN"),
		Org:         os.Getenv("PRTOOL_ORG"),
		Team:        os.Getenv("PRTOOL_TEAM"),
		User:        os.Getenv("PRTOOL_USER"),
		Repo:        os.Getenv("PRTOOL_REPO"),
		Since:       os.Getenv("PRTOOL_SINCE"),
		LLMProvider: os.Getenv("PRTOOL_LLM_PROVIDER"),
		LLMAPIKey:   os.Getenv("PRTOOL_LLM_API_KEY"),
		LLMModel:    os.Getenv("PRTOOL_LLM_MODEL"),
		Prompt:      os.Getenv("PRTOOL_PROMPT"),
		Output:      os.Getenv("PRTOOL_OUTPUT"),
		DryRun:      os.Getenv("PRTOOL_DRY_RUN") == "true",
		Verbose:     os.Getenv("PRTOOL_VERBOSE") == "true",
		CI:          os.Getenv("PRTOOL_CI") == "true",
		LogFile:     os.Getenv("PRTOOL_LOG_FILE"),
	}

	return config
}

// MergeConfig merges configurations with precedence: CLI > env > YAML
func MergeConfig(cliConfig, envConfig, yamlConfig *Config) *Config {
	if cliConfig == nil {
		cliConfig = &Config{}
	}
	if envConfig == nil {
		envConfig = &Config{}
	}
	if yamlConfig == nil {
		yamlConfig = &Config{}
	}

	merged := &Config{}

	// GitHub configuration
	merged.GitHubToken = firstNonEmpty(cliConfig.GitHubToken, envConfig.GitHubToken, yamlConfig.GitHubToken)

	// Scope configuration
	merged.Org = firstNonEmpty(cliConfig.Org, envConfig.Org, yamlConfig.Org)
	merged.Team = firstNonEmpty(cliConfig.Team, envConfig.Team, yamlConfig.Team)
	merged.User = firstNonEmpty(cliConfig.User, envConfig.User, yamlConfig.User)
	merged.Repo = firstNonEmpty(cliConfig.Repo, envConfig.Repo, yamlConfig.Repo)

	// Time range
	merged.Since = firstNonEmpty(cliConfig.Since, envConfig.Since, yamlConfig.Since)

	// LLM configuration
	merged.LLMProvider = firstNonEmpty(cliConfig.LLMProvider, envConfig.LLMProvider, yamlConfig.LLMProvider)
	merged.LLMAPIKey = firstNonEmpty(cliConfig.LLMAPIKey, envConfig.LLMAPIKey, yamlConfig.LLMAPIKey)
	merged.LLMModel = firstNonEmpty(cliConfig.LLMModel, envConfig.LLMModel, yamlConfig.LLMModel)
	merged.Prompt = firstNonEmpty(cliConfig.Prompt, envConfig.Prompt, yamlConfig.Prompt)

	// Output configuration
	merged.Output = firstNonEmpty(cliConfig.Output, envConfig.Output, yamlConfig.Output)
	merged.DryRun = firstBool(cliConfig.DryRun, envConfig.DryRun, yamlConfig.DryRun)
	merged.Verbose = firstBool(cliConfig.Verbose, envConfig.Verbose, yamlConfig.Verbose)
	merged.CI = firstBool(cliConfig.CI, envConfig.CI, yamlConfig.CI)

	// Logging
	merged.LogFile = firstNonEmpty(cliConfig.LogFile, envConfig.LogFile, yamlConfig.LogFile)

	return merged
}

// firstNonEmpty returns the first non-empty string from the given values
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstBool returns the first true boolean or the last boolean if none are true
func firstBool(values ...bool) bool {
	for i, v := range values {
		if v || i == len(values)-1 {
			return v
		}
	}
	return false
}

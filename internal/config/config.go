package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GitHub    GitHubConfig  `yaml:"github"`
	LLM       LLMConfig     `yaml:"llm"`
	Output    OutputConfig  `yaml:"output"`
	Verbose   bool          `yaml:"verbose"`
	DryRun    bool          `yaml:"dry_run"`
	TimeRange string        `yaml:"time_range"`
	CI        bool          `yaml:"ci"`
	LogFile   string        `yaml:"log_file"`
}

type GitHubConfig struct {
	Token        string   `yaml:"token"`
	Organization string   `yaml:"organization"`
	Team         string   `yaml:"team"`
	User         string   `yaml:"user"`
	Repositories []string `yaml:"repositories"`
}

type LLMConfig struct {
	Provider    string `yaml:"provider"`
	Model       string `yaml:"model"`
	APIKey      string `yaml:"api_key"`
	BaseURL     string `yaml:"base_url"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int    `yaml:"max_tokens"`
}

type OutputConfig struct {
	Format string `yaml:"format"`
	File   string `yaml:"file"`
}

func New() *Config {
	return &Config{
		LLM: LLMConfig{
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Output: OutputConfig{
			Format: "markdown",
		},
	}
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := New()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return config, nil
}

func LoadFromEnv() *Config {
	config := New()

	// GitHub configuration
	if token := os.Getenv("PRTOOL_GITHUB_TOKEN"); token != "" {
		config.GitHub.Token = token
	}
	if org := os.Getenv("PRTOOL_GITHUB_ORGANIZATION"); org != "" {
		config.GitHub.Organization = org
	}
	if team := os.Getenv("PRTOOL_GITHUB_TEAM"); team != "" {
		config.GitHub.Team = team
	}
	if user := os.Getenv("PRTOOL_GITHUB_USER"); user != "" {
		config.GitHub.User = user
	}
	if repos := os.Getenv("PRTOOL_GITHUB_REPOSITORIES"); repos != "" {
		config.GitHub.Repositories = strings.Split(repos, ",")
	}

	// LLM configuration
	if provider := os.Getenv("PRTOOL_LLM_PROVIDER"); provider != "" {
		config.LLM.Provider = provider
	}
	if model := os.Getenv("PRTOOL_LLM_MODEL"); model != "" {
		config.LLM.Model = model
	}
	if apiKey := os.Getenv("PRTOOL_LLM_API_KEY"); apiKey != "" {
		config.LLM.APIKey = apiKey
	}
	if baseURL := os.Getenv("PRTOOL_LLM_BASE_URL"); baseURL != "" {
		config.LLM.BaseURL = baseURL
	}
	if temp := os.Getenv("PRTOOL_LLM_TEMPERATURE"); temp != "" {
		if val, err := strconv.ParseFloat(temp, 64); err == nil {
			config.LLM.Temperature = val
		}
	}
	if maxTokens := os.Getenv("PRTOOL_LLM_MAX_TOKENS"); maxTokens != "" {
		if val, err := strconv.Atoi(maxTokens); err == nil {
			config.LLM.MaxTokens = val
		}
	}

	// Output configuration
	if format := os.Getenv("PRTOOL_OUTPUT_FORMAT"); format != "" {
		config.Output.Format = format
	}
	if file := os.Getenv("PRTOOL_OUTPUT_FILE"); file != "" {
		config.Output.File = file
	}

	// General flags
	if verbose := os.Getenv("PRTOOL_VERBOSE"); verbose != "" {
		config.Verbose = verbose == "true" || verbose == "1"
	}
	if dryRun := os.Getenv("PRTOOL_DRY_RUN"); dryRun != "" {
		config.DryRun = dryRun == "true" || dryRun == "1"
	}
	if ci := os.Getenv("PRTOOL_CI"); ci != "" {
		config.CI = ci == "true" || ci == "1"
	}
	if logFile := os.Getenv("PRTOOL_LOG_FILE"); logFile != "" {
		config.LogFile = logFile
	}

	return config
}

func MergeConfig(base, overlay *Config) *Config {
	if base == nil {
		base = New()
	}
	if overlay == nil {
		return base
	}

	merged := *base

	// Merge GitHub config
	if overlay.GitHub.Token != "" {
		merged.GitHub.Token = overlay.GitHub.Token
	}
	if overlay.GitHub.Organization != "" {
		merged.GitHub.Organization = overlay.GitHub.Organization
	}
	if overlay.GitHub.Team != "" {
		merged.GitHub.Team = overlay.GitHub.Team
	}
	if overlay.GitHub.User != "" {
		merged.GitHub.User = overlay.GitHub.User
	}
	if len(overlay.GitHub.Repositories) > 0 {
		merged.GitHub.Repositories = overlay.GitHub.Repositories
	}

	// Merge LLM config
	if overlay.LLM.Provider != "" {
		merged.LLM.Provider = overlay.LLM.Provider
	}
	if overlay.LLM.Model != "" {
		merged.LLM.Model = overlay.LLM.Model
	}
	if overlay.LLM.APIKey != "" {
		merged.LLM.APIKey = overlay.LLM.APIKey
	}
	if overlay.LLM.BaseURL != "" {
		merged.LLM.BaseURL = overlay.LLM.BaseURL
	}
	if overlay.LLM.Temperature != 0 {
		merged.LLM.Temperature = overlay.LLM.Temperature
	}
	if overlay.LLM.MaxTokens != 0 {
		merged.LLM.MaxTokens = overlay.LLM.MaxTokens
	}

	// Merge Output config
	if overlay.Output.Format != "" {
		merged.Output.Format = overlay.Output.Format
	}
	if overlay.Output.File != "" {
		merged.Output.File = overlay.Output.File
	}

	// Merge general flags
	if overlay.Verbose {
		merged.Verbose = overlay.Verbose
	}
	if overlay.DryRun {
		merged.DryRun = overlay.DryRun
	}
	if overlay.TimeRange != "" {
		merged.TimeRange = overlay.TimeRange
	}
	if overlay.CI {
		merged.CI = overlay.CI
	}
	if overlay.LogFile != "" {
		merged.LogFile = overlay.LogFile
	}

	return &merged
}
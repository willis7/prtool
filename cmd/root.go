package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourorg/prtool/internal/config"
)

var (
	version    = "dev"
	cfgFile    string
	cfg        = config.New()
)

var rootCmd = &cobra.Command{
	Use:   "prtool",
	Short: "A tool for fetching and summarizing GitHub pull requests",
	Long:  `prtool fetches merged GitHub pull requests and generates AI-powered summaries.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.prtool.yaml)")
	rootCmd.PersistentFlags().BoolVar(&cfg.Verbose, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&cfg.DryRun, "dry-run", false, "dry run mode")

	// GitHub flags
	rootCmd.PersistentFlags().StringVar(&cfg.GitHub.Token, "github-token", "", "GitHub personal access token")
	rootCmd.PersistentFlags().StringVar(&cfg.GitHub.Organization, "github-org", "", "GitHub organization")
	rootCmd.PersistentFlags().StringVar(&cfg.GitHub.Team, "github-team", "", "GitHub team")
	rootCmd.PersistentFlags().StringVar(&cfg.GitHub.User, "github-user", "", "GitHub user")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.GitHub.Repositories, "github-repos", nil, "GitHub repositories")

	// LLM flags
	rootCmd.PersistentFlags().StringVar(&cfg.LLM.Provider, "llm-provider", "", "LLM provider (openai, ollama)")
	rootCmd.PersistentFlags().StringVar(&cfg.LLM.Model, "llm-model", "", "LLM model")
	rootCmd.PersistentFlags().StringVar(&cfg.LLM.APIKey, "llm-api-key", "", "LLM API key")
	rootCmd.PersistentFlags().StringVar(&cfg.LLM.BaseURL, "llm-base-url", "", "LLM base URL")

	// Output flags
	rootCmd.PersistentFlags().StringVar(&cfg.Output.Format, "output-format", "markdown", "output format")
	rootCmd.PersistentFlags().StringVar(&cfg.Output.File, "output-file", "", "output file")

	// Time range flag
	rootCmd.PersistentFlags().StringVar(&cfg.TimeRange, "since", "-7d", "time range for PRs (e.g., -7d, -1m, -1yr)")

	// Version flag
	rootCmd.Flags().BoolP("version", "v", false, "version for prtool")
	rootCmd.Version = version
}

func initConfig() {
	var fileConfig *config.Config
	var err error

	// Store CLI values
	cliConfig := cfg

	// Load from file if specified
	if cfgFile != "" {
		fileConfig, err = config.LoadFromFile(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Try default locations
		home, err := os.UserHomeDir()
		if err == nil {
			defaultPath := home + "/.prtool.yaml"
			if _, err := os.Stat(defaultPath); err == nil {
				fileConfig, _ = config.LoadFromFile(defaultPath)
			}
		}
	}

	// Load from environment
	envConfig := config.LoadFromEnv()

	// Merge configs: file -> env -> CLI (cfg already has CLI values)
	if fileConfig != nil {
		cfg = config.MergeConfig(fileConfig, envConfig)
		cfg = config.MergeConfig(cfg, cliConfig)
	} else {
		cfg = config.MergeConfig(envConfig, cliConfig)
	}
}

func GetConfig() *config.Config {
	return cfg
}
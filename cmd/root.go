package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/version"
)

var (
	appVersion   = "dev"
	cfgFile      string
	cfg          = config.New()
	versionCheck bool
)

var rootCmd = &cobra.Command{
	Use:   "prtool",
	Short: "A tool for fetching and summarizing GitHub pull requests",
	Long:  `prtool fetches merged GitHub pull requests and generates AI-powered summaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If --version-check is specified, it's handled in PersistentPreRunE
		// Otherwise, show help
		if !versionCheck {
			cmd.Help()
		}
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Handle version check if requested
		if versionCheck {
			return runVersionCheck()
		}
		return nil
	},
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
	rootCmd.Version = appVersion
	
	// Version check flag
	rootCmd.PersistentFlags().BoolVar(&versionCheck, "version-check", false, "check for latest version")
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

func runVersionCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("Current version: %s\n", appVersion)
	fmt.Println("Checking for updates...")

	checker := version.NewGitHubChecker()
	hasUpdate, latest, err := version.CheckForUpdate(ctx, checker, appVersion, "yourorg", "prtool")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
		os.Exit(0) // Exit cleanly after showing error
	}

	if hasUpdate {
		fmt.Printf("\nA new version is available: %s (published %s)\n", latest.TagName, latest.PublishedAt.Format("2006-01-02"))
		fmt.Printf("Release: %s\n", latest.HTMLURL)
		fmt.Println("\nTo update, run:")
		fmt.Println("  go install github.com/yourorg/prtool@latest")
	} else {
		fmt.Println("You are running the latest version!")
	}

	os.Exit(0) // Exit after version check
	return nil // This won't be reached but satisfies the compiler
}
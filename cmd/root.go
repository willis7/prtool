package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/willis7/prtool/internal/config"
	"github.com/willis7/prtool/internal/gh"
	"github.com/willis7/prtool/internal/llm"
	"github.com/willis7/prtool/internal/render"
	"github.com/willis7/prtool/internal/service"
)

var (
	version = "dev"

	cfgFile string
	cliConfig = &config.Config{}
)

var rootCmd = &cobra.Command{
	Use:     "prtool",
	Short:   "A tool to summarize pull requests.",
	Long:    "prtool is a command-line tool to fetch and summarize GitHub pull requests.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config from file, env, and merge with CLI flags
		var fileConfig *config.Config
		if cfgFile != "" {
			var err error
			fileConfig, err = config.LoadFromFile(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config from file %s: %w", cfgFile, err)
			}
		}
		envConfig := config.LoadFromEnv()
		finalConfig := config.MergeConfig(cliConfig, envConfig, fileConfig)

		// Initialize GitHub client and LLM client
		ghClient := gh.NewRestClient(finalConfig.GitHubToken)

		var llmClient llm.LLM
		switch finalConfig.LLMProvider {
		case "openai":
			llmClient = llm.NewOpenAIClient(finalConfig.LLMAPIKey, finalConfig.LLMModel)
		case "ollama":
			llmClient = llm.NewOllamaClient("http://localhost:11434", finalConfig.LLMModel)
		default:
			return fmt.Errorf("unsupported LLM provider: %s", finalConfig.LLMProvider)
		}
		fetcher := service.NewFetcher(ghClient, llmClient)

		// Fetch PRs
		prs, err := fetcher.Fetch(finalConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch pull requests: %w", err)
		}

		var summary string
		if !finalConfig.DryRun {
			summary, err = fetcher.GenerateSummary(prs)
			if err != nil {
				return fmt.Errorf("failed to generate summary: %w", err)
			}
		}

		if finalConfig.DryRun {
			fmt.Println("--- Dry Run: Fetched PRs ---")
			for _, pr := range prs {
				fmt.Printf("Title: %s\n", pr.Title)
				fmt.Printf("URL: %s\n", pr.URL)
				fmt.Printf("Author: %s\n", pr.Author)
				fmt.Printf("Merged At: %s\n", pr.MergedAt.Format("2006-01-02"))
				fmt.Printf("Labels: %s\n", pr.Labels)
				fmt.Println("--------------------------")
			}
			return nil
		}

		// Prepare metadata for rendering
		meta := render.Metadata{
			Timeframe: finalConfig.Since,
			PRCount:   len(prs),
			Date:      time.Now(),
		}

		// Determine scope string for metadata
		if finalConfig.Org != "" {
			meta.Scope = "org: " + finalConfig.Org
		} else if finalConfig.Team != "" {
			meta.Scope = "team: " + finalConfig.Team
		} else if finalConfig.User != "" {
			meta.Scope = "user: " + finalConfig.User
		} else if finalConfig.Repo != "" {
			meta.Scope = "repo: " + finalConfig.Repo
		}

		// Render Markdown
		markdownOutput := render.Render(meta, prs, summary)

		// Write to output
		if finalConfig.Output != "" {
			err := os.WriteFile(finalConfig.Output, []byte(markdownOutput), 0644)
			if err != nil {
				return fmt.Errorf("failed to write output to file %s: %w", finalConfig.Output, err)
			}
			fmt.Printf("PR summary written to %s\n", finalConfig.Output)
		} else {
			fmt.Println(markdownOutput)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	// Define global flags and bind them to cliConfig
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.prtool.yaml)")
	config.BindFlags(rootCmd.PersistentFlags())
}

func initConfig() {
	// This function is called after flags are parsed, so cliConfig is populated.
	// No need to do anything here for now, as config merging happens in RunE.
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourorg/prtool/internal/gh"
	"github.com/yourorg/prtool/internal/render"
	"github.com/yourorg/prtool/internal/scope"
	"github.com/yourorg/prtool/internal/service"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch and summarize merged pull requests",
	Long:  `Fetches merged pull requests from GitHub repositories and generates summaries.`,
	RunE:  runExecute,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runExecute(cmd *cobra.Command, args []string) error {
	// Get configuration
	config := GetConfig()

	// Create GitHub client
	if config.GitHub.Token == "" {
		return fmt.Errorf("GitHub token is required. Set via --github-token, PRTOOL_GITHUB_TOKEN, or config file")
	}

	client, err := gh.NewRestClient(config.GitHub.Token)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Fetch PRs
	if config.Verbose {
		fmt.Fprintln(os.Stderr, "Fetching pull requests...")
	}

	prs, err := service.Fetch(config, client)
	if err != nil {
		return fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	// Get repositories for metadata
	repos, err := scope.ResolveRepos(config.GitHub, client)
	if err != nil {
		return fmt.Errorf("failed to resolve repositories: %w", err)
	}

	// Handle dry-run mode
	if config.DryRun {
		output := render.RenderDryRun(prs)
		fmt.Print(output)
		return nil
	}

	// Generate summaries using LLM
	summarizer, err := service.NewSummarizer(config)
	if err != nil {
		return fmt.Errorf("failed to create summarizer: %w", err)
	}

	ctx := context.Background()
	prs, err = summarizer.SummarizePRs(ctx, prs, config.Verbose)
	if err != nil {
		return fmt.Errorf("failed to generate summaries: %w", err)
	}

	// Prepare metadata
	meta := render.Metadata{
		GeneratedAt:  time.Now(),
		TimeRange:    config.TimeRange,
		TotalPRs:     len(prs),
		Repositories: repos,
		LLMProvider:  config.LLM.Provider,
		LLMModel:     config.LLM.Model,
	}

	// Render output
	output := render.Render(meta, prs)

	// Write output
	if config.Output.File != "" {
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "Writing output to %s\n", config.Output.File)
		}
		
		err := os.WriteFile(config.Output.File, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "Successfully wrote %d bytes to %s\n", len(output), config.Output.File)
		}
	} else {
		// Write to stdout
		fmt.Print(output)
	}

	return nil
}
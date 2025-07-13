package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourorg/prtool/internal/config"
	"github.com/yourorg/prtool/internal/gh"
	"github.com/yourorg/prtool/internal/llm"
	"github.com/yourorg/prtool/internal/render"
	"github.com/yourorg/prtool/internal/service"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch PRs and render Markdown summary",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadFromEnv()
		client := &gh.MockClient{}
		prs, err := service.Fetch(*cfg, client, "mock-repo", time.Now().Add(-7*24*time.Hour))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching PRs: %v\n", err)
			os.Exit(1)
		}
		llmClient := llm.NewLLM(*cfg)
		summary, err := llmClient.Summarise("context")
		if err != nil {
			fmt.Fprintf(os.Stderr, "LLM error: %v\n", err)
			os.Exit(1)
		}
		meta := render.Metadata{Title: "PR Summary", Date: time.Now().Format("2006-01-02")}
		md := render.Render(meta, prs)
		md += "\n\n" + summary + "\n"
		fmt.Print(md)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

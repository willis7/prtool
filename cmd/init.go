package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate annotated YAML config in current directory",
	Run: func(cmd *cobra.Command, args []string) {
		path := "prtool.config.yaml"
		f, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		annotated := `# prtool configuration file
# github_token: Personal access token for GitHub API
# scope: org/team/user/repo (see docs)
# llm_provider: openai, ollama, stub

github_token: ""
scope: ""
llm_provider: "stub"
`
		if _, err := f.WriteString(annotated); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config written to %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

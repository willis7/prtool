package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prtool",
	Short: "A CLI tool for summarizing GitHub pull requests",
	Long: `prtool is a command-line tool that fetches GitHub pull requests (PRs) 
for a specified time period and scope (organization, team, user, or repository), 
summarizes them using an LLM (OpenAI or Ollama), and outputs the result in Markdown format.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	
	// Handle version flag
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println(version)
			return
		}
		
		// Show help if no flags provided
		cmd.Help()
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "prtool",
	Short:   "A tool to summarize pull requests.",
	Long:    `prtool is a command-line tool to fetch and summarize GitHub pull requests.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// Main logic will go here.
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

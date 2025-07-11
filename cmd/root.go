package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

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
	rootCmd.Flags().BoolP("version", "v", false, "version for prtool")
	rootCmd.Version = version
}
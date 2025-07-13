package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "prtool",
	Short: "prtool CLI",
	Run: func(cmd *cobra.Command, args []string) {
		// Default action: print help
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version")
	rootCmd.Flags().Bool("version-check", false, "Check for latest release on GitHub")
	rootCmd.Flags().Bool("ci", false, "CI mode: disables interactive output/spinners, sets exit codes")
	rootCmd.Flags().Bool("verbose", false, "Enable verbose logging")
	rootCmd.Flags().String("log-file", "", "Path to log file")

	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Println(version)
			os.Exit(0)
		}
		if vc, _ := cmd.Flags().GetBool("version-check"); vc {
			latest := getLatestRelease()
			fmt.Printf("Latest release: %s\n", latest)
			os.Exit(0)
		}
		if ci, _ := cmd.Flags().GetBool("ci"); ci {
			// In CI mode, disable interactive output/spinners
			os.Setenv("PRTOOL_CI", "1")
		}
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			os.Setenv("PRTOOL_VERBOSE", "1")
		}
		if logFile, _ := cmd.Flags().GetString("log-file"); logFile != "" {
			os.Setenv("PRTOOL_LOG_FILE", logFile)
		}
	}
}

func getLatestRelease() string {
	// In production, this would query GitHub API. For now, return stub.
	return "v0.1.0"
}

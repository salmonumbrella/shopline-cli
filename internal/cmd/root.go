package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "shopline",
	Short: "Shopline CLI - Interact with the Shopline API",
	Long:  `A command-line interface for the Shopline e-commerce platform API.`,
}

func Execute(version, commit, date string) error {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("store", "s", "", "Store profile name (or set SHOPLINE_STORE)")
	rootCmd.PersistentFlags().StringP("output", "o", getDefaultOutput(), "Output format: text, json (or set SHOPLINE_OUTPUT)")
	rootCmd.PersistentFlags().String("color", "auto", "Color mode: auto, always, never")
	rootCmd.PersistentFlags().String("query", "", "JQ filter for JSON output")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().Int("limit", 0, "Limit number of results")
	rootCmd.PersistentFlags().String("sort-by", "", "Field to sort by")
	rootCmd.PersistentFlags().Bool("desc", false, "Sort in descending order")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Preview changes without executing them")
}

// getDefaultOutput returns the default output format from SHOPLINE_OUTPUT env var or "text".
func getDefaultOutput() string {
	if output := os.Getenv("SHOPLINE_OUTPUT"); output == "json" || output == "text" {
		return output
	}
	return "text"
}

package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "shopline",
	Short: "Shopline CLI - Interact with the Shopline API",
	Long:  `A command-line interface for the Shopline e-commerce platform API.`,
}

func Execute(version, commit, date string) error {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
	setupRootCommand()
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("store", "s", "", "Store profile name (or set SHOPLINE_STORE)")
	rootCmd.PersistentFlags().StringP("output", "o", getDefaultOutput(), "Output format: text, json (or set SHOPLINE_OUTPUT)")
	rootCmd.PersistentFlags().String("color", "auto", "Color mode: auto, always, never")
	rootCmd.PersistentFlags().String("query", "", "JQ filter for JSON output")
	rootCmd.PersistentFlags().String("jq", "", "Alias for --query (JQ filter for JSON output)")
	rootCmd.PersistentFlags().String("fields", "", "Select fields in JSON output (shorthand for --query)")
	rootCmd.PersistentFlags().Bool("items-only", false, "For JSON list output, emit only the items array (no pagination envelope)")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().Int("limit", 0, "Limit number of results (sets page size for list commands)")
	rootCmd.PersistentFlags().String("sort-by", "", "Field to sort by")
	rootCmd.PersistentFlags().Bool("desc", false, "Sort in descending order")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Preview changes without executing them")
}

var rootSetupOnce sync.Once

func setupRootCommand() {
	rootSetupOnce.Do(func() {
		rootCmd.SetHelpCommand(helpCmd)
		rootCmd.PersistentPreRunE = chainPersistentPreRunE(rootCmd.PersistentPreRunE, preRunNormalizeIDs, preRunApplyLimit, preRunSetupQuery)
		applyDesirePathAliases(rootCmd)
	})
}

func chainPersistentPreRunE(existing func(*cobra.Command, []string) error, next ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	if existing == nil && len(next) == 1 {
		return next[0]
	}
	return func(cmd *cobra.Command, args []string) error {
		if existing != nil {
			if err := existing(cmd, args); err != nil {
				return err
			}
		}
		for _, fn := range next {
			if fn == nil {
				continue
			}
			if err := fn(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func preRunNormalizeIDs(cmd *cobra.Command, args []string) error {
	normalizeIDArgs(args)
	return normalizeIDFlags(cmd)
}

func preRunApplyLimit(cmd *cobra.Command, _ []string) error {
	return applyLimitToPageSize(cmd)
}

func preRunSetupQuery(cmd *cobra.Command, _ []string) error {
	// If any jq/fields are used, ensure JSON output so the flags "do something".
	output, _ := cmd.Flags().GetString("output")
	query, _ := cmd.Flags().GetString("query")
	jq, _ := cmd.Flags().GetString("jq")
	fields, _ := cmd.Flags().GetString("fields")

	needsJSON := query != "" || jq != "" || fields != ""
	if needsJSON && output != "json" {
		if cmd.Flags().Changed("output") {
			return fmt.Errorf("--jq/--query/--fields require --output json")
		}
		if err := cmd.Flags().Set("output", "json"); err != nil {
			return err
		}
	}

	if query != "" && jq != "" {
		return fmt.Errorf("--jq and --query cannot be used together (use one)")
	}

	effective := query
	if jq != "" {
		effective = jq
	}

	if fields != "" {
		if effective != "" {
			return fmt.Errorf("--fields and --query/--jq cannot be used together (use one)")
		}
		fs, err := parseFieldsFlag(fields)
		if err != nil {
			return err
		}
		effective = buildFieldsQuery(fs)
	}

	// Normalize to --query so the rest of the code only needs to read one flag.
	if effective != "" && effective != query {
		if err := cmd.Flags().Set("query", effective); err != nil {
			return err
		}
	}

	return nil
}

// getDefaultOutput returns the default output format from SHOPLINE_OUTPUT env var or "text".
func getDefaultOutput() string {
	if output := os.Getenv("SHOPLINE_OUTPUT"); output == "json" || output == "text" {
		return output
	}
	return "text"
}

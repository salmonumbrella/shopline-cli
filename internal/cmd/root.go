package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

//go:embed help.txt
var helpText string

var rootCmd = &cobra.Command{
	Use:   "spl",
	Short: "Shopline CLI - Interact with the Shopline API",
	Long:  `A command-line interface for the Shopline e-commerce platform API.`,
}

var (
	rootItemsOnly bool
	rootYes       bool
)

const outputModeFlagName = "output-mode"

func Execute(version, commit, date string) error {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
	setupRootCommand()
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("store", "s", "", "Store profile name (or set SHOPLINE_STORE)")
	rootCmd.PersistentFlags().StringP("output", "o", getDefaultOutput(), "Output format: text|json|jsonl|ndjson (env SHOPLINE_OUTPUT)")
	rootCmd.PersistentFlags().String(outputModeFlagName, "text", "Internal: requested output mode before normalization")
	_ = rootCmd.PersistentFlags().MarkHidden(outputModeFlagName)
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Shorthand for --output json")
	rootCmd.PersistentFlags().String("color", "auto", "Color mode: auto, always, never")
	rootCmd.PersistentFlags().StringP("query", "q", "", "JQ expression to filter JSON output")
	rootCmd.PersistentFlags().String("output-query", "", "Internal: resolved query expression for JSON output filtering")
	_ = rootCmd.PersistentFlags().MarkHidden("output-query")
	rootCmd.PersistentFlags().String("jq", "", "Alias for --query")
	rootCmd.PersistentFlags().String("query-file", "", "Read JQ expression from file ('-' for stdin)")
	rootCmd.PersistentFlags().StringP("fields", "F", "", "Select fields in JSON output (shorthand for --query)")
	rootCmd.PersistentFlags().BoolVar(&rootItemsOnly, "items-only", false, "Output only the items/results array when present (JSON output)")
	rootCmd.PersistentFlags().BoolVar(&rootItemsOnly, "results-only", false, "Alias for --items-only")
	rootCmd.PersistentFlags().BoolVarP(&rootYes, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&rootYes, "force", false, "Alias for --yes")
	rootCmd.PersistentFlags().Bool("no-input", false, "Disable interactive prompts")
	rootCmd.PersistentFlags().IntP("limit", "l", 0, "Limit number of results (0 = API defaults; sets page size for list commands)")
	rootCmd.PersistentFlags().String("sort-by", "", "Sort results by field")
	rootCmd.PersistentFlags().BoolP("desc", "D", false, "Sort in descending order")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Preview changes without executing them")
	rootCmd.PersistentFlags().String("admin-token", "", "Admin API token (env: SHOPLINE_ADMIN_TOKEN)")
	rootCmd.PersistentFlags().String("admin-merchant-id", "", "Admin merchant ID (env: SHOPLINE_ADMIN_MERCHANT_ID)")
}

var rootSetupOnce sync.Once

func setupRootCommand() {
	rootSetupOnce.Do(func() {
		rootCmd.SetHelpCommand(helpCmd)

		// Override root help with static help.txt; subcommands keep Cobra's default.
		defaultHelp := rootCmd.HelpFunc()
		rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			if cmd.Name() == rootCmd.Name() && !cmd.HasParent() {
				_, _ = fmt.Fprint(outWriter(cmd), helpText)
				return
			}
			defaultHelp(cmd, args)
		})

		rootCmd.PersistentPreRunE = chainPersistentPreRunE(rootCmd.PersistentPreRunE, preRunNormalizeIDs, preRunApplyLimit, preRunApplyNonInteractive, preRunSetupQuery)
		applyDesirePathAliases(rootCmd)
		applyRootFlagAliases(rootCmd)
		applyCommonFlagAliases(rootCmd)
		applyLocalYesShorthandRecursive(rootCmd)
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

func preRunApplyNonInteractive(cmd *cobra.Command, _ []string) error {
	force, _ := cmd.Flags().GetBool("force")
	if force {
		if err := cmd.Flags().Set("yes", "true"); err != nil {
			return err
		}
	}
	resultsOnly, _ := cmd.Flags().GetBool("results-only")
	if resultsOnly && !cmd.Flags().Changed("items-only") {
		if err := cmd.Flags().Set("items-only", "true"); err != nil {
			return err
		}
	}
	return nil
}

func preRunSetupQuery(cmd *cobra.Command, _ []string) error {
	if err := normalizeOutputFlag(cmd); err != nil {
		return err
	}

	// If any jq/fields are used, ensure JSON output so the flags "do something".
	output, _ := cmd.Flags().GetString("output")
	jsonOut, _ := cmd.Flags().GetBool("json")
	query := getOutputQuery(cmd)
	jq, _ := cmd.Flags().GetString("jq")
	queryFile, _ := cmd.Flags().GetString("query-file")
	fields, _ := cmd.Flags().GetString("fields")

	needsJSON := jsonOut || query != "" || jq != "" || queryFile != "" || fields != ""
	if needsJSON && output != "json" {
		if cmd.Flags().Changed("output") {
			return fmt.Errorf("--jq/--query/--query-file/--fields require --output json")
		}
		if err := cmd.Flags().Set("output", "json"); err != nil {
			return err
		}
		if err := setRequestedOutputMode(cmd, "json"); err != nil {
			return err
		}
	}

	if query != "" && jq != "" {
		return fmt.Errorf("--jq and --query cannot be used together (use one)")
	}
	if queryFile != "" && (query != "" || jq != "") {
		return fmt.Errorf("--query-file and --query/--jq cannot be used together (use one)")
	}

	effective := query
	if jq != "" {
		effective = jq
	}
	if queryFile != "" {
		loadedQuery, err := readQueryFile(cmd, queryFile)
		if err != nil {
			return err
		}
		effective = loadedQuery
	}

	if fields != "" {
		if effective != "" {
			return fmt.Errorf("--fields and --query/--jq/--query-file cannot be used together (use one)")
		}
		fs, err := parseFieldsWithPresets(cmd, fields)
		if err != nil {
			return err
		}
		effective = buildFieldsQuery(fs)
	}

	// Persist normalized output query in internal storage used by formatter/input layers.
	// Some unit tests construct minimal commands without this hidden flag.
	if cmd.Flags().Lookup("output-query") != nil {
		if err := cmd.Flags().Set("output-query", effective); err != nil {
			return err
		}
	}

	// Keep legacy --query behavior for commands without local --query collisions.
	if shouldUseLegacyOutputQuery(cmd) && effective != "" && effective != query {
		if err := cmd.Flags().Set("query", effective); err != nil {
			return err
		}
	}

	return nil
}

func getOutputQuery(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	val, _ := cmd.Flags().GetString("query")
	return val
}

func shouldUseLegacyOutputQuery(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if cmd == cmd.Root() || cmd.Parent() == nil {
		return true
	}
	return cmd.LocalFlags().Lookup("query") == nil
}

func applyLocalYesShorthandRecursive(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	if yes := cmd.LocalFlags().Lookup("yes"); yes != nil && yes.Shorthand == "" && cmd.LocalFlags().ShorthandLookup("y") == nil {
		yes.Shorthand = "y"
	}

	for _, sub := range cmd.Commands() {
		applyLocalYesShorthandRecursive(sub)
	}
}

func normalizeOutputFlag(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}

	output, _ := cmd.Flags().GetString("output")
	if err := setRequestedOutputMode(cmd, output); err != nil {
		return err
	}
	normalized := normalizeOutputValue(output)
	if normalized == output {
		return nil
	}
	return cmd.Flags().Set("output", normalized)
}

func normalizeOutputValue(value string) string {
	normalized := normalizeRequestedOutputValue(value)
	switch normalized {
	case "json", "jsonl", "ndjson":
		return "json"
	case "text":
		return "text"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizeRequestedOutputValue(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "json", "jsonl", "ndjson", "text":
		return normalized
	default:
		return strings.TrimSpace(value)
	}
}

func setRequestedOutputMode(cmd *cobra.Command, value string) error {
	if cmd == nil {
		return nil
	}
	if cmd.Flags().Lookup(outputModeFlagName) == nil {
		return nil
	}
	mode := normalizeRequestedOutputValue(value)
	if mode == "" {
		mode = "text"
	}
	return cmd.Flags().Set(outputModeFlagName, mode)
}

const maxQueryFileSize = 1 << 20 // 1MB

func readQueryFile(cmd *cobra.Command, path string) (string, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		return "", fmt.Errorf("--query-file cannot be empty")
	}

	var (
		b   []byte
		err error
	)
	if target == "-" {
		b, err = io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxQueryFileSize+1))
		if err != nil {
			return "", fmt.Errorf("failed to read --query-file from stdin: %w", err)
		}
	} else {
		b, err = os.ReadFile(target)
		if err != nil {
			return "", fmt.Errorf("failed to read --query-file: %w", err)
		}
	}

	if len(b) > maxQueryFileSize {
		return "", fmt.Errorf("--query-file too large (%d bytes, max %d)", len(b), maxQueryFileSize)
	}

	query := strings.TrimSpace(string(b))
	if query == "" {
		return "", fmt.Errorf("--query-file is empty")
	}
	return query, nil
}

// getDefaultOutput returns the default output format from SHOPLINE_OUTPUT env var or "text".
func getDefaultOutput() string {
	if output := normalizeRequestedOutputValue(os.Getenv("SHOPLINE_OUTPUT")); output == "json" || output == "jsonl" || output == "ndjson" || output == "text" {
		return output
	}
	return "text"
}

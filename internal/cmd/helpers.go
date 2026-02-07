package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resolveOrArg returns the first positional arg, or resolves via --by flag.
func resolveOrArg(cmd *cobra.Command, args []string, resolver func(query string) (string, error)) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	by, _ := cmd.Flags().GetString("by")
	if by == "" {
		return "", fmt.Errorf("provide a resource ID as argument, or use --by to search by name/email")
	}
	id, err := resolver(by)
	if err != nil {
		return "", err
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Resolved to %s\n", id)
	return id, nil
}

// enrichError wraps an API error with suggestions based on resource context.
func enrichError(err error, resource, resourceID string) error {
	return api.EnrichError(err, resource, resourceID)
}

// handleError formats and prints a rich error, returning the enriched error.
func handleError(cmd *cobra.Command, err error, resource, resourceID string) error {
	enriched := enrichError(err, resource, resourceID)
	formatted := api.FormatRichError(enriched)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), formatted)
	return enriched
}

var nonResourceCommands = map[string]struct{}{
	"auth":       {},
	"schema":     {},
	"docs":       {},
	"completion": {},
	"help":       {},
}

func idPrefixForCommand(cmd *cobra.Command) string {
	if cmd == nil || cmd.Name() != "list" {
		return ""
	}
	parent := cmd.Parent()
	if parent == nil {
		return ""
	}
	if _, skip := nonResourceCommands[parent.Name()]; skip {
		return ""
	}
	if res, ok := schema.Get(parent.Name()); ok && res.IDPrefix != "" {
		return res.IDPrefix
	}
	return parent.Name()
}

func normalizeIDToken(token string) (string, bool) {
	_, id, ok := outfmt.ParseID(token)
	if !ok {
		return token, false
	}
	return id, true
}

func normalizeIDArgs(args []string) {
	for i, arg := range args {
		if id, ok := normalizeIDToken(arg); ok {
			args[i] = id
		}
	}
}

func isIDFlag(name string) bool {
	if name == "id" || name == "ids" {
		return true
	}
	return strings.HasSuffix(name, "-id") || strings.HasSuffix(name, "-ids")
}

func normalizeIDFlags(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}

	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !isIDFlag(f.Name) {
			return
		}

		switch f.Value.Type() {
		case "string":
			val, _ := cmd.Flags().GetString(f.Name)
			if id, ok := normalizeIDToken(val); ok {
				_ = cmd.Flags().Set(f.Name, id)
			}
		case "stringSlice":
			values, _ := cmd.Flags().GetStringSlice(f.Name)
			changed := false
			for i, v := range values {
				if id, ok := normalizeIDToken(v); ok {
					values[i] = id
					changed = true
				}
			}
			if changed {
				_ = cmd.Flags().Set(f.Name, strings.Join(values, ","))
			}
		}
	})

	return nil
}

func applyLimitToPageSize(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}

	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		limitFlag = cmd.InheritedFlags().Lookup("limit")
	}
	if limitFlag == nil || !limitFlag.Changed {
		return nil
	}
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("limit must be >= 0")
	}
	if cmd.Flags().Lookup("page-size") == nil {
		return nil
	}
	return cmd.Flags().Set("page-size", strconv.Itoa(limit))
}

func readSortOptions(cmd *cobra.Command) (string, string) {
	if cmd == nil {
		return "", ""
	}
	sortBy, _ := cmd.Flags().GetString("sort-by")
	if sortBy == "" {
		return "", ""
	}
	desc, _ := cmd.Flags().GetBool("desc")
	if desc {
		return sortBy, "desc"
	}
	return sortBy, "asc"
}

func parseTimeFlag(value, label string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t, err = time.Parse("2006-01-02", value)
		if err != nil {
			return nil, fmt.Errorf("invalid %s date format, use RFC3339 or YYYY-MM-DD: %w", label, err)
		}
	}
	return &t, nil
}

package cmd

import (
	"fmt"
	"os"
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

// getAdminClient creates an AdminClient from flags or environment variables.
// Used by admin CLI commands (orders, products, shipping, livestreams, messages).
func getAdminClient(cmd *cobra.Command) (*api.AdminClient, error) {
	if os.Getenv("SHOPLINE_ADMIN_BASE_URL") == "" {
		return nil, fmt.Errorf("admin API base URL required: set SHOPLINE_ADMIN_BASE_URL env var")
	}

	token, _ := cmd.Flags().GetString("admin-token")
	if token == "" {
		token = os.Getenv("SHOPLINE_ADMIN_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var")
	}

	merchantID, _ := cmd.Flags().GetString("admin-merchant-id")
	if merchantID == "" {
		merchantID = os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	}
	if merchantID == "" {
		return nil, fmt.Errorf("admin merchant ID required: set --admin-merchant-id or SHOPLINE_ADMIN_MERCHANT_ID env var")
	}

	return api.NewAdminClient(token, merchantID), nil
}

// aliasFlagValue wraps a pflag.Value so that setting the alias also marks the
// original flag as Changed. This is needed because cobra's MarkFlagRequired
// checks the original flag's Changed field, not the alias's.
type aliasFlagValue struct {
	pflag.Value
	original *pflag.Flag
}

func (v *aliasFlagValue) Set(s string) error {
	if err := v.Value.Set(s); err != nil {
		return err
	}
	v.original.Changed = true
	return nil
}

// flagAlias registers a hidden alias for an existing flag.
// Both flags share the same underlying Value, so setting either one sets both.
// The alias wraps the value to propagate Changed to the original flag,
// ensuring MarkFlagRequired works correctly with either name.
func flagAlias(fs *pflag.FlagSet, name, alias string) {
	f := fs.Lookup(name)
	if f == nil {
		// Defensive no-op: alias wiring should never crash the CLI.
		return
	}
	if fs.Lookup(alias) != nil {
		return // alias name already taken, skip silently
	}
	a := *f // shallow copy
	a.Name = alias
	a.Shorthand = ""
	a.Usage = ""
	a.Hidden = true
	a.Value = &aliasFlagValue{Value: f.Value, original: f}
	// Strip annotations (especially "required") — only the original should be checked.
	a.Annotations = map[string][]string{"alias-of": {name}}
	fs.AddFlag(&a)
}

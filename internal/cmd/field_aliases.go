package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/salmonumbrella/shopline-cli/internal/queryalias"
	"github.com/spf13/cobra"
)

func newFieldAliasesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "aliases",
		Aliases: []string{"al"},
		Short:   "List jq field aliases",
		Long:    "Show all available short aliases for JSON field names, usable in --jq expressions.\n\nFunction aliases: sl() = select()",
		RunE: func(cmd *cobra.Command, args []string) error {
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "ALIAS\tFIELD\n")
			for _, e := range queryalias.Entries() {
				_, _ = fmt.Fprintf(tw, "%s\t%s\n", e.Alias, e.Canonical)
			}
			return tw.Flush()
		},
	}
}

func init() {
	rootCmd.AddCommand(newFieldAliasesCmd())
}

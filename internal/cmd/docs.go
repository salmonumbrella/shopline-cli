package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation",
	Hidden: true,
}

var docsManCmd = &cobra.Command{
	Use:   "man [output-dir]",
	Short: "Generate man pages",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "./man"
		if len(args) > 0 {
			dir = args[0]
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		header := &doc.GenManHeader{
			Title:   "SHOPLINE",
			Section: "1",
			Source:  "Shopline CLI",
		}

		if err := doc.GenManTree(rootCmd, header, dir); err != nil {
			return fmt.Errorf("failed to generate man pages: %w", err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Man pages generated in %s\n", dir)
		return nil
	},
}

var docsMarkdownCmd = &cobra.Command{
	Use:   "markdown [output-dir]",
	Short: "Generate markdown documentation",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "./docs/cli"
		if len(args) > 0 {
			dir = args[0]
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := doc.GenMarkdownTree(rootCmd, dir); err != nil {
			return fmt.Errorf("failed to generate markdown docs: %w", err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Markdown docs generated in %s\n", dir)
		return nil
	},
}

func init() {
	docsCmd.AddCommand(docsManCmd)
	docsCmd.AddCommand(docsMarkdownCmd)
	rootCmd.AddCommand(docsCmd)
}

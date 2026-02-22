package cmd

import (
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema [resource]",
	Short: "Show available resources and their commands",
	Long: `Display information about available CLI resources.

Without arguments, lists all resources.
With a resource name, shows detailed information about that resource.

Examples:
  shopline schema              # List all resources
  shopline schema orders       # Show orders resource details
  shopline schema list         # Explicit list subcommand
  shopline schema get orders   # Explicit get subcommand
  shopline schema --output json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSchema,
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		formatter := getFormatter(cmd)
		return listResources(cmd, formatter)
	},
}

var schemaGetCmd = &cobra.Command{
	Use:   "get <resource>",
	Short: "Show details for a resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		formatter := getFormatter(cmd)
		return showResource(cmd, formatter, args[0])
	},
}

func init() {
	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaGetCmd)
	rootCmd.AddCommand(schemaCmd)
}

func runSchema(cmd *cobra.Command, args []string) error {
	formatter := getFormatter(cmd)

	if len(args) == 0 {
		return listResources(cmd, formatter)
	}
	return showResource(cmd, formatter, args[0])
}

func listResources(cmd *cobra.Command, formatter interface{ JSON(interface{}) error }) error {
	resources := schema.All()

	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "json" {
		return formatter.JSON(resources)
	}

	// Text output
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Available resources:")
	_, _ = fmt.Fprintln(out)

	for _, res := range resources {
		_, _ = fmt.Fprintf(out, "  %-20s %s\n", res.Name, res.Description)
		if len(res.Commands) > 0 {
			_, _ = fmt.Fprintf(out, "  %-20s commands: %s\n", "", strings.Join(res.Commands, ", "))
		}
		_, _ = fmt.Fprintln(out)
	}

	_, _ = fmt.Fprintf(out, "Use 'spl schema <resource>' for details.\n")
	return nil
}

func showResource(cmd *cobra.Command, formatter interface{ JSON(interface{}) error }, name string) error {
	res, ok := schema.Get(name)
	if !ok {
		return fmt.Errorf("unknown resource: %s\n\nRun 'spl schema' to see available resources", name)
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "json" {
		return formatter.JSON(res)
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Resource: %s\n", res.Name)
	_, _ = fmt.Fprintf(out, "Description: %s\n", res.Description)
	_, _ = fmt.Fprintf(out, "ID Prefix: [%s:$id]\n", res.IDPrefix)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Commands:")
	for _, c := range res.Commands {
		_, _ = fmt.Fprintf(out, "  spl %s %s\n", res.Name, c)
	}

	return nil
}

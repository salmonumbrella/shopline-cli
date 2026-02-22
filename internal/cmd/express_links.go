package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var expressLinksCmd = &cobra.Command{
	Use:   "express-links",
	Short: "Manage express links (via Admin API)",
}

var expressLinksCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new", "generate"},
	Short:   "Create an express link (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var req api.AdminCreateExpressLinkRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create express link") {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.CreateExpressLink(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create express link: %w", err)
		}
		return getFormatter(cmd).JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(expressLinksCmd)

	expressLinksCmd.AddCommand(expressLinksCreateCmd)
	addJSONBodyFlags(expressLinksCreateCmd)
	expressLinksCreateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	schema.Register(schema.Resource{
		Name:        "express-links",
		Description: "Manage express links (via Admin API)",
		Commands:    []string{"create"},
	})
}

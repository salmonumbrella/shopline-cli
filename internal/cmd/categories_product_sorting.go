package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var categoriesProductsSortingCmd = &cobra.Command{
	Use:   "products-sorting",
	Short: "Manage category product sorting",
}

var categoriesProductsSortingUpdateCmd = &cobra.Command{
	Use:   "update <category-id>",
	Short: "Bulk update category product sorting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update category product sorting for %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkUpdateCategoryProductSorting(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update category product sorting: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	categoriesCmd.AddCommand(categoriesProductsSortingCmd)
	categoriesProductsSortingCmd.AddCommand(categoriesProductsSortingUpdateCmd)
	addJSONBodyFlags(categoriesProductsSortingUpdateCmd)
}

package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var userCreditsCmd = &cobra.Command{
	Use:   "user-credits",
	Short: "Manage user store credits",
}

var userCreditsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List user store credits records",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListUserCredits(cmd.Context(), &api.UserCreditsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list user credits: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var userCreditsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Bulk update user store credits (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would bulk update user credits") {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.BulkUpdateUserCredits(cmd.Context(), body)
		if err != nil {
			return err
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(userCreditsCmd)

	userCreditsCmd.AddCommand(userCreditsListCmd)
	userCreditsListCmd.Flags().Int("page", 1, "Page number")
	userCreditsListCmd.Flags().Int("page-size", 20, "Results per page")

	userCreditsCmd.AddCommand(userCreditsBulkUpdateCmd)
	addJSONBodyFlags(userCreditsBulkUpdateCmd)

	schema.Register(schema.Resource{
		Name:        "user-credits",
		Description: "Manage user store credits",
		Commands:    []string{"list", "bulk-update"},
		IDPrefix:    "user_credit",
	})
}

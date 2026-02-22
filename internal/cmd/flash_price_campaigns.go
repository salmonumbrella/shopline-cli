package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var flashPriceCampaignsCmd = &cobra.Command{
	Use:   "flash-price-campaigns",
	Short: "Manage flash price campaigns (documented endpoints)",
}

var flashPriceCampaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List flash price campaigns (raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListFlashPriceCampaigns(cmd.Context(), &api.FlashPriceCampaignsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list flash price campaigns: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var flashPriceCampaignsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a flash price campaign (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetFlashPriceCampaign(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get flash price campaign: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var flashPriceCampaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a flash price campaign (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create flash price campaign") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.CreateFlashPriceCampaign(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create flash price campaign: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var flashPriceCampaignsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a flash price campaign (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update flash price campaign") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateFlashPriceCampaign(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update flash price campaign: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var flashPriceCampaignsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a flash price campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !confirmAction(cmd, fmt.Sprintf("Delete flash price campaign %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if checkDryRun(cmd, "[DRY-RUN] Would delete flash price campaign") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if err := client.DeleteFlashPriceCampaign(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete flash price campaign: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted flash price campaign %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(flashPriceCampaignsCmd)

	flashPriceCampaignsCmd.AddCommand(flashPriceCampaignsListCmd)
	flashPriceCampaignsListCmd.Flags().Int("page", 1, "Page number")
	flashPriceCampaignsListCmd.Flags().Int("page-size", 20, "Results per page")

	flashPriceCampaignsCmd.AddCommand(flashPriceCampaignsGetCmd)

	flashPriceCampaignsCmd.AddCommand(flashPriceCampaignsCreateCmd)
	addJSONBodyFlags(flashPriceCampaignsCreateCmd)

	flashPriceCampaignsCmd.AddCommand(flashPriceCampaignsUpdateCmd)
	addJSONBodyFlags(flashPriceCampaignsUpdateCmd)

	flashPriceCampaignsCmd.AddCommand(flashPriceCampaignsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "flash-price-campaigns",
		Description: "Manage flash price campaigns",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "flash_price_campaign",
	})
}

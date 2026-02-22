package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

// ============================
// customers store-credits
// ============================

var customersStoreCreditsCmd = &cobra.Command{
	Use:   "store-credits",
	Short: "Manage customer store credits",
}

var customersStoreCreditsListCmd = &cobra.Command{
	Use:   "list <customer-id>",
	Short: "Get customer store credit history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		resp, err := client.ListCustomerStoreCredits(cmd.Context(), args[0], page, perPage)
		if err != nil {
			return storeCreditError("list", args[0], err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersStoreCreditsUpdateCmd = &cobra.Command{
	Use:   "update <customer-id>",
	Short: "Update customer store credits",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update store credits for customer %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			raw, err := readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
			var scReq api.StoreCreditUpdateRequest
			if err := json.Unmarshal(raw, &scReq); err != nil {
				return fmt.Errorf("failed to parse store credit request: %w", err)
			}
			resp, err := client.UpdateCustomerStoreCredits(cmd.Context(), args[0], &scReq)
			if err != nil {
				return storeCreditError("update", args[0], err)
			}
			return getFormatter(cmd).JSON(resp)
		}

		if !cmd.Flags().Changed("value") {
			return fmt.Errorf("request body required (use --body/--body-file or provide --value)")
		}
		value, _ := cmd.Flags().GetInt("value")
		remarks, _ := cmd.Flags().GetString("remarks")
		expiresAt, _ := cmd.Flags().GetString("expires-at")

		req := &api.StoreCreditUpdateRequest{
			Value:     value,
			Remarks:   remarks,
			ExpiresAt: expiresAt,
			Type:      "manual_credit",
		}

		resp, err := client.UpdateCustomerStoreCredits(cmd.Context(), args[0], req)
		if err != nil {
			return storeCreditError("update", args[0], err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	customersCmd.AddCommand(customersStoreCreditsCmd)
	customersStoreCreditsCmd.AddCommand(customersStoreCreditsListCmd)
	customersStoreCreditsListCmd.Flags().Int("page", 0, "Page number")
	customersStoreCreditsListCmd.Flags().Int("per-page", 0, "Results per page")

	customersStoreCreditsCmd.AddCommand(customersStoreCreditsUpdateCmd)
	addJSONBodyFlags(customersStoreCreditsUpdateCmd)
	customersStoreCreditsUpdateCmd.Flags().Int("value", 0, "Credits to add (positive) or deduct (negative), -999999~999999 (ignored when --body/--body-file set)")
	customersStoreCreditsUpdateCmd.Flags().String("remarks", "", "Reason for credit adjustment, max 50 chars (ignored when --body/--body-file set)")
	customersStoreCreditsUpdateCmd.Flags().String("expires-at", "", "Expiry date, ISO 8601 (ignored when --body/--body-file set)")
}

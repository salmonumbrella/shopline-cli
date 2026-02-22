package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var storeCreditsCmd = &cobra.Command{
	Use:   "store-credits",
	Short: "Manage customer store credits",
}

var storeCreditsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customer store credit history",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := client.ListCustomerStoreCredits(cmd.Context(), customerID, page, perPage)
		if err != nil {
			return storeCreditError("list", customerID, err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var storeCreditsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add or deduct store credits for a customer",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}
		value, _ := cmd.Flags().GetInt("value")
		remarks, _ := cmd.Flags().GetString("remarks")
		expiresAt, _ := cmd.Flags().GetString("expires-at")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update store credits: customer=%s, value=%d, remarks=%s", customerID, value, remarks)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StoreCreditUpdateRequest{
			Value:     value,
			Remarks:   remarks,
			ExpiresAt: expiresAt,
			Type:      "manual_credit",
		}

		resp, err := client.UpdateCustomerStoreCredits(cmd.Context(), customerID, req)
		if err != nil {
			return storeCreditError("update", customerID, err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

// storeCreditError wraps store credit API errors with actionable hints.
// The most common failure mode is a 404 caused by an invalid customer ID,
// which Shopline returns as a bare NotFoundError with no message body.
func storeCreditError(op, customerID string, err error) error {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr.Status == 404 {
		return fmt.Errorf(
			"customer %s not found in this store (HTTP 404); verify the customer ID belongs to the selected store profile (run 'spl customers ls' to list customers)",
			customerID,
		)
	}
	return fmt.Errorf("failed to %s store credits: %w", op, err)
}

func init() {
	rootCmd.AddCommand(storeCreditsCmd)

	storeCreditsCmd.PersistentFlags().String("customer-id", "", "Customer ID")

	storeCreditsCmd.AddCommand(storeCreditsListCmd)
	storeCreditsListCmd.Flags().Int("page", 0, "Page number (0 = server default)")
	storeCreditsListCmd.Flags().Int("per-page", 0, "Results per page (0 = server default)")

	storeCreditsCmd.AddCommand(storeCreditsCreateCmd)
	storeCreditsCreateCmd.Flags().Int("value", 0, "Credits to add (positive) or deduct (negative)")
	storeCreditsCreateCmd.Flags().String("remarks", "", "Reason for adding or deducting credits (max 50 chars)")
	storeCreditsCreateCmd.Flags().String("expires-at", "", "Expiry date (ISO 8601, e.g. 2026-02-21)")
	_ = storeCreditsCreateCmd.MarkFlagRequired("customer-id")
	_ = storeCreditsCreateCmd.MarkFlagRequired("value")
	_ = storeCreditsCreateCmd.MarkFlagRequired("remarks")
}

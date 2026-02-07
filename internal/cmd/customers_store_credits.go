package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

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
	Use:     "list <customer-id>",
	Aliases: []string{"ls"},
	Short:   "Get customer store credit history",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomerStoreCredits(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get customer store credits: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersStoreCreditsUpdateCmd = &cobra.Command{
	Use:   "update <customer-id>",
	Short: "Update customer store credits",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update store credits for customer %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		var req json.RawMessage
		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			req, err = readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
		} else {
			if !cmd.Flags().Changed("amount") {
				return fmt.Errorf("request body required (use --body/--body-file or provide --amount)")
			}
			amount, _ := cmd.Flags().GetFloat64("amount")
			reason, _ := cmd.Flags().GetString("reason")
			req, err = json.Marshal(map[string]any{
				"amount": amount,
				"reason": reason,
			})
			if err != nil {
				return fmt.Errorf("failed to build request body: %w", err)
			}
		}

		resp, err := client.CreateCustomerStoreCredits(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update customer store credits: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	customersCmd.AddCommand(customersStoreCreditsCmd)
	customersStoreCreditsCmd.AddCommand(customersStoreCreditsListCmd)
	customersStoreCreditsCmd.AddCommand(customersStoreCreditsUpdateCmd)

	addJSONBodyFlags(customersStoreCreditsUpdateCmd)
	customersStoreCreditsUpdateCmd.Flags().Float64("amount", 0, "Credit amount (positive or negative) (ignored when --body/--body-file set)")
	customersStoreCreditsUpdateCmd.Flags().String("reason", "", "Reason for credit adjustment (ignored when --body/--body-file set)")
}

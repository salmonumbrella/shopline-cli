package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var abandonedCheckoutsCmd = &cobra.Command{
	Use:   "abandoned-checkouts",
	Short: "Manage abandoned checkouts",
}

var abandonedCheckoutsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List abandoned checkouts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		customerID, _ := cmd.Flags().GetString("customer-id")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.AbandonedCheckoutsListOptions{
			Page:       page,
			PageSize:   pageSize,
			Status:     status,
			CustomerID: customerID,
		}
		if from != "" {
			since, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			opts.Since = since
		}
		if to != "" {
			until, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			opts.Until = until
		}

		resp, err := client.ListAbandonedCheckouts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list abandoned checkouts: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "TOTAL", "ITEMS", "RECOVERY_EMAILS", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				c.ID,
				c.Email,
				c.TotalPrice + " " + c.Currency,
				fmt.Sprintf("%d", len(c.LineItems)),
				fmt.Sprintf("%d", c.RecoveryEmailSentCount),
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d abandoned checkouts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var abandonedCheckoutsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get abandoned checkout details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		checkout, err := client.GetAbandonedCheckout(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get abandoned checkout: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(checkout)
		}

		fmt.Printf("Checkout ID:       %s\n", checkout.ID)
		fmt.Printf("Email:             %s\n", checkout.Email)
		if checkout.Phone != "" {
			fmt.Printf("Phone:             %s\n", checkout.Phone)
		}
		fmt.Printf("Customer ID:       %s\n", checkout.CustomerID)
		fmt.Printf("Customer Locale:   %s\n", checkout.CustomerLocale)
		fmt.Printf("Total:             %s %s\n", checkout.TotalPrice, checkout.Currency)
		fmt.Printf("Subtotal:          %s %s\n", checkout.SubtotalPrice, checkout.Currency)
		fmt.Printf("Tax:               %s %s\n", checkout.TotalTax, checkout.Currency)
		if checkout.TotalDiscounts != "" && checkout.TotalDiscounts != "0" {
			fmt.Printf("Discounts:         %s %s\n", checkout.TotalDiscounts, checkout.Currency)
		}
		fmt.Printf("Recovery Emails:   %d\n", checkout.RecoveryEmailSentCount)
		if checkout.RecoveryURL != "" {
			fmt.Printf("Recovery URL:      %s\n", checkout.RecoveryURL)
		}
		if checkout.CompletedAt != nil {
			fmt.Printf("Completed:         %s\n", checkout.CompletedAt.Format(time.RFC3339))
		}
		if checkout.ClosedAt != nil {
			fmt.Printf("Closed:            %s\n", checkout.ClosedAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:           %s\n", checkout.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:           %s\n", checkout.UpdatedAt.Format(time.RFC3339))

		if len(checkout.LineItems) > 0 {
			fmt.Printf("\nLine Items (%d):\n", len(checkout.LineItems))
			for _, item := range checkout.LineItems {
				fmt.Printf("  - %s (%s) x%d @ %.2f\n",
					item.Title, item.VariantName, item.Quantity, item.Price)
			}
		}
		return nil
	},
}

var abandonedCheckoutsSendRecoveryCmd = &cobra.Command{
	Use:   "send-recovery <id>",
	Short: "Send recovery email for an abandoned checkout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.SendAbandonedCheckoutRecoveryEmail(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to send recovery email: %w", err)
		}

		fmt.Printf("Recovery email sent for checkout %s.\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(abandonedCheckoutsCmd)

	abandonedCheckoutsCmd.AddCommand(abandonedCheckoutsListCmd)
	abandonedCheckoutsListCmd.Flags().String("status", "", "Filter by status")
	abandonedCheckoutsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	abandonedCheckoutsListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	abandonedCheckoutsListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	abandonedCheckoutsListCmd.Flags().Int("page", 1, "Page number")
	abandonedCheckoutsListCmd.Flags().Int("page-size", 20, "Results per page")

	abandonedCheckoutsCmd.AddCommand(abandonedCheckoutsGetCmd)
	abandonedCheckoutsCmd.AddCommand(abandonedCheckoutsSendRecoveryCmd)

	schema.Register(schema.Resource{
		Name:        "abandoned-checkouts",
		Description: "Manage abandoned checkouts",
		Commands:    []string{"list", "get", "send-recovery"},
		IDPrefix:    "checkout",
	})
}

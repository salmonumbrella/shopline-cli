package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightAbandonedCheckout)
				return formatter.JSON(api.ListResponse[lightAbandonedCheckout]{
					Items:      lightItems,
					Pagination: resp.Pagination,
					Page:       resp.Page,
					PageSize:   resp.PageSize,
					TotalCount: resp.TotalCount,
					HasMore:    resp.HasMore,
				})
			}
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "TOTAL", "ITEMS", "RECOVERY_EMAILS", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("checkout", c.ID),
				c.Email,
				c.TotalPrice + " " + c.Currency,
				fmt.Sprintf("%d", len(c.LineItems)),
				fmt.Sprintf("%d", c.RecoveryEmailSentCount),
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d abandoned checkouts\n", len(resp.Items), resp.TotalCount)
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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightAbandonedCheckout(checkout))
			}
			return formatter.JSON(checkout)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Checkout ID:       %s\n", checkout.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:             %s\n", checkout.Email)
		if checkout.Phone != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Phone:             %s\n", checkout.Phone)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:       %s\n", checkout.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Locale:   %s\n", checkout.CustomerLocale)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total:             %s %s\n", checkout.TotalPrice, checkout.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Subtotal:          %s %s\n", checkout.SubtotalPrice, checkout.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax:               %s %s\n", checkout.TotalTax, checkout.Currency)
		if checkout.TotalDiscounts != "" && checkout.TotalDiscounts != "0" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Discounts:         %s %s\n", checkout.TotalDiscounts, checkout.Currency)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Recovery Emails:   %d\n", checkout.RecoveryEmailSentCount)
		if checkout.RecoveryURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Recovery URL:      %s\n", checkout.RecoveryURL)
		}
		if checkout.CompletedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Completed:         %s\n", checkout.CompletedAt.Format(time.RFC3339))
		}
		if checkout.ClosedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Closed:            %s\n", checkout.ClosedAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:           %s\n", checkout.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:           %s\n", checkout.UpdatedAt.Format(time.RFC3339))

		if len(checkout.LineItems) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLine Items (%d):\n", len(checkout.LineItems))
			for _, item := range checkout.LineItems {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (%s) x%d @ %.2f\n",
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Recovery email sent for checkout %s.\n", args[0])
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
	abandonedCheckoutsListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(abandonedCheckoutsListCmd.Flags(), "light", "li")

	abandonedCheckoutsCmd.AddCommand(abandonedCheckoutsGetCmd)
	abandonedCheckoutsGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(abandonedCheckoutsGetCmd.Flags(), "light", "li")
	abandonedCheckoutsCmd.AddCommand(abandonedCheckoutsSendRecoveryCmd)

	schema.Register(schema.Resource{
		Name:        "abandoned-checkouts",
		Description: "Manage abandoned checkouts",
		Commands:    []string{"list", "get", "send-recovery"},
		IDPrefix:    "checkout",
	})
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var subscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage customer subscriptions",
}

var subscriptionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		productID, _ := cmd.Flags().GetString("product-id")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.SubscriptionsListOptions{
			Page:       page,
			PageSize:   pageSize,
			CustomerID: customerID,
			ProductID:  productID,
			Status:     status,
		}

		resp, err := client.ListSubscriptions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list subscriptions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER ID", "PRODUCT ID", "STATUS", "INTERVAL", "PRICE", "NEXT BILLING", "CREATED"}
		var rows [][]string
		for _, s := range resp.Items {
			interval := s.Interval
			if s.IntervalCount > 1 {
				interval = fmt.Sprintf("%d %ss", s.IntervalCount, s.Interval)
			}
			nextBilling := ""
			if !s.NextBillingAt.IsZero() {
				nextBilling = s.NextBillingAt.Format("2006-01-02")
			}
			price := s.Price
			if s.Currency != "" {
				price = s.Price + " " + s.Currency
			}
			rows = append(rows, []string{
				outfmt.FormatID("subscription", s.ID),
				s.CustomerID,
				s.ProductID,
				string(s.Status),
				interval,
				price,
				nextBilling,
				s.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d subscriptions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var subscriptionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get subscription details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		subscription, err := client.GetSubscription(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get subscription: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(subscription)
		}

		nextBilling := "N/A"
		if !subscription.NextBillingAt.IsZero() {
			nextBilling = subscription.NextBillingAt.Format(time.RFC3339)
		}
		cancelledAt := "N/A"
		if !subscription.CancelledAt.IsZero() {
			cancelledAt = subscription.CancelledAt.Format(time.RFC3339)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Subscription ID:  %s\n", subscription.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:      %s\n", subscription.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:       %s\n", subscription.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Variant ID:       %s\n", subscription.VariantID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:           %s\n", subscription.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Interval:         %d %s(s)\n", subscription.IntervalCount, subscription.Interval)
		_, _ = fmt.Fprintf(outWriter(cmd), "Price:            %s %s\n", subscription.Price, subscription.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Next Billing:     %s\n", nextBilling)
		_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled At:     %s\n", cancelledAt)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:          %s\n", subscription.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", subscription.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var subscriptionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a subscription",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		productID, _ := cmd.Flags().GetString("product-id")
		variantID, _ := cmd.Flags().GetString("variant-id")
		interval, _ := cmd.Flags().GetString("interval")
		intervalCount, _ := cmd.Flags().GetInt("interval-count")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create subscription: customer=%s, product=%s, interval=%d %s(s)", customerID, productID, intervalCount, interval)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.SubscriptionCreateRequest{
			CustomerID:    customerID,
			ProductID:     productID,
			VariantID:     variantID,
			Interval:      interval,
			IntervalCount: intervalCount,
		}

		subscription, err := client.CreateSubscription(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(subscription)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created subscription: %s\n", subscription.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer:  %s\n", subscription.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product:   %s\n", subscription.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:    %s\n", subscription.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Interval:  %s\n", subscription.Interval)

		return nil
	},
}

var subscriptionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Cancel a subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel subscription: %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteSubscription(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to cancel subscription: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled subscription: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(subscriptionsCmd)

	subscriptionsCmd.AddCommand(subscriptionsListCmd)
	subscriptionsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	subscriptionsListCmd.Flags().String("product-id", "", "Filter by product ID")
	subscriptionsListCmd.Flags().String("status", "", "Filter by status (active, paused, cancelled, expired)")
	subscriptionsListCmd.Flags().Int("page", 1, "Page number")
	subscriptionsListCmd.Flags().Int("page-size", 20, "Results per page")

	subscriptionsCmd.AddCommand(subscriptionsGetCmd)

	subscriptionsCmd.AddCommand(subscriptionsCreateCmd)
	subscriptionsCreateCmd.Flags().String("customer-id", "", "Customer ID")
	subscriptionsCreateCmd.Flags().String("product-id", "", "Product ID")
	subscriptionsCreateCmd.Flags().String("variant-id", "", "Variant ID (optional)")
	subscriptionsCreateCmd.Flags().String("interval", "month", "Billing interval (day, week, month, year)")
	subscriptionsCreateCmd.Flags().Int("interval-count", 1, "Number of intervals between billings")
	_ = subscriptionsCreateCmd.MarkFlagRequired("customer-id")
	_ = subscriptionsCreateCmd.MarkFlagRequired("product-id")

	subscriptionsCmd.AddCommand(subscriptionsDeleteCmd)
}

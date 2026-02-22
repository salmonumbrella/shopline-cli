package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var productSubscriptionsCmd = &cobra.Command{
	Use:   "product-subscriptions",
	Short: "Manage product subscriptions",
}

var productSubscriptionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List product subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		productID, _ := cmd.Flags().GetString("product-id")
		customerID, _ := cmd.Flags().GetString("customer-id")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ProductSubscriptionsListOptions{
			Page:       page,
			PageSize:   pageSize,
			ProductID:  productID,
			CustomerID: customerID,
			Status:     status,
		}

		resp, err := client.ListProductSubscriptions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list product subscriptions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PRODUCT ID", "CUSTOMER ID", "STATUS", "FREQUENCY", "PRICE", "CYCLES", "NEXT BILLING"}
		var rows [][]string
		for _, s := range resp.Items {
			price := s.Price
			if s.Currency != "" {
				price = s.Price + " " + s.Currency
			}
			frequency := s.Frequency
			if s.FrequencyInterval > 1 {
				frequency = fmt.Sprintf("every %d %s", s.FrequencyInterval, s.Frequency)
			}
			cycles := fmt.Sprintf("%d/%d", s.CompletedCycles, s.TotalCycles)
			if s.TotalCycles == 0 {
				cycles = fmt.Sprintf("%d/unlimited", s.CompletedCycles)
			}
			nextBilling := s.NextBillingDate.Format("2006-01-02")
			if s.NextBillingDate.IsZero() {
				nextBilling = "-"
			}
			rows = append(rows, []string{
				outfmt.FormatID("product_subscription", s.ID),
				s.ProductID,
				s.CustomerID,
				s.Status,
				frequency,
				price,
				cycles,
				nextBilling,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d subscriptions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var productSubscriptionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get product subscription details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		subscription, err := client.GetProductSubscription(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product subscription: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(subscription)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Subscription ID:     %s\n", subscription.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:          %s\n", subscription.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Variant ID:          %s\n", subscription.VariantID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:         %s\n", subscription.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Selling Plan ID:     %s\n", subscription.SellingPlanID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:              %s\n", subscription.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Frequency:           %s\n", subscription.Frequency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Frequency Interval:  %d\n", subscription.FrequencyInterval)
		_, _ = fmt.Fprintf(outWriter(cmd), "Price:               %s %s\n", subscription.Price, subscription.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Quantity:            %d\n", subscription.Quantity)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Cycles:        %d\n", subscription.TotalCycles)
		_, _ = fmt.Fprintf(outWriter(cmd), "Completed Cycles:    %d\n", subscription.CompletedCycles)
		if !subscription.NextBillingDate.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Next Billing Date:   %s\n", subscription.NextBillingDate.Format(time.RFC3339))
		}
		if !subscription.LastBillingDate.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Last Billing Date:   %s\n", subscription.LastBillingDate.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:             %s\n", subscription.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:             %s\n", subscription.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var productSubscriptionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a product subscription",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product-id")
		variantID, _ := cmd.Flags().GetString("variant-id")
		customerID, _ := cmd.Flags().GetString("customer-id")
		sellingPlanID, _ := cmd.Flags().GetString("selling-plan-id")
		quantity, _ := cmd.Flags().GetInt("quantity")
		nextBillingDate, _ := cmd.Flags().GetString("next-billing-date")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create subscription for product %s, customer %s", productID, customerID)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.ProductSubscriptionCreateRequest{
			ProductID:       productID,
			VariantID:       variantID,
			CustomerID:      customerID,
			SellingPlanID:   sellingPlanID,
			Quantity:        quantity,
			NextBillingDate: nextBillingDate,
		}

		subscription, err := client.CreateProductSubscription(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create product subscription: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(subscription)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created subscription %s\n", subscription.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:      %s\n", subscription.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:     %s\n", subscription.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Selling Plan ID: %s\n", subscription.SellingPlanID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", subscription.Status)

		return nil
	},
}

var productSubscriptionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Cancel a product subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel subscription %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteProductSubscription(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to cancel product subscription: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled subscription %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(productSubscriptionsCmd)

	productSubscriptionsCmd.AddCommand(productSubscriptionsListCmd)
	productSubscriptionsListCmd.Flags().String("product-id", "", "Filter by product ID")
	productSubscriptionsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	productSubscriptionsListCmd.Flags().String("status", "", "Filter by status (active, paused, cancelled)")
	productSubscriptionsListCmd.Flags().Int("page", 1, "Page number")
	productSubscriptionsListCmd.Flags().Int("page-size", 20, "Results per page")

	productSubscriptionsCmd.AddCommand(productSubscriptionsGetCmd)

	productSubscriptionsCmd.AddCommand(productSubscriptionsCreateCmd)
	productSubscriptionsCreateCmd.Flags().String("product-id", "", "Product ID")
	productSubscriptionsCreateCmd.Flags().String("variant-id", "", "Variant ID (optional)")
	productSubscriptionsCreateCmd.Flags().String("customer-id", "", "Customer ID")
	productSubscriptionsCreateCmd.Flags().String("selling-plan-id", "", "Selling plan ID")
	productSubscriptionsCreateCmd.Flags().Int("quantity", 1, "Quantity")
	productSubscriptionsCreateCmd.Flags().String("next-billing-date", "", "Next billing date (YYYY-MM-DD)")
	_ = productSubscriptionsCreateCmd.MarkFlagRequired("product-id")
	_ = productSubscriptionsCreateCmd.MarkFlagRequired("customer-id")
	_ = productSubscriptionsCreateCmd.MarkFlagRequired("selling-plan-id")

	productSubscriptionsCmd.AddCommand(productSubscriptionsDeleteCmd)
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var discountCodesCmd = &cobra.Command{
	Use:   "discount-codes",
	Short: "Manage discount codes",
}

var discountCodesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List discount codes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		priceRuleID, _ := cmd.Flags().GetString("price-rule-id")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.DiscountCodesListOptions{
			Page:        page,
			PageSize:    pageSize,
			PriceRuleID: priceRuleID,
			Status:      status,
		}

		resp, err := client.ListDiscountCodes(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list discount codes: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CODE", "TYPE", "VALUE", "USAGE", "STATUS", "STARTS", "ENDS"}
		var rows [][]string
		for _, dc := range resp.Items {
			value := fmt.Sprintf("%.0f", dc.DiscountValue)
			if dc.DiscountType == "percentage" {
				value += "%"
			}
			usage := fmt.Sprintf("%d", dc.UsageCount)
			if dc.UsageLimit > 0 {
				usage = fmt.Sprintf("%d/%d", dc.UsageCount, dc.UsageLimit)
			}
			startsAt := "-"
			if !dc.StartsAt.IsZero() {
				startsAt = dc.StartsAt.Format("2006-01-02")
			}
			endsAt := "-"
			if !dc.EndsAt.IsZero() {
				endsAt = dc.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("discount_code", dc.ID),
				dc.Code,
				dc.DiscountType,
				value,
				usage,
				dc.Status,
				startsAt,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d discount codes\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var discountCodesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get discount code details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		discountCode, err := client.GetDiscountCode(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get discount code: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(discountCode)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Code ID: %s\n", discountCode.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code:             %s\n", discountCode.Code)
		if discountCode.PriceRuleID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Price Rule ID:    %s\n", discountCode.PriceRuleID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:    %s\n", discountCode.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value:   %.2f\n", discountCode.DiscountValue)
		if discountCode.MinPurchase > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Min Purchase:     %.2f\n", discountCode.MinPurchase)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Usage:            %d", discountCode.UsageCount)
		if discountCode.UsageLimit > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), " / %d", discountCode.UsageLimit)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:           %s\n", discountCode.Status)
		if !discountCode.StartsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:        %s\n", discountCode.StartsAt.Format(time.RFC3339))
		}
		if !discountCode.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:          %s\n", discountCode.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:          %s\n", discountCode.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var discountCodesLookupCmd = &cobra.Command{
	Use:   "lookup <code>",
	Short: "Lookup a discount code by code string",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		discountCode, err := client.GetDiscountCodeByCode(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to lookup discount code: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(discountCode)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Code ID: %s\n", discountCode.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code:             %s\n", discountCode.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount:         %.2f (%s)\n", discountCode.DiscountValue, discountCode.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:           %s\n", discountCode.Status)
		return nil
	},
}

var discountCodesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a discount code",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create discount code") {
			return nil
		}

		code, _ := cmd.Flags().GetString("code")
		priceRuleID, _ := cmd.Flags().GetString("price-rule-id")
		discountType, _ := cmd.Flags().GetString("discount-type")
		discountValue, _ := cmd.Flags().GetFloat64("discount-value")
		usageLimit, _ := cmd.Flags().GetInt("usage-limit")
		minPurchase, _ := cmd.Flags().GetFloat64("min-purchase")
		startsAt, _ := cmd.Flags().GetString("starts-at")
		endsAt, _ := cmd.Flags().GetString("ends-at")

		req := &api.DiscountCodeCreateRequest{
			Code:          code,
			PriceRuleID:   priceRuleID,
			DiscountType:  discountType,
			DiscountValue: discountValue,
			UsageLimit:    usageLimit,
			MinPurchase:   minPurchase,
		}

		if startsAt != "" {
			t, err := time.Parse(time.RFC3339, startsAt)
			if err != nil {
				return fmt.Errorf("invalid starts-at format (use RFC3339): %w", err)
			}
			req.StartsAt = t
		}

		if endsAt != "" {
			t, err := time.Parse(time.RFC3339, endsAt)
			if err != nil {
				return fmt.Errorf("invalid ends-at format (use RFC3339): %w", err)
			}
			req.EndsAt = t
		}

		discountCode, err := client.CreateDiscountCode(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create discount code: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(discountCode)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created discount code %s (code: %s)\n", discountCode.ID, discountCode.Code)
		return nil
	},
}

var discountCodesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a discount code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would delete discount code") {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete discount code %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteDiscountCode(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete discount code: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted discount code %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(discountCodesCmd)

	discountCodesCmd.AddCommand(discountCodesListCmd)
	discountCodesListCmd.Flags().Int("page", 1, "Page number")
	discountCodesListCmd.Flags().Int("page-size", 20, "Results per page")
	discountCodesListCmd.Flags().String("price-rule-id", "", "Filter by price rule ID")
	discountCodesListCmd.Flags().String("status", "", "Filter by status (active, inactive, expired)")

	discountCodesCmd.AddCommand(discountCodesGetCmd)
	discountCodesCmd.AddCommand(discountCodesLookupCmd)

	discountCodesCmd.AddCommand(discountCodesCreateCmd)
	discountCodesCreateCmd.Flags().String("code", "", "Discount code (required)")
	discountCodesCreateCmd.Flags().String("price-rule-id", "", "Price rule ID")
	discountCodesCreateCmd.Flags().String("discount-type", "", "Discount type: percentage or fixed_amount (required)")
	discountCodesCreateCmd.Flags().Float64("discount-value", 0, "Discount value (required)")
	discountCodesCreateCmd.Flags().Int("usage-limit", 0, "Maximum number of uses")
	discountCodesCreateCmd.Flags().Float64("min-purchase", 0, "Minimum purchase amount")
	discountCodesCreateCmd.Flags().String("starts-at", "", "Start date (RFC3339 format)")
	discountCodesCreateCmd.Flags().String("ends-at", "", "End date (RFC3339 format)")
	_ = discountCodesCreateCmd.MarkFlagRequired("code")
	_ = discountCodesCreateCmd.MarkFlagRequired("discount-type")
	_ = discountCodesCreateCmd.MarkFlagRequired("discount-value")

	discountCodesCmd.AddCommand(discountCodesDeleteCmd)
	discountCodesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var couponsCmd = &cobra.Command{
	Use:   "coupons",
	Short: "Manage coupons",
}

var couponsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List coupons",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.CouponsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListCoupons(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list coupons: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CODE", "TITLE", "TYPE", "VALUE", "USAGE", "STATUS", "STARTS", "ENDS"}
		var rows [][]string
		for _, c := range resp.Items {
			value := fmt.Sprintf("%.0f", c.DiscountValue)
			if c.DiscountType == "percentage" {
				value += "%"
			}
			usage := fmt.Sprintf("%d", c.UsageCount)
			if c.UsageLimit > 0 {
				usage = fmt.Sprintf("%d/%d", c.UsageCount, c.UsageLimit)
			}
			endsAt := "-"
			if !c.EndsAt.IsZero() {
				endsAt = c.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				c.ID,
				c.Code,
				c.Title,
				c.DiscountType,
				value,
				usage,
				c.Status,
				c.StartsAt.Format("2006-01-02"),
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d coupons\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var couponsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get coupon details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		coupon, err := client.GetCoupon(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get coupon: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(coupon)
		}

		fmt.Printf("Coupon ID:      %s\n", coupon.ID)
		fmt.Printf("Code:           %s\n", coupon.Code)
		fmt.Printf("Title:          %s\n", coupon.Title)
		fmt.Printf("Description:    %s\n", coupon.Description)
		fmt.Printf("Discount Type:  %s\n", coupon.DiscountType)
		fmt.Printf("Discount Value: %.2f\n", coupon.DiscountValue)
		if coupon.MinPurchase > 0 {
			fmt.Printf("Min Purchase:   %.2f\n", coupon.MinPurchase)
		}
		if coupon.MaxDiscount > 0 {
			fmt.Printf("Max Discount:   %.2f\n", coupon.MaxDiscount)
		}
		fmt.Printf("Usage:          %d", coupon.UsageCount)
		if coupon.UsageLimit > 0 {
			fmt.Printf(" / %d", coupon.UsageLimit)
		}
		fmt.Println()
		if coupon.PerCustomer > 0 {
			fmt.Printf("Per Customer:   %d\n", coupon.PerCustomer)
		}
		fmt.Printf("Status:         %s\n", coupon.Status)
		fmt.Printf("Starts At:      %s\n", coupon.StartsAt.Format(time.RFC3339))
		if !coupon.EndsAt.IsZero() {
			fmt.Printf("Ends At:        %s\n", coupon.EndsAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", coupon.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var couponsLookupCmd = &cobra.Command{
	Use:   "lookup <code>",
	Short: "Lookup a coupon by code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		coupon, err := client.GetCouponByCode(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to lookup coupon: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(coupon)
		}

		fmt.Printf("Coupon ID:      %s\n", coupon.ID)
		fmt.Printf("Code:           %s\n", coupon.Code)
		fmt.Printf("Title:          %s\n", coupon.Title)
		fmt.Printf("Discount:       %.2f (%s)\n", coupon.DiscountValue, coupon.DiscountType)
		fmt.Printf("Status:         %s\n", coupon.Status)
		return nil
	},
}

var couponsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a coupon",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		code, _ := cmd.Flags().GetString("code")
		discountType, _ := cmd.Flags().GetString("discount-type")
		discountValue, _ := cmd.Flags().GetFloat64("discount-value")
		title, _ := cmd.Flags().GetString("title")
		minPurchase, _ := cmd.Flags().GetFloat64("min-purchase")
		usageLimit, _ := cmd.Flags().GetInt("usage-limit")
		perCustomer, _ := cmd.Flags().GetInt("per-customer")

		req := &api.CouponCreateRequest{
			Code:          code,
			DiscountType:  discountType,
			DiscountValue: discountValue,
			Title:         title,
			MinPurchase:   minPurchase,
			UsageLimit:    usageLimit,
			PerCustomer:   perCustomer,
		}

		coupon, err := client.CreateCoupon(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create coupon: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(coupon)
		}

		fmt.Printf("Created coupon %s (code: %s)\n", coupon.ID, coupon.Code)
		return nil
	},
}

var couponsActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		coupon, err := client.ActivateCoupon(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate coupon: %w", err)
		}

		fmt.Printf("Activated coupon %s (status: %s)\n", coupon.ID, coupon.Status)
		return nil
	},
}

var couponsDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		coupon, err := client.DeactivateCoupon(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate coupon: %w", err)
		}

		fmt.Printf("Deactivated coupon %s (status: %s)\n", coupon.ID, coupon.Status)
		return nil
	},
}

var couponsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete coupon %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCoupon(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete coupon: %w", err)
		}

		fmt.Printf("Deleted coupon %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(couponsCmd)

	couponsCmd.AddCommand(couponsListCmd)
	couponsListCmd.Flags().Int("page", 1, "Page number")
	couponsListCmd.Flags().Int("page-size", 20, "Results per page")
	couponsListCmd.Flags().String("status", "", "Filter by status (active, inactive, expired)")

	couponsCmd.AddCommand(couponsGetCmd)
	couponsCmd.AddCommand(couponsLookupCmd)

	couponsCmd.AddCommand(couponsCreateCmd)
	couponsCreateCmd.Flags().String("code", "", "Coupon code (required)")
	couponsCreateCmd.Flags().String("discount-type", "", "Discount type: percentage or fixed_amount (required)")
	couponsCreateCmd.Flags().Float64("discount-value", 0, "Discount value (required)")
	couponsCreateCmd.Flags().String("title", "", "Coupon title")
	couponsCreateCmd.Flags().Float64("min-purchase", 0, "Minimum purchase amount")
	couponsCreateCmd.Flags().Int("usage-limit", 0, "Maximum number of uses")
	couponsCreateCmd.Flags().Int("per-customer", 0, "Maximum uses per customer")
	_ = couponsCreateCmd.MarkFlagRequired("code")
	_ = couponsCreateCmd.MarkFlagRequired("discount-type")
	_ = couponsCreateCmd.MarkFlagRequired("discount-value")

	couponsCmd.AddCommand(couponsActivateCmd)
	couponsCmd.AddCommand(couponsDeactivateCmd)

	couponsCmd.AddCommand(couponsDeleteCmd)
	couponsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

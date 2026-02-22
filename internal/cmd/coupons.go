package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightCoupon)
				return formatter.JSON(api.ListResponse[lightCoupon]{
					Items:      lightItems,
					Page:       resp.Page,
					PageSize:   resp.PageSize,
					TotalCount: resp.TotalCount,
					HasMore:    resp.HasMore,
				})
			}
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
				outfmt.FormatID("coupon", c.ID),
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
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d coupons\n", len(resp.Items), resp.TotalCount)
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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightCoupon(coupon))
			}
			return formatter.JSON(coupon)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Coupon ID:      %s\n", coupon.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code:           %s\n", coupon.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:          %s\n", coupon.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", coupon.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:  %s\n", coupon.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value: %.2f\n", coupon.DiscountValue)
		if coupon.MinPurchase > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Min Purchase:   %.2f\n", coupon.MinPurchase)
		}
		if coupon.MaxDiscount > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Max Discount:   %.2f\n", coupon.MaxDiscount)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Usage:          %d", coupon.UsageCount)
		if coupon.UsageLimit > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), " / %d", coupon.UsageLimit)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		if coupon.PerCustomer > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Per Customer:   %d\n", coupon.PerCustomer)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", coupon.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:      %s\n", coupon.StartsAt.Format(time.RFC3339))
		if !coupon.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:        %s\n", coupon.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", coupon.CreatedAt.Format(time.RFC3339))
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Coupon ID:      %s\n", coupon.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code:           %s\n", coupon.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:          %s\n", coupon.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount:       %.2f (%s)\n", coupon.DiscountValue, coupon.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", coupon.Status)
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
		if checkDryRun(cmd, "[DRY-RUN] Would create coupon") {
			return nil
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created coupon %s (code: %s)\n", coupon.ID, coupon.Code)
		return nil
	},
}

var couponsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update coupon %s", args[0])) {
			return nil
		}

		var req api.CouponUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		coupon, err := client.UpdateCoupon(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update coupon: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(coupon)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated coupon %s (status: %s)\n", coupon.ID, coupon.Status)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would activate coupon %s", args[0])) {
			return nil
		}

		coupon, err := client.ActivateCoupon(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate coupon: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Activated coupon %s (status: %s)\n", coupon.ID, coupon.Status)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would deactivate coupon %s", args[0])) {
			return nil
		}

		coupon, err := client.DeactivateCoupon(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate coupon: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deactivated coupon %s (status: %s)\n", coupon.ID, coupon.Status)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete coupon %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete coupon %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCoupon(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete coupon: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted coupon %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(couponsCmd)

	couponsCmd.AddCommand(couponsListCmd)
	couponsListCmd.Flags().Int("page", 1, "Page number")
	couponsListCmd.Flags().Int("page-size", 20, "Results per page")
	couponsListCmd.Flags().String("status", "", "Filter by status (active, inactive, expired)")
	couponsListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(couponsListCmd.Flags(), "light", "li")

	couponsCmd.AddCommand(couponsGetCmd)
	couponsGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(couponsGetCmd.Flags(), "light", "li")
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

	couponsCmd.AddCommand(couponsUpdateCmd)
	addJSONBodyFlags(couponsUpdateCmd)

	couponsCmd.AddCommand(couponsActivateCmd)
	couponsCmd.AddCommand(couponsDeactivateCmd)

	couponsCmd.AddCommand(couponsDeleteCmd)
	couponsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	schema.Register(schema.Resource{
		Name:        "coupons",
		Description: "Manage coupons",
		Commands:    []string{"list", "get", "lookup", "create", "update", "activate", "deactivate", "delete"},
		IDPrefix:    "coupon",
	})
}

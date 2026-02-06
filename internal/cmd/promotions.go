package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var promotionsCmd = &cobra.Command{
	Use:   "promotions",
	Short: "Manage promotions",
}

var promotionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		promoType, _ := cmd.Flags().GetString("type")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.PromotionsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Type:     promoType,
		}

		resp, err := client.ListPromotions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list promotions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "TYPE", "STATUS", "DISCOUNT", "USAGE", "STARTS", "ENDS"}
		var rows [][]string
		for _, p := range resp.Items {
			discount := fmt.Sprintf("%.0f", p.DiscountValue)
			if p.DiscountType == "percentage" {
				discount += "%"
			}
			usage := fmt.Sprintf("%d", p.UsageCount)
			if p.UsageLimit > 0 {
				usage = fmt.Sprintf("%d/%d", p.UsageCount, p.UsageLimit)
			}
			endsAt := "-"
			if !p.EndsAt.IsZero() {
				endsAt = p.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				p.ID,
				p.Title,
				p.Type,
				p.Status,
				discount,
				usage,
				p.StartsAt.Format("2006-01-02"),
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d promotions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var promotionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get promotion details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.GetPromotion(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get promotion: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(promotion)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Promotion ID:    %s\n", promotion.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:           %s\n", promotion.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", promotion.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:            %s\n", promotion.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", promotion.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:   %s\n", promotion.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value:  %.2f\n", promotion.DiscountValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Min Purchase:    %.2f\n", promotion.MinPurchase)
		_, _ = fmt.Fprintf(outWriter(cmd), "Usage:           %d", promotion.UsageCount)
		if promotion.UsageLimit > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), " / %d", promotion.UsageLimit)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:       %s\n", promotion.StartsAt.Format(time.RFC3339))
		if !promotion.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:         %s\n", promotion.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", promotion.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var promotionsActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.ActivatePromotion(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate promotion: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Activated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.DeactivatePromotion(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate promotion: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deactivated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Delete promotion %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
				return nil
			}
		}

		if err := client.DeletePromotion(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete promotion: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted promotion %s\n", args[0])
		return nil
	},
}

var promotionsCouponCenterCmd = &cobra.Command{
	Use:   "coupon-center",
	Short: "Get coupon center promotions (documented endpoint; raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetPromotionsCouponCenter(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get promotions coupon center: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(promotionsCmd)

	promotionsCmd.AddCommand(promotionsListCmd)
	promotionsListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")
	promotionsListCmd.Flags().String("type", "", "Filter by type")
	promotionsListCmd.Flags().Int("page", 1, "Page number")
	promotionsListCmd.Flags().Int("page-size", 20, "Results per page")

	promotionsCmd.AddCommand(promotionsGetCmd)
	promotionsCmd.AddCommand(promotionsActivateCmd)
	promotionsCmd.AddCommand(promotionsDeactivateCmd)
	promotionsCmd.AddCommand(promotionsDeleteCmd)
	promotionsCmd.AddCommand(promotionsCouponCenterCmd)
}

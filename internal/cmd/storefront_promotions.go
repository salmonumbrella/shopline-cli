package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var storefrontPromotionsCmd = &cobra.Command{
	Use:   "storefront-promotions",
	Short: "View storefront promotion information",
}

var storefrontPromotionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storefront promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		promoType, _ := cmd.Flags().GetString("type")
		discountType, _ := cmd.Flags().GetString("discount-type")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StorefrontPromotionsListOptions{
			Page:         page,
			PageSize:     pageSize,
			Status:       status,
			Type:         promoType,
			DiscountType: discountType,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListStorefrontPromotions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list storefront promotions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "TYPE", "DISCOUNT", "USAGE", "STATUS", "ENDS AT"}
		var rows [][]string
		for _, p := range resp.Items {
			discount := p.DiscountValue
			if p.DiscountType == "percentage" {
				discount += "%"
			}
			usage := fmt.Sprintf("%d", p.UsageCount)
			if p.UsageLimit > 0 {
				usage = fmt.Sprintf("%d/%d", p.UsageCount, p.UsageLimit)
			}
			endsAt := "No end"
			if p.EndsAt != nil {
				endsAt = p.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				p.ID,
				p.Title,
				p.DiscountType,
				discount,
				usage,
				p.Status,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d promotions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storefrontPromotionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get storefront promotion details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		byCode, _ := cmd.Flags().GetBool("by-code")

		var promo *api.StorefrontPromotion
		if byCode {
			promo, err = client.GetStorefrontPromotionByCode(cmd.Context(), args[0])
		} else {
			promo, err = client.GetStorefrontPromotion(cmd.Context(), args[0])
		}
		if err != nil {
			return fmt.Errorf("failed to get storefront promotion: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(promo)
		}

		fmt.Printf("Promotion ID:    %s\n", promo.ID)
		fmt.Printf("Title:           %s\n", promo.Title)
		fmt.Printf("Description:     %s\n", promo.Description)
		fmt.Printf("Type:            %s\n", promo.Type)
		fmt.Printf("Status:          %s\n", promo.Status)
		fmt.Printf("Discount Type:   %s\n", promo.DiscountType)
		fmt.Printf("Discount Value:  %s\n", promo.DiscountValue)
		if promo.Code != "" {
			fmt.Printf("Code:            %s\n", promo.Code)
		}
		if promo.MinPurchase != "" {
			fmt.Printf("Min Purchase:    %s\n", promo.MinPurchase)
		}
		if promo.MaxDiscount != "" {
			fmt.Printf("Max Discount:    %s\n", promo.MaxDiscount)
		}
		fmt.Printf("Usage:           %d", promo.UsageCount)
		if promo.UsageLimit > 0 {
			fmt.Printf(" / %d", promo.UsageLimit)
		}
		fmt.Println()
		if promo.CustomerLimit > 0 {
			fmt.Printf("Customer Limit:  %d per customer\n", promo.CustomerLimit)
		}
		fmt.Printf("Stackable:       %v\n", promo.Stackable)
		fmt.Printf("Auto Apply:      %v\n", promo.AutoApply)
		fmt.Printf("Target Type:     %s\n", promo.TargetType)
		fmt.Printf("Starts At:       %s\n", promo.StartsAt.Format(time.RFC3339))
		if promo.EndsAt != nil {
			fmt.Printf("Ends At:         %s\n", promo.EndsAt.Format(time.RFC3339))
		} else {
			fmt.Printf("Ends At:         No end date\n")
		}
		fmt.Printf("Created:         %s\n", promo.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:         %s\n", promo.UpdatedAt.Format(time.RFC3339))

		if promo.Banner != nil && promo.Banner.Enabled {
			fmt.Printf("\nBanner:\n")
			fmt.Printf("  Text:          %s\n", promo.Banner.Text)
			fmt.Printf("  Position:      %s\n", promo.Banner.Position)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(storefrontPromotionsCmd)

	storefrontPromotionsCmd.AddCommand(storefrontPromotionsListCmd)
	storefrontPromotionsListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired)")
	storefrontPromotionsListCmd.Flags().String("type", "", "Filter by promotion type")
	storefrontPromotionsListCmd.Flags().String("discount-type", "", "Filter by discount type (percentage, fixed, shipping)")
	storefrontPromotionsListCmd.Flags().Int("page", 1, "Page number")
	storefrontPromotionsListCmd.Flags().Int("page-size", 20, "Results per page")

	storefrontPromotionsCmd.AddCommand(storefrontPromotionsGetCmd)
	storefrontPromotionsGetCmd.Flags().Bool("by-code", false, "Get promotion by code instead of ID")
}

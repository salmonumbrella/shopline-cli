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
		fmt.Printf("\nShowing %d of %d promotions\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Promotion ID:    %s\n", promotion.ID)
		fmt.Printf("Title:           %s\n", promotion.Title)
		fmt.Printf("Description:     %s\n", promotion.Description)
		fmt.Printf("Type:            %s\n", promotion.Type)
		fmt.Printf("Status:          %s\n", promotion.Status)
		fmt.Printf("Discount Type:   %s\n", promotion.DiscountType)
		fmt.Printf("Discount Value:  %.2f\n", promotion.DiscountValue)
		fmt.Printf("Min Purchase:    %.2f\n", promotion.MinPurchase)
		fmt.Printf("Usage:           %d", promotion.UsageCount)
		if promotion.UsageLimit > 0 {
			fmt.Printf(" / %d", promotion.UsageLimit)
		}
		fmt.Println()
		fmt.Printf("Starts At:       %s\n", promotion.StartsAt.Format(time.RFC3339))
		if !promotion.EndsAt.IsZero() {
			fmt.Printf("Ends At:         %s\n", promotion.EndsAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:         %s\n", promotion.CreatedAt.Format(time.RFC3339))
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

		fmt.Printf("Activated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
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

		fmt.Printf("Deactivated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
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
			fmt.Printf("Delete promotion %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeletePromotion(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete promotion: %w", err)
		}

		fmt.Printf("Deleted promotion %s\n", args[0])
		return nil
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
}

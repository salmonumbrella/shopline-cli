package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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

var promotionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a promotion",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintln(outWriter(cmd), "[DRY-RUN] Would create promotion")
			return nil
		}

		var req api.PromotionCreateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.CreatePromotion(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create promotion: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(promotion)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created promotion %s (status: %s)\n", promotion.ID, promotion.Status)
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

var promotionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would update promotion %s\n", args[0])
			return nil
		}

		var req api.PromotionUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.UpdatePromotion(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update promotion: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(promotion)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("query")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.PromotionSearchOptions{
			Query:    query,
			Status:   status,
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.SearchPromotions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search promotions: %w", err)
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
	promotionsCmd.AddCommand(promotionsCreateCmd)
	addJSONBodyFlags(promotionsCreateCmd)

	promotionsCmd.AddCommand(promotionsUpdateCmd)
	addJSONBodyFlags(promotionsUpdateCmd)

	promotionsCmd.AddCommand(promotionsSearchCmd)
	promotionsSearchCmd.Flags().String("query", "", "Search query")
	promotionsSearchCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")
	promotionsSearchCmd.Flags().Int("page", 1, "Page number")
	promotionsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	promotionsCmd.AddCommand(promotionsActivateCmd)
	promotionsCmd.AddCommand(promotionsDeactivateCmd)
	promotionsCmd.AddCommand(promotionsDeleteCmd)
	promotionsCmd.AddCommand(promotionsCouponCenterCmd)

	schema.Register(schema.Resource{
		Name:        "promotions",
		Description: "Manage promotions",
		Commands:    []string{"list", "get", "create", "update", "search", "activate", "deactivate", "delete", "coupon-center"},
		IDPrefix:    "promotion",
	})
}

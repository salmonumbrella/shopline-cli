package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var salesCmd = &cobra.Command{
	Use:   "sales",
	Short: "Manage sale campaigns",
}

var salesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sale campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.SalesListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListSales(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list sales: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "DISCOUNT", "APPLIES TO", "STATUS", "STARTS", "ENDS"}
		var rows [][]string
		for _, s := range resp.Items {
			discount := fmt.Sprintf("%.0f", s.DiscountValue)
			if s.DiscountType == "percentage" {
				discount += "%"
			}
			startsAt := "-"
			if !s.StartsAt.IsZero() {
				startsAt = s.StartsAt.Format("2006-01-02")
			}
			endsAt := "-"
			if !s.EndsAt.IsZero() {
				endsAt = s.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				s.ID,
				s.Title,
				discount,
				s.AppliesTo,
				s.Status,
				startsAt,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d sales\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var salesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get sale details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		sale, err := client.GetSale(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get sale: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(sale)
		}

		fmt.Printf("Sale ID:         %s\n", sale.ID)
		fmt.Printf("Title:           %s\n", sale.Title)
		fmt.Printf("Description:     %s\n", sale.Description)
		fmt.Printf("Discount Type:   %s\n", sale.DiscountType)
		fmt.Printf("Discount Value:  %.2f\n", sale.DiscountValue)
		fmt.Printf("Applies To:      %s\n", sale.AppliesTo)
		if len(sale.ProductIDs) > 0 {
			fmt.Printf("Products:        %s\n", strings.Join(sale.ProductIDs, ", "))
		}
		if len(sale.CollectionIDs) > 0 {
			fmt.Printf("Collections:     %s\n", strings.Join(sale.CollectionIDs, ", "))
		}
		fmt.Printf("Status:          %s\n", sale.Status)
		if !sale.StartsAt.IsZero() {
			fmt.Printf("Starts At:       %s\n", sale.StartsAt.Format(time.RFC3339))
		}
		if !sale.EndsAt.IsZero() {
			fmt.Printf("Ends At:         %s\n", sale.EndsAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:         %s\n", sale.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var salesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a sale campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		discountType, _ := cmd.Flags().GetString("discount-type")
		discountValue, _ := cmd.Flags().GetFloat64("discount-value")
		appliesTo, _ := cmd.Flags().GetString("applies-to")
		productIDs, _ := cmd.Flags().GetStringSlice("product-ids")
		collectionIDs, _ := cmd.Flags().GetStringSlice("collection-ids")
		startsAtStr, _ := cmd.Flags().GetString("starts-at")
		endsAtStr, _ := cmd.Flags().GetString("ends-at")

		req := &api.SaleCreateRequest{
			Title:         title,
			Description:   description,
			DiscountType:  discountType,
			DiscountValue: discountValue,
			AppliesTo:     appliesTo,
			ProductIDs:    productIDs,
			CollectionIDs: collectionIDs,
		}

		if startsAtStr != "" {
			startsAt, err := time.Parse(time.RFC3339, startsAtStr)
			if err != nil {
				return fmt.Errorf("invalid starts-at format (use RFC3339): %w", err)
			}
			req.StartsAt = &startsAt
		}

		if endsAtStr != "" {
			endsAt, err := time.Parse(time.RFC3339, endsAtStr)
			if err != nil {
				return fmt.Errorf("invalid ends-at format (use RFC3339): %w", err)
			}
			req.EndsAt = &endsAt
		}

		sale, err := client.CreateSale(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create sale: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(sale)
		}

		fmt.Printf("Created sale %s (%s)\n", sale.ID, sale.Title)
		return nil
	},
}

var salesActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a sale campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		sale, err := client.ActivateSale(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate sale: %w", err)
		}

		fmt.Printf("Activated sale %s (status: %s)\n", sale.ID, sale.Status)
		return nil
	},
}

var salesDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a sale campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		sale, err := client.DeactivateSale(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate sale: %w", err)
		}

		fmt.Printf("Deactivated sale %s (status: %s)\n", sale.ID, sale.Status)
		return nil
	},
}

var salesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a sale campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete sale %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteSale(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete sale: %w", err)
		}

		fmt.Printf("Deleted sale %s\n", args[0])
		return nil
	},
}

var salesDeleteProductsCmd = &cobra.Command{
	Use:   "delete-products <id>",
	Short: "Delete products from a sale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		productIDs, _ := cmd.Flags().GetStringSlice("product-ids")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete products from sale %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		err = client.DeleteSaleProducts(cmd.Context(), args[0], &api.SaleDeleteProductsRequest{
			ProductIDs: productIDs,
		})
		if err != nil {
			return fmt.Errorf("failed to delete sale products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(map[string]any{"ok": true})
		}

		fmt.Printf("Deleted %d product(s) from sale %s\n", len(productIDs), args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(salesCmd)

	salesCmd.AddCommand(salesListCmd)
	salesListCmd.Flags().Int("page", 1, "Page number")
	salesListCmd.Flags().Int("page-size", 20, "Results per page")
	salesListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")

	salesCmd.AddCommand(salesGetCmd)

	salesCmd.AddCommand(salesCreateCmd)
	salesCreateCmd.Flags().String("title", "", "Sale title (required)")
	salesCreateCmd.Flags().String("description", "", "Sale description")
	salesCreateCmd.Flags().String("discount-type", "", "Discount type: percentage or fixed_amount (required)")
	salesCreateCmd.Flags().Float64("discount-value", 0, "Discount value (required)")
	salesCreateCmd.Flags().String("applies-to", "all", "Applies to: all, products, collections")
	salesCreateCmd.Flags().StringSlice("product-ids", nil, "Product IDs (comma-separated)")
	salesCreateCmd.Flags().StringSlice("collection-ids", nil, "Collection IDs (comma-separated)")
	salesCreateCmd.Flags().String("starts-at", "", "Start time (RFC3339 format)")
	salesCreateCmd.Flags().String("ends-at", "", "End time (RFC3339 format)")
	_ = salesCreateCmd.MarkFlagRequired("title")
	_ = salesCreateCmd.MarkFlagRequired("discount-type")
	_ = salesCreateCmd.MarkFlagRequired("discount-value")

	salesCmd.AddCommand(salesActivateCmd)
	salesCmd.AddCommand(salesDeactivateCmd)

	salesCmd.AddCommand(salesDeleteCmd)
	salesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	salesCmd.AddCommand(salesDeleteProductsCmd)
	salesDeleteProductsCmd.Flags().StringSlice("product-ids", nil, "Product IDs (comma-separated) (required)")
	_ = salesDeleteProductsCmd.MarkFlagRequired("product-ids")
}

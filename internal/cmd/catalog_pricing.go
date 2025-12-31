package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var catalogPricingCmd = &cobra.Command{
	Use:   "catalog-pricing",
	Short: "Manage B2B catalog pricing",
}

var catalogPricingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List catalog pricing entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		catalogID, _ := cmd.Flags().GetString("catalog-id")
		productID, _ := cmd.Flags().GetString("product-id")

		opts := &api.CatalogPricingListOptions{
			Page:      page,
			PageSize:  pageSize,
			CatalogID: catalogID,
			ProductID: productID,
		}

		resp, err := client.ListCatalogPricing(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list catalog pricing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CATALOG ID", "PRODUCT ID", "ORIGINAL", "CATALOG PRICE", "DISCOUNT", "QTY RANGE"}
		var rows [][]string
		for _, p := range resp.Items {
			discount := fmt.Sprintf("%.0f%%", p.DiscountPct)
			qtyRange := "-"
			if p.MinQuantity > 0 || p.MaxQuantity > 0 {
				if p.MaxQuantity > 0 {
					qtyRange = fmt.Sprintf("%d-%d", p.MinQuantity, p.MaxQuantity)
				} else {
					qtyRange = fmt.Sprintf("%d+", p.MinQuantity)
				}
			}
			rows = append(rows, []string{
				p.ID,
				p.CatalogID,
				p.ProductID,
				fmt.Sprintf("%.2f", p.OriginalPrice),
				fmt.Sprintf("%.2f", p.CatalogPrice),
				discount,
				qtyRange,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d catalog pricing entries\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var catalogPricingGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get catalog pricing details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		pricing, err := client.GetCatalogPricing(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get catalog pricing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(pricing)
		}

		fmt.Printf("Pricing ID:      %s\n", pricing.ID)
		fmt.Printf("Catalog ID:      %s\n", pricing.CatalogID)
		fmt.Printf("Product ID:      %s\n", pricing.ProductID)
		if pricing.VariantID != "" {
			fmt.Printf("Variant ID:      %s\n", pricing.VariantID)
		}
		fmt.Printf("Original Price:  %.2f\n", pricing.OriginalPrice)
		fmt.Printf("Catalog Price:   %.2f\n", pricing.CatalogPrice)
		fmt.Printf("Discount:        %.0f%%\n", pricing.DiscountPct)
		if pricing.MinQuantity > 0 {
			fmt.Printf("Min Quantity:    %d\n", pricing.MinQuantity)
		}
		if pricing.MaxQuantity > 0 {
			fmt.Printf("Max Quantity:    %d\n", pricing.MaxQuantity)
		}
		fmt.Printf("Created:         %s\n", pricing.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:         %s\n", pricing.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var catalogPricingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create catalog pricing",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		catalogID, _ := cmd.Flags().GetString("catalog-id")
		productID, _ := cmd.Flags().GetString("product-id")
		variantID, _ := cmd.Flags().GetString("variant-id")
		catalogPrice, _ := cmd.Flags().GetFloat64("catalog-price")
		minQuantity, _ := cmd.Flags().GetInt("min-quantity")
		maxQuantity, _ := cmd.Flags().GetInt("max-quantity")

		req := &api.CatalogPricingCreateRequest{
			CatalogID:    catalogID,
			ProductID:    productID,
			VariantID:    variantID,
			CatalogPrice: catalogPrice,
			MinQuantity:  minQuantity,
			MaxQuantity:  maxQuantity,
		}

		pricing, err := client.CreateCatalogPricing(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create catalog pricing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(pricing)
		}

		fmt.Printf("Created catalog pricing %s (price: %.2f)\n", pricing.ID, pricing.CatalogPrice)
		return nil
	},
}

var catalogPricingUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update catalog pricing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.CatalogPricingUpdateRequest{}

		if cmd.Flags().Changed("catalog-price") {
			catalogPrice, _ := cmd.Flags().GetFloat64("catalog-price")
			req.CatalogPrice = &catalogPrice
		}
		if cmd.Flags().Changed("min-quantity") {
			minQuantity, _ := cmd.Flags().GetInt("min-quantity")
			req.MinQuantity = &minQuantity
		}
		if cmd.Flags().Changed("max-quantity") {
			maxQuantity, _ := cmd.Flags().GetInt("max-quantity")
			req.MaxQuantity = &maxQuantity
		}

		pricing, err := client.UpdateCatalogPricing(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update catalog pricing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(pricing)
		}

		fmt.Printf("Updated catalog pricing %s\n", pricing.ID)
		return nil
	},
}

var catalogPricingDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete catalog pricing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete catalog pricing %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCatalogPricing(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete catalog pricing: %w", err)
		}

		fmt.Printf("Deleted catalog pricing %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(catalogPricingCmd)

	catalogPricingCmd.AddCommand(catalogPricingListCmd)
	catalogPricingListCmd.Flags().Int("page", 1, "Page number")
	catalogPricingListCmd.Flags().Int("page-size", 20, "Results per page")
	catalogPricingListCmd.Flags().String("catalog-id", "", "Filter by catalog ID")
	catalogPricingListCmd.Flags().String("product-id", "", "Filter by product ID")

	catalogPricingCmd.AddCommand(catalogPricingGetCmd)

	catalogPricingCmd.AddCommand(catalogPricingCreateCmd)
	catalogPricingCreateCmd.Flags().String("catalog-id", "", "Catalog ID (required)")
	catalogPricingCreateCmd.Flags().String("product-id", "", "Product ID (required)")
	catalogPricingCreateCmd.Flags().String("variant-id", "", "Variant ID")
	catalogPricingCreateCmd.Flags().Float64("catalog-price", 0, "Catalog price (required)")
	catalogPricingCreateCmd.Flags().Int("min-quantity", 0, "Minimum quantity")
	catalogPricingCreateCmd.Flags().Int("max-quantity", 0, "Maximum quantity")
	_ = catalogPricingCreateCmd.MarkFlagRequired("catalog-id")
	_ = catalogPricingCreateCmd.MarkFlagRequired("product-id")
	_ = catalogPricingCreateCmd.MarkFlagRequired("catalog-price")

	catalogPricingCmd.AddCommand(catalogPricingUpdateCmd)
	catalogPricingUpdateCmd.Flags().Float64("catalog-price", 0, "Catalog price")
	catalogPricingUpdateCmd.Flags().Int("min-quantity", 0, "Minimum quantity")
	catalogPricingUpdateCmd.Flags().Int("max-quantity", 0, "Maximum quantity")

	catalogPricingCmd.AddCommand(catalogPricingDeleteCmd)
	catalogPricingDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

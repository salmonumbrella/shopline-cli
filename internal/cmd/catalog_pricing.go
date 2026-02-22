package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
				outfmt.FormatID("catalog_price", p.ID),
				p.CatalogID,
				p.ProductID,
				fmt.Sprintf("%.2f", p.OriginalPrice),
				fmt.Sprintf("%.2f", p.CatalogPrice),
				discount,
				qtyRange,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d catalog pricing entries\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Pricing ID:      %s\n", pricing.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Catalog ID:      %s\n", pricing.CatalogID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:      %s\n", pricing.ProductID)
		if pricing.VariantID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Variant ID:      %s\n", pricing.VariantID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Original Price:  %.2f\n", pricing.OriginalPrice)
		_, _ = fmt.Fprintf(outWriter(cmd), "Catalog Price:   %.2f\n", pricing.CatalogPrice)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount:        %.0f%%\n", pricing.DiscountPct)
		if pricing.MinQuantity > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Min Quantity:    %d\n", pricing.MinQuantity)
		}
		if pricing.MaxQuantity > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Max Quantity:    %d\n", pricing.MaxQuantity)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", pricing.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", pricing.UpdatedAt.Format(time.RFC3339))
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
		if checkDryRun(cmd, "[DRY-RUN] Would create catalog pricing") {
			return nil
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created catalog pricing %s (price: %.2f)\n", pricing.ID, pricing.CatalogPrice)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update catalog pricing %s", args[0])) {
			return nil
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated catalog pricing %s\n", pricing.ID)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete catalog pricing %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete catalog pricing %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCatalogPricing(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete catalog pricing: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted catalog pricing %s\n", args[0])
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

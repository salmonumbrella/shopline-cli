package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var productListingsCmd = &cobra.Command{
	Use:   "product-listings",
	Short: "Manage product listings (products published to sales channels)",
}

var productListingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List product listings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ProductListingsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListProductListings(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list product listings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PRODUCT ID", "TITLE", "VENDOR", "TYPE", "AVAILABLE", "PUBLISHED"}
		var rows [][]string
		for _, pl := range resp.Items {
			available := "No"
			if pl.Available {
				available = "Yes"
			}
			publishedAt := ""
			if !pl.PublishedAt.IsZero() {
				publishedAt = pl.PublishedAt.Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{
				outfmt.FormatID("product_listing", pl.ID),
				pl.ProductID,
				pl.Title,
				pl.Vendor,
				pl.ProductType,
				available,
				publishedAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d product listings\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var productListingsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get product listing details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		listing, err := client.GetProductListing(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product listing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(listing)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Listing ID:    %s\n", listing.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:    %s\n", listing.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:         %s\n", listing.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:        %s\n", listing.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Vendor:        %s\n", listing.Vendor)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product Type:  %s\n", listing.ProductType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:     %t\n", listing.Available)
		if listing.BodyHTML != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Body HTML:     %s\n", listing.BodyHTML)
		}
		if !listing.PublishedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Published:     %s\n", listing.PublishedAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", listing.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", listing.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var productListingsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Publish a product to a sales channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product-id")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create product listing for product %s", productID)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		listing, err := client.CreateProductListing(cmd.Context(), productID)
		if err != nil {
			return fmt.Errorf("failed to create product listing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(listing)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created product listing %s\n", listing.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:  %s\n", listing.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:       %s\n", listing.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:   %t\n", listing.Available)

		return nil
	},
}

var productListingsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a product listing from a sales channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete product listing %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete product listing %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteProductListing(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete product listing: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted product listing %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(productListingsCmd)

	productListingsCmd.AddCommand(productListingsListCmd)
	productListingsListCmd.Flags().Int("page", 1, "Page number")
	productListingsListCmd.Flags().Int("page-size", 20, "Results per page")

	productListingsCmd.AddCommand(productListingsGetCmd)

	productListingsCmd.AddCommand(productListingsCreateCmd)
	productListingsCreateCmd.Flags().String("product-id", "", "Product ID to publish")
	_ = productListingsCreateCmd.MarkFlagRequired("product-id")

	productListingsCmd.AddCommand(productListingsDeleteCmd)
	productListingsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var storefrontProductsCmd = &cobra.Command{
	Use:   "storefront-products",
	Short: "View storefront product information",
}

var storefrontProductsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storefront products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		collection, _ := cmd.Flags().GetString("collection")
		category, _ := cmd.Flags().GetString("category")
		vendor, _ := cmd.Flags().GetString("vendor")
		productType, _ := cmd.Flags().GetString("product-type")
		tag, _ := cmd.Flags().GetString("tag")
		query, _ := cmd.Flags().GetString("query")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StorefrontProductsListOptions{
			Page:        page,
			PageSize:    pageSize,
			Collection:  collection,
			Category:    category,
			Vendor:      vendor,
			ProductType: productType,
			Tag:         tag,
			Query:       query,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListStorefrontProducts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list storefront products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "VENDOR", "PRICE", "AVAILABLE", "VIEWS", "SALES"}
		var rows [][]string
		for _, p := range resp.Items {
			price := p.Price
			if p.Currency != "" {
				price = p.Price + " " + p.Currency
			}
			available := "No"
			if p.Available {
				available = "Yes"
			}
			rows = append(rows, []string{
				p.ID,
				p.Title,
				p.Vendor,
				price,
				available,
				fmt.Sprintf("%d", p.ViewCount),
				fmt.Sprintf("%d", p.SalesCount),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d products\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storefrontProductsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get storefront product details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		byHandle, _ := cmd.Flags().GetBool("by-handle")

		var product *api.StorefrontProduct
		if byHandle {
			product, err = client.GetStorefrontProductByHandle(cmd.Context(), args[0])
		} else {
			product, err = client.GetStorefrontProduct(cmd.Context(), args[0])
		}
		if err != nil {
			return fmt.Errorf("failed to get storefront product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		fmt.Printf("Product ID:     %s\n", product.ID)
		fmt.Printf("Handle:         %s\n", product.Handle)
		fmt.Printf("Title:          %s\n", product.Title)
		fmt.Printf("Description:    %s\n", product.Description)
		fmt.Printf("Vendor:         %s\n", product.Vendor)
		fmt.Printf("Product Type:   %s\n", product.ProductType)
		fmt.Printf("Tags:           %s\n", strings.Join(product.Tags, ", "))
		fmt.Printf("Status:         %s\n", product.Status)
		fmt.Printf("Available:      %v\n", product.Available)
		fmt.Printf("Price:          %s %s\n", product.Price, product.Currency)
		if product.CompareAtPrice != "" {
			fmt.Printf("Compare At:     %s %s\n", product.CompareAtPrice, product.Currency)
		}
		fmt.Printf("View Count:     %d\n", product.ViewCount)
		fmt.Printf("Sales Count:    %d\n", product.SalesCount)
		fmt.Printf("Review Count:   %d\n", product.ReviewCount)
		fmt.Printf("Average Rating: %.1f\n", product.AverageRating)
		if product.PublishedAt != nil {
			fmt.Printf("Published:      %s\n", product.PublishedAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", product.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", product.UpdatedAt.Format(time.RFC3339))

		if len(product.Variants) > 0 {
			fmt.Printf("\nVariants (%d):\n", len(product.Variants))
			for _, v := range product.Variants {
				available := "unavailable"
				if v.Available {
					available = "available"
				}
				fmt.Printf("  - %s (SKU: %s) %s - %s\n", v.Title, v.SKU, v.Price, available)
			}
		}

		if len(product.Images) > 0 {
			fmt.Printf("\nImages (%d):\n", len(product.Images))
			for _, img := range product.Images {
				fmt.Printf("  - %s (%dx%d)\n", img.URL, img.Width, img.Height)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(storefrontProductsCmd)

	storefrontProductsCmd.AddCommand(storefrontProductsListCmd)
	storefrontProductsListCmd.Flags().String("collection", "", "Filter by collection")
	storefrontProductsListCmd.Flags().String("category", "", "Filter by category")
	storefrontProductsListCmd.Flags().String("vendor", "", "Filter by vendor")
	storefrontProductsListCmd.Flags().String("product-type", "", "Filter by product type")
	storefrontProductsListCmd.Flags().String("tag", "", "Filter by tag")
	storefrontProductsListCmd.Flags().String("query", "", "Search query")
	storefrontProductsListCmd.Flags().Int("page", 1, "Page number")
	storefrontProductsListCmd.Flags().Int("page-size", 20, "Results per page")

	storefrontProductsCmd.AddCommand(storefrontProductsGetCmd)
	storefrontProductsGetCmd.Flags().Bool("by-handle", false, "Get product by handle instead of ID")
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage products",
}

var productsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		vendor, _ := cmd.Flags().GetString("vendor")
		productType, _ := cmd.Flags().GetString("product-type")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ProductsListOptions{
			Page:        page,
			PageSize:    pageSize,
			Status:      status,
			Vendor:      vendor,
			ProductType: productType,
		}

		resp, err := client.ListProducts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "STATUS", "VENDOR", "TYPE", "PRICE", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			priceStr := ""
			if p.Price != nil {
				priceStr = p.Price.Label
			}
			rows = append(rows, []string{
				p.ID,
				p.Title,
				p.Status,
				p.Vendor,
				p.ProductType,
				priceStr,
				p.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d products\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var productsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get product details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.GetProduct(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		fmt.Printf("Product ID:     %s\n", product.ID)
		fmt.Printf("Title:          %s\n", product.Title)
		fmt.Printf("Handle:         %s\n", product.Handle)
		fmt.Printf("Status:         %s\n", product.Status)
		fmt.Printf("Description:    %s\n", product.Description)
		fmt.Printf("Vendor:         %s\n", product.Vendor)
		fmt.Printf("Product Type:   %s\n", product.ProductType)
		fmt.Printf("Tags:           %s\n", strings.Join(product.Tags, ", "))
		priceStr := ""
		if product.Price != nil {
			priceStr = product.Price.Label
		}
		fmt.Printf("Price:          %s\n", priceStr)
		fmt.Printf("Created:        %s\n", product.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", product.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(productsCmd)

	productsCmd.AddCommand(productsListCmd)
	productsListCmd.Flags().String("status", "", "Filter by status (active, draft, archived)")
	productsListCmd.Flags().String("vendor", "", "Filter by vendor")
	productsListCmd.Flags().String("product-type", "", "Filter by product type")
	productsListCmd.Flags().Int("page", 1, "Page number")
	productsListCmd.Flags().Int("page-size", 20, "Results per page")

	productsCmd.AddCommand(productsGetCmd)
}

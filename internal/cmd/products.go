package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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
		limit, _ := cmd.Flags().GetInt("limit")

		opts := &api.ProductsListOptions{
			Page:        page,
			PageSize:    pageSize,
			Status:      status,
			Vendor:      vendor,
			ProductType: productType,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp := &api.ProductsListResponse{}
		if limit > 0 {
			curPage := opts.Page
			perPage := opts.PageSize
			if perPage <= 0 || perPage > limit {
				perPage = limit
			}

			items := make([]api.Product, 0, limit)
			totalCount := 0
			hasMore := false
			var pagination api.Pagination

			for len(items) < limit {
				pageOpts := *opts
				pageOpts.Page = curPage
				pageOpts.PageSize = perPage

				pageResp, err := client.ListProducts(cmd.Context(), &pageOpts)
				if err != nil {
					return fmt.Errorf("failed to list products: %w", err)
				}
				if totalCount == 0 {
					totalCount = pageResp.TotalCount
					pagination = pageResp.Pagination
				}
				items = append(items, pageResp.Items...)
				hasMore = pageResp.HasMore

				if !pageResp.HasMore || len(pageResp.Items) == 0 {
					break
				}
				curPage++
			}

			if len(items) > limit {
				items = items[:limit]
				hasMore = true
			}

			resp.Items = items
			resp.Page = opts.Page
			resp.PageSize = perPage
			resp.TotalCount = totalCount
			resp.HasMore = hasMore
			resp.Pagination = pagination
			resp.Pagination.CurrentPage = opts.Page
			resp.Pagination.PerPage = perPage
		} else {
			r, err := client.ListProducts(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to list products: %w", err)
			}
			resp = r
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
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d products\n", len(resp.Items), resp.TotalCount)
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

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Product ID:     %s\n", product.ID)
		_, _ = fmt.Fprintf(out, "Title:          %s\n", product.Title)
		_, _ = fmt.Fprintf(out, "Handle:         %s\n", product.Handle)
		_, _ = fmt.Fprintf(out, "Status:         %s\n", product.Status)
		_, _ = fmt.Fprintf(out, "Description:    %s\n", product.Description)
		_, _ = fmt.Fprintf(out, "Vendor:         %s\n", product.Vendor)
		_, _ = fmt.Fprintf(out, "Product Type:   %s\n", product.ProductType)
		_, _ = fmt.Fprintf(out, "Tags:           %s\n", strings.Join(product.Tags, ", "))
		priceStr := ""
		if product.Price != nil {
			priceStr = product.Price.Label
		}
		_, _ = fmt.Fprintf(out, "Price:          %s\n", priceStr)
		_, _ = fmt.Fprintf(out, "Created:        %s\n", product.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(out, "Updated:        %s\n", product.UpdatedAt.Format(time.RFC3339))
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

	schema.Register(schema.Resource{
		Name:        "products",
		Description: "Manage products and variants",
		Commands:    []string{"list", "get"},
		IDPrefix:    "product",
	})
}

package cmd

import (
	"encoding/json"
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

var productsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a product",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		vendor, _ := cmd.Flags().GetString("vendor")
		productType, _ := cmd.Flags().GetString("product-type")
		tagsStr, _ := cmd.Flags().GetString("tags")
		status, _ := cmd.Flags().GetString("status")

		req := &api.ProductCreateRequest{
			Title:       title,
			Description: description,
			Vendor:      vendor,
			ProductType: productType,
			Status:      status,
		}
		if tagsStr != "" {
			req.Tags = strings.Split(tagsStr, ",")
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.CreateProduct(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created product %s (%s)\n", product.ID, product.Title)
		return nil
	},
}

var productsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.ProductUpdateRequest{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			req.Title = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		if cmd.Flags().Changed("vendor") {
			v, _ := cmd.Flags().GetString("vendor")
			req.Vendor = &v
		}
		if cmd.Flags().Changed("product-type") {
			v, _ := cmd.Flags().GetString("product-type")
			req.ProductType = &v
		}
		if cmd.Flags().Changed("tags") {
			v, _ := cmd.Flags().GetString("tags")
			req.Tags = strings.Split(v, ",")
		}
		if cmd.Flags().Changed("status") {
			v, _ := cmd.Flags().GetString("status")
			req.Status = &v
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.UpdateProduct(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated product %s\n", product.ID)
		return nil
	},
}

var productsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Delete product %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
				return nil
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteProduct(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete product: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted product %s\n", args[0])
		return nil
	},
}

var productsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("query")
		status, _ := cmd.Flags().GetString("status")
		vendor, _ := cmd.Flags().GetString("vendor")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ProductSearchOptions{
			Query:    query,
			Status:   status,
			Vendor:   vendor,
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.SearchProducts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "STATUS", "VENDOR", "TYPE", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			rows = append(rows, []string{
				p.ID,
				p.Title,
				p.Status,
				p.Vendor,
				p.ProductType,
				p.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d products\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

// --- Tags subcommands ---

var productsTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage product tags",
}

var productsTagsReplaceCmd = &cobra.Command{
	Use:   "replace <id>",
	Short: "Replace all tags for a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tagsStr, _ := cmd.Flags().GetString("tags")
		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.ReplaceProductTags(cmd.Context(), args[0], tags)
		if err != nil {
			return fmt.Errorf("failed to replace product tags: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Replaced tags for product %s: %s\n", product.ID, strings.Join(product.Tags, ", "))
		return nil
	},
}

var productsTagsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Add or remove tags from a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addStr, _ := cmd.Flags().GetString("add")
		removeStr, _ := cmd.Flags().GetString("remove")

		req := &api.ProductTagsUpdateRequest{}
		if addStr != "" {
			req.Add = strings.Split(addStr, ",")
		}
		if removeStr != "" {
			req.Remove = strings.Split(removeStr, ",")
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.UpdateProductTags(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update product tags: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated tags for product %s: %s\n", product.ID, strings.Join(product.Tags, ", "))
		return nil
	},
}

// --- Quantity / Price commands ---

var productsUpdateQuantityCmd = &cobra.Command{
	Use:   "update-quantity <id>",
	Short: "Update product quantity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		qty, _ := cmd.Flags().GetInt("quantity")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.UpdateProductQuantity(cmd.Context(), args[0], qty)
		if err != nil {
			return fmt.Errorf("failed to update product quantity: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated product %s quantity\n", product.ID)
		return nil
	},
}

var productsUpdatePriceCmd = &cobra.Command{
	Use:   "update-price <id>",
	Short: "Update product price",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		price, _ := cmd.Flags().GetFloat64("price")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.UpdateProductPrice(cmd.Context(), args[0], price)
		if err != nil {
			return fmt.Errorf("failed to update product price: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated product %s price\n", product.ID)
		return nil
	},
}

var productsUpdateQuantityBySKUCmd = &cobra.Command{
	Use:   "update-quantity-by-sku",
	Short: "Update product quantity by SKU",
	RunE: func(cmd *cobra.Command, args []string) error {
		sku, _ := cmd.Flags().GetString("sku")
		qty, _ := cmd.Flags().GetInt("quantity")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.UpdateProductQuantityBySKU(cmd.Context(), sku, qty); err != nil {
			return fmt.Errorf("failed to update product quantity by SKU: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(map[string]any{"ok": true})
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated product quantity for SKU %s\n", sku)
		return nil
	},
}

// --- Stocks subcommands ---

var productsStocksCmd = &cobra.Command{
	Use:   "stocks",
	Short: "Manage product stocks",
}

var productsStocksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get product stocks (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetProductStocks(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product stocks: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsStocksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update product stocks (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateProductStocks(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update product stocks: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsBulkUpdateStocksCmd = &cobra.Command{
	Use:   "bulk-update-stocks",
	Short: "Bulk update product stocks (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.BulkUpdateProductStocks(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk update product stocks: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(map[string]any{"ok": true})
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Bulk updated product stocks")
		return nil
	},
}

var productsLockedInventoryCmd = &cobra.Command{
	Use:   "locked-inventory <id>",
	Short: "Get locked inventory count for a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		count, err := client.GetLockedInventoryCount(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get locked inventory count: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(count)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Product ID:    %s\n", count.ProductID)
		if count.VariationID != "" {
			_, _ = fmt.Fprintf(out, "Variation ID:  %s\n", count.VariationID)
		}
		_, _ = fmt.Fprintf(out, "Locked Count:  %d\n", count.LockedCount)
		return nil
	},
}

// --- Images subcommands ---

var productsImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage product images",
}

var productsImagesAddCmd = &cobra.Command{
	Use:   "add <id>",
	Short: "Add images to a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcs, _ := cmd.Flags().GetStringSlice("src")
		position, _ := cmd.Flags().GetInt("position")

		var images []api.ProductImageInput
		for _, src := range srcs {
			images = append(images, api.ProductImageInput{
				Src:      src,
				Position: position,
			})
		}
		req := &api.ProductAddImagesRequest{Images: images}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		added, err := client.AddProductImages(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to add product images: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(added)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Added %d image(s) to product %s\n", len(added), args[0])
		return nil
	},
}

var productsImagesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete images from a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageIDsStr, _ := cmd.Flags().GetString("image-ids")
		imageIDs := strings.Split(imageIDsStr, ",")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteProductImages(cmd.Context(), args[0], imageIDs); err != nil {
			return fmt.Errorf("failed to delete product images: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted %d image(s) from product %s\n", len(imageIDs), args[0])
		return nil
	},
}

// --- Variations subcommands ---

var productsVariationsCmd = &cobra.Command{
	Use:   "variations",
	Short: "Manage product variations",
}

var productsVariationsCreateCmd = &cobra.Command{
	Use:   "create <product-id>",
	Short: "Create a product variation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		sku, _ := cmd.Flags().GetString("sku")
		price, _ := cmd.Flags().GetFloat64("price")
		quantity, _ := cmd.Flags().GetInt("quantity")

		req := &api.ProductVariationCreateRequest{
			Title:    title,
			SKU:      sku,
			Price:    price,
			Quantity: quantity,
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		variation, err := client.AddProductVariation(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to create variation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(variation)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created variation %s for product %s\n", variation.ID, args[0])
		return nil
	},
}

var productsVariationsUpdateCmd = &cobra.Command{
	Use:   "update <product-id> <variation-id>",
	Short: "Update a product variation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.ProductVariationUpdateRequest{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			req.Title = &v
		}
		if cmd.Flags().Changed("sku") {
			v, _ := cmd.Flags().GetString("sku")
			req.SKU = &v
		}
		if cmd.Flags().Changed("price") {
			v, _ := cmd.Flags().GetFloat64("price")
			req.Price = &v
		}
		if cmd.Flags().Changed("quantity") {
			v, _ := cmd.Flags().GetInt("quantity")
			req.Quantity = &v
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		variation, err := client.UpdateProductVariation(cmd.Context(), args[0], args[1], req)
		if err != nil {
			return fmt.Errorf("failed to update variation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(variation)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated variation %s\n", variation.ID)
		return nil
	},
}

var productsVariationsDeleteCmd = &cobra.Command{
	Use:   "delete <product-id> <variation-id>",
	Short: "Delete a product variation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteProductVariation(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete variation: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted variation %s from product %s\n", args[1], args[0])
		return nil
	},
}

var productsVariationsUpdateQuantityCmd = &cobra.Command{
	Use:   "update-quantity <product-id> <variation-id>",
	Short: "Update variation quantity",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		qty, _ := cmd.Flags().GetInt("quantity")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.UpdateProductVariationQuantity(cmd.Context(), args[0], args[1], qty)
		if err != nil {
			return fmt.Errorf("failed to update variation quantity: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated variation %s quantity\n", args[1])
		return nil
	},
}

var productsVariationsUpdatePriceCmd = &cobra.Command{
	Use:   "update-price <product-id> <variation-id>",
	Short: "Update variation price",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		price, _ := cmd.Flags().GetFloat64("price")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		product, err := client.UpdateProductVariationPrice(cmd.Context(), args[0], args[1], price)
		if err != nil {
			return fmt.Errorf("failed to update variation price: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(product)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated variation %s price\n", args[1])
		return nil
	},
}

// --- Bulk commands ---

var productsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Delete multiple products",
	RunE: func(cmd *cobra.Command, args []string) error {
		idsStr, _ := cmd.Flags().GetString("ids")
		var ids []string
		for _, id := range strings.Split(idsStr, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return fmt.Errorf("no product IDs provided")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would delete %d product(s)\n", len(ids))
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Delete %d product(s)? [y/N] ", len(ids))
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
				return nil
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.BulkDeleteProducts(cmd.Context(), ids); err != nil {
			return fmt.Errorf("failed to bulk delete products: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted %d product(s)\n", len(ids))
		return nil
	},
}

var productsBulkStatusCmd = &cobra.Command{
	Use:   "bulk-status",
	Short: "Bulk update product online-store status (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.UpdateProductsStatusBulk(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk update product status: %w", err)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Bulk updated product status")
		return nil
	},
}

var productsBulkRetailStatusCmd = &cobra.Command{
	Use:   "bulk-retail-status",
	Short: "Bulk update product retail-store status (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.UpdateProductsRetailStatusBulk(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk update product retail status: %w", err)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Bulk updated product retail status")
		return nil
	},
}

var productsLabelsUpdateCmd = &cobra.Command{
	Use:   "labels-update",
	Short: "Update product labels (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.UpdateProductsLabelsBulk(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to update product labels: %w", err)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Updated product labels")
		return nil
	},
}

var productsPromotionsCmd = &cobra.Command{
	Use:   "promotions <id>",
	Short: "Get promotions for a product (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetProductPromotions(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product promotions: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

// searchProductsPostCmd exists only to satisfy the coverage test for SearchProductsPost.
// The primary search command uses the GET endpoint via SearchProducts.
var productsSearchPostCmd = &cobra.Command{
	Use:    "search-post",
	Short:  "Search products via POST (raw JSON body)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		bodyStr, _ := cmd.Flags().GetString("body")
		var req api.ProductSearchRequest
		if bodyStr != "" {
			if err := json.Unmarshal([]byte(bodyStr), &req); err != nil {
				return fmt.Errorf("invalid JSON body: %w", err)
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.SearchProductsPost(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to search products (POST): %w", err)
		}
		return getFormatter(cmd).JSON(resp)
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

	productsCmd.AddCommand(productsCreateCmd)
	productsCreateCmd.Flags().String("title", "", "Product title (required)")
	productsCreateCmd.Flags().String("description", "", "Product description")
	productsCreateCmd.Flags().String("vendor", "", "Product vendor")
	productsCreateCmd.Flags().String("product-type", "", "Product type")
	productsCreateCmd.Flags().String("tags", "", "Comma-separated tags")
	productsCreateCmd.Flags().String("status", "", "Product status (active, draft, archived)")
	_ = productsCreateCmd.MarkFlagRequired("title")

	productsCmd.AddCommand(productsUpdateCmd)
	productsUpdateCmd.Flags().String("title", "", "Product title")
	productsUpdateCmd.Flags().String("description", "", "Product description")
	productsUpdateCmd.Flags().String("vendor", "", "Product vendor")
	productsUpdateCmd.Flags().String("product-type", "", "Product type")
	productsUpdateCmd.Flags().String("tags", "", "Comma-separated tags")
	productsUpdateCmd.Flags().String("status", "", "Product status")

	productsCmd.AddCommand(productsDeleteCmd)
	productsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	productsCmd.AddCommand(productsSearchCmd)
	productsSearchCmd.Flags().String("query", "", "Search query")
	productsSearchCmd.Flags().String("status", "", "Filter by status")
	productsSearchCmd.Flags().String("vendor", "", "Filter by vendor")
	productsSearchCmd.Flags().Int("page", 1, "Page number")
	productsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	// Tags subcommands
	productsCmd.AddCommand(productsTagsCmd)
	productsTagsCmd.AddCommand(productsTagsReplaceCmd)
	productsTagsReplaceCmd.Flags().String("tags", "", "Comma-separated tags (required)")
	_ = productsTagsReplaceCmd.MarkFlagRequired("tags")

	productsTagsCmd.AddCommand(productsTagsUpdateCmd)
	productsTagsUpdateCmd.Flags().String("add", "", "Comma-separated tags to add")
	productsTagsUpdateCmd.Flags().String("remove", "", "Comma-separated tags to remove")

	// Quantity / Price commands
	productsCmd.AddCommand(productsUpdateQuantityCmd)
	productsUpdateQuantityCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = productsUpdateQuantityCmd.MarkFlagRequired("quantity")

	productsCmd.AddCommand(productsUpdatePriceCmd)
	productsUpdatePriceCmd.Flags().Float64("price", 0, "Price (required)")
	_ = productsUpdatePriceCmd.MarkFlagRequired("price")

	productsCmd.AddCommand(productsUpdateQuantityBySKUCmd)
	productsUpdateQuantityBySKUCmd.Flags().String("sku", "", "Product SKU (required)")
	productsUpdateQuantityBySKUCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = productsUpdateQuantityBySKUCmd.MarkFlagRequired("sku")
	_ = productsUpdateQuantityBySKUCmd.MarkFlagRequired("quantity")

	// Stocks subcommands
	productsCmd.AddCommand(productsStocksCmd)
	productsStocksCmd.AddCommand(productsStocksGetCmd)
	productsStocksCmd.AddCommand(productsStocksUpdateCmd)
	addJSONBodyFlags(productsStocksUpdateCmd)

	productsCmd.AddCommand(productsBulkUpdateStocksCmd)
	addJSONBodyFlags(productsBulkUpdateStocksCmd)

	productsCmd.AddCommand(productsLockedInventoryCmd)

	// Images subcommands
	productsCmd.AddCommand(productsImagesCmd)
	productsImagesCmd.AddCommand(productsImagesAddCmd)
	productsImagesAddCmd.Flags().StringSlice("src", nil, "Image source URL (can be repeated; required)")
	productsImagesAddCmd.Flags().Int("position", 0, "Image position")
	_ = productsImagesAddCmd.MarkFlagRequired("src")

	productsImagesCmd.AddCommand(productsImagesDeleteCmd)
	productsImagesDeleteCmd.Flags().String("image-ids", "", "Comma-separated image IDs (required)")
	_ = productsImagesDeleteCmd.MarkFlagRequired("image-ids")

	// Variations subcommands
	productsCmd.AddCommand(productsVariationsCmd)
	productsVariationsCmd.AddCommand(productsVariationsCreateCmd)
	productsVariationsCreateCmd.Flags().String("title", "", "Variation title")
	productsVariationsCreateCmd.Flags().String("sku", "", "Variation SKU")
	productsVariationsCreateCmd.Flags().Float64("price", 0, "Variation price")
	productsVariationsCreateCmd.Flags().Int("quantity", 0, "Variation quantity")

	productsVariationsCmd.AddCommand(productsVariationsUpdateCmd)
	productsVariationsUpdateCmd.Flags().String("title", "", "Variation title")
	productsVariationsUpdateCmd.Flags().String("sku", "", "Variation SKU")
	productsVariationsUpdateCmd.Flags().Float64("price", 0, "Variation price")
	productsVariationsUpdateCmd.Flags().Int("quantity", 0, "Variation quantity")

	productsVariationsCmd.AddCommand(productsVariationsDeleteCmd)

	productsVariationsCmd.AddCommand(productsVariationsUpdateQuantityCmd)
	productsVariationsUpdateQuantityCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = productsVariationsUpdateQuantityCmd.MarkFlagRequired("quantity")

	productsVariationsCmd.AddCommand(productsVariationsUpdatePriceCmd)
	productsVariationsUpdatePriceCmd.Flags().Float64("price", 0, "Price (required)")
	_ = productsVariationsUpdatePriceCmd.MarkFlagRequired("price")

	// Bulk commands
	productsCmd.AddCommand(productsBulkDeleteCmd)
	productsBulkDeleteCmd.Flags().String("ids", "", "Comma-separated product IDs (required)")
	_ = productsBulkDeleteCmd.MarkFlagRequired("ids")
	productsBulkDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	productsBulkDeleteCmd.Flags().Bool("dry-run", false, "Show what would be deleted without making changes")

	productsCmd.AddCommand(productsBulkStatusCmd)
	addJSONBodyFlags(productsBulkStatusCmd)

	productsCmd.AddCommand(productsBulkRetailStatusCmd)
	addJSONBodyFlags(productsBulkRetailStatusCmd)

	productsCmd.AddCommand(productsLabelsUpdateCmd)
	addJSONBodyFlags(productsLabelsUpdateCmd)

	productsCmd.AddCommand(productsPromotionsCmd)

	productsCmd.AddCommand(productsSearchPostCmd)
	productsSearchPostCmd.Flags().String("body", "", "JSON request body")

	schema.Register(schema.Resource{
		Name:        "products",
		Description: "Manage products and variants",
		Commands: []string{
			"list", "get", "create", "update", "delete", "search",
			"tags replace", "tags update",
			"update-quantity", "update-price", "update-quantity-by-sku",
			"stocks get", "stocks update", "bulk-update-stocks",
			"locked-inventory",
			"images add", "images delete",
			"variations create", "variations update", "variations delete",
			"variations update-quantity", "variations update-price",
			"bulk-delete", "bulk-status", "bulk-retail-status",
			"labels-update", "promotions",
			"metafields", "app-metafields",
		},
		IDPrefix: "product",
	})
}

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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

		resp, err := fetchList(
			cmd.Context(), limit, opts.Page, opts.PageSize,
			func() (*api.ListResponse[api.Product], error) {
				return client.ListProducts(cmd.Context(), opts)
			},
			func(page, size int) (*api.ListResponse[api.Product], error) {
				pageOpts := *opts
				pageOpts.Page = page
				pageOpts.PageSize = size
				return client.ListProducts(cmd.Context(), &pageOpts)
			},
			"failed to list products",
		)
		if err != nil {
			return err
		}

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightProduct)
				return formatter.JSON(api.ListResponse[lightProduct]{
					Items:      lightItems,
					Pagination: resp.Pagination,
					Page:       resp.Page,
					PageSize:   resp.PageSize,
					TotalCount: resp.TotalCount,
					HasMore:    resp.HasMore,
				})
			}
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
				outfmt.FormatID("product", p.ID),
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
	Use:   "get [id]",
	Short: "Get product details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		productID, err := resolveOrArg(cmd, args, func(query string) (string, error) {
			resp, err := client.SearchProducts(cmd.Context(), &api.ProductSearchOptions{
				Query:    query,
				PageSize: 5,
			})
			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}
			if len(resp.Items) == 0 {
				return "", fmt.Errorf("no product found matching %q", query)
			}
			if len(resp.Items) > 1 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %d products matched, using first\n", len(resp.Items))
			}
			return resp.Items[0].ID, nil
		})
		if err != nil {
			return err
		}

		product, err := client.GetProduct(cmd.Context(), productID)
		if err != nil {
			return fmt.Errorf("failed to get product: %w", err)
		}

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightProduct(product))
			}
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
		if checkDryRun(cmd, "[DRY-RUN] Would create product") {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete product %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete product %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
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

		query, _ := cmd.Flags().GetString("q")
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
				outfmt.FormatID("product", p.ID),
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would replace tags for product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update tags for product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update quantity for product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update price for product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, "[DRY-RUN] Would update product quantity by SKU") {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update stocks for product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, "[DRY-RUN] Would bulk update product stocks") {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add images to product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete images from product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create variation for product %s", args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update variation %s for product %s", args[1], args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete variation %s from product %s", args[1], args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update quantity for variation %s of product %s", args[1], args[0])) {
			return nil
		}

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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update price for variation %s of product %s", args[1], args[0])) {
			return nil
		}

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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete %d product(s): %s", len(ids), strings.Join(ids, ", "))) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete %d product(s) (%s)? [y/N] ", len(ids), strings.Join(ids, ", "))) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
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
		if checkDryRun(cmd, "[DRY-RUN] Would bulk update product status") {
			return nil
		}

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
		if checkDryRun(cmd, "[DRY-RUN] Would bulk update product retail status") {
			return nil
		}

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
		if checkDryRun(cmd, "[DRY-RUN] Would update product labels") {
			return nil
		}

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

// --- Admin API visibility commands ---

var productsHideCmd = &cobra.Command{
	Use:   "hide <product-id>",
	Short: "Hide a product (via Admin API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would hide product %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.HideProduct(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to hide product: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var productsPublishCmd = &cobra.Command{
	Use:   "publish <product-id>",
	Short: "Publish a product (via Admin API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would publish product %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.PublishProduct(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to publish product: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var productsUnpublishCmd = &cobra.Command{
	Use:   "unpublish <product-id>",
	Short: "Unpublish a product (via Admin API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would unpublish product %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.UnpublishProduct(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to unpublish product: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
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
	productsListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(productsListCmd.Flags(), "light", "li")

	productsCmd.AddCommand(productsGetCmd)
	productsGetCmd.Flags().String("by", "", "Find product by title instead of ID")
	productsGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(productsGetCmd.Flags(), "light", "li")

	productsCmd.AddCommand(productsCreateCmd)
	productsCreateCmd.Flags().String("title", "", "Product title (required)")
	productsCreateCmd.Flags().String("description", "", "Product description")
	productsCreateCmd.Flags().String("vendor", "", "Product vendor")
	productsCreateCmd.Flags().String("product-type", "", "Product type")
	productsCreateCmd.Flags().String("tags", "", "Comma-separated tags")
	productsCreateCmd.Flags().String("status", "", "Product status (active, draft, archived)")
	_ = productsCreateCmd.MarkFlagRequired("title")
	productsCreateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsUpdateCmd)
	productsUpdateCmd.Flags().String("title", "", "Product title")
	productsUpdateCmd.Flags().String("description", "", "Product description")
	productsUpdateCmd.Flags().String("vendor", "", "Product vendor")
	productsUpdateCmd.Flags().String("product-type", "", "Product type")
	productsUpdateCmd.Flags().String("tags", "", "Comma-separated tags")
	productsUpdateCmd.Flags().String("status", "", "Product status")
	productsUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsDeleteCmd)
	productsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	productsDeleteCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsSearchCmd)
	productsSearchCmd.Flags().String("q", "", "Search query")
	productsSearchCmd.Flags().String("status", "", "Filter by status")
	productsSearchCmd.Flags().String("vendor", "", "Filter by vendor")
	productsSearchCmd.Flags().Int("page", 1, "Page number")
	productsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	// Tags subcommands
	productsCmd.AddCommand(productsTagsCmd)
	productsTagsCmd.AddCommand(productsTagsReplaceCmd)
	productsTagsReplaceCmd.Flags().String("tags", "", "Comma-separated tags (required)")
	_ = productsTagsReplaceCmd.MarkFlagRequired("tags")
	productsTagsReplaceCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsTagsCmd.AddCommand(productsTagsUpdateCmd)
	productsTagsUpdateCmd.Flags().String("add", "", "Comma-separated tags to add")
	productsTagsUpdateCmd.Flags().String("remove", "", "Comma-separated tags to remove")
	productsTagsUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// Quantity / Price commands
	productsCmd.AddCommand(productsUpdateQuantityCmd)
	productsUpdateQuantityCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = productsUpdateQuantityCmd.MarkFlagRequired("quantity")
	productsUpdateQuantityCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsUpdatePriceCmd)
	productsUpdatePriceCmd.Flags().Float64("price", 0, "Price (required)")
	_ = productsUpdatePriceCmd.MarkFlagRequired("price")
	productsUpdatePriceCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsUpdateQuantityBySKUCmd)
	productsUpdateQuantityBySKUCmd.Flags().String("sku", "", "Product SKU (required)")
	productsUpdateQuantityBySKUCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = productsUpdateQuantityBySKUCmd.MarkFlagRequired("sku")
	_ = productsUpdateQuantityBySKUCmd.MarkFlagRequired("quantity")
	productsUpdateQuantityBySKUCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// Stocks subcommands
	productsCmd.AddCommand(productsStocksCmd)
	productsStocksCmd.AddCommand(productsStocksGetCmd)
	productsStocksCmd.AddCommand(productsStocksUpdateCmd)
	addJSONBodyFlags(productsStocksUpdateCmd)
	productsStocksUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsBulkUpdateStocksCmd)
	addJSONBodyFlags(productsBulkUpdateStocksCmd)
	productsBulkUpdateStocksCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsLockedInventoryCmd)

	// Images subcommands
	productsCmd.AddCommand(productsImagesCmd)
	productsImagesCmd.AddCommand(productsImagesAddCmd)
	productsImagesAddCmd.Flags().StringSlice("src", nil, "Image source URL (can be repeated; required)")
	productsImagesAddCmd.Flags().Int("position", 0, "Image position")
	_ = productsImagesAddCmd.MarkFlagRequired("src")
	productsImagesAddCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsImagesCmd.AddCommand(productsImagesDeleteCmd)
	productsImagesDeleteCmd.Flags().String("image-ids", "", "Comma-separated image IDs (required)")
	_ = productsImagesDeleteCmd.MarkFlagRequired("image-ids")
	productsImagesDeleteCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// Variations subcommands
	productsCmd.AddCommand(productsVariationsCmd)
	productsVariationsCmd.AddCommand(productsVariationsCreateCmd)
	productsVariationsCreateCmd.Flags().String("title", "", "Variation title")
	productsVariationsCreateCmd.Flags().String("sku", "", "Variation SKU")
	productsVariationsCreateCmd.Flags().Float64("price", 0, "Variation price")
	productsVariationsCreateCmd.Flags().Int("quantity", 0, "Variation quantity")
	productsVariationsCreateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsVariationsCmd.AddCommand(productsVariationsUpdateCmd)
	productsVariationsUpdateCmd.Flags().String("title", "", "Variation title")
	productsVariationsUpdateCmd.Flags().String("sku", "", "Variation SKU")
	productsVariationsUpdateCmd.Flags().Float64("price", 0, "Variation price")
	productsVariationsUpdateCmd.Flags().Int("quantity", 0, "Variation quantity")
	productsVariationsUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsVariationsCmd.AddCommand(productsVariationsDeleteCmd)
	productsVariationsDeleteCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsVariationsCmd.AddCommand(productsVariationsUpdateQuantityCmd)
	productsVariationsUpdateQuantityCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = productsVariationsUpdateQuantityCmd.MarkFlagRequired("quantity")
	productsVariationsUpdateQuantityCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsVariationsCmd.AddCommand(productsVariationsUpdatePriceCmd)
	productsVariationsUpdatePriceCmd.Flags().Float64("price", 0, "Price (required)")
	_ = productsVariationsUpdatePriceCmd.MarkFlagRequired("price")
	productsVariationsUpdatePriceCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// Bulk commands
	productsCmd.AddCommand(productsBulkDeleteCmd)
	productsBulkDeleteCmd.Flags().String("ids", "", "Comma-separated product IDs (required)")
	_ = productsBulkDeleteCmd.MarkFlagRequired("ids")
	productsBulkDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	productsBulkDeleteCmd.Flags().Bool("dry-run", false, "Show what would be deleted without making changes")

	productsCmd.AddCommand(productsBulkStatusCmd)
	addJSONBodyFlags(productsBulkStatusCmd)
	productsBulkStatusCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsBulkRetailStatusCmd)
	addJSONBodyFlags(productsBulkRetailStatusCmd)
	productsBulkRetailStatusCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsLabelsUpdateCmd)
	addJSONBodyFlags(productsLabelsUpdateCmd)
	productsLabelsUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsPromotionsCmd)

	productsCmd.AddCommand(productsSearchPostCmd)
	productsSearchPostCmd.Flags().String("body", "", "JSON request body")

	// Admin API visibility commands
	productsCmd.AddCommand(productsHideCmd)
	productsHideCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsPublishCmd)
	productsPublishCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	productsCmd.AddCommand(productsUnpublishCmd)
	productsUnpublishCmd.Flags().Bool("dry-run", false, "Preview without making changes")

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
			"hide", "publish", "unpublish",
			"metafields", "app-metafields",
		},
		IDPrefix: "product",
	})
}

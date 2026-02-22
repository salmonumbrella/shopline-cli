package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var addonProductsCmd = &cobra.Command{
	Use:   "addon-products",
	Short: "Manage add-on product bundles",
}

var addonProductsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List add-on products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		productID, _ := cmd.Flags().GetString("product-id")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.AddonProductsListOptions{
			Page:      page,
			PageSize:  pageSize,
			ProductID: productID,
			Status:    status,
		}

		resp, err := client.ListAddonProducts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list addon products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "PRODUCT ID", "PRICE", "STATUS", "CREATED"}
		var rows [][]string
		for _, ap := range resp.Items {
			price := ap.Price
			if ap.Currency != "" {
				price = ap.Price + " " + ap.Currency
			}
			rows = append(rows, []string{
				outfmt.FormatID("addon_product", ap.ID),
				ap.Title,
				ap.ProductID,
				price,
				ap.Status,
				ap.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d addon products\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var addonProductsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get add-on product details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		addonProductID, err := resolveOrArg(cmd, args, func(query string) (string, error) {
			resp, err := client.SearchAddonProducts(cmd.Context(), &api.AddonProductSearchOptions{
				Query: query, PageSize: 1,
			})
			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}
			if len(resp.Items) == 0 {
				return "", fmt.Errorf("no addon product found matching %q", query)
			}
			if len(resp.Items) > 1 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %d matches found, using first\n", len(resp.Items))
			}
			return resp.Items[0].ID, nil
		})
		if err != nil {
			return err
		}

		addonProduct, err := client.GetAddonProduct(cmd.Context(), addonProductID)
		if err != nil {
			return fmt.Errorf("failed to get addon product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(addonProduct)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Addon Product ID:  %s\n", addonProduct.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:             %s\n", addonProduct.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:        %s\n", addonProduct.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Variant ID:        %s\n", addonProduct.VariantID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Price:             %s %s\n", addonProduct.Price, addonProduct.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Quantity:          %d\n", addonProduct.Quantity)
		_, _ = fmt.Fprintf(outWriter(cmd), "Position:          %d\n", addonProduct.Position)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:            %s\n", addonProduct.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:       %s\n", addonProduct.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:           %s\n", addonProduct.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:           %s\n", addonProduct.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var addonProductsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an add-on product",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		productID, _ := cmd.Flags().GetString("product-id")
		variantID, _ := cmd.Flags().GetString("variant-id")
		price, _ := cmd.Flags().GetString("price")
		quantity, _ := cmd.Flags().GetInt("quantity")
		position, _ := cmd.Flags().GetInt("position")
		description, _ := cmd.Flags().GetString("description")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create addon product '%s' for product %s", title, productID)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AddonProductCreateRequest{
			Title:       title,
			ProductID:   productID,
			VariantID:   variantID,
			Price:       price,
			Quantity:    quantity,
			Position:    position,
			Description: description,
		}

		addonProduct, err := client.CreateAddonProduct(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create addon product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(addonProduct)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created addon product %s\n", addonProduct.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:       %s\n", addonProduct.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:  %s\n", addonProduct.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Price:       %s %s\n", addonProduct.Price, addonProduct.Currency)

		return nil
	},
}

var addonProductsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an add-on product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete addon product %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteAddonProduct(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete addon product: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted addon product %s\n", args[0])
		return nil
	},
}

var addonProductsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search add-on products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("q")
		productID, _ := cmd.Flags().GetString("product-id")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.AddonProductSearchOptions{
			Query:     query,
			ProductID: productID,
			Status:    status,
			Page:      page,
			PageSize:  pageSize,
		}

		resp, err := client.SearchAddonProducts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search addon products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "PRODUCT ID", "PRICE", "STATUS", "CREATED"}
		var rows [][]string
		for _, ap := range resp.Items {
			price := ap.Price
			if ap.Currency != "" {
				price = ap.Price + " " + ap.Currency
			}
			rows = append(rows, []string{
				outfmt.FormatID("addon_product", ap.ID),
				ap.Title,
				ap.ProductID,
				price,
				ap.Status,
				ap.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d addon products\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var addonProductsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an add-on product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update addon product %s", args[0])) {
			return nil
		}

		req := &api.AddonProductUpdateRequest{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			req.Title = &v
		}
		if cmd.Flags().Changed("price") {
			v, _ := cmd.Flags().GetString("price")
			req.Price = &v
		}
		if cmd.Flags().Changed("quantity") {
			v, _ := cmd.Flags().GetInt("quantity")
			req.Quantity = &v
		}
		if cmd.Flags().Changed("status") {
			v, _ := cmd.Flags().GetString("status")
			req.Status = &v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			req.Description = &v
		}
		if cmd.Flags().Changed("product-id") {
			v, _ := cmd.Flags().GetString("product-id")
			req.ProductID = &v
		}
		if cmd.Flags().Changed("variant-id") {
			v, _ := cmd.Flags().GetString("variant-id")
			req.VariantID = &v
		}
		if cmd.Flags().Changed("position") {
			v, _ := cmd.Flags().GetInt("position")
			req.Position = &v
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		addonProduct, err := client.UpdateAddonProduct(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update addon product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(addonProduct)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated addon product %s\n", addonProduct.ID)
		return nil
	},
}

var addonProductsUpdateQuantityCmd = &cobra.Command{
	Use:   "update-quantity <id>",
	Short: "Update add-on product quantity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		qty, _ := cmd.Flags().GetInt("quantity")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update addon product quantity for %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AddonProductQuantityRequest{Quantity: qty}
		addonProduct, err := client.UpdateAddonProductQuantity(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update addon product quantity: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(addonProduct)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated addon product %s quantity to %d\n", addonProduct.ID, addonProduct.Quantity)
		return nil
	},
}

var addonProductsUpdateQuantityBySKUCmd = &cobra.Command{
	Use:   "update-quantity-by-sku",
	Short: "Bulk update add-on product quantity by SKU",
	RunE: func(cmd *cobra.Command, args []string) error {
		sku, _ := cmd.Flags().GetString("sku")
		qty, _ := cmd.Flags().GetInt("quantity")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update addon product quantity for SKU %s", sku)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AddonProductQuantityBySKURequest{SKU: sku, Quantity: qty}
		addonProduct, err := client.UpdateAddonProductsQuantityBySKU(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to update addon products quantity by sku: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(addonProduct)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated addon product quantity for SKU %s\n", sku)
		return nil
	},
}

var addonProductsStocksCmd = &cobra.Command{
	Use:   "stocks",
	Short: "Manage add-on product stocks",
}

var addonProductsStocksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get add-on product stocks (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetAddonProductStocks(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get addon product stocks: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var addonProductsStocksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update add-on product stocks (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update addon product stocks for %s", args[0])) {
			return nil
		}

		var req api.AddonProductStocksUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateAddonProductStocks(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update addon product stocks: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(addonProductsCmd)

	addonProductsCmd.AddCommand(addonProductsListCmd)
	addonProductsListCmd.Flags().String("product-id", "", "Filter by parent product ID")
	addonProductsListCmd.Flags().String("status", "", "Filter by status (active, inactive)")
	addonProductsListCmd.Flags().Int("page", 1, "Page number")
	addonProductsListCmd.Flags().Int("page-size", 20, "Results per page")

	addonProductsCmd.AddCommand(addonProductsGetCmd)
	addonProductsGetCmd.Flags().String("by", "", "Find addon product by title instead of ID")

	addonProductsCmd.AddCommand(addonProductsCreateCmd)
	addonProductsCreateCmd.Flags().String("title", "", "Add-on product title")
	addonProductsCreateCmd.Flags().String("product-id", "", "Parent product ID")
	addonProductsCreateCmd.Flags().String("variant-id", "", "Variant ID (optional)")
	addonProductsCreateCmd.Flags().String("price", "", "Price")
	addonProductsCreateCmd.Flags().Int("quantity", 1, "Quantity")
	addonProductsCreateCmd.Flags().Int("position", 0, "Display position")
	addonProductsCreateCmd.Flags().String("description", "", "Description")
	_ = addonProductsCreateCmd.MarkFlagRequired("title")
	_ = addonProductsCreateCmd.MarkFlagRequired("product-id")

	addonProductsCmd.AddCommand(addonProductsDeleteCmd)

	addonProductsCmd.AddCommand(addonProductsSearchCmd)
	addonProductsSearchCmd.Flags().String("q", "", "Search query")
	addonProductsSearchCmd.Flags().String("product-id", "", "Filter by parent product ID")
	addonProductsSearchCmd.Flags().String("status", "", "Filter by status (active, inactive)")
	addonProductsSearchCmd.Flags().Int("page", 1, "Page number")
	addonProductsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	addonProductsCmd.AddCommand(addonProductsUpdateCmd)
	addonProductsUpdateCmd.Flags().String("title", "", "Add-on product title")
	addonProductsUpdateCmd.Flags().String("price", "", "Price")
	addonProductsUpdateCmd.Flags().Int("quantity", 0, "Quantity")
	addonProductsUpdateCmd.Flags().String("status", "", "Status (active, inactive)")
	addonProductsUpdateCmd.Flags().String("description", "", "Description")
	addonProductsUpdateCmd.Flags().String("product-id", "", "Parent product ID")
	addonProductsUpdateCmd.Flags().String("variant-id", "", "Variant ID")
	addonProductsUpdateCmd.Flags().Int("position", 0, "Display position")

	addonProductsCmd.AddCommand(addonProductsUpdateQuantityCmd)
	addonProductsUpdateQuantityCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = addonProductsUpdateQuantityCmd.MarkFlagRequired("quantity")

	addonProductsCmd.AddCommand(addonProductsUpdateQuantityBySKUCmd)
	addonProductsUpdateQuantityBySKUCmd.Flags().String("sku", "", "Add-on product SKU (required)")
	addonProductsUpdateQuantityBySKUCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = addonProductsUpdateQuantityBySKUCmd.MarkFlagRequired("sku")
	_ = addonProductsUpdateQuantityBySKUCmd.MarkFlagRequired("quantity")

	addonProductsCmd.AddCommand(addonProductsStocksCmd)
	addonProductsStocksCmd.AddCommand(addonProductsStocksGetCmd)
	addonProductsStocksCmd.AddCommand(addonProductsStocksUpdateCmd)
	addJSONBodyFlags(addonProductsStocksUpdateCmd)

	schema.Register(schema.Resource{
		Name:        "addon-products",
		Description: "Manage add-on product bundles",
		Commands:    []string{"list", "get", "search", "create", "delete", "update", "update-quantity", "update-quantity-by-sku", "stocks"},
		IDPrefix:    "addon_product",
	})
}

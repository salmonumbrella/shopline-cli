package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				ap.ID,
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
	Use:   "get <id>",
	Short: "Get add-on product details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		addonProduct, err := client.GetAddonProduct(cmd.Context(), args[0])
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would create addon product '%s' for product %s\n", title, productID)
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
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, _ = fmt.Fprintf(outWriter(cmd), "[DRY-RUN] Would delete addon product %s\n", args[0])
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

func init() {
	rootCmd.AddCommand(addonProductsCmd)

	addonProductsCmd.AddCommand(addonProductsListCmd)
	addonProductsListCmd.Flags().String("product-id", "", "Filter by parent product ID")
	addonProductsListCmd.Flags().String("status", "", "Filter by status (active, inactive)")
	addonProductsListCmd.Flags().Int("page", 1, "Page number")
	addonProductsListCmd.Flags().Int("page-size", 20, "Results per page")

	addonProductsCmd.AddCommand(addonProductsGetCmd)

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
}

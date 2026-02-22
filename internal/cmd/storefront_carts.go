package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var storefrontCartsCmd = &cobra.Command{
	Use:   "storefront-carts",
	Short: "Manage storefront shopping carts",
}

var storefrontCartsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storefront carts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StorefrontCartsListOptions{
			Page:       page,
			PageSize:   pageSize,
			CustomerID: customerID,
			Status:     status,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListStorefrontCarts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list storefront carts: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER ID", "ITEMS", "SUBTOTAL", "TOTAL", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("storefront_cart", c.ID),
				c.CustomerID,
				fmt.Sprintf("%d", c.ItemCount),
				c.Subtotal,
				c.TotalPrice,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d carts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storefrontCartsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get storefront cart details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		cart, err := client.GetStorefrontCart(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get storefront cart: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(cart)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Cart ID:       %s\n", cart.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:   %s\n", cart.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:         %s\n", cart.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:      %s\n", cart.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Item Count:    %d\n", cart.ItemCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Subtotal:      %s\n", cart.Subtotal)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Tax:     %s\n", cart.TotalTax)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Discount:%s\n", cart.TotalDiscount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Price:   %s\n", cart.TotalPrice)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", cart.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", cart.UpdatedAt.Format(time.RFC3339))

		if len(cart.Items) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nItems:\n")
			for _, item := range cart.Items {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (%s) x%d @ %s = %s\n",
					item.Title, item.VariantTitle, item.Quantity, item.Price, item.LineTotal)
			}
		}

		return nil
	},
}

var storefrontCartsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a storefront cart",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		email, _ := cmd.Flags().GetString("email")
		currency, _ := cmd.Flags().GetString("currency")

		msg := "[DRY-RUN] Would create storefront cart"
		if customerID != "" {
			msg += fmt.Sprintf(" for customer %s", customerID)
		}
		if checkDryRun(cmd, msg) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StorefrontCartCreateRequest{
			CustomerID: customerID,
			Email:      email,
			Currency:   currency,
		}

		cart, err := client.CreateStorefrontCart(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create storefront cart: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(cart)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created storefront cart %s\n", cart.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID: %s\n", cart.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:    %s\n", cart.Currency)

		return nil
	},
}

var storefrontCartsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a storefront cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete storefront cart %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteStorefrontCart(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete storefront cart: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted storefront cart %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storefrontCartsCmd)

	storefrontCartsCmd.AddCommand(storefrontCartsListCmd)
	storefrontCartsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	storefrontCartsListCmd.Flags().String("status", "", "Filter by status (active, abandoned, completed)")
	storefrontCartsListCmd.Flags().Int("page", 1, "Page number")
	storefrontCartsListCmd.Flags().Int("page-size", 20, "Results per page")

	storefrontCartsCmd.AddCommand(storefrontCartsGetCmd)

	storefrontCartsCmd.AddCommand(storefrontCartsCreateCmd)
	storefrontCartsCreateCmd.Flags().String("customer-id", "", "Customer ID")
	storefrontCartsCreateCmd.Flags().String("email", "", "Customer email")
	storefrontCartsCreateCmd.Flags().String("currency", "", "Cart currency")

	storefrontCartsCmd.AddCommand(storefrontCartsDeleteCmd)
}

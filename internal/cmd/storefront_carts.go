package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				c.ID,
				c.CustomerID,
				fmt.Sprintf("%d", c.ItemCount),
				c.Subtotal,
				c.TotalPrice,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d carts\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Cart ID:       %s\n", cart.ID)
		fmt.Printf("Customer ID:   %s\n", cart.CustomerID)
		fmt.Printf("Email:         %s\n", cart.Email)
		fmt.Printf("Currency:      %s\n", cart.Currency)
		fmt.Printf("Item Count:    %d\n", cart.ItemCount)
		fmt.Printf("Subtotal:      %s\n", cart.Subtotal)
		fmt.Printf("Total Tax:     %s\n", cart.TotalTax)
		fmt.Printf("Total Discount:%s\n", cart.TotalDiscount)
		fmt.Printf("Total Price:   %s\n", cart.TotalPrice)
		fmt.Printf("Created:       %s\n", cart.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:       %s\n", cart.UpdatedAt.Format(time.RFC3339))

		if len(cart.Items) > 0 {
			fmt.Printf("\nItems:\n")
			for _, item := range cart.Items {
				fmt.Printf("  - %s (%s) x%d @ %s = %s\n",
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create storefront cart")
			if customerID != "" {
				fmt.Printf(" for customer %s", customerID)
			}
			fmt.Println()
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

		fmt.Printf("Created storefront cart %s\n", cart.ID)
		fmt.Printf("Customer ID: %s\n", cart.CustomerID)
		fmt.Printf("Currency:    %s\n", cart.Currency)

		return nil
	},
}

var storefrontCartsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a storefront cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete storefront cart %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteStorefrontCart(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete storefront cart: %w", err)
		}

		fmt.Printf("Deleted storefront cart %s\n", args[0])
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

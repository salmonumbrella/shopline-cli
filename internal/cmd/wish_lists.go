package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var wishListsCmd = &cobra.Command{
	Use:   "wish-lists",
	Short: "Manage customer wish lists",
}

var wishListsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List wish lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.WishListsListOptions{
			Page:       page,
			PageSize:   pageSize,
			CustomerID: customerID,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListWishLists(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list wish lists: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER", "NAME", "ITEMS", "PUBLIC", "CREATED"}
		var rows [][]string
		for _, wl := range resp.Items {
			isPublic := "No"
			if wl.IsPublic {
				isPublic = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("wish_list", wl.ID),
				wl.CustomerID,
				wl.Name,
				fmt.Sprintf("%d", wl.ItemCount),
				isPublic,
				wl.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d wish lists\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var wishListsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get wish list details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		wishList, err := client.GetWishList(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get wish list: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(wishList)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Wish List ID:  %s\n", wishList.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:   %s\n", wishList.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s\n", wishList.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:   %s\n", wishList.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Default:       %v\n", wishList.IsDefault)
		_, _ = fmt.Fprintf(outWriter(cmd), "Public:        %v\n", wishList.IsPublic)
		if wishList.IsPublic && wishList.ShareURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Share URL:     %s\n", wishList.ShareURL)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Item Count:    %d\n", wishList.ItemCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", wishList.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", wishList.UpdatedAt.Format(time.RFC3339))

		if len(wishList.Items) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nItems:\n")
			for _, item := range wishList.Items {
				available := "unavailable"
				if item.Available {
					available = "available"
				}
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s", item.Title)
				if item.VariantTitle != "" {
					_, _ = fmt.Fprintf(outWriter(cmd), " (%s)", item.VariantTitle)
				}
				_, _ = fmt.Fprintf(outWriter(cmd), " - %s %s [%s]\n", item.Price, item.Currency, available)
				if item.Notes != "" {
					_, _ = fmt.Fprintf(outWriter(cmd), "    Notes: %s\n", item.Notes)
				}
			}
		}

		return nil
	},
}

var wishListsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a wish list",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		isDefault, _ := cmd.Flags().GetBool("default")
		isPublic, _ := cmd.Flags().GetBool("public")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create wish list '%s' for customer %s", name, customerID)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.WishListCreateRequest{
			CustomerID:  customerID,
			Name:        name,
			Description: description,
			IsDefault:   isDefault,
			IsPublic:    isPublic,
		}

		wishList, err := client.CreateWishList(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create wish list: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(wishList)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created wish list %s\n", wishList.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:        %s\n", wishList.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID: %s\n", wishList.CustomerID)
		if wishList.IsPublic && wishList.ShareURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Share URL:   %s\n", wishList.ShareURL)
		}

		return nil
	},
}

var wishListsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a wish list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete wish list %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteWishList(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete wish list: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted wish list %s\n", args[0])
		return nil
	},
}

var wishListsAddItemCmd = &cobra.Command{
	Use:   "add-item <wish-list-id>",
	Short: "Add an item to a wish list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product-id")
		variantID, _ := cmd.Flags().GetString("variant-id")
		quantity, _ := cmd.Flags().GetInt("quantity")
		priority, _ := cmd.Flags().GetInt("priority")
		notes, _ := cmd.Flags().GetString("notes")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add product %s to wish list %s", productID, args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.WishListItemCreateRequest{
			ProductID: productID,
			VariantID: variantID,
			Quantity:  quantity,
			Priority:  priority,
			Notes:     notes,
		}

		item, err := client.AddWishListItem(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to add item to wish list: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(item)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Added item %s to wish list\n", item.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product: %s\n", item.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Price:   %s %s\n", item.Price, item.Currency)

		return nil
	},
}

var wishListsRemoveItemCmd = &cobra.Command{
	Use:   "remove-item <wish-list-id> <item-id>",
	Short: "Remove an item from a wish list",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would remove item %s from wish list %s", args[1], args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.RemoveWishListItem(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to remove item from wish list: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Removed item %s from wish list %s\n", args[1], args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(wishListsCmd)

	wishListsCmd.AddCommand(wishListsListCmd)
	wishListsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	wishListsListCmd.Flags().Int("page", 1, "Page number")
	wishListsListCmd.Flags().Int("page-size", 20, "Results per page")

	wishListsCmd.AddCommand(wishListsGetCmd)

	wishListsCmd.AddCommand(wishListsCreateCmd)
	wishListsCreateCmd.Flags().String("customer-id", "", "Customer ID")
	wishListsCreateCmd.Flags().String("name", "", "Wish list name")
	wishListsCreateCmd.Flags().String("description", "", "Wish list description")
	wishListsCreateCmd.Flags().Bool("default", false, "Set as default wish list")
	wishListsCreateCmd.Flags().Bool("public", false, "Make wish list public")
	_ = wishListsCreateCmd.MarkFlagRequired("customer-id")
	_ = wishListsCreateCmd.MarkFlagRequired("name")

	wishListsCmd.AddCommand(wishListsDeleteCmd)

	wishListsCmd.AddCommand(wishListsAddItemCmd)
	wishListsAddItemCmd.Flags().String("product-id", "", "Product ID to add")
	wishListsAddItemCmd.Flags().String("variant-id", "", "Variant ID (optional)")
	wishListsAddItemCmd.Flags().Int("quantity", 1, "Quantity")
	wishListsAddItemCmd.Flags().Int("priority", 0, "Priority (higher = more wanted)")
	wishListsAddItemCmd.Flags().String("notes", "", "Notes for this item")
	_ = wishListsAddItemCmd.MarkFlagRequired("product-id")

	wishListsCmd.AddCommand(wishListsRemoveItemCmd)
}

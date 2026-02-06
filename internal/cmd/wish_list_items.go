package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var wishListItemsCmd = &cobra.Command{
	Use:     "wish-list-items",
	Aliases: []string{"wishlist-items", "wishitems"},
	Short:   "Manage wish list items (documented endpoints)",
}

var wishListItemsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List wish list items",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListWishListItems(cmd.Context(), &api.WishListItemsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list wish list items: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var wishListItemsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new", "add"},
	Short:   "Create a wish list item",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.CreateWishListItem(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create wish list item: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var wishListItemsDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "rm"},
	Short:   "Delete wish list items",
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete wish list items? (use --yes to confirm)\n")
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				return nil
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.DeleteWishListItems(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to delete wish list items: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(wishListItemsCmd)

	wishListItemsCmd.AddCommand(wishListItemsListCmd)
	wishListItemsListCmd.Flags().Int("page", 1, "Page number")
	wishListItemsListCmd.Flags().Int("page-size", 20, "Results per page")

	wishListItemsCmd.AddCommand(wishListItemsCreateCmd)
	addJSONBodyFlags(wishListItemsCreateCmd)

	wishListItemsCmd.AddCommand(wishListItemsDeleteCmd)
	addJSONBodyFlags(wishListItemsDeleteCmd)
}

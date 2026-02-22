package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

// ============================
// carts (Open API)
// ============================

var cartsCmd = &cobra.Command{
	Use:     "carts",
	Aliases: []string{"cart"},
	Short:   "Manage carts (Open API)",
}

var cartsExchangeCmd = &cobra.Command{
	Use:   "exchange",
	Short: "Exchange a cart",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would exchange cart") {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ExchangeCart(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to exchange cart: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var cartsPrepareCmd = &cobra.Command{
	Use:   "prepare <cart-id>",
	Short: "Prepare a cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would prepare cart %s", args[0])) {
			return nil
		}

		// Some stores may require a body, some may not. Allow optional JSON.
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		var req json.RawMessage
		var hasBody bool
		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			req, err = readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
			hasBody = true
		}

		var anyBody any
		if hasBody {
			anyBody = req
		}

		resp, err := client.PrepareCart(cmd.Context(), args[0], anyBody)
		if err != nil {
			return fmt.Errorf("failed to prepare cart: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

// carts items

var cartsItemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Manage cart items",
}

var cartsItemsAddCmd = &cobra.Command{
	Use:     "add <cart-id>",
	Aliases: []string{"create", "new"},
	Short:   "Add items to a cart",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add items to cart %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.AddCartItems(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to add cart items: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var cartsItemsUpdateCmd = &cobra.Command{
	Use:   "update <cart-id>",
	Short: "Update items in a cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update items in cart %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateCartItems(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update cart items: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var cartsItemsDeleteCmd = &cobra.Command{
	Use:   "delete <cart-id>",
	Short: "Delete items from a cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete items from cart %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Are you sure you want to delete cart items from %s? (use --yes to confirm)\n", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.DeleteCartItems(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to delete cart items: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

// carts items metafields

var cartsItemsMetafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage cart item metafields",
}

var cartsItemsMetafieldsListCmd = &cobra.Command{
	Use:   "list <cart-id>",
	Short: "List cart item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListCartItemMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list cart item metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var cartsItemsMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <cart-id>",
	Short: "Bulk create cart item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk create cart item metafields for %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkCreateCartItemMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create cart item metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var cartsItemsMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <cart-id>",
	Short: "Bulk update cart item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk update cart item metafields for %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkUpdateCartItemMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update cart item metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var cartsItemsMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <cart-id>",
	Short: "Bulk delete cart item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk delete cart item metafields for %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Are you sure you want to bulk delete cart item metafields for %s? (use --yes to confirm)\n", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkDeleteCartItemMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete cart item metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// carts items app-metafields

var cartsItemsAppMetafieldsCmd = &cobra.Command{
	Use:     "app-metafields",
	Aliases: []string{"app-metafield"},
	Short:   "Manage cart item app metafields",
}

var cartsItemsAppMetafieldsListCmd = &cobra.Command{
	Use:   "list <cart-id>",
	Short: "List cart item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListCartItemAppMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list cart item app metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var cartsItemsAppMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <cart-id>",
	Short: "Bulk create cart item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk create cart item app metafields for %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkCreateCartItemAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create cart item app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var cartsItemsAppMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <cart-id>",
	Short: "Bulk update cart item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk update cart item app metafields for %s", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkUpdateCartItemAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update cart item app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var cartsItemsAppMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <cart-id>",
	Short: "Bulk delete cart item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk delete cart item app metafields for %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Are you sure you want to bulk delete cart item app metafields for %s? (use --yes to confirm)\n", args[0])) {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkDeleteCartItemAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete cart item app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cartsCmd)

	cartsCmd.AddCommand(cartsExchangeCmd)
	addJSONBodyFlags(cartsExchangeCmd)

	cartsCmd.AddCommand(cartsPrepareCmd)
	addJSONBodyFlags(cartsPrepareCmd)

	cartsCmd.AddCommand(cartsItemsCmd)
	cartsItemsCmd.AddCommand(cartsItemsAddCmd)
	cartsItemsCmd.AddCommand(cartsItemsUpdateCmd)
	cartsItemsCmd.AddCommand(cartsItemsDeleteCmd)
	addJSONBodyFlags(cartsItemsAddCmd)
	addJSONBodyFlags(cartsItemsUpdateCmd)
	addJSONBodyFlags(cartsItemsDeleteCmd)

	cartsItemsCmd.AddCommand(cartsItemsMetafieldsCmd)
	cartsItemsMetafieldsCmd.AddCommand(cartsItemsMetafieldsListCmd)
	cartsItemsMetafieldsCmd.AddCommand(cartsItemsMetafieldsBulkCreateCmd)
	cartsItemsMetafieldsCmd.AddCommand(cartsItemsMetafieldsBulkUpdateCmd)
	cartsItemsMetafieldsCmd.AddCommand(cartsItemsMetafieldsBulkDeleteCmd)
	addJSONBodyFlags(cartsItemsMetafieldsBulkCreateCmd)
	addJSONBodyFlags(cartsItemsMetafieldsBulkUpdateCmd)
	addJSONBodyFlags(cartsItemsMetafieldsBulkDeleteCmd)

	cartsItemsCmd.AddCommand(cartsItemsAppMetafieldsCmd)
	cartsItemsAppMetafieldsCmd.AddCommand(cartsItemsAppMetafieldsListCmd)
	cartsItemsAppMetafieldsCmd.AddCommand(cartsItemsAppMetafieldsBulkCreateCmd)
	cartsItemsAppMetafieldsCmd.AddCommand(cartsItemsAppMetafieldsBulkUpdateCmd)
	cartsItemsAppMetafieldsCmd.AddCommand(cartsItemsAppMetafieldsBulkDeleteCmd)
	addJSONBodyFlags(cartsItemsAppMetafieldsBulkCreateCmd)
	addJSONBodyFlags(cartsItemsAppMetafieldsBulkUpdateCmd)
	addJSONBodyFlags(cartsItemsAppMetafieldsBulkDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "carts",
		Description: "Manage carts",
		Commands:    []string{"exchange", "prepare", "items"},
		IDPrefix:    "cart",
	})
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ============================
// orders metafields (non-app)
// ============================

var ordersMetafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage order metafields",
}

var ordersMetafieldsListCmd = &cobra.Command{
	Use:   "list <order-id>",
	Short: "List metafields attached to an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListOrderMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersMetafieldsGetCmd = &cobra.Command{
	Use:   "get <order-id> <metafield-id>",
	Short: "Get a specific order metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetOrderMetafield(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get order metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersMetafieldsCreateCmd = &cobra.Command{
	Use:   "create <order-id>",
	Short: "Create an order metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.CreateOrderMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create order metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <order-id> <metafield-id>",
	Short: "Update an order metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateOrderMetafield(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to update order metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <order-id> <metafield-id>",
	Short: "Delete an order metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete order metafield %s for order %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteOrderMetafield(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete order metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted order metafield %s (order %s)\n", args[1], args[0])
		return nil
	},
}

var ordersMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <order-id>",
	Short: "Bulk create order metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-create metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkCreateOrderMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create order metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <order-id>",
	Short: "Bulk update order metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-update metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkUpdateOrderMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update order metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <order-id>",
	Short: "Bulk delete order metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-delete metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkDeleteOrderMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete order metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// ============================
// orders app-metafields (app)
// ============================

var ordersAppMetafieldsCmd = &cobra.Command{
	Use:   "app-metafields",
	Short: "Manage order app metafields",
}

var ordersAppMetafieldsListCmd = &cobra.Command{
	Use:   "list <order-id>",
	Short: "List app metafields attached to an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListOrderAppMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order app metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersAppMetafieldsGetCmd = &cobra.Command{
	Use:   "get <order-id> <metafield-id>",
	Short: "Get a specific order app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetOrderAppMetafield(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get order app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersAppMetafieldsCreateCmd = &cobra.Command{
	Use:   "create <order-id>",
	Short: "Create an order app metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.CreateOrderAppMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create order app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersAppMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <order-id> <metafield-id>",
	Short: "Update an order app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateOrderAppMetafield(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to update order app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersAppMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <order-id> <metafield-id>",
	Short: "Delete an order app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete order app metafield %s for order %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteOrderAppMetafield(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete order app metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted order app metafield %s (order %s)\n", args[1], args[0])
		return nil
	},
}

var ordersAppMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <order-id>",
	Short: "Bulk create order app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-create app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkCreateOrderAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create order app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersAppMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <order-id>",
	Short: "Bulk update order app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-update app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkUpdateOrderAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update order app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersAppMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <order-id>",
	Short: "Bulk delete order app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-delete app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkDeleteOrderAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete order app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// ==================================
// orders item-metafields (non-app)
// ==================================

var ordersItemMetafieldsCmd = &cobra.Command{
	Use:   "item-metafields",
	Short: "Manage order item metafields",
}

var ordersItemMetafieldsListCmd = &cobra.Command{
	Use:   "list <order-id>",
	Short: "List metafields attached to order items of an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListOrderItemMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order item metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersItemMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <order-id>",
	Short: "Bulk create order item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-create item metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkCreateOrderItemMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create order item metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersItemMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <order-id>",
	Short: "Bulk update order item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-update item metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkUpdateOrderItemMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update order item metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersItemMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <order-id>",
	Short: "Bulk delete order item metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-delete item metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkDeleteOrderItemMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete order item metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// ==================================
// orders item app-metafields (app)
// ==================================

var ordersItemAppMetafieldsCmd = &cobra.Command{
	Use:   "item-app-metafields",
	Short: "Manage order item app metafields",
}

var ordersItemAppMetafieldsListCmd = &cobra.Command{
	Use:   "list <order-id>",
	Short: "List app metafields attached to order items of an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListOrderItemAppMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order item app metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var ordersItemAppMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <order-id>",
	Short: "Bulk create order item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-create item app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkCreateOrderItemAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create order item app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersItemAppMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <order-id>",
	Short: "Bulk update order item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-update item app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkUpdateOrderItemAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update order item app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var ordersItemAppMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <order-id>",
	Short: "Bulk delete order item app metafields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would bulk-delete item app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		if err := client.BulkDeleteOrderItemAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete order item app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

func init() {
	// orders metafields
	ordersCmd.AddCommand(ordersMetafieldsCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsListCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsGetCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsCreateCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsUpdateCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsDeleteCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsBulkCreateCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsBulkUpdateCmd)
	ordersMetafieldsCmd.AddCommand(ordersMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		ordersMetafieldsCreateCmd,
		ordersMetafieldsUpdateCmd,
		ordersMetafieldsBulkCreateCmd,
		ordersMetafieldsBulkUpdateCmd,
		ordersMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}

	// orders app-metafields
	ordersCmd.AddCommand(ordersAppMetafieldsCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsListCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsGetCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsCreateCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsUpdateCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsDeleteCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsBulkCreateCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsBulkUpdateCmd)
	ordersAppMetafieldsCmd.AddCommand(ordersAppMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		ordersAppMetafieldsCreateCmd,
		ordersAppMetafieldsUpdateCmd,
		ordersAppMetafieldsBulkCreateCmd,
		ordersAppMetafieldsBulkUpdateCmd,
		ordersAppMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}

	// orders item-metafields
	ordersCmd.AddCommand(ordersItemMetafieldsCmd)
	ordersItemMetafieldsCmd.AddCommand(ordersItemMetafieldsListCmd)
	ordersItemMetafieldsCmd.AddCommand(ordersItemMetafieldsBulkCreateCmd)
	ordersItemMetafieldsCmd.AddCommand(ordersItemMetafieldsBulkUpdateCmd)
	ordersItemMetafieldsCmd.AddCommand(ordersItemMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		ordersItemMetafieldsBulkCreateCmd,
		ordersItemMetafieldsBulkUpdateCmd,
		ordersItemMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}

	// orders item-app-metafields
	ordersCmd.AddCommand(ordersItemAppMetafieldsCmd)
	ordersItemAppMetafieldsCmd.AddCommand(ordersItemAppMetafieldsListCmd)
	ordersItemAppMetafieldsCmd.AddCommand(ordersItemAppMetafieldsBulkCreateCmd)
	ordersItemAppMetafieldsCmd.AddCommand(ordersItemAppMetafieldsBulkUpdateCmd)
	ordersItemAppMetafieldsCmd.AddCommand(ordersItemAppMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		ordersItemAppMetafieldsBulkCreateCmd,
		ordersItemAppMetafieldsBulkUpdateCmd,
		ordersItemAppMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}
}

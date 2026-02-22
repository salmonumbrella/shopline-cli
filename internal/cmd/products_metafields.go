package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ============================
// products metafields (non-app)
// ============================

var productsMetafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage product metafields",
}

var productsMetafieldsListCmd = &cobra.Command{
	Use:   "list <product-id>",
	Short: "List metafields attached to a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListProductMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list product metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsMetafieldsGetCmd = &cobra.Command{
	Use:   "get <product-id> <metafield-id>",
	Short: "Get a specific product metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetProductMetafield(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get product metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsMetafieldsCreateCmd = &cobra.Command{
	Use:   "create <product-id>",
	Short: "Create a product metafield",
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
		resp, err := client.CreateProductMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create product metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <product-id> <metafield-id>",
	Short: "Update a product metafield",
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
		resp, err := client.UpdateProductMetafield(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to update product metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <product-id> <metafield-id>",
	Short: "Delete a product metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete product metafield %s for product %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteProductMetafield(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete product metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted product metafield %s (product %s)\n", args[1], args[0])
		return nil
	},
}

var productsMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <product-id>",
	Short: "Bulk create product metafields",
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
		if err := client.BulkCreateProductMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create product metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var productsMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <product-id>",
	Short: "Bulk update product metafields",
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
		if err := client.BulkUpdateProductMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update product metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var productsMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <product-id>",
	Short: "Bulk delete product metafields",
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
		if err := client.BulkDeleteProductMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete product metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// ============================
// products app-metafields (app)
// ============================

var productsAppMetafieldsCmd = &cobra.Command{
	Use:   "app-metafields",
	Short: "Manage product app metafields",
}

var productsAppMetafieldsListCmd = &cobra.Command{
	Use:   "list <product-id>",
	Short: "List app metafields attached to a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListProductAppMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list product app metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsAppMetafieldsGetCmd = &cobra.Command{
	Use:   "get <product-id> <metafield-id>",
	Short: "Get a specific product app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetProductAppMetafield(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get product app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsAppMetafieldsCreateCmd = &cobra.Command{
	Use:   "create <product-id>",
	Short: "Create a product app metafield",
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
		resp, err := client.CreateProductAppMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create product app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsAppMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <product-id> <metafield-id>",
	Short: "Update a product app metafield",
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
		resp, err := client.UpdateProductAppMetafield(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to update product app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productsAppMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <product-id> <metafield-id>",
	Short: "Delete a product app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete product app metafield %s for product %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteProductAppMetafield(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete product app metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted product app metafield %s (product %s)\n", args[1], args[0])
		return nil
	},
}

var productsAppMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <product-id>",
	Short: "Bulk create product app metafields",
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
		if err := client.BulkCreateProductAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create product app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var productsAppMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <product-id>",
	Short: "Bulk update product app metafields",
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
		if err := client.BulkUpdateProductAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update product app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var productsAppMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <product-id>",
	Short: "Bulk delete product app metafields",
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
		if err := client.BulkDeleteProductAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete product app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

func init() {
	// products metafields
	productsCmd.AddCommand(productsMetafieldsCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsListCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsGetCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsCreateCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsUpdateCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsDeleteCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsBulkCreateCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsBulkUpdateCmd)
	productsMetafieldsCmd.AddCommand(productsMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		productsMetafieldsCreateCmd,
		productsMetafieldsUpdateCmd,
		productsMetafieldsBulkCreateCmd,
		productsMetafieldsBulkUpdateCmd,
		productsMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}

	// products app-metafields
	productsCmd.AddCommand(productsAppMetafieldsCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsListCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsGetCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsCreateCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsUpdateCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsDeleteCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsBulkCreateCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsBulkUpdateCmd)
	productsAppMetafieldsCmd.AddCommand(productsAppMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		productsAppMetafieldsCreateCmd,
		productsAppMetafieldsUpdateCmd,
		productsAppMetafieldsBulkCreateCmd,
		productsAppMetafieldsBulkUpdateCmd,
		productsAppMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}
}

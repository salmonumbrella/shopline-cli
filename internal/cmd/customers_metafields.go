package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ============================
// customers metafields (non-app)
// ============================

var customersMetafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage customer metafields",
}

var customersMetafieldsListCmd = &cobra.Command{
	Use:   "list <customer-id>",
	Short: "List metafields attached to a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListCustomerMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list customer metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersMetafieldsGetCmd = &cobra.Command{
	Use:   "get <customer-id> <metafield-id>",
	Short: "Get a specific customer metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomerMetafield(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get customer metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersMetafieldsCreateCmd = &cobra.Command{
	Use:   "create <customer-id>",
	Short: "Create a customer metafield",
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
		resp, err := client.CreateCustomerMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create customer metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <customer-id> <metafield-id>",
	Short: "Update a customer metafield",
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
		resp, err := client.UpdateCustomerMetafield(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to update customer metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <customer-id> <metafield-id>",
	Short: "Delete a customer metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete customer metafield %s for customer %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteCustomerMetafield(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete customer metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted customer metafield %s (customer %s)\n", args[1], args[0])
		return nil
	},
}

var customersMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <customer-id>",
	Short: "Bulk create customer metafields",
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
		if err := client.BulkCreateCustomerMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create customer metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var customersMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <customer-id>",
	Short: "Bulk update customer metafields",
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
		if err := client.BulkUpdateCustomerMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update customer metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var customersMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <customer-id>",
	Short: "Bulk delete customer metafields",
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
		if err := client.BulkDeleteCustomerMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete customer metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// ============================
// customers app-metafields (app)
// ============================

var customersAppMetafieldsCmd = &cobra.Command{
	Use:   "app-metafields",
	Short: "Manage customer app metafields",
}

var customersAppMetafieldsListCmd = &cobra.Command{
	Use:   "list <customer-id>",
	Short: "List app metafields attached to a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListCustomerAppMetafields(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list customer app metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersAppMetafieldsGetCmd = &cobra.Command{
	Use:   "get <customer-id> <metafield-id>",
	Short: "Get a specific customer app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomerAppMetafield(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get customer app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersAppMetafieldsCreateCmd = &cobra.Command{
	Use:   "create <customer-id>",
	Short: "Create a customer app metafield",
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
		resp, err := client.CreateCustomerAppMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create customer app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersAppMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <customer-id> <metafield-id>",
	Short: "Update a customer app metafield",
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
		resp, err := client.UpdateCustomerAppMetafield(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("failed to update customer app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersAppMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <customer-id> <metafield-id>",
	Short: "Delete a customer app metafield",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete customer app metafield %s for customer %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteCustomerAppMetafield(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete customer app metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted customer app metafield %s (customer %s)\n", args[1], args[0])
		return nil
	},
}

var customersAppMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create <customer-id>",
	Short: "Bulk create customer app metafields",
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
		if err := client.BulkCreateCustomerAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk create customer app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var customersAppMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update <customer-id>",
	Short: "Bulk update customer app metafields",
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
		if err := client.BulkUpdateCustomerAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk update customer app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var customersAppMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <customer-id>",
	Short: "Bulk delete customer app metafields",
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
		if err := client.BulkDeleteCustomerAppMetafields(cmd.Context(), args[0], body); err != nil {
			return fmt.Errorf("failed to bulk delete customer app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

func init() {
	// customers metafields
	customersCmd.AddCommand(customersMetafieldsCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsListCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsGetCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsCreateCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsUpdateCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsDeleteCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsBulkCreateCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsBulkUpdateCmd)
	customersMetafieldsCmd.AddCommand(customersMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		customersMetafieldsCreateCmd,
		customersMetafieldsUpdateCmd,
		customersMetafieldsBulkCreateCmd,
		customersMetafieldsBulkUpdateCmd,
		customersMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}

	// customers app-metafields
	customersCmd.AddCommand(customersAppMetafieldsCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsListCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsGetCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsCreateCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsUpdateCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsDeleteCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsBulkCreateCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsBulkUpdateCmd)
	customersAppMetafieldsCmd.AddCommand(customersAppMetafieldsBulkDeleteCmd)

	for _, c := range []*cobra.Command{
		customersAppMetafieldsCreateCmd,
		customersAppMetafieldsUpdateCmd,
		customersAppMetafieldsBulkCreateCmd,
		customersAppMetafieldsBulkUpdateCmd,
		customersAppMetafieldsBulkDeleteCmd,
	} {
		addJSONBodyFlags(c)
	}
}

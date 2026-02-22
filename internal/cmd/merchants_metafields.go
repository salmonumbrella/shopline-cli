package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ============================
// merchants metafields (current merchant)
// ============================

var merchantsMetafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage merchant metafields (current merchant)",
}

var merchantsMetafieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List merchant metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListMerchantMetafields(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list merchant metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsMetafieldsGetCmd = &cobra.Command{
	Use:   "get <metafield-id>",
	Short: "Get a merchant metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetMerchantMetafield(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get merchant metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsMetafieldsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a merchant metafield",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create metafield for current merchant") {
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
		resp, err := client.CreateMerchantMetafield(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create merchant metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <metafield-id>",
	Short: "Update a merchant metafield",
	Args:  cobra.ExactArgs(1),
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
		resp, err := client.UpdateMerchantMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update merchant metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <metafield-id>",
	Short: "Delete a merchant metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete merchant metafield %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteMerchantMetafield(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete merchant metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted merchant metafield %s\n", args[0])
		return nil
	},
}

var merchantsMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create",
	Short: "Bulk create merchant metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would bulk-create metafield for current merchant") {
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
		if err := client.BulkCreateMerchantMetafields(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk create merchant metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var merchantsMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Bulk update merchant metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would bulk-update metafield for current merchant") {
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
		if err := client.BulkUpdateMerchantMetafields(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk update merchant metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var merchantsMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Bulk delete merchant metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would bulk-delete metafield for current merchant") {
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
		if err := client.BulkDeleteMerchantMetafields(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk delete merchant metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

// ============================
// merchants app-metafields (current merchant)
// ============================

var merchantsAppMetafieldsCmd = &cobra.Command{
	Use:   "app-metafields",
	Short: "Manage merchant app metafields (current merchant)",
}

var merchantsAppMetafieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List merchant app metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListMerchantAppMetafields(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list merchant app metafields: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsAppMetafieldsGetCmd = &cobra.Command{
	Use:   "get <metafield-id>",
	Short: "Get a merchant app metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetMerchantAppMetafield(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get merchant app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsAppMetafieldsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a merchant app metafield",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create app-metafield for current merchant") {
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
		resp, err := client.CreateMerchantAppMetafield(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create merchant app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsAppMetafieldsUpdateCmd = &cobra.Command{
	Use:   "update <metafield-id>",
	Short: "Update a merchant app metafield",
	Args:  cobra.ExactArgs(1),
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
		resp, err := client.UpdateMerchantAppMetafield(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update merchant app metafield: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsAppMetafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <metafield-id>",
	Short: "Delete a merchant app metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete app-metafield for %s", args[0])) {
			return nil
		}
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if !confirmAction(cmd, fmt.Sprintf("Delete merchant app metafield %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}
		if err := client.DeleteMerchantAppMetafield(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete merchant app metafield: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted merchant app metafield %s\n", args[0])
		return nil
	},
}

var merchantsAppMetafieldsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create",
	Short: "Bulk create merchant app metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would bulk-create app-metafield for current merchant") {
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
		if err := client.BulkCreateMerchantAppMetafields(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk create merchant app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var merchantsAppMetafieldsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Bulk update merchant app metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would bulk-update app-metafield for current merchant") {
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
		if err := client.BulkUpdateMerchantAppMetafields(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk update merchant app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

var merchantsAppMetafieldsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Bulk delete merchant app metafields",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would bulk-delete app-metafield for current merchant") {
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
		if err := client.BulkDeleteMerchantAppMetafields(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to bulk delete merchant app metafields: %w", err)
		}
		_, _ = fmt.Fprintln(outWriter(cmd), "OK")
		return nil
	},
}

func init() {
	merchantsCmd.AddCommand(merchantsMetafieldsCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsListCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsGetCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsCreateCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsUpdateCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsDeleteCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsBulkCreateCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsBulkUpdateCmd)
	merchantsMetafieldsCmd.AddCommand(merchantsMetafieldsBulkDeleteCmd)

	addJSONBodyFlags(merchantsMetafieldsCreateCmd)
	addJSONBodyFlags(merchantsMetafieldsUpdateCmd)
	addJSONBodyFlags(merchantsMetafieldsBulkCreateCmd)
	addJSONBodyFlags(merchantsMetafieldsBulkUpdateCmd)
	addJSONBodyFlags(merchantsMetafieldsBulkDeleteCmd)

	merchantsCmd.AddCommand(merchantsAppMetafieldsCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsListCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsGetCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsCreateCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsUpdateCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsDeleteCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsBulkCreateCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsBulkUpdateCmd)
	merchantsAppMetafieldsCmd.AddCommand(merchantsAppMetafieldsBulkDeleteCmd)

	addJSONBodyFlags(merchantsAppMetafieldsCreateCmd)
	addJSONBodyFlags(merchantsAppMetafieldsUpdateCmd)
	addJSONBodyFlags(merchantsAppMetafieldsBulkCreateCmd)
	addJSONBodyFlags(merchantsAppMetafieldsBulkUpdateCmd)
	addJSONBodyFlags(merchantsAppMetafieldsBulkDeleteCmd)
}

package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var posCmd = &cobra.Command{
	Use:   "pos",
	Short: "Manage point-of-sale (POS) resources",
}

var posPurchaseOrdersCmd = &cobra.Command{
	Use:   "purchase-orders",
	Short: "Manage POS purchase orders (documented endpoints)",
}

var posPurchaseOrdersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List POS purchase orders (raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListPOSPurchaseOrders(cmd.Context(), &api.POSPurchaseOrdersListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list POS purchase orders: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var posPurchaseOrdersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get POS purchase order details (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetPOSPurchaseOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get POS purchase order: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var posPurchaseOrdersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create POS purchase order (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create POS purchase order") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreatePOSPurchaseOrder(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create POS purchase order: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var posPurchaseOrdersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update POS purchase order (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update POS purchase order %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdatePOSPurchaseOrder(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update POS purchase order: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var posPurchaseOrdersBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Bulk delete POS purchase orders (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would bulk delete POS purchase orders") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkDeletePOSPurchaseOrders(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to bulk delete POS purchase orders: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var posPurchaseOrdersCreateChildCmd = &cobra.Command{
	Use:   "create-child <id>",
	Short: "Create child POS purchase order (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create child POS purchase order for %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreatePOSPurchaseOrderChild(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to create child POS purchase order: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(posCmd)

	posCmd.AddCommand(posPurchaseOrdersCmd)

	posPurchaseOrdersCmd.AddCommand(posPurchaseOrdersListCmd)
	posPurchaseOrdersListCmd.Flags().Int("page", 1, "Page number")
	posPurchaseOrdersListCmd.Flags().Int("page-size", 20, "Results per page")

	posPurchaseOrdersCmd.AddCommand(posPurchaseOrdersGetCmd)

	posPurchaseOrdersCmd.AddCommand(posPurchaseOrdersCreateCmd)
	addJSONBodyFlags(posPurchaseOrdersCreateCmd)

	posPurchaseOrdersCmd.AddCommand(posPurchaseOrdersUpdateCmd)
	addJSONBodyFlags(posPurchaseOrdersUpdateCmd)

	posPurchaseOrdersCmd.AddCommand(posPurchaseOrdersBulkDeleteCmd)
	addJSONBodyFlags(posPurchaseOrdersBulkDeleteCmd)

	posPurchaseOrdersCmd.AddCommand(posPurchaseOrdersCreateChildCmd)
	addJSONBodyFlags(posPurchaseOrdersCreateChildCmd)
}

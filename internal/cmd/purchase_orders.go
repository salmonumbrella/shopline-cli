package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var purchaseOrdersCmd = &cobra.Command{
	Use:   "purchase-orders",
	Short: "Manage purchase orders",
}

var purchaseOrdersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List purchase orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		supplierID, _ := cmd.Flags().GetString("supplier-id")
		warehouseID, _ := cmd.Flags().GetString("warehouse-id")

		opts := &api.PurchaseOrdersListOptions{
			Page:        page,
			PageSize:    pageSize,
			Status:      status,
			SupplierID:  supplierID,
			WarehouseID: warehouseID,
		}

		resp, err := client.ListPurchaseOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list purchase orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NUMBER", "STATUS", "SUPPLIER", "WAREHOUSE", "TOTAL", "EXPECTED", "CREATED"}
		var rows [][]string
		for _, po := range resp.Items {
			expectedAt := "-"
			if !po.ExpectedAt.IsZero() {
				expectedAt = po.ExpectedAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("purchase_order", po.ID),
				po.Number,
				po.Status,
				po.SupplierName,
				po.WarehouseName,
				po.Total,
				expectedAt,
				po.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d purchase orders\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var purchaseOrdersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get purchase order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		po, err := client.GetPurchaseOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get purchase order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(po)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Purchase Order ID: %s\n", po.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Number:            %s\n", po.Number)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:            %s\n", po.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Supplier:          %s (%s)\n", po.SupplierName, po.SupplierID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Warehouse:         %s (%s)\n", po.WarehouseName, po.WarehouseID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:          %s\n", po.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Subtotal:          %s\n", po.Subtotal)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax:               %s\n", po.Tax)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total:             %s\n", po.Total)
		if po.Note != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Note:              %s\n", po.Note)
		}
		if !po.ExpectedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Expected At:       %s\n", po.ExpectedAt.Format(time.RFC3339))
		}
		if !po.ReceivedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Received At:       %s\n", po.ReceivedAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:           %s\n", po.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:           %s\n", po.UpdatedAt.Format(time.RFC3339))

		if len(po.LineItems) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLine Items (%d):\n", len(po.LineItems))
			for _, item := range po.LineItems {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (SKU: %s)\n", item.Title, item.SKU)
				_, _ = fmt.Fprintf(outWriter(cmd), "    Quantity: %d, Received: %d, Unit Cost: %s, Total: %s\n",
					item.Quantity, item.ReceivedQty, item.UnitCost, item.Total)
			}
		}
		return nil
	},
}

var purchaseOrdersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a purchase order",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create purchase order") {
			return nil
		}

		var req api.PurchaseOrderCreateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		po, err := client.CreatePurchaseOrder(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create purchase order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(po)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created purchase order %s (status: %s)\n", po.ID, po.Status)
		return nil
	},
}

var purchaseOrdersReceiveCmd = &cobra.Command{
	Use:   "receive <id>",
	Short: "Mark a purchase order as received",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		po, err := client.ReceivePurchaseOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to receive purchase order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Marked purchase order %s as received (status: %s)\n", po.ID, po.Status)
		return nil
	},
}

var purchaseOrdersCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a purchase order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel purchase order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Cancel purchase order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		po, err := client.CancelPurchaseOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to cancel purchase order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled purchase order %s (status: %s)\n", po.ID, po.Status)
		return nil
	},
}

var purchaseOrdersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a purchase order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete purchase order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete purchase order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeletePurchaseOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete purchase order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted purchase order %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(purchaseOrdersCmd)

	purchaseOrdersCmd.AddCommand(purchaseOrdersListCmd)
	purchaseOrdersListCmd.Flags().Int("page", 1, "Page number")
	purchaseOrdersListCmd.Flags().Int("page-size", 20, "Results per page")
	purchaseOrdersListCmd.Flags().String("status", "", "Filter by status (draft, pending, received, cancelled)")
	purchaseOrdersListCmd.Flags().String("supplier-id", "", "Filter by supplier ID")
	purchaseOrdersListCmd.Flags().String("warehouse-id", "", "Filter by warehouse ID")

	purchaseOrdersCmd.AddCommand(purchaseOrdersGetCmd)
	purchaseOrdersCmd.AddCommand(purchaseOrdersCreateCmd)
	addJSONBodyFlags(purchaseOrdersCreateCmd)
	purchaseOrdersCmd.AddCommand(purchaseOrdersReceiveCmd)
	purchaseOrdersCmd.AddCommand(purchaseOrdersCancelCmd)
	purchaseOrdersCmd.AddCommand(purchaseOrdersDeleteCmd)
}

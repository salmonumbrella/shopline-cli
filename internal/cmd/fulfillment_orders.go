package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var fulfillmentOrdersCmd = &cobra.Command{
	Use:   "fulfillment-orders",
	Short: "Manage fulfillment orders",
}

var fulfillmentOrdersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fulfillment orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		orderID, _ := cmd.Flags().GetString("order-id")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := &api.FulfillmentOrdersListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			OrderID:  orderID,
		}

		resp := &api.FulfillmentOrdersListResponse{}
		if limit > 0 {
			curPage := opts.Page
			perPage := opts.PageSize
			if perPage <= 0 || perPage > limit {
				perPage = limit
			}

			items := make([]api.FulfillmentOrder, 0, limit)
			totalCount := 0
			hasMore := false

			for len(items) < limit {
				pageOpts := *opts
				pageOpts.Page = curPage
				pageOpts.PageSize = perPage

				pageResp, err := client.ListFulfillmentOrders(cmd.Context(), &pageOpts)
				if err != nil {
					return fmt.Errorf("failed to list fulfillment orders: %w", err)
				}
				if totalCount == 0 {
					totalCount = pageResp.TotalCount
				}
				items = append(items, pageResp.Items...)
				hasMore = pageResp.HasMore

				if !pageResp.HasMore || len(pageResp.Items) == 0 {
					break
				}
				curPage++
			}

			if len(items) > limit {
				items = items[:limit]
				hasMore = true
			}

			resp.Items = items
			resp.Page = opts.Page
			resp.PageSize = perPage
			resp.TotalCount = totalCount
			resp.HasMore = hasMore
		} else {
			r, err := client.ListFulfillmentOrders(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to list fulfillment orders: %w", err)
			}
			resp = r
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER ID", "STATUS", "FULFILLMENT STATUS", "LOCATION", "ITEMS", "CREATED"}
		var rows [][]string
		for _, fo := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("fulfillment_order", fo.ID),
				fo.OrderID,
				fo.Status,
				fo.FulfillmentStatus,
				fo.AssignedLocationID,
				fmt.Sprintf("%d", len(fo.LineItems)),
				fo.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d fulfillment orders\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var fulfillmentOrdersOrderCmd = &cobra.Command{
	Use:   "order <order-id>",
	Short: "List fulfillment orders for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListOrderFulfillmentOrders(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order fulfillment orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER ID", "STATUS", "FULFILLMENT STATUS", "LOCATION", "ITEMS", "CREATED"}
		var rows [][]string
		for _, fo := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("fulfillment_order", fo.ID),
				fo.OrderID,
				fo.Status,
				fo.FulfillmentStatus,
				fo.AssignedLocationID,
				fmt.Sprintf("%d", len(fo.LineItems)),
				fo.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d fulfillment orders for order %s\n", len(resp.Items), args[0])
		return nil
	},
}

var fulfillmentOrdersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get fulfillment order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		fo, err := client.GetFulfillmentOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get fulfillment order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(fo)
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Fulfillment Order ID: %s\n", fo.ID)
		_, _ = fmt.Fprintf(out, "Order ID:             %s\n", fo.OrderID)
		_, _ = fmt.Fprintf(out, "Status:               %s\n", fo.Status)
		_, _ = fmt.Fprintf(out, "Fulfillment Status:   %s\n", fo.FulfillmentStatus)
		_, _ = fmt.Fprintf(out, "Assigned Location:    %s\n", fo.AssignedLocationID)
		_, _ = fmt.Fprintf(out, "Request Status:       %s\n", fo.RequestStatus)
		_, _ = fmt.Fprintf(out, "Delivery Method:      %s (%s)\n", fo.DeliveryMethod.MethodType, fo.DeliveryMethod.ServiceCode)
		_, _ = fmt.Fprintf(out, "Created:              %s\n", fo.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(out, "Updated:              %s\n", fo.UpdatedAt.Format(time.RFC3339))

		if len(fo.LineItems) > 0 {
			_, _ = fmt.Fprintf(out, "\nLine Items (%d):\n", len(fo.LineItems))
			for _, item := range fo.LineItems {
				_, _ = fmt.Fprintf(out, "  - Variant: %s, Qty: %d (Fulfillable: %d, Fulfilled: %d)\n",
					item.VariantID, item.Quantity, item.FulfillableQuantity, item.FulfilledQuantity)
			}
		}
		return nil
	},
}

var fulfillmentOrdersMoveCmd = &cobra.Command{
	Use:   "move <id>",
	Short: "Move a fulfillment order to a new location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would move fulfillment order %s", args[0])) {
			return nil
		}

		locationID, _ := cmd.Flags().GetString("location-id")

		fo, err := client.MoveFulfillmentOrder(cmd.Context(), args[0], locationID)
		if err != nil {
			return fmt.Errorf("failed to move fulfillment order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Moved fulfillment order %s to location %s\n", fo.ID, fo.AssignedLocationID)
		return nil
	},
}

var fulfillmentOrdersCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a fulfillment order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel fulfillment order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Cancel fulfillment order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		fo, err := client.CancelFulfillmentOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to cancel fulfillment order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled fulfillment order %s (status: %s)\n", fo.ID, fo.Status)
		return nil
	},
}

var fulfillmentOrdersCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a fulfillment order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would close fulfillment order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Close fulfillment order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		fo, err := client.CloseFulfillmentOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to close fulfillment order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Closed fulfillment order %s (status: %s)\n", fo.ID, fo.Status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fulfillmentOrdersCmd)

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersListCmd)
	fulfillmentOrdersListCmd.Flags().Int("page", 1, "Page number")
	fulfillmentOrdersListCmd.Flags().Int("page-size", 20, "Results per page")
	fulfillmentOrdersListCmd.Flags().String("status", "", "Filter by status")
	fulfillmentOrdersListCmd.Flags().String("order-id", "", "Filter by order ID")

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersOrderCmd)

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersGetCmd)

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersMoveCmd)
	fulfillmentOrdersMoveCmd.Flags().String("location-id", "", "New location ID")
	_ = fulfillmentOrdersMoveCmd.MarkFlagRequired("location-id")

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersCancelCmd)
	fulfillmentOrdersCancelCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersCloseCmd)
	fulfillmentOrdersCloseCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

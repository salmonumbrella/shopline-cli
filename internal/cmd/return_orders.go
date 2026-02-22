package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var returnOrdersCmd = &cobra.Command{
	Use:   "return-orders",
	Short: "Manage return orders",
}

var returnOrdersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List return orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		orderID, _ := cmd.Flags().GetString("order-id")
		customerID, _ := cmd.Flags().GetString("customer-id")
		returnType, _ := cmd.Flags().GetString("type")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ReturnOrdersListOptions{
			Page:       page,
			PageSize:   pageSize,
			Status:     status,
			OrderID:    orderID,
			CustomerID: customerID,
			ReturnType: returnType,
		}
		if from != "" {
			since, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			opts.Since = since
		}
		if to != "" {
			until, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			opts.Until = until
		}

		resp, err := client.ListReturnOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list return orders: %w", err)
		}

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightReturnOrder)
				return formatter.JSON(api.ListResponse[lightReturnOrder]{
					Items:      lightItems,
					Pagination: resp.Pagination,
					Page:       resp.Page,
					PageSize:   resp.PageSize,
					TotalCount: resp.TotalCount,
					HasMore:    resp.HasMore,
				})
			}
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "STATUS", "TYPE", "ITEMS", "AMOUNT", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("return_order", r.ID),
				r.OrderNumber,
				r.Status,
				r.ReturnType,
				fmt.Sprintf("%d", len(r.LineItems)),
				r.TotalAmount + " " + r.Currency,
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d return orders\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var returnOrdersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get return order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		returnOrder, err := client.GetReturnOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get return order: %w", err)
		}

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightReturnOrder(returnOrder))
			}
			return formatter.JSON(returnOrder)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Return Order ID:  %s\n", returnOrder.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order ID:         %s\n", returnOrder.OrderID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order Number:     %s\n", returnOrder.OrderNumber)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:           %s\n", returnOrder.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Return Type:      %s\n", returnOrder.ReturnType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:      %s\n", returnOrder.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Email:   %s\n", returnOrder.CustomerEmail)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Amount:     %s %s\n", returnOrder.TotalAmount, returnOrder.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Refund Amount:    %s %s\n", returnOrder.RefundAmount, returnOrder.Currency)
		if returnOrder.Reason != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Reason:           %s\n", returnOrder.Reason)
		}
		if returnOrder.Note != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Note:             %s\n", returnOrder.Note)
		}
		if returnOrder.TrackingNumber != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Tracking Number:  %s\n", returnOrder.TrackingNumber)
			_, _ = fmt.Fprintf(outWriter(cmd), "Tracking Company: %s\n", returnOrder.TrackingCompany)
		}
		if returnOrder.ReceivedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Received:         %s\n", returnOrder.ReceivedAt.Format(time.RFC3339))
		}
		if returnOrder.CompletedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Completed:        %s\n", returnOrder.CompletedAt.Format(time.RFC3339))
		}
		if returnOrder.CancelledAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled:        %s\n", returnOrder.CancelledAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:          %s\n", returnOrder.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", returnOrder.UpdatedAt.Format(time.RFC3339))

		if len(returnOrder.LineItems) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLine Items (%d):\n", len(returnOrder.LineItems))
			for _, item := range returnOrder.LineItems {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s x%d (Reason: %s)\n",
					item.Title, item.Quantity, item.ReturnReason)
			}
		}
		return nil
	},
}

var returnOrdersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a return order",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create return order") {
			return nil
		}

		var req api.ReturnOrderCreateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		ret, err := client.CreateReturnOrder(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create return order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(ret)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created return order %s (status: %s)\n", ret.ID, ret.Status)
		return nil
	},
}

var returnOrdersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a return order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update return order %s", args[0])) {
			return nil
		}

		var req api.ReturnOrderUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		ret, err := client.UpdateReturnOrder(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update return order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(ret)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated return order %s (status: %s)\n", ret.ID, ret.Status)
		return nil
	},
}

var returnOrdersCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a return order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel return order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Cancel return order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.CancelReturnOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to cancel return order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Return order %s cancelled.\n", args[0])
		return nil
	},
}

var returnOrdersCompleteCmd = &cobra.Command{
	Use:   "complete <id>",
	Short: "Mark a return order as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would complete return order %s", args[0])) {
			return nil
		}

		returnOrder, err := client.CompleteReturnOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to complete return order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Return order %s completed. Status: %s\n", returnOrder.ID, returnOrder.Status)
		return nil
	},
}

var returnOrdersReceiveCmd = &cobra.Command{
	Use:   "receive <id>",
	Short: "Mark returned items as received",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		returnOrder, err := client.ReceiveReturnOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to receive return order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Return order %s marked as received. Status: %s\n", returnOrder.ID, returnOrder.Status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(returnOrdersCmd)

	returnOrdersCmd.AddCommand(returnOrdersListCmd)
	returnOrdersListCmd.Flags().String("status", "", "Filter by status (pending, received, completed, cancelled)")
	returnOrdersListCmd.Flags().String("order-id", "", "Filter by original order ID")
	returnOrdersListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	returnOrdersListCmd.Flags().String("type", "", "Filter by return type (return, exchange)")
	returnOrdersListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	returnOrdersListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	returnOrdersListCmd.Flags().Int("page", 1, "Page number")
	returnOrdersListCmd.Flags().Int("page-size", 20, "Results per page")
	returnOrdersListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(returnOrdersListCmd.Flags(), "light", "li")

	returnOrdersCmd.AddCommand(returnOrdersGetCmd)
	returnOrdersGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(returnOrdersGetCmd.Flags(), "light", "li")
	returnOrdersCmd.AddCommand(returnOrdersCreateCmd)
	addJSONBodyFlags(returnOrdersCreateCmd)
	returnOrdersCmd.AddCommand(returnOrdersUpdateCmd)
	addJSONBodyFlags(returnOrdersUpdateCmd)
	returnOrdersCmd.AddCommand(returnOrdersCancelCmd)
	returnOrdersCmd.AddCommand(returnOrdersCompleteCmd)
	returnOrdersCmd.AddCommand(returnOrdersReceiveCmd)

	schema.Register(schema.Resource{
		Name:        "return-orders",
		Description: "Manage return orders",
		Commands:    []string{"list", "get", "create", "update", "cancel", "complete", "receive"},
		IDPrefix:    "return",
	})
}

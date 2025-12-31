package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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

		resp, err := client.ListReturnOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list return orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "STATUS", "TYPE", "ITEMS", "AMOUNT", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			rows = append(rows, []string{
				r.ID,
				r.OrderNumber,
				r.Status,
				r.ReturnType,
				fmt.Sprintf("%d", len(r.LineItems)),
				r.TotalAmount + " " + r.Currency,
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d return orders\n", len(resp.Items), resp.TotalCount)
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

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(returnOrder)
		}

		fmt.Printf("Return Order ID:  %s\n", returnOrder.ID)
		fmt.Printf("Order ID:         %s\n", returnOrder.OrderID)
		fmt.Printf("Order Number:     %s\n", returnOrder.OrderNumber)
		fmt.Printf("Status:           %s\n", returnOrder.Status)
		fmt.Printf("Return Type:      %s\n", returnOrder.ReturnType)
		fmt.Printf("Customer ID:      %s\n", returnOrder.CustomerID)
		fmt.Printf("Customer Email:   %s\n", returnOrder.CustomerEmail)
		fmt.Printf("Total Amount:     %s %s\n", returnOrder.TotalAmount, returnOrder.Currency)
		fmt.Printf("Refund Amount:    %s %s\n", returnOrder.RefundAmount, returnOrder.Currency)
		if returnOrder.Reason != "" {
			fmt.Printf("Reason:           %s\n", returnOrder.Reason)
		}
		if returnOrder.Note != "" {
			fmt.Printf("Note:             %s\n", returnOrder.Note)
		}
		if returnOrder.TrackingNumber != "" {
			fmt.Printf("Tracking Number:  %s\n", returnOrder.TrackingNumber)
			fmt.Printf("Tracking Company: %s\n", returnOrder.TrackingCompany)
		}
		if returnOrder.ReceivedAt != nil {
			fmt.Printf("Received:         %s\n", returnOrder.ReceivedAt.Format(time.RFC3339))
		}
		if returnOrder.CompletedAt != nil {
			fmt.Printf("Completed:        %s\n", returnOrder.CompletedAt.Format(time.RFC3339))
		}
		if returnOrder.CancelledAt != nil {
			fmt.Printf("Cancelled:        %s\n", returnOrder.CancelledAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:          %s\n", returnOrder.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:          %s\n", returnOrder.UpdatedAt.Format(time.RFC3339))

		if len(returnOrder.LineItems) > 0 {
			fmt.Printf("\nLine Items (%d):\n", len(returnOrder.LineItems))
			for _, item := range returnOrder.LineItems {
				fmt.Printf("  - %s x%d (Reason: %s)\n",
					item.Title, item.Quantity, item.ReturnReason)
			}
		}
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

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Cancel return order %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.CancelReturnOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to cancel return order: %w", err)
		}

		fmt.Printf("Return order %s cancelled.\n", args[0])
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

		returnOrder, err := client.CompleteReturnOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to complete return order: %w", err)
		}

		fmt.Printf("Return order %s completed. Status: %s\n", returnOrder.ID, returnOrder.Status)
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

		fmt.Printf("Return order %s marked as received. Status: %s\n", returnOrder.ID, returnOrder.Status)
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
	returnOrdersListCmd.Flags().Int("page", 1, "Page number")
	returnOrdersListCmd.Flags().Int("page-size", 20, "Results per page")

	returnOrdersCmd.AddCommand(returnOrdersGetCmd)
	returnOrdersCmd.AddCommand(returnOrdersCancelCmd)
	returnOrdersCmd.AddCommand(returnOrdersCompleteCmd)
	returnOrdersCmd.AddCommand(returnOrdersReceiveCmd)
}

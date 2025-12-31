package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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

		opts := &api.FulfillmentOrdersListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			OrderID:  orderID,
		}

		resp, err := client.ListFulfillmentOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list fulfillment orders: %w", err)
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
				fo.ID,
				fo.OrderID,
				fo.Status,
				fo.FulfillmentStatus,
				fo.AssignedLocationID,
				fmt.Sprintf("%d", len(fo.LineItems)),
				fo.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d fulfillment orders\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Fulfillment Order ID: %s\n", fo.ID)
		fmt.Printf("Order ID:             %s\n", fo.OrderID)
		fmt.Printf("Status:               %s\n", fo.Status)
		fmt.Printf("Fulfillment Status:   %s\n", fo.FulfillmentStatus)
		fmt.Printf("Assigned Location:    %s\n", fo.AssignedLocationID)
		fmt.Printf("Request Status:       %s\n", fo.RequestStatus)
		fmt.Printf("Delivery Method:      %s (%s)\n", fo.DeliveryMethod.MethodType, fo.DeliveryMethod.ServiceCode)
		fmt.Printf("Created:              %s\n", fo.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:              %s\n", fo.UpdatedAt.Format(time.RFC3339))

		if len(fo.LineItems) > 0 {
			fmt.Printf("\nLine Items (%d):\n", len(fo.LineItems))
			for _, item := range fo.LineItems {
				fmt.Printf("  - Variant: %s, Qty: %d (Fulfillable: %d, Fulfilled: %d)\n",
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

		locationID, _ := cmd.Flags().GetString("location-id")

		fo, err := client.MoveFulfillmentOrder(cmd.Context(), args[0], locationID)
		if err != nil {
			return fmt.Errorf("failed to move fulfillment order: %w", err)
		}

		fmt.Printf("Moved fulfillment order %s to location %s\n", fo.ID, fo.AssignedLocationID)
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

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Cancel fulfillment order %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		fo, err := client.CancelFulfillmentOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to cancel fulfillment order: %w", err)
		}

		fmt.Printf("Cancelled fulfillment order %s (status: %s)\n", fo.ID, fo.Status)
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

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Close fulfillment order %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		fo, err := client.CloseFulfillmentOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to close fulfillment order: %w", err)
		}

		fmt.Printf("Closed fulfillment order %s (status: %s)\n", fo.ID, fo.Status)
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

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersGetCmd)

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersMoveCmd)
	fulfillmentOrdersMoveCmd.Flags().String("location-id", "", "New location ID")
	_ = fulfillmentOrdersMoveCmd.MarkFlagRequired("location-id")

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersCancelCmd)
	fulfillmentOrdersCancelCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	fulfillmentOrdersCmd.AddCommand(fulfillmentOrdersCloseCmd)
	fulfillmentOrdersCloseCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

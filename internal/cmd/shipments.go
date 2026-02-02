package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var shipmentsCmd = &cobra.Command{
	Use:   "shipments",
	Short: "Manage shipments",
}

var shipmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shipments",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		orderID, _ := cmd.Flags().GetString("order-id")
		fulfillmentID, _ := cmd.Flags().GetString("fulfillment-id")
		status, _ := cmd.Flags().GetString("status")
		trackingNumber, _ := cmd.Flags().GetString("tracking-number")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ShipmentsListOptions{
			Page:           page,
			PageSize:       pageSize,
			OrderID:        orderID,
			FulfillmentID:  fulfillmentID,
			Status:         status,
			TrackingNumber: trackingNumber,
		}

		resp, err := client.ListShipments(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list shipments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "FULFILLMENT", "CARRIER", "TRACKING", "STATUS", "CREATED"}
		var rows [][]string
		for _, s := range resp.Items {
			rows = append(rows, []string{
				s.ID,
				s.OrderID,
				s.FulfillmentID,
				s.TrackingCompany,
				s.TrackingNumber,
				s.Status,
				s.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d shipments\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var shipmentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get shipment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		shipment, err := client.GetShipment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get shipment: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(shipment)
		}

		fmt.Printf("Shipment ID:        %s\n", shipment.ID)
		fmt.Printf("Order ID:           %s\n", shipment.OrderID)
		fmt.Printf("Fulfillment ID:     %s\n", shipment.FulfillmentID)
		fmt.Printf("Tracking Company:   %s\n", shipment.TrackingCompany)
		fmt.Printf("Tracking Number:    %s\n", shipment.TrackingNumber)
		if shipment.TrackingURL != "" {
			fmt.Printf("Tracking URL:       %s\n", shipment.TrackingURL)
		}
		fmt.Printf("Status:             %s\n", shipment.Status)
		if !shipment.EstimatedDelivery.IsZero() {
			fmt.Printf("Estimated Delivery: %s\n", shipment.EstimatedDelivery.Format(time.RFC3339))
		}
		if !shipment.DeliveredAt.IsZero() {
			fmt.Printf("Delivered At:       %s\n", shipment.DeliveredAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:            %s\n", shipment.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:            %s\n", shipment.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var shipmentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a shipment",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		orderID, _ := cmd.Flags().GetString("order-id")
		fulfillmentID, _ := cmd.Flags().GetString("fulfillment-id")
		trackingCompany, _ := cmd.Flags().GetString("tracking-company")
		trackingNumber, _ := cmd.Flags().GetString("tracking-number")
		trackingURL, _ := cmd.Flags().GetString("tracking-url")

		req := &api.ShipmentCreateRequest{
			OrderID:         orderID,
			FulfillmentID:   fulfillmentID,
			TrackingCompany: trackingCompany,
			TrackingNumber:  trackingNumber,
			TrackingURL:     trackingURL,
		}

		shipment, err := client.CreateShipment(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create shipment: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(shipment)
		}

		fmt.Printf("Created shipment %s\n", shipment.ID)
		fmt.Printf("Order ID:         %s\n", shipment.OrderID)
		fmt.Printf("Fulfillment ID:   %s\n", shipment.FulfillmentID)
		fmt.Printf("Tracking Company: %s\n", shipment.TrackingCompany)
		fmt.Printf("Tracking Number:  %s\n", shipment.TrackingNumber)
		return nil
	},
}

var shipmentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a shipment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		trackingCompany, _ := cmd.Flags().GetString("tracking-company")
		trackingNumber, _ := cmd.Flags().GetString("tracking-number")
		trackingURL, _ := cmd.Flags().GetString("tracking-url")
		status, _ := cmd.Flags().GetString("status")

		req := &api.ShipmentUpdateRequest{
			TrackingCompany: trackingCompany,
			TrackingNumber:  trackingNumber,
			TrackingURL:     trackingURL,
			Status:          status,
		}

		shipment, err := client.UpdateShipment(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update shipment: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(shipment)
		}

		fmt.Printf("Updated shipment %s\n", shipment.ID)
		fmt.Printf("Tracking Company: %s\n", shipment.TrackingCompany)
		fmt.Printf("Tracking Number:  %s\n", shipment.TrackingNumber)
		fmt.Printf("Status:           %s\n", shipment.Status)
		return nil
	},
}

var shipmentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a shipment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete shipment %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteShipment(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete shipment: %w", err)
		}

		fmt.Printf("Deleted shipment %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shipmentsCmd)

	shipmentsCmd.AddCommand(shipmentsListCmd)
	shipmentsListCmd.Flags().String("order-id", "", "Filter by order ID")
	shipmentsListCmd.Flags().String("fulfillment-id", "", "Filter by fulfillment ID")
	shipmentsListCmd.Flags().String("status", "", "Filter by status")
	shipmentsListCmd.Flags().String("tracking-number", "", "Filter by tracking number")
	shipmentsListCmd.Flags().Int("page", 1, "Page number")
	shipmentsListCmd.Flags().Int("page-size", 20, "Results per page")

	shipmentsCmd.AddCommand(shipmentsGetCmd)

	shipmentsCmd.AddCommand(shipmentsCreateCmd)
	shipmentsCreateCmd.Flags().String("order-id", "", "Order ID")
	shipmentsCreateCmd.Flags().String("fulfillment-id", "", "Fulfillment ID")
	shipmentsCreateCmd.Flags().String("tracking-company", "", "Tracking company/carrier name")
	shipmentsCreateCmd.Flags().String("tracking-number", "", "Tracking number")
	shipmentsCreateCmd.Flags().String("tracking-url", "", "Tracking URL (optional)")
	_ = shipmentsCreateCmd.MarkFlagRequired("order-id")
	_ = shipmentsCreateCmd.MarkFlagRequired("fulfillment-id")
	_ = shipmentsCreateCmd.MarkFlagRequired("tracking-company")
	_ = shipmentsCreateCmd.MarkFlagRequired("tracking-number")

	shipmentsCmd.AddCommand(shipmentsUpdateCmd)
	shipmentsUpdateCmd.Flags().String("tracking-company", "", "Tracking company/carrier name")
	shipmentsUpdateCmd.Flags().String("tracking-number", "", "Tracking number")
	shipmentsUpdateCmd.Flags().String("tracking-url", "", "Tracking URL")
	shipmentsUpdateCmd.Flags().String("status", "", "Shipment status")

	shipmentsCmd.AddCommand(shipmentsDeleteCmd)
	shipmentsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	schema.Register(schema.Resource{
		Name:        "shipments",
		Description: "Manage shipments",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "shipment",
	})
}

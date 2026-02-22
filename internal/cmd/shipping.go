package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var shippingCmd = &cobra.Command{
	Use:     "shipping",
	Aliases: []string{"ship"},
	Short:   "Manage order shipments, tracking, and labels (via Admin API)",
}

var shippingStatusCmd = &cobra.Command{
	Use:   "status <order-id>",
	Short: "Check if a shipment has been executed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.GetShipmentStatus(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get shipment status: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var shippingTrackingCmd = &cobra.Command{
	Use:   "tracking <order-id>",
	Short: "Get tracking number for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.GetTrackingNumber(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get tracking number: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var shippingExecuteCmd = &cobra.Command{
	Use:   "execute <order-id>",
	Short: "Execute shipment for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would execute shipment for order %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		orderNumber, _ := cmd.Flags().GetString("order-number")
		performerID, _ := cmd.Flags().GetString("performer-id")

		req := &api.AdminExecuteShipmentRequest{
			OrderNumber: orderNumber,
			PerformerID: performerID,
		}

		result, err := client.ExecuteShipment(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to execute shipment: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var shippingPrintLabelCmd = &cobra.Command{
	Use:     "print-label <order-id>",
	Aliases: []string{"label", "print"},
	Short:   "Generate and retrieve packing label",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		upsert, _ := cmd.Flags().GetBool("upsert")

		req := &api.AdminPrintLabelRequest{
			Upsert: upsert,
		}

		result, err := client.PrintPackingLabel(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to print label: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(shippingCmd)
	shippingCmd.AddCommand(shippingStatusCmd)
	shippingCmd.AddCommand(shippingTrackingCmd)
	shippingCmd.AddCommand(shippingExecuteCmd)
	shippingCmd.AddCommand(shippingPrintLabelCmd)

	shippingExecuteCmd.Flags().String("order-number", "", "Order number (e.g., W-12345) (required)")
	shippingExecuteCmd.Flags().String("performer-id", "", "ID of person executing shipment (required)")
	_ = shippingExecuteCmd.MarkFlagRequired("order-number")
	_ = shippingExecuteCmd.MarkFlagRequired("performer-id")
	shippingExecuteCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	shippingPrintLabelCmd.Flags().Bool("upsert", false, "Force re-execution even if already executed")

	schema.Register(schema.Resource{
		Name:        "shipping",
		Description: "Manage order shipments, tracking, and labels (via Admin API)",
		Commands:    []string{"status", "tracking", "execute", "print-label"},
	})
}

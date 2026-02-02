package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var fulfillmentsCmd = &cobra.Command{
	Use:   "fulfillments",
	Short: "Manage fulfillments",
}

var fulfillmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fulfillments",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		orderID, _ := cmd.Flags().GetString("order-id")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.FulfillmentsListOptions{
			Page:     page,
			PageSize: pageSize,
			OrderID:  orderID,
			Status:   status,
		}

		resp, err := client.ListFulfillments(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list fulfillments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "STATUS", "CARRIER", "TRACKING", "CREATED"}
		var rows [][]string
		for _, f := range resp.Items {
			rows = append(rows, []string{
				f.ID,
				f.OrderID,
				string(f.Status),
				f.TrackingCompany,
				f.TrackingNumber,
				f.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d fulfillments\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var fulfillmentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get fulfillment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		fulfillment, err := client.GetFulfillment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get fulfillment: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(fulfillment)
		}

		fmt.Printf("Fulfillment ID:    %s\n", fulfillment.ID)
		fmt.Printf("Order ID:          %s\n", fulfillment.OrderID)
		fmt.Printf("Status:            %s\n", fulfillment.Status)
		fmt.Printf("Tracking Company:  %s\n", fulfillment.TrackingCompany)
		fmt.Printf("Tracking Number:   %s\n", fulfillment.TrackingNumber)
		if fulfillment.TrackingURL != "" {
			fmt.Printf("Tracking URL:      %s\n", fulfillment.TrackingURL)
		}
		fmt.Printf("Created:           %s\n", fulfillment.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:           %s\n", fulfillment.UpdatedAt.Format(time.RFC3339))

		if len(fulfillment.LineItems) > 0 {
			fmt.Printf("\nLine Items:\n")
			for _, item := range fulfillment.LineItems {
				fmt.Printf("  - %s (qty: %d, SKU: %s)\n", item.Title, item.Quantity, item.SKU)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(fulfillmentsCmd)

	fulfillmentsCmd.AddCommand(fulfillmentsListCmd)
	fulfillmentsListCmd.Flags().String("order-id", "", "Filter by order ID")
	fulfillmentsListCmd.Flags().String("status", "", "Filter by status (pending/open/success/cancelled/failure)")
	fulfillmentsListCmd.Flags().Int("page", 1, "Page number")
	fulfillmentsListCmd.Flags().Int("page-size", 20, "Results per page")

	fulfillmentsCmd.AddCommand(fulfillmentsGetCmd)

	schema.Register(schema.Resource{
		Name:        "fulfillments",
		Description: "Manage order fulfillments",
		Commands:    []string{"list", "get"},
		IDPrefix:    "fulfillment",
	})
}

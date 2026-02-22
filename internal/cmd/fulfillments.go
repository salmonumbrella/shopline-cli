package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
				outfmt.FormatID("fulfillment", f.ID),
				f.OrderID,
				string(f.Status),
				f.TrackingCompany,
				f.TrackingNumber,
				f.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d fulfillments\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Fulfillment ID:    %s\n", fulfillment.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order ID:          %s\n", fulfillment.OrderID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:            %s\n", fulfillment.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tracking Company:  %s\n", fulfillment.TrackingCompany)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tracking Number:   %s\n", fulfillment.TrackingNumber)
		if fulfillment.TrackingURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Tracking URL:      %s\n", fulfillment.TrackingURL)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:           %s\n", fulfillment.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:           %s\n", fulfillment.UpdatedAt.Format(time.RFC3339))

		if len(fulfillment.LineItems) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLine Items:\n")
			for _, item := range fulfillment.LineItems {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (qty: %d, SKU: %s)\n", item.Title, item.Quantity, item.SKU)
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

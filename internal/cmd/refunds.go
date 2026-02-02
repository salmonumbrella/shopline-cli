package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var refundsCmd = &cobra.Command{
	Use:   "refunds",
	Short: "Manage order refunds",
}

var refundsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List refunds",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.RefundsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListRefunds(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list refunds: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "STATUS", "AMOUNT", "CURRENCY", "NOTE", "ITEMS", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			note := r.Note
			if len(note) > 20 {
				note = note[:17] + "..."
			}
			rows = append(rows, []string{
				r.ID,
				r.OrderID,
				r.Status,
				r.Amount,
				r.Currency,
				note,
				fmt.Sprintf("%d", len(r.LineItems)),
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d refunds\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var refundsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get refund details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		refund, err := client.GetRefund(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get refund: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(refund)
		}

		fmt.Printf("Refund ID:      %s\n", refund.ID)
		fmt.Printf("Order ID:       %s\n", refund.OrderID)
		fmt.Printf("Status:         %s\n", refund.Status)
		fmt.Printf("Amount:         %s %s\n", refund.Amount, refund.Currency)
		if refund.Note != "" {
			fmt.Printf("Note:           %s\n", refund.Note)
		}
		fmt.Printf("Restock:        %t\n", refund.Restock)
		if !refund.ProcessedAt.IsZero() {
			fmt.Printf("Processed At:   %s\n", refund.ProcessedAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", refund.CreatedAt.Format(time.RFC3339))

		if len(refund.LineItems) > 0 {
			fmt.Printf("\nLine Items (%d):\n", len(refund.LineItems))
			for _, item := range refund.LineItems {
				fmt.Printf("  - Line Item: %s, Qty: %d, Subtotal: %.2f\n",
					item.LineItemID, item.Quantity, item.Subtotal)
			}
		}
		return nil
	},
}

var refundsOrderCmd = &cobra.Command{
	Use:   "order <order-id>",
	Short: "List refunds for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListOrderRefunds(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order refunds: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "STATUS", "AMOUNT", "CURRENCY", "NOTE", "ITEMS", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			note := r.Note
			if len(note) > 20 {
				note = note[:17] + "..."
			}
			rows = append(rows, []string{
				r.ID,
				r.Status,
				r.Amount,
				r.Currency,
				note,
				fmt.Sprintf("%d", len(r.LineItems)),
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d refunds for order %s\n", len(resp.Items), args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(refundsCmd)

	refundsCmd.AddCommand(refundsListCmd)
	refundsListCmd.Flags().Int("page", 1, "Page number")
	refundsListCmd.Flags().Int("page-size", 20, "Results per page")

	refundsCmd.AddCommand(refundsGetCmd)
	refundsCmd.AddCommand(refundsOrderCmd)

	schema.Register(schema.Resource{
		Name:        "refunds",
		Description: "Manage order refunds",
		Commands:    []string{"list", "get", "order"},
		IDPrefix:    "refund",
	})
}

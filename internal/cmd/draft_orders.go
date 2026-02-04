package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var draftOrdersCmd = &cobra.Command{
	Use:   "draft-orders",
	Short: "Manage draft orders",
}

var draftOrdersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List draft orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		customerID, _ := cmd.Flags().GetString("customer-id")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.DraftOrdersListOptions{
			Page:       page,
			PageSize:   pageSize,
			Status:     status,
			CustomerID: customerID,
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

		resp, err := client.ListDraftOrders(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list draft orders: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "STATUS", "CUSTOMER", "TOTAL", "CREATED"}
		var rows [][]string
		for _, d := range resp.Items {
			rows = append(rows, []string{
				d.ID,
				d.Name,
				d.Status,
				d.CustomerEmail,
				d.TotalPrice + " " + d.Currency,
				d.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d draft orders\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var draftOrdersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get draft order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		draftOrder, err := client.GetDraftOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get draft order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(draftOrder)
		}

		fmt.Printf("Draft Order ID:  %s\n", draftOrder.ID)
		fmt.Printf("Name:            %s\n", draftOrder.Name)
		fmt.Printf("Status:          %s\n", draftOrder.Status)
		fmt.Printf("Customer ID:     %s\n", draftOrder.CustomerID)
		fmt.Printf("Customer Email:  %s\n", draftOrder.CustomerEmail)
		fmt.Printf("Total:           %s %s\n", draftOrder.TotalPrice, draftOrder.Currency)
		fmt.Printf("Subtotal:        %s %s\n", draftOrder.SubtotalPrice, draftOrder.Currency)
		fmt.Printf("Tax:             %s %s\n", draftOrder.TotalTax, draftOrder.Currency)
		if draftOrder.Note != "" {
			fmt.Printf("Note:            %s\n", draftOrder.Note)
		}
		if draftOrder.InvoiceURL != "" {
			fmt.Printf("Invoice URL:     %s\n", draftOrder.InvoiceURL)
		}
		if draftOrder.InvoiceSentAt != nil {
			fmt.Printf("Invoice Sent:    %s\n", draftOrder.InvoiceSentAt.Format(time.RFC3339))
		}
		if draftOrder.CompletedAt != nil {
			fmt.Printf("Completed:       %s\n", draftOrder.CompletedAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:         %s\n", draftOrder.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:         %s\n", draftOrder.UpdatedAt.Format(time.RFC3339))

		if len(draftOrder.LineItems) > 0 {
			fmt.Printf("\nLine Items (%d):\n", len(draftOrder.LineItems))
			for _, item := range draftOrder.LineItems {
				fmt.Printf("  - %s (Variant: %s) x%d @ %.2f\n",
					item.Title, item.VariantID, item.Quantity, item.Price)
			}
		}
		return nil
	},
}

var draftOrdersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a draft order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete draft order %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteDraftOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete draft order: %w", err)
		}

		fmt.Printf("Draft order %s deleted.\n", args[0])
		return nil
	},
}

var draftOrdersCompleteCmd = &cobra.Command{
	Use:   "complete <id>",
	Short: "Complete a draft order (convert to order)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Complete draft order %s? This will create a real order. [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		draftOrder, err := client.CompleteDraftOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to complete draft order: %w", err)
		}

		fmt.Printf("Draft order %s completed. Status: %s\n", draftOrder.ID, draftOrder.Status)
		return nil
	},
}

var draftOrdersSendInvoiceCmd = &cobra.Command{
	Use:   "send-invoice <id>",
	Short: "Send invoice for a draft order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.SendDraftOrderInvoice(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to send invoice: %w", err)
		}

		fmt.Printf("Invoice sent for draft order %s.\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(draftOrdersCmd)

	draftOrdersCmd.AddCommand(draftOrdersListCmd)
	draftOrdersListCmd.Flags().String("status", "", "Filter by status (open, invoice_sent, completed)")
	draftOrdersListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	draftOrdersListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	draftOrdersListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	draftOrdersListCmd.Flags().Int("page", 1, "Page number")
	draftOrdersListCmd.Flags().Int("page-size", 20, "Results per page")

	draftOrdersCmd.AddCommand(draftOrdersGetCmd)
	draftOrdersCmd.AddCommand(draftOrdersDeleteCmd)
	draftOrdersCmd.AddCommand(draftOrdersCompleteCmd)
	draftOrdersCmd.AddCommand(draftOrdersSendInvoiceCmd)

	schema.Register(schema.Resource{
		Name:        "draft-orders",
		Description: "Manage draft orders",
		Commands:    []string{"list", "get", "delete", "complete", "send-invoice"},
		IDPrefix:    "draft_order",
	})
}

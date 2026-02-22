package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightDraftOrder)
				return formatter.JSON(api.ListResponse[lightDraftOrder]{
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

		headers := []string{"ID", "NAME", "STATUS", "CUSTOMER", "TOTAL", "CREATED"}
		var rows [][]string
		for _, d := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("draft_order", d.ID),
				d.Name,
				d.Status,
				d.CustomerEmail,
				d.TotalPrice + " " + d.Currency,
				d.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d draft orders\n", len(resp.Items), resp.TotalCount)
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

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightDraftOrder(draftOrder))
			}
			return formatter.JSON(draftOrder)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Draft Order ID:  %s\n", draftOrder.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", draftOrder.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", draftOrder.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:     %s\n", draftOrder.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Email:  %s\n", draftOrder.CustomerEmail)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total:           %s %s\n", draftOrder.TotalPrice, draftOrder.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Subtotal:        %s %s\n", draftOrder.SubtotalPrice, draftOrder.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax:             %s %s\n", draftOrder.TotalTax, draftOrder.Currency)
		if draftOrder.Note != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Note:            %s\n", draftOrder.Note)
		}
		if draftOrder.InvoiceURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Invoice URL:     %s\n", draftOrder.InvoiceURL)
		}
		if draftOrder.InvoiceSentAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Invoice Sent:    %s\n", draftOrder.InvoiceSentAt.Format(time.RFC3339))
		}
		if draftOrder.CompletedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Completed:       %s\n", draftOrder.CompletedAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", draftOrder.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", draftOrder.UpdatedAt.Format(time.RFC3339))

		if len(draftOrder.LineItems) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nLine Items (%d):\n", len(draftOrder.LineItems))
			for _, item := range draftOrder.LineItems {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (Variant: %s) x%d @ %.2f\n",
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete draft order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete draft order %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteDraftOrder(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete draft order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Draft order %s deleted.\n", args[0])
		return nil
	},
}

var draftOrdersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a draft order",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create draft order") {
			return nil
		}

		var req api.DraftOrderCreateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		draftOrder, err := client.CreateDraftOrder(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create draft order: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(draftOrder)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created draft order %s (status: %s)\n", draftOrder.ID, draftOrder.Status)
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would complete draft order %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Complete draft order %s? This will create a real order. [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		draftOrder, err := client.CompleteDraftOrder(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to complete draft order: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Draft order %s completed. Status: %s\n", draftOrder.ID, draftOrder.Status)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Invoice sent for draft order %s.\n", args[0])
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
	draftOrdersListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(draftOrdersListCmd.Flags(), "light", "li")

	draftOrdersCmd.AddCommand(draftOrdersGetCmd)
	draftOrdersGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(draftOrdersGetCmd.Flags(), "light", "li")

	draftOrdersCmd.AddCommand(draftOrdersCreateCmd)
	addJSONBodyFlags(draftOrdersCreateCmd)

	draftOrdersCmd.AddCommand(draftOrdersDeleteCmd)
	draftOrdersCmd.AddCommand(draftOrdersCompleteCmd)
	draftOrdersCmd.AddCommand(draftOrdersSendInvoiceCmd)

	schema.Register(schema.Resource{
		Name:        "draft-orders",
		Description: "Manage draft orders",
		Commands:    []string{"list", "get", "create", "delete", "complete", "send-invoice"},
		IDPrefix:    "draft_order",
	})
}

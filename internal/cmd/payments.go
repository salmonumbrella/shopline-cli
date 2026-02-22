package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var paymentsCmd = &cobra.Command{
	Use:   "payments",
	Short: "Manage payments",
}

var paymentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List payments",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		gateway, _ := cmd.Flags().GetString("gateway")

		opts := &api.PaymentsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Gateway:  gateway,
		}

		resp, err := client.ListPayments(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list payments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "AMOUNT", "CURRENCY", "STATUS", "GATEWAY", "METHOD", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("payment", p.ID),
				p.OrderID,
				p.Amount,
				p.Currency,
				p.Status,
				p.Gateway,
				p.PaymentMethod,
				p.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d payments\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var paymentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get payment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		payment, err := client.GetPayment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get payment: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(payment)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Payment ID:     %s\n", payment.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order ID:       %s\n", payment.OrderID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Amount:         %s %s\n", payment.Amount, payment.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", payment.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Gateway:        %s\n", payment.Gateway)
		_, _ = fmt.Fprintf(outWriter(cmd), "Payment Method: %s\n", payment.PaymentMethod)
		if payment.TransactionID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Transaction ID: %s\n", payment.TransactionID)
		}
		if payment.ErrorMessage != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Error:          %s\n", payment.ErrorMessage)
		}
		if payment.CreditCard != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Card:           %s ****%s (%02d/%d)\n",
				payment.CreditCard.Brand,
				payment.CreditCard.Last4,
				payment.CreditCard.ExpiryMonth,
				payment.CreditCard.ExpiryYear)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", payment.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", payment.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var paymentsOrderCmd = &cobra.Command{
	Use:   "order <order-id>",
	Short: "List payments for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListOrderPayments(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order payments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "AMOUNT", "CURRENCY", "STATUS", "GATEWAY", "METHOD", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("payment", p.ID),
				p.Amount,
				p.Currency,
				p.Status,
				p.Gateway,
				p.PaymentMethod,
				p.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d payments for order %s\n", len(resp.Items), args[0])
		return nil
	},
}

var paymentsAccountSummaryCmd = &cobra.Command{
	Use:   "account-summary",
	Short: "Get payments account summary (via Admin API)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.GetPaymentsAccountSummary(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get payments account summary: %w", err)
		}
		return getFormatter(cmd).JSON(result)
	},
}

var paymentsPayoutsCmd = &cobra.Command{
	Use:   "payouts",
	Short: "List payment payouts (via Admin API)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		from, _ := cmd.Flags().GetInt64("from")
		if from <= 0 {
			return fmt.Errorf("--from must be a positive Unix timestamp in milliseconds")
		}

		result, err := client.GetPaymentsPayouts(cmd.Context(), &api.AdminPaymentsPayoutsOptions{From: from})
		if err != nil {
			return fmt.Errorf("failed to get payment payouts: %w", err)
		}
		return getFormatter(cmd).JSON(result)
	},
}

var paymentsCaptureCmd = &cobra.Command{
	Use:   "capture <id>",
	Short: "Capture an authorized payment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		amount, _ := cmd.Flags().GetString("amount")

		payment, err := client.CapturePayment(cmd.Context(), args[0], amount)
		if err != nil {
			return fmt.Errorf("failed to capture payment: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Captured payment %s (status: %s)\n", payment.ID, payment.Status)
		return nil
	},
}

var paymentsVoidCmd = &cobra.Command{
	Use:   "void <id>",
	Short: "Void an authorized payment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Void payment %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		payment, err := client.VoidPayment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to void payment: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Voided payment %s (status: %s)\n", payment.ID, payment.Status)
		return nil
	},
}

var paymentsRefundCmd = &cobra.Command{
	Use:   "refund <id>",
	Short: "Refund a captured payment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		amount, _ := cmd.Flags().GetString("amount")
		reason, _ := cmd.Flags().GetString("reason")

		if !confirmAction(cmd, fmt.Sprintf("Refund payment %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		payment, err := client.RefundPayment(cmd.Context(), args[0], amount, reason)
		if err != nil {
			return fmt.Errorf("failed to refund payment: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Refunded payment %s (status: %s)\n", payment.ID, payment.Status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(paymentsCmd)

	paymentsCmd.AddCommand(paymentsListCmd)
	paymentsListCmd.Flags().Int("page", 1, "Page number")
	paymentsListCmd.Flags().Int("page-size", 20, "Results per page")
	paymentsListCmd.Flags().String("status", "", "Filter by status (authorized, captured, refunded, voided)")
	paymentsListCmd.Flags().String("gateway", "", "Filter by payment gateway")

	paymentsCmd.AddCommand(paymentsGetCmd)
	paymentsCmd.AddCommand(paymentsOrderCmd)
	paymentsCmd.AddCommand(paymentsAccountSummaryCmd)
	paymentsCmd.AddCommand(paymentsPayoutsCmd)
	paymentsPayoutsCmd.Flags().Int64("from", 0, "Unix timestamp in milliseconds (required)")
	_ = paymentsPayoutsCmd.MarkFlagRequired("from")

	paymentsCmd.AddCommand(paymentsCaptureCmd)
	paymentsCaptureCmd.Flags().String("amount", "", "Amount to capture (defaults to full amount)")

	paymentsCmd.AddCommand(paymentsVoidCmd)

	paymentsCmd.AddCommand(paymentsRefundCmd)
	paymentsRefundCmd.Flags().String("amount", "", "Amount to refund (defaults to full amount)")
	paymentsRefundCmd.Flags().String("reason", "", "Reason for refund")

	schema.Register(schema.Resource{
		Name:        "payments",
		Description: "Manage payments",
		Commands:    []string{"list", "get", "order", "account-summary", "payouts", "capture", "void", "refund"},
		IDPrefix:    "payment",
	})
}

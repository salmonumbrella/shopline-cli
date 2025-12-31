package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Manage payment transactions",
}

var transactionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		kind, _ := cmd.Flags().GetString("kind")

		opts := &api.TransactionsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Kind:     kind,
		}

		resp, err := client.ListTransactions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list transactions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "KIND", "STATUS", "AMOUNT", "CURRENCY", "GATEWAY", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			rows = append(rows, []string{
				t.ID,
				t.OrderID,
				t.Kind,
				t.Status,
				t.Amount,
				t.Currency,
				t.Gateway,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d transactions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var transactionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get transaction details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		transaction, err := client.GetTransaction(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(transaction)
		}

		fmt.Printf("Transaction ID:    %s\n", transaction.ID)
		fmt.Printf("Order ID:          %s\n", transaction.OrderID)
		fmt.Printf("Kind:              %s\n", transaction.Kind)
		fmt.Printf("Status:            %s\n", transaction.Status)
		fmt.Printf("Amount:            %s %s\n", transaction.Amount, transaction.Currency)
		fmt.Printf("Gateway:           %s\n", transaction.Gateway)
		if transaction.ErrorCode != "" {
			fmt.Printf("Error Code:        %s\n", transaction.ErrorCode)
		}
		if transaction.Message != "" {
			fmt.Printf("Message:           %s\n", transaction.Message)
		}
		fmt.Printf("Created:           %s\n", transaction.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var transactionsOrderCmd = &cobra.Command{
	Use:   "order <order-id>",
	Short: "List transactions for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListOrderTransactions(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list order transactions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "KIND", "STATUS", "AMOUNT", "CURRENCY", "GATEWAY", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			rows = append(rows, []string{
				t.ID,
				t.Kind,
				t.Status,
				t.Amount,
				t.Currency,
				t.Gateway,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d transactions for order %s\n", len(resp.Items), args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(transactionsCmd)

	transactionsCmd.AddCommand(transactionsListCmd)
	transactionsListCmd.Flags().Int("page", 1, "Page number")
	transactionsListCmd.Flags().Int("page-size", 20, "Results per page")
	transactionsListCmd.Flags().String("status", "", "Filter by status (success, failure, pending)")
	transactionsListCmd.Flags().String("kind", "", "Filter by kind (sale, refund, capture, void)")

	transactionsCmd.AddCommand(transactionsGetCmd)
	transactionsCmd.AddCommand(transactionsOrderCmd)
}

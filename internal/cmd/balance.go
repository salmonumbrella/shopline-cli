package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Manage account balance",
}

var balanceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current account balance",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		balance, err := client.GetBalance(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(balance)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:   %s\n", balance.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:  %s\n", balance.Available)
		_, _ = fmt.Fprintf(outWriter(cmd), "Pending:    %s\n", balance.Pending)
		if balance.Reserved != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Reserved:   %s\n", balance.Reserved)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Total:      %s\n", balance.Total)
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:    %s\n", balance.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var balanceTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "List balance transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		txnType, _ := cmd.Flags().GetString("type")
		sourceType, _ := cmd.Flags().GetString("source-type")

		opts := &api.BalanceTransactionsListOptions{
			Page:       page,
			PageSize:   pageSize,
			Type:       txnType,
			SourceType: sourceType,
		}

		resp, err := client.ListBalanceTransactions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list balance transactions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TYPE", "AMOUNT", "CURRENCY", "NET", "STATUS", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("transaction", t.ID),
				t.Type,
				t.Amount,
				t.Currency,
				t.Net,
				t.Status,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d transactions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var balanceTransactionGetCmd = &cobra.Command{
	Use:   "transaction <id>",
	Short: "Get balance transaction details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		txn, err := client.GetBalanceTransaction(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get balance transaction: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(txn)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Transaction ID: %s\n", txn.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:           %s\n", txn.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Amount:         %s %s\n", txn.Amount, txn.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Net:            %s\n", txn.Net)
		if txn.Fee != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Fee:            %s\n", txn.Fee)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", txn.Status)
		if txn.Description != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", txn.Description)
		}
		if txn.SourceID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Source ID:      %s\n", txn.SourceID)
			_, _ = fmt.Fprintf(outWriter(cmd), "Source Type:    %s\n", txn.SourceType)
		}
		if txn.AvailableOn != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Available On:   %s\n", txn.AvailableOn.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", txn.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)

	balanceCmd.AddCommand(balanceGetCmd)

	balanceCmd.AddCommand(balanceTransactionsCmd)
	balanceTransactionsCmd.Flags().Int("page", 1, "Page number")
	balanceTransactionsCmd.Flags().Int("page-size", 20, "Results per page")
	balanceTransactionsCmd.Flags().String("type", "", "Filter by type (payment, refund, payout, adjustment)")
	balanceTransactionsCmd.Flags().String("source-type", "", "Filter by source type (order, payout, refund)")

	balanceCmd.AddCommand(balanceTransactionGetCmd)
}

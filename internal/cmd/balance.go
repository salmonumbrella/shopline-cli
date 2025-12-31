package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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

		fmt.Printf("Currency:   %s\n", balance.Currency)
		fmt.Printf("Available:  %s\n", balance.Available)
		fmt.Printf("Pending:    %s\n", balance.Pending)
		if balance.Reserved != "" {
			fmt.Printf("Reserved:   %s\n", balance.Reserved)
		}
		fmt.Printf("Total:      %s\n", balance.Total)
		fmt.Printf("Updated:    %s\n", balance.UpdatedAt.Format(time.RFC3339))
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
				t.ID,
				t.Type,
				t.Amount,
				t.Currency,
				t.Net,
				t.Status,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d transactions\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Transaction ID: %s\n", txn.ID)
		fmt.Printf("Type:           %s\n", txn.Type)
		fmt.Printf("Amount:         %s %s\n", txn.Amount, txn.Currency)
		fmt.Printf("Net:            %s\n", txn.Net)
		if txn.Fee != "" {
			fmt.Printf("Fee:            %s\n", txn.Fee)
		}
		fmt.Printf("Status:         %s\n", txn.Status)
		if txn.Description != "" {
			fmt.Printf("Description:    %s\n", txn.Description)
		}
		if txn.SourceID != "" {
			fmt.Printf("Source ID:      %s\n", txn.SourceID)
			fmt.Printf("Source Type:    %s\n", txn.SourceType)
		}
		if txn.AvailableOn != nil {
			fmt.Printf("Available On:   %s\n", txn.AvailableOn.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", txn.CreatedAt.Format(time.RFC3339))
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

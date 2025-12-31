package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var payoutsCmd = &cobra.Command{
	Use:   "payouts",
	Short: "Manage payment payouts",
}

var payoutsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List payouts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.PayoutsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListPayouts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list payouts: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "AMOUNT", "CURRENCY", "STATUS", "TYPE", "BANK", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			rows = append(rows, []string{
				p.ID,
				p.Amount,
				p.Currency,
				p.Status,
				p.Type,
				p.BankAccount,
				p.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d payouts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var payoutsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get payout details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		payout, err := client.GetPayout(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get payout: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(payout)
		}

		fmt.Printf("Payout ID:      %s\n", payout.ID)
		fmt.Printf("Amount:         %s %s\n", payout.Amount, payout.Currency)
		fmt.Printf("Status:         %s\n", payout.Status)
		fmt.Printf("Type:           %s\n", payout.Type)
		fmt.Printf("Bank Account:   %s\n", payout.BankAccount)
		if payout.TransactionID != "" {
			fmt.Printf("Transaction ID: %s\n", payout.TransactionID)
		}
		if payout.Fee != "" {
			fmt.Printf("Fee:            %s\n", payout.Fee)
		}
		if payout.Net != "" {
			fmt.Printf("Net:            %s\n", payout.Net)
		}
		if payout.Summary != nil {
			fmt.Printf("Summary:\n")
			fmt.Printf("  Sales:        %s\n", payout.Summary.Sales)
			fmt.Printf("  Refunds:      %s\n", payout.Summary.Refunds)
			fmt.Printf("  Adjustments:  %s\n", payout.Summary.Adjustments)
			fmt.Printf("  Charges:      %s\n", payout.Summary.Charges)
		}
		if payout.ScheduledDate != nil {
			fmt.Printf("Scheduled:      %s\n", payout.ScheduledDate.Format(time.RFC3339))
		}
		if payout.ArrivalDate != nil {
			fmt.Printf("Arrival:        %s\n", payout.ArrivalDate.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", payout.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", payout.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(payoutsCmd)

	payoutsCmd.AddCommand(payoutsListCmd)
	payoutsListCmd.Flags().Int("page", 1, "Page number")
	payoutsListCmd.Flags().Int("page-size", 20, "Results per page")
	payoutsListCmd.Flags().String("status", "", "Filter by status (pending, in_transit, paid, failed, cancelled)")

	payoutsCmd.AddCommand(payoutsGetCmd)
}

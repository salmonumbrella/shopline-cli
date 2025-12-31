package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var memberPointsCmd = &cobra.Command{
	Use:   "member-points",
	Short: "Manage customer member points",
}

var memberPointsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get customer points balance",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")

		points, err := client.GetMemberPoints(cmd.Context(), customerID)
		if err != nil {
			return fmt.Errorf("failed to get member points: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(points)
		}

		fmt.Printf("Customer ID:      %s\n", points.CustomerID)
		fmt.Printf("Total Points:     %d\n", points.TotalPoints)
		fmt.Printf("Available Points: %d\n", points.AvailablePoints)
		fmt.Printf("Pending Points:   %d\n", points.PendingPoints)
		fmt.Printf("Expired Points:   %d\n", points.ExpiredPoints)
		fmt.Printf("Updated:          %s\n", points.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var memberPointsTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "List points transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		txnType, _ := cmd.Flags().GetString("type")

		opts := &api.PointsTransactionsListOptions{
			Page:     page,
			PageSize: pageSize,
			Type:     txnType,
		}

		resp, err := client.ListPointsTransactions(cmd.Context(), customerID, opts)
		if err != nil {
			return fmt.Errorf("failed to list points transactions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TYPE", "POINTS", "BALANCE", "DESCRIPTION", "ORDER", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			pointsStr := fmt.Sprintf("%+d", t.Points)
			orderID := t.OrderID
			if orderID == "" {
				orderID = "-"
			}
			desc := t.Description
			if len(desc) > 25 {
				desc = desc[:22] + "..."
			}
			rows = append(rows, []string{
				t.ID,
				t.Type,
				pointsStr,
				fmt.Sprintf("%d", t.Balance),
				desc,
				orderID,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d transactions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var memberPointsAdjustCmd = &cobra.Command{
	Use:   "adjust",
	Short: "Adjust customer points",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		points, _ := cmd.Flags().GetInt("points")
		description, _ := cmd.Flags().GetString("description")

		result, err := client.AdjustMemberPoints(cmd.Context(), customerID, points, description)
		if err != nil {
			return fmt.Errorf("failed to adjust member points: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(result)
		}

		fmt.Printf("Adjusted points by %+d\n", points)
		fmt.Printf("New Balance:\n")
		fmt.Printf("  Total Points:     %d\n", result.TotalPoints)
		fmt.Printf("  Available Points: %d\n", result.AvailablePoints)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(memberPointsCmd)

	memberPointsCmd.PersistentFlags().String("customer-id", "", "Customer ID")
	_ = memberPointsCmd.MarkPersistentFlagRequired("customer-id")

	memberPointsCmd.AddCommand(memberPointsGetCmd)

	memberPointsCmd.AddCommand(memberPointsTransactionsCmd)
	memberPointsTransactionsCmd.Flags().Int("page", 1, "Page number")
	memberPointsTransactionsCmd.Flags().Int("page-size", 20, "Results per page")
	memberPointsTransactionsCmd.Flags().String("type", "", "Filter by transaction type")

	memberPointsCmd.AddCommand(memberPointsAdjustCmd)
	memberPointsAdjustCmd.Flags().Int("points", 0, "Points to add (positive) or deduct (negative)")
	memberPointsAdjustCmd.Flags().String("description", "", "Description for the adjustment")
	_ = memberPointsAdjustCmd.MarkFlagRequired("points")
}

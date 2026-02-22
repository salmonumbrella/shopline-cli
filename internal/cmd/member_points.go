package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}

		points, err := client.GetMemberPoints(cmd.Context(), customerID)
		if err != nil {
			return fmt.Errorf("failed to get member points: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(points)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:      %s\n", points.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Points:     %d\n", points.TotalPoints)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available Points: %d\n", points.AvailablePoints)
		_, _ = fmt.Fprintf(outWriter(cmd), "Pending Points:   %d\n", points.PendingPoints)
		_, _ = fmt.Fprintf(outWriter(cmd), "Expired Points:   %d\n", points.ExpiredPoints)
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", points.UpdatedAt.Format(time.RFC3339))
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
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}
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
				outfmt.FormatID("member_point_transaction", t.ID),
				t.Type,
				pointsStr,
				fmt.Sprintf("%d", t.Balance),
				desc,
				orderID,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d transactions\n", len(resp.Items), resp.TotalCount)
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
		if checkDryRun(cmd, "[DRY-RUN] Would adjust member points") {
			return nil
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Adjusted points by %+d\n", points)
		_, _ = fmt.Fprintf(outWriter(cmd), "New Balance:\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "  Total Points:     %d\n", result.TotalPoints)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Available Points: %d\n", result.AvailablePoints)
		return nil
	},
}

var memberPointsHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Get customer member points history (Open API)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		customerID, _ := cmd.Flags().GetString("customer-id")
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}
		resp, err := client.GetCustomerMemberPointsHistory(cmd.Context(), customerID)
		if err != nil {
			return fmt.Errorf("failed to get member points history: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var memberPointsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update customer member points (Open API)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would update customer member points") {
			return nil
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		if strings.TrimSpace(customerID) == "" {
			return fmt.Errorf("customer id is required (use --customer-id)")
		}

		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		var req json.RawMessage
		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			req, err = readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
		} else {
			if !cmd.Flags().Changed("points") {
				return fmt.Errorf("request body required (use --body/--body-file or provide --points)")
			}
			points, _ := cmd.Flags().GetInt("points")
			description, _ := cmd.Flags().GetString("description")
			req, err = json.Marshal(map[string]any{
				"points":      points,
				"description": description,
			})
			if err != nil {
				return fmt.Errorf("failed to build request body: %w", err)
			}
		}

		resp, err := client.UpdateCustomerMemberPoints(cmd.Context(), customerID, req)
		if err != nil {
			return fmt.Errorf("failed to update member points: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var memberPointsRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Member point rules (Open API)",
}

var memberPointsRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List member point rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.ListMemberPointRules(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list member point rules: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var memberPointsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Bulk update member points (Open API) (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would bulk update member points") {
			return nil
		}

		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.BulkUpdateMemberPoints(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to bulk update member points: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(memberPointsCmd)

	memberPointsCmd.PersistentFlags().String("customer-id", "", "Customer ID")

	memberPointsCmd.AddCommand(memberPointsGetCmd)

	memberPointsCmd.AddCommand(memberPointsTransactionsCmd)
	memberPointsTransactionsCmd.Flags().Int("page", 1, "Page number")
	memberPointsTransactionsCmd.Flags().Int("page-size", 20, "Results per page")
	memberPointsTransactionsCmd.Flags().String("type", "", "Filter by transaction type")

	memberPointsCmd.AddCommand(memberPointsAdjustCmd)
	memberPointsAdjustCmd.Flags().Int("points", 0, "Points to add (positive) or deduct (negative)")
	memberPointsAdjustCmd.Flags().String("description", "", "Description for the adjustment")
	_ = memberPointsAdjustCmd.MarkFlagRequired("points")

	memberPointsCmd.AddCommand(memberPointsHistoryCmd)

	memberPointsCmd.AddCommand(memberPointsUpdateCmd)
	addJSONBodyFlags(memberPointsUpdateCmd)
	memberPointsUpdateCmd.Flags().Int("points", 0, "Points to add (positive) or deduct (negative) (ignored when --body/--body-file set)")
	memberPointsUpdateCmd.Flags().String("description", "", "Description for the update (ignored when --body/--body-file set)")

	memberPointsCmd.AddCommand(memberPointsRulesCmd)
	memberPointsRulesCmd.AddCommand(memberPointsRulesListCmd)

	memberPointsCmd.AddCommand(memberPointsBulkUpdateCmd)
	addJSONBodyFlags(memberPointsBulkUpdateCmd)
}

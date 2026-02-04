package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var operationLogsCmd = &cobra.Command{
	Use:     "operation-logs",
	Aliases: []string{"audit-logs", "audit", "logs"},
	Short:   "View operation audit logs",
}

var operationLogsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List operation logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		action, _ := cmd.Flags().GetString("action")
		resourceType, _ := cmd.Flags().GetString("resource-type")
		resourceID, _ := cmd.Flags().GetString("resource-id")
		userID, _ := cmd.Flags().GetString("user-id")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		since, _ := cmd.Flags().GetString("since")
		until, _ := cmd.Flags().GetString("until")

		if from == "" {
			from = since
		}
		if to == "" {
			to = until
		}

		opts := &api.OperationLogsListOptions{
			Page:         page,
			PageSize:     pageSize,
			Action:       api.OperationLogAction(action),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			UserID:       userID,
		}

		if from != "" {
			t, err := time.Parse(time.RFC3339, from)
			if err != nil {
				t, err = time.Parse("2006-01-02", from)
				if err != nil {
					return fmt.Errorf("invalid from date format, use RFC3339 or YYYY-MM-DD: %w", err)
				}
			}
			opts.StartDate = &t
		}
		if to != "" {
			t, err := time.Parse(time.RFC3339, to)
			if err != nil {
				t, err = time.Parse("2006-01-02", to)
				if err != nil {
					return fmt.Errorf("invalid to date format, use RFC3339 or YYYY-MM-DD: %w", err)
				}
			}
			opts.EndDate = &t
		}

		resp, err := client.ListOperationLogs(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list operation logs: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ACTION", "RESOURCE", "USER", "IP", "CREATED"}
		var rows [][]string
		for _, l := range resp.Items {
			resource := l.ResourceType
			if l.ResourceID != "" {
				resource += ":" + l.ResourceID
			}
			user := l.UserEmail
			if user == "" {
				user = l.UserName
			}
			rows = append(rows, []string{
				l.ID,
				string(l.Action),
				resource,
				user,
				l.IPAddress,
				l.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d operation logs\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var operationLogsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get operation log details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		log, err := client.GetOperationLog(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get operation log: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(log)
		}

		fmt.Printf("Operation Log ID: %s\n", log.ID)
		fmt.Printf("Action:           %s\n", log.Action)
		fmt.Printf("Resource Type:    %s\n", log.ResourceType)
		fmt.Printf("Resource ID:      %s\n", log.ResourceID)
		if log.ResourceName != "" {
			fmt.Printf("Resource Name:    %s\n", log.ResourceName)
		}
		fmt.Println()
		fmt.Printf("User ID:          %s\n", log.UserID)
		fmt.Printf("User Email:       %s\n", log.UserEmail)
		if log.UserName != "" {
			fmt.Printf("User Name:        %s\n", log.UserName)
		}
		fmt.Printf("IP Address:       %s\n", log.IPAddress)
		if log.UserAgent != "" {
			fmt.Printf("User Agent:       %s\n", log.UserAgent)
		}
		fmt.Printf("Created:          %s\n", log.CreatedAt.Format(time.RFC3339))

		if len(log.Changes) > 0 {
			fmt.Printf("\nChanges:\n")
			for field, change := range log.Changes {
				fmt.Printf("  %s: %v -> %v\n", field, change.From, change.To)
			}
		}

		if len(log.Metadata) > 0 {
			fmt.Printf("\nMetadata:\n")
			for key, value := range log.Metadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(operationLogsCmd)

	operationLogsCmd.AddCommand(operationLogsListCmd)
	operationLogsListCmd.Flags().Int("page", 1, "Page number")
	operationLogsListCmd.Flags().Int("page-size", 20, "Results per page")
	operationLogsListCmd.Flags().String("action", "", "Filter by action (create, update, delete, login, logout, export, import)")
	operationLogsListCmd.Flags().String("resource-type", "", "Filter by resource type (product, order, customer, etc.)")
	operationLogsListCmd.Flags().String("resource-id", "", "Filter by resource ID")
	operationLogsListCmd.Flags().String("user-id", "", "Filter by user ID")
	operationLogsListCmd.Flags().String("from", "", "Filter by start date (YYYY-MM-DD or RFC3339)")
	operationLogsListCmd.Flags().String("to", "", "Filter by end date (YYYY-MM-DD or RFC3339)")
	operationLogsListCmd.Flags().String("since", "", "Filter by start date (YYYY-MM-DD or RFC3339)")
	operationLogsListCmd.Flags().String("until", "", "Filter by end date (YYYY-MM-DD or RFC3339)")

	operationLogsCmd.AddCommand(operationLogsGetCmd)
}

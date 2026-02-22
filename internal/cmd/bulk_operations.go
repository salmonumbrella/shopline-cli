package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var bulkOperationsCmd = &cobra.Command{
	Use:   "bulk-operations",
	Short: "Manage bulk operations",
}

var bulkOperationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bulk operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		opType, _ := cmd.Flags().GetString("type")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.BulkOperationsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Type:     opType,
		}

		resp, err := client.ListBulkOperations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list bulk operations: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TYPE", "STATUS", "OBJECTS", "FILE SIZE", "CREATED"}
		var rows [][]string
		for _, op := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("bulk_operation", op.ID),
				op.Type,
				op.Status,
				fmt.Sprintf("%d", op.ObjectCount),
				formatBytes(op.FileSize),
				op.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d bulk operations\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var bulkOperationsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get bulk operation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		op, err := client.GetBulkOperation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get bulk operation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(op)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Operation ID:   %s\n", op.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:           %s\n", op.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", op.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Object Count:   %d\n", op.ObjectCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "File Size:      %s\n", formatBytes(op.FileSize))
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", op.CreatedAt.Format(time.RFC3339))
		if op.CompletedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Completed:      %s\n", op.CompletedAt.Format(time.RFC3339))
		}
		if op.URL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Result URL:     %s\n", op.URL)
		}
		if op.ErrorCode != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Error Code:     %s\n", op.ErrorCode)
		}
		if op.PartialDataURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Partial Data:   %s\n", op.PartialDataURL)
		}
		return nil
	},
}

var bulkOperationsCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Get the currently running bulk operation",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		op, err := client.GetCurrentBulkOperation(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get current bulk operation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(op)
		}

		if op.ID == "" {
			_, _ = fmt.Fprintln(outWriter(cmd), "No bulk operation currently running")
			return nil
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Operation ID:   %s\n", op.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:           %s\n", op.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", op.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Object Count:   %d\n", op.ObjectCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", op.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var bulkOperationsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Create a bulk query operation",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create bulk query operation") {
			return nil
		}

		query, _ := cmd.Flags().GetString("graphql")

		req := &api.BulkOperationCreateRequest{
			Query: query,
		}

		op, err := client.CreateBulkQuery(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create bulk query: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(op)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created bulk query %s\n", op.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status: %s\n", op.Status)
		return nil
	},
}

var bulkOperationsMutationCmd = &cobra.Command{
	Use:   "mutation",
	Short: "Create a bulk mutation operation",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create bulk mutation operation") {
			return nil
		}

		mutation, _ := cmd.Flags().GetString("graphql")
		stagedUploadPath, _ := cmd.Flags().GetString("staged-upload-path")

		req := &api.BulkOperationMutationRequest{
			Mutation:         mutation,
			StagedUploadPath: stagedUploadPath,
		}

		op, err := client.CreateBulkMutation(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create bulk mutation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(op)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created bulk mutation %s\n", op.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status: %s\n", op.Status)
		return nil
	},
}

var bulkOperationsCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a bulk operation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would cancel bulk operation %s", args[0])) {
			return nil
		}

		op, err := client.CancelBulkOperation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to cancel bulk operation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(op)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Cancelled bulk operation %s\n", op.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status: %s\n", op.Status)
		return nil
	},
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	rootCmd.AddCommand(bulkOperationsCmd)

	bulkOperationsCmd.AddCommand(bulkOperationsListCmd)
	bulkOperationsListCmd.Flags().String("status", "", "Filter by status (created, running, completed, failed, cancelled)")
	bulkOperationsListCmd.Flags().String("type", "", "Filter by type (query, mutation)")
	bulkOperationsListCmd.Flags().Int("page", 1, "Page number")
	bulkOperationsListCmd.Flags().Int("page-size", 20, "Results per page")

	bulkOperationsCmd.AddCommand(bulkOperationsGetCmd)
	bulkOperationsCmd.AddCommand(bulkOperationsCurrentCmd)

	bulkOperationsCmd.AddCommand(bulkOperationsQueryCmd)
	bulkOperationsQueryCmd.Flags().String("graphql", "", "GraphQL query to execute for bulk operation")
	_ = bulkOperationsQueryCmd.MarkFlagRequired("graphql")

	bulkOperationsCmd.AddCommand(bulkOperationsMutationCmd)
	bulkOperationsMutationCmd.Flags().String("graphql", "", "GraphQL mutation to execute for bulk operation")
	_ = bulkOperationsMutationCmd.MarkFlagRequired("graphql")
	bulkOperationsMutationCmd.Flags().String("staged-upload-path", "", "Staged upload path for the mutation payload (e.g. tmp/bulk-mutation.jsonl)")
	_ = bulkOperationsMutationCmd.MarkFlagRequired("staged-upload-path")

	bulkOperationsCmd.AddCommand(bulkOperationsCancelCmd)
}

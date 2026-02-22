package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var orderRisksCmd = &cobra.Command{
	Use:   "order-risks",
	Short: "Manage order risk assessments",
}

var orderRisksListCmd = &cobra.Command{
	Use:   "list <order-id>",
	Short: "List risks for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.OrderRisksListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListOrderRisks(cmd.Context(), args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to list order risks: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "SCORE", "RECOMMENDATION", "SOURCE", "DISPLAY", "CANCEL", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("order_risk", r.ID),
				fmt.Sprintf("%.2f", r.Score),
				r.Recommendation,
				r.Source,
				strconv.FormatBool(r.Display),
				strconv.FormatBool(r.CauseCancel),
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d risks for order %s\n", len(resp.Items), resp.TotalCount, args[0])
		return nil
	},
}

var orderRisksGetCmd = &cobra.Command{
	Use:   "get <order-id> <risk-id>",
	Short: "Get order risk details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		risk, err := client.GetOrderRisk(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get order risk: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(risk)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Risk ID:         %s\n", risk.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order ID:        %s\n", risk.OrderID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Score:           %.2f\n", risk.Score)
		_, _ = fmt.Fprintf(outWriter(cmd), "Recommendation:  %s\n", risk.Recommendation)
		_, _ = fmt.Fprintf(outWriter(cmd), "Source:          %s\n", risk.Source)
		if risk.Message != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Message:         %s\n", risk.Message)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Display:         %t\n", risk.Display)
		_, _ = fmt.Fprintf(outWriter(cmd), "Cause Cancel:    %t\n", risk.CauseCancel)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", risk.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", risk.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var orderRisksCreateCmd = &cobra.Command{
	Use:   "create <order-id>",
	Short: "Create a risk assessment for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create order risk for order %s", args[0])) {
			return nil
		}

		score, _ := cmd.Flags().GetFloat64("score")
		recommendation, _ := cmd.Flags().GetString("recommendation")
		source, _ := cmd.Flags().GetString("source")
		message, _ := cmd.Flags().GetString("message")
		display, _ := cmd.Flags().GetBool("display")
		causeCancel, _ := cmd.Flags().GetBool("cause-cancel")

		req := &api.OrderRiskCreateRequest{
			Score:          score,
			Recommendation: recommendation,
			Source:         source,
			Message:        message,
			Display:        display,
			CauseCancel:    causeCancel,
		}

		risk, err := client.CreateOrderRisk(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to create order risk: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(risk)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created risk %s for order %s\n", risk.ID, args[0])
		_, _ = fmt.Fprintf(outWriter(cmd), "Score: %.2f, Recommendation: %s\n", risk.Score, risk.Recommendation)
		return nil
	},
}

var orderRisksUpdateCmd = &cobra.Command{
	Use:   "update <order-id> <risk-id>",
	Short: "Update an order risk assessment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update risk %s for order %s", args[1], args[0])) {
			return nil
		}

		var req api.OrderRiskUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		risk, err := client.UpdateOrderRisk(cmd.Context(), args[0], args[1], &req)
		if err != nil {
			return fmt.Errorf("failed to update order risk: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(risk)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated risk %s for order %s\n", risk.ID, risk.OrderID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Score: %.2f, Recommendation: %s\n", risk.Score, risk.Recommendation)
		return nil
	},
}

var orderRisksDeleteCmd = &cobra.Command{
	Use:   "delete <order-id> <risk-id>",
	Short: "Delete an order risk assessment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete risk %s from order %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteOrderRisk(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to delete order risk: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted risk %s from order %s.\n", args[1], args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orderRisksCmd)

	orderRisksCmd.AddCommand(orderRisksListCmd)
	orderRisksListCmd.Flags().Int("page", 1, "Page number")
	orderRisksListCmd.Flags().Int("page-size", 20, "Results per page")

	orderRisksCmd.AddCommand(orderRisksGetCmd)

	orderRisksCmd.AddCommand(orderRisksCreateCmd)
	orderRisksCreateCmd.Flags().Float64("score", 0, "Risk score (0.0 to 1.0)")
	orderRisksCreateCmd.Flags().String("recommendation", "", "Recommendation (accept, investigate, cancel)")
	orderRisksCreateCmd.Flags().String("source", "", "Risk source")
	orderRisksCreateCmd.Flags().String("message", "", "Risk message")
	orderRisksCreateCmd.Flags().Bool("display", false, "Display risk to merchant")
	orderRisksCreateCmd.Flags().Bool("cause-cancel", false, "Risk should cause order cancellation")
	_ = orderRisksCreateCmd.MarkFlagRequired("score")
	_ = orderRisksCreateCmd.MarkFlagRequired("recommendation")

	orderRisksCmd.AddCommand(orderRisksUpdateCmd)
	addJSONBodyFlags(orderRisksUpdateCmd)
	orderRisksUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	orderRisksCmd.AddCommand(orderRisksDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "order-risks",
		Description: "Manage order risk assessments",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "order_risk",
	})
}

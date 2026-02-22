package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var affiliateCampaignsCmd = &cobra.Command{
	Use:   "affiliate-campaigns",
	Short: "Manage affiliate marketing campaigns",
}

var affiliateCampaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List affiliate campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.AffiliateCampaignsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListAffiliateCampaigns(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list affiliate campaigns: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "STATUS", "COMMISSION", "SALES", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			commission := fmt.Sprintf("%.2f%%", c.CommissionValue)
			if c.CommissionType == "fixed" {
				commission = fmt.Sprintf("$%.2f", c.CommissionValue)
			}
			rows = append(rows, []string{
				outfmt.FormatID("affiliate_campaign", c.ID),
				c.Name,
				c.Status,
				commission,
				fmt.Sprintf("%d", c.TotalSales),
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d affiliate campaigns\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var affiliateCampaignsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get affiliate campaign details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		campaign, err := client.GetAffiliateCampaign(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get affiliate campaign: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(campaign)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Campaign ID:     %s\n", campaign.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", campaign.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", campaign.Status)
		if campaign.Description != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", campaign.Description)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Commission Type: %s\n", campaign.CommissionType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Commission:      %.2f\n", campaign.CommissionValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Clicks:    %d\n", campaign.TotalClicks)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Sales:     %d\n", campaign.TotalSales)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Revenue:   $%.2f\n", campaign.TotalRevenue)
		if !campaign.StartDate.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Start Date:      %s\n", campaign.StartDate.Format(time.RFC3339))
		}
		if !campaign.EndDate.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "End Date:        %s\n", campaign.EndDate.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", campaign.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", campaign.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var affiliateCampaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an affiliate campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		commissionType, _ := cmd.Flags().GetString("commission-type")
		commissionValue, _ := cmd.Flags().GetFloat64("commission-value")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create affiliate campaign: %s", name)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AffiliateCampaignCreateRequest{
			Name:            name,
			Description:     description,
			CommissionType:  commissionType,
			CommissionValue: commissionValue,
		}

		campaign, err := client.CreateAffiliateCampaign(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create affiliate campaign: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(campaign)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created affiliate campaign %s\n", campaign.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:       %s\n", campaign.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Commission: %.2f (%s)\n", campaign.CommissionValue, campaign.CommissionType)

		return nil
	},
}

var affiliateCampaignsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an affiliate campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update affiliate campaign %s", args[0])) {
			return nil
		}

		var req api.AffiliateCampaignUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		campaign, err := client.UpdateAffiliateCampaign(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update affiliate campaign: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(campaign)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated affiliate campaign %s\n", campaign.ID)
		return nil
	},
}

var affiliateCampaignsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an affiliate campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete affiliate campaign %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete affiliate campaign %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteAffiliateCampaign(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete affiliate campaign: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted affiliate campaign %s\n", args[0])
		return nil
	},
}

var affiliateCampaignsOrdersCmd = &cobra.Command{
	Use:   "orders <id>",
	Short: "Get affiliate campaign orders (documented endpoint; raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.GetAffiliateCampaignOrders(cmd.Context(), args[0], &api.AffiliateCampaignOrdersOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to get affiliate campaign orders: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var affiliateCampaignsSummaryCmd = &cobra.Command{
	Use:   "summary <id>",
	Short: "Get affiliate campaign summary (documented endpoint; raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetAffiliateCampaignSummary(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get affiliate campaign summary: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var affiliateCampaignsProductsSalesRankingCmd = &cobra.Command{
	Use:   "products-sales-ranking <id>",
	Short: "Get products sales ranking of campaign (documented endpoint; raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.GetAffiliateCampaignProductsSalesRanking(cmd.Context(), args[0], &api.AffiliateCampaignProductsSalesRankingOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to get affiliate campaign products sales ranking: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var affiliateCampaignsExportReportCmd = &cobra.Command{
	Use:   "export-report <id>",
	Short: "Export affiliate campaign report to partner (documented endpoint; raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would export affiliate campaign report for %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ExportAffiliateCampaignReport(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to export affiliate campaign report: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(affiliateCampaignsCmd)

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsListCmd)
	affiliateCampaignsListCmd.Flags().Int("page", 1, "Page number")
	affiliateCampaignsListCmd.Flags().Int("page-size", 20, "Results per page")
	affiliateCampaignsListCmd.Flags().String("status", "", "Filter by status (active, paused, ended)")

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsGetCmd)

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsCreateCmd)
	affiliateCampaignsCreateCmd.Flags().String("name", "", "Campaign name (required)")
	affiliateCampaignsCreateCmd.Flags().String("description", "", "Campaign description")
	affiliateCampaignsCreateCmd.Flags().String("commission-type", "percentage", "Commission type (percentage, fixed)")
	affiliateCampaignsCreateCmd.Flags().Float64("commission-value", 0, "Commission value (required)")
	_ = affiliateCampaignsCreateCmd.MarkFlagRequired("name")
	_ = affiliateCampaignsCreateCmd.MarkFlagRequired("commission-value")

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsUpdateCmd)
	addJSONBodyFlags(affiliateCampaignsUpdateCmd)

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsDeleteCmd)
	affiliateCampaignsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsOrdersCmd)
	affiliateCampaignsOrdersCmd.Flags().Int("page", 1, "Page number")
	affiliateCampaignsOrdersCmd.Flags().Int("page-size", 20, "Results per page")

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsSummaryCmd)

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsProductsSalesRankingCmd)
	affiliateCampaignsProductsSalesRankingCmd.Flags().Int("page", 1, "Page number")
	affiliateCampaignsProductsSalesRankingCmd.Flags().Int("page-size", 20, "Results per page")

	affiliateCampaignsCmd.AddCommand(affiliateCampaignsExportReportCmd)
	addJSONBodyFlags(affiliateCampaignsExportReportCmd)
}

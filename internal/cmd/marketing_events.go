package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var marketingEventsCmd = &cobra.Command{
	Use:   "marketing-events",
	Short: "Manage marketing event tracking",
}

var marketingEventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List marketing events",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		eventType, _ := cmd.Flags().GetString("event-type")
		marketingType, _ := cmd.Flags().GetString("marketing-type")

		opts := &api.MarketingEventsListOptions{
			Page:          page,
			PageSize:      pageSize,
			EventType:     eventType,
			MarketingType: marketingType,
		}

		resp, err := client.ListMarketingEvents(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list marketing events: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EVENT TYPE", "MARKETING TYPE", "UTM CAMPAIGN", "BUDGET", "CREATED"}
		var rows [][]string
		for _, e := range resp.Items {
			budget := "-"
			if e.Budget > 0 {
				budget = fmt.Sprintf("%.2f %s", e.Budget, e.Currency)
			}
			rows = append(rows, []string{
				outfmt.FormatID("marketing_event", e.ID),
				e.EventType,
				e.MarketingType,
				e.UTMCampaign,
				budget,
				e.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d marketing events\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var marketingEventsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get marketing event details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		event, err := client.GetMarketingEvent(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get marketing event: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(event)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Event ID:       %s\n", event.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Event Type:     %s\n", event.EventType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Marketing Type: %s\n", event.MarketingType)
		if event.RemoteID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Remote ID:      %s\n", event.RemoteID)
		}
		if event.Description != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", event.Description)
		}
		if event.Budget > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Budget:         %.2f %s\n", event.Budget, event.Currency)
		}
		if event.UTMCampaign != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "UTM Campaign:   %s\n", event.UTMCampaign)
		}
		if event.UTMSource != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "UTM Source:     %s\n", event.UTMSource)
		}
		if event.UTMMedium != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "UTM Medium:     %s\n", event.UTMMedium)
		}
		if event.ManageURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Manage URL:     %s\n", event.ManageURL)
		}
		if event.PreviewURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Preview URL:    %s\n", event.PreviewURL)
		}
		if !event.StartedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Started At:     %s\n", event.StartedAt.Format(time.RFC3339))
		}
		if !event.EndedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ended At:       %s\n", event.EndedAt.Format(time.RFC3339))
		}
		if len(event.MarketedResources) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Resources:      %d marketed resources\n", len(event.MarketedResources))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", event.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", event.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var marketingEventsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a marketing event",
	RunE: func(cmd *cobra.Command, args []string) error {
		eventType, _ := cmd.Flags().GetString("event-type")
		marketingType, _ := cmd.Flags().GetString("marketing-type")
		utmCampaign, _ := cmd.Flags().GetString("utm-campaign")
		utmSource, _ := cmd.Flags().GetString("utm-source")
		utmMedium, _ := cmd.Flags().GetString("utm-medium")
		budget, _ := cmd.Flags().GetFloat64("budget")
		currency, _ := cmd.Flags().GetString("currency")
		description, _ := cmd.Flags().GetString("description")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create marketing event: %s (%s)", eventType, marketingType)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.MarketingEventCreateRequest{
			EventType:     eventType,
			MarketingType: marketingType,
			UTMCampaign:   utmCampaign,
			UTMSource:     utmSource,
			UTMMedium:     utmMedium,
			Budget:        budget,
			Currency:      currency,
			Description:   description,
		}

		event, err := client.CreateMarketingEvent(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create marketing event: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(event)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created marketing event %s\n", event.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Event Type:     %s\n", event.EventType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Marketing Type: %s\n", event.MarketingType)

		return nil
	},
}

var marketingEventsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a marketing event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update marketing event %s", args[0])) {
			return nil
		}

		var req api.MarketingEventUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		event, err := client.UpdateMarketingEvent(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update marketing event: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(event)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated marketing event %s\n", event.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Event Type:     %s\n", event.EventType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Marketing Type: %s\n", event.MarketingType)
		return nil
	},
}

var marketingEventsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a marketing event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete marketing event %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete marketing event %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteMarketingEvent(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete marketing event: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted marketing event %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(marketingEventsCmd)

	marketingEventsCmd.AddCommand(marketingEventsListCmd)
	marketingEventsListCmd.Flags().Int("page", 1, "Page number")
	marketingEventsListCmd.Flags().Int("page-size", 20, "Results per page")
	marketingEventsListCmd.Flags().String("event-type", "", "Filter by event type (ad, campaign, email, social)")
	marketingEventsListCmd.Flags().String("marketing-type", "", "Filter by marketing type (cpc, display, social, search, email)")

	marketingEventsCmd.AddCommand(marketingEventsGetCmd)

	marketingEventsCmd.AddCommand(marketingEventsCreateCmd)
	marketingEventsCreateCmd.Flags().String("event-type", "", "Event type (ad, campaign, email, social) (required)")
	marketingEventsCreateCmd.Flags().String("marketing-type", "", "Marketing type (cpc, display, social, search, email) (required)")
	marketingEventsCreateCmd.Flags().String("utm-campaign", "", "UTM campaign parameter")
	marketingEventsCreateCmd.Flags().String("utm-source", "", "UTM source parameter")
	marketingEventsCreateCmd.Flags().String("utm-medium", "", "UTM medium parameter")
	marketingEventsCreateCmd.Flags().Float64("budget", 0, "Campaign budget")
	marketingEventsCreateCmd.Flags().String("currency", "USD", "Budget currency")
	marketingEventsCreateCmd.Flags().String("description", "", "Event description")
	_ = marketingEventsCreateCmd.MarkFlagRequired("event-type")
	_ = marketingEventsCreateCmd.MarkFlagRequired("marketing-type")

	marketingEventsCmd.AddCommand(marketingEventsUpdateCmd)
	addJSONBodyFlags(marketingEventsUpdateCmd)
	marketingEventsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	marketingEventsCmd.AddCommand(marketingEventsDeleteCmd)
	marketingEventsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	schema.Register(schema.Resource{
		Name:        "marketing-events",
		Description: "Manage marketing event tracking",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "marketing_event",
	})
}

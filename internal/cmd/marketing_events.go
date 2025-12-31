package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				e.ID,
				e.EventType,
				e.MarketingType,
				e.UTMCampaign,
				budget,
				e.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d marketing events\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Event ID:       %s\n", event.ID)
		fmt.Printf("Event Type:     %s\n", event.EventType)
		fmt.Printf("Marketing Type: %s\n", event.MarketingType)
		if event.RemoteID != "" {
			fmt.Printf("Remote ID:      %s\n", event.RemoteID)
		}
		if event.Description != "" {
			fmt.Printf("Description:    %s\n", event.Description)
		}
		if event.Budget > 0 {
			fmt.Printf("Budget:         %.2f %s\n", event.Budget, event.Currency)
		}
		if event.UTMCampaign != "" {
			fmt.Printf("UTM Campaign:   %s\n", event.UTMCampaign)
		}
		if event.UTMSource != "" {
			fmt.Printf("UTM Source:     %s\n", event.UTMSource)
		}
		if event.UTMMedium != "" {
			fmt.Printf("UTM Medium:     %s\n", event.UTMMedium)
		}
		if event.ManageURL != "" {
			fmt.Printf("Manage URL:     %s\n", event.ManageURL)
		}
		if event.PreviewURL != "" {
			fmt.Printf("Preview URL:    %s\n", event.PreviewURL)
		}
		if !event.StartedAt.IsZero() {
			fmt.Printf("Started At:     %s\n", event.StartedAt.Format(time.RFC3339))
		}
		if !event.EndedAt.IsZero() {
			fmt.Printf("Ended At:       %s\n", event.EndedAt.Format(time.RFC3339))
		}
		if len(event.MarketedResources) > 0 {
			fmt.Printf("Resources:      %d marketed resources\n", len(event.MarketedResources))
		}
		fmt.Printf("Created:        %s\n", event.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", event.UpdatedAt.Format(time.RFC3339))
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create marketing event: %s (%s)\n", eventType, marketingType)
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

		fmt.Printf("Created marketing event %s\n", event.ID)
		fmt.Printf("Event Type:     %s\n", event.EventType)
		fmt.Printf("Marketing Type: %s\n", event.MarketingType)

		return nil
	},
}

var marketingEventsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a marketing event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete marketing event %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete marketing event %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteMarketingEvent(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete marketing event: %w", err)
		}

		fmt.Printf("Deleted marketing event %s\n", args[0])
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

	marketingEventsCmd.AddCommand(marketingEventsDeleteCmd)
	marketingEventsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var cdpCmd = &cobra.Command{
	Use:   "cdp",
	Short: "Access Customer Data Platform analytics",
}

// CDP Profiles commands
var cdpProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage CDP customer profiles",
}

var cdpProfilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDP customer profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		segment, _ := cmd.Flags().GetString("segment")
		tag, _ := cmd.Flags().GetString("tag")
		churnRisk, _ := cmd.Flags().GetString("churn-risk")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CDPProfilesListOptions{
			Page:      page,
			PageSize:  pageSize,
			Segment:   segment,
			Tag:       tag,
			ChurnRisk: churnRisk,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListCDPProfiles(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list CDP profiles: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "ORDERS", "TOTAL SPENT", "LTV", "CHURN RISK", "SEGMENTS"}
		var rows [][]string
		for _, p := range resp.Items {
			segments := strings.Join(p.Segments, ", ")
			if len(segments) > 30 {
				segments = segments[:27] + "..."
			}
			rows = append(rows, []string{
				p.ID,
				p.Email,
				fmt.Sprintf("%d", p.TotalOrders),
				p.TotalSpent,
				p.LifetimeValue,
				p.ChurnRisk,
				segments,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d profiles\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var cdpProfilesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get CDP profile details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		profile, err := client.GetCDPProfile(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get CDP profile: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(profile)
		}

		fmt.Printf("Profile ID:       %s\n", profile.ID)
		fmt.Printf("Customer ID:      %s\n", profile.CustomerID)
		fmt.Printf("Email:            %s\n", profile.Email)
		fmt.Printf("Phone:            %s\n", profile.Phone)
		fmt.Printf("Name:             %s %s\n", profile.FirstName, profile.LastName)
		fmt.Printf("Total Orders:     %d\n", profile.TotalOrders)
		fmt.Printf("Total Spent:      %s\n", profile.TotalSpent)
		fmt.Printf("Avg Order Value:  %s\n", profile.AverageOrderValue)
		fmt.Printf("Lifetime Value:   %s\n", profile.LifetimeValue)
		fmt.Printf("Predicted LTV:    %s\n", profile.PredictedLTV)
		fmt.Printf("Churn Risk:       %s\n", profile.ChurnRisk)
		fmt.Printf("Segments:         %s\n", strings.Join(profile.Segments, ", "))
		fmt.Printf("Tags:             %s\n", strings.Join(profile.Tags, ", "))

		if profile.RFMScore != nil {
			fmt.Printf("\nRFM Analysis:\n")
			fmt.Printf("  Recency:        %d\n", profile.RFMScore.Recency)
			fmt.Printf("  Frequency:      %d\n", profile.RFMScore.Frequency)
			fmt.Printf("  Monetary:       %d\n", profile.RFMScore.Monetary)
			fmt.Printf("  Total Score:    %d\n", profile.RFMScore.Total)
			fmt.Printf("  Segment:        %s\n", profile.RFMScore.Segment)
		}

		if profile.Preferences != nil {
			fmt.Printf("\nPreferences:\n")
			fmt.Printf("  Email Marketing:  %v\n", profile.Preferences.EmailMarketing)
			fmt.Printf("  SMS Marketing:    %v\n", profile.Preferences.SMSMarketing)
			fmt.Printf("  Push Notifications: %v\n", profile.Preferences.PushNotifications)
			fmt.Printf("  Preferred Channel: %s\n", profile.Preferences.PreferredChannel)
		}

		if profile.FirstOrderAt != nil {
			fmt.Printf("\nFirst Order:      %s\n", profile.FirstOrderAt.Format(time.RFC3339))
		}
		if profile.LastOrderAt != nil {
			fmt.Printf("Last Order:       %s\n", profile.LastOrderAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:          %s\n", profile.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:          %s\n", profile.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

// CDP Events commands
var cdpEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "View CDP events",
}

var cdpEventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDP events",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		eventType, _ := cmd.Flags().GetString("event-type")
		eventName, _ := cmd.Flags().GetString("event-name")
		source, _ := cmd.Flags().GetString("source")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CDPEventsListOptions{
			Page:       page,
			PageSize:   pageSize,
			CustomerID: customerID,
			EventType:  eventType,
			EventName:  eventName,
			Source:     source,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListCDPEvents(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list CDP events: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER", "TYPE", "NAME", "SOURCE", "CHANNEL", "TIMESTAMP"}
		var rows [][]string
		for _, e := range resp.Items {
			rows = append(rows, []string{
				e.ID,
				e.CustomerID,
				e.EventType,
				e.EventName,
				e.Source,
				e.Channel,
				e.Timestamp.Format("2006-01-02 15:04:05"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d events\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var cdpEventsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get CDP event details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		event, err := client.GetCDPEvent(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get CDP event: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(event)
		}

		fmt.Printf("Event ID:     %s\n", event.ID)
		fmt.Printf("Customer ID:  %s\n", event.CustomerID)
		fmt.Printf("Session ID:   %s\n", event.SessionID)
		fmt.Printf("Event Type:   %s\n", event.EventType)
		fmt.Printf("Event Name:   %s\n", event.EventName)
		fmt.Printf("Source:       %s\n", event.Source)
		fmt.Printf("Channel:      %s\n", event.Channel)
		fmt.Printf("Timestamp:    %s\n", event.Timestamp.Format(time.RFC3339))
		fmt.Printf("Created:      %s\n", event.CreatedAt.Format(time.RFC3339))

		if len(event.Properties) > 0 {
			fmt.Printf("\nProperties:\n")
			for k, v := range event.Properties {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}

		return nil
	},
}

// CDP Segments commands
var cdpSegmentsCmd = &cobra.Command{
	Use:   "segments",
	Short: "Manage CDP segments",
}

var cdpSegmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDP segments",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		segmentType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CDPSegmentsListOptions{
			Page:     page,
			PageSize: pageSize,
			Type:     segmentType,
			Status:   status,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListCDPSegments(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list CDP segments: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "TYPE", "CUSTOMERS", "STATUS", "UPDATED"}
		var rows [][]string
		for _, s := range resp.Items {
			rows = append(rows, []string{
				s.ID,
				s.Name,
				s.Type,
				fmt.Sprintf("%d", s.CustomerCount),
				s.Status,
				s.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d segments\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var cdpSegmentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get CDP segment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		segment, err := client.GetCDPSegment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get CDP segment: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(segment)
		}

		fmt.Printf("Segment ID:      %s\n", segment.ID)
		fmt.Printf("Name:            %s\n", segment.Name)
		fmt.Printf("Description:     %s\n", segment.Description)
		fmt.Printf("Type:            %s\n", segment.Type)
		fmt.Printf("Status:          %s\n", segment.Status)
		fmt.Printf("Customer Count:  %d\n", segment.CustomerCount)
		fmt.Printf("Created:         %s\n", segment.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:         %s\n", segment.UpdatedAt.Format(time.RFC3339))

		if len(segment.Conditions) > 0 {
			fmt.Printf("\nConditions:\n")
			for i, c := range segment.Conditions {
				fmt.Printf("  %d. %s %s %v\n", i+1, c.Field, c.Operator, c.Value)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cdpCmd)

	// Profiles subcommands
	cdpCmd.AddCommand(cdpProfilesCmd)

	cdpProfilesCmd.AddCommand(cdpProfilesListCmd)
	cdpProfilesListCmd.Flags().String("segment", "", "Filter by segment")
	cdpProfilesListCmd.Flags().String("tag", "", "Filter by tag")
	cdpProfilesListCmd.Flags().String("churn-risk", "", "Filter by churn risk (low, medium, high)")
	cdpProfilesListCmd.Flags().Int("page", 1, "Page number")
	cdpProfilesListCmd.Flags().Int("page-size", 20, "Results per page")

	cdpProfilesCmd.AddCommand(cdpProfilesGetCmd)

	// Events subcommands
	cdpCmd.AddCommand(cdpEventsCmd)

	cdpEventsCmd.AddCommand(cdpEventsListCmd)
	cdpEventsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	cdpEventsListCmd.Flags().String("event-type", "", "Filter by event type")
	cdpEventsListCmd.Flags().String("event-name", "", "Filter by event name")
	cdpEventsListCmd.Flags().String("source", "", "Filter by source")
	cdpEventsListCmd.Flags().Int("page", 1, "Page number")
	cdpEventsListCmd.Flags().Int("page-size", 20, "Results per page")

	cdpEventsCmd.AddCommand(cdpEventsGetCmd)

	// Segments subcommands
	cdpCmd.AddCommand(cdpSegmentsCmd)

	cdpSegmentsCmd.AddCommand(cdpSegmentsListCmd)
	cdpSegmentsListCmd.Flags().String("type", "", "Filter by segment type")
	cdpSegmentsListCmd.Flags().String("status", "", "Filter by status")
	cdpSegmentsListCmd.Flags().Int("page", 1, "Page number")
	cdpSegmentsListCmd.Flags().Int("page-size", 20, "Results per page")

	cdpSegmentsCmd.AddCommand(cdpSegmentsGetCmd)

	schema.Register(schema.Resource{
		Name:        "cdp",
		Description: "Access Customer Data Platform analytics (profiles, events, segments)",
		Commands:    []string{"profiles list", "profiles get", "events list", "events get", "segments list", "segments get"},
		IDPrefix:    "cdp",
	})
}

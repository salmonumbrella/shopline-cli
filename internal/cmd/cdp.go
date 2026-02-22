package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
				outfmt.FormatID("cdp_property", p.ID),
				p.Email,
				fmt.Sprintf("%d", p.TotalOrders),
				p.TotalSpent,
				p.LifetimeValue,
				p.ChurnRisk,
				segments,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d profiles\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Profile ID:       %s\n", profile.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:      %s\n", profile.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:            %s\n", profile.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Phone:            %s\n", profile.Phone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:             %s %s\n", profile.FirstName, profile.LastName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Orders:     %d\n", profile.TotalOrders)
		_, _ = fmt.Fprintf(outWriter(cmd), "Total Spent:      %s\n", profile.TotalSpent)
		_, _ = fmt.Fprintf(outWriter(cmd), "Avg Order Value:  %s\n", profile.AverageOrderValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Lifetime Value:   %s\n", profile.LifetimeValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Predicted LTV:    %s\n", profile.PredictedLTV)
		_, _ = fmt.Fprintf(outWriter(cmd), "Churn Risk:       %s\n", profile.ChurnRisk)
		_, _ = fmt.Fprintf(outWriter(cmd), "Segments:         %s\n", strings.Join(profile.Segments, ", "))
		_, _ = fmt.Fprintf(outWriter(cmd), "Tags:             %s\n", strings.Join(profile.Tags, ", "))

		if profile.RFMScore != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nRFM Analysis:\n")
			_, _ = fmt.Fprintf(outWriter(cmd), "  Recency:        %d\n", profile.RFMScore.Recency)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Frequency:      %d\n", profile.RFMScore.Frequency)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Monetary:       %d\n", profile.RFMScore.Monetary)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Total Score:    %d\n", profile.RFMScore.Total)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Segment:        %s\n", profile.RFMScore.Segment)
		}

		if profile.Preferences != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nPreferences:\n")
			_, _ = fmt.Fprintf(outWriter(cmd), "  Email Marketing:  %v\n", profile.Preferences.EmailMarketing)
			_, _ = fmt.Fprintf(outWriter(cmd), "  SMS Marketing:    %v\n", profile.Preferences.SMSMarketing)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Push Notifications: %v\n", profile.Preferences.PushNotifications)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Preferred Channel: %s\n", profile.Preferences.PreferredChannel)
		}

		if profile.FirstOrderAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nFirst Order:      %s\n", profile.FirstOrderAt.Format(time.RFC3339))
		}
		if profile.LastOrderAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Last Order:       %s\n", profile.LastOrderAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:          %s\n", profile.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", profile.UpdatedAt.Format(time.RFC3339))

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
				outfmt.FormatID("cdp_event", e.ID),
				e.CustomerID,
				e.EventType,
				e.EventName,
				e.Source,
				e.Channel,
				e.Timestamp.Format("2006-01-02 15:04:05"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d events\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Event ID:     %s\n", event.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:  %s\n", event.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Session ID:   %s\n", event.SessionID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Event Type:   %s\n", event.EventType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Event Name:   %s\n", event.EventName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Source:       %s\n", event.Source)
		_, _ = fmt.Fprintf(outWriter(cmd), "Channel:      %s\n", event.Channel)
		_, _ = fmt.Fprintf(outWriter(cmd), "Timestamp:    %s\n", event.Timestamp.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", event.CreatedAt.Format(time.RFC3339))

		if len(event.Properties) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nProperties:\n")
			for k, v := range event.Properties {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s: %v\n", k, v)
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
				outfmt.FormatID("cdp_segment", s.ID),
				s.Name,
				s.Type,
				fmt.Sprintf("%d", s.CustomerCount),
				s.Status,
				s.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d segments\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Segment ID:      %s\n", segment.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", segment.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", segment.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:            %s\n", segment.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", segment.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Count:  %d\n", segment.CustomerCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", segment.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", segment.UpdatedAt.Format(time.RFC3339))

		if len(segment.Conditions) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nConditions:\n")
			for i, c := range segment.Conditions {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %d. %s %s %v\n", i+1, c.Field, c.Operator, c.Value)
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

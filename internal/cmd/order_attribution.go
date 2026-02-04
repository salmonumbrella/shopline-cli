package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var orderAttributionCmd = &cobra.Command{
	Use:   "order-attribution",
	Short: "Manage order attribution tracking",
}

var orderAttributionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List order attributions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		source, _ := cmd.Flags().GetString("source")
		medium, _ := cmd.Flags().GetString("medium")
		campaign, _ := cmd.Flags().GetString("campaign")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.OrderAttributionListOptions{
			Page:     page,
			PageSize: pageSize,
			Source:   source,
			Medium:   medium,
			Campaign: campaign,
		}
		if from != "" {
			since, err := parseTimeFlag(from, "from")
			if err != nil {
				return err
			}
			opts.Since = since
		}
		if to != "" {
			until, err := parseTimeFlag(to, "to")
			if err != nil {
				return err
			}
			opts.Until = until
		}

		resp, err := client.ListOrderAttributions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list order attributions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ORDER", "SOURCE", "MEDIUM", "CAMPAIGN", "TOUCHPOINTS", "CREATED"}
		var rows [][]string
		for _, a := range resp.Items {
			rows = append(rows, []string{
				a.ID,
				a.OrderID,
				a.Source,
				a.Medium,
				a.Campaign,
				fmt.Sprintf("%d", a.TouchpointCount),
				a.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d attributions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var orderAttributionGetCmd = &cobra.Command{
	Use:   "get <order-id>",
	Short: "Get attribution for an order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		attribution, err := client.GetOrderAttribution(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get order attribution: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(attribution)
		}

		fmt.Printf("Attribution ID:   %s\n", attribution.ID)
		fmt.Printf("Order ID:         %s\n", attribution.OrderID)
		fmt.Printf("\n--- Traffic Source ---\n")
		fmt.Printf("Source:           %s\n", attribution.Source)
		fmt.Printf("Medium:           %s\n", attribution.Medium)
		if attribution.Campaign != "" {
			fmt.Printf("Campaign:         %s\n", attribution.Campaign)
		}
		if attribution.Content != "" {
			fmt.Printf("Content:          %s\n", attribution.Content)
		}
		if attribution.Term != "" {
			fmt.Printf("Term:             %s\n", attribution.Term)
		}
		fmt.Printf("\n--- UTM Parameters ---\n")
		if attribution.UtmSource != "" {
			fmt.Printf("utm_source:       %s\n", attribution.UtmSource)
		}
		if attribution.UtmMedium != "" {
			fmt.Printf("utm_medium:       %s\n", attribution.UtmMedium)
		}
		if attribution.UtmCampaign != "" {
			fmt.Printf("utm_campaign:     %s\n", attribution.UtmCampaign)
		}
		if attribution.UtmContent != "" {
			fmt.Printf("utm_content:      %s\n", attribution.UtmContent)
		}
		if attribution.UtmTerm != "" {
			fmt.Printf("utm_term:         %s\n", attribution.UtmTerm)
		}
		fmt.Printf("\n--- Journey ---\n")
		if attribution.ReferrerURL != "" {
			fmt.Printf("Referrer:         %s\n", attribution.ReferrerURL)
		}
		if attribution.LandingPage != "" {
			fmt.Printf("Landing Page:     %s\n", attribution.LandingPage)
		}
		fmt.Printf("Touchpoints:      %d\n", attribution.TouchpointCount)
		if attribution.FirstVisitAt != nil {
			fmt.Printf("First Visit:      %s\n", attribution.FirstVisitAt.Format(time.RFC3339))
		}
		if attribution.LastVisitAt != nil {
			fmt.Printf("Last Visit:       %s\n", attribution.LastVisitAt.Format(time.RFC3339))
		}
		fmt.Printf("\nCreated:          %s\n", attribution.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orderAttributionCmd)

	orderAttributionCmd.AddCommand(orderAttributionListCmd)
	orderAttributionListCmd.Flags().String("source", "", "Filter by source (e.g., google, facebook)")
	orderAttributionListCmd.Flags().String("medium", "", "Filter by medium (e.g., cpc, organic, social)")
	orderAttributionListCmd.Flags().String("campaign", "", "Filter by campaign")
	orderAttributionListCmd.Flags().String("from", "", "Filter by created date from (YYYY-MM-DD or RFC3339)")
	orderAttributionListCmd.Flags().String("to", "", "Filter by created date to (YYYY-MM-DD or RFC3339)")
	orderAttributionListCmd.Flags().Int("page", 1, "Page number")
	orderAttributionListCmd.Flags().Int("page-size", 20, "Results per page")

	orderAttributionCmd.AddCommand(orderAttributionGetCmd)

	schema.Register(schema.Resource{
		Name:        "order-attribution",
		Description: "Manage order attribution tracking",
		Commands:    []string{"list", "get"},
		IDPrefix:    "attribution",
	})
}

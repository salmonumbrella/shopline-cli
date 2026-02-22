package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var shippingZonesCmd = &cobra.Command{
	Use:   "shipping-zones",
	Short: "Manage shipping zones",
}

var shippingZonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shipping zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ShippingZonesListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListShippingZones(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list shipping zones: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "COUNTRIES", "PRICE RATES", "WEIGHT RATES", "CREATED"}
		var rows [][]string
		for _, z := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("shipping_zone", z.ID),
				z.Name,
				fmt.Sprintf("%d", len(z.Countries)),
				fmt.Sprintf("%d", len(z.PriceBasedRates)),
				fmt.Sprintf("%d", len(z.WeightBasedRates)),
				z.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d shipping zones\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var shippingZonesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get shipping zone details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		zone, err := client.GetShippingZone(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get shipping zone: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(zone)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Zone ID:  %s\n", zone.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:     %s\n", zone.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:  %s\n", zone.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:  %s\n", zone.UpdatedAt.Format(time.RFC3339))

		if len(zone.Countries) > 0 {
			_, _ = fmt.Fprintln(outWriter(cmd), "\nCountries:")
			for _, c := range zone.Countries {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s - %s\n", c.Code, c.Name)
			}
		}

		if len(zone.PriceBasedRates) > 0 {
			_, _ = fmt.Fprintln(outWriter(cmd), "\nPrice-Based Rates:")
			for _, r := range zone.PriceBasedRates {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s: %s (%s - %s)\n", r.Name, r.Price, r.MinValue, r.MaxValue)
			}
		}

		if len(zone.WeightBasedRates) > 0 {
			_, _ = fmt.Fprintln(outWriter(cmd), "\nWeight-Based Rates:")
			for _, r := range zone.WeightBasedRates {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s: %s (%.2f - %.2f kg)\n", r.Name, r.Price, r.MinWeight, r.MaxWeight)
			}
		}
		return nil
	},
}

var shippingZonesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a shipping zone",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create shipping zone") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")

		req := &api.ShippingZoneCreateRequest{
			Name: name,
		}

		zone, err := client.CreateShippingZone(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create shipping zone: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(zone)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created shipping zone %s\n", zone.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", zone.Name)
		return nil
	},
}

var shippingZonesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a shipping zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete shipping zone %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete shipping zone %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteShippingZone(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete shipping zone: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted shipping zone %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shippingZonesCmd)

	shippingZonesCmd.AddCommand(shippingZonesListCmd)
	shippingZonesListCmd.Flags().Int("page", 1, "Page number")
	shippingZonesListCmd.Flags().Int("page-size", 20, "Results per page")

	shippingZonesCmd.AddCommand(shippingZonesGetCmd)

	shippingZonesCmd.AddCommand(shippingZonesCreateCmd)
	shippingZonesCreateCmd.Flags().String("name", "", "Shipping zone name")
	_ = shippingZonesCreateCmd.MarkFlagRequired("name")

	shippingZonesCmd.AddCommand(shippingZonesDeleteCmd)
}

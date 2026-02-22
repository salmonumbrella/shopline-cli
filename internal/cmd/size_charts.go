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

var sizeChartsCmd = &cobra.Command{
	Use:   "size-charts",
	Short: "Manage size charts",
}

var sizeChartsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List size charts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.SizeChartsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListSizeCharts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list size charts: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "UNIT", "ROWS", "ACTIVE", "CREATED"}
		var rows [][]string
		for _, sc := range resp.Items {
			active := "No"
			if sc.Active {
				active = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("size_chart", sc.ID),
				sc.Name,
				sc.Unit,
				fmt.Sprintf("%d", len(sc.Rows)),
				active,
				sc.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d size charts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var sizeChartsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get size chart details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		sizeChart, err := client.GetSizeChart(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get size chart: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(sizeChart)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Size Chart ID:  %s\n", sizeChart.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", sizeChart.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", sizeChart.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Unit:           %s\n", sizeChart.Unit)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:         %t\n", sizeChart.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", sizeChart.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", sizeChart.UpdatedAt.Format(time.RFC3339))

		if len(sizeChart.Headers) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nHeaders: %s\n", strings.Join(sizeChart.Headers, ", "))
		}

		if len(sizeChart.Rows) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nSizes:\n")
			for _, row := range sizeChart.Rows {
				_, _ = fmt.Fprintf(outWriter(cmd), "  %s: %s\n", row.Size, strings.Join(row.Values, ", "))
			}
		}

		if len(sizeChart.ProductIDs) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nAssociated Products: %d\n", len(sizeChart.ProductIDs))
		}

		return nil
	},
}

var sizeChartsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a size chart",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		unit, _ := cmd.Flags().GetString("unit")
		active, _ := cmd.Flags().GetBool("active")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create size chart '%s'", name)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.SizeChartCreateRequest{
			Name:        name,
			Description: description,
			Unit:        unit,
			Active:      active,
		}

		sizeChart, err := client.CreateSizeChart(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create size chart: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(sizeChart)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created size chart %s\n", sizeChart.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", sizeChart.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Unit:   %s\n", sizeChart.Unit)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", sizeChart.Active)

		return nil
	},
}

var sizeChartsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a size chart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update size chart %s", args[0])) {
			return nil
		}

		var req api.SizeChartUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		sizeChart, err := client.UpdateSizeChart(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update size chart: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(sizeChart)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated size chart %s\n", sizeChart.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", sizeChart.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Unit:   %s\n", sizeChart.Unit)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", sizeChart.Active)
		return nil
	},
}

var sizeChartsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a size chart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete size chart %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete size chart %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteSizeChart(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete size chart: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted size chart %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sizeChartsCmd)

	sizeChartsCmd.AddCommand(sizeChartsListCmd)
	sizeChartsListCmd.Flags().Int("page", 1, "Page number")
	sizeChartsListCmd.Flags().Int("page-size", 20, "Results per page")

	sizeChartsCmd.AddCommand(sizeChartsGetCmd)

	sizeChartsCmd.AddCommand(sizeChartsCreateCmd)
	sizeChartsCreateCmd.Flags().String("name", "", "Size chart name (required)")
	sizeChartsCreateCmd.Flags().String("description", "", "Size chart description")
	sizeChartsCreateCmd.Flags().String("unit", "cm", "Measurement unit (cm, inches, etc.)")
	sizeChartsCreateCmd.Flags().Bool("active", true, "Size chart active status")
	_ = sizeChartsCreateCmd.MarkFlagRequired("name")

	sizeChartsCmd.AddCommand(sizeChartsUpdateCmd)
	addJSONBodyFlags(sizeChartsUpdateCmd)
	sizeChartsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	sizeChartsCmd.AddCommand(sizeChartsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "size-charts",
		Description: "Manage size charts",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "size_chart",
	})
}

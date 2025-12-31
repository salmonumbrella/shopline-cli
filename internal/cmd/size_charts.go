package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				sc.ID,
				sc.Name,
				sc.Unit,
				fmt.Sprintf("%d", len(sc.Rows)),
				active,
				sc.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d size charts\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Size Chart ID:  %s\n", sizeChart.ID)
		fmt.Printf("Name:           %s\n", sizeChart.Name)
		fmt.Printf("Description:    %s\n", sizeChart.Description)
		fmt.Printf("Unit:           %s\n", sizeChart.Unit)
		fmt.Printf("Active:         %t\n", sizeChart.Active)
		fmt.Printf("Created:        %s\n", sizeChart.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", sizeChart.UpdatedAt.Format(time.RFC3339))

		if len(sizeChart.Headers) > 0 {
			fmt.Printf("\nHeaders: %s\n", strings.Join(sizeChart.Headers, ", "))
		}

		if len(sizeChart.Rows) > 0 {
			fmt.Printf("\nSizes:\n")
			for _, row := range sizeChart.Rows {
				fmt.Printf("  %s: %s\n", row.Size, strings.Join(row.Values, ", "))
			}
		}

		if len(sizeChart.ProductIDs) > 0 {
			fmt.Printf("\nAssociated Products: %d\n", len(sizeChart.ProductIDs))
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create size chart '%s'\n", name)
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

		fmt.Printf("Created size chart %s\n", sizeChart.ID)
		fmt.Printf("Name:   %s\n", sizeChart.Name)
		fmt.Printf("Unit:   %s\n", sizeChart.Unit)
		fmt.Printf("Active: %t\n", sizeChart.Active)

		return nil
	},
}

var sizeChartsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a size chart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete size chart %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete size chart %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteSizeChart(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete size chart: %w", err)
		}

		fmt.Printf("Deleted size chart %s\n", args[0])
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

	sizeChartsCmd.AddCommand(sizeChartsDeleteCmd)
}

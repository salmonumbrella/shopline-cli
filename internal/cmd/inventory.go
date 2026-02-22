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

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Manage inventory levels",
}

var inventoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List inventory levels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		locationID, _ := cmd.Flags().GetString("location-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.InventoryListOptions{
			LocationID: locationID,
			Page:       page,
			PageSize:   pageSize,
		}

		resp, err := client.ListInventoryLevels(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list inventory levels: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "ITEM ID", "LOCATION ID", "AVAILABLE", "UPDATED"}
		var rows [][]string
		for _, l := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("inventory_item", l.ID),
				l.InventoryItemID,
				l.LocationID,
				strconv.Itoa(l.Available),
				l.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d inventory levels\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var inventoryGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get inventory level details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		level, err := client.GetInventoryLevel(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get inventory level: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(level)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Inventory ID:     %s\n", level.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Inventory Item:   %s\n", level.InventoryItemID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Location:         %s\n", level.LocationID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:        %d\n", level.Available)
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", level.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var inventoryAdjustCmd = &cobra.Command{
	Use:   "adjust <id>",
	Short: "Adjust inventory quantity",
	Long:  `Adjust the available quantity of an inventory level by a delta value (positive or negative).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		delta, _ := cmd.Flags().GetInt("delta")
		if delta == 0 {
			return fmt.Errorf("--delta is required and must be non-zero")
		}

		action := "increase"
		if delta < 0 {
			action = "decrease"
		}
		if !confirmAction(cmd, fmt.Sprintf("Adjust inventory %s by %d (%s)? [y/N] ", args[0], delta, action)) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		level, err := client.AdjustInventory(cmd.Context(), args[0], delta)
		if err != nil {
			return fmt.Errorf("failed to adjust inventory: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(level)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Inventory adjusted. New available quantity: %d\n", level.Available)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inventoryCmd)

	inventoryCmd.AddCommand(inventoryListCmd)
	inventoryListCmd.Flags().String("location-id", "", "Filter by location ID")
	inventoryListCmd.Flags().Int("page", 1, "Page number")
	inventoryListCmd.Flags().Int("page-size", 20, "Results per page")

	inventoryCmd.AddCommand(inventoryGetCmd)

	inventoryCmd.AddCommand(inventoryAdjustCmd)
	inventoryAdjustCmd.Flags().Int("delta", 0, "Quantity adjustment (positive or negative)")
	_ = inventoryAdjustCmd.MarkFlagRequired("delta")

	schema.Register(schema.Resource{
		Name:        "inventory",
		Description: "Manage inventory levels",
		Commands:    []string{"list", "get", "adjust"},
		IDPrefix:    "inventory",
	})
}

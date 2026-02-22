package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var inventoryLevelsCmd = &cobra.Command{
	Use:   "inventory-levels",
	Short: "Manage inventory levels",
}

var inventoryLevelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List inventory levels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		locationID, _ := cmd.Flags().GetString("location-id")

		opts := &api.InventoryListOptions{
			Page:       page,
			PageSize:   pageSize,
			LocationID: locationID,
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

		headers := []string{"ID", "INVENTORY ITEM", "LOCATION", "AVAILABLE", "RESERVED", "INCOMING", "ON HAND", "UPDATED"}
		var rows [][]string
		for _, l := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("inventory_level", l.ID),
				l.InventoryItemID,
				l.LocationID,
				fmt.Sprintf("%d", l.Available),
				fmt.Sprintf("%d", l.Reserved),
				fmt.Sprintf("%d", l.Incoming),
				fmt.Sprintf("%d", l.OnHand),
				l.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d inventory levels\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var inventoryLevelsGetCmd = &cobra.Command{
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

		_, _ = fmt.Fprintf(outWriter(cmd), "ID:              %s\n", level.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Inventory Item:  %s\n", level.InventoryItemID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Location:        %s\n", level.LocationID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:       %d\n", level.Available)
		_, _ = fmt.Fprintf(outWriter(cmd), "Reserved:        %d\n", level.Reserved)
		_, _ = fmt.Fprintf(outWriter(cmd), "Incoming:        %d\n", level.Incoming)
		_, _ = fmt.Fprintf(outWriter(cmd), "On Hand:         %d\n", level.OnHand)
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", level.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var inventoryLevelsAdjustCmd = &cobra.Command{
	Use:   "adjust",
	Short: "Adjust inventory level",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		inventoryItemID, _ := cmd.Flags().GetString("inventory-item-id")
		locationID, _ := cmd.Flags().GetString("location-id")
		adjustment, _ := cmd.Flags().GetInt("adjustment")

		req := &api.InventoryLevelAdjustRequest{
			InventoryItemID:     inventoryItemID,
			LocationID:          locationID,
			AvailableAdjustment: adjustment,
		}

		level, err := client.AdjustInventoryLevel(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to adjust inventory level: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(level)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Adjusted inventory level %s\n", level.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available: %d\n", level.Available)
		_, _ = fmt.Fprintf(outWriter(cmd), "On Hand:   %d\n", level.OnHand)
		return nil
	},
}

var inventoryLevelsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set inventory level",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		inventoryItemID, _ := cmd.Flags().GetString("inventory-item-id")
		locationID, _ := cmd.Flags().GetString("location-id")
		available, _ := cmd.Flags().GetInt("available")

		req := &api.InventoryLevelSetRequest{
			InventoryItemID: inventoryItemID,
			LocationID:      locationID,
			Available:       available,
		}

		level, err := client.SetInventoryLevel(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to set inventory level: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(level)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Set inventory level %s\n", level.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available: %d\n", level.Available)
		_, _ = fmt.Fprintf(outWriter(cmd), "On Hand:   %d\n", level.OnHand)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inventoryLevelsCmd)

	inventoryLevelsCmd.AddCommand(inventoryLevelsListCmd)
	inventoryLevelsListCmd.Flags().Int("page", 1, "Page number")
	inventoryLevelsListCmd.Flags().Int("page-size", 20, "Results per page")
	inventoryLevelsListCmd.Flags().String("location-id", "", "Filter by location ID")

	inventoryLevelsCmd.AddCommand(inventoryLevelsGetCmd)

	inventoryLevelsCmd.AddCommand(inventoryLevelsAdjustCmd)
	inventoryLevelsAdjustCmd.Flags().String("inventory-item-id", "", "Inventory item ID")
	inventoryLevelsAdjustCmd.Flags().String("location-id", "", "Location ID")
	inventoryLevelsAdjustCmd.Flags().Int("adjustment", 0, "Quantity adjustment (positive or negative)")
	_ = inventoryLevelsAdjustCmd.MarkFlagRequired("inventory-item-id")
	_ = inventoryLevelsAdjustCmd.MarkFlagRequired("location-id")
	_ = inventoryLevelsAdjustCmd.MarkFlagRequired("adjustment")

	inventoryLevelsCmd.AddCommand(inventoryLevelsSetCmd)
	inventoryLevelsSetCmd.Flags().String("inventory-item-id", "", "Inventory item ID")
	inventoryLevelsSetCmd.Flags().String("location-id", "", "Location ID")
	inventoryLevelsSetCmd.Flags().Int("available", 0, "Available quantity to set")
	_ = inventoryLevelsSetCmd.MarkFlagRequired("inventory-item-id")
	_ = inventoryLevelsSetCmd.MarkFlagRequired("location-id")
	_ = inventoryLevelsSetCmd.MarkFlagRequired("available")
}

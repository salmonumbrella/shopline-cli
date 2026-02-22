package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var localDeliveryCmd = &cobra.Command{
	Use:   "local-delivery",
	Short: "Manage local delivery options",
}

var localDeliveryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local delivery options",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		locationID, _ := cmd.Flags().GetString("location-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		activeFlag, _ := cmd.Flags().GetString("active")

		opts := &api.LocalDeliveryListOptions{
			Page:       page,
			PageSize:   pageSize,
			LocationID: locationID,
		}

		if activeFlag != "" {
			active := activeFlag == "true"
			opts.Active = &active
		}

		resp, err := client.ListLocalDeliveryOptions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list local delivery options: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "PRICE", "FREE_ABOVE", "ACTIVE", "ZONES", "DELIVERY_TIME"}
		var rows [][]string
		for _, o := range resp.Items {
			deliveryTime := ""
			if o.DeliveryTimeMin > 0 || o.DeliveryTimeMax > 0 {
				deliveryTime = fmt.Sprintf("%d-%d %s", o.DeliveryTimeMin, o.DeliveryTimeMax, o.DeliveryTimeUnit)
			}
			rows = append(rows, []string{
				outfmt.FormatID("local_delivery", o.ID),
				o.Name,
				o.Price + " " + o.Currency,
				o.FreeAbove,
				strconv.FormatBool(o.Active),
				fmt.Sprintf("%d", len(o.Zones)),
				deliveryTime,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d local delivery options\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var localDeliveryGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get local delivery option details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		option, err := client.GetLocalDeliveryOption(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get local delivery option: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(option)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Option ID:        %s\n", option.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:             %s\n", option.Name)
		if option.Description != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Description:      %s\n", option.Description)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:           %t\n", option.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "Price:            %s %s\n", option.Price, option.Currency)
		if option.FreeAbove != "" && option.FreeAbove != "0" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Free Above:       %s %s\n", option.FreeAbove, option.Currency)
		}
		if option.MinOrderAmount != "" && option.MinOrderAmount != "0" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Min Order:        %s %s\n", option.MinOrderAmount, option.Currency)
		}
		if option.MaxOrderAmount != "" && option.MaxOrderAmount != "0" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Max Order:        %s %s\n", option.MaxOrderAmount, option.Currency)
		}
		if option.DeliveryTimeMin > 0 || option.DeliveryTimeMax > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Delivery Time:    %d-%d %s\n", option.DeliveryTimeMin, option.DeliveryTimeMax, option.DeliveryTimeUnit)
		}
		if option.LocationID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Location ID:      %s\n", option.LocationID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:          %s\n", option.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", option.UpdatedAt.Format(time.RFC3339))

		if len(option.Zones) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nDelivery Zones (%d):\n", len(option.Zones))
			for _, zone := range option.Zones {
				if zone.Type == "zip_code" {
					_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (%s): %s\n", zone.Name, zone.Type, strings.Join(zone.ZipCodes, ", "))
				} else {
					_, _ = fmt.Fprintf(outWriter(cmd), "  - %s (%s): %.1f - %.1f km\n", zone.Name, zone.Type, zone.MinDistance, zone.MaxDistance)
				}
			}
		}
		return nil
	},
}

var localDeliveryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a local delivery option",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create local delivery option") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		price, _ := cmd.Flags().GetString("price")
		freeAbove, _ := cmd.Flags().GetString("free-above")
		active, _ := cmd.Flags().GetBool("active")
		locationID, _ := cmd.Flags().GetString("location-id")

		req := &api.LocalDeliveryCreateRequest{
			Name:        name,
			Description: description,
			Price:       price,
			FreeAbove:   freeAbove,
			Active:      active,
			LocationID:  locationID,
		}

		option, err := client.CreateLocalDeliveryOption(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create local delivery option: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(option)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created local delivery option %s: %s\n", option.ID, option.Name)
		return nil
	},
}

var localDeliveryUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a local delivery option",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update local delivery option %s", args[0])) {
			return nil
		}

		var req api.LocalDeliveryUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		option, err := client.UpdateLocalDeliveryOption(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update local delivery option: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(option)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated local delivery option %s: %s\n", option.ID, option.Name)
		return nil
	},
}

var localDeliveryDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a local delivery option",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete local delivery option %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteLocalDeliveryOption(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete local delivery option: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Local delivery option %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(localDeliveryCmd)

	localDeliveryCmd.AddCommand(localDeliveryListCmd)
	localDeliveryListCmd.Flags().String("location-id", "", "Filter by location ID")
	localDeliveryListCmd.Flags().String("active", "", "Filter by active status (true/false)")
	localDeliveryListCmd.Flags().Int("page", 1, "Page number")
	localDeliveryListCmd.Flags().Int("page-size", 20, "Results per page")

	localDeliveryCmd.AddCommand(localDeliveryGetCmd)

	localDeliveryCmd.AddCommand(localDeliveryCreateCmd)
	localDeliveryCreateCmd.Flags().String("name", "", "Delivery option name")
	localDeliveryCreateCmd.Flags().String("description", "", "Delivery option description")
	localDeliveryCreateCmd.Flags().String("price", "", "Delivery price")
	localDeliveryCreateCmd.Flags().String("free-above", "", "Free delivery above this order amount")
	localDeliveryCreateCmd.Flags().Bool("active", true, "Whether the option is active")
	localDeliveryCreateCmd.Flags().String("location-id", "", "Location ID for this delivery option")
	_ = localDeliveryCreateCmd.MarkFlagRequired("name")
	_ = localDeliveryCreateCmd.MarkFlagRequired("price")

	localDeliveryCmd.AddCommand(localDeliveryUpdateCmd)
	addJSONBodyFlags(localDeliveryUpdateCmd)
	localDeliveryUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	localDeliveryCmd.AddCommand(localDeliveryDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "local-delivery",
		Description: "Manage local delivery options",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "local_delivery_option",
	})
}

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

var deliveryOptionsCmd = &cobra.Command{
	Use:   "delivery-options",
	Short: "Manage delivery options",
}

var deliveryOptionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List delivery options",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		optType, _ := cmd.Flags().GetString("type")

		opts := &api.DeliveryOptionsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Type:     optType,
		}

		resp, err := client.ListDeliveryOptions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list delivery options: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "TYPE", "STATUS", "COUNTRIES", "CREATED"}
		var rows [][]string
		for _, d := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("delivery_option", d.ID),
				d.Name,
				d.Type,
				d.Status,
				fmt.Sprintf("%d", len(d.SupportedCountries)),
				d.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d delivery options\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var deliveryOptionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get delivery option details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		opt, err := client.GetDeliveryOption(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get delivery option: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(opt)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Option ID:    %s\n", opt.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", opt.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:         %s\n", opt.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:       %s\n", opt.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:  %s\n", opt.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", opt.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", opt.UpdatedAt.Format(time.RFC3339))

		if len(opt.SupportedCountries) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\nSupported Countries:\n  %s\n", strings.Join(opt.SupportedCountries, ", "))
		}
		return nil
	},
}

var deliveryOptionsTimeSlotsCmd = &cobra.Command{
	Use:   "time-slots <id>",
	Short: "List time slots for a delivery option",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")

		opts := &api.DeliveryTimeSlotsListOptions{
			Page:      page,
			PageSize:  pageSize,
			StartDate: startDate,
			EndDate:   endDate,
		}

		resp, err := client.ListDeliveryTimeSlots(cmd.Context(), args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to list delivery time slots: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "DATE", "START", "END", "AVAILABLE", "CAPACITY", "BOOKED"}
		var rows [][]string
		for _, s := range resp.Items {
			available := "No"
			if s.Available {
				available = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("delivery_setting", s.ID),
				s.Date,
				s.StartTime,
				s.EndTime,
				available,
				fmt.Sprintf("%d", s.Capacity),
				fmt.Sprintf("%d", s.Booked),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d time slots\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var deliveryOptionsUpdatePickupCmd = &cobra.Command{
	Use:   "update-pickup <id>",
	Short: "Update pickup store for a delivery option",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would update delivery option pickup store") {
			return nil
		}

		storeID, _ := cmd.Flags().GetString("store-id")
		storeName, _ := cmd.Flags().GetString("store-name")
		address, _ := cmd.Flags().GetString("address")
		phone, _ := cmd.Flags().GetString("phone")

		req := &api.PickupStoreUpdateRequest{
			StoreID:   storeID,
			StoreName: storeName,
			Address:   address,
			Phone:     phone,
		}

		opt, err := client.UpdateDeliveryOptionPickupStore(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update pickup store: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(opt)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated pickup store for delivery option %s\n", opt.ID)
		return nil
	},
}

var deliveryOptionsConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage delivery config (documented endpoint)",
}

var deliveryOptionsConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get delivery config (via /delivery_options/delivery_config)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		optType, _ := cmd.Flags().GetString("type")
		deliveryOptionID, _ := cmd.Flags().GetString("delivery-option-id")
		resp, err := client.GetDeliveryConfig(cmd.Context(), &api.DeliveryConfigOptions{
			Type:             optType,
			DeliveryOptionID: deliveryOptionID,
		})
		if err != nil {
			return fmt.Errorf("failed to get delivery config: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var deliveryOptionsDeliveryTimeSlotsCmd = &cobra.Command{
	Use:   "delivery-time-slots <id>",
	Short: "Get delivery time slots (documented endpoint; raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetDeliveryTimeSlotsOpenAPI(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get delivery time slots: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var deliveryOptionsStoresInfoCmd = &cobra.Command{
	Use:   "stores-info",
	Short: "Manage store pickup store info (documented endpoint)",
}

var deliveryOptionsStoresInfoUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update delivery option stores info (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update delivery option stores info") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateDeliveryOptionStoresInfo(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update delivery option stores info: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(deliveryOptionsCmd)

	deliveryOptionsCmd.AddCommand(deliveryOptionsListCmd)
	deliveryOptionsListCmd.Flags().Int("page", 1, "Page number")
	deliveryOptionsListCmd.Flags().Int("page-size", 20, "Results per page")
	deliveryOptionsListCmd.Flags().String("status", "", "Filter by status")
	deliveryOptionsListCmd.Flags().String("type", "", "Filter by type")

	deliveryOptionsCmd.AddCommand(deliveryOptionsGetCmd)

	deliveryOptionsCmd.AddCommand(deliveryOptionsTimeSlotsCmd)
	deliveryOptionsTimeSlotsCmd.Flags().Int("page", 1, "Page number")
	deliveryOptionsTimeSlotsCmd.Flags().Int("page-size", 20, "Results per page")
	deliveryOptionsTimeSlotsCmd.Flags().String("start-date", "", "Filter by start date (YYYY-MM-DD)")
	deliveryOptionsTimeSlotsCmd.Flags().String("end-date", "", "Filter by end date (YYYY-MM-DD)")

	deliveryOptionsCmd.AddCommand(deliveryOptionsUpdatePickupCmd)
	deliveryOptionsUpdatePickupCmd.Flags().String("store-id", "", "Store ID")
	deliveryOptionsUpdatePickupCmd.Flags().String("store-name", "", "Store name")
	deliveryOptionsUpdatePickupCmd.Flags().String("address", "", "Store address")
	deliveryOptionsUpdatePickupCmd.Flags().String("phone", "", "Store phone")
	_ = deliveryOptionsUpdatePickupCmd.MarkFlagRequired("store-id")

	deliveryOptionsCmd.AddCommand(deliveryOptionsConfigCmd)
	deliveryOptionsConfigCmd.AddCommand(deliveryOptionsConfigGetCmd)
	deliveryOptionsConfigGetCmd.Flags().String("type", "", "Delivery option type (required by API)")
	deliveryOptionsConfigGetCmd.Flags().String("delivery-option-id", "", "Delivery option id (required for some types, e.g. store pickup)")
	_ = deliveryOptionsConfigGetCmd.MarkFlagRequired("type")

	deliveryOptionsCmd.AddCommand(deliveryOptionsDeliveryTimeSlotsCmd)

	deliveryOptionsCmd.AddCommand(deliveryOptionsStoresInfoCmd)
	deliveryOptionsStoresInfoCmd.AddCommand(deliveryOptionsStoresInfoUpdateCmd)
	addJSONBodyFlags(deliveryOptionsStoresInfoUpdateCmd)

	schema.Register(schema.Resource{
		Name:        "delivery-options",
		Description: "Manage delivery options",
		Commands:    []string{"list", "get", "time-slots", "delivery-time-slots", "update-pickup", "config", "stores-info"},
		IDPrefix:    "delivery_option",
	})
}

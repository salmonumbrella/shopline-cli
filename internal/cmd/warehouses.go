package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var warehousesCmd = &cobra.Command{
	Use:   "warehouses",
	Short: "Manage warehouses",
}

var warehousesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List warehouses",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.WarehousesListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListWarehouses(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list warehouses: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "CODE", "CITY", "COUNTRY", "ACTIVE", "DEFAULT", "CREATED"}
		var rows [][]string
		for _, w := range resp.Items {
			active := "no"
			if w.Active {
				active = "yes"
			}
			isDefault := "no"
			if w.IsDefault {
				isDefault = "yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("warehouse", w.ID),
				w.Name,
				w.Code,
				w.City,
				w.Country,
				active,
				isDefault,
				w.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d warehouses\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var warehousesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get warehouse details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		warehouse, err := client.GetWarehouse(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get warehouse: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(warehouse)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Warehouse ID:   %s\n", warehouse.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", warehouse.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code:           %s\n", warehouse.Code)
		_, _ = fmt.Fprintf(outWriter(cmd), "Address:        %s\n", warehouse.Address1)
		if warehouse.Address2 != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "                %s\n", warehouse.Address2)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "City:           %s\n", warehouse.City)
		if warehouse.Province != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Province:       %s (%s)\n", warehouse.Province, warehouse.ProvinceCode)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Country:        %s (%s)\n", warehouse.Country, warehouse.CountryCode)
		_, _ = fmt.Fprintf(outWriter(cmd), "ZIP:            %s\n", warehouse.Zip)
		if warehouse.Phone != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Phone:          %s\n", warehouse.Phone)
		}
		if warehouse.Email != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Email:          %s\n", warehouse.Email)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:         %t\n", warehouse.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "Default:        %t\n", warehouse.IsDefault)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", warehouse.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", warehouse.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var warehousesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a warehouse",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create warehouse") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		code, _ := cmd.Flags().GetString("code")
		address, _ := cmd.Flags().GetString("address")
		city, _ := cmd.Flags().GetString("city")
		country, _ := cmd.Flags().GetString("country")
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")

		req := &api.WarehouseCreateRequest{
			Name:     name,
			Code:     code,
			Address1: address,
			City:     city,
			Country:  country,
			Phone:    phone,
			Email:    email,
		}

		warehouse, err := client.CreateWarehouse(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create warehouse: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(warehouse)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created warehouse %s\n", warehouse.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", warehouse.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code: %s\n", warehouse.Code)
		return nil
	},
}

var warehousesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a warehouse",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update warehouse %s", args[0])) {
			return nil
		}

		var req api.WarehouseUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		warehouse, err := client.UpdateWarehouse(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update warehouse: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(warehouse)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated warehouse %s\n", warehouse.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", warehouse.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Code: %s\n", warehouse.Code)
		return nil
	},
}

var warehousesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a warehouse",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete warehouse %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteWarehouse(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete warehouse: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted warehouse %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(warehousesCmd)

	warehousesCmd.AddCommand(warehousesListCmd)
	warehousesListCmd.Flags().Int("page", 1, "Page number")
	warehousesListCmd.Flags().Int("page-size", 20, "Results per page")

	warehousesCmd.AddCommand(warehousesGetCmd)

	warehousesCmd.AddCommand(warehousesCreateCmd)
	warehousesCreateCmd.Flags().String("name", "", "Warehouse name")
	warehousesCreateCmd.Flags().String("code", "", "Warehouse code")
	warehousesCreateCmd.Flags().String("address", "", "Street address")
	warehousesCreateCmd.Flags().String("city", "", "City")
	warehousesCreateCmd.Flags().String("country", "", "Country")
	warehousesCreateCmd.Flags().String("phone", "", "Phone number")
	warehousesCreateCmd.Flags().String("email", "", "Email address")
	_ = warehousesCreateCmd.MarkFlagRequired("name")
	_ = warehousesCreateCmd.MarkFlagRequired("address")
	_ = warehousesCreateCmd.MarkFlagRequired("city")
	_ = warehousesCreateCmd.MarkFlagRequired("country")

	warehousesCmd.AddCommand(warehousesUpdateCmd)
	addJSONBodyFlags(warehousesUpdateCmd)
	warehousesUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	warehousesCmd.AddCommand(warehousesDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "warehouses",
		Description: "Manage warehouses",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "warehouse",
	})
}

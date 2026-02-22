package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var fulfillmentServicesCmd = &cobra.Command{
	Use:   "fulfillment-services",
	Short: "Manage fulfillment services",
}

var fulfillmentServicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fulfillment services",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.FulfillmentServicesListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListFulfillmentServices(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list fulfillment services: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "HANDLE", "CALLBACK URL", "INVENTORY", "TRACKING", "CREATED"}
		var rows [][]string
		for _, fs := range resp.Items {
			inventory := "No"
			if fs.InventoryManagement {
				inventory = "Yes"
			}
			tracking := "No"
			if fs.TrackingSupport {
				tracking = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("fulfillment_service", fs.ID),
				fs.Name,
				fs.Handle,
				fs.CallbackURL,
				inventory,
				tracking,
				fs.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d fulfillment services\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var fulfillmentServicesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get fulfillment service details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		fs, err := client.GetFulfillmentService(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get fulfillment service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(fs)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "ID:                      %s\n", fs.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:                    %s\n", fs.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:                  %s\n", fs.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Callback URL:            %s\n", fs.CallbackURL)
		_, _ = fmt.Fprintf(outWriter(cmd), "Inventory Management:    %v\n", fs.InventoryManagement)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tracking Support:        %v\n", fs.TrackingSupport)
		_, _ = fmt.Fprintf(outWriter(cmd), "Requires Shipping Method: %v\n", fs.RequiresShippingMethod)
		_, _ = fmt.Fprintf(outWriter(cmd), "Format:                  %s\n", fs.Format)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:                 %s\n", fs.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:                 %s\n", fs.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var fulfillmentServicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a fulfillment service",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		callbackURL, _ := cmd.Flags().GetString("callback-url")
		inventoryManagement, _ := cmd.Flags().GetBool("inventory-management")
		trackingSupport, _ := cmd.Flags().GetBool("tracking-support")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create fulfillment service %q with callback URL %s", name, callbackURL)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.FulfillmentServiceCreateRequest{
			Name:                name,
			CallbackURL:         callbackURL,
			InventoryManagement: inventoryManagement,
			TrackingSupport:     trackingSupport,
		}

		fs, err := client.CreateFulfillmentService(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create fulfillment service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(fs)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created fulfillment service %s\n", fs.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", fs.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Callback URL: %s\n", fs.CallbackURL)

		return nil
	},
}

var fulfillmentServicesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a fulfillment service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update fulfillment service %s", args[0])) {
			return nil
		}

		var req api.FulfillmentServiceUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		fs, err := client.UpdateFulfillmentService(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update fulfillment service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(fs)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated fulfillment service %s\n", fs.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", fs.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Callback URL: %s\n", fs.CallbackURL)
		return nil
	},
}

var fulfillmentServicesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a fulfillment service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete fulfillment service %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")

		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete fulfillment service %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteFulfillmentService(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete fulfillment service: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted fulfillment service %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fulfillmentServicesCmd)

	fulfillmentServicesCmd.AddCommand(fulfillmentServicesListCmd)
	fulfillmentServicesListCmd.Flags().Int("page", 1, "Page number")
	fulfillmentServicesListCmd.Flags().Int("page-size", 20, "Results per page")

	fulfillmentServicesCmd.AddCommand(fulfillmentServicesGetCmd)

	fulfillmentServicesCmd.AddCommand(fulfillmentServicesCreateCmd)
	fulfillmentServicesCreateCmd.Flags().String("name", "", "Fulfillment service name")
	fulfillmentServicesCreateCmd.Flags().String("callback-url", "", "Callback URL for fulfillment requests")
	fulfillmentServicesCreateCmd.Flags().Bool("inventory-management", false, "Enable inventory management")
	fulfillmentServicesCreateCmd.Flags().Bool("tracking-support", false, "Enable tracking support")
	_ = fulfillmentServicesCreateCmd.MarkFlagRequired("name")
	_ = fulfillmentServicesCreateCmd.MarkFlagRequired("callback-url")

	fulfillmentServicesCmd.AddCommand(fulfillmentServicesUpdateCmd)
	addJSONBodyFlags(fulfillmentServicesUpdateCmd)
	fulfillmentServicesUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	fulfillmentServicesCmd.AddCommand(fulfillmentServicesDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "fulfillment-services",
		Description: "Manage fulfillment services",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "fulfillment_service",
	})
}

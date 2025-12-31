package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				fs.ID,
				fs.Name,
				fs.Handle,
				fs.CallbackURL,
				inventory,
				tracking,
				fs.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d fulfillment services\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("ID:                      %s\n", fs.ID)
		fmt.Printf("Name:                    %s\n", fs.Name)
		fmt.Printf("Handle:                  %s\n", fs.Handle)
		fmt.Printf("Callback URL:            %s\n", fs.CallbackURL)
		fmt.Printf("Inventory Management:    %v\n", fs.InventoryManagement)
		fmt.Printf("Tracking Support:        %v\n", fs.TrackingSupport)
		fmt.Printf("Requires Shipping Method: %v\n", fs.RequiresShippingMethod)
		fmt.Printf("Format:                  %s\n", fs.Format)
		fmt.Printf("Created:                 %s\n", fs.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:                 %s\n", fs.UpdatedAt.Format(time.RFC3339))

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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create fulfillment service %q with callback URL %s\n", name, callbackURL)
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

		fmt.Printf("Created fulfillment service %s\n", fs.ID)
		fmt.Printf("Name:         %s\n", fs.Name)
		fmt.Printf("Callback URL: %s\n", fs.CallbackURL)

		return nil
	},
}

var fulfillmentServicesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a fulfillment service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete fulfillment service %s\n", args[0])
			return nil
		}

		if !yes {
			fmt.Printf("Are you sure you want to delete fulfillment service %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteFulfillmentService(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete fulfillment service: %w", err)
		}

		fmt.Printf("Deleted fulfillment service %s\n", args[0])
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

	fulfillmentServicesCmd.AddCommand(fulfillmentServicesDeleteCmd)
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var carrierServicesCmd = &cobra.Command{
	Use:   "carrier-services",
	Short: "Manage carrier services",
}

var carrierServicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List carrier services",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CarrierServicesListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListCarrierServices(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list carrier services: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "TYPE", "CALLBACK URL", "ACTIVE", "DISCOVERY", "CREATED"}
		var rows [][]string
		for _, cs := range resp.Items {
			active := "No"
			if cs.Active {
				active = "Yes"
			}
			discovery := "No"
			if cs.ServiceDiscovery {
				discovery = "Yes"
			}
			rows = append(rows, []string{
				cs.ID,
				cs.Name,
				cs.CarrierServiceType,
				cs.CallbackURL,
				active,
				discovery,
				cs.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d carrier services\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var carrierServicesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get carrier service details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		cs, err := client.GetCarrierService(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get carrier service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(cs)
		}

		fmt.Printf("ID:                %s\n", cs.ID)
		fmt.Printf("Name:              %s\n", cs.Name)
		fmt.Printf("Type:              %s\n", cs.CarrierServiceType)
		fmt.Printf("Callback URL:      %s\n", cs.CallbackURL)
		fmt.Printf("Active:            %v\n", cs.Active)
		fmt.Printf("Service Discovery: %v\n", cs.ServiceDiscovery)
		fmt.Printf("Format:            %s\n", cs.Format)
		fmt.Printf("Created:           %s\n", cs.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:           %s\n", cs.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var carrierServicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a carrier service",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		callbackURL, _ := cmd.Flags().GetString("callback-url")
		serviceDiscovery, _ := cmd.Flags().GetBool("service-discovery")
		carrierType, _ := cmd.Flags().GetString("type")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create carrier service %q with callback URL %s\n", name, callbackURL)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.CarrierServiceCreateRequest{
			Name:               name,
			CallbackURL:        callbackURL,
			ServiceDiscovery:   serviceDiscovery,
			CarrierServiceType: carrierType,
		}

		cs, err := client.CreateCarrierService(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create carrier service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(cs)
		}

		fmt.Printf("Created carrier service %s\n", cs.ID)
		fmt.Printf("Name:         %s\n", cs.Name)
		fmt.Printf("Callback URL: %s\n", cs.CallbackURL)

		return nil
	},
}

var carrierServicesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a carrier service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete carrier service %s\n", args[0])
			return nil
		}

		if !yes {
			fmt.Printf("Are you sure you want to delete carrier service %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteCarrierService(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete carrier service: %w", err)
		}

		fmt.Printf("Deleted carrier service %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(carrierServicesCmd)

	carrierServicesCmd.AddCommand(carrierServicesListCmd)
	carrierServicesListCmd.Flags().Int("page", 1, "Page number")
	carrierServicesListCmd.Flags().Int("page-size", 20, "Results per page")

	carrierServicesCmd.AddCommand(carrierServicesGetCmd)

	carrierServicesCmd.AddCommand(carrierServicesCreateCmd)
	carrierServicesCreateCmd.Flags().String("name", "", "Carrier service name")
	carrierServicesCreateCmd.Flags().String("callback-url", "", "Callback URL for rate requests")
	carrierServicesCreateCmd.Flags().Bool("service-discovery", false, "Enable service discovery")
	carrierServicesCreateCmd.Flags().String("type", "api", "Carrier service type (api, legacy)")
	_ = carrierServicesCreateCmd.MarkFlagRequired("name")
	_ = carrierServicesCreateCmd.MarkFlagRequired("callback-url")

	carrierServicesCmd.AddCommand(carrierServicesDeleteCmd)
	carrierServicesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

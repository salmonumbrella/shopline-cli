package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
				outfmt.FormatID("carrier_service", cs.ID),
				cs.Name,
				cs.CarrierServiceType,
				cs.CallbackURL,
				active,
				discovery,
				cs.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d carrier services\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "ID:                %s\n", cs.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:              %s\n", cs.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:              %s\n", cs.CarrierServiceType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Callback URL:      %s\n", cs.CallbackURL)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:            %v\n", cs.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "Service Discovery: %v\n", cs.ServiceDiscovery)
		_, _ = fmt.Fprintf(outWriter(cmd), "Format:            %s\n", cs.Format)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:           %s\n", cs.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:           %s\n", cs.UpdatedAt.Format(time.RFC3339))

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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create carrier service %q with callback URL %s", name, callbackURL)) {
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created carrier service %s\n", cs.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", cs.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Callback URL: %s\n", cs.CallbackURL)

		return nil
	},
}

var carrierServicesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a carrier service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update carrier service %s", args[0])) {
			return nil
		}

		var req api.CarrierServiceUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		cs, err := client.UpdateCarrierService(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update carrier service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(cs)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated carrier service %s\n", cs.ID)
		return nil
	},
}

var carrierServicesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a carrier service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete carrier service %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")

		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete carrier service %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteCarrierService(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete carrier service: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted carrier service %s\n", args[0])
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

	carrierServicesCmd.AddCommand(carrierServicesUpdateCmd)
	addJSONBodyFlags(carrierServicesUpdateCmd)

	carrierServicesCmd.AddCommand(carrierServicesDeleteCmd)
	carrierServicesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

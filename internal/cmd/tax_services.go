package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var taxServicesCmd = &cobra.Command{
	Use:     "tax-services",
	Aliases: []string{"tax-service"},
	Short:   "Manage tax service providers",
}

var taxServicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tax services",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		provider, _ := cmd.Flags().GetString("provider")

		opts := &api.TaxServicesListOptions{
			Page:     page,
			PageSize: pageSize,
			Provider: provider,
		}

		if cmd.Flags().Changed("active") {
			active, _ := cmd.Flags().GetBool("active")
			opts.Active = &active
		}

		resp, err := client.ListTaxServices(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list tax services: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "PROVIDER", "SANDBOX", "ACTIVE", "COUNTRIES"}
		var rows [][]string
		for _, s := range resp.Items {
			sandbox := "no"
			if s.Sandbox {
				sandbox = "yes"
			}
			active := "no"
			if s.Active {
				active = "yes"
			}
			countries := strings.Join(s.Countries, ", ")
			if len(countries) > 20 {
				countries = countries[:17] + "..."
			}
			rows = append(rows, []string{
				outfmt.FormatID("tax_service", s.ID),
				s.Name,
				s.Provider,
				sandbox,
				active,
				countries,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d tax services\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var taxServicesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get tax service details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		service, err := client.GetTaxService(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get tax service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(service)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Tax Service ID: %s\n", service.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", service.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Provider:       %s\n", service.Provider)
		_, _ = fmt.Fprintf(outWriter(cmd), "Sandbox:        %t\n", service.Sandbox)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:         %t\n", service.Active)
		if service.CallbackURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Callback URL:   %s\n", service.CallbackURL)
		}
		if len(service.Countries) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Countries:      %s\n", strings.Join(service.Countries, ", "))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", service.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", service.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var taxServicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a tax service",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		provider, _ := cmd.Flags().GetString("provider")
		apiKey, _ := cmd.Flags().GetString("api-key")
		apiSecret, _ := cmd.Flags().GetString("api-secret")
		sandbox, _ := cmd.Flags().GetBool("sandbox")
		active, _ := cmd.Flags().GetBool("active")
		callbackURL, _ := cmd.Flags().GetString("callback-url")
		countriesStr, _ := cmd.Flags().GetString("countries")

		var countries []string
		if countriesStr != "" {
			countries = strings.Split(countriesStr, ",")
			for i := range countries {
				countries[i] = strings.TrimSpace(countries[i])
			}
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create tax service: %s (%s)", name, provider)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.TaxServiceCreateRequest{
			Name:        name,
			Provider:    provider,
			APIKey:      apiKey,
			APISecret:   apiSecret,
			Sandbox:     sandbox,
			Active:      active,
			CallbackURL: callbackURL,
			Countries:   countries,
		}

		service, err := client.CreateTaxService(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create tax service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(service)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created tax service %s\n", service.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:     %s\n", service.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Provider: %s\n", service.Provider)

		return nil
	},
}

var taxServicesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a tax service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.TaxServiceUpdateRequest{}

		if cmd.Flags().Changed("name") {
			req.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("api-key") {
			req.APIKey, _ = cmd.Flags().GetString("api-key")
		}
		if cmd.Flags().Changed("api-secret") {
			req.APISecret, _ = cmd.Flags().GetString("api-secret")
		}
		if cmd.Flags().Changed("sandbox") {
			v, _ := cmd.Flags().GetBool("sandbox")
			req.Sandbox = &v
		}
		if cmd.Flags().Changed("active") {
			v, _ := cmd.Flags().GetBool("active")
			req.Active = &v
		}
		if cmd.Flags().Changed("callback-url") {
			req.CallbackURL, _ = cmd.Flags().GetString("callback-url")
		}
		if cmd.Flags().Changed("countries") {
			countriesStr, _ := cmd.Flags().GetString("countries")
			countries := strings.Split(countriesStr, ",")
			for i := range countries {
				countries[i] = strings.TrimSpace(countries[i])
			}
			req.Countries = countries
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update tax service %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		service, err := client.UpdateTaxService(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update tax service: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(service)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated tax service %s\n", service.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", service.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", service.Active)

		return nil
	},
}

var taxServicesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a tax service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete tax service %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete tax service %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteTaxService(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete tax service: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted tax service %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(taxServicesCmd)

	taxServicesCmd.AddCommand(taxServicesListCmd)
	taxServicesListCmd.Flags().Int("page", 1, "Page number")
	taxServicesListCmd.Flags().Int("page-size", 20, "Results per page")
	taxServicesListCmd.Flags().String("provider", "", "Filter by provider (avalara, taxjar, etc.)")
	taxServicesListCmd.Flags().Bool("active", false, "Filter by active status")

	taxServicesCmd.AddCommand(taxServicesGetCmd)

	taxServicesCmd.AddCommand(taxServicesCreateCmd)
	taxServicesCreateCmd.Flags().String("name", "", "Tax service name (required)")
	taxServicesCreateCmd.Flags().String("provider", "", "Provider (avalara, taxjar, etc.) (required)")
	taxServicesCreateCmd.Flags().String("api-key", "", "API key (required)")
	taxServicesCreateCmd.Flags().String("api-secret", "", "API secret")
	taxServicesCreateCmd.Flags().Bool("sandbox", false, "Use sandbox mode")
	taxServicesCreateCmd.Flags().Bool("active", true, "Activate the service")
	taxServicesCreateCmd.Flags().String("callback-url", "", "Callback URL")
	taxServicesCreateCmd.Flags().String("countries", "", "Comma-separated country codes")
	_ = taxServicesCreateCmd.MarkFlagRequired("name")
	_ = taxServicesCreateCmd.MarkFlagRequired("provider")
	_ = taxServicesCreateCmd.MarkFlagRequired("api-key")

	taxServicesCmd.AddCommand(taxServicesUpdateCmd)
	taxServicesUpdateCmd.Flags().String("name", "", "Tax service name")
	taxServicesUpdateCmd.Flags().String("api-key", "", "API key")
	taxServicesUpdateCmd.Flags().String("api-secret", "", "API secret")
	taxServicesUpdateCmd.Flags().Bool("sandbox", false, "Use sandbox mode")
	taxServicesUpdateCmd.Flags().Bool("active", false, "Activate the service")
	taxServicesUpdateCmd.Flags().String("callback-url", "", "Callback URL")
	taxServicesUpdateCmd.Flags().String("countries", "", "Comma-separated country codes")

	taxServicesCmd.AddCommand(taxServicesDeleteCmd)
	taxServicesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

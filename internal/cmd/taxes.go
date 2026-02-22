package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var taxesCmd = &cobra.Command{
	Use:   "taxes",
	Short: "Manage tax settings",
}

var taxesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List taxes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		countryCode, _ := cmd.Flags().GetString("country")

		opts := &api.TaxesListOptions{
			Page:        page,
			PageSize:    pageSize,
			CountryCode: countryCode,
		}

		if cmd.Flags().Changed("enabled") {
			enabled, _ := cmd.Flags().GetBool("enabled")
			opts.Enabled = &enabled
		}

		resp, err := client.ListTaxes(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list taxes: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "RATE", "COUNTRY", "PROVINCE", "SHIPPING", "ENABLED"}
		var rows [][]string
		for _, t := range resp.Items {
			shipping := "no"
			if t.Shipping {
				shipping = "yes"
			}
			enabled := "no"
			if t.Enabled {
				enabled = "yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("tax", t.ID),
				t.Name,
				fmt.Sprintf("%.2f%%", t.Rate),
				t.CountryCode,
				t.ProvinceCode,
				shipping,
				enabled,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d taxes\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var taxesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get tax details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		tax, err := client.GetTax(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get tax: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tax)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Tax ID:       %s\n", tax.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", tax.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Rate:         %.2f%%\n", tax.Rate)
		_, _ = fmt.Fprintf(outWriter(cmd), "Country:      %s\n", tax.CountryCode)
		if tax.ProvinceCode != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Province:     %s\n", tax.ProvinceCode)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Priority:     %d\n", tax.Priority)
		_, _ = fmt.Fprintf(outWriter(cmd), "Compound:     %t\n", tax.Compound)
		_, _ = fmt.Fprintf(outWriter(cmd), "Shipping:     %t\n", tax.Shipping)
		_, _ = fmt.Fprintf(outWriter(cmd), "Enabled:      %t\n", tax.Enabled)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", tax.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", tax.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var taxesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a tax",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		rate, _ := cmd.Flags().GetFloat64("rate")
		countryCode, _ := cmd.Flags().GetString("country")
		provinceCode, _ := cmd.Flags().GetString("province")
		priority, _ := cmd.Flags().GetInt("priority")
		compound, _ := cmd.Flags().GetBool("compound")
		shipping, _ := cmd.Flags().GetBool("shipping")
		enabled, _ := cmd.Flags().GetBool("enabled")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create tax: %s (%.2f%%) for %s", name, rate, countryCode)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.TaxCreateRequest{
			Name:         name,
			Rate:         rate,
			CountryCode:  countryCode,
			ProvinceCode: provinceCode,
			Priority:     priority,
			Compound:     compound,
			Shipping:     shipping,
			Enabled:      enabled,
		}

		tax, err := client.CreateTax(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create tax: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tax)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created tax %s\n", tax.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:    %s\n", tax.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Rate:    %.2f%%\n", tax.Rate)
		_, _ = fmt.Fprintf(outWriter(cmd), "Country: %s\n", tax.CountryCode)

		return nil
	},
}

var taxesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a tax",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.TaxUpdateRequest{}

		if cmd.Flags().Changed("name") {
			req.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("rate") {
			rate, _ := cmd.Flags().GetFloat64("rate")
			req.Rate = &rate
		}
		if cmd.Flags().Changed("priority") {
			req.Priority, _ = cmd.Flags().GetInt("priority")
		}
		if cmd.Flags().Changed("compound") {
			v, _ := cmd.Flags().GetBool("compound")
			req.Compound = &v
		}
		if cmd.Flags().Changed("shipping") {
			v, _ := cmd.Flags().GetBool("shipping")
			req.Shipping = &v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			req.Enabled = &v
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update tax %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		tax, err := client.UpdateTax(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update tax: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tax)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated tax %s\n", tax.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", tax.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Rate: %.2f%%\n", tax.Rate)

		return nil
	},
}

var taxesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a tax",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete tax %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete tax %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteTax(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete tax: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted tax %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(taxesCmd)

	taxesCmd.AddCommand(taxesListCmd)
	taxesListCmd.Flags().Int("page", 1, "Page number")
	taxesListCmd.Flags().Int("page-size", 20, "Results per page")
	taxesListCmd.Flags().String("country", "", "Filter by country code")
	taxesListCmd.Flags().Bool("enabled", false, "Filter by enabled status")

	taxesCmd.AddCommand(taxesGetCmd)

	taxesCmd.AddCommand(taxesCreateCmd)
	taxesCreateCmd.Flags().String("name", "", "Tax name (required)")
	taxesCreateCmd.Flags().Float64("rate", 0, "Tax rate percentage (required)")
	taxesCreateCmd.Flags().String("country", "", "Country code (required)")
	taxesCreateCmd.Flags().String("province", "", "Province code")
	taxesCreateCmd.Flags().Int("priority", 1, "Tax priority")
	taxesCreateCmd.Flags().Bool("compound", false, "Compound tax")
	taxesCreateCmd.Flags().Bool("shipping", false, "Apply to shipping")
	taxesCreateCmd.Flags().Bool("enabled", true, "Enable the tax")
	_ = taxesCreateCmd.MarkFlagRequired("name")
	_ = taxesCreateCmd.MarkFlagRequired("rate")
	_ = taxesCreateCmd.MarkFlagRequired("country")

	taxesCmd.AddCommand(taxesUpdateCmd)
	taxesUpdateCmd.Flags().String("name", "", "Tax name")
	taxesUpdateCmd.Flags().Float64("rate", 0, "Tax rate percentage")
	taxesUpdateCmd.Flags().Int("priority", 0, "Tax priority")
	taxesUpdateCmd.Flags().Bool("compound", false, "Compound tax")
	taxesUpdateCmd.Flags().Bool("shipping", false, "Apply to shipping")
	taxesUpdateCmd.Flags().Bool("enabled", false, "Enable the tax")

	taxesCmd.AddCommand(taxesDeleteCmd)
	taxesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var companyCatalogsCmd = &cobra.Command{
	Use:   "company-catalogs",
	Short: "Manage B2B company catalogs",
}

var companyCatalogsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List company catalogs",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		companyID, _ := cmd.Flags().GetString("company-id")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.CompanyCatalogsListOptions{
			Page:      page,
			PageSize:  pageSize,
			CompanyID: companyID,
			Status:    status,
		}

		resp, err := client.ListCompanyCatalogs(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list company catalogs: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "COMPANY", "NAME", "PRODUCTS", "DEFAULT", "STATUS", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			isDefault := "No"
			if c.IsDefault {
				isDefault = "Yes"
			}
			createdAt := "-"
			if !c.CreatedAt.IsZero() {
				createdAt = c.CreatedAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("company_catalog", c.ID),
				c.CompanyName,
				c.Name,
				fmt.Sprintf("%d", len(c.ProductIDs)),
				isDefault,
				c.Status,
				createdAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d company catalogs\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var companyCatalogsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get company catalog details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		catalog, err := client.GetCompanyCatalog(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get company catalog: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(catalog)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Catalog ID:      %s\n", catalog.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Company ID:      %s\n", catalog.CompanyID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Company Name:    %s\n", catalog.CompanyName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", catalog.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", catalog.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Products:        %d\n", len(catalog.ProductIDs))
		if len(catalog.ProductIDs) > 0 && len(catalog.ProductIDs) <= 10 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Product IDs:     %s\n", strings.Join(catalog.ProductIDs, ", "))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Default:         %t\n", catalog.IsDefault)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", catalog.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", catalog.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", catalog.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var companyCatalogsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a company catalog",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create company catalog") {
			return nil
		}

		companyID, _ := cmd.Flags().GetString("company-id")
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		productIDs, _ := cmd.Flags().GetStringSlice("product-ids")
		isDefault, _ := cmd.Flags().GetBool("default")

		req := &api.CompanyCatalogCreateRequest{
			CompanyID:   companyID,
			Name:        name,
			Description: description,
			ProductIDs:  productIDs,
			IsDefault:   isDefault,
		}

		catalog, err := client.CreateCompanyCatalog(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create company catalog: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(catalog)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created company catalog %s (%s)\n", catalog.ID, catalog.Name)
		return nil
	},
}

var companyCatalogsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a company catalog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update company catalog %s", args[0])) {
			return nil
		}

		req := &api.CompanyCatalogUpdateRequest{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			req.Name = name
		}
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			req.Description = description
		}
		if cmd.Flags().Changed("product-ids") {
			productIDs, _ := cmd.Flags().GetStringSlice("product-ids")
			req.ProductIDs = productIDs
		}
		if cmd.Flags().Changed("default") {
			isDefault, _ := cmd.Flags().GetBool("default")
			req.IsDefault = &isDefault
		}

		catalog, err := client.UpdateCompanyCatalog(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update company catalog: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(catalog)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated company catalog %s\n", catalog.ID)
		return nil
	},
}

var companyCatalogsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a company catalog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete company catalog %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete company catalog %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCompanyCatalog(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete company catalog: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted company catalog %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(companyCatalogsCmd)

	companyCatalogsCmd.AddCommand(companyCatalogsListCmd)
	companyCatalogsListCmd.Flags().Int("page", 1, "Page number")
	companyCatalogsListCmd.Flags().Int("page-size", 20, "Results per page")
	companyCatalogsListCmd.Flags().String("company-id", "", "Filter by company ID")
	companyCatalogsListCmd.Flags().String("status", "", "Filter by status (active, inactive)")

	companyCatalogsCmd.AddCommand(companyCatalogsGetCmd)

	companyCatalogsCmd.AddCommand(companyCatalogsCreateCmd)
	companyCatalogsCreateCmd.Flags().String("company-id", "", "Company ID (required)")
	companyCatalogsCreateCmd.Flags().String("name", "", "Catalog name (required)")
	companyCatalogsCreateCmd.Flags().String("description", "", "Catalog description")
	companyCatalogsCreateCmd.Flags().StringSlice("product-ids", nil, "Product IDs (comma-separated)")
	companyCatalogsCreateCmd.Flags().Bool("default", false, "Set as default catalog")
	_ = companyCatalogsCreateCmd.MarkFlagRequired("company-id")
	_ = companyCatalogsCreateCmd.MarkFlagRequired("name")

	companyCatalogsCmd.AddCommand(companyCatalogsUpdateCmd)
	companyCatalogsUpdateCmd.Flags().String("name", "", "Catalog name")
	companyCatalogsUpdateCmd.Flags().String("description", "", "Catalog description")
	companyCatalogsUpdateCmd.Flags().StringSlice("product-ids", nil, "Product IDs (comma-separated)")
	companyCatalogsUpdateCmd.Flags().Bool("default", false, "Set as default catalog")

	companyCatalogsCmd.AddCommand(companyCatalogsDeleteCmd)
	companyCatalogsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

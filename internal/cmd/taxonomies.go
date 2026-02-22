package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var taxonomiesCmd = &cobra.Command{
	Use:   "taxonomies",
	Short: "Manage product taxonomies/categories",
}

var taxonomiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List taxonomies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		parentID, _ := cmd.Flags().GetString("parent-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.TaxonomiesListOptions{
			Page:     page,
			PageSize: pageSize,
			ParentID: parentID,
		}

		resp, err := client.ListTaxonomies(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list taxonomies: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "HANDLE", "LEVEL", "PRODUCTS", "ACTIVE", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			active := "No"
			if t.Active {
				active = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("taxonomy", t.ID),
				t.Name,
				t.Handle,
				fmt.Sprintf("%d", t.Level),
				fmt.Sprintf("%d", t.ProductCount),
				active,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d taxonomies\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var taxonomiesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get taxonomy details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		taxonomy, err := client.GetTaxonomy(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get taxonomy: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(taxonomy)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Taxonomy ID:    %s\n", taxonomy.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", taxonomy.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:         %s\n", taxonomy.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", taxonomy.Description)
		if taxonomy.ParentID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Parent ID:      %s\n", taxonomy.ParentID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Level:          %d\n", taxonomy.Level)
		_, _ = fmt.Fprintf(outWriter(cmd), "Position:       %d\n", taxonomy.Position)
		if taxonomy.Path != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Path:           %s\n", taxonomy.Path)
		}
		if taxonomy.FullPath != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Full Path:      %s\n", taxonomy.FullPath)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Product Count:  %d\n", taxonomy.ProductCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:         %t\n", taxonomy.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", taxonomy.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", taxonomy.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var taxonomiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a taxonomy",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		handle, _ := cmd.Flags().GetString("handle")
		description, _ := cmd.Flags().GetString("description")
		parentID, _ := cmd.Flags().GetString("parent-id")
		position, _ := cmd.Flags().GetInt("position")
		active, _ := cmd.Flags().GetBool("active")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create taxonomy '%s'", name)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.TaxonomyCreateRequest{
			Name:        name,
			Handle:      handle,
			Description: description,
			ParentID:    parentID,
			Position:    position,
			Active:      active,
		}

		taxonomy, err := client.CreateTaxonomy(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create taxonomy: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(taxonomy)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created taxonomy %s\n", taxonomy.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", taxonomy.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle: %s\n", taxonomy.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Level:  %d\n", taxonomy.Level)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", taxonomy.Active)

		return nil
	},
}

var taxonomiesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a taxonomy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update taxonomy %s", args[0])) {
			return nil
		}

		var req api.TaxonomyUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		taxonomy, err := client.UpdateTaxonomy(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update taxonomy: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(taxonomy)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated taxonomy %s\n", taxonomy.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", taxonomy.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle: %s\n", taxonomy.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", taxonomy.Active)
		return nil
	},
}

var taxonomiesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a taxonomy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete taxonomy %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete taxonomy %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteTaxonomy(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete taxonomy: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted taxonomy %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(taxonomiesCmd)

	taxonomiesCmd.AddCommand(taxonomiesListCmd)
	taxonomiesListCmd.Flags().String("parent-id", "", "Filter by parent taxonomy ID")
	taxonomiesListCmd.Flags().Int("page", 1, "Page number")
	taxonomiesListCmd.Flags().Int("page-size", 20, "Results per page")

	taxonomiesCmd.AddCommand(taxonomiesGetCmd)

	taxonomiesCmd.AddCommand(taxonomiesCreateCmd)
	taxonomiesCreateCmd.Flags().String("name", "", "Taxonomy name (required)")
	taxonomiesCreateCmd.Flags().String("handle", "", "URL handle (auto-generated if not provided)")
	taxonomiesCreateCmd.Flags().String("description", "", "Taxonomy description")
	taxonomiesCreateCmd.Flags().String("parent-id", "", "Parent taxonomy ID for nested categories")
	taxonomiesCreateCmd.Flags().Int("position", 0, "Position in the list")
	taxonomiesCreateCmd.Flags().Bool("active", true, "Taxonomy active status")
	_ = taxonomiesCreateCmd.MarkFlagRequired("name")

	taxonomiesCmd.AddCommand(taxonomiesUpdateCmd)
	addJSONBodyFlags(taxonomiesUpdateCmd)
	taxonomiesUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	taxonomiesCmd.AddCommand(taxonomiesDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "taxonomies",
		Description: "Manage product taxonomies/categories",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "taxonomy",
	})
}

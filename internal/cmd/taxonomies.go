package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				t.ID,
				t.Name,
				t.Handle,
				fmt.Sprintf("%d", t.Level),
				fmt.Sprintf("%d", t.ProductCount),
				active,
				t.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d taxonomies\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Taxonomy ID:    %s\n", taxonomy.ID)
		fmt.Printf("Name:           %s\n", taxonomy.Name)
		fmt.Printf("Handle:         %s\n", taxonomy.Handle)
		fmt.Printf("Description:    %s\n", taxonomy.Description)
		if taxonomy.ParentID != "" {
			fmt.Printf("Parent ID:      %s\n", taxonomy.ParentID)
		}
		fmt.Printf("Level:          %d\n", taxonomy.Level)
		fmt.Printf("Position:       %d\n", taxonomy.Position)
		if taxonomy.Path != "" {
			fmt.Printf("Path:           %s\n", taxonomy.Path)
		}
		if taxonomy.FullPath != "" {
			fmt.Printf("Full Path:      %s\n", taxonomy.FullPath)
		}
		fmt.Printf("Product Count:  %d\n", taxonomy.ProductCount)
		fmt.Printf("Active:         %t\n", taxonomy.Active)
		fmt.Printf("Created:        %s\n", taxonomy.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", taxonomy.UpdatedAt.Format(time.RFC3339))

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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create taxonomy '%s'\n", name)
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

		fmt.Printf("Created taxonomy %s\n", taxonomy.ID)
		fmt.Printf("Name:   %s\n", taxonomy.Name)
		fmt.Printf("Handle: %s\n", taxonomy.Handle)
		fmt.Printf("Level:  %d\n", taxonomy.Level)
		fmt.Printf("Active: %t\n", taxonomy.Active)

		return nil
	},
}

var taxonomiesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a taxonomy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete taxonomy %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete taxonomy %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteTaxonomy(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete taxonomy: %w", err)
		}

		fmt.Printf("Deleted taxonomy %s\n", args[0])
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

	taxonomiesCmd.AddCommand(taxonomiesDeleteCmd)
}

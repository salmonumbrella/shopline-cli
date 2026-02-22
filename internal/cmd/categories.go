package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Manage product categories",
}

var categoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		parentID, _ := cmd.Flags().GetString("parent-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CategoriesListOptions{
			Page:     page,
			PageSize: pageSize,
			ParentID: parentID,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListCategories(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list categories: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "HANDLE", "PARENT", "POSITION", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("category", c.ID),
				c.Title,
				c.Handle,
				c.ParentID,
				fmt.Sprintf("%d", c.Position),
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d categories\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var categoriesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get category details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		category, err := client.GetCategory(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get category: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(category)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Category ID:    %s\n", category.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:          %s\n", category.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:         %s\n", category.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", category.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Parent ID:      %s\n", category.ParentID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Position:       %d\n", category.Position)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", category.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", category.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var categoriesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a category",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create category") {
			return nil
		}

		title, _ := cmd.Flags().GetString("title")
		handle, _ := cmd.Flags().GetString("handle")
		description, _ := cmd.Flags().GetString("description")
		parentID, _ := cmd.Flags().GetString("parent-id")

		req := &api.CategoryCreateRequest{
			Title:       title,
			Handle:      handle,
			Description: description,
			ParentID:    parentID,
		}

		category, err := client.CreateCategory(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(category)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created category %s\n", category.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:  %s\n", category.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle: %s\n", category.Handle)
		return nil
	},
}

var categoriesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update category %s", args[0])) {
			return nil
		}

		var req api.CategoryUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		category, err := client.UpdateCategory(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update category: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(category)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated category %s\n", category.ID)
		return nil
	},
}

var categoriesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete category %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete category %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCategory(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete category: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted category %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(categoriesCmd)

	categoriesCmd.AddCommand(categoriesListCmd)
	categoriesListCmd.Flags().String("parent-id", "", "Filter by parent category ID")
	categoriesListCmd.Flags().Int("page", 1, "Page number")
	categoriesListCmd.Flags().Int("page-size", 20, "Results per page")

	categoriesCmd.AddCommand(categoriesGetCmd)

	categoriesCmd.AddCommand(categoriesCreateCmd)
	categoriesCreateCmd.Flags().String("title", "", "Category title")
	categoriesCreateCmd.Flags().String("handle", "", "Category handle (URL slug)")
	categoriesCreateCmd.Flags().String("description", "", "Category description")
	categoriesCreateCmd.Flags().String("parent-id", "", "Parent category ID")
	_ = categoriesCreateCmd.MarkFlagRequired("title")

	categoriesCmd.AddCommand(categoriesUpdateCmd)
	addJSONBodyFlags(categoriesUpdateCmd)

	categoriesCmd.AddCommand(categoriesDeleteCmd)
}

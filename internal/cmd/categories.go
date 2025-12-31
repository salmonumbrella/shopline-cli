package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
				c.ID,
				c.Title,
				c.Handle,
				c.ParentID,
				fmt.Sprintf("%d", c.Position),
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d categories\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Category ID:    %s\n", category.ID)
		fmt.Printf("Title:          %s\n", category.Title)
		fmt.Printf("Handle:         %s\n", category.Handle)
		fmt.Printf("Description:    %s\n", category.Description)
		fmt.Printf("Parent ID:      %s\n", category.ParentID)
		fmt.Printf("Position:       %d\n", category.Position)
		fmt.Printf("Created:        %s\n", category.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", category.UpdatedAt.Format(time.RFC3339))
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

		fmt.Printf("Created category %s\n", category.ID)
		fmt.Printf("Title:  %s\n", category.Title)
		fmt.Printf("Handle: %s\n", category.Handle)
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

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete category %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCategory(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete category: %w", err)
		}

		fmt.Printf("Deleted category %s\n", args[0])
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

	categoriesCmd.AddCommand(categoriesDeleteCmd)
}

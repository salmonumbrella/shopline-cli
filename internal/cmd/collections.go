package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage product collections",
}

var collectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		handle, _ := cmd.Flags().GetString("handle")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CollectionsListOptions{
			Page:     page,
			PageSize: pageSize,
			Title:    title,
			Handle:   handle,
		}

		resp, err := client.ListCollections(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list collections: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "HANDLE", "PRODUCTS", "SORT ORDER", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			rows = append(rows, []string{
				c.ID,
				c.Title,
				c.Handle,
				fmt.Sprintf("%d", c.ProductsCount),
				c.SortOrder,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d collections\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var collectionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get collection details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		collection, err := client.GetCollection(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get collection: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(collection)
		}

		fmt.Printf("Collection ID:    %s\n", collection.ID)
		fmt.Printf("Title:            %s\n", collection.Title)
		fmt.Printf("Handle:           %s\n", collection.Handle)
		fmt.Printf("Description:      %s\n", collection.Description)
		fmt.Printf("Sort Order:       %s\n", collection.SortOrder)
		fmt.Printf("Products Count:   %d\n", collection.ProductsCount)
		fmt.Printf("Published Scope:  %s\n", collection.PublishedScope)
		fmt.Printf("Published At:     %s\n", collection.PublishedAt.Format(time.RFC3339))
		fmt.Printf("Created:          %s\n", collection.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:          %s\n", collection.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var collectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a collection",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		handle, _ := cmd.Flags().GetString("handle")
		description, _ := cmd.Flags().GetString("description")
		sortOrder, _ := cmd.Flags().GetString("sort-order")

		req := &api.CollectionCreateRequest{
			Title:       title,
			Handle:      handle,
			Description: description,
			SortOrder:   sortOrder,
		}

		collection, err := client.CreateCollection(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(collection)
		}

		fmt.Printf("Created collection %s\n", collection.ID)
		fmt.Printf("Title:  %s\n", collection.Title)
		fmt.Printf("Handle: %s\n", collection.Handle)
		return nil
	},
}

var collectionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete collection %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCollection(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete collection: %w", err)
		}

		fmt.Printf("Deleted collection %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(collectionsCmd)

	collectionsCmd.AddCommand(collectionsListCmd)
	collectionsListCmd.Flags().String("title", "", "Filter by title")
	collectionsListCmd.Flags().String("handle", "", "Filter by handle")
	collectionsListCmd.Flags().Int("page", 1, "Page number")
	collectionsListCmd.Flags().Int("page-size", 20, "Results per page")

	collectionsCmd.AddCommand(collectionsGetCmd)

	collectionsCmd.AddCommand(collectionsCreateCmd)
	collectionsCreateCmd.Flags().String("title", "", "Collection title")
	collectionsCreateCmd.Flags().String("handle", "", "Collection handle (URL slug)")
	collectionsCreateCmd.Flags().String("description", "", "Collection description")
	collectionsCreateCmd.Flags().String("sort-order", "", "Product sort order (alpha-asc, alpha-desc, best-selling, etc)")
	_ = collectionsCreateCmd.MarkFlagRequired("title")

	collectionsCmd.AddCommand(collectionsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "collections",
		Description: "Manage product collections",
		Commands:    []string{"list", "get", "create", "delete"},
		IDPrefix:    "collection",
	})
}

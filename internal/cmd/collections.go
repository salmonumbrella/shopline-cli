package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
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
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
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
				outfmt.FormatID("collection", c.ID),
				c.Title,
				c.Handle,
				fmt.Sprintf("%d", c.ProductsCount),
				c.SortOrder,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d collections\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Collection ID:    %s\n", collection.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:            %s\n", collection.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:           %s\n", collection.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:      %s\n", collection.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Sort Order:       %s\n", collection.SortOrder)
		_, _ = fmt.Fprintf(outWriter(cmd), "Products Count:   %d\n", collection.ProductsCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Published Scope:  %s\n", collection.PublishedScope)
		_, _ = fmt.Fprintf(outWriter(cmd), "Published At:     %s\n", collection.PublishedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:          %s\n", collection.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:          %s\n", collection.UpdatedAt.Format(time.RFC3339))
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
		if checkDryRun(cmd, "[DRY-RUN] Would create collection") {
			return nil
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created collection %s\n", collection.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:  %s\n", collection.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle: %s\n", collection.Handle)
		return nil
	},
}

var collectionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a collection",
	Long:  "Update a collection using either --body/--body-file (raw JSON) or individual flags (--title, --handle, --description, --sort-order).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update collection %s", args[0])) {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("title") || cmd.Flags().Changed("handle") ||
			cmd.Flags().Changed("description") || cmd.Flags().Changed("sort-order")

		if hasBody && hasFlags {
			return fmt.Errorf("use either --body/--body-file or individual flags, not both")
		}
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide collection data via --body/--body-file or individual flags (--title, --handle, --description, --sort-order)")
		}

		var req api.CollectionUpdateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		} else {
			if cmd.Flags().Changed("title") {
				req.Title, _ = cmd.Flags().GetString("title")
			}
			if cmd.Flags().Changed("handle") {
				req.Handle, _ = cmd.Flags().GetString("handle")
			}
			if cmd.Flags().Changed("description") {
				req.Description, _ = cmd.Flags().GetString("description")
			}
			if cmd.Flags().Changed("sort-order") {
				req.SortOrder, _ = cmd.Flags().GetString("sort-order")
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		collection, err := client.UpdateCollection(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update collection: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(collection)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated collection %s\n", collection.ID)
		return nil
	},
}

var collectionsAddProductsCmd = &cobra.Command{
	Use:   "add-products <collection-id>",
	Short: "Add products to a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		productIDs, _ := cmd.Flags().GetStringSlice("product-id")
		if len(productIDs) == 0 {
			return fmt.Errorf("at least one --product-id is required")
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add %d products to collection %s", len(productIDs), args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.AddProductsToCollection(cmd.Context(), args[0], productIDs); err != nil {
			return fmt.Errorf("failed to add products to collection: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Added %d products to collection %s\n", len(productIDs), args[0])
		return nil
	},
}

var collectionsRemoveProductCmd = &cobra.Command{
	Use:   "remove-product <collection-id> <product-id>",
	Short: "Remove a product from a collection",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would remove product %s from collection %s", args[1], args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.RemoveProductFromCollection(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to remove product from collection: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Removed product %s from collection %s\n", args[1], args[0])
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
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete collection %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete collection %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCollection(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete collection: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted collection %s\n", args[0])
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

	collectionsCmd.AddCommand(collectionsUpdateCmd)
	addJSONBodyFlags(collectionsUpdateCmd)
	collectionsUpdateCmd.Flags().String("title", "", "Collection title")
	collectionsUpdateCmd.Flags().String("handle", "", "Collection handle (URL slug)")
	collectionsUpdateCmd.Flags().String("description", "", "Collection description")
	collectionsUpdateCmd.Flags().String("sort-order", "", "Product sort order (alpha-asc, alpha-desc, best-selling, etc)")

	collectionsCmd.AddCommand(collectionsAddProductsCmd)
	collectionsAddProductsCmd.Flags().StringSlice("product-id", nil, "Product IDs to add (repeatable)")

	collectionsCmd.AddCommand(collectionsRemoveProductCmd)

	collectionsCmd.AddCommand(collectionsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "collections",
		Description: "Manage product collections",
		Commands:    []string{"list", "get", "create", "update", "add-products", "remove-product", "delete"},
		IDPrefix:    "collection",
	})
}

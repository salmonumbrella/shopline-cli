package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var salesCmd = &cobra.Command{
	Use:   "sales",
	Short: "Manage sale campaigns",
}

var salesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sale campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.SalesListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListSales(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list sales: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "DISCOUNT", "APPLIES TO", "STATUS", "STARTS", "ENDS"}
		var rows [][]string
		for _, s := range resp.Items {
			discount := fmt.Sprintf("%.0f", s.DiscountValue)
			if s.DiscountType == "percentage" {
				discount += "%"
			}
			startsAt := "-"
			if !s.StartsAt.IsZero() {
				startsAt = s.StartsAt.Format("2006-01-02")
			}
			endsAt := "-"
			if !s.EndsAt.IsZero() {
				endsAt = s.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("sale", s.ID),
				s.Title,
				discount,
				s.AppliesTo,
				s.Status,
				startsAt,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d sales\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var salesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get sale details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		sale, err := client.GetSale(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get sale: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(sale)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Sale ID:         %s\n", sale.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:           %s\n", sale.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", sale.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:   %s\n", sale.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value:  %.2f\n", sale.DiscountValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Applies To:      %s\n", sale.AppliesTo)
		if len(sale.ProductIDs) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Products:        %s\n", strings.Join(sale.ProductIDs, ", "))
		}
		if len(sale.CollectionIDs) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Collections:     %s\n", strings.Join(sale.CollectionIDs, ", "))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", sale.Status)
		if !sale.StartsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:       %s\n", sale.StartsAt.Format(time.RFC3339))
		}
		if !sale.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:         %s\n", sale.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", sale.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var salesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a sale campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create sale") {
			return nil
		}

		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		discountType, _ := cmd.Flags().GetString("discount-type")
		discountValue, _ := cmd.Flags().GetFloat64("discount-value")
		appliesTo, _ := cmd.Flags().GetString("applies-to")
		productIDs, _ := cmd.Flags().GetStringSlice("product-ids")
		collectionIDs, _ := cmd.Flags().GetStringSlice("collection-ids")
		startsAtStr, _ := cmd.Flags().GetString("starts-at")
		endsAtStr, _ := cmd.Flags().GetString("ends-at")

		req := &api.SaleCreateRequest{
			Title:         title,
			Description:   description,
			DiscountType:  discountType,
			DiscountValue: discountValue,
			AppliesTo:     appliesTo,
			ProductIDs:    productIDs,
			CollectionIDs: collectionIDs,
		}

		if startsAtStr != "" {
			startsAt, err := time.Parse(time.RFC3339, startsAtStr)
			if err != nil {
				return fmt.Errorf("invalid starts-at format (use RFC3339): %w", err)
			}
			req.StartsAt = &startsAt
		}

		if endsAtStr != "" {
			endsAt, err := time.Parse(time.RFC3339, endsAtStr)
			if err != nil {
				return fmt.Errorf("invalid ends-at format (use RFC3339): %w", err)
			}
			req.EndsAt = &endsAt
		}

		sale, err := client.CreateSale(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create sale: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(sale)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created sale %s (%s)\n", sale.ID, sale.Title)
		return nil
	},
}

var salesActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a sale campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would activate sale %s", args[0])) {
			return nil
		}

		sale, err := client.ActivateSale(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate sale: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Activated sale %s (status: %s)\n", sale.ID, sale.Status)
		return nil
	},
}

var salesDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a sale campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would deactivate sale %s", args[0])) {
			return nil
		}

		sale, err := client.DeactivateSale(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate sale: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deactivated sale %s (status: %s)\n", sale.ID, sale.Status)
		return nil
	},
}

var salesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a sale campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete sale %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete sale %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteSale(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete sale: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted sale %s\n", args[0])
		return nil
	},
}

var salesDeleteProductsCmd = &cobra.Command{
	Use:   "delete-products <id>",
	Short: "Delete products from a sale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		productIDs, _ := cmd.Flags().GetStringSlice("product-ids")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete products from sale %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		err = client.DeleteSaleProducts(cmd.Context(), args[0], &api.SaleDeleteProductsRequest{
			ProductIDs: productIDs,
		})
		if err != nil {
			return fmt.Errorf("failed to delete sale products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(map[string]any{"ok": true})
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted %d product(s) from sale %s\n", len(productIDs), args[0])
		return nil
	},
}

// -- Sale products sub-commands --

var salesProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage products in a live sale",
}

var salesProductsListCmd = &cobra.Command{
	Use:   "list <sale-id>",
	Short: "List products in a sale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.GetSaleProducts(cmd.Context(), args[0], &api.SaleProductsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list sale products: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var salesProductsAddCmd = &cobra.Command{
	Use:   "add <sale-id>",
	Short: "Add products to a sale (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var req api.SaleAddProductsRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add products to sale %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.AddSaleProducts(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to add sale products: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

var salesProductsUpdateCmd = &cobra.Command{
	Use:   "update <sale-id>",
	Short: "Update products in a sale (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var req api.SaleUpdateProductsRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update products in sale %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateSaleProducts(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update sale products: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

// -- Sale comments sub-command --

var salesCommentsCmd = &cobra.Command{
	Use:   "comments <sale-id>",
	Short: "List comments from a live stream sale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.GetSaleComments(cmd.Context(), args[0], &api.SaleCommentsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list sale comments: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

// -- Sale customers sub-command --

var salesCustomersCmd = &cobra.Command{
	Use:   "customers <sale-id>",
	Short: "List customers who commented in a live stream sale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.GetSaleCustomers(cmd.Context(), args[0], &api.SaleCustomersListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list sale customers: %w", err)
		}

		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(salesCmd)

	salesCmd.AddCommand(salesListCmd)
	salesListCmd.Flags().Int("page", 1, "Page number")
	salesListCmd.Flags().Int("page-size", 20, "Results per page")
	salesListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")

	salesCmd.AddCommand(salesGetCmd)

	salesCmd.AddCommand(salesCreateCmd)
	salesCreateCmd.Flags().String("title", "", "Sale title (required)")
	salesCreateCmd.Flags().String("description", "", "Sale description")
	salesCreateCmd.Flags().String("discount-type", "", "Discount type: percentage or fixed_amount (required)")
	salesCreateCmd.Flags().Float64("discount-value", 0, "Discount value (required)")
	salesCreateCmd.Flags().String("applies-to", "all", "Applies to: all, products, collections")
	salesCreateCmd.Flags().StringSlice("product-ids", nil, "Product IDs (comma-separated)")
	salesCreateCmd.Flags().StringSlice("collection-ids", nil, "Collection IDs (comma-separated)")
	salesCreateCmd.Flags().String("starts-at", "", "Start time (RFC3339 format)")
	salesCreateCmd.Flags().String("ends-at", "", "End time (RFC3339 format)")
	_ = salesCreateCmd.MarkFlagRequired("title")
	_ = salesCreateCmd.MarkFlagRequired("discount-type")
	_ = salesCreateCmd.MarkFlagRequired("discount-value")

	salesCmd.AddCommand(salesActivateCmd)
	salesCmd.AddCommand(salesDeactivateCmd)

	salesCmd.AddCommand(salesDeleteCmd)
	salesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	salesCmd.AddCommand(salesDeleteProductsCmd)
	salesDeleteProductsCmd.Flags().StringSlice("product-ids", nil, "Product IDs (comma-separated) (required)")
	_ = salesDeleteProductsCmd.MarkFlagRequired("product-ids")

	// Products sub-commands
	salesCmd.AddCommand(salesProductsCmd)
	salesProductsCmd.AddCommand(salesProductsListCmd)
	salesProductsListCmd.Flags().Int("page", 1, "Page number")
	salesProductsListCmd.Flags().Int("page-size", 20, "Results per page")

	salesProductsCmd.AddCommand(salesProductsAddCmd)
	addJSONBodyFlags(salesProductsAddCmd)

	salesProductsCmd.AddCommand(salesProductsUpdateCmd)
	addJSONBodyFlags(salesProductsUpdateCmd)

	// Comments sub-command
	salesCmd.AddCommand(salesCommentsCmd)
	salesCommentsCmd.Flags().Int("page", 1, "Page number")
	salesCommentsCmd.Flags().Int("page-size", 20, "Results per page")

	// Customers sub-command
	salesCmd.AddCommand(salesCustomersCmd)
	salesCustomersCmd.Flags().Int("page", 1, "Page number")
	salesCustomersCmd.Flags().Int("page-size", 20, "Results per page")

	schema.Register(schema.Resource{
		Name:        "sales",
		Description: "Manage sale campaigns",
		Commands: []string{
			"list", "get", "create", "delete", "activate", "deactivate",
			"delete-products",
			"products list", "products add", "products update",
			"comments", "customers",
		},
		IDPrefix: "sale",
	})
}

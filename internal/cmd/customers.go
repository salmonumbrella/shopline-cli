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

var customersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage customers",
}

var customersSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search customers",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		q, _ := cmd.Flags().GetString("q")
		if len(args) == 1 {
			if q != "" {
				return fmt.Errorf("use either a positional query or --q, not both")
			}
			q = args[0]
		}
		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := &api.CustomerSearchOptions{
			Query:    q,
			Email:    email,
			Phone:    phone,
			Page:     page,
			PageSize: pageSize,
		}

		resp := &api.CustomersListResponse{}
		if limit > 0 {
			curPage := opts.Page
			perPage := opts.PageSize
			if perPage <= 0 || perPage > limit {
				perPage = limit
			}

			items := make([]api.Customer, 0, limit)
			totalCount := 0
			hasMore := false
			var pagination api.Pagination

			for len(items) < limit {
				pageOpts := *opts
				pageOpts.Page = curPage
				pageOpts.PageSize = perPage

				pageResp, err := client.SearchCustomers(cmd.Context(), &pageOpts)
				if err != nil {
					return fmt.Errorf("failed to search customers: %w", err)
				}
				if totalCount == 0 {
					totalCount = pageResp.TotalCount
					pagination = pageResp.Pagination
				}
				items = append(items, pageResp.Items...)
				hasMore = pageResp.HasMore

				if !pageResp.HasMore || len(pageResp.Items) == 0 {
					break
				}
				curPage++
			}

			if len(items) > limit {
				items = items[:limit]
				hasMore = true
			}

			resp.Items = items
			resp.Page = opts.Page
			resp.PageSize = perPage
			resp.TotalCount = totalCount
			resp.HasMore = hasMore
			resp.Pagination = pagination
			resp.Pagination.CurrentPage = opts.Page
			resp.Pagination.PerPage = perPage
		} else {
			r, err := client.SearchCustomers(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to search customers: %w", err)
			}
			resp = r
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "NAME", "STATE", "ORDERS", "TOTAL SPENT", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			name := strings.TrimSpace(strings.TrimSpace(c.FirstName) + " " + strings.TrimSpace(c.LastName))
			totalSpent := c.TotalSpent
			if c.Currency != "" && c.TotalSpent != "" {
				totalSpent = c.TotalSpent + " " + c.Currency
			}
			rows = append(rows, []string{
				outfmt.FormatID("customer", c.ID),
				c.Email,
				name,
				c.State,
				fmt.Sprintf("%d", c.OrdersCount),
				totalSpent,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d customers\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		state, _ := cmd.Flags().GetString("state")
		tags, _ := cmd.Flags().GetString("tags")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := &api.CustomersListOptions{
			Page:     page,
			PageSize: pageSize,
			Email:    email,
			State:    state,
			Tags:     tags,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp := &api.CustomersListResponse{}
		if limit > 0 {
			curPage := opts.Page
			perPage := opts.PageSize
			if perPage <= 0 || perPage > limit {
				perPage = limit
			}

			items := make([]api.Customer, 0, limit)
			totalCount := 0
			hasMore := false
			var pagination api.Pagination

			for len(items) < limit {
				pageOpts := *opts
				pageOpts.Page = curPage
				pageOpts.PageSize = perPage

				pageResp, err := client.ListCustomers(cmd.Context(), &pageOpts)
				if err != nil {
					return fmt.Errorf("failed to list customers: %w", err)
				}
				if totalCount == 0 {
					totalCount = pageResp.TotalCount
					pagination = pageResp.Pagination
				}
				items = append(items, pageResp.Items...)
				hasMore = pageResp.HasMore

				if !pageResp.HasMore || len(pageResp.Items) == 0 {
					break
				}
				curPage++
			}

			if len(items) > limit {
				items = items[:limit]
				hasMore = true
			}

			resp.Items = items
			resp.Page = opts.Page
			resp.PageSize = perPage
			resp.TotalCount = totalCount
			resp.HasMore = hasMore
			resp.Pagination = pagination
			resp.Pagination.CurrentPage = opts.Page
			resp.Pagination.PerPage = perPage
		} else {
			r, err := client.ListCustomers(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to list customers: %w", err)
			}
			resp = r
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "NAME", "STATE", "ORDERS", "TOTAL SPENT", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			name := c.FirstName
			if c.LastName != "" {
				if name != "" {
					name += " "
				}
				name += c.LastName
			}
			totalSpent := c.TotalSpent
			if c.Currency != "" {
				totalSpent = c.TotalSpent + " " + c.Currency
			}
			rows = append(rows, []string{
				c.ID,
				c.Email,
				name,
				c.State,
				fmt.Sprintf("%d", c.OrdersCount),
				totalSpent,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d customers\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get customer details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customer, err := client.GetCustomer(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get customer: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(customer)
		}

		name := customer.FirstName
		if customer.LastName != "" {
			if name != "" {
				name += " "
			}
			name += customer.LastName
		}

		out := outWriter(cmd)
		_, _ = fmt.Fprintf(out, "Customer ID:      %s\n", customer.ID)
		_, _ = fmt.Fprintf(out, "Email:            %s\n", customer.Email)
		_, _ = fmt.Fprintf(out, "Name:             %s\n", name)
		_, _ = fmt.Fprintf(out, "Phone:            %s\n", customer.Phone)
		_, _ = fmt.Fprintf(out, "State:            %s\n", customer.State)
		_, _ = fmt.Fprintf(out, "Accepts Marketing: %t\n", customer.AcceptsMarketing)
		_, _ = fmt.Fprintf(out, "Credit Balance:   %s\n", formatCustomerCreditBalance(customer))
		_, _ = fmt.Fprintf(out, "Subscriptions:    %s\n", formatCustomerSubscriptions(customer))
		_, _ = fmt.Fprintf(out, "Orders Count:     %d\n", customer.OrdersCount)
		_, _ = fmt.Fprintf(out, "Total Spent:      %s %s\n", customer.TotalSpent, customer.Currency)
		_, _ = fmt.Fprintf(out, "Tags:             %s\n", customer.Tags)
		_, _ = fmt.Fprintf(out, "Note:             %s\n", customer.Note)
		_, _ = fmt.Fprintf(out, "Created:          %s\n", customer.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(out, "Updated:          %s\n", customer.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var customersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a customer",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		phone, _ := cmd.Flags().GetString("phone")
		acceptsMarketing, _ := cmd.Flags().GetBool("accepts-marketing")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		note, _ := cmd.Flags().GetString("note")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create customer %s\n", email)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreateCustomer(cmd.Context(), &api.CustomerCreateRequest{
			Email:            email,
			FirstName:        firstName,
			LastName:         lastName,
			Phone:            phone,
			AcceptsMarketing: acceptsMarketing,
			Tags:             tags,
			Note:             note,
		})
		if err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}
		fmt.Printf("Created customer %s\n", resp.ID)
		fmt.Printf("Email: %s\n", resp.Email)
		return nil
	},
}

var customersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		var req api.CustomerUpdateRequest
		if cmd.Flags().Changed("email") {
			v, _ := cmd.Flags().GetString("email")
			req.Email = &v
		}
		if cmd.Flags().Changed("first-name") {
			v, _ := cmd.Flags().GetString("first-name")
			req.FirstName = &v
		}
		if cmd.Flags().Changed("last-name") {
			v, _ := cmd.Flags().GetString("last-name")
			req.LastName = &v
		}
		if cmd.Flags().Changed("phone") {
			v, _ := cmd.Flags().GetString("phone")
			req.Phone = &v
		}
		if cmd.Flags().Changed("accepts-marketing") {
			v, _ := cmd.Flags().GetBool("accepts-marketing")
			req.AcceptsMarketing = &v
		}
		if cmd.Flags().Changed("tag") {
			v, _ := cmd.Flags().GetStringSlice("tag")
			req.Tags = v
		}
		if cmd.Flags().Changed("note") {
			v, _ := cmd.Flags().GetString("note")
			req.Note = &v
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would update customer %s\n", id)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateCustomer(cmd.Context(), id, &req)
		if err != nil {
			return fmt.Errorf("failed to update customer: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}
		fmt.Printf("Updated customer %s\n", resp.ID)
		return nil
	},
}

var customersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete customer %s\n", id)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete customer %s? [y/N] ", id)
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCustomer(cmd.Context(), id); err != nil {
			return fmt.Errorf("failed to delete customer: %w", err)
		}
		fmt.Printf("Deleted customer %s\n", id)
		return nil
	},
}

var customersTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage customer tags",
}

var customersTagsSetCmd = &cobra.Command{
	Use:   "set <id>",
	Short: "Replace all customer tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		tags, _ := cmd.Flags().GetStringSlice("tag")
		resp, err := client.SetCustomerTags(cmd.Context(), args[0], tags)
		if err != nil {
			return fmt.Errorf("failed to set customer tags: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersTagsAddCmd = &cobra.Command{
	Use:   "add <id>",
	Short: "Add customer tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		tags, _ := cmd.Flags().GetStringSlice("tag")
		resp, err := client.UpdateCustomerTags(cmd.Context(), args[0], &api.CustomerTagsUpdateRequest{Add: tags})
		if err != nil {
			return fmt.Errorf("failed to add customer tags: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersTagsRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove customer tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		tags, _ := cmd.Flags().GetStringSlice("tag")
		resp, err := client.UpdateCustomerTags(cmd.Context(), args[0], &api.CustomerTagsUpdateRequest{Remove: tags})
		if err != nil {
			return fmt.Errorf("failed to remove customer tags: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersSubscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage customer subscriptions",
}

var customersSubscriptionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update customer subscriptions (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateCustomerSubscriptions(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update customer subscriptions: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersLineCmd = &cobra.Command{
	Use:   "line",
	Short: "Lookup customers by LINE ID",
}

var customersLineGetCmd = &cobra.Command{
	Use:   "get <line-id>",
	Short: "Get customer by LINE ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetLineCustomer(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get line customer: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersCouponPromotionsCmd = &cobra.Command{
	Use:   "coupon-promotions <customer-id>",
	Short: "Get coupon promotions available to a customer (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomerCouponPromotions(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get customer coupon promotions: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func formatCustomerCreditBalance(customer *api.Customer) string {
	if customer == nil || customer.CreditBalance == nil {
		return "N/A"
	}
	amount := fmt.Sprintf("%.2f", *customer.CreditBalance)
	if customer.Currency != "" {
		return amount + " " + customer.Currency
	}
	return amount
}

func formatCustomerSubscriptions(customer *api.Customer) string {
	if customer == nil || len(customer.Subscriptions) == 0 {
		return "N/A"
	}
	parts := make([]string, 0, len(customer.Subscriptions))
	for _, sub := range customer.Subscriptions {
		status := "inactive"
		if sub.IsActive {
			status = "active"
		}
		if sub.Platform != "" {
			parts = append(parts, sub.Platform+"="+status)
		} else {
			parts = append(parts, status)
		}
	}
	return strings.Join(parts, ", ")
}

func init() {
	rootCmd.AddCommand(customersCmd)

	customersCmd.AddCommand(customersSearchCmd)
	customersSearchCmd.Flags().String("q", "", "Search query (or provide as positional arg)")
	customersSearchCmd.Flags().String("email", "", "Filter by email")
	customersSearchCmd.Flags().String("phone", "", "Filter by phone")
	customersSearchCmd.Flags().Int("page", 1, "Page number")
	customersSearchCmd.Flags().Int("page-size", 20, "Results per page")

	customersCmd.AddCommand(customersListCmd)
	customersListCmd.Flags().String("email", "", "Filter by email (partial match)")
	customersListCmd.Flags().String("state", "", "Filter by state (enabled, disabled, invited)")
	customersListCmd.Flags().String("tags", "", "Filter by tags")
	customersListCmd.Flags().Int("page", 1, "Page number")
	customersListCmd.Flags().Int("page-size", 20, "Results per page")

	customersCmd.AddCommand(customersGetCmd)

	customersCmd.AddCommand(customersCreateCmd)
	customersCreateCmd.Flags().String("email", "", "Customer email")
	customersCreateCmd.Flags().String("first-name", "", "First name")
	customersCreateCmd.Flags().String("last-name", "", "Last name")
	customersCreateCmd.Flags().String("phone", "", "Phone number")
	customersCreateCmd.Flags().Bool("accepts-marketing", false, "Accepts marketing")
	customersCreateCmd.Flags().StringSlice("tag", nil, "Customer tag (repeatable)")
	customersCreateCmd.Flags().String("note", "", "Customer note")
	_ = customersCreateCmd.MarkFlagRequired("email")

	customersCmd.AddCommand(customersUpdateCmd)
	customersUpdateCmd.Flags().String("email", "", "Customer email")
	customersUpdateCmd.Flags().String("first-name", "", "First name")
	customersUpdateCmd.Flags().String("last-name", "", "Last name")
	customersUpdateCmd.Flags().String("phone", "", "Phone number")
	customersUpdateCmd.Flags().Bool("accepts-marketing", false, "Accepts marketing")
	customersUpdateCmd.Flags().StringSlice("tag", nil, "Customer tag (repeatable; replaces tags when set)")
	customersUpdateCmd.Flags().String("note", "", "Customer note")

	customersCmd.AddCommand(customersDeleteCmd)

	customersCmd.AddCommand(customersTagsCmd)
	customersTagsCmd.AddCommand(customersTagsSetCmd)
	customersTagsSetCmd.Flags().StringSlice("tag", nil, "Customer tag (repeatable)")
	customersTagsCmd.AddCommand(customersTagsAddCmd)
	customersTagsAddCmd.Flags().StringSlice("tag", nil, "Customer tag (repeatable)")
	customersTagsCmd.AddCommand(customersTagsRemoveCmd)
	customersTagsRemoveCmd.Flags().StringSlice("tag", nil, "Customer tag (repeatable)")

	customersCmd.AddCommand(customersSubscriptionsCmd)
	customersSubscriptionsCmd.AddCommand(customersSubscriptionsUpdateCmd)
	addJSONBodyFlags(customersSubscriptionsUpdateCmd)

	customersCmd.AddCommand(customersLineCmd)
	customersLineCmd.AddCommand(customersLineGetCmd)

	customersCmd.AddCommand(customersCouponPromotionsCmd)

	schema.Register(schema.Resource{
		Name:        "customers",
		Description: "Manage customer accounts",
		Commands:    []string{"list", "get", "search", "create", "update", "delete", "tags", "subscriptions", "line", "coupon-promotions", "metafields", "app-metafields", "store-credits", "membership-info", "membership-tier"},
		IDPrefix:    "customer",
	})
}

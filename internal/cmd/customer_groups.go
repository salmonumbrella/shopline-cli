package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var customerGroupsCmd = &cobra.Command{
	Use:   "customer-groups",
	Short: "Manage customer groups",
}

var customerGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customer groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CustomerGroupsListOptions{
			Page:     page,
			PageSize: pageSize,
		}
		if sortBy, sortOrder := readSortOptions(cmd); sortBy != "" {
			opts.SortBy = sortBy
			opts.SortOrder = sortOrder
		}

		resp, err := client.ListCustomerGroups(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list customer groups: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		renderCustomerGroupsTable(formatter, resp.Items, "")
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d customer groups\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customerGroupsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get customer group details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		groupID, err := resolveOrArg(cmd, args, func(query string) (string, error) {
			resp, err := client.SearchCustomerGroups(cmd.Context(), &api.CustomerGroupSearchOptions{
				Query: query, PageSize: 1,
			})
			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}
			if len(resp.Items) == 0 {
				return "", fmt.Errorf("no customer group found matching %q", query)
			}
			if len(resp.Items) > 1 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %d matches found, using first\n", len(resp.Items))
			}
			return resp.Items[0].ID, nil
		})
		if err != nil {
			return err
		}

		group, err := client.GetCustomerGroup(cmd.Context(), groupID)
		if err != nil {
			return fmt.Errorf("failed to get customer group: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(group)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Group ID:       %s\n", group.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:           %s\n", group.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:    %s\n", group.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Count: %d\n", group.CustomerCount)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", group.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", group.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var customerGroupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a customer group",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create customer group") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		req := &api.CustomerGroupCreateRequest{
			Name:        name,
			Description: description,
		}

		group, err := client.CreateCustomerGroup(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create customer group: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(group)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created customer group %s\n", group.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", group.Name)
		return nil
	},
}

var customerGroupsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a customer group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update customer group %s", args[0])) {
			return nil
		}

		var req api.CustomerGroupUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		group, err := client.UpdateCustomerGroup(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update customer group: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(group)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated customer group %s\n", group.ID)
		return nil
	},
}

var customerGroupsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a customer group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete customer group %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete customer group %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCustomerGroup(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete customer group: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted customer group %s\n", args[0])
		return nil
	},
}

var customerGroupsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search customer groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("q")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CustomerGroupSearchOptions{
			Query:    query,
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.SearchCustomerGroups(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search customer groups: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		renderCustomerGroupsTable(formatter, resp.Items, "customer_group")
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d customer groups\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

// renderCustomerGroupsTable renders a table of customer groups.
// idPrefix is used for FormatID on search commands; pass "" for list commands (auto-prefix handles it).
func renderCustomerGroupsTable(formatter *outfmt.Formatter, groups []api.CustomerGroup, idPrefix string) {
	headers := []string{"ID", "NAME", "DESCRIPTION", "CUSTOMERS", "CREATED"}
	var rows [][]string
	for _, g := range groups {
		desc := g.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		id := g.ID
		if idPrefix != "" {
			id = outfmt.FormatID(idPrefix, g.ID)
		}
		rows = append(rows, []string{
			id,
			g.Name,
			desc,
			fmt.Sprintf("%d", g.CustomerCount),
			g.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	formatter.Table(headers, rows)
}

var customerGroupsChildrenCmd = &cobra.Command{
	Use:   "children",
	Short: "Work with child customer groups (documented endpoints)",
}

var customerGroupsChildrenListCmd = &cobra.Command{
	Use:   "list <parent-id>",
	Short: "List child customer groups of a parent group (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomerGroupChildren(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to list customer group children: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customerGroupsChildrenCustomerIDsCmd = &cobra.Command{
	Use:   "customer-ids <parent-id> <child-id>",
	Short: "Get customer IDs in a child customer group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetCustomerGroupChildCustomerIDs(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get customer ids for child customer group: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		formatter = formatter.WithIDPrefix("customer")
		headers := []string{"ID"}
		rows := make([][]string, 0, len(resp.CustomerIDs))
		for _, id := range resp.CustomerIDs {
			rows = append(rows, []string{id})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d customers\n", len(resp.CustomerIDs), resp.TotalCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customerGroupsCmd)

	customerGroupsCmd.AddCommand(customerGroupsListCmd)
	customerGroupsListCmd.Flags().Int("page", 1, "Page number")
	customerGroupsListCmd.Flags().Int("page-size", 20, "Results per page")

	customerGroupsCmd.AddCommand(customerGroupsGetCmd)
	customerGroupsGetCmd.Flags().String("by", "", "Find customer group by name instead of ID")

	customerGroupsCmd.AddCommand(customerGroupsCreateCmd)
	customerGroupsCreateCmd.Flags().String("name", "", "Group name")
	customerGroupsCreateCmd.Flags().String("description", "", "Group description")
	_ = customerGroupsCreateCmd.MarkFlagRequired("name")

	customerGroupsCmd.AddCommand(customerGroupsUpdateCmd)
	addJSONBodyFlags(customerGroupsUpdateCmd)

	customerGroupsCmd.AddCommand(customerGroupsDeleteCmd)

	customerGroupsCmd.AddCommand(customerGroupsSearchCmd)
	customerGroupsSearchCmd.Flags().String("q", "", "Search query")
	customerGroupsSearchCmd.Flags().Int("page", 1, "Page number")
	customerGroupsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	customerGroupsCmd.AddCommand(customerGroupsChildrenCmd)
	customerGroupsChildrenCmd.AddCommand(customerGroupsChildrenListCmd)
	customerGroupsChildrenCmd.AddCommand(customerGroupsChildrenCustomerIDsCmd)

	schema.Register(schema.Resource{
		Name:        "customer-groups",
		Description: "Manage customer groups",
		Commands:    []string{"list", "get", "search", "create", "update", "delete", "children"},
		IDPrefix:    "customer_group",
	})
}

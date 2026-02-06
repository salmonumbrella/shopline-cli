package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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

		headers := []string{"ID", "NAME", "DESCRIPTION", "CUSTOMERS", "CREATED"}
		var rows [][]string
		for _, g := range resp.Items {
			desc := g.Description
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
			rows = append(rows, []string{
				g.ID,
				g.Name,
				desc,
				fmt.Sprintf("%d", g.CustomerCount),
				g.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d customer groups\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customerGroupsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get customer group details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		group, err := client.GetCustomerGroup(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get customer group: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(group)
		}

		fmt.Printf("Group ID:       %s\n", group.ID)
		fmt.Printf("Name:           %s\n", group.Name)
		fmt.Printf("Description:    %s\n", group.Description)
		fmt.Printf("Customer Count: %d\n", group.CustomerCount)
		fmt.Printf("Created:        %s\n", group.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", group.UpdatedAt.Format(time.RFC3339))
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

		fmt.Printf("Created customer group %s\n", group.ID)
		fmt.Printf("Name: %s\n", group.Name)
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

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete customer group %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCustomerGroup(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete customer group: %w", err)
		}

		fmt.Printf("Deleted customer group %s\n", args[0])
		return nil
	},
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
		fmt.Printf("\nShowing %d of %d customers\n", len(resp.CustomerIDs), resp.TotalCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customerGroupsCmd)

	customerGroupsCmd.AddCommand(customerGroupsListCmd)
	customerGroupsListCmd.Flags().Int("page", 1, "Page number")
	customerGroupsListCmd.Flags().Int("page-size", 20, "Results per page")

	customerGroupsCmd.AddCommand(customerGroupsGetCmd)

	customerGroupsCmd.AddCommand(customerGroupsCreateCmd)
	customerGroupsCreateCmd.Flags().String("name", "", "Group name")
	customerGroupsCreateCmd.Flags().String("description", "", "Group description")
	_ = customerGroupsCreateCmd.MarkFlagRequired("name")

	customerGroupsCmd.AddCommand(customerGroupsDeleteCmd)

	customerGroupsCmd.AddCommand(customerGroupsChildrenCmd)
	customerGroupsChildrenCmd.AddCommand(customerGroupsChildrenListCmd)
	customerGroupsChildrenCmd.AddCommand(customerGroupsChildrenCustomerIDsCmd)

	schema.Register(schema.Resource{
		Name:        "customer-groups",
		Description: "Manage customer groups",
		Commands:    []string{"list", "get", "create", "delete", "children"},
		IDPrefix:    "customer_group",
	})
}

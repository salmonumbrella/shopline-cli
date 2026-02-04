package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var customersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage customers",
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

		resp, err := client.ListCustomers(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list customers: %w", err)
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
		fmt.Printf("\nShowing %d of %d customers\n", len(resp.Items), resp.TotalCount)
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

		fmt.Printf("Customer ID:      %s\n", customer.ID)
		fmt.Printf("Email:            %s\n", customer.Email)
		fmt.Printf("Name:             %s\n", name)
		fmt.Printf("Phone:            %s\n", customer.Phone)
		fmt.Printf("State:            %s\n", customer.State)
		fmt.Printf("Accepts Marketing: %t\n", customer.AcceptsMarketing)
		fmt.Printf("Credit Balance:   %s\n", formatCustomerCreditBalance(customer))
		fmt.Printf("Subscriptions:    %s\n", formatCustomerSubscriptions(customer))
		fmt.Printf("Orders Count:     %d\n", customer.OrdersCount)
		fmt.Printf("Total Spent:      %s %s\n", customer.TotalSpent, customer.Currency)
		fmt.Printf("Tags:             %s\n", customer.Tags)
		fmt.Printf("Note:             %s\n", customer.Note)
		fmt.Printf("Created:          %s\n", customer.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:          %s\n", customer.UpdatedAt.Format(time.RFC3339))
		return nil
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

	customersCmd.AddCommand(customersListCmd)
	customersListCmd.Flags().String("email", "", "Filter by email (partial match)")
	customersListCmd.Flags().String("state", "", "Filter by state (enabled, disabled, invited)")
	customersListCmd.Flags().String("tags", "", "Filter by tags")
	customersListCmd.Flags().Int("page", 1, "Page number")
	customersListCmd.Flags().Int("page-size", 20, "Results per page")

	customersCmd.AddCommand(customersGetCmd)

	schema.Register(schema.Resource{
		Name:        "customers",
		Description: "Manage customer accounts",
		Commands:    []string{"list", "get"},
		IDPrefix:    "customer",
	})
}

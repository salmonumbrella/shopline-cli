package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var customerSavedSearchesCmd = &cobra.Command{
	Use:   "customer-saved-searches",
	Short: "Manage customer saved searches",
}

var customerSavedSearchesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customer saved searches",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CustomerSavedSearchesListOptions{
			Page:     page,
			PageSize: pageSize,
			Name:     name,
		}

		resp, err := client.ListCustomerSavedSearches(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list customer saved searches: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "QUERY", "CREATED"}
		var rows [][]string
		for _, s := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("saved_search", s.ID),
				s.Name,
				s.Query,
				s.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d saved searches\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customerSavedSearchesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get saved search details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		search, err := client.GetCustomerSavedSearch(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get saved search: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(search)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Search ID:  %s\n", search.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:       %s\n", search.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Query:      %s\n", search.Query)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:    %s\n", search.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:    %s\n", search.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var customerSavedSearchesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a customer saved search",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		query, _ := cmd.Flags().GetString("q")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create saved search: name=%s, query=%s", name, query)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.CustomerSavedSearchCreateRequest{
			Name:  name,
			Query: query,
		}

		search, err := client.CreateCustomerSavedSearch(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create saved search: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(search)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created saved search: %s\n", search.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:  %s\n", search.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Query: %s\n", search.Query)

		return nil
	},
}

var customerSavedSearchesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a customer saved search",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete saved search: %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteCustomerSavedSearch(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete saved search: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted saved search: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customerSavedSearchesCmd)

	customerSavedSearchesCmd.AddCommand(customerSavedSearchesListCmd)
	customerSavedSearchesListCmd.Flags().String("name", "", "Filter by name")
	customerSavedSearchesListCmd.Flags().Int("page", 1, "Page number")
	customerSavedSearchesListCmd.Flags().Int("page-size", 20, "Results per page")

	customerSavedSearchesCmd.AddCommand(customerSavedSearchesGetCmd)

	customerSavedSearchesCmd.AddCommand(customerSavedSearchesCreateCmd)
	customerSavedSearchesCreateCmd.Flags().String("name", "", "Search name")
	customerSavedSearchesCreateCmd.Flags().String("q", "", "Search query")
	_ = customerSavedSearchesCreateCmd.MarkFlagRequired("name")
	_ = customerSavedSearchesCreateCmd.MarkFlagRequired("q")

	customerSavedSearchesCmd.AddCommand(customerSavedSearchesDeleteCmd)
}

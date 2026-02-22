package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var customerBlacklistCmd = &cobra.Command{
	Use:   "customer-blacklist",
	Short: "Manage customer blacklist",
}

var customerBlacklistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List blacklisted customers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.CustomerBlacklistListOptions{
			Page:     page,
			PageSize: pageSize,
			Email:    email,
			Phone:    phone,
		}

		resp, err := client.ListCustomerBlacklist(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list customer blacklist: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER ID", "EMAIL", "PHONE", "REASON", "CREATED"}
		var rows [][]string
		for _, e := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("blacklist_entry", e.ID),
				e.CustomerID,
				e.Email,
				e.Phone,
				e.Reason,
				e.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d blacklist entries\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var customerBlacklistGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get blacklist entry details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		entry, err := client.GetCustomerBlacklist(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get blacklist entry: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(entry)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Entry ID:     %s\n", entry.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer ID:  %s\n", entry.CustomerID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:        %s\n", entry.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Phone:        %s\n", entry.Phone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Reason:       %s\n", entry.Reason)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", entry.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", entry.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var customerBlacklistCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a customer to blacklist",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")
		reason, _ := cmd.Flags().GetString("reason")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add to blacklist: customer=%s, email=%s, phone=%s", customerID, email, phone)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.CustomerBlacklistCreateRequest{
			CustomerID: customerID,
			Email:      email,
			Phone:      phone,
			Reason:     reason,
		}

		entry, err := client.CreateCustomerBlacklist(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to add to blacklist: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(entry)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Added to blacklist: %s\n", entry.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:  %s\n", entry.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Reason: %s\n", entry.Reason)

		return nil
	},
}

var customerBlacklistDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a customer from blacklist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would remove from blacklist: %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteCustomerBlacklist(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to remove from blacklist: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Removed from blacklist: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customerBlacklistCmd)

	customerBlacklistCmd.AddCommand(customerBlacklistListCmd)
	customerBlacklistListCmd.Flags().String("email", "", "Filter by email")
	customerBlacklistListCmd.Flags().String("phone", "", "Filter by phone")
	customerBlacklistListCmd.Flags().Int("page", 1, "Page number")
	customerBlacklistListCmd.Flags().Int("page-size", 20, "Results per page")

	customerBlacklistCmd.AddCommand(customerBlacklistGetCmd)

	customerBlacklistCmd.AddCommand(customerBlacklistCreateCmd)
	customerBlacklistCreateCmd.Flags().String("customer-id", "", "Customer ID to blacklist")
	customerBlacklistCreateCmd.Flags().String("email", "", "Email to blacklist")
	customerBlacklistCreateCmd.Flags().String("phone", "", "Phone to blacklist")
	customerBlacklistCreateCmd.Flags().String("reason", "", "Reason for blacklisting")

	customerBlacklistCmd.AddCommand(customerBlacklistDeleteCmd)
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var storeCreditsCmd = &cobra.Command{
	Use:   "store-credits",
	Short: "Manage customer store credits",
}

var storeCreditsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List store credits",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		customerID, _ := cmd.Flags().GetString("customer-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StoreCreditsListOptions{
			Page:       page,
			PageSize:   pageSize,
			CustomerID: customerID,
		}

		resp, err := client.ListStoreCredits(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list store credits: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER ID", "AMOUNT", "BALANCE", "CURRENCY", "EXPIRES", "CREATED"}
		var rows [][]string
		for _, c := range resp.Items {
			expiresAt := ""
			if !c.ExpiresAt.IsZero() {
				expiresAt = c.ExpiresAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				c.ID,
				c.CustomerID,
				c.Amount,
				c.Balance,
				c.Currency,
				expiresAt,
				c.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d store credits\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var storeCreditsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get store credit details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		credit, err := client.GetStoreCredit(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get store credit: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(credit)
		}

		expiresAt := "N/A"
		if !credit.ExpiresAt.IsZero() {
			expiresAt = credit.ExpiresAt.Format(time.RFC3339)
		}

		fmt.Printf("Credit ID:    %s\n", credit.ID)
		fmt.Printf("Customer ID:  %s\n", credit.CustomerID)
		fmt.Printf("Amount:       %s %s\n", credit.Amount, credit.Currency)
		fmt.Printf("Balance:      %s %s\n", credit.Balance, credit.Currency)
		fmt.Printf("Description:  %s\n", credit.Description)
		fmt.Printf("Expires:      %s\n", expiresAt)
		fmt.Printf("Created:      %s\n", credit.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", credit.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var storeCreditsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a store credit",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		amount, _ := cmd.Flags().GetString("amount")
		currency, _ := cmd.Flags().GetString("currency")
		description, _ := cmd.Flags().GetString("description")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create store credit: customer=%s, amount=%s %s\n", customerID, amount, currency)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StoreCreditCreateRequest{
			CustomerID:  customerID,
			Amount:      amount,
			Currency:    currency,
			Description: description,
		}

		credit, err := client.CreateStoreCredit(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create store credit: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(credit)
		}

		fmt.Printf("Created store credit: %s\n", credit.ID)
		fmt.Printf("Customer: %s\n", credit.CustomerID)
		fmt.Printf("Amount:   %s %s\n", credit.Amount, credit.Currency)

		return nil
	},
}

var storeCreditsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a store credit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete store credit: %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteStoreCredit(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete store credit: %w", err)
		}

		fmt.Printf("Deleted store credit: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storeCreditsCmd)

	storeCreditsCmd.AddCommand(storeCreditsListCmd)
	storeCreditsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	storeCreditsListCmd.Flags().Int("page", 1, "Page number")
	storeCreditsListCmd.Flags().Int("page-size", 20, "Results per page")

	storeCreditsCmd.AddCommand(storeCreditsGetCmd)

	storeCreditsCmd.AddCommand(storeCreditsCreateCmd)
	storeCreditsCreateCmd.Flags().String("customer-id", "", "Customer ID")
	storeCreditsCreateCmd.Flags().String("amount", "", "Credit amount")
	storeCreditsCreateCmd.Flags().String("currency", "USD", "Currency code")
	storeCreditsCreateCmd.Flags().String("description", "", "Description")
	_ = storeCreditsCreateCmd.MarkFlagRequired("customer-id")
	_ = storeCreditsCreateCmd.MarkFlagRequired("amount")

	storeCreditsCmd.AddCommand(storeCreditsDeleteCmd)
}

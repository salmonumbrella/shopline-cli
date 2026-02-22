package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var companyCreditsCmd = &cobra.Command{
	Use:   "company-credits",
	Short: "Manage B2B company credits",
}

var companyCreditsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List company credits",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		companyID, _ := cmd.Flags().GetString("company-id")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.CompanyCreditsListOptions{
			Page:      page,
			PageSize:  pageSize,
			CompanyID: companyID,
			Status:    status,
		}

		resp, err := client.ListCompanyCredits(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list company credits: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "COMPANY", "BALANCE", "LIMIT", "CURRENCY", "STATUS", "UPDATED"}
		var rows [][]string
		for _, c := range resp.Items {
			updatedAt := "-"
			if !c.UpdatedAt.IsZero() {
				updatedAt = c.UpdatedAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("company_credit", c.ID),
				c.CompanyName,
				fmt.Sprintf("%.2f", c.CreditBalance),
				fmt.Sprintf("%.2f", c.CreditLimit),
				c.Currency,
				c.Status,
				updatedAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d company credits\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var companyCreditsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get company credit details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		credit, err := client.GetCompanyCredit(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get company credit: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(credit)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Credit ID:       %s\n", credit.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Company ID:      %s\n", credit.CompanyID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Company Name:    %s\n", credit.CompanyName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Credit Balance:  %.2f %s\n", credit.CreditBalance, credit.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Credit Limit:    %.2f %s\n", credit.CreditLimit, credit.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:       %.2f %s\n", credit.CreditLimit-credit.CreditBalance, credit.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", credit.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", credit.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", credit.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var companyCreditsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create company credit",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create company credit") {
			return nil
		}

		companyID, _ := cmd.Flags().GetString("company-id")
		creditLimit, _ := cmd.Flags().GetFloat64("credit-limit")
		currency, _ := cmd.Flags().GetString("currency")

		req := &api.CompanyCreditCreateRequest{
			CompanyID:   companyID,
			CreditLimit: creditLimit,
			Currency:    currency,
		}

		credit, err := client.CreateCompanyCredit(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create company credit: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(credit)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created company credit %s (limit: %.2f %s)\n", credit.ID, credit.CreditLimit, credit.Currency)
		return nil
	},
}

var companyCreditsAdjustCmd = &cobra.Command{
	Use:   "adjust <id>",
	Short: "Adjust company credit balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would adjust company credit %s", args[0])) {
			return nil
		}

		amount, _ := cmd.Flags().GetFloat64("amount")
		description, _ := cmd.Flags().GetString("description")
		referenceID, _ := cmd.Flags().GetString("reference-id")

		req := &api.CompanyCreditAdjustRequest{
			Amount:      amount,
			Description: description,
			ReferenceID: referenceID,
		}

		credit, err := client.AdjustCompanyCredit(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to adjust company credit: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(credit)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Adjusted credit %s by %.2f (new balance: %.2f)\n", credit.ID, amount, credit.CreditBalance)
		return nil
	},
}

var companyCreditsTransactionsCmd = &cobra.Command{
	Use:   "transactions <id>",
	Short: "List credit transactions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListCompanyCreditTransactions(cmd.Context(), args[0], page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list credit transactions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TYPE", "AMOUNT", "BALANCE", "DESCRIPTION", "REFERENCE", "DATE"}
		var rows [][]string
		for _, tx := range resp.Items {
			createdAt := "-"
			if !tx.CreatedAt.IsZero() {
				createdAt = tx.CreatedAt.Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{
				outfmt.FormatID("credit_transaction", tx.ID),
				tx.Type,
				fmt.Sprintf("%+.2f", tx.Amount),
				fmt.Sprintf("%.2f", tx.Balance),
				tx.Description,
				tx.ReferenceID,
				createdAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d transactions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var companyCreditsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete company credit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete company credit %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete company credit %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteCompanyCredit(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete company credit: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted company credit %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(companyCreditsCmd)

	companyCreditsCmd.AddCommand(companyCreditsListCmd)
	companyCreditsListCmd.Flags().Int("page", 1, "Page number")
	companyCreditsListCmd.Flags().Int("page-size", 20, "Results per page")
	companyCreditsListCmd.Flags().String("company-id", "", "Filter by company ID")
	companyCreditsListCmd.Flags().String("status", "", "Filter by status (active, suspended)")

	companyCreditsCmd.AddCommand(companyCreditsGetCmd)

	companyCreditsCmd.AddCommand(companyCreditsCreateCmd)
	companyCreditsCreateCmd.Flags().String("company-id", "", "Company ID (required)")
	companyCreditsCreateCmd.Flags().Float64("credit-limit", 0, "Credit limit (required)")
	companyCreditsCreateCmd.Flags().String("currency", "USD", "Currency code")
	_ = companyCreditsCreateCmd.MarkFlagRequired("company-id")
	_ = companyCreditsCreateCmd.MarkFlagRequired("credit-limit")

	companyCreditsCmd.AddCommand(companyCreditsAdjustCmd)
	companyCreditsAdjustCmd.Flags().Float64("amount", 0, "Amount to adjust (positive for credit, negative for debit) (required)")
	companyCreditsAdjustCmd.Flags().String("description", "", "Transaction description")
	companyCreditsAdjustCmd.Flags().String("reference-id", "", "External reference ID")
	_ = companyCreditsAdjustCmd.MarkFlagRequired("amount")

	companyCreditsCmd.AddCommand(companyCreditsTransactionsCmd)
	companyCreditsTransactionsCmd.Flags().Int("page", 1, "Page number")
	companyCreditsTransactionsCmd.Flags().Int("page-size", 20, "Results per page")

	companyCreditsCmd.AddCommand(companyCreditsDeleteCmd)
	companyCreditsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

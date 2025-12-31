package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var giftCardsCmd = &cobra.Command{
	Use:   "gift-cards",
	Short: "Manage gift cards",
}

var giftCardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List gift cards",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		customerID, _ := cmd.Flags().GetString("customer-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.GiftCardsListOptions{
			Page:       page,
			PageSize:   pageSize,
			Status:     status,
			CustomerID: customerID,
		}

		resp, err := client.ListGiftCards(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list gift cards: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CODE", "INITIAL VALUE", "BALANCE", "CURRENCY", "STATUS", "EXPIRES", "CREATED"}
		var rows [][]string
		for _, gc := range resp.Items {
			expiresAt := ""
			if !gc.ExpiresAt.IsZero() {
				expiresAt = gc.ExpiresAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				gc.ID,
				gc.MaskedCode,
				gc.InitialValue,
				gc.Balance,
				gc.Currency,
				string(gc.Status),
				expiresAt,
				gc.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d gift cards\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var giftCardsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get gift card details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		giftCard, err := client.GetGiftCard(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get gift card: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(giftCard)
		}

		expiresAt := "N/A"
		if !giftCard.ExpiresAt.IsZero() {
			expiresAt = giftCard.ExpiresAt.Format(time.RFC3339)
		}
		disabledAt := "N/A"
		if !giftCard.DisabledAt.IsZero() {
			disabledAt = giftCard.DisabledAt.Format(time.RFC3339)
		}

		fmt.Printf("Gift Card ID:   %s\n", giftCard.ID)
		fmt.Printf("Code:           %s\n", giftCard.Code)
		fmt.Printf("Masked Code:    %s\n", giftCard.MaskedCode)
		fmt.Printf("Initial Value:  %s %s\n", giftCard.InitialValue, giftCard.Currency)
		fmt.Printf("Balance:        %s %s\n", giftCard.Balance, giftCard.Currency)
		fmt.Printf("Status:         %s\n", giftCard.Status)
		fmt.Printf("Customer ID:    %s\n", giftCard.CustomerID)
		fmt.Printf("Note:           %s\n", giftCard.Note)
		fmt.Printf("Expires:        %s\n", expiresAt)
		fmt.Printf("Disabled At:    %s\n", disabledAt)
		fmt.Printf("Created:        %s\n", giftCard.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", giftCard.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var giftCardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a gift card",
	RunE: func(cmd *cobra.Command, args []string) error {
		initialValue, _ := cmd.Flags().GetString("initial-value")
		currency, _ := cmd.Flags().GetString("currency")
		code, _ := cmd.Flags().GetString("code")
		customerID, _ := cmd.Flags().GetString("customer-id")
		note, _ := cmd.Flags().GetString("note")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create gift card: value=%s %s\n", initialValue, currency)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.GiftCardCreateRequest{
			InitialValue: initialValue,
			Currency:     currency,
			Code:         code,
			CustomerID:   customerID,
			Note:         note,
		}

		giftCard, err := client.CreateGiftCard(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create gift card: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(giftCard)
		}

		fmt.Printf("Created gift card: %s\n", giftCard.ID)
		fmt.Printf("Code:    %s\n", giftCard.Code)
		fmt.Printf("Value:   %s %s\n", giftCard.InitialValue, giftCard.Currency)
		fmt.Printf("Status:  %s\n", giftCard.Status)

		return nil
	},
}

var giftCardsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Disable a gift card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would disable gift card: %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteGiftCard(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to disable gift card: %w", err)
		}

		fmt.Printf("Disabled gift card: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(giftCardsCmd)

	giftCardsCmd.AddCommand(giftCardsListCmd)
	giftCardsListCmd.Flags().String("status", "", "Filter by status (enabled, disabled)")
	giftCardsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	giftCardsListCmd.Flags().Int("page", 1, "Page number")
	giftCardsListCmd.Flags().Int("page-size", 20, "Results per page")

	giftCardsCmd.AddCommand(giftCardsGetCmd)

	giftCardsCmd.AddCommand(giftCardsCreateCmd)
	giftCardsCreateCmd.Flags().String("initial-value", "", "Initial value of the gift card")
	giftCardsCreateCmd.Flags().String("currency", "USD", "Currency code")
	giftCardsCreateCmd.Flags().String("code", "", "Custom gift card code (optional)")
	giftCardsCreateCmd.Flags().String("customer-id", "", "Customer ID to assign gift card to (optional)")
	giftCardsCreateCmd.Flags().String("note", "", "Note about the gift card (optional)")
	_ = giftCardsCreateCmd.MarkFlagRequired("initial-value")

	giftCardsCmd.AddCommand(giftCardsDeleteCmd)
}

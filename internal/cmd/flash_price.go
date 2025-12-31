package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var flashPriceCmd = &cobra.Command{
	Use:   "flash-price",
	Short: "Manage flash sale pricing",
}

var flashPriceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List flash prices",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		productID, _ := cmd.Flags().GetString("product-id")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.FlashPriceListOptions{
			Page:      page,
			PageSize:  pageSize,
			ProductID: productID,
			Status:    status,
		}

		resp, err := client.ListFlashPrices(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list flash prices: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PRODUCT ID", "ORIGINAL", "FLASH PRICE", "DISCOUNT", "SOLD", "STATUS", "STARTS", "ENDS"}
		var rows [][]string
		for _, fp := range resp.Items {
			discount := fmt.Sprintf("%.0f%%", fp.DiscountPct)
			sold := fmt.Sprintf("%d", fp.QuantitySold)
			if fp.Quantity > 0 {
				sold = fmt.Sprintf("%d/%d", fp.QuantitySold, fp.Quantity)
			}
			startsAt := "-"
			if !fp.StartsAt.IsZero() {
				startsAt = fp.StartsAt.Format("2006-01-02 15:04")
			}
			endsAt := "-"
			if !fp.EndsAt.IsZero() {
				endsAt = fp.EndsAt.Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{
				fp.ID,
				fp.ProductID,
				fmt.Sprintf("%.2f", fp.OriginalPrice),
				fmt.Sprintf("%.2f", fp.FlashPrice),
				discount,
				sold,
				fp.Status,
				startsAt,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d flash prices\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var flashPriceGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get flash price details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		flashPrice, err := client.GetFlashPrice(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get flash price: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(flashPrice)
		}

		fmt.Printf("Flash Price ID:  %s\n", flashPrice.ID)
		fmt.Printf("Product ID:      %s\n", flashPrice.ProductID)
		if flashPrice.VariantID != "" {
			fmt.Printf("Variant ID:      %s\n", flashPrice.VariantID)
		}
		fmt.Printf("Original Price:  %.2f\n", flashPrice.OriginalPrice)
		fmt.Printf("Flash Price:     %.2f\n", flashPrice.FlashPrice)
		fmt.Printf("Discount:        %.0f%%\n", flashPrice.DiscountPct)
		fmt.Printf("Sold:            %d", flashPrice.QuantitySold)
		if flashPrice.Quantity > 0 {
			fmt.Printf(" / %d", flashPrice.Quantity)
		}
		fmt.Println()
		if flashPrice.LimitPerUser > 0 {
			fmt.Printf("Limit Per User:  %d\n", flashPrice.LimitPerUser)
		}
		fmt.Printf("Status:          %s\n", flashPrice.Status)
		if !flashPrice.StartsAt.IsZero() {
			fmt.Printf("Starts At:       %s\n", flashPrice.StartsAt.Format(time.RFC3339))
		}
		if !flashPrice.EndsAt.IsZero() {
			fmt.Printf("Ends At:         %s\n", flashPrice.EndsAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:         %s\n", flashPrice.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var flashPriceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a flash price",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		productID, _ := cmd.Flags().GetString("product-id")
		variantID, _ := cmd.Flags().GetString("variant-id")
		flashPrice, _ := cmd.Flags().GetFloat64("flash-price")
		quantity, _ := cmd.Flags().GetInt("quantity")
		limitPerUser, _ := cmd.Flags().GetInt("limit-per-user")
		startsAtStr, _ := cmd.Flags().GetString("starts-at")
		endsAtStr, _ := cmd.Flags().GetString("ends-at")

		req := &api.FlashPriceCreateRequest{
			ProductID:    productID,
			VariantID:    variantID,
			FlashPrice:   flashPrice,
			Quantity:     quantity,
			LimitPerUser: limitPerUser,
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

		fp, err := client.CreateFlashPrice(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create flash price: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(fp)
		}

		fmt.Printf("Created flash price %s for product %s (price: %.2f)\n", fp.ID, fp.ProductID, fp.FlashPrice)
		return nil
	},
}

var flashPriceActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a flash price",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		flashPrice, err := client.ActivateFlashPrice(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate flash price: %w", err)
		}

		fmt.Printf("Activated flash price %s (status: %s)\n", flashPrice.ID, flashPrice.Status)
		return nil
	},
}

var flashPriceDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a flash price",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		flashPrice, err := client.DeactivateFlashPrice(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate flash price: %w", err)
		}

		fmt.Printf("Deactivated flash price %s (status: %s)\n", flashPrice.ID, flashPrice.Status)
		return nil
	},
}

var flashPriceUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a flash price campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.FlashPriceUpdateRequest{}

		if cmd.Flags().Changed("flash-price") {
			flashPrice, _ := cmd.Flags().GetFloat64("flash-price")
			req.FlashPrice = &flashPrice
		}

		if cmd.Flags().Changed("quantity") {
			quantity, _ := cmd.Flags().GetInt("quantity")
			req.Quantity = &quantity
		}

		if cmd.Flags().Changed("limit-per-user") {
			limitPerUser, _ := cmd.Flags().GetInt("limit-per-user")
			req.LimitPerUser = &limitPerUser
		}

		if startsAtStr, _ := cmd.Flags().GetString("starts-at"); startsAtStr != "" {
			startsAt, err := time.Parse(time.RFC3339, startsAtStr)
			if err != nil {
				return fmt.Errorf("invalid starts-at format (use RFC3339): %w", err)
			}
			req.StartsAt = &startsAt
		}

		if endsAtStr, _ := cmd.Flags().GetString("ends-at"); endsAtStr != "" {
			endsAt, err := time.Parse(time.RFC3339, endsAtStr)
			if err != nil {
				return fmt.Errorf("invalid ends-at format (use RFC3339): %w", err)
			}
			req.EndsAt = &endsAt
		}

		fp, err := client.UpdateFlashPrice(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update flash price: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(fp)
		}

		fmt.Printf("Updated flash price %s (price: %.2f)\n", fp.ID, fp.FlashPrice)
		return nil
	},
}

var flashPriceDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a flash price",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Delete flash price %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteFlashPrice(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete flash price: %w", err)
		}

		fmt.Printf("Deleted flash price %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(flashPriceCmd)

	flashPriceCmd.AddCommand(flashPriceListCmd)
	flashPriceListCmd.Flags().Int("page", 1, "Page number")
	flashPriceListCmd.Flags().Int("page-size", 20, "Results per page")
	flashPriceListCmd.Flags().String("product-id", "", "Filter by product ID")
	flashPriceListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")

	flashPriceCmd.AddCommand(flashPriceGetCmd)

	flashPriceCmd.AddCommand(flashPriceCreateCmd)
	flashPriceCreateCmd.Flags().String("product-id", "", "Product ID (required)")
	flashPriceCreateCmd.Flags().String("variant-id", "", "Variant ID")
	flashPriceCreateCmd.Flags().Float64("flash-price", 0, "Flash sale price (required)")
	flashPriceCreateCmd.Flags().Int("quantity", 0, "Available quantity")
	flashPriceCreateCmd.Flags().Int("limit-per-user", 0, "Limit per user")
	flashPriceCreateCmd.Flags().String("starts-at", "", "Start time (RFC3339 format)")
	flashPriceCreateCmd.Flags().String("ends-at", "", "End time (RFC3339 format)")
	_ = flashPriceCreateCmd.MarkFlagRequired("product-id")
	_ = flashPriceCreateCmd.MarkFlagRequired("flash-price")

	flashPriceCmd.AddCommand(flashPriceActivateCmd)
	flashPriceCmd.AddCommand(flashPriceDeactivateCmd)

	flashPriceCmd.AddCommand(flashPriceUpdateCmd)
	flashPriceUpdateCmd.Flags().Float64("flash-price", 0, "Flash sale price")
	flashPriceUpdateCmd.Flags().Int("quantity", 0, "Available quantity")
	flashPriceUpdateCmd.Flags().Int("limit-per-user", 0, "Limit per user")
	flashPriceUpdateCmd.Flags().String("starts-at", "", "Start time (RFC3339 format)")
	flashPriceUpdateCmd.Flags().String("ends-at", "", "End time (RFC3339 format)")

	flashPriceCmd.AddCommand(flashPriceDeleteCmd)
	flashPriceDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var giftsCmd = &cobra.Command{
	Use:   "gifts",
	Short: "Manage gift promotions",
}

var giftsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List gift promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.GiftsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		resp, err := client.ListGifts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list gifts: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		renderGiftsTable(formatter, resp.Items, "")
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d gifts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var giftsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get gift details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		giftID, err := resolveOrArg(cmd, args, func(query string) (string, error) {
			resp, err := client.SearchGifts(cmd.Context(), &api.GiftSearchOptions{
				Query: query, PageSize: 1,
			})
			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}
			if len(resp.Items) == 0 {
				return "", fmt.Errorf("no gift found matching %q", query)
			}
			if len(resp.Items) > 1 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %d matches found, using first\n", len(resp.Items))
			}
			return resp.Items[0].ID, nil
		})
		if err != nil {
			return err
		}

		gift, err := client.GetGift(cmd.Context(), giftID)
		if err != nil {
			return fmt.Errorf("failed to get gift: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(gift)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Gift ID:         %s\n", gift.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:           %s\n", gift.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", gift.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Gift Product:    %s (%s)\n", gift.GiftProductName, gift.GiftProductID)
		if gift.GiftVariantID != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Gift Variant:    %s\n", gift.GiftVariantID)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Trigger Type:    %s\n", gift.TriggerType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Trigger Value:   %.2f\n", gift.TriggerValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Used:            %d", gift.QuantityUsed)
		if gift.Quantity > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), " / %d", gift.Quantity)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		if gift.LimitPerUser > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Limit Per User:  %d\n", gift.LimitPerUser)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", gift.Status)
		if !gift.StartsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:       %s\n", gift.StartsAt.Format(time.RFC3339))
		}
		if !gift.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:         %s\n", gift.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", gift.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var giftsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a gift promotion",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create gift") {
			return nil
		}

		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		giftProductID, _ := cmd.Flags().GetString("gift-product-id")
		giftVariantID, _ := cmd.Flags().GetString("gift-variant-id")
		triggerType, _ := cmd.Flags().GetString("trigger-type")
		triggerValue, _ := cmd.Flags().GetFloat64("trigger-value")
		quantity, _ := cmd.Flags().GetInt("quantity")
		limitPerUser, _ := cmd.Flags().GetInt("limit-per-user")
		startsAtStr, _ := cmd.Flags().GetString("starts-at")
		endsAtStr, _ := cmd.Flags().GetString("ends-at")

		req := &api.GiftCreateRequest{
			Title:         title,
			Description:   description,
			GiftProductID: giftProductID,
			GiftVariantID: giftVariantID,
			TriggerType:   triggerType,
			TriggerValue:  triggerValue,
			Quantity:      quantity,
			LimitPerUser:  limitPerUser,
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

		gift, err := client.CreateGift(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create gift: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(gift)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created gift promotion %s (%s)\n", gift.ID, gift.Title)
		return nil
	},
}

var giftsActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a gift promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would activate gift %s", args[0])) {
			return nil
		}

		gift, err := client.ActivateGift(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate gift: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Activated gift %s (status: %s)\n", gift.ID, gift.Status)
		return nil
	},
}

var giftsDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a gift promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would deactivate gift %s", args[0])) {
			return nil
		}

		gift, err := client.DeactivateGift(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate gift: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deactivated gift %s (status: %s)\n", gift.ID, gift.Status)
		return nil
	},
}

var giftsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a gift promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete gift %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete gift %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteGift(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete gift: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted gift %s\n", args[0])
		return nil
	},
}

var giftsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a gift promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		giftProductID, _ := cmd.Flags().GetString("gift-product-id")
		giftVariantID, _ := cmd.Flags().GetString("gift-variant-id")
		triggerType, _ := cmd.Flags().GetString("trigger-type")
		triggerValue, _ := cmd.Flags().GetFloat64("trigger-value")
		quantity, _ := cmd.Flags().GetInt("quantity")
		limitPerUser, _ := cmd.Flags().GetInt("limit-per-user")
		startsAtStr, _ := cmd.Flags().GetString("starts-at")
		endsAtStr, _ := cmd.Flags().GetString("ends-at")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update gift %s", args[0])) {
			return nil
		}

		req := &api.GiftUpdateRequest{}
		if cmd.Flags().Changed("title") {
			req.Title = title
		}
		if cmd.Flags().Changed("description") {
			req.Description = description
		}
		if cmd.Flags().Changed("gift-product-id") {
			req.GiftProductID = giftProductID
		}
		if cmd.Flags().Changed("gift-variant-id") {
			req.GiftVariantID = giftVariantID
		}
		if cmd.Flags().Changed("trigger-type") {
			req.TriggerType = triggerType
		}
		if cmd.Flags().Changed("trigger-value") {
			req.TriggerValue = &triggerValue
		}
		if cmd.Flags().Changed("quantity") {
			req.Quantity = &quantity
		}
		if cmd.Flags().Changed("limit-per-user") {
			req.LimitPerUser = &limitPerUser
		}
		if cmd.Flags().Changed("starts-at") && startsAtStr != "" {
			startsAt, err := time.Parse(time.RFC3339, startsAtStr)
			if err != nil {
				return fmt.Errorf("invalid starts-at format (use RFC3339): %w", err)
			}
			req.StartsAt = &startsAt
		}
		if cmd.Flags().Changed("ends-at") && endsAtStr != "" {
			endsAt, err := time.Parse(time.RFC3339, endsAtStr)
			if err != nil {
				return fmt.Errorf("invalid ends-at format (use RFC3339): %w", err)
			}
			req.EndsAt = &endsAt
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		gift, err := client.UpdateGift(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update gift: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(gift)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated gift %s\n", gift.ID)
		return nil
	},
}

var giftsUpdateQuantityCmd = &cobra.Command{
	Use:   "update-quantity <id>",
	Short: "Update gift quantity (documented endpoint)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		qty, _ := cmd.Flags().GetInt("quantity")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update gift quantity for %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		gift, err := client.UpdateGiftQuantity(cmd.Context(), args[0], qty)
		if err != nil {
			return fmt.Errorf("failed to update gift quantity: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(gift)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated gift %s quantity to %d\n", gift.ID, gift.Quantity)
		return nil
	},
}

var giftsUpdateQuantityBySKUCmd = &cobra.Command{
	Use:   "update-quantity-by-sku",
	Short: "Bulk update gift quantity by SKU (documented endpoint)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sku, _ := cmd.Flags().GetString("sku")
		qty, _ := cmd.Flags().GetInt("quantity")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update gift quantity for SKU %s", sku)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.UpdateGiftsQuantityBySKU(cmd.Context(), sku, qty); err != nil {
			return fmt.Errorf("failed to update gifts quantity by sku: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(map[string]any{"ok": true})
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated gift quantity for SKU %s\n", sku)
		return nil
	},
}

var giftsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search gift promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("q")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.GiftSearchOptions{
			Query:    query,
			Status:   status,
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.SearchGifts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search gifts: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		renderGiftsTable(formatter, resp.Items, "gift")
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d gifts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

// renderGiftsTable renders a table of gift promotions using the given formatter.
// idPrefix is used for FormatID on search commands; pass "" for list commands (auto-prefix handles it).
func renderGiftsTable(formatter *outfmt.Formatter, gifts []api.Gift, idPrefix string) {
	headers := []string{"ID", "TITLE", "GIFT PRODUCT", "TRIGGER", "USED", "STATUS", "STARTS", "ENDS"}
	var rows [][]string
	for _, g := range gifts {
		trigger := fmt.Sprintf("%s: %.2f", g.TriggerType, g.TriggerValue)
		used := fmt.Sprintf("%d", g.QuantityUsed)
		if g.Quantity > 0 {
			used = fmt.Sprintf("%d/%d", g.QuantityUsed, g.Quantity)
		}
		startsAt := "-"
		if !g.StartsAt.IsZero() {
			startsAt = g.StartsAt.Format("2006-01-02")
		}
		endsAt := "-"
		if !g.EndsAt.IsZero() {
			endsAt = g.EndsAt.Format("2006-01-02")
		}
		id := g.ID
		if idPrefix != "" {
			id = outfmt.FormatID(idPrefix, g.ID)
		}
		rows = append(rows, []string{
			id,
			g.Title,
			g.GiftProductName,
			trigger,
			used,
			g.Status,
			startsAt,
			endsAt,
		})
	}
	formatter.Table(headers, rows)
}

var giftsStocksCmd = &cobra.Command{
	Use:   "stocks",
	Short: "Manage gift stocks (documented endpoints)",
}

var giftsStocksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get gift stocks (documented endpoint; raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetGiftStocks(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get gift stocks: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var giftsStocksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update gift stocks (documented endpoint; raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update gift stocks for %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateGiftStocks(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update gift stocks: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(giftsCmd)

	giftsCmd.AddCommand(giftsListCmd)
	giftsListCmd.Flags().Int("page", 1, "Page number")
	giftsListCmd.Flags().Int("page-size", 20, "Results per page")
	giftsListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")

	giftsCmd.AddCommand(giftsGetCmd)
	giftsGetCmd.Flags().String("by", "", "Find gift by title instead of ID")

	giftsCmd.AddCommand(giftsCreateCmd)
	giftsCreateCmd.Flags().String("title", "", "Gift title (required)")
	giftsCreateCmd.Flags().String("description", "", "Gift description")
	giftsCreateCmd.Flags().String("gift-product-id", "", "Product ID for the gift (required)")
	giftsCreateCmd.Flags().String("gift-variant-id", "", "Variant ID for the gift")
	giftsCreateCmd.Flags().String("trigger-type", "", "Trigger type: min_purchase, product_purchase (required)")
	giftsCreateCmd.Flags().Float64("trigger-value", 0, "Trigger value (required)")
	giftsCreateCmd.Flags().Int("quantity", 0, "Available quantity")
	giftsCreateCmd.Flags().Int("limit-per-user", 0, "Limit per user")
	giftsCreateCmd.Flags().String("starts-at", "", "Start time (RFC3339 format)")
	giftsCreateCmd.Flags().String("ends-at", "", "End time (RFC3339 format)")
	_ = giftsCreateCmd.MarkFlagRequired("title")
	_ = giftsCreateCmd.MarkFlagRequired("gift-product-id")
	_ = giftsCreateCmd.MarkFlagRequired("trigger-type")
	_ = giftsCreateCmd.MarkFlagRequired("trigger-value")

	giftsCmd.AddCommand(giftsSearchCmd)
	giftsSearchCmd.Flags().String("q", "", "Search query")
	giftsSearchCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")
	giftsSearchCmd.Flags().Int("page", 1, "Page number")
	giftsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	giftsCmd.AddCommand(giftsActivateCmd)
	giftsCmd.AddCommand(giftsDeactivateCmd)

	giftsCmd.AddCommand(giftsDeleteCmd)
	giftsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	giftsCmd.AddCommand(giftsUpdateCmd)
	giftsUpdateCmd.Flags().String("title", "", "Gift title")
	giftsUpdateCmd.Flags().String("description", "", "Gift description")
	giftsUpdateCmd.Flags().String("gift-product-id", "", "Product ID for the gift")
	giftsUpdateCmd.Flags().String("gift-variant-id", "", "Variant ID for the gift")
	giftsUpdateCmd.Flags().String("trigger-type", "", "Trigger type: min_purchase, product_purchase")
	giftsUpdateCmd.Flags().Float64("trigger-value", 0, "Trigger value")
	giftsUpdateCmd.Flags().Int("quantity", 0, "Available quantity")
	giftsUpdateCmd.Flags().Int("limit-per-user", 0, "Limit per user")
	giftsUpdateCmd.Flags().String("starts-at", "", "Start time (RFC3339 format)")
	giftsUpdateCmd.Flags().String("ends-at", "", "End time (RFC3339 format)")

	giftsCmd.AddCommand(giftsUpdateQuantityCmd)
	giftsUpdateQuantityCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = giftsUpdateQuantityCmd.MarkFlagRequired("quantity")

	giftsCmd.AddCommand(giftsUpdateQuantityBySKUCmd)
	giftsUpdateQuantityBySKUCmd.Flags().String("sku", "", "Gift SKU (required)")
	giftsUpdateQuantityBySKUCmd.Flags().Int("quantity", 0, "Quantity (required)")
	_ = giftsUpdateQuantityBySKUCmd.MarkFlagRequired("sku")
	_ = giftsUpdateQuantityBySKUCmd.MarkFlagRequired("quantity")

	giftsCmd.AddCommand(giftsStocksCmd)
	giftsStocksCmd.AddCommand(giftsStocksGetCmd)
	giftsStocksCmd.AddCommand(giftsStocksUpdateCmd)
	addJSONBodyFlags(giftsStocksUpdateCmd)

	schema.Register(schema.Resource{
		Name:        "gifts",
		Description: "Manage gift promotions",
		Commands:    []string{"list", "get", "search", "create", "update", "activate", "deactivate", "delete", "update-quantity", "update-quantity-by-sku", "stocks"},
		IDPrefix:    "gift",
	})
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var promotionsCmd = &cobra.Command{
	Use:   "promotions",
	Short: "Manage promotions",
}

var promotionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		promoType, _ := cmd.Flags().GetString("type")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.PromotionsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
			Type:     promoType,
		}

		resp, err := client.ListPromotions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list promotions: %w", err)
		}

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				lightItems := toLightSlice(resp.Items, toLightPromotion)
				return formatter.JSON(api.ListResponse[lightPromotion]{
					Items:      lightItems,
					Pagination: resp.Pagination,
					Page:       resp.Page,
					PageSize:   resp.PageSize,
					TotalCount: resp.TotalCount,
					HasMore:    resp.HasMore,
				})
			}
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "TYPE", "STATUS", "DISCOUNT", "USAGE", "STARTS", "ENDS"}
		var rows [][]string
		for _, p := range resp.Items {
			discount := fmt.Sprintf("%.0f", p.DiscountValue)
			if p.DiscountType == "percentage" {
				discount += "%"
			}
			usage := fmt.Sprintf("%d", p.UsageCount)
			if p.UsageLimit > 0 {
				usage = fmt.Sprintf("%d/%d", p.UsageCount, p.UsageLimit)
			}
			startsAt := "-"
			if !p.StartsAt.IsZero() {
				startsAt = p.StartsAt.Format("2006-01-02")
			}
			endsAt := "-"
			if !p.EndsAt.IsZero() {
				endsAt = p.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("promotion", p.ID),
				p.Title,
				p.Type,
				p.Status,
				discount,
				usage,
				startsAt,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d promotions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var promotionsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get promotion details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotionID, err := resolveOrArg(cmd, args, func(query string) (string, error) {
			resp, err := client.SearchPromotions(cmd.Context(), &api.PromotionSearchOptions{
				Query: query, PageSize: 1,
			})
			if err != nil {
				return "", fmt.Errorf("search failed: %w", err)
			}
			if len(resp.Items) == 0 {
				return "", fmt.Errorf("no promotion found matching %q", query)
			}
			if len(resp.Items) > 1 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %d matches found, using first\n", len(resp.Items))
			}
			return resp.Items[0].ID, nil
		})
		if err != nil {
			return err
		}

		promotion, err := client.GetPromotion(cmd.Context(), promotionID)
		if err != nil {
			return fmt.Errorf("failed to get promotion: %w", err)
		}

		light, _ := cmd.Flags().GetBool("light")
		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			if light {
				return formatter.JSON(toLightPromotion(promotion))
			}
			return formatter.JSON(promotion)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Promotion ID:    %s\n", promotion.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:           %s\n", promotion.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:     %s\n", promotion.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:            %s\n", promotion.Type)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", promotion.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:   %s\n", promotion.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value:  %.2f\n", promotion.DiscountValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Min Purchase:    %.2f\n", promotion.MinPurchase)
		_, _ = fmt.Fprintf(outWriter(cmd), "Usage:           %d", promotion.UsageCount)
		if promotion.UsageLimit > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), " / %d", promotion.UsageLimit)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Starts At:       %s\n", promotion.StartsAt.Format(time.RFC3339))
		if !promotion.EndsAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Ends At:         %s\n", promotion.EndsAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", promotion.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var promotionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a promotion",
	Long:  "Create a promotion using either --body/--body-file (raw JSON) or individual flags (--title, --discount-type, --discount-value, etc.).",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create promotion") {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("title") || cmd.Flags().Changed("discount-type") ||
			cmd.Flags().Changed("discount-value") || cmd.Flags().Changed("starts-at") ||
			cmd.Flags().Changed("ends-at") || cmd.Flags().Changed("usage-limit") ||
			cmd.Flags().Changed("status")

		if hasBody && hasFlags {
			return fmt.Errorf("use either --body/--body-file or individual flags, not both")
		}
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide promotion data via --body/--body-file or individual flags (--title, --discount-type, --discount-value, --starts-at)")
		}

		var req api.PromotionCreateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		} else {
			title, _ := cmd.Flags().GetString("title")
			discountType, _ := cmd.Flags().GetString("discount-type")
			discountValue, _ := cmd.Flags().GetFloat64("discount-value")
			startsAtStr, _ := cmd.Flags().GetString("starts-at")
			endsAtStr, _ := cmd.Flags().GetString("ends-at")
			usageLimit, _ := cmd.Flags().GetInt("usage-limit")
			status, _ := cmd.Flags().GetString("status")

			req.Title = title
			req.DiscountType = discountType
			req.DiscountValue = discountValue
			if status != "" {
				req.Type = status
			}
			if usageLimit > 0 {
				req.UsageLimit = usageLimit
			}

			if startsAtStr != "" {
				t, err := parsePromotionTime(startsAtStr, "starts-at")
				if err != nil {
					return err
				}
				req.StartsAt = t
			}
			if endsAtStr != "" {
				t, err := parsePromotionTime(endsAtStr, "ends-at")
				if err != nil {
					return err
				}
				req.EndsAt = t
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.CreatePromotion(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to create promotion: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(promotion)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsActivateCmd = &cobra.Command{
	Use:   "activate <id>",
	Short: "Activate a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would activate promotion %s", args[0])) {
			return nil
		}

		promotion, err := client.ActivatePromotion(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to activate promotion: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Activated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would deactivate promotion %s", args[0])) {
			return nil
		}

		promotion, err := client.DeactivatePromotion(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to deactivate promotion: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deactivated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a promotion",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete promotion %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete promotion %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeletePromotion(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete promotion: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted promotion %s\n", args[0])
		return nil
	},
}

var promotionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a promotion",
	Long:  "Update a promotion using either --body/--body-file (raw JSON) or individual flags (--title, --discount-type, --discount-value, etc.).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update promotion %s", args[0])) {
			return nil
		}

		hasBody := cmd.Flags().Changed("body") || cmd.Flags().Changed("body-file")
		hasFlags := cmd.Flags().Changed("title") || cmd.Flags().Changed("discount-type") ||
			cmd.Flags().Changed("discount-value") || cmd.Flags().Changed("starts-at") ||
			cmd.Flags().Changed("ends-at") || cmd.Flags().Changed("usage-limit") ||
			cmd.Flags().Changed("status")

		if hasBody && hasFlags {
			return fmt.Errorf("use either --body/--body-file or individual flags, not both")
		}
		if !hasBody && !hasFlags {
			return fmt.Errorf("provide promotion data via --body/--body-file or individual flags (--title, --discount-type, --discount-value, --starts-at)")
		}

		var req api.PromotionUpdateRequest
		if hasBody {
			if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
				return err
			}
		} else {
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				req.Title = &v
			}
			if cmd.Flags().Changed("discount-type") {
				v, _ := cmd.Flags().GetString("discount-type")
				req.DiscountType = &v
			}
			if cmd.Flags().Changed("discount-value") {
				v, _ := cmd.Flags().GetFloat64("discount-value")
				req.DiscountValue = &v
			}
			if cmd.Flags().Changed("usage-limit") {
				v, _ := cmd.Flags().GetInt("usage-limit")
				req.UsageLimit = &v
			}
			if cmd.Flags().Changed("status") {
				v, _ := cmd.Flags().GetString("status")
				req.Type = &v
			}
			if cmd.Flags().Changed("starts-at") {
				v, _ := cmd.Flags().GetString("starts-at")
				t, err := parsePromotionTime(v, "starts-at")
				if err != nil {
					return err
				}
				req.StartsAt = &t
			}
			if cmd.Flags().Changed("ends-at") {
				v, _ := cmd.Flags().GetString("ends-at")
				t, err := parsePromotionTime(v, "ends-at")
				if err != nil {
					return err
				}
				req.EndsAt = &t
			}
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		promotion, err := client.UpdatePromotion(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update promotion: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(promotion)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated promotion %s (status: %s)\n", promotion.ID, promotion.Status)
		return nil
	},
}

var promotionsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search promotions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("q")
		status, _ := cmd.Flags().GetString("status")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.PromotionSearchOptions{
			Query:    query,
			Status:   status,
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.SearchPromotions(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search promotions: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "TYPE", "STATUS", "DISCOUNT", "USAGE", "STARTS", "ENDS"}
		var rows [][]string
		for _, p := range resp.Items {
			discount := fmt.Sprintf("%.0f", p.DiscountValue)
			if p.DiscountType == "percentage" {
				discount += "%"
			}
			usage := fmt.Sprintf("%d", p.UsageCount)
			if p.UsageLimit > 0 {
				usage = fmt.Sprintf("%d/%d", p.UsageCount, p.UsageLimit)
			}
			startsAt := "-"
			if !p.StartsAt.IsZero() {
				startsAt = p.StartsAt.Format("2006-01-02")
			}
			endsAt := "-"
			if !p.EndsAt.IsZero() {
				endsAt = p.EndsAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("promotion", p.ID),
				p.Title,
				p.Type,
				p.Status,
				discount,
				usage,
				startsAt,
				endsAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d promotions\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var promotionsCouponCenterCmd = &cobra.Command{
	Use:   "coupon-center",
	Short: "Get coupon center promotions (documented endpoint; raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetPromotionsCouponCenter(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get promotions coupon center: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

// parsePromotionTime parses a time string in RFC3339 or YYYY-MM-DD format.
func parsePromotionTime(value, label string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t, err = time.Parse("2006-01-02", value)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid --%s format, use RFC3339 or YYYY-MM-DD: %w", label, err)
		}
	}
	return t, nil
}

func init() {
	rootCmd.AddCommand(promotionsCmd)

	promotionsCmd.AddCommand(promotionsListCmd)
	promotionsListCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")
	promotionsListCmd.Flags().String("type", "", "Filter by type")
	promotionsListCmd.Flags().Int("page", 1, "Page number")
	promotionsListCmd.Flags().Int("page-size", 20, "Results per page")
	promotionsListCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(promotionsListCmd.Flags(), "light", "li")

	promotionsCmd.AddCommand(promotionsGetCmd)
	promotionsGetCmd.Flags().String("by", "", "Find promotion by title instead of ID")
	promotionsGetCmd.Flags().Bool("light", false, "Minimal payload (saves tokens)")
	flagAlias(promotionsGetCmd.Flags(), "light", "li")

	promotionsCmd.AddCommand(promotionsCreateCmd)
	addJSONBodyFlags(promotionsCreateCmd)
	promotionsCreateCmd.Flags().String("title", "", "Promotion title")
	promotionsCreateCmd.Flags().String("discount-type", "", "Discount type: percentage or fixed_amount")
	promotionsCreateCmd.Flags().Float64("discount-value", 0, "Discount amount")
	promotionsCreateCmd.Flags().String("starts-at", "", "Start date (RFC3339 or YYYY-MM-DD)")
	promotionsCreateCmd.Flags().String("ends-at", "", "End date (RFC3339 or YYYY-MM-DD)")
	promotionsCreateCmd.Flags().Int("usage-limit", 0, "Max uses (0 = unlimited)")
	promotionsCreateCmd.Flags().String("status", "", "Promotion status: active, inactive")

	promotionsCmd.AddCommand(promotionsUpdateCmd)
	addJSONBodyFlags(promotionsUpdateCmd)
	promotionsUpdateCmd.Flags().String("title", "", "Promotion title")
	promotionsUpdateCmd.Flags().String("discount-type", "", "Discount type: percentage or fixed_amount")
	promotionsUpdateCmd.Flags().Float64("discount-value", 0, "Discount amount")
	promotionsUpdateCmd.Flags().String("starts-at", "", "Start date (RFC3339 or YYYY-MM-DD)")
	promotionsUpdateCmd.Flags().String("ends-at", "", "End date (RFC3339 or YYYY-MM-DD)")
	promotionsUpdateCmd.Flags().Int("usage-limit", 0, "Max uses (0 = unlimited)")
	promotionsUpdateCmd.Flags().String("status", "", "Promotion status: active, inactive")

	promotionsCmd.AddCommand(promotionsSearchCmd)
	promotionsSearchCmd.Flags().String("q", "", "Search query")
	promotionsSearchCmd.Flags().String("status", "", "Filter by status (active, scheduled, expired, inactive)")
	promotionsSearchCmd.Flags().Int("page", 1, "Page number")
	promotionsSearchCmd.Flags().Int("page-size", 20, "Results per page")

	promotionsCmd.AddCommand(promotionsActivateCmd)
	promotionsCmd.AddCommand(promotionsDeactivateCmd)
	promotionsCmd.AddCommand(promotionsDeleteCmd)
	promotionsCmd.AddCommand(promotionsCouponCenterCmd)

	schema.Register(schema.Resource{
		Name:        "promotions",
		Description: "Manage promotions",
		Commands:    []string{"list", "get", "create", "update", "search", "activate", "deactivate", "delete", "coupon-center"},
		IDPrefix:    "promotion",
	})
}

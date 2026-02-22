package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var userCouponsCmd = &cobra.Command{
	Use:   "user-coupons",
	Short: "Manage user-assigned coupons",
}

var userCouponsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List user coupons",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		promotionID, _ := cmd.Flags().GetString("promotion-id")
		userID, _ := cmd.Flags().GetString("user-id")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.UserCouponsListOptions{
			Page:        page,
			PageSize:    pageSize,
			PromotionID: promotionID,
			UserID:      userID,
			Status:      status,
		}

		resp, err := client.ListUserCoupons(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list user coupons: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "USER ID", "COUPON CODE", "TITLE", "DISCOUNT", "STATUS", "EXPIRES"}
		var rows [][]string
		for _, uc := range resp.Items {
			discount := fmt.Sprintf("%.0f", uc.DiscountValue)
			if uc.DiscountType == "percentage" {
				discount += "%"
			}
			expiresAt := "-"
			if !uc.ExpiresAt.IsZero() {
				expiresAt = uc.ExpiresAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				outfmt.FormatID("user_coupon", uc.ID),
				uc.UserID,
				uc.CouponCode,
				uc.Title,
				discount,
				uc.Status,
				expiresAt,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d user coupons\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var userCouponsListDocsCmd = &cobra.Command{
	Use:     "list-docs",
	Aliases: []string{"list-v2", "ls-docs"},
	Short:   "List user coupons (via documented /user_coupons/list endpoint)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListUserCouponsListEndpoint(cmd.Context(), &api.UserCouponsListEndpointOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list user coupons: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var userCouponsClaimCmd = &cobra.Command{
	Use:   "claim <coupon-code>",
	Short: "Claim a user coupon (by coupon code)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		var req json.RawMessage
		var hasBody bool
		var err error
		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			req, err = readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
			hasBody = true
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		var anyBody any
		if hasBody {
			anyBody = req
		}
		resp, err := client.ClaimUserCoupon(cmd.Context(), args[0], anyBody)
		if err != nil {
			return fmt.Errorf("failed to claim user coupon: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var userCouponsRedeemCmd = &cobra.Command{
	Use:   "redeem <coupon-code>",
	Short: "Redeem a user coupon (by coupon code)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")

		var req json.RawMessage
		var hasBody bool
		var err error
		if strings.TrimSpace(body) != "" || strings.TrimSpace(bodyFile) != "" {
			req, err = readJSONBodyFlags(cmd)
			if err != nil {
				return err
			}
			hasBody = true
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		var anyBody any
		if hasBody {
			anyBody = req
		}
		resp, err := client.RedeemUserCoupon(cmd.Context(), args[0], anyBody)
		if err != nil {
			return fmt.Errorf("failed to redeem user coupon: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var userCouponsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get user coupon details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		userCoupon, err := client.GetUserCoupon(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get user coupon: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(userCoupon)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "User Coupon ID: %s\n", userCoupon.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "User ID:        %s\n", userCoupon.UserID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Coupon ID:      %s\n", userCoupon.CouponID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Coupon Code:    %s\n", userCoupon.CouponCode)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:          %s\n", userCoupon.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Type:  %s\n", userCoupon.DiscountType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Discount Value: %.2f\n", userCoupon.DiscountValue)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:         %s\n", userCoupon.Status)
		if !userCoupon.UsedAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Used At:        %s\n", userCoupon.UsedAt.Format(time.RFC3339))
		}
		if !userCoupon.ExpiresAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Expires At:     %s\n", userCoupon.ExpiresAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", userCoupon.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var userCouponsAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a coupon to a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		userID, _ := cmd.Flags().GetString("user-id")
		couponID, _ := cmd.Flags().GetString("coupon-id")

		req := &api.UserCouponAssignRequest{
			UserID:   userID,
			CouponID: couponID,
		}

		userCoupon, err := client.AssignUserCoupon(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to assign coupon: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(userCoupon)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Assigned coupon %s to user %s (id: %s)\n", userCoupon.CouponCode, userCoupon.UserID, userCoupon.ID)
		return nil
	},
}

var userCouponsRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke a user's coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if !confirmAction(cmd, fmt.Sprintf("Revoke user coupon %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.RevokeUserCoupon(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to revoke user coupon: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Revoked user coupon %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCouponsCmd)

	userCouponsCmd.AddCommand(userCouponsListCmd)
	userCouponsListCmd.Flags().Int("page", 1, "Page number")
	userCouponsListCmd.Flags().Int("page-size", 20, "Results per page")
	userCouponsListCmd.Flags().String("promotion-id", "", "Promotion/coupon campaign ID (required for /user_coupons endpoint)")
	userCouponsListCmd.Flags().String("user-id", "", "Filter by user ID")
	userCouponsListCmd.Flags().String("status", "", "Filter by status (active, used, expired, revoked)")
	_ = userCouponsListCmd.MarkFlagRequired("promotion-id")

	userCouponsCmd.AddCommand(userCouponsListDocsCmd)
	userCouponsListDocsCmd.Flags().Int("page", 1, "Page number")
	userCouponsListDocsCmd.Flags().Int("page-size", 20, "Results per page")

	userCouponsCmd.AddCommand(userCouponsGetCmd)

	userCouponsCmd.AddCommand(userCouponsAssignCmd)
	userCouponsAssignCmd.Flags().String("user-id", "", "User ID (required)")
	userCouponsAssignCmd.Flags().String("coupon-id", "", "Coupon ID (required)")
	_ = userCouponsAssignCmd.MarkFlagRequired("user-id")
	_ = userCouponsAssignCmd.MarkFlagRequired("coupon-id")

	userCouponsCmd.AddCommand(userCouponsRevokeCmd)
	userCouponsRevokeCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	userCouponsCmd.AddCommand(userCouponsClaimCmd)
	addJSONBodyFlags(userCouponsClaimCmd)

	userCouponsCmd.AddCommand(userCouponsRedeemCmd)
	addJSONBodyFlags(userCouponsRedeemCmd)
}

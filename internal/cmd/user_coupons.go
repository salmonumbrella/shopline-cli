package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
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
		userID, _ := cmd.Flags().GetString("user-id")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.UserCouponsListOptions{
			Page:     page,
			PageSize: pageSize,
			UserID:   userID,
			Status:   status,
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
				uc.ID,
				uc.UserID,
				uc.CouponCode,
				uc.Title,
				discount,
				uc.Status,
				expiresAt,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d user coupons\n", len(resp.Items), resp.TotalCount)
		return nil
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

		fmt.Printf("User Coupon ID: %s\n", userCoupon.ID)
		fmt.Printf("User ID:        %s\n", userCoupon.UserID)
		fmt.Printf("Coupon ID:      %s\n", userCoupon.CouponID)
		fmt.Printf("Coupon Code:    %s\n", userCoupon.CouponCode)
		fmt.Printf("Title:          %s\n", userCoupon.Title)
		fmt.Printf("Discount Type:  %s\n", userCoupon.DiscountType)
		fmt.Printf("Discount Value: %.2f\n", userCoupon.DiscountValue)
		fmt.Printf("Status:         %s\n", userCoupon.Status)
		if !userCoupon.UsedAt.IsZero() {
			fmt.Printf("Used At:        %s\n", userCoupon.UsedAt.Format(time.RFC3339))
		}
		if !userCoupon.ExpiresAt.IsZero() {
			fmt.Printf("Expires At:     %s\n", userCoupon.ExpiresAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", userCoupon.CreatedAt.Format(time.RFC3339))
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

		fmt.Printf("Assigned coupon %s to user %s (id: %s)\n", userCoupon.CouponCode, userCoupon.UserID, userCoupon.ID)
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

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Revoke user coupon %s? [y/N] ", args[0])
			var confirm string
			_, _ = fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.RevokeUserCoupon(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to revoke user coupon: %w", err)
		}

		fmt.Printf("Revoked user coupon %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCouponsCmd)

	userCouponsCmd.AddCommand(userCouponsListCmd)
	userCouponsListCmd.Flags().Int("page", 1, "Page number")
	userCouponsListCmd.Flags().Int("page-size", 20, "Results per page")
	userCouponsListCmd.Flags().String("user-id", "", "Filter by user ID")
	userCouponsListCmd.Flags().String("status", "", "Filter by status (active, used, expired, revoked)")

	userCouponsCmd.AddCommand(userCouponsGetCmd)

	userCouponsCmd.AddCommand(userCouponsAssignCmd)
	userCouponsAssignCmd.Flags().String("user-id", "", "User ID (required)")
	userCouponsAssignCmd.Flags().String("coupon-id", "", "Coupon ID (required)")
	_ = userCouponsAssignCmd.MarkFlagRequired("user-id")
	_ = userCouponsAssignCmd.MarkFlagRequired("coupon-id")

	userCouponsCmd.AddCommand(userCouponsRevokeCmd)
	userCouponsRevokeCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

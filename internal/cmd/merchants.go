package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var merchantsCmd = &cobra.Command{
	Use:   "merchants",
	Short: "View merchant information",
}

var merchantsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current merchant details",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		merchant, err := client.GetMerchant(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get merchant: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(merchant)
		}

		fmt.Printf("Merchant ID:     %s\n", merchant.ID)
		fmt.Printf("Name:            %s\n", merchant.Name)
		fmt.Printf("Handle:          %s\n", merchant.Handle)
		fmt.Printf("Owner:           %s\n", merchant.ShopOwner)
		fmt.Printf("Email:           %s\n", merchant.Email)
		fmt.Printf("Phone:           %s\n", merchant.Phone)
		fmt.Printf("Domain:          %s\n", merchant.Domain)
		fmt.Printf("Primary Domain:  %s\n", merchant.PrimaryDomain)
		fmt.Printf("\n--- Location ---\n")
		fmt.Printf("Country:         %s (%s)\n", merchant.Country, merchant.CountryCode)
		fmt.Printf("Province:        %s\n", merchant.Province)
		fmt.Printf("City:            %s\n", merchant.City)
		fmt.Printf("Address:         %s\n", merchant.Address1)
		if merchant.Address2 != "" {
			fmt.Printf("                 %s\n", merchant.Address2)
		}
		fmt.Printf("ZIP:             %s\n", merchant.Zip)
		fmt.Printf("Timezone:        %s\n", merchant.Timezone)
		fmt.Printf("\n--- Settings ---\n")
		fmt.Printf("Plan:            %s (%s)\n", merchant.PlanDisplayName, merchant.Plan)
		fmt.Printf("Currency:        %s\n", merchant.Currency)
		fmt.Printf("Weight Unit:     %s\n", merchant.WeightUnit)
		fmt.Printf("Taxes Included:  %v\n", merchant.TaxesIncluded)
		fmt.Printf("Tax Shipping:    %v\n", merchant.TaxShipping)
		fmt.Printf("Password Enabled:%v\n", merchant.PasswordEnabled)
		fmt.Printf("Setup Required:  %v\n", merchant.SetupRequired)

		if merchant.Features != nil {
			fmt.Printf("\n--- Features ---\n")
			fmt.Printf("Checkout:        %v\n", merchant.Features.Checkout)
			fmt.Printf("Multi-Location:  %v\n", merchant.Features.MultiLocation)
			fmt.Printf("Multi-Currency:  %v\n", merchant.Features.MultiCurrency)
			fmt.Printf("Gift Cards:      %v\n", merchant.Features.GiftCards)
			fmt.Printf("Subscriptions:   %v\n", merchant.Features.Subscriptions)
			fmt.Printf("Discounts:       %v\n", merchant.Features.Discounts)
		}

		if merchant.Finances != nil && len(merchant.Finances.EnabledPresentmentCurrencies) > 0 {
			fmt.Printf("\n--- Currencies ---\n")
			fmt.Printf("Presentment Currencies: %s\n", strings.Join(merchant.Finances.EnabledPresentmentCurrencies, ", "))
		}

		fmt.Printf("\nCreated:         %s\n", merchant.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:         %s\n", merchant.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

// Staff subcommands
var merchantsStaffCmd = &cobra.Command{
	Use:   "staff",
	Short: "Manage merchant staff",
}

var merchantsStaffListCmd = &cobra.Command{
	Use:   "list",
	Short: "List merchant staff",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		role, _ := cmd.Flags().GetString("role")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.MerchantStaffListOptions{
			Page:     page,
			PageSize: pageSize,
			Role:     role,
		}

		resp, err := client.ListMerchantStaff(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list merchant staff: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "NAME", "ROLE", "OWNER", "ACTIVE", "LAST LOGIN"}
		var rows [][]string
		for _, s := range resp.Items {
			name := fmt.Sprintf("%s %s", s.FirstName, s.LastName)
			owner := "No"
			if s.AccountOwner {
				owner = "Yes"
			}
			active := "No"
			if s.Active {
				active = "Yes"
			}
			lastLogin := "Never"
			if s.LastLoginAt != nil {
				lastLogin = s.LastLoginAt.Format("2006-01-02")
			}
			rows = append(rows, []string{
				s.ID,
				s.Email,
				name,
				s.Role,
				owner,
				active,
				lastLogin,
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d staff members\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var merchantsStaffGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get staff member details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		staff, err := client.GetMerchantStaff(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get staff member: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(staff)
		}

		fmt.Printf("Staff ID:      %s\n", staff.ID)
		fmt.Printf("Email:         %s\n", staff.Email)
		fmt.Printf("First Name:    %s\n", staff.FirstName)
		fmt.Printf("Last Name:     %s\n", staff.LastName)
		fmt.Printf("Phone:         %s\n", staff.Phone)
		fmt.Printf("Role:          %s\n", staff.Role)
		fmt.Printf("Account Owner: %v\n", staff.AccountOwner)
		fmt.Printf("Active:        %v\n", staff.Active)

		if len(staff.Permissions) > 0 {
			fmt.Printf("Permissions:   %s\n", strings.Join(staff.Permissions, ", "))
		}

		if staff.LastLoginAt != nil {
			fmt.Printf("Last Login:    %s\n", staff.LastLoginAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:       %s\n", staff.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:       %s\n", staff.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(merchantsCmd)

	merchantsCmd.AddCommand(merchantsGetCmd)

	// Staff subcommands
	merchantsCmd.AddCommand(merchantsStaffCmd)

	merchantsStaffCmd.AddCommand(merchantsStaffListCmd)
	merchantsStaffListCmd.Flags().String("role", "", "Filter by role")
	merchantsStaffListCmd.Flags().Int("page", 1, "Page number")
	merchantsStaffListCmd.Flags().Int("page-size", 20, "Results per page")

	merchantsStaffCmd.AddCommand(merchantsStaffGetCmd)
}

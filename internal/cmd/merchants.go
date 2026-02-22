package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var merchantsCmd = &cobra.Command{
	Use:   "merchants",
	Short: "View merchant information",
}

var merchantsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List merchants",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.ListMerchants(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list merchants: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "HANDLE", "DOMAIN", "CREATED"}
		var rows [][]string
		for _, m := range resp {
			rows = append(rows, []string{
				outfmt.FormatID("merchant", m.ID),
				m.Name,
				m.Handle,
				m.Domain,
				m.CreatedAt.Format("2006-01-02 15:04"),
			})
		}
		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d merchants\n", len(resp))
		return nil
	},
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Merchant ID:     %s\n", merchant.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", merchant.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:          %s\n", merchant.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Owner:           %s\n", merchant.ShopOwner)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:           %s\n", merchant.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Phone:           %s\n", merchant.Phone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Domain:          %s\n", merchant.Domain)
		_, _ = fmt.Fprintf(outWriter(cmd), "Primary Domain:  %s\n", merchant.PrimaryDomain)
		_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Location ---\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "Country:         %s (%s)\n", merchant.Country, merchant.CountryCode)
		_, _ = fmt.Fprintf(outWriter(cmd), "Province:        %s\n", merchant.Province)
		_, _ = fmt.Fprintf(outWriter(cmd), "City:            %s\n", merchant.City)
		_, _ = fmt.Fprintf(outWriter(cmd), "Address:         %s\n", merchant.Address1)
		if merchant.Address2 != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "                 %s\n", merchant.Address2)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "ZIP:             %s\n", merchant.Zip)
		_, _ = fmt.Fprintf(outWriter(cmd), "Timezone:        %s\n", merchant.Timezone)
		_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Settings ---\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "Plan:            %s (%s)\n", merchant.PlanDisplayName, merchant.Plan)
		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:        %s\n", merchant.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Weight Unit:     %s\n", merchant.WeightUnit)
		_, _ = fmt.Fprintf(outWriter(cmd), "Taxes Included:  %v\n", merchant.TaxesIncluded)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax Shipping:    %v\n", merchant.TaxShipping)
		_, _ = fmt.Fprintf(outWriter(cmd), "Password Enabled:%v\n", merchant.PasswordEnabled)
		_, _ = fmt.Fprintf(outWriter(cmd), "Setup Required:  %v\n", merchant.SetupRequired)

		if merchant.Features != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Features ---\n")
			_, _ = fmt.Fprintf(outWriter(cmd), "Checkout:        %v\n", merchant.Features.Checkout)
			_, _ = fmt.Fprintf(outWriter(cmd), "Multi-Location:  %v\n", merchant.Features.MultiLocation)
			_, _ = fmt.Fprintf(outWriter(cmd), "Multi-Currency:  %v\n", merchant.Features.MultiCurrency)
			_, _ = fmt.Fprintf(outWriter(cmd), "Gift Cards:      %v\n", merchant.Features.GiftCards)
			_, _ = fmt.Fprintf(outWriter(cmd), "Subscriptions:   %v\n", merchant.Features.Subscriptions)
			_, _ = fmt.Fprintf(outWriter(cmd), "Discounts:       %v\n", merchant.Features.Discounts)
		}

		if merchant.Finances != nil && len(merchant.Finances.EnabledPresentmentCurrencies) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Currencies ---\n")
			_, _ = fmt.Fprintf(outWriter(cmd), "Presentment Currencies: %s\n", strings.Join(merchant.Finances.EnabledPresentmentCurrencies, ", "))
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "\nCreated:         %s\n", merchant.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", merchant.UpdatedAt.Format(time.RFC3339))

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
				outfmt.FormatID("merchant_setting", s.ID),
				s.Email,
				name,
				s.Role,
				owner,
				active,
				lastLogin,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d staff members\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Staff ID:      %s\n", staff.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:         %s\n", staff.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "First Name:    %s\n", staff.FirstName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Last Name:     %s\n", staff.LastName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Phone:         %s\n", staff.Phone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Role:          %s\n", staff.Role)
		_, _ = fmt.Fprintf(outWriter(cmd), "Account Owner: %v\n", staff.AccountOwner)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:        %v\n", staff.Active)

		if len(staff.Permissions) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Permissions:   %s\n", strings.Join(staff.Permissions, ", "))
		}

		if staff.LastLoginAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Last Login:    %s\n", staff.LastLoginAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", staff.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", staff.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

// Documented merchant endpoints

var merchantsGetByIDCmd = &cobra.Command{
	Use:   "get-by-id <merchant-id>",
	Short: "Get merchant details by id (documented endpoint; raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetMerchantByID(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get merchant by id: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var merchantsExpressLinkCmd = &cobra.Command{
	Use:   "express-link",
	Short: "Generate express cart link (documented endpoint)",
}

var merchantsExpressLinkGenerateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"create", "new"},
	Short:   "Generate merchant's express cart link (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would generate express cart link") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GenerateMerchantExpressLink(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to generate express link: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(merchantsCmd)

	merchantsCmd.AddCommand(merchantsListCmd)
	merchantsCmd.AddCommand(merchantsGetCmd)

	// Staff subcommands
	merchantsCmd.AddCommand(merchantsStaffCmd)

	merchantsStaffCmd.AddCommand(merchantsStaffListCmd)
	merchantsStaffListCmd.Flags().String("role", "", "Filter by role")
	merchantsStaffListCmd.Flags().Int("page", 1, "Page number")
	merchantsStaffListCmd.Flags().Int("page-size", 20, "Results per page")

	merchantsStaffCmd.AddCommand(merchantsStaffGetCmd)

	// Documented endpoints
	merchantsCmd.AddCommand(merchantsGetByIDCmd)

	merchantsCmd.AddCommand(merchantsExpressLinkCmd)
	merchantsExpressLinkCmd.AddCommand(merchantsExpressLinkGenerateCmd)
	addJSONBodyFlags(merchantsExpressLinkGenerateCmd)

	schema.Register(schema.Resource{
		Name:        "merchants",
		Description: "View merchant information",
		Commands:    []string{"list", "get", "get-by-id", "staff", "metafields", "app-metafields", "express-link"},
		IDPrefix:    "merchant",
	})
}

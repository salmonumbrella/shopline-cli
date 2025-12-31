package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var checkoutSettingsCmd = &cobra.Command{
	Use:     "checkout-settings",
	Aliases: []string{"checkout"},
	Short:   "Manage checkout settings",
}

var checkoutSettingsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get checkout settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		settings, err := client.GetCheckoutSettings(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get checkout settings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(settings)
		}

		fmt.Printf("Checkout Settings\n")
		fmt.Printf("=================\n\n")
		fmt.Printf("Customer Requirements\n")
		fmt.Printf("  Require Phone:            %t\n", settings.RequirePhone)
		fmt.Printf("  Require Shipping Address: %t\n", settings.RequireShippingAddress)
		fmt.Printf("  Require Billing Address:  %t\n", settings.RequireBillingAddress)
		fmt.Printf("  Require Company:          %t\n", settings.RequireCompany)
		fmt.Printf("  Require Full Name:        %t\n", settings.RequireFullName)
		fmt.Println()
		fmt.Printf("Checkout Options\n")
		fmt.Printf("  Guest Checkout:           %t\n", settings.EnableGuestCheckout)
		fmt.Printf("  Express Checkout:         %t\n", settings.EnableExpressCheckout)
		fmt.Printf("  Order Notes:              %t\n", settings.EnableOrderNotes)
		fmt.Printf("  Address Autofill:         %t\n", settings.EnableAddressAutofill)
		fmt.Printf("  Multi-Currency:           %t\n", settings.EnableMultiCurrency)
		fmt.Println()
		fmt.Printf("Tipping\n")
		fmt.Printf("  Enabled:                  %t\n", settings.EnableTipping)
		if settings.EnableTipping && len(settings.TippingOptions) > 0 {
			fmt.Printf("  Options:                  %v%%\n", settings.TippingOptions)
			fmt.Printf("  Default:                  %.0f%%\n", settings.DefaultTippingOption)
		}
		fmt.Println()
		fmt.Printf("Abandoned Cart\n")
		fmt.Printf("  Enabled:                  %t\n", settings.AbandonedCartEnabled)
		if settings.AbandonedCartEnabled {
			fmt.Printf("  Delay (hours):            %d\n", settings.AbandonedCartDelay)
		}
		fmt.Println()
		fmt.Printf("Policies\n")
		if settings.TermsOfServiceURL != "" {
			fmt.Printf("  Terms of Service:         %s\n", settings.TermsOfServiceURL)
		}
		if settings.PrivacyPolicyURL != "" {
			fmt.Printf("  Privacy Policy:           %s\n", settings.PrivacyPolicyURL)
		}
		if settings.RefundPolicyURL != "" {
			fmt.Printf("  Refund Policy:            %s\n", settings.RefundPolicyURL)
		}
		fmt.Println()
		fmt.Printf("Updated:                    %s\n", settings.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var checkoutSettingsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update checkout settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.CheckoutSettingsUpdateRequest{}

		// Handle boolean flags
		if cmd.Flags().Changed("require-phone") {
			v, _ := cmd.Flags().GetBool("require-phone")
			req.RequirePhone = &v
		}
		if cmd.Flags().Changed("guest-checkout") {
			v, _ := cmd.Flags().GetBool("guest-checkout")
			req.EnableGuestCheckout = &v
		}
		if cmd.Flags().Changed("express-checkout") {
			v, _ := cmd.Flags().GetBool("express-checkout")
			req.EnableExpressCheckout = &v
		}
		if cmd.Flags().Changed("order-notes") {
			v, _ := cmd.Flags().GetBool("order-notes")
			req.EnableOrderNotes = &v
		}
		if cmd.Flags().Changed("tipping") {
			v, _ := cmd.Flags().GetBool("tipping")
			req.EnableTipping = &v
		}
		if cmd.Flags().Changed("abandoned-cart") {
			v, _ := cmd.Flags().GetBool("abandoned-cart")
			req.AbandonedCartEnabled = &v
		}

		// Handle string flags
		if cmd.Flags().Changed("terms-url") {
			req.TermsOfServiceURL, _ = cmd.Flags().GetString("terms-url")
		}
		if cmd.Flags().Changed("privacy-url") {
			req.PrivacyPolicyURL, _ = cmd.Flags().GetString("privacy-url")
		}
		if cmd.Flags().Changed("refund-url") {
			req.RefundPolicyURL, _ = cmd.Flags().GetString("refund-url")
		}

		// Handle int flags
		if cmd.Flags().Changed("abandoned-cart-delay") {
			req.AbandonedCartDelay, _ = cmd.Flags().GetInt("abandoned-cart-delay")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("[DRY-RUN] Would update checkout settings")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		settings, err := client.UpdateCheckoutSettings(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to update checkout settings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(settings)
		}

		fmt.Println("Checkout settings updated successfully")
		fmt.Printf("Guest Checkout:    %t\n", settings.EnableGuestCheckout)
		fmt.Printf("Express Checkout:  %t\n", settings.EnableExpressCheckout)
		fmt.Printf("Tipping:           %t\n", settings.EnableTipping)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkoutSettingsCmd)

	checkoutSettingsCmd.AddCommand(checkoutSettingsGetCmd)

	checkoutSettingsCmd.AddCommand(checkoutSettingsUpdateCmd)
	checkoutSettingsUpdateCmd.Flags().Bool("require-phone", false, "Require phone number")
	checkoutSettingsUpdateCmd.Flags().Bool("guest-checkout", false, "Enable guest checkout")
	checkoutSettingsUpdateCmd.Flags().Bool("express-checkout", false, "Enable express checkout")
	checkoutSettingsUpdateCmd.Flags().Bool("order-notes", false, "Enable order notes")
	checkoutSettingsUpdateCmd.Flags().Bool("tipping", false, "Enable tipping")
	checkoutSettingsUpdateCmd.Flags().Bool("abandoned-cart", false, "Enable abandoned cart recovery")
	checkoutSettingsUpdateCmd.Flags().Int("abandoned-cart-delay", 24, "Abandoned cart email delay (hours)")
	checkoutSettingsUpdateCmd.Flags().String("terms-url", "", "Terms of service URL")
	checkoutSettingsUpdateCmd.Flags().String("privacy-url", "", "Privacy policy URL")
	checkoutSettingsUpdateCmd.Flags().String("refund-url", "", "Refund policy URL")
}

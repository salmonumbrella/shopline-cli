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

		_, _ = fmt.Fprintf(outWriter(cmd), "Checkout Settings\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "=================\n\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer Requirements\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "  Require Phone:            %t\n", settings.RequirePhone)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Require Shipping Address: %t\n", settings.RequireShippingAddress)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Require Billing Address:  %t\n", settings.RequireBillingAddress)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Require Company:          %t\n", settings.RequireCompany)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Require Full Name:        %t\n", settings.RequireFullName)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Checkout Options\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "  Guest Checkout:           %t\n", settings.EnableGuestCheckout)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Express Checkout:         %t\n", settings.EnableExpressCheckout)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Order Notes:              %t\n", settings.EnableOrderNotes)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Address Autofill:         %t\n", settings.EnableAddressAutofill)
		_, _ = fmt.Fprintf(outWriter(cmd), "  Multi-Currency:           %t\n", settings.EnableMultiCurrency)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Tipping\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "  Enabled:                  %t\n", settings.EnableTipping)
		if settings.EnableTipping && len(settings.TippingOptions) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "  Options:                  %v%%\n", settings.TippingOptions)
			_, _ = fmt.Fprintf(outWriter(cmd), "  Default:                  %.0f%%\n", settings.DefaultTippingOption)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Abandoned Cart\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "  Enabled:                  %t\n", settings.AbandonedCartEnabled)
		if settings.AbandonedCartEnabled {
			_, _ = fmt.Fprintf(outWriter(cmd), "  Delay (hours):            %d\n", settings.AbandonedCartDelay)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Policies\n")
		if settings.TermsOfServiceURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "  Terms of Service:         %s\n", settings.TermsOfServiceURL)
		}
		if settings.PrivacyPolicyURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "  Privacy Policy:           %s\n", settings.PrivacyPolicyURL)
		}
		if settings.RefundPolicyURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "  Refund Policy:            %s\n", settings.RefundPolicyURL)
		}
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:                    %s\n", settings.UpdatedAt.Format(time.RFC3339))

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

		if checkDryRun(cmd, "[DRY-RUN] Would update checkout settings") {
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

		_, _ = fmt.Fprintln(outWriter(cmd), "Checkout settings updated successfully")
		_, _ = fmt.Fprintf(outWriter(cmd), "Guest Checkout:    %t\n", settings.EnableGuestCheckout)
		_, _ = fmt.Fprintf(outWriter(cmd), "Express Checkout:  %t\n", settings.EnableExpressCheckout)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tipping:           %t\n", settings.EnableTipping)

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

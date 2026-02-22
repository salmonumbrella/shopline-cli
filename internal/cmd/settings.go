package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage store settings",
}

var settingsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get store settings",
	Long:  "Get store settings. Store info comes from /merchants, user settings from /settings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		// Get merchant info for store settings
		merchant, err := client.GetMerchant(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get merchant settings: %w", err)
		}

		// Get user-specific settings
		userSettings, err := client.GetSettings(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get user settings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			combined := map[string]interface{}{
				"merchant":      merchant,
				"user_settings": userSettings,
			}
			return formatter.JSON(combined)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Store Settings\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "==============\n\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", merchant.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:           %s\n", merchant.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Domain:          %s\n", merchant.Domain)
		_, _ = fmt.Fprintf(outWriter(cmd), "Phone:           %s\n", merchant.Phone)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Address:         %s\n", merchant.Address1)
		if merchant.Address2 != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "                 %s\n", merchant.Address2)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "City:            %s\n", merchant.City)
		_, _ = fmt.Fprintf(outWriter(cmd), "Province:        %s\n", merchant.Province)
		_, _ = fmt.Fprintf(outWriter(cmd), "Country:         %s\n", merchant.CountryCode)
		_, _ = fmt.Fprintf(outWriter(cmd), "ZIP:             %s\n", merchant.Zip)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:        %s\n", merchant.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Timezone:        %s\n", merchant.Timezone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Weight Unit:     %s\n", merchant.WeightUnit)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Taxes Included:  %t\n", merchant.TaxesIncluded)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax Shipping:    %t\n", merchant.TaxShipping)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "User Settings\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "-------------\n")
		_, _ = fmt.Fprintf(outWriter(cmd), "Min Age Limit:   %s\n", userSettings.Users.MinimumAgeLimit)
		_, _ = fmt.Fprintf(outWriter(cmd), "POS Apply Credit: %t\n", userSettings.Users.PosApplyCredit)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", merchant.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", merchant.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var settingsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update user settings",
	Long:  "Update user-specific settings (minimum age limit, POS apply credit). Store settings are managed via the merchant endpoint.",
	RunE: func(cmd *cobra.Command, args []string) error {
		minAgeLimit, _ := cmd.Flags().GetString("min-age-limit")

		req := &api.UserSettingsUpdateRequest{
			Users: api.UserSettingsUpdate{
				MinimumAgeLimit: minAgeLimit,
			},
		}

		if cmd.Flags().Changed("pos-apply-credit") {
			posApplyCredit, _ := cmd.Flags().GetBool("pos-apply-credit")
			req.Users.PosApplyCredit = &posApplyCredit
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update user settings") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		settings, err := client.UpdateSettings(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to update settings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(settings)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "User settings updated successfully")
		_, _ = fmt.Fprintf(outWriter(cmd), "Min Age Limit:    %s\n", settings.Users.MinimumAgeLimit)
		_, _ = fmt.Fprintf(outWriter(cmd), "POS Apply Credit: %t\n", settings.Users.PosApplyCredit)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(settingsCmd)

	settingsCmd.AddCommand(settingsGetCmd)

	settingsCmd.AddCommand(settingsUpdateCmd)
	settingsUpdateCmd.Flags().String("min-age-limit", "", "Minimum age limit for customers")
	settingsUpdateCmd.Flags().Bool("pos-apply-credit", false, "Apply store credit in POS transactions")
}

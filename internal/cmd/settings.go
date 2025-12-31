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

		fmt.Printf("Store Settings\n")
		fmt.Printf("==============\n\n")
		fmt.Printf("Name:            %s\n", merchant.Name)
		fmt.Printf("Email:           %s\n", merchant.Email)
		fmt.Printf("Domain:          %s\n", merchant.Domain)
		fmt.Printf("Phone:           %s\n", merchant.Phone)
		fmt.Println()
		fmt.Printf("Address:         %s\n", merchant.Address1)
		if merchant.Address2 != "" {
			fmt.Printf("                 %s\n", merchant.Address2)
		}
		fmt.Printf("City:            %s\n", merchant.City)
		fmt.Printf("Province:        %s\n", merchant.Province)
		fmt.Printf("Country:         %s\n", merchant.CountryCode)
		fmt.Printf("ZIP:             %s\n", merchant.Zip)
		fmt.Println()
		fmt.Printf("Currency:        %s\n", merchant.Currency)
		fmt.Printf("Timezone:        %s\n", merchant.Timezone)
		fmt.Printf("Weight Unit:     %s\n", merchant.WeightUnit)
		fmt.Println()
		fmt.Printf("Taxes Included:  %t\n", merchant.TaxesIncluded)
		fmt.Printf("Tax Shipping:    %t\n", merchant.TaxShipping)
		fmt.Println()
		fmt.Printf("User Settings\n")
		fmt.Printf("-------------\n")
		fmt.Printf("Min Age Limit:   %s\n", userSettings.Users.MinimumAgeLimit)
		fmt.Printf("POS Apply Credit: %t\n", userSettings.Users.PosApplyCredit)
		fmt.Println()
		fmt.Printf("Created:         %s\n", merchant.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:         %s\n", merchant.UpdatedAt.Format(time.RFC3339))

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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Println("[DRY-RUN] Would update user settings")
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

		fmt.Println("User settings updated successfully")
		fmt.Printf("Min Age Limit:    %s\n", settings.Users.MinimumAgeLimit)
		fmt.Printf("POS Apply Credit: %t\n", settings.Users.PosApplyCredit)

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

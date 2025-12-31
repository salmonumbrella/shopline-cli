package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var shopCmd = &cobra.Command{
	Use:   "shop",
	Short: "Manage shop settings",
}

var shopInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get shop information",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		shop, err := client.GetShop(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get shop info: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(shop)
		}

		fmt.Printf("Shop ID:         %s\n", shop.ID)
		fmt.Printf("Name:            %s\n", shop.Name)
		fmt.Printf("Email:           %s\n", shop.Email)
		fmt.Printf("Domain:          %s\n", shop.Domain)
		fmt.Printf("Shopline Domain: %s\n", shop.MyshoplineDomain)
		fmt.Printf("Phone:           %s\n", shop.Phone)
		fmt.Printf("Owner:           %s\n", shop.ShopOwner)
		fmt.Println()
		fmt.Printf("Address:         %s\n", shop.Address1)
		if shop.Address2 != "" {
			fmt.Printf("                 %s\n", shop.Address2)
		}
		fmt.Printf("City:            %s\n", shop.City)
		fmt.Printf("Province:        %s (%s)\n", shop.Province, shop.ProvinceCode)
		fmt.Printf("Country:         %s (%s)\n", shop.Country, shop.CountryCode)
		fmt.Printf("ZIP:             %s\n", shop.Zip)
		fmt.Println()
		fmt.Printf("Currency:        %s\n", shop.Currency)
		fmt.Printf("Timezone:        %s\n", shop.Timezone)
		fmt.Printf("Weight Unit:     %s\n", shop.WeightUnit)
		fmt.Println()
		fmt.Printf("Plan:            %s (%s)\n", shop.PlanDisplayName, shop.PlanName)
		fmt.Printf("Created:         %s\n", shop.CreatedAt.Format(time.RFC3339))
		return nil
	},
}

var shopSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Get shop settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		settings, err := client.GetShopSettings(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get shop settings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(settings)
		}

		fmt.Printf("Currency:             %s\n", settings.Currency)
		fmt.Printf("Weight Unit:          %s\n", settings.WeightUnit)
		fmt.Printf("Timezone:             %s\n", settings.Timezone)
		fmt.Printf("Order Prefix:         %s\n", settings.OrderPrefix)
		fmt.Printf("Order Suffix:         %s\n", settings.OrderSuffix)
		fmt.Printf("Taxes Included:       %t\n", settings.TaxesIncluded)
		fmt.Printf("Tax Shipping:         %t\n", settings.TaxShipping)
		fmt.Printf("Auto Fulfillment:     %t\n", settings.AutomaticFulfillment)
		if len(settings.EnabledPresentmentCurrencies) > 0 {
			fmt.Printf("Currencies Enabled:   %v\n", settings.EnabledPresentmentCurrencies)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shopCmd)

	shopCmd.AddCommand(shopInfoCmd)
	shopCmd.AddCommand(shopSettingsCmd)
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Shop ID:         %s\n", shop.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:            %s\n", shop.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:           %s\n", shop.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Domain:          %s\n", shop.Domain)
		_, _ = fmt.Fprintf(outWriter(cmd), "Shopline Domain: %s\n", shop.MyshoplineDomain)
		_, _ = fmt.Fprintf(outWriter(cmd), "Phone:           %s\n", shop.Phone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Owner:           %s\n", shop.ShopOwner)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Address:         %s\n", shop.Address1)
		if shop.Address2 != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "                 %s\n", shop.Address2)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "City:            %s\n", shop.City)
		_, _ = fmt.Fprintf(outWriter(cmd), "Province:        %s (%s)\n", shop.Province, shop.ProvinceCode)
		_, _ = fmt.Fprintf(outWriter(cmd), "Country:         %s (%s)\n", shop.Country, shop.CountryCode)
		_, _ = fmt.Fprintf(outWriter(cmd), "ZIP:             %s\n", shop.Zip)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:        %s\n", shop.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Timezone:        %s\n", shop.Timezone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Weight Unit:     %s\n", shop.WeightUnit)
		_, _ = fmt.Fprintln(outWriter(cmd))
		_, _ = fmt.Fprintf(outWriter(cmd), "Plan:            %s (%s)\n", shop.PlanDisplayName, shop.PlanName)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", shop.CreatedAt.Format(time.RFC3339))
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Currency:             %s\n", settings.Currency)
		_, _ = fmt.Fprintf(outWriter(cmd), "Weight Unit:          %s\n", settings.WeightUnit)
		_, _ = fmt.Fprintf(outWriter(cmd), "Timezone:             %s\n", settings.Timezone)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order Prefix:         %s\n", settings.OrderPrefix)
		_, _ = fmt.Fprintf(outWriter(cmd), "Order Suffix:         %s\n", settings.OrderSuffix)
		_, _ = fmt.Fprintf(outWriter(cmd), "Taxes Included:       %t\n", settings.TaxesIncluded)
		_, _ = fmt.Fprintf(outWriter(cmd), "Tax Shipping:         %t\n", settings.TaxShipping)
		_, _ = fmt.Fprintf(outWriter(cmd), "Auto Fulfillment:     %t\n", settings.AutomaticFulfillment)
		if len(settings.EnabledPresentmentCurrencies) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Currencies Enabled:   %v\n", settings.EnabledPresentmentCurrencies)
		}
		return nil
	},
}

var shopSettingsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update shop settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would update shop settings") {
			return nil
		}

		var req api.ShopSettingsUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		settings, err := client.UpdateShopSettings(cmd.Context(), &req)
		if err != nil {
			return fmt.Errorf("failed to update shop settings: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(settings)
		}

		_, _ = fmt.Fprintln(outWriter(cmd), "Updated shop settings")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shopCmd)

	shopCmd.AddCommand(shopInfoCmd)
	shopCmd.AddCommand(shopSettingsCmd)
	shopSettingsCmd.AddCommand(shopSettingsUpdateCmd)
	addJSONBodyFlags(shopSettingsUpdateCmd)
	shopSettingsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	schema.Register(schema.Resource{
		Name:        "shop",
		Description: "Manage shop settings",
		Commands:    []string{"info", "settings"},
	})
}

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ============================
// settings/* (documented endpoints)
// ============================

var settingsCheckoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Manage checkout settings (documented /settings/checkout endpoint)",
}

var settingsCheckoutGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get checkout settings (via /settings/checkout)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetSettingsCheckout(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get settings checkout: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsDomainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage domain settings",
}

var settingsDomainsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get domain settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetSettingsDomains(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get settings domains: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsDomainsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update domain settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update settings domains") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateSettingsDomains(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to update settings domains: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsLayoutsCmd = &cobra.Command{
	Use:   "layouts",
	Short: "Manage layout settings",
}

var settingsLayoutsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get layout settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetSettingsLayouts(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get settings layouts: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsLayoutsDraftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Manage layout draft settings",
}

var settingsLayoutsDraftGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get layout draft settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetSettingsLayoutsDraft(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get settings layouts draft: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsLayoutsDraftUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update layout draft settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update settings layouts draft") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateSettingsLayoutsDraft(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to update settings layouts draft: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsLayoutsPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish layout draft settings",
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

		if checkDryRun(cmd, "[DRY-RUN] Would publish settings layouts") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		var anyBody any
		if hasBody {
			anyBody = req
		}

		resp, err := client.PublishSettingsLayouts(cmd.Context(), anyBody)
		if err != nil {
			return fmt.Errorf("failed to publish settings layouts: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage theme settings",
}

var settingsThemeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get theme settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetSettingsTheme(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get settings theme: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsThemeDraftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Manage theme draft settings",
}

var settingsThemeDraftGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get theme draft settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetSettingsThemeDraft(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get settings theme draft: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsThemeDraftUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update theme draft settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would update settings theme draft") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.UpdateSettingsThemeDraft(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to update settings theme draft: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var settingsThemePublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish theme draft settings",
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

		if checkDryRun(cmd, "[DRY-RUN] Would publish settings theme") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		var anyBody any
		if hasBody {
			anyBody = req
		}

		resp, err := client.PublishSettingsTheme(cmd.Context(), anyBody)
		if err != nil {
			return fmt.Errorf("failed to publish settings theme: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

// Simple GET wrappers

func newSettingsGetJSONCmd(use, short string, fn func(cmd *cobra.Command) (json.RawMessage, error)) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := fn(cmd)
			if err != nil {
				return err
			}
			return getFormatter(cmd).JSON(resp)
		},
	}
}

var settingsOrdersGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get order settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsOrders(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings orders: %w", err)
		}
		return resp, nil
	},
)

var settingsPaymentsGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get payment settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsPayments(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings payments: %w", err)
		}
		return resp, nil
	},
)

var settingsPOSGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get POS settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsPOS(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings pos: %w", err)
		}
		return resp, nil
	},
)

var settingsProductReviewGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get product review settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsProductReview(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings product review: %w", err)
		}
		return resp, nil
	},
)

var settingsProductsGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get product settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsProducts(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings products: %w", err)
		}
		return resp, nil
	},
)

var settingsPromotionsGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get promotion settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsPromotions(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings promotions: %w", err)
		}
		return resp, nil
	},
)

var settingsShopGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get shop settings (via /settings/shop)",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsShop(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings shop: %w", err)
		}
		return resp, nil
	},
)

var settingsTaxGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get tax settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsTax(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings tax: %w", err)
		}
		return resp, nil
	},
)

var settingsThirdPartyAdsGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get third-party ads settings",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsThirdPartyAds(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings third party ads: %w", err)
		}
		return resp, nil
	},
)

var settingsUsersGetCmd = newSettingsGetJSONCmd(
	"get",
	"Get user settings (via /settings/users)",
	func(cmd *cobra.Command) (json.RawMessage, error) {
		client, err := getClient(cmd)
		if err != nil {
			return nil, err
		}
		resp, err := client.GetSettingsUsers(cmd.Context())
		if err != nil {
			return nil, fmt.Errorf("failed to get settings users: %w", err)
		}
		return resp, nil
	},
)

func init() {
	// settings checkout
	settingsCmd.AddCommand(settingsCheckoutCmd)
	settingsCheckoutCmd.AddCommand(settingsCheckoutGetCmd)

	// settings domains
	settingsCmd.AddCommand(settingsDomainsCmd)
	settingsDomainsCmd.AddCommand(settingsDomainsGetCmd)
	settingsDomainsCmd.AddCommand(settingsDomainsUpdateCmd)
	addJSONBodyFlags(settingsDomainsUpdateCmd)
	settingsDomainsUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// settings layouts
	settingsCmd.AddCommand(settingsLayoutsCmd)
	settingsLayoutsCmd.AddCommand(settingsLayoutsGetCmd)
	settingsLayoutsCmd.AddCommand(settingsLayoutsDraftCmd)
	settingsLayoutsDraftCmd.AddCommand(settingsLayoutsDraftGetCmd)
	settingsLayoutsDraftCmd.AddCommand(settingsLayoutsDraftUpdateCmd)
	addJSONBodyFlags(settingsLayoutsDraftUpdateCmd)
	settingsLayoutsDraftUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")
	settingsLayoutsCmd.AddCommand(settingsLayoutsPublishCmd)
	addJSONBodyFlags(settingsLayoutsPublishCmd)
	settingsLayoutsPublishCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// settings theme
	settingsCmd.AddCommand(settingsThemeCmd)
	settingsThemeCmd.AddCommand(settingsThemeGetCmd)
	settingsThemeCmd.AddCommand(settingsThemeDraftCmd)
	settingsThemeDraftCmd.AddCommand(settingsThemeDraftGetCmd)
	settingsThemeDraftCmd.AddCommand(settingsThemeDraftUpdateCmd)
	addJSONBodyFlags(settingsThemeDraftUpdateCmd)
	settingsThemeDraftUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")
	settingsThemeCmd.AddCommand(settingsThemePublishCmd)
	addJSONBodyFlags(settingsThemePublishCmd)
	settingsThemePublishCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	// settings orders/payments/pos/product-review/products/promotions/shop/tax/third-party-ads/users
	settingsOrdersCmd := &cobra.Command{Use: "orders", Short: "Manage order settings"}
	settingsCmd.AddCommand(settingsOrdersCmd)
	settingsOrdersCmd.AddCommand(settingsOrdersGetCmd)

	settingsPaymentsCmd := &cobra.Command{Use: "payments", Short: "Manage payment settings"}
	settingsCmd.AddCommand(settingsPaymentsCmd)
	settingsPaymentsCmd.AddCommand(settingsPaymentsGetCmd)

	settingsPOSCmd := &cobra.Command{Use: "pos", Short: "Manage POS settings"}
	settingsCmd.AddCommand(settingsPOSCmd)
	settingsPOSCmd.AddCommand(settingsPOSGetCmd)

	settingsProductReviewCmd := &cobra.Command{Use: "product-review", Short: "Manage product review settings"}
	settingsCmd.AddCommand(settingsProductReviewCmd)
	settingsProductReviewCmd.AddCommand(settingsProductReviewGetCmd)

	settingsProductsCmd := &cobra.Command{Use: "products", Short: "Manage product settings"}
	settingsCmd.AddCommand(settingsProductsCmd)
	settingsProductsCmd.AddCommand(settingsProductsGetCmd)

	settingsPromotionsCmd := &cobra.Command{Use: "promotions", Short: "Manage promotion settings"}
	settingsCmd.AddCommand(settingsPromotionsCmd)
	settingsPromotionsCmd.AddCommand(settingsPromotionsGetCmd)

	settingsShopCmd := &cobra.Command{Use: "shop", Short: "Manage shop settings (via /settings/shop)"}
	settingsCmd.AddCommand(settingsShopCmd)
	settingsShopCmd.AddCommand(settingsShopGetCmd)

	settingsTaxCmd := &cobra.Command{Use: "tax", Short: "Manage tax settings (via /settings/tax)"}
	settingsCmd.AddCommand(settingsTaxCmd)
	settingsTaxCmd.AddCommand(settingsTaxGetCmd)

	settingsThirdPartyAdsCmd := &cobra.Command{Use: "third-party-ads", Short: "Manage third-party ads settings"}
	settingsCmd.AddCommand(settingsThirdPartyAdsCmd)
	settingsThirdPartyAdsCmd.AddCommand(settingsThirdPartyAdsGetCmd)

	settingsUsersCmd := &cobra.Command{Use: "users", Short: "Manage user settings (via /settings/users)"}
	settingsCmd.AddCommand(settingsUsersCmd)
	settingsUsersCmd.AddCommand(settingsUsersGetCmd)
}

package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var shoplyticsCmd = &cobra.Command{
	Use:   "shoplytics",
	Short: "Access Shoplytics analytics (via Admin API)",
}

var shoplyticsNewReturningCmd = &cobra.Command{
	Use:   "new-and-returning",
	Short: "Get new vs returning customers by date range",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")

		result, err := client.GetShoplyticsCustomersNewAndReturning(cmd.Context(), &api.AdminShoplyticsNewReturningOptions{
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return fmt.Errorf("failed to get new/returning customers: %w", err)
		}
		return getFormatter(cmd).JSON(result)
	},
}

var shoplyticsFirstOrderChannelsCmd = &cobra.Command{
	Use:   "first-order-channels",
	Short: "Get first-order channel analytics",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.GetShoplyticsCustomersFirstOrderChannels(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get first-order channels: %w", err)
		}
		return getFormatter(cmd).JSON(result)
	},
}

var shoplyticsPaymentsMethodsGridCmd = &cobra.Command{
	Use:   "payments-methods-grid",
	Short: "Get payment methods grid analytics",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.GetShoplyticsPaymentsMethodsGrid(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get payment methods grid: %w", err)
		}
		return getFormatter(cmd).JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(shoplyticsCmd)

	shoplyticsCmd.AddCommand(shoplyticsNewReturningCmd)
	shoplyticsNewReturningCmd.Flags().String("start-date", "", "Start date in YYYY-MM-DD (required)")
	shoplyticsNewReturningCmd.Flags().String("end-date", "", "End date in YYYY-MM-DD (required)")
	_ = shoplyticsNewReturningCmd.MarkFlagRequired("start-date")
	_ = shoplyticsNewReturningCmd.MarkFlagRequired("end-date")

	shoplyticsCmd.AddCommand(shoplyticsFirstOrderChannelsCmd)
	shoplyticsCmd.AddCommand(shoplyticsPaymentsMethodsGridCmd)

	schema.Register(schema.Resource{
		Name:        "shoplytics",
		Description: "Access Shoplytics analytics (via Admin API)",
		Commands:    []string{"new-and-returning", "first-order-channels", "payments-methods-grid"},
	})
}

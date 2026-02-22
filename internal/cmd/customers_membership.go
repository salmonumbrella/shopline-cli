package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var customersMembershipInfoCmd = &cobra.Command{
	Use:   "membership-info",
	Short: "Get customers membership info",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomersMembershipInfo(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get customers membership info: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var customersMembershipTierCmd = &cobra.Command{
	Use:   "membership-tier",
	Short: "Customer membership tier tools",
}

var customersMembershipTierActionLogsCmd = &cobra.Command{
	Use:   "action-logs <customer-id>",
	Short: "Get customer membership tier action logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetCustomerMembershipTierActionLogs(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get membership tier action logs: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	customersCmd.AddCommand(customersMembershipInfoCmd)

	customersCmd.AddCommand(customersMembershipTierCmd)
	customersMembershipTierCmd.AddCommand(customersMembershipTierActionLogsCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ============================
// staffs permissions
// ============================

var staffsPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "View staff permissions",
}

var staffsPermissionsGetCmd = &cobra.Command{
	Use:   "get <staff-id>",
	Short: "Get staff permission details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		resp, err := client.GetStaffPermissions(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get staff permissions: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	staffsCmd.AddCommand(staffsPermissionsCmd)
	staffsPermissionsCmd.AddCommand(staffsPermissionsGetCmd)
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var staffsCmd = &cobra.Command{
	Use:   "staffs",
	Short: "Manage staff accounts",
}

var staffsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List staff accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.StaffsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListStaffs(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list staffs: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "EMAIL", "NAME", "OWNER", "PERMISSIONS", "CREATED"}
		var rows [][]string
		for _, s := range resp.Items {
			owner := "No"
			if s.AccountOwner {
				owner = "Yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("staff", s.ID),
				s.Email,
				fmt.Sprintf("%s %s", s.FirstName, s.LastName),
				owner,
				strings.Join(s.Permissions, ", "),
				s.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d staff accounts\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var staffsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get staff account details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		staff, err := client.GetStaff(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get staff: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(staff)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Staff ID:      %s\n", staff.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Email:         %s\n", staff.Email)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:          %s %s\n", staff.FirstName, staff.LastName)
		if staff.Phone != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Phone:         %s\n", staff.Phone)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Account Owner: %t\n", staff.AccountOwner)
		_, _ = fmt.Fprintf(outWriter(cmd), "Locale:        %s\n", staff.Locale)
		_, _ = fmt.Fprintf(outWriter(cmd), "Permissions:   %s\n", strings.Join(staff.Permissions, ", "))
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", staff.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", staff.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var staffsInviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a new staff member (not supported by Shopline API)",
	Long: `Invite a new staff member.

NOTE: The Shopline Open API does not support inviting staff members.
Staff must be added manually through the Shopline Admin panel:

  Settings > Staff and Permissions > Add Staff

This command is provided for forward compatibility in case Shopline
adds this functionality to their API in the future.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")

		// The Shopline Open API does not support staff invites.
		// Return a helpful error message instead of hitting the API.
		return fmt.Errorf(`staff invite is not supported by the Shopline API

To add %s as a staff member, use the Shopline Admin panel:
  1. Go to Settings > Staff and Permissions
  2. Click "Add Staff"
  3. Enter the email address and set permissions

See: https://help.shopline.com/hc/en-001/articles/900004300606`, email)
	},
}

var staffsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a staff account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		phone, _ := cmd.Flags().GetString("phone")
		locale, _ := cmd.Flags().GetString("locale")
		permissionsStr, _ := cmd.Flags().GetString("permissions")

		var permissions []string
		if permissionsStr != "" {
			permissions = strings.Split(permissionsStr, ",")
			for i := range permissions {
				permissions[i] = strings.TrimSpace(permissions[i])
			}
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update staff %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.StaffUpdateRequest{
			FirstName:   firstName,
			LastName:    lastName,
			Phone:       phone,
			Locale:      locale,
			Permissions: permissions,
		}

		staff, err := client.UpdateStaff(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update staff: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(staff)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated staff %s\n", staff.ID)
		return nil
	},
}

var staffsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a staff account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete staff %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete staff account %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteStaff(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete staff: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted staff %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(staffsCmd)

	staffsCmd.AddCommand(staffsListCmd)
	staffsListCmd.Flags().Int("page", 1, "Page number")
	staffsListCmd.Flags().Int("page-size", 20, "Results per page")

	staffsCmd.AddCommand(staffsGetCmd)

	staffsCmd.AddCommand(staffsInviteCmd)
	staffsInviteCmd.Flags().String("email", "", "Staff email address (for error message only)")
	// Note: Other flags removed since Shopline API doesn't support staff invites

	staffsCmd.AddCommand(staffsUpdateCmd)
	staffsUpdateCmd.Flags().String("first-name", "", "First name")
	staffsUpdateCmd.Flags().String("last-name", "", "Last name")
	staffsUpdateCmd.Flags().String("phone", "", "Phone number")
	staffsUpdateCmd.Flags().String("locale", "", "Locale (e.g., en, zh)")
	staffsUpdateCmd.Flags().String("permissions", "", "Comma-separated list of permissions")

	staffsCmd.AddCommand(staffsDeleteCmd)
}

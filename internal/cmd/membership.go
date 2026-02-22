package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var membershipCmd = &cobra.Command{
	Use:   "membership",
	Short: "Manage membership tiers",
}

var membershipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List membership tiers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.MembershipTiersListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListMembershipTiers(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list membership tiers: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "LEVEL", "MIN POINTS", "MAX POINTS", "DISCOUNT", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			maxPoints := "-"
			if t.MaxPoints > 0 {
				maxPoints = fmt.Sprintf("%d", t.MaxPoints)
			}
			discount := "-"
			if t.Discount > 0 {
				discount = fmt.Sprintf("%.0f%%", t.Discount*100)
			}
			rows = append(rows, []string{
				outfmt.FormatID("membership_tier", t.ID),
				t.Name,
				fmt.Sprintf("%d", t.Level),
				fmt.Sprintf("%d", t.MinPoints),
				maxPoints,
				discount,
				t.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d membership tiers\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var membershipGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get membership tier details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		tier, err := client.GetMembershipTier(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get membership tier: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tier)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Tier ID:      %s\n", tier.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", tier.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Level:        %d\n", tier.Level)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:  %s\n", tier.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Min Points:   %d\n", tier.MinPoints)
		if tier.MaxPoints > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Max Points:   %d\n", tier.MaxPoints)
		}
		if tier.Discount > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Discount:     %.0f%%\n", tier.Discount*100)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", tier.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", tier.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var membershipCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a membership tier",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create membership tier") {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		level, _ := cmd.Flags().GetInt("level")
		description, _ := cmd.Flags().GetString("description")
		minPoints, _ := cmd.Flags().GetInt("min-points")
		maxPoints, _ := cmd.Flags().GetInt("max-points")
		discount, _ := cmd.Flags().GetFloat64("discount")

		req := &api.MembershipTierCreateRequest{
			Name:        name,
			Level:       level,
			Description: description,
			MinPoints:   minPoints,
			MaxPoints:   maxPoints,
			Discount:    discount,
		}

		tier, err := client.CreateMembershipTier(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create membership tier: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(tier)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created membership tier %s\n", tier.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:  %s\n", tier.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Level: %d\n", tier.Level)
		return nil
	},
}

var membershipDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a membership tier",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete membership tier %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete membership tier %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteMembershipTier(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete membership tier: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted membership tier %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(membershipCmd)

	membershipCmd.AddCommand(membershipListCmd)
	membershipListCmd.Flags().Int("page", 1, "Page number")
	membershipListCmd.Flags().Int("page-size", 20, "Results per page")

	membershipCmd.AddCommand(membershipGetCmd)

	membershipCmd.AddCommand(membershipCreateCmd)
	membershipCreateCmd.Flags().String("name", "", "Tier name")
	membershipCreateCmd.Flags().Int("level", 0, "Tier level (higher = better)")
	membershipCreateCmd.Flags().String("description", "", "Tier description")
	membershipCreateCmd.Flags().Int("min-points", 0, "Minimum points required")
	membershipCreateCmd.Flags().Int("max-points", 0, "Maximum points for tier")
	membershipCreateCmd.Flags().Float64("discount", 0, "Discount percentage (0.0-1.0)")
	_ = membershipCreateCmd.MarkFlagRequired("name")
	_ = membershipCreateCmd.MarkFlagRequired("level")
	_ = membershipCreateCmd.MarkFlagRequired("min-points")

	membershipCmd.AddCommand(membershipDeleteCmd)
}

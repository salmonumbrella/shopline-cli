package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var labelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "Manage product labels",
}

var labelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List labels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.LabelsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListLabels(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list labels: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "COLOR", "ACTIVE", "CREATED"}
		var rows [][]string
		for _, l := range resp.Items {
			active := "No"
			if l.Active {
				active = "Yes"
			}
			rows = append(rows, []string{
				l.ID,
				l.Name,
				l.Color,
				active,
				l.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d labels\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var labelsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get label details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		label, err := client.GetLabel(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get label: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(label)
		}

		fmt.Printf("Label ID:     %s\n", label.ID)
		fmt.Printf("Name:         %s\n", label.Name)
		fmt.Printf("Description:  %s\n", label.Description)
		fmt.Printf("Color:        %s\n", label.Color)
		fmt.Printf("Icon:         %s\n", label.Icon)
		fmt.Printf("Active:       %t\n", label.Active)
		fmt.Printf("Created:      %s\n", label.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", label.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var labelsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a label",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		color, _ := cmd.Flags().GetString("color")
		icon, _ := cmd.Flags().GetString("icon")
		active, _ := cmd.Flags().GetBool("active")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create label '%s'\n", name)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.LabelCreateRequest{
			Name:        name,
			Description: description,
			Color:       color,
			Icon:        icon,
			Active:      active,
		}

		label, err := client.CreateLabel(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(label)
		}

		fmt.Printf("Created label %s\n", label.ID)
		fmt.Printf("Name:   %s\n", label.Name)
		fmt.Printf("Color:  %s\n", label.Color)
		fmt.Printf("Active: %t\n", label.Active)

		return nil
	},
}

var labelsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a label",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete label %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete label %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteLabel(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete label: %w", err)
		}

		fmt.Printf("Deleted label %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(labelsCmd)

	labelsCmd.AddCommand(labelsListCmd)
	labelsListCmd.Flags().Int("page", 1, "Page number")
	labelsListCmd.Flags().Int("page-size", 20, "Results per page")

	labelsCmd.AddCommand(labelsGetCmd)

	labelsCmd.AddCommand(labelsCreateCmd)
	labelsCreateCmd.Flags().String("name", "", "Label name (required)")
	labelsCreateCmd.Flags().String("description", "", "Label description")
	labelsCreateCmd.Flags().String("color", "", "Label color (hex code)")
	labelsCreateCmd.Flags().String("icon", "", "Label icon")
	labelsCreateCmd.Flags().Bool("active", true, "Label active status")
	_ = labelsCreateCmd.MarkFlagRequired("name")

	labelsCmd.AddCommand(labelsDeleteCmd)
}

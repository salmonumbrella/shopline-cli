package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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
				outfmt.FormatID("label", l.ID),
				l.Name,
				l.Color,
				active,
				l.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d labels\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Label ID:     %s\n", label.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:         %s\n", label.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Description:  %s\n", label.Description)
		_, _ = fmt.Fprintf(outWriter(cmd), "Color:        %s\n", label.Color)
		_, _ = fmt.Fprintf(outWriter(cmd), "Icon:         %s\n", label.Icon)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active:       %t\n", label.Active)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", label.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", label.UpdatedAt.Format(time.RFC3339))

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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create label '%s'", name)) {
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created label %s\n", label.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", label.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Color:  %s\n", label.Color)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", label.Active)

		return nil
	},
}

var labelsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a label",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update label %s", args[0])) {
			return nil
		}

		var req api.LabelUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		label, err := client.UpdateLabel(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update label: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(label)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated label %s\n", label.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:   %s\n", label.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Color:  %s\n", label.Color)
		_, _ = fmt.Fprintf(outWriter(cmd), "Active: %t\n", label.Active)
		return nil
	},
}

var labelsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a label",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete label %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete label %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteLabel(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete label: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted label %s\n", args[0])
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

	labelsCmd.AddCommand(labelsUpdateCmd)
	addJSONBodyFlags(labelsUpdateCmd)
	labelsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	labelsCmd.AddCommand(labelsDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "labels",
		Description: "Manage product labels",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "label",
	})
}

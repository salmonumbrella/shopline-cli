package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "Manage themes",
}

var themesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		role, _ := cmd.Flags().GetString("role")

		opts := &api.ThemesListOptions{
			Page:     page,
			PageSize: pageSize,
			Role:     role,
		}

		resp, err := client.ListThemes(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list themes: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "NAME", "ROLE", "PREVIEWABLE", "PROCESSING", "CREATED"}
		var rows [][]string
		for _, t := range resp.Items {
			previewable := "no"
			if t.Previewable {
				previewable = "yes"
			}
			processing := "no"
			if t.Processing {
				processing = "yes"
			}
			rows = append(rows, []string{
				outfmt.FormatID("theme", t.ID),
				t.Name,
				t.Role,
				previewable,
				processing,
				t.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d themes\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var themesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get theme details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		theme, err := client.GetTheme(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get theme: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(theme)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Theme ID:    %s\n", theme.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name:        %s\n", theme.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Role:        %s\n", theme.Role)
		_, _ = fmt.Fprintf(outWriter(cmd), "Previewable: %t\n", theme.Previewable)
		_, _ = fmt.Fprintf(outWriter(cmd), "Processing:  %t\n", theme.Processing)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:     %s\n", theme.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:     %s\n", theme.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var themesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a theme",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		role, _ := cmd.Flags().GetString("role")
		src, _ := cmd.Flags().GetString("src")

		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would create theme") {
			return nil
		}

		req := &api.ThemeCreateRequest{
			Name: name,
			Role: role,
			Src:  src,
		}

		theme, err := client.CreateTheme(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create theme: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(theme)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created theme %s\n", theme.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", theme.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Role: %s\n", theme.Role)

		return nil
	},
}

var themesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update theme %s", args[0])) {
			return nil
		}

		var req api.ThemeUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		theme, err := client.UpdateTheme(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update theme: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(theme)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated theme %s\n", theme.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Name: %s\n", theme.Name)
		_, _ = fmt.Fprintf(outWriter(cmd), "Role: %s\n", theme.Role)
		return nil
	},
}

var themesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete theme %s", args[0])) {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Delete theme %s? [y/N] ", args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.DeleteTheme(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete theme: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted theme %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(themesCmd)

	themesCmd.AddCommand(themesListCmd)
	themesListCmd.Flags().Int("page", 1, "Page number")
	themesListCmd.Flags().Int("page-size", 20, "Results per page")
	themesListCmd.Flags().String("role", "", "Filter by role (main, mobile, unpublished)")

	themesCmd.AddCommand(themesGetCmd)

	themesCmd.AddCommand(themesCreateCmd)
	themesCreateCmd.Flags().String("name", "", "Theme name")
	themesCreateCmd.Flags().String("role", "", "Theme role (main, mobile, unpublished)")
	themesCreateCmd.Flags().String("src", "", "URL to theme zip file")
	_ = themesCreateCmd.MarkFlagRequired("name")

	themesCmd.AddCommand(themesUpdateCmd)
	addJSONBodyFlags(themesUpdateCmd)
	themesUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	themesCmd.AddCommand(themesDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "themes",
		Description: "Manage themes",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "theme",
	})
}

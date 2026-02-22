package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var redirectsCmd = &cobra.Command{
	Use:   "redirects",
	Short: "Manage URL redirects",
}

var redirectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List redirects",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		path, _ := cmd.Flags().GetString("path")
		target, _ := cmd.Flags().GetString("target")

		opts := &api.RedirectsListOptions{
			Page:     page,
			PageSize: pageSize,
			Path:     path,
			Target:   target,
		}

		resp, err := client.ListRedirects(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list redirects: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PATH", "TARGET", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("redirect", r.ID),
				r.Path,
				r.Target,
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d redirects\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var redirectsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get redirect details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		redirect, err := client.GetRedirect(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get redirect: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(redirect)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Redirect ID: %s\n", redirect.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Path:        %s\n", redirect.Path)
		_, _ = fmt.Fprintf(outWriter(cmd), "Target:      %s\n", redirect.Target)
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:     %s\n", redirect.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:     %s\n", redirect.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var redirectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a redirect",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		target, _ := cmd.Flags().GetString("target")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create redirect: %s -> %s", path, target)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.RedirectCreateRequest{
			Path:   path,
			Target: target,
		}

		redirect, err := client.CreateRedirect(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create redirect: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(redirect)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created redirect %s\n", redirect.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Path:   %s\n", redirect.Path)
		_, _ = fmt.Fprintf(outWriter(cmd), "Target: %s\n", redirect.Target)

		return nil
	},
}

var redirectsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a redirect",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update redirect %s", args[0])) {
			return nil
		}

		var req api.RedirectUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		redirect, err := client.UpdateRedirect(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update redirect: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(redirect)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated redirect %s\n", redirect.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Path:   %s\n", redirect.Path)
		_, _ = fmt.Fprintf(outWriter(cmd), "Target: %s\n", redirect.Target)
		return nil
	},
}

var redirectsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a redirect",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete redirect %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete redirect %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteRedirect(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete redirect: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted redirect %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(redirectsCmd)

	redirectsCmd.AddCommand(redirectsListCmd)
	redirectsListCmd.Flags().Int("page", 1, "Page number")
	redirectsListCmd.Flags().Int("page-size", 20, "Results per page")
	redirectsListCmd.Flags().String("path", "", "Filter by source path")
	redirectsListCmd.Flags().String("target", "", "Filter by target URL")

	redirectsCmd.AddCommand(redirectsGetCmd)

	redirectsCmd.AddCommand(redirectsCreateCmd)
	redirectsCreateCmd.Flags().String("path", "", "Source path (required)")
	redirectsCreateCmd.Flags().String("target", "", "Target URL (required)")
	_ = redirectsCreateCmd.MarkFlagRequired("path")
	_ = redirectsCreateCmd.MarkFlagRequired("target")

	redirectsCmd.AddCommand(redirectsUpdateCmd)
	addJSONBodyFlags(redirectsUpdateCmd)
	redirectsUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	redirectsCmd.AddCommand(redirectsDeleteCmd)
	redirectsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	schema.Register(schema.Resource{
		Name:        "redirects",
		Description: "Manage URL redirects",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "redirect",
	})
}

package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var pagesCmd = &cobra.Command{
	Use:   "pages",
	Short: "Manage pages",
}

var pagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		title, _ := cmd.Flags().GetString("title")

		opts := &api.PagesListOptions{
			Page:     page,
			PageSize: pageSize,
			Title:    title,
		}

		// Handle --published flag (tri-state: true, false, or unset)
		if cmd.Flags().Changed("published") {
			published, _ := cmd.Flags().GetBool("published")
			opts.Published = &published
		}

		resp, err := client.ListPages(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list pages: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "HANDLE", "PUBLISHED", "AUTHOR", "CREATED"}
		var rows [][]string
		for _, p := range resp.Items {
			published := "no"
			if p.Published {
				published = "yes"
			}
			rows = append(rows, []string{
				p.ID,
				p.Title,
				p.Handle,
				published,
				p.Author,
				p.CreatedAt.Format("2006-01-02"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d pages\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var pagesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get page details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		pg, err := client.GetPage(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get page: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(pg)
		}

		fmt.Printf("Page ID:        %s\n", pg.ID)
		fmt.Printf("Title:          %s\n", pg.Title)
		fmt.Printf("Handle:         %s\n", pg.Handle)
		fmt.Printf("Author:         %s\n", pg.Author)
		fmt.Printf("Published:      %t\n", pg.Published)
		if !pg.PublishedAt.IsZero() {
			fmt.Printf("Published At:   %s\n", pg.PublishedAt.Format(time.RFC3339))
		}
		if pg.TemplateSuffix != "" {
			fmt.Printf("Template:       %s\n", pg.TemplateSuffix)
		}
		fmt.Printf("Created:        %s\n", pg.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", pg.UpdatedAt.Format(time.RFC3339))

		showBody, _ := cmd.Flags().GetBool("body")
		if showBody && pg.BodyHTML != "" {
			fmt.Printf("\nBody HTML:\n%s\n", pg.BodyHTML)
		}
		return nil
	},
}

var pagesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a page",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		handle, _ := cmd.Flags().GetString("handle")
		author, _ := cmd.Flags().GetString("author")
		published, _ := cmd.Flags().GetBool("published")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create page: %s\n", title)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.PageCreateRequest{
			Title:     title,
			BodyHTML:  body,
			Handle:    handle,
			Author:    author,
			Published: published,
		}

		pg, err := client.CreatePage(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create page: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(pg)
		}

		fmt.Printf("Created page %s\n", pg.ID)
		fmt.Printf("Title:   %s\n", pg.Title)
		fmt.Printf("Handle:  %s\n", pg.Handle)

		return nil
	},
}

var pagesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete page %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete page %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeletePage(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete page: %w", err)
		}

		fmt.Printf("Deleted page %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pagesCmd)

	pagesCmd.AddCommand(pagesListCmd)
	pagesListCmd.Flags().Int("page", 1, "Page number")
	pagesListCmd.Flags().Int("page-size", 20, "Results per page")
	pagesListCmd.Flags().String("title", "", "Filter by title")
	pagesListCmd.Flags().Bool("published", false, "Filter by published status")

	pagesCmd.AddCommand(pagesGetCmd)
	pagesGetCmd.Flags().Bool("body", false, "Show body HTML content")

	pagesCmd.AddCommand(pagesCreateCmd)
	pagesCreateCmd.Flags().String("title", "", "Page title (required)")
	pagesCreateCmd.Flags().String("body", "", "Page body HTML (required)")
	pagesCreateCmd.Flags().String("handle", "", "URL handle (auto-generated if not provided)")
	pagesCreateCmd.Flags().String("author", "", "Page author")
	pagesCreateCmd.Flags().Bool("published", false, "Publish the page immediately")
	_ = pagesCreateCmd.MarkFlagRequired("title")
	_ = pagesCreateCmd.MarkFlagRequired("body")

	pagesCmd.AddCommand(pagesDeleteCmd)
	pagesDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

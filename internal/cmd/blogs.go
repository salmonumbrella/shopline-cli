package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var blogsCmd = &cobra.Command{
	Use:   "blogs",
	Short: "Manage blogs",
}

var blogsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List blogs",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.BlogsListOptions{
			Page:     page,
			PageSize: pageSize,
		}

		resp, err := client.ListBlogs(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list blogs: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "TITLE", "HANDLE", "COMMENTABLE", "CREATED"}
		var rows [][]string
		for _, b := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("blog", b.ID),
				b.Title,
				b.Handle,
				b.Commentable,
				b.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d blogs\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var blogsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get blog details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		blog, err := client.GetBlog(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get blog: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(blog)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Blog ID:        %s\n", blog.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:          %s\n", blog.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:         %s\n", blog.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Commentable:    %s\n", blog.Commentable)
		if blog.Tags != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Tags:           %s\n", blog.Tags)
		}
		if blog.TemplateSuffix != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Template:       %s\n", blog.TemplateSuffix)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:        %s\n", blog.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:        %s\n", blog.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var blogsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a blog",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		handle, _ := cmd.Flags().GetString("handle")
		commentable, _ := cmd.Flags().GetString("commentable")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create blog: %s", title)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.BlogCreateRequest{
			Title:       title,
			Handle:      handle,
			Commentable: commentable,
		}

		blog, err := client.CreateBlog(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create blog: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(blog)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created blog %s\n", blog.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:   %s\n", blog.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:  %s\n", blog.Handle)

		return nil
	},
}

var blogsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a blog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update blog %s", args[0])) {
			return nil
		}

		var req api.BlogUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		blog, err := client.UpdateBlog(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update blog: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(blog)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated blog %s\n", blog.ID)
		return nil
	},
}

var blogsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a blog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete blog %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete blog %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteBlog(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete blog: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted blog %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blogsCmd)

	blogsCmd.AddCommand(blogsListCmd)
	blogsListCmd.Flags().Int("page", 1, "Page number")
	blogsListCmd.Flags().Int("page-size", 20, "Results per page")

	blogsCmd.AddCommand(blogsGetCmd)

	blogsCmd.AddCommand(blogsCreateCmd)
	blogsCreateCmd.Flags().String("title", "", "Blog title (required)")
	blogsCreateCmd.Flags().String("handle", "", "URL handle (auto-generated if not provided)")
	blogsCreateCmd.Flags().String("commentable", "moderate", "Comment setting (no, moderate, yes)")
	_ = blogsCreateCmd.MarkFlagRequired("title")

	blogsCmd.AddCommand(blogsUpdateCmd)
	addJSONBodyFlags(blogsUpdateCmd)

	blogsCmd.AddCommand(blogsDeleteCmd)
	blogsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

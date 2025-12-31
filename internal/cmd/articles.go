package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var articlesCmd = &cobra.Command{
	Use:   "articles",
	Short: "Manage blog articles",
}

var articlesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List articles",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		blogID, _ := cmd.Flags().GetString("blog-id")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ArticlesListOptions{
			Page:     page,
			PageSize: pageSize,
			BlogID:   blogID,
		}

		resp, err := client.ListArticles(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list articles: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "BLOG ID", "TITLE", "AUTHOR", "PUBLISHED", "CREATED"}
		var rows [][]string
		for _, a := range resp.Items {
			published := "No"
			if a.Published {
				published = "Yes"
			}
			rows = append(rows, []string{
				a.ID,
				a.BlogID,
				a.Title,
				a.Author,
				published,
				a.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d articles\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var articlesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get article details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		article, err := client.GetArticle(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get article: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(article)
		}

		fmt.Printf("Article ID:     %s\n", article.ID)
		fmt.Printf("Blog ID:        %s\n", article.BlogID)
		fmt.Printf("Title:          %s\n", article.Title)
		fmt.Printf("Handle:         %s\n", article.Handle)
		fmt.Printf("Author:         %s\n", article.Author)
		fmt.Printf("Tags:           %s\n", article.Tags)
		fmt.Printf("Published:      %t\n", article.Published)
		if !article.PublishedAt.IsZero() {
			fmt.Printf("Published At:   %s\n", article.PublishedAt.Format(time.RFC3339))
		}
		fmt.Printf("Created:        %s\n", article.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", article.UpdatedAt.Format(time.RFC3339))
		if article.Image != nil {
			fmt.Printf("Image:          %s\n", article.Image.Src)
		}
		fmt.Printf("\nBody HTML:\n%s\n", article.BodyHTML)

		return nil
	},
}

var articlesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an article",
	RunE: func(cmd *cobra.Command, args []string) error {
		blogID, _ := cmd.Flags().GetString("blog-id")
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		author, _ := cmd.Flags().GetString("author")
		tags, _ := cmd.Flags().GetString("tags")
		published, _ := cmd.Flags().GetBool("published")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create article '%s' in blog %s\n", title, blogID)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.ArticleCreateRequest{
			BlogID:    blogID,
			Title:     title,
			BodyHTML:  body,
			Author:    author,
			Tags:      tags,
			Published: published,
		}

		article, err := client.CreateArticle(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create article: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(article)
		}

		fmt.Printf("Created article %s\n", article.ID)
		fmt.Printf("Title:     %s\n", article.Title)
		fmt.Printf("Blog ID:   %s\n", article.BlogID)
		fmt.Printf("Published: %t\n", article.Published)

		return nil
	},
}

var articlesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete article %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete article %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteArticle(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete article: %w", err)
		}

		fmt.Printf("Deleted article %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(articlesCmd)

	articlesCmd.AddCommand(articlesListCmd)
	articlesListCmd.Flags().String("blog-id", "", "Filter by blog ID")
	articlesListCmd.Flags().Int("page", 1, "Page number")
	articlesListCmd.Flags().Int("page-size", 20, "Results per page")

	articlesCmd.AddCommand(articlesGetCmd)

	articlesCmd.AddCommand(articlesCreateCmd)
	articlesCreateCmd.Flags().String("blog-id", "", "Blog ID (required)")
	articlesCreateCmd.Flags().String("title", "", "Article title (required)")
	articlesCreateCmd.Flags().String("body", "", "Article body HTML (required)")
	articlesCreateCmd.Flags().String("author", "", "Article author")
	articlesCreateCmd.Flags().String("tags", "", "Comma-separated tags")
	articlesCreateCmd.Flags().Bool("published", false, "Publish the article")
	_ = articlesCreateCmd.MarkFlagRequired("blog-id")
	_ = articlesCreateCmd.MarkFlagRequired("title")
	_ = articlesCreateCmd.MarkFlagRequired("body")

	articlesCmd.AddCommand(articlesDeleteCmd)
}

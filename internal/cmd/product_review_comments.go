package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var productReviewCommentsCmd = &cobra.Command{
	Use:   "product-review-comments",
	Short: "Manage product review comments (documented endpoints)",
}

var productReviewCommentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List product review comments (raw JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListProductReviewComments(cmd.Context(), &api.ProductReviewCommentsListOptions{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return fmt.Errorf("failed to list product review comments: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get product review comment details (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetProductReviewComment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product review comment: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create product review comment (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create product review comment") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreateProductReviewComment(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create product review comment: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update product review comment (raw JSON body)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update product review comment %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.UpdateProductReviewComment(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update product review comment: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete product review comment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete product review comment %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete product review comment %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.DeleteProductReviewComment(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete product review comment: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create",
	Short: "Bulk create product review comments (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would bulk create product review comments") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkCreateProductReviewComments(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to bulk create product review comments: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Bulk update product review comments (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would bulk update product review comments") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkUpdateProductReviewComments(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to bulk update product review comments: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

var productReviewCommentsBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Bulk delete product review comments (raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to bulk delete product review comments? Use --yes to confirm.\n")
			return nil
		}

		if checkDryRun(cmd, "[DRY-RUN] Would bulk delete product review comments") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.BulkDeleteProductReviewComments(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to bulk delete product review comments: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(productReviewCommentsCmd)

	productReviewCommentsCmd.AddCommand(productReviewCommentsListCmd)
	productReviewCommentsListCmd.Flags().Int("page", 1, "Page number")
	productReviewCommentsListCmd.Flags().Int("page-size", 20, "Results per page")

	productReviewCommentsCmd.AddCommand(productReviewCommentsGetCmd)

	productReviewCommentsCmd.AddCommand(productReviewCommentsCreateCmd)
	addJSONBodyFlags(productReviewCommentsCreateCmd)

	productReviewCommentsCmd.AddCommand(productReviewCommentsUpdateCmd)
	addJSONBodyFlags(productReviewCommentsUpdateCmd)

	productReviewCommentsCmd.AddCommand(productReviewCommentsDeleteCmd)
	productReviewCommentsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	productReviewCommentsCmd.AddCommand(productReviewCommentsBulkCreateCmd)
	addJSONBodyFlags(productReviewCommentsBulkCreateCmd)

	productReviewCommentsCmd.AddCommand(productReviewCommentsBulkUpdateCmd)
	addJSONBodyFlags(productReviewCommentsBulkUpdateCmd)

	productReviewCommentsCmd.AddCommand(productReviewCommentsBulkDeleteCmd)
	addJSONBodyFlags(productReviewCommentsBulkDeleteCmd)
	productReviewCommentsBulkDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")
}

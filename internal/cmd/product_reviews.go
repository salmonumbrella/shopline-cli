package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var productReviewsCmd = &cobra.Command{
	Use:   "product-reviews",
	Short: "Manage product reviews",
}

var productReviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List product reviews",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		productID, _ := cmd.Flags().GetString("product-id")
		status, _ := cmd.Flags().GetString("status")
		rating, _ := cmd.Flags().GetInt("rating")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.ProductReviewsListOptions{
			Page:      page,
			PageSize:  pageSize,
			ProductID: productID,
			Status:    status,
			Rating:    rating,
		}

		resp, err := client.ListProductReviews(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list product reviews: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PRODUCT ID", "CUSTOMER", "RATING", "TITLE", "STATUS", "CREATED"}
		var rows [][]string
		for _, r := range resp.Items {
			verifiedMark := ""
			if r.Verified {
				verifiedMark = " (verified)"
			}
			rows = append(rows, []string{
				r.ID,
				r.ProductID,
				r.CustomerName + verifiedMark,
				fmt.Sprintf("%d/5", r.Rating),
				r.Title,
				r.Status,
				r.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d reviews\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var productReviewsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get product review details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		review, err := client.GetProductReview(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get product review: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(review)
		}

		fmt.Printf("Review ID:      %s\n", review.ID)
		fmt.Printf("Product ID:     %s\n", review.ProductID)
		fmt.Printf("Customer ID:    %s\n", review.CustomerID)
		fmt.Printf("Customer Name:  %s\n", review.CustomerName)
		fmt.Printf("Rating:         %d/5\n", review.Rating)
		fmt.Printf("Title:          %s\n", review.Title)
		fmt.Printf("Content:        %s\n", review.Content)
		fmt.Printf("Status:         %s\n", review.Status)
		fmt.Printf("Verified:       %t\n", review.Verified)
		fmt.Printf("Helpful Count:  %d\n", review.HelpfulCount)
		fmt.Printf("Created:        %s\n", review.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:        %s\n", review.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var productReviewsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a product review",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product-id")
		customerID, _ := cmd.Flags().GetString("customer-id")
		customerName, _ := cmd.Flags().GetString("customer-name")
		rating, _ := cmd.Flags().GetInt("rating")
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create review for product %s with rating %d/5\n", productID, rating)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.ProductReviewCreateRequest{
			ProductID:    productID,
			CustomerID:   customerID,
			CustomerName: customerName,
			Rating:       rating,
			Title:        title,
			Content:      content,
		}

		review, err := client.CreateProductReview(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create product review: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(review)
		}

		fmt.Printf("Created review %s\n", review.ID)
		fmt.Printf("Product ID:  %s\n", review.ProductID)
		fmt.Printf("Rating:      %d/5\n", review.Rating)
		fmt.Printf("Status:      %s\n", review.Status)

		return nil
	},
}

var productReviewsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a product review",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete review %s\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteProductReview(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete product review: %w", err)
		}

		fmt.Printf("Deleted review %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(productReviewsCmd)

	productReviewsCmd.AddCommand(productReviewsListCmd)
	productReviewsListCmd.Flags().String("product-id", "", "Filter by product ID")
	productReviewsListCmd.Flags().String("status", "", "Filter by status (pending, approved, rejected)")
	productReviewsListCmd.Flags().Int("rating", 0, "Filter by rating (1-5)")
	productReviewsListCmd.Flags().Int("page", 1, "Page number")
	productReviewsListCmd.Flags().Int("page-size", 20, "Results per page")

	productReviewsCmd.AddCommand(productReviewsGetCmd)

	productReviewsCmd.AddCommand(productReviewsCreateCmd)
	productReviewsCreateCmd.Flags().String("product-id", "", "Product ID")
	productReviewsCreateCmd.Flags().String("customer-id", "", "Customer ID (optional)")
	productReviewsCreateCmd.Flags().String("customer-name", "", "Customer name")
	productReviewsCreateCmd.Flags().Int("rating", 0, "Rating (1-5)")
	productReviewsCreateCmd.Flags().String("title", "", "Review title (optional)")
	productReviewsCreateCmd.Flags().String("content", "", "Review content")
	_ = productReviewsCreateCmd.MarkFlagRequired("product-id")
	_ = productReviewsCreateCmd.MarkFlagRequired("rating")
	_ = productReviewsCreateCmd.MarkFlagRequired("content")

	productReviewsCmd.AddCommand(productReviewsDeleteCmd)
}

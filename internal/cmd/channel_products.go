package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var channelProductsCmd = &cobra.Command{
	Use:   "channel-products",
	Short: "Manage multi-channel product listings",
}

var channelProductsListCmd = &cobra.Command{
	Use:   "list <channel-id>",
	Short: "List product listings in a channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")

		opts := &api.ChannelProductsListOptions{
			Page:     page,
			PageSize: pageSize,
			Status:   status,
		}

		if cmd.Flags().Changed("published") {
			published, _ := cmd.Flags().GetBool("published")
			opts.Published = &published
		}

		if cmd.Flags().Changed("available") {
			available, _ := cmd.Flags().GetBool("available")
			opts.AvailableForSale = &available
		}

		resp, err := client.ListChannelProductListings(cmd.Context(), args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to list channel products: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PRODUCT", "TITLE", "STATUS", "PUBLISHED", "AVAILABLE", "UPDATED"}
		var rows [][]string
		for _, l := range resp.Items {
			rows = append(rows, []string{
				outfmt.FormatID("channel_product", l.ID),
				l.ProductID,
				l.Title,
				l.Status,
				fmt.Sprintf("%t", l.Published),
				fmt.Sprintf("%t", l.AvailableForSale),
				l.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d product listings in channel %s\n", len(resp.Items), resp.TotalCount, args[0])
		return nil
	},
}

var channelProductsGetCmd = &cobra.Command{
	Use:   "get <channel-id> <product-id>",
	Short: "Get product listing details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		listing, err := client.GetChannelProductListing(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("failed to get product listing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(listing)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Listing ID:    %s\n", listing.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:    %s\n", listing.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Channel ID:    %s\n", listing.ChannelID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Title:         %s\n", listing.Title)
		_, _ = fmt.Fprintf(outWriter(cmd), "Handle:        %s\n", listing.Handle)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:        %s\n", listing.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Published:     %t\n", listing.Published)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available:     %t\n", listing.AvailableForSale)
		if listing.PublishedAt != nil {
			_, _ = fmt.Fprintf(outWriter(cmd), "Published At:  %s\n", listing.PublishedAt.Format(time.RFC3339))
		}
		if len(listing.Variants) > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Variants:      %d\n", len(listing.Variants))
			for _, v := range listing.Variants {
				_, _ = fmt.Fprintf(outWriter(cmd), "  - %s: %s (%s, qty: %d)\n", v.VariantID, v.Title, v.Price, v.InventoryQuantity)
			}
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:       %s\n", listing.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:       %s\n", listing.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var channelProductsPublishCmd = &cobra.Command{
	Use:   "publish <channel-id> <product-id>",
	Short: "Publish a product to a channel",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would publish product to channel listing") {
			return nil
		}

		req := &api.ChannelProductPublishRequest{
			ProductID: args[1],
		}

		listing, err := client.PublishProductToChannelListing(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to publish product: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(listing)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Published product %s to channel %s\n", listing.ProductID, listing.ChannelID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Listing ID: %s\n", listing.ID)
		return nil
	},
}

var channelProductsUnpublishCmd = &cobra.Command{
	Use:   "unpublish <channel-id> <product-id>",
	Short: "Unpublish a product from a channel",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would unpublish product from channel listing") {
			return nil
		}

		if !confirmAction(cmd, fmt.Sprintf("Unpublish product %s from channel %s? [y/N] ", args[1], args[0])) {
			_, _ = fmt.Fprintln(outWriter(cmd), "Cancelled.")
			return nil
		}

		if err := client.UnpublishProductFromChannelListing(cmd.Context(), args[0], args[1]); err != nil {
			return fmt.Errorf("failed to unpublish product: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Unpublished product %s from channel %s\n", args[1], args[0])
		return nil
	},
}

var channelProductsUpdateCmd = &cobra.Command{
	Use:   "update <channel-id> <product-id>",
	Short: "Update a product listing",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}
		if checkDryRun(cmd, "[DRY-RUN] Would update channel product listing") {
			return nil
		}

		req := &api.ChannelProductUpdateRequest{}

		if cmd.Flags().Changed("published") {
			published, _ := cmd.Flags().GetBool("published")
			req.Published = &published
		}

		if cmd.Flags().Changed("available") {
			available, _ := cmd.Flags().GetBool("available")
			req.AvailableForSale = &available
		}

		listing, err := client.UpdateChannelProductListing(cmd.Context(), args[0], args[1], req)
		if err != nil {
			return fmt.Errorf("failed to update product listing: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(listing)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated product listing %s\n", listing.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Published: %t\n", listing.Published)
		_, _ = fmt.Fprintf(outWriter(cmd), "Available: %t\n", listing.AvailableForSale)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(channelProductsCmd)

	channelProductsCmd.AddCommand(channelProductsListCmd)
	channelProductsListCmd.Flags().Int("page", 1, "Page number")
	channelProductsListCmd.Flags().Int("page-size", 20, "Results per page")
	channelProductsListCmd.Flags().Bool("published", false, "Filter by published status")
	channelProductsListCmd.Flags().Bool("available", false, "Filter by availability")
	channelProductsListCmd.Flags().String("status", "", "Filter by status (active, draft, archived)")

	channelProductsCmd.AddCommand(channelProductsGetCmd)

	channelProductsCmd.AddCommand(channelProductsPublishCmd)

	channelProductsCmd.AddCommand(channelProductsUnpublishCmd)

	channelProductsCmd.AddCommand(channelProductsUpdateCmd)
	channelProductsUpdateCmd.Flags().Bool("published", false, "Set published status")
	channelProductsUpdateCmd.Flags().Bool("available", false, "Set availability status")
}

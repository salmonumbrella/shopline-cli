package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var mediasCmd = &cobra.Command{
	Use:   "medias",
	Short: "Manage product media files",
}

var mediasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List medias",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		productID, _ := cmd.Flags().GetString("product-id")
		mediaType, _ := cmd.Flags().GetString("type")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		opts := &api.MediasListOptions{
			Page:      page,
			PageSize:  pageSize,
			ProductID: productID,
			MediaType: mediaType,
		}

		resp, err := client.ListMedias(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list medias: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "PRODUCT ID", "TYPE", "ALT", "DIMENSIONS", "CREATED"}
		var rows [][]string
		for _, m := range resp.Items {
			dimensions := ""
			if m.Width > 0 && m.Height > 0 {
				dimensions = fmt.Sprintf("%dx%d", m.Width, m.Height)
			}
			rows = append(rows, []string{
				m.ID,
				m.ProductID,
				string(m.MediaType),
				truncate(m.Alt, 30),
				dimensions,
				m.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		fmt.Printf("\nShowing %d of %d medias\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var mediasGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get media details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		media, err := client.GetMedia(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get media: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(media)
		}

		fmt.Printf("Media ID:     %s\n", media.ID)
		fmt.Printf("Product ID:   %s\n", media.ProductID)
		fmt.Printf("Type:         %s\n", media.MediaType)
		fmt.Printf("Position:     %d\n", media.Position)
		fmt.Printf("Alt:          %s\n", media.Alt)
		fmt.Printf("Src:          %s\n", media.Src)
		if media.Width > 0 && media.Height > 0 {
			fmt.Printf("Dimensions:   %dx%d\n", media.Width, media.Height)
		}
		if media.MimeType != "" {
			fmt.Printf("MIME Type:    %s\n", media.MimeType)
		}
		if media.FileSize > 0 {
			fmt.Printf("File Size:    %d bytes\n", media.FileSize)
		}
		if media.Duration > 0 {
			fmt.Printf("Duration:     %d seconds\n", media.Duration)
		}
		if media.PreviewURL != "" {
			fmt.Printf("Preview URL:  %s\n", media.PreviewURL)
		}
		if media.ExternalURL != "" {
			fmt.Printf("External URL: %s\n", media.ExternalURL)
		}
		fmt.Printf("Created:      %s\n", media.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:      %s\n", media.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

var mediasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a media",
	RunE: func(cmd *cobra.Command, args []string) error {
		productID, _ := cmd.Flags().GetString("product-id")
		mediaType, _ := cmd.Flags().GetString("type")
		src, _ := cmd.Flags().GetString("src")
		alt, _ := cmd.Flags().GetString("alt")
		position, _ := cmd.Flags().GetInt("position")
		externalURL, _ := cmd.Flags().GetString("external-url")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create %s media for product %s\n", mediaType, productID)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.MediaCreateRequest{
			ProductID:   productID,
			MediaType:   api.MediaType(mediaType),
			Src:         src,
			Alt:         alt,
			Position:    position,
			ExternalURL: externalURL,
		}

		media, err := client.CreateMedia(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create media: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(media)
		}

		fmt.Printf("Created media %s\n", media.ID)
		fmt.Printf("Product ID: %s\n", media.ProductID)
		fmt.Printf("Type:       %s\n", media.MediaType)
		fmt.Printf("Src:        %s\n", media.Src)

		return nil
	},
}

var mediasDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a media",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete media %s\n", args[0])
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete media %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteMedia(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete media: %w", err)
		}

		fmt.Printf("Deleted media %s\n", args[0])
		return nil
	},
}

// truncate shortens a string to the specified length with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(mediasCmd)

	mediasCmd.AddCommand(mediasListCmd)
	mediasListCmd.Flags().String("product-id", "", "Filter by product ID")
	mediasListCmd.Flags().String("type", "", "Filter by media type (image, video, model_3d, external_video)")
	mediasListCmd.Flags().Int("page", 1, "Page number")
	mediasListCmd.Flags().Int("page-size", 20, "Results per page")

	mediasCmd.AddCommand(mediasGetCmd)

	mediasCmd.AddCommand(mediasCreateCmd)
	mediasCreateCmd.Flags().String("product-id", "", "Product ID (required)")
	mediasCreateCmd.Flags().String("type", "image", "Media type (image, video, model_3d, external_video)")
	mediasCreateCmd.Flags().String("src", "", "Media source URL")
	mediasCreateCmd.Flags().String("alt", "", "Alt text for the media")
	mediasCreateCmd.Flags().Int("position", 0, "Position in the media list")
	mediasCreateCmd.Flags().String("external-url", "", "External URL for external_video type")
	_ = mediasCreateCmd.MarkFlagRequired("product-id")

	mediasCmd.AddCommand(mediasDeleteCmd)
}

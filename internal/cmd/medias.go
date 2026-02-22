package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
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
				outfmt.FormatID("media", m.ID),
				m.ProductID,
				string(m.MediaType),
				truncate(m.Alt, 30),
				dimensions,
				m.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d medias\n", len(resp.Items), resp.TotalCount)
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Media ID:     %s\n", media.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID:   %s\n", media.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:         %s\n", media.MediaType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Position:     %d\n", media.Position)
		_, _ = fmt.Fprintf(outWriter(cmd), "Alt:          %s\n", media.Alt)
		_, _ = fmt.Fprintf(outWriter(cmd), "Src:          %s\n", media.Src)
		if media.Width > 0 && media.Height > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Dimensions:   %dx%d\n", media.Width, media.Height)
		}
		if media.MimeType != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "MIME Type:    %s\n", media.MimeType)
		}
		if media.FileSize > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "File Size:    %d bytes\n", media.FileSize)
		}
		if media.Duration > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Duration:     %d seconds\n", media.Duration)
		}
		if media.PreviewURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Preview URL:  %s\n", media.PreviewURL)
		}
		if media.ExternalURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "External URL: %s\n", media.ExternalURL)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", media.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", media.UpdatedAt.Format(time.RFC3339))

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

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create %s media for product %s", mediaType, productID)) {
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

		_, _ = fmt.Fprintf(outWriter(cmd), "Created media %s\n", media.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID: %s\n", media.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:       %s\n", media.MediaType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Src:        %s\n", media.Src)

		return nil
	},
}

var mediasUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a media",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update media %s", args[0])) {
			return nil
		}

		var req api.MediaUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		media, err := client.UpdateMedia(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update media: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(media)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated media %s\n", media.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Product ID: %s\n", media.ProductID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Type:       %s\n", media.MediaType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Src:        %s\n", media.Src)
		return nil
	},
}

var mediasDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a media",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete media %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete media %s? Use --yes to confirm.\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteMedia(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete media: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted media %s\n", args[0])
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

	mediasCmd.AddCommand(mediasUpdateCmd)
	addJSONBodyFlags(mediasUpdateCmd)
	mediasUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	mediasCmd.AddCommand(mediasDeleteCmd)

	schema.Register(schema.Resource{
		Name:        "medias",
		Description: "Manage product media files",
		Commands:    []string{"list", "get", "create", "update", "delete"},
		IDPrefix:    "media",
	})
}

package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage theme assets",
}

var assetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List assets for a theme",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		themeID, _ := cmd.Flags().GetString("theme-id")

		resp, err := client.ListAssets(cmd.Context(), themeID)
		if err != nil {
			return fmt.Errorf("failed to list assets: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"KEY", "CONTENT TYPE", "SIZE", "UPDATED"}
		var rows [][]string
		for _, a := range resp.Items {
			rows = append(rows, []string{
				a.Key,
				a.ContentType,
				fmt.Sprintf("%d", a.Size),
				a.UpdatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d assets\n", len(resp.Items))
		return nil
	},
}

var assetsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get asset details",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		themeID, _ := cmd.Flags().GetString("theme-id")
		key, _ := cmd.Flags().GetString("key")

		asset, err := client.GetAsset(cmd.Context(), themeID, key)
		if err != nil {
			return fmt.Errorf("failed to get asset: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(asset)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Key:          %s\n", asset.Key)
		_, _ = fmt.Fprintf(outWriter(cmd), "Theme ID:     %s\n", asset.ThemeID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Content Type: %s\n", asset.ContentType)
		_, _ = fmt.Fprintf(outWriter(cmd), "Size:         %d bytes\n", asset.Size)
		_, _ = fmt.Fprintf(outWriter(cmd), "Checksum:     %s\n", asset.Checksum)
		if asset.PublicURL != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Public URL:   %s\n", asset.PublicURL)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:      %s\n", asset.CreatedAt.Format("2006-01-02 15:04:05"))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:      %s\n", asset.UpdatedAt.Format("2006-01-02 15:04:05"))
		if asset.Value != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "\n--- Content ---\n%s\n", asset.Value)
		}
		return nil
	},
}

var assetsPutCmd = &cobra.Command{
	Use:   "put",
	Short: "Create or update an asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		themeID, _ := cmd.Flags().GetString("theme-id")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create/update asset %s in theme %s", key, themeID)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AssetUpdateRequest{
			Key:   key,
			Value: value,
		}

		asset, err := client.UpdateAsset(cmd.Context(), themeID, req)
		if err != nil {
			return fmt.Errorf("failed to update asset: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(asset)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated asset %s\n", asset.Key)
		return nil
	},
}

var assetsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		themeID, _ := cmd.Flags().GetString("theme-id")
		key, _ := cmd.Flags().GetString("key")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete asset %s from theme %s", key, themeID)) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete asset %s? (use --yes to confirm)\n", key)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteAsset(cmd.Context(), themeID, key); err != nil {
			return fmt.Errorf("failed to delete asset: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted asset %s\n", key)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(assetsCmd)

	assetsCmd.AddCommand(assetsListCmd)
	assetsListCmd.Flags().String("theme-id", "", "Theme ID (required)")
	_ = assetsListCmd.MarkFlagRequired("theme-id")

	assetsCmd.AddCommand(assetsGetCmd)
	assetsGetCmd.Flags().String("theme-id", "", "Theme ID (required)")
	assetsGetCmd.Flags().String("key", "", "Asset key (required)")
	_ = assetsGetCmd.MarkFlagRequired("theme-id")
	_ = assetsGetCmd.MarkFlagRequired("key")

	assetsCmd.AddCommand(assetsPutCmd)
	assetsPutCmd.Flags().String("theme-id", "", "Theme ID (required)")
	assetsPutCmd.Flags().String("key", "", "Asset key (required)")
	assetsPutCmd.Flags().String("value", "", "Asset content")
	_ = assetsPutCmd.MarkFlagRequired("theme-id")
	_ = assetsPutCmd.MarkFlagRequired("key")

	assetsCmd.AddCommand(assetsDeleteCmd)
	assetsDeleteCmd.Flags().String("theme-id", "", "Theme ID (required)")
	assetsDeleteCmd.Flags().String("key", "", "Asset key (required)")
	_ = assetsDeleteCmd.MarkFlagRequired("theme-id")
	_ = assetsDeleteCmd.MarkFlagRequired("key")
}

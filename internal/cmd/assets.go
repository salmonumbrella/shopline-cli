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
		fmt.Printf("\nShowing %d assets\n", len(resp.Items))
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

		fmt.Printf("Key:          %s\n", asset.Key)
		fmt.Printf("Theme ID:     %s\n", asset.ThemeID)
		fmt.Printf("Content Type: %s\n", asset.ContentType)
		fmt.Printf("Size:         %d bytes\n", asset.Size)
		fmt.Printf("Checksum:     %s\n", asset.Checksum)
		if asset.PublicURL != "" {
			fmt.Printf("Public URL:   %s\n", asset.PublicURL)
		}
		fmt.Printf("Created:      %s\n", asset.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:      %s\n", asset.UpdatedAt.Format("2006-01-02 15:04:05"))
		if asset.Value != "" {
			fmt.Printf("\n--- Content ---\n%s\n", asset.Value)
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would create/update asset %s in theme %s\n", key, themeID)
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

		fmt.Printf("Updated asset %s\n", asset.Key)
		return nil
	},
}

var assetsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		themeID, _ := cmd.Flags().GetString("theme-id")
		key, _ := cmd.Flags().GetString("key")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("[DRY-RUN] Would delete asset %s from theme %s\n", key, themeID)
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to delete asset %s? (use --yes to confirm)\n", key)
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteAsset(cmd.Context(), themeID, key); err != nil {
			return fmt.Errorf("failed to delete asset: %w", err)
		}

		fmt.Printf("Deleted asset %s\n", key)
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

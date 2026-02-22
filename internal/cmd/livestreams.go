package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var livestreamsCmd = &cobra.Command{
	Use:     "livestreams",
	Aliases: []string{"live", "livestream", "streams"},
	Short:   "Manage Shopline livestream sales (via Admin API)",
}

var livestreamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List livestreams",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		salesType, _ := cmd.Flags().GetString("type")

		opts := &api.AdminListStreamsOptions{
			PageNum:   page,
			PageSize:  pageSize,
			SalesType: salesType,
		}

		result, err := client.ListLivestreams(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list livestreams: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsGetCmd = &cobra.Command{
	Use:   "get <stream-id>",
	Short: "Get livestream details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.GetLivestream(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new livestream",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create livestream") {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		owner, _ := cmd.Flags().GetString("owner")
		description, _ := cmd.Flags().GetString("description")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		lockTime, _ := cmd.Flags().GetString("lock-inventory-time")
		checkoutTime, _ := cmd.Flags().GetString("checkout-time")
		checkoutMsg, _ := cmd.Flags().GetString("checkout-message")
		platform, _ := cmd.Flags().GetString("platform")
		image, _ := cmd.Flags().GetString("image")

		req := &api.AdminCreateStreamRequest{
			Title:             title,
			SalesOwner:        owner,
			SalesDescription:  description,
			StartDate:         startDate,
			EndDate:           endDate,
			LockInventoryTime: lockTime,
			CheckoutTime:      checkoutTime,
			CheckoutMessage:   checkoutMsg,
			Platform:          platform,
			ImageServePath:    image,
		}

		result, err := client.CreateLivestream(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create livestream: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsUpdateCmd = &cobra.Command{
	Use:   "update <stream-id>",
	Short: "Update a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update livestream %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		postTitle, _ := cmd.Flags().GetString("post-title")
		postOwner, _ := cmd.Flags().GetString("post-owner")
		postDescription, _ := cmd.Flags().GetString("post-description")
		checkoutTime, _ := cmd.Flags().GetString("checkout-time")
		lockTime, _ := cmd.Flags().GetString("lock-inventory-time")
		archiveTime, _ := cmd.Flags().GetString("archive-visible-time")

		req := &api.AdminUpdateStreamRequest{
			PostSalesTitle:            postTitle,
			PostSalesOwner:            postOwner,
			PostSalesDescription:      postDescription,
			CheckoutTime:              checkoutTime,
			LockInventoryTime:         lockTime,
			ArchivedStreamVisibleTime: archiveTime,
		}

		result, err := client.UpdateLivestream(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to update livestream: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsDeleteCmd = &cobra.Command{
	Use:   "delete <stream-id>",
	Short: "Delete a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete livestream %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		err = client.DeleteLivestream(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete livestream: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted livestream %s\n", args[0])
		return nil
	},
}

var livestreamsAddProductsCmd = &cobra.Command{
	Use:   "add-products <stream-id>",
	Short: "Add products to a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add products to livestream %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		var req api.AdminAddStreamProductsRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		result, err := client.AddStreamProducts(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to add products: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsRemoveProductsCmd = &cobra.Command{
	Use:   "remove-products <stream-id>",
	Short: "Remove products from a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		productIDs, _ := cmd.Flags().GetStringSlice("product-ids")
		if len(productIDs) == 0 {
			return fmt.Errorf("--product-ids is required")
		}

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would remove products from livestream %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		req := &api.AdminRemoveStreamProductsRequest{
			ProductIDs: productIDs,
		}

		err = client.RemoveStreamProducts(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to remove products: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Products removed from livestream %s\n", args[0])
		return nil
	},
}

var livestreamsStartCmd = &cobra.Command{
	Use:   "start <stream-id>",
	Short: "Start a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would start livestream %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		platform, _ := cmd.Flags().GetString("platform")
		videoDataStr, _ := cmd.Flags().GetString("video-data")

		req := &api.AdminStartStreamRequest{
			Platform: platform,
		}

		if videoDataStr != "" {
			var vd api.AdminVideoData
			if err := json.Unmarshal([]byte(videoDataStr), &vd); err != nil {
				return fmt.Errorf("invalid --video-data JSON: %w", err)
			}
			req.VideoData = &vd
		}

		result, err := client.StartLivestream(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to start livestream: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsEndCmd = &cobra.Command{
	Use:   "end <stream-id>",
	Short: "End a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would end livestream %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		err = client.EndLivestream(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to end livestream: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Ended livestream %s\n", args[0])
		return nil
	},
}

var livestreamsCommentsCmd = &cobra.Command{
	Use:   "comments <stream-id>",
	Short: "Get comments for a livestream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")

		result, err := client.GetStreamComments(cmd.Context(), args[0], page)
		if err != nil {
			return fmt.Errorf("failed to get stream comments: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsActiveVideosCmd = &cobra.Command{
	Use:   "active-videos <stream-id>",
	Short: "Get active livestream videos for a platform",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		platform, _ := cmd.Flags().GetString("platform")
		platform = strings.ToUpper(strings.TrimSpace(platform))
		if platform != "FACEBOOK" && platform != "INSTAGRAM" {
			return fmt.Errorf("--platform must be FACEBOOK or INSTAGRAM")
		}
		result, err := client.GetStreamActiveVideos(cmd.Context(), args[0], platform)
		if err != nil {
			return fmt.Errorf("failed to get active videos: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var livestreamsToggleProductCmd = &cobra.Command{
	Use:   "toggle-product <stream-id> <product-id>",
	Short: "Toggle livestream product display status",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would toggle product %s for livestream %s", args[1], args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		status = strings.ToUpper(strings.TrimSpace(status))
		if status != "HIDDEN" && status != "DISPLAYING" {
			return fmt.Errorf("--status must be HIDDEN or DISPLAYING")
		}
		req := &api.AdminToggleStreamProductRequest{Status: status}
		result, err := client.ToggleStreamProductDisplay(cmd.Context(), args[0], args[1], req)
		if err != nil {
			return fmt.Errorf("failed to toggle product display: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(livestreamsCmd)

	livestreamsCmd.AddCommand(livestreamsListCmd)
	livestreamsListCmd.Flags().Int("page", 1, "Page number")
	livestreamsListCmd.Flags().Int("page-size", 20, "Results per page")
	livestreamsListCmd.Flags().String("type", "", "Sales type: LIVE or POST")

	livestreamsCmd.AddCommand(livestreamsGetCmd)

	livestreamsCmd.AddCommand(livestreamsCreateCmd)
	livestreamsCreateCmd.Flags().String("title", "", "Livestream title (required)")
	livestreamsCreateCmd.Flags().String("owner", "", "Sales owner (required)")
	livestreamsCreateCmd.Flags().String("description", "", "Sales description")
	livestreamsCreateCmd.Flags().String("start-date", "", "Start date (ISO 8601)")
	livestreamsCreateCmd.Flags().String("end-date", "", "End date (ISO 8601)")
	livestreamsCreateCmd.Flags().String("lock-inventory-time", "", "Lock inventory time")
	livestreamsCreateCmd.Flags().String("checkout-time", "", "Checkout time")
	livestreamsCreateCmd.Flags().String("checkout-message", "", "Checkout message")
	livestreamsCreateCmd.Flags().String("platform", "", "Platform: FACEBOOK, LINE, or INSTAGRAM (required)")
	livestreamsCreateCmd.Flags().String("image", "", "Image serve path")
	_ = livestreamsCreateCmd.MarkFlagRequired("title")
	_ = livestreamsCreateCmd.MarkFlagRequired("owner")
	_ = livestreamsCreateCmd.MarkFlagRequired("platform")
	livestreamsCreateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsUpdateCmd)
	livestreamsUpdateCmd.Flags().String("post-title", "", "Post-sales title")
	livestreamsUpdateCmd.Flags().String("post-owner", "", "Post-sales owner")
	livestreamsUpdateCmd.Flags().String("post-description", "", "Post-sales description")
	livestreamsUpdateCmd.Flags().String("checkout-time", "", "Checkout time")
	livestreamsUpdateCmd.Flags().String("lock-inventory-time", "", "Lock inventory time")
	livestreamsUpdateCmd.Flags().String("archive-visible-time", "", "Archived stream visible time")
	livestreamsUpdateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsDeleteCmd)
	livestreamsDeleteCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsAddProductsCmd)
	addJSONBodyFlags(livestreamsAddProductsCmd)
	livestreamsAddProductsCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsRemoveProductsCmd)
	livestreamsRemoveProductsCmd.Flags().StringSlice("product-ids", nil, "Product IDs to remove (comma-separated)")
	_ = livestreamsRemoveProductsCmd.MarkFlagRequired("product-ids")
	livestreamsRemoveProductsCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsStartCmd)
	livestreamsStartCmd.Flags().String("platform", "", "Platform: FACEBOOK, LINE, or INSTAGRAM (required)")
	livestreamsStartCmd.Flags().String("video-data", "", "Video data as JSON (for Facebook/Instagram)")
	_ = livestreamsStartCmd.MarkFlagRequired("platform")
	livestreamsStartCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsEndCmd)
	livestreamsEndCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	livestreamsCmd.AddCommand(livestreamsCommentsCmd)
	livestreamsCommentsCmd.Flags().Int("page", 1, "Page number")

	livestreamsCmd.AddCommand(livestreamsActiveVideosCmd)
	livestreamsActiveVideosCmd.Flags().String("platform", "", "Platform: FACEBOOK or INSTAGRAM (required)")
	_ = livestreamsActiveVideosCmd.MarkFlagRequired("platform")

	livestreamsCmd.AddCommand(livestreamsToggleProductCmd)
	livestreamsToggleProductCmd.Flags().String("status", "", "Display status: HIDDEN or DISPLAYING (required)")
	_ = livestreamsToggleProductCmd.MarkFlagRequired("status")
	livestreamsToggleProductCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	schema.Register(schema.Resource{
		Name:        "livestreams",
		Description: "Manage Shopline livestream sales (via Admin API)",
		Commands:    []string{"list", "get", "create", "update", "delete", "add-products", "remove-products", "start", "end", "comments", "active-videos", "toggle-product"},
	})
}

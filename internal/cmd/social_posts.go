package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var socialPostsCmd = &cobra.Command{
	Use:     "social-posts",
	Aliases: []string{"social"},
	Short:   "Manage social media sales events and channels (via Admin API)",
}

// --- Top-level commands ---

var socialPostsChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "List social channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.GetSocialChannels(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get social channels: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsChannelPostsCmd = &cobra.Command{
	Use:     "channel-posts",
	Aliases: []string{"chposts", "cposts"},
	Short:   "List posts for a social channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		channelID, _ := cmd.Flags().GetString("channel-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		postType, _ := cmd.Flags().GetString("type")
		since, _ := cmd.Flags().GetString("since")
		before, _ := cmd.Flags().GetString("before")
		after, _ := cmd.Flags().GetString("after")

		opts := &api.SocialChannelPostsOptions{
			PartyChannelID: channelID,
			PageSize:       pageSize,
			Type:           postType,
			Since:          since,
			Before:         before,
			After:          after,
		}

		result, err := client.GetChannelPosts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to get channel posts: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List social post categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		result, err := client.GetSocialCategories(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get social categories: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

// --- Products subgroup ---

var socialPostsProductsCmd = &cobra.Command{
	Use:   "products",
	Short: "Social posts product operations",
}

var socialPostsProductsSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search social products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		query, _ := cmd.Flags().GetString("q")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		searchType, _ := cmd.Flags().GetString("search-type")
		categoryIDs, _ := cmd.Flags().GetStringSlice("category-ids")

		opts := &api.SocialProductSearchOptions{
			Query:       query,
			Page:        page,
			PageSize:    pageSize,
			SearchType:  searchType,
			CategoryIDs: categoryIDs,
		}

		result, err := client.SearchSocialProducts(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to search social products: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

// --- Events subgroup ---

var socialPostsEventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event", "ev"},
	Short:   "Sales event operations",
}

var socialPostsEventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sales events",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		salesType, _ := cmd.Flags().GetString("type")

		opts := &api.SalesEventListOptions{
			PageNum:   page,
			PageSize:  pageSize,
			SalesType: salesType,
		}

		result, err := client.ListSalesEvents(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list sales events: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get sales event details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		fieldScopes, _ := cmd.Flags().GetString("field-scopes")

		result, err := client.GetSalesEvent(cmd.Context(), args[0], fieldScopes)
		if err != nil {
			return fmt.Errorf("failed to get sales event: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a sales event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, "[DRY-RUN] Would create sales event") {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		platform, _ := cmd.Flags().GetString("platform")
		title, _ := cmd.Flags().GetString("title")
		patternModel, _ := cmd.Flags().GetString("pattern-model")
		platforms, _ := cmd.Flags().GetStringSlice("platforms")
		postSubType, _ := cmd.Flags().GetString("post-sub-type")

		if len(platforms) == 0 {
			platforms = []string{platform}
		}

		req := &api.CreateSalesEventRequest{
			Type:         1,
			Platform:     platform,
			Title:        title,
			PatternModel: patternModel,
			Platforms:    platforms,
			PostSubType:  postSubType,
		}

		result, err := client.CreateSalesEvent(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create sales event: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsScheduleCmd = &cobra.Command{
	Use:   "schedule <id>",
	Short: "Schedule a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would schedule sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		startTime, _ := cmd.Flags().GetInt64("start-time")
		endTime, _ := cmd.Flags().GetInt64("end-time")

		req := &api.ScheduleSalesEventRequest{
			StartTime: startTime,
			EndTime:   endTime,
		}

		err = client.ScheduleSalesEvent(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to schedule sales event: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Scheduled sales event %s\n", args[0])
		return nil
	},
}

var socialPostsEventsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		err = client.DeleteSalesEvent(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete sales event: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted sales event %s\n", args[0])
		return nil
	},
}

var socialPostsEventsPublishCmd = &cobra.Command{
	Use:   "publish <id>",
	Short: "Publish a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would publish sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		err = client.PublishSalesEvent(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to publish sales event: %w", err)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Published sales event %s\n", args[0])
		return nil
	},
}

var socialPostsEventsAddProductsCmd = &cobra.Command{
	Use:   "add-products <id>",
	Short: "Add products to a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would add products to sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		var req api.AddSalesEventProductsRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		result, err := client.AddSalesEventProducts(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to add products to sales event: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsUpdateKeysCmd = &cobra.Command{
	Use:   "update-keys <id>",
	Short: "Update product keywords in a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update product keys for sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		var req api.UpdateProductKeysRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		result, err := client.UpdateSalesEventProductKeys(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update product keys: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsLinkFacebookCmd = &cobra.Command{
	Use:   "link-facebook <id>",
	Short: "Link a Facebook post to a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would link Facebook post to sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		var req api.LinkFacebookPostRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		result, err := client.LinkFacebookPost(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to link Facebook post: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsLinkInstagramCmd = &cobra.Command{
	Use:   "link-instagram <id>",
	Short: "Link an Instagram post to a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would link Instagram post to sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		var req api.LinkInstagramPostRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		result, err := client.LinkInstagramPost(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to link Instagram post: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var socialPostsEventsLinkFBGroupCmd = &cobra.Command{
	Use:   "link-fb-group <id>",
	Short: "Link a Facebook Group post to a sales event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would link FB group post to sales event %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		pageID, _ := cmd.Flags().GetString("page-id")
		relationURL, _ := cmd.Flags().GetString("relation-url")

		req := &api.LinkFBGroupPostRequest{
			PageID:      pageID,
			RelationURL: relationURL,
		}

		result, err := client.LinkFBGroupPost(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to link FB group post: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(socialPostsCmd)

	// Channels
	socialPostsCmd.AddCommand(socialPostsChannelsCmd)

	// Channel Posts
	socialPostsCmd.AddCommand(socialPostsChannelPostsCmd)
	socialPostsChannelPostsCmd.Flags().String("channel-id", "", "Party channel ID (required)")
	socialPostsChannelPostsCmd.Flags().Int("page-size", 10, "Results per page")
	socialPostsChannelPostsCmd.Flags().String("type", "POST", "Post type")
	socialPostsChannelPostsCmd.Flags().String("since", "", "Cursor: since")
	socialPostsChannelPostsCmd.Flags().String("before", "", "Cursor: before")
	socialPostsChannelPostsCmd.Flags().String("after", "", "Cursor: after")
	_ = socialPostsChannelPostsCmd.MarkFlagRequired("channel-id")

	// Categories
	socialPostsCmd.AddCommand(socialPostsCategoriesCmd)

	// Products
	socialPostsCmd.AddCommand(socialPostsProductsCmd)
	socialPostsProductsCmd.AddCommand(socialPostsProductsSearchCmd)
	socialPostsProductsSearchCmd.Flags().String("q", "", "Search query")
	socialPostsProductsSearchCmd.Flags().Int("page", 1, "Page number")
	socialPostsProductsSearchCmd.Flags().Int("page-size", 100, "Results per page")
	socialPostsProductsSearchCmd.Flags().String("search-type", "", "Search type")
	socialPostsProductsSearchCmd.Flags().StringSlice("category-ids", nil, "Category IDs")

	// Events
	socialPostsCmd.AddCommand(socialPostsEventsCmd)

	socialPostsEventsCmd.AddCommand(socialPostsEventsListCmd)
	socialPostsEventsListCmd.Flags().Int("page", 1, "Page number")
	socialPostsEventsListCmd.Flags().Int("page-size", 20, "Results per page")
	socialPostsEventsListCmd.Flags().String("type", "POST", "Sales type")

	socialPostsEventsCmd.AddCommand(socialPostsEventsGetCmd)
	socialPostsEventsGetCmd.Flags().String("field-scopes", "DETAILS,PRODUCT_NUM,SALES_CONFIG,PRODUCT_LIST", "Field scopes to include")

	socialPostsEventsCmd.AddCommand(socialPostsEventsCreateCmd)
	socialPostsEventsCreateCmd.Flags().String("platform", "", "Platform (required)")
	socialPostsEventsCreateCmd.Flags().String("title", "", "Event title (required)")
	socialPostsEventsCreateCmd.Flags().String("pattern-model", "EXACT_MATCH", "Pattern model")
	socialPostsEventsCreateCmd.Flags().StringSlice("platforms", nil, "Platforms (defaults to [platform])")
	socialPostsEventsCreateCmd.Flags().String("post-sub-type", "", "Post sub-type")
	_ = socialPostsEventsCreateCmd.MarkFlagRequired("platform")
	_ = socialPostsEventsCreateCmd.MarkFlagRequired("title")
	socialPostsEventsCreateCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsScheduleCmd)
	socialPostsEventsScheduleCmd.Flags().Int64("start-time", 0, "Start time (Unix seconds, required)")
	socialPostsEventsScheduleCmd.Flags().Int64("end-time", 0, "End time (Unix seconds, required)")
	_ = socialPostsEventsScheduleCmd.MarkFlagRequired("start-time")
	_ = socialPostsEventsScheduleCmd.MarkFlagRequired("end-time")
	socialPostsEventsScheduleCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsDeleteCmd)
	socialPostsEventsDeleteCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsPublishCmd)
	socialPostsEventsPublishCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsAddProductsCmd)
	addJSONBodyFlags(socialPostsEventsAddProductsCmd)
	socialPostsEventsAddProductsCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsUpdateKeysCmd)
	addJSONBodyFlags(socialPostsEventsUpdateKeysCmd)
	socialPostsEventsUpdateKeysCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsLinkFacebookCmd)
	addJSONBodyFlags(socialPostsEventsLinkFacebookCmd)
	socialPostsEventsLinkFacebookCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsLinkInstagramCmd)
	addJSONBodyFlags(socialPostsEventsLinkInstagramCmd)
	socialPostsEventsLinkInstagramCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	socialPostsEventsCmd.AddCommand(socialPostsEventsLinkFBGroupCmd)
	socialPostsEventsLinkFBGroupCmd.Flags().String("page-id", "", "Facebook page ID (required)")
	socialPostsEventsLinkFBGroupCmd.Flags().String("relation-url", "", "Relation URL (required)")
	_ = socialPostsEventsLinkFBGroupCmd.MarkFlagRequired("page-id")
	_ = socialPostsEventsLinkFBGroupCmd.MarkFlagRequired("relation-url")
	socialPostsEventsLinkFBGroupCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	schema.Register(schema.Resource{
		Name:        "social-posts",
		Description: "Manage social media sales events and channels (via Admin API)",
		Commands: []string{
			"channels", "channel-posts", "categories",
			"products search",
			"events list", "events get", "events create", "events schedule",
			"events delete", "events publish", "events add-products",
			"events update-keys", "events link-facebook", "events link-instagram",
			"events link-fb-group",
		},
	})
}

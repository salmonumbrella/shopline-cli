package cmd

import (
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/schema"
	"github.com/spf13/cobra"
)

var messageCenterCmd = &cobra.Command{
	Use:     "message-center",
	Aliases: []string{"mc", "messages"},
	Short:   "Manage Shopline message center conversations (via Admin API)",
}

var messageCenterListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "conversations"},
	Short:   "List message center conversations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		platform, _ := cmd.Flags().GetString("platform")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		state, _ := cmd.Flags().GetString("state")
		searchType, _ := cmd.Flags().GetString("search-type")
		query, _ := cmd.Flags().GetString("search-query")

		opts := &api.AdminListConversationsOptions{
			Platform:    platform,
			PageNum:     page,
			PageSize:    pageSize,
			StateFilter: state,
			SearchType:  searchType,
			Query:       query,
		}

		if cmd.Flags().Changed("archived") {
			archived, _ := cmd.Flags().GetBool("archived")
			opts.IsArchived = &archived
		}

		result, err := client.ListConversations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list conversations: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterSendCmd = &cobra.Command{
	Use:   "send <conversation-id>",
	Short: "Send a shop/order message reply",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would send message to conversation %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		platform, _ := cmd.Flags().GetString("platform")
		content, _ := cmd.Flags().GetString("content")

		req := &api.AdminSendMessageRequest{
			Platform: platform,
			Type:     "message",
			Content:  content,
		}

		result, err := client.SendMessage(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "List connected message-center channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.GetMessageCenterChannels(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get channels: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterStaffInfoCmd = &cobra.Command{
	Use:   "staff-info",
	Short: "Get current message-center staff info",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.GetMessageCenterStaffInfo(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get staff info: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterProfileCmd = &cobra.Command{
	Use:   "profile <scope-id>",
	Short: "Get message-center customer profile by scope ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}
		result, err := client.GetMessageCenterProfile(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get profile: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterInstantListCmd = &cobra.Command{
	Use:   "instant-list",
	Short: "List instant-message conversations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		searchType, _ := cmd.Flags().GetString("search-type")
		route, _ := cmd.Flags().GetString("route")
		unreadType, _ := cmd.Flags().GetString("unread-type")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		partyChannelIDs, _ := cmd.Flags().GetStringSlice("party-channel-id-list")

		opts := &api.AdminListInstantMessagesOptions{
			Page:               page,
			SearchType:         searchType,
			Route:              route,
			UnreadType:         unreadType,
			PageSize:           pageSize,
			PartyChannelIDList: partyChannelIDs,
		}

		result, err := client.ListInstantMessages(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list instant messages: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterInstantMessagesCmd = &cobra.Command{
	Use:   "instant-messages <conversation-id>",
	Short: "Get messages from an instant-message conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		searchType, _ := cmd.Flags().GetString("search-type")
		useMessageID, _ := cmd.Flags().GetString("use-message-id")
		createTime, _ := cmd.Flags().GetString("create-time")

		query := &api.AdminInstantMessagesQuery{
			SearchType:   searchType,
			UseMessageID: useMessageID,
			CreateTime:   createTime,
		}

		result, err := client.GetInstantMessages(cmd.Context(), args[0], query)
		if err != nil {
			return fmt.Errorf("failed to get instant messages: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

var messageCenterInstantSendCmd = &cobra.Command{
	Use:   "instant-send <conversation-id>",
	Short: "Send an instant message reply",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would send instant message to conversation %s", args[0])) {
			return nil
		}

		client, err := getAdminClient(cmd)
		if err != nil {
			return err
		}

		content, _ := cmd.Flags().GetString("content")
		senderType, _ := cmd.Flags().GetString("sender-type")
		source, _ := cmd.Flags().GetString("source")

		req := &api.AdminSendInstantMessageRequest{
			ConversationID: args[0],
			Content:        content,
			SenderTypeEnum: senderType,
			MessageSource:  source,
		}

		result, err := client.SendInstantMessage(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to send instant message: %w", err)
		}
		formatter := getFormatter(cmd)
		return formatter.JSON(result)
	},
}

func init() {
	rootCmd.AddCommand(messageCenterCmd)

	messageCenterCmd.AddCommand(messageCenterListCmd)
	messageCenterListCmd.Flags().String("platform", "", "Platform: shop_messages or order_messages")
	messageCenterListCmd.Flags().Int("page", 1, "Page number")
	messageCenterListCmd.Flags().Int("page-size", 24, "Results per page")
	messageCenterListCmd.Flags().String("state", "", "Filter: all, unread, read, follow_up")
	messageCenterListCmd.Flags().Bool("archived", false, "Show archived conversations")
	messageCenterListCmd.Flags().String("search-type", "", "Search type: message, conversation, or order")
	messageCenterListCmd.Flags().String("search-query", "", "Search query text")

	messageCenterCmd.AddCommand(messageCenterSendCmd)
	messageCenterSendCmd.Flags().String("platform", "", "Platform: order_messages or shop_messages (required)")
	messageCenterSendCmd.Flags().String("content", "", "Message content (required)")
	_ = messageCenterSendCmd.MarkFlagRequired("platform")
	_ = messageCenterSendCmd.MarkFlagRequired("content")
	messageCenterSendCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	messageCenterCmd.AddCommand(messageCenterChannelsCmd)
	messageCenterCmd.AddCommand(messageCenterStaffInfoCmd)
	messageCenterCmd.AddCommand(messageCenterProfileCmd)

	messageCenterCmd.AddCommand(messageCenterInstantListCmd)
	messageCenterInstantListCmd.Flags().Int("page", 1, "Page number")
	messageCenterInstantListCmd.Flags().String("search-type", "chat_history", "Search type")
	messageCenterInstantListCmd.Flags().String("route", "hot", "Conversation route")
	messageCenterInstantListCmd.Flags().String("unread-type", "last_unread_message", "Unread filter type")
	messageCenterInstantListCmd.Flags().Int("page-size", 20, "Results per page")
	messageCenterInstantListCmd.Flags().StringSlice("party-channel-id-list", nil, "Filter by party channel IDs")

	messageCenterCmd.AddCommand(messageCenterInstantMessagesCmd)
	messageCenterInstantMessagesCmd.Flags().String("search-type", "", "Pagination direction: up or down (required)")
	messageCenterInstantMessagesCmd.Flags().String("use-message-id", "", "Message ID cursor (required)")
	messageCenterInstantMessagesCmd.Flags().String("create-time", "", "Message create time cursor (required)")
	_ = messageCenterInstantMessagesCmd.MarkFlagRequired("search-type")
	_ = messageCenterInstantMessagesCmd.MarkFlagRequired("use-message-id")
	_ = messageCenterInstantMessagesCmd.MarkFlagRequired("create-time")

	messageCenterCmd.AddCommand(messageCenterInstantSendCmd)
	messageCenterInstantSendCmd.Flags().String("content", "", "Message content (required)")
	messageCenterInstantSendCmd.Flags().String("sender-type", "", "Optional sender type enum")
	messageCenterInstantSendCmd.Flags().String("source", "", "Optional message source")
	_ = messageCenterInstantSendCmd.MarkFlagRequired("content")
	messageCenterInstantSendCmd.Flags().Bool("dry-run", false, "Preview without making changes")

	schema.Register(schema.Resource{
		Name:        "message-center",
		Description: "Manage Shopline message center conversations (via Admin API)",
		Commands:    []string{"list", "send", "channels", "staff-info", "profile", "instant-list", "instant-messages", "instant-send"},
	})
}

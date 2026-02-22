package cmd

import (
	"fmt"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var conversationsCmd = &cobra.Command{
	Use:   "conversations",
	Short: "Manage customer conversations/chat",
}

var conversationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List conversations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		status, _ := cmd.Flags().GetString("status")
		channel, _ := cmd.Flags().GetString("channel")
		customerID, _ := cmd.Flags().GetString("customer-id")
		assigneeID, _ := cmd.Flags().GetString("assignee-id")

		opts := &api.ConversationsListOptions{
			Page:       page,
			PageSize:   pageSize,
			Status:     status,
			Channel:    channel,
			CustomerID: customerID,
			AssigneeID: assigneeID,
		}

		resp, err := client.ListConversations(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("failed to list conversations: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "CUSTOMER", "STATUS", "CHANNEL", "MESSAGES", "LAST MESSAGE"}
		var rows [][]string
		for _, c := range resp.Items {
			lastMsg := "-"
			if !c.LastMessageAt.IsZero() {
				lastMsg = c.LastMessageAt.Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{
				outfmt.FormatID("conversation", c.ID),
				c.CustomerName,
				c.Status,
				c.Channel,
				fmt.Sprintf("%d", c.MessageCount),
				lastMsg,
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d conversations\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var conversationsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get conversation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		conversation, err := client.GetConversation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get conversation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(conversation)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Conversation ID: %s\n", conversation.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Customer:        %s\n", conversation.CustomerName)
		if conversation.CustomerEmail != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Email:           %s\n", conversation.CustomerEmail)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:          %s\n", conversation.Status)
		_, _ = fmt.Fprintf(outWriter(cmd), "Channel:         %s\n", conversation.Channel)
		if conversation.Subject != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Subject:         %s\n", conversation.Subject)
		}
		if conversation.AssigneeName != "" {
			_, _ = fmt.Fprintf(outWriter(cmd), "Assignee:        %s\n", conversation.AssigneeName)
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Messages:        %d\n", conversation.MessageCount)
		if conversation.UnreadCount > 0 {
			_, _ = fmt.Fprintf(outWriter(cmd), "Unread:          %d\n", conversation.UnreadCount)
		}
		if !conversation.LastMessageAt.IsZero() {
			_, _ = fmt.Fprintf(outWriter(cmd), "Last Message:    %s\n", conversation.LastMessageAt.Format(time.RFC3339))
		}
		_, _ = fmt.Fprintf(outWriter(cmd), "Created:         %s\n", conversation.CreatedAt.Format(time.RFC3339))
		_, _ = fmt.Fprintf(outWriter(cmd), "Updated:         %s\n", conversation.UpdatedAt.Format(time.RFC3339))
		return nil
	},
}

var conversationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a conversation",
	RunE: func(cmd *cobra.Command, args []string) error {
		customerID, _ := cmd.Flags().GetString("customer-id")
		channel, _ := cmd.Flags().GetString("channel")
		subject, _ := cmd.Flags().GetString("subject")
		message, _ := cmd.Flags().GetString("message")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would create conversation for customer %s via %s", customerID, channel)) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.ConversationCreateRequest{
			CustomerID: customerID,
			Channel:    channel,
			Subject:    subject,
			Message:    message,
		}

		conversation, err := client.CreateConversation(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create conversation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(conversation)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Created conversation %s\n", conversation.ID)
		_, _ = fmt.Fprintf(outWriter(cmd), "Channel: %s\n", conversation.Channel)
		_, _ = fmt.Fprintf(outWriter(cmd), "Status:  %s\n", conversation.Status)

		return nil
	},
}

var conversationsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would update conversation %s", args[0])) {
			return nil
		}

		var req api.ConversationUpdateRequest
		if err := readJSONBodyFlagsInto(cmd, &req); err != nil {
			return err
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		conversation, err := client.UpdateConversation(cmd.Context(), args[0], &req)
		if err != nil {
			return fmt.Errorf("failed to update conversation: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			return formatter.JSON(conversation)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Updated conversation %s (status: %s)\n", conversation.ID, conversation.Status)
		return nil
	},
}

var conversationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would delete conversation %s", args[0])) {
			return nil
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			_, _ = fmt.Fprintf(outWriter(cmd), "Are you sure you want to delete conversation %s? (use --yes to confirm)\n", args[0])
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := client.DeleteConversation(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete conversation: %w", err)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Deleted conversation %s\n", args[0])
		return nil
	},
}

var conversationsMessagesCmd = &cobra.Command{
	Use:   "messages <conversation-id>",
	Short: "List messages in a conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		resp, err := client.ListConversationMessages(cmd.Context(), args[0], page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list messages: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(resp)
		}

		headers := []string{"ID", "SENDER", "TYPE", "MESSAGE", "CREATED"}
		var rows [][]string
		for _, m := range resp.Items {
			body := m.Body
			if len(body) > 50 {
				body = body[:47] + "..."
			}
			rows = append(rows, []string{
				outfmt.FormatID("message", m.ID),
				m.SenderName,
				m.SenderType,
				body,
				m.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		formatter.Table(headers, rows)
		_, _ = fmt.Fprintf(outWriter(cmd), "\nShowing %d of %d messages\n", len(resp.Items), resp.TotalCount)
		return nil
	},
}

var conversationsSendCmd = &cobra.Command{
	Use:   "send <conversation-id>",
	Short: "Send a message to a conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, _ := cmd.Flags().GetString("body")

		if checkDryRun(cmd, fmt.Sprintf("[DRY-RUN] Would send message to conversation %s", args[0])) {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		req := &api.ConversationMessageCreateRequest{
			Body: body,
		}

		message, err := client.SendConversationMessage(cmd.Context(), args[0], req)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		formatter := getFormatter(cmd)
		outputFormat, _ := cmd.Flags().GetString("output")

		if outputFormat == "json" {
			return formatter.JSON(message)
		}

		_, _ = fmt.Fprintf(outWriter(cmd), "Sent message %s\n", message.ID)
		return nil
	},
}

var conversationsShopMessageCmd = &cobra.Command{
	Use:   "shop-message",
	Short: "Create shop message (documented endpoint; raw JSON body)",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := readJSONBodyFlags(cmd)
		if err != nil {
			return err
		}

		if checkDryRun(cmd, "[DRY-RUN] Would create shop message") {
			return nil
		}

		client, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := client.CreateConversationShopMessage(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create shop message: %w", err)
		}
		return getFormatter(cmd).JSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(conversationsCmd)

	conversationsCmd.AddCommand(conversationsListCmd)
	conversationsListCmd.Flags().Int("page", 1, "Page number")
	conversationsListCmd.Flags().Int("page-size", 20, "Results per page")
	conversationsListCmd.Flags().String("status", "", "Filter by status (open, closed, pending)")
	conversationsListCmd.Flags().String("channel", "", "Filter by channel (chat, email, messenger, whatsapp)")
	conversationsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	conversationsListCmd.Flags().String("assignee-id", "", "Filter by assignee ID")

	conversationsCmd.AddCommand(conversationsGetCmd)

	conversationsCmd.AddCommand(conversationsCreateCmd)
	conversationsCreateCmd.Flags().String("customer-id", "", "Customer ID (required)")
	conversationsCreateCmd.Flags().String("channel", "chat", "Channel (chat, email, messenger, whatsapp)")
	conversationsCreateCmd.Flags().String("subject", "", "Conversation subject")
	conversationsCreateCmd.Flags().String("message", "", "Initial message")
	_ = conversationsCreateCmd.MarkFlagRequired("customer-id")

	conversationsCmd.AddCommand(conversationsUpdateCmd)
	addJSONBodyFlags(conversationsUpdateCmd)

	conversationsCmd.AddCommand(conversationsDeleteCmd)
	conversationsDeleteCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	conversationsCmd.AddCommand(conversationsMessagesCmd)
	conversationsMessagesCmd.Flags().Int("page", 1, "Page number")
	conversationsMessagesCmd.Flags().Int("page-size", 20, "Results per page")

	conversationsCmd.AddCommand(conversationsSendCmd)
	conversationsSendCmd.Flags().String("body", "", "Message body (required)")
	_ = conversationsSendCmd.MarkFlagRequired("body")

	conversationsCmd.AddCommand(conversationsShopMessageCmd)
	addJSONBodyFlags(conversationsShopMessageCmd)
}

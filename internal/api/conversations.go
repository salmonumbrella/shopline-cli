package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Conversation represents a Shopline customer conversation/chat.
type Conversation struct {
	ID            string    `json:"id"`
	CustomerID    string    `json:"customer_id"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`  // open, closed, pending
	Channel       string    `json:"channel"` // chat, email, messenger, whatsapp
	AssigneeID    string    `json:"assignee_id"`
	AssigneeName  string    `json:"assignee_name"`
	MessageCount  int       `json:"message_count"`
	UnreadCount   int       `json:"unread_count"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ConversationMessage represents a message within a conversation.
type ConversationMessage struct {
	ID         string    `json:"id"`
	Body       string    `json:"body"`
	SenderID   string    `json:"sender_id"`
	SenderType string    `json:"sender_type"` // customer, staff
	SenderName string    `json:"sender_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// ConversationsListOptions contains options for listing conversations.
type ConversationsListOptions struct {
	Page       int
	PageSize   int
	Status     string
	Channel    string
	CustomerID string
	AssigneeID string
}

// ConversationsListResponse is the paginated response for conversations.
type ConversationsListResponse = ListResponse[Conversation]

// ConversationMessagesListResponse is the paginated response for conversation messages.
type ConversationMessagesListResponse = ListResponse[ConversationMessage]

// ConversationCreateRequest contains the data for creating a conversation.
type ConversationCreateRequest struct {
	CustomerID string `json:"customer_id"`
	Subject    string `json:"subject,omitempty"`
	Channel    string `json:"channel"`
	Message    string `json:"message,omitempty"`
}

// ConversationUpdateRequest contains the data for updating a conversation.
type ConversationUpdateRequest struct {
	Status     string `json:"status,omitempty"`
	AssigneeID string `json:"assignee_id,omitempty"`
	Subject    string `json:"subject,omitempty"`
}

// ConversationMessageCreateRequest contains the data for sending a message.
type ConversationMessageCreateRequest struct {
	Body string `json:"body"`
}

// ListConversations retrieves a list of conversations.
func (c *Client) ListConversations(ctx context.Context, opts *ConversationsListOptions) (*ConversationsListResponse, error) {
	path := "/conversations"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			String("channel", opts.Channel).
			String("customer_id", opts.CustomerID).
			String("assignee_id", opts.AssigneeID).
			Build()
	}

	var resp ConversationsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetConversation retrieves a single conversation by ID.
func (c *Client) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("conversation id is required")
	}
	var conversation Conversation
	if err := c.Get(ctx, fmt.Sprintf("/conversations/%s", id), &conversation); err != nil {
		return nil, err
	}
	return &conversation, nil
}

// CreateConversation creates a new conversation.
func (c *Client) CreateConversation(ctx context.Context, req *ConversationCreateRequest) (*Conversation, error) {
	var conversation Conversation
	if err := c.Post(ctx, "/conversations", req, &conversation); err != nil {
		return nil, err
	}
	return &conversation, nil
}

// UpdateConversation updates an existing conversation.
func (c *Client) UpdateConversation(ctx context.Context, id string, req *ConversationUpdateRequest) (*Conversation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("conversation id is required")
	}
	var conversation Conversation
	if err := c.Put(ctx, fmt.Sprintf("/conversations/%s", id), req, &conversation); err != nil {
		return nil, err
	}
	return &conversation, nil
}

// DeleteConversation deletes a conversation.
func (c *Client) DeleteConversation(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("conversation id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/conversations/%s", id))
}

// ListConversationMessages retrieves messages for a conversation.
func (c *Client) ListConversationMessages(ctx context.Context, conversationID string, page, pageSize int) (*ConversationMessagesListResponse, error) {
	if strings.TrimSpace(conversationID) == "" {
		return nil, fmt.Errorf("conversation id is required")
	}
	path := fmt.Sprintf("/conversations/%s/messages", conversationID)
	path += NewQuery().
		Int("page", page).
		Int("page_size", pageSize).
		Build()

	var resp ConversationMessagesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendConversationMessage sends a message to a conversation.
func (c *Client) SendConversationMessage(ctx context.Context, conversationID string, req *ConversationMessageCreateRequest) (*ConversationMessage, error) {
	if strings.TrimSpace(conversationID) == "" {
		return nil, fmt.Errorf("conversation id is required")
	}
	var message ConversationMessage
	if err := c.Post(ctx, fmt.Sprintf("/conversations/%s/messages", conversationID), req, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

// OrderMessageCreateRequest contains the request body for creating an order message.
type OrderMessageCreateRequest struct {
	OrderID string `json:"order_id"`
	Content string `json:"content"`
}

// ShopMessageCreateRequest contains the request body for creating a shop message.
type ShopMessageCreateRequest struct {
	CustomerID string `json:"customer_id"`
	Content    string `json:"content"`
}

// CreateOrderMessage creates a message for an order conversation.
func (c *Client) CreateOrderMessage(ctx context.Context, orderID, content string) error {
	if strings.TrimSpace(orderID) == "" {
		return fmt.Errorf("order id is required")
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is required")
	}
	req := &OrderMessageCreateRequest{OrderID: orderID, Content: content}
	return c.Post(ctx, "/conversations/order-message", req, nil)
}

// CreateShopMessage creates a shop message to a customer.
func (c *Client) CreateShopMessage(ctx context.Context, customerID, content string) error {
	if strings.TrimSpace(customerID) == "" {
		return fmt.Errorf("customer id is required")
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is required")
	}
	req := &ShopMessageCreateRequest{CustomerID: customerID, Content: content}
	return c.Post(ctx, "/conversations/shop-message", req, nil)
}

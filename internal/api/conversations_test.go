package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConversationsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/conversations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ConversationsListResponse{
			Items: []Conversation{
				{ID: "conv_123", CustomerName: "John Doe", Status: "open", Channel: "chat", MessageCount: 5},
				{ID: "conv_456", CustomerName: "Jane Smith", Status: "closed", Channel: "email", MessageCount: 10},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	conversations, err := client.ListConversations(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListConversations failed: %v", err)
	}

	if len(conversations.Items) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(conversations.Items))
	}
	if conversations.Items[0].ID != "conv_123" {
		t.Errorf("Unexpected conversation ID: %s", conversations.Items[0].ID)
	}
}

func TestConversationsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "open" {
			t.Errorf("Expected status=open, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("channel") != "chat" {
			t.Errorf("Expected channel=chat, got %s", r.URL.Query().Get("channel"))
		}

		resp := ConversationsListResponse{
			Items:      []Conversation{},
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &ConversationsListOptions{
		Status:  "open",
		Channel: "chat",
	}
	_, err := client.ListConversations(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListConversations failed: %v", err)
	}
}

func TestConversationsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/conversations/conv_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		conversation := Conversation{ID: "conv_123", CustomerName: "John Doe", Status: "open", Channel: "chat"}
		_ = json.NewEncoder(w).Encode(conversation)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	conversation, err := client.GetConversation(context.Background(), "conv_123")
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}

	if conversation.ID != "conv_123" {
		t.Errorf("Unexpected conversation ID: %s", conversation.ID)
	}
}

func TestGetConversationEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetConversation(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "conversation id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestConversationsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/conversations" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ConversationCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerID != "cust_123" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}

		conversation := Conversation{
			ID:         "conv_new",
			CustomerID: req.CustomerID,
			Channel:    req.Channel,
			Status:     "open",
		}
		_ = json.NewEncoder(w).Encode(conversation)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ConversationCreateRequest{
		CustomerID: "cust_123",
		Channel:    "chat",
		Subject:    "Help needed",
	}
	conversation, err := client.CreateConversation(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateConversation failed: %v", err)
	}

	if conversation.ID != "conv_new" {
		t.Errorf("Unexpected conversation ID: %s", conversation.ID)
	}
	if conversation.CustomerID != "cust_123" {
		t.Errorf("Unexpected customer ID: %s", conversation.CustomerID)
	}
}

func TestConversationsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/conv_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ConversationUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		conversation := Conversation{
			ID:         "conv_123",
			Status:     req.Status,
			AssigneeID: req.AssigneeID,
		}
		_ = json.NewEncoder(w).Encode(conversation)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ConversationUpdateRequest{
		Status:     "closed",
		AssigneeID: "staff_123",
	}
	conversation, err := client.UpdateConversation(context.Background(), "conv_123", req)
	if err != nil {
		t.Fatalf("UpdateConversation failed: %v", err)
	}

	if conversation.Status != "closed" {
		t.Errorf("Unexpected status: %s", conversation.Status)
	}
}

func TestConversationsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/conv_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteConversation(context.Background(), "conv_123")
	if err != nil {
		t.Fatalf("DeleteConversation failed: %v", err)
	}
}

func TestUpdateConversationEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.UpdateConversation(context.Background(), tc.id, &ConversationUpdateRequest{Status: "closed"})
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "conversation id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteConversationEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteConversation(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "conversation id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestListConversationMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/conv_123/messages" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ConversationMessagesListResponse{
			Items: []ConversationMessage{
				{ID: "msg_1", Body: "Hello", SenderType: "customer", SenderName: "John"},
				{ID: "msg_2", Body: "Hi there!", SenderType: "staff", SenderName: "Support"},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	messages, err := client.ListConversationMessages(context.Background(), "conv_123", 1, 20)
	if err != nil {
		t.Fatalf("ListConversationMessages failed: %v", err)
	}

	if len(messages.Items) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages.Items))
	}
}

func TestListConversationMessagesEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListConversationMessages(context.Background(), "", 1, 20)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "conversation id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestSendConversationMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/conv_123/messages" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ConversationMessageCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		message := ConversationMessage{
			ID:         "msg_new",
			Body:       req.Body,
			SenderType: "staff",
			SenderName: "Support",
		}
		_ = json.NewEncoder(w).Encode(message)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ConversationMessageCreateRequest{
		Body: "Thank you for contacting us!",
	}
	message, err := client.SendConversationMessage(context.Background(), "conv_123", req)
	if err != nil {
		t.Fatalf("SendConversationMessage failed: %v", err)
	}

	if message.ID != "msg_new" {
		t.Errorf("Unexpected message ID: %s", message.ID)
	}
}

func TestSendConversationMessageEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.SendConversationMessage(context.Background(), "", &ConversationMessageCreateRequest{Body: "Test"})
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "conversation id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestCreateOrderMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/order-message" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req OrderMessageCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.OrderID != "order_123" {
			t.Errorf("Unexpected order ID: %s", req.OrderID)
		}
		if req.Content != "Your order has been shipped!" {
			t.Errorf("Unexpected content: %s", req.Content)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CreateOrderMessage(context.Background(), "order_123", "Your order has been shipped!")
	if err != nil {
		t.Fatalf("CreateOrderMessage failed: %v", err)
	}
}

func TestCreateOrderMessageEmptyOrderID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name    string
		orderID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.CreateOrderMessage(context.Background(), tc.orderID, "Test message")
			if err == nil {
				t.Error("Expected error for empty order ID, got nil")
			}
			if err != nil && err.Error() != "order id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateOrderMessageEmptyContent(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name    string
		content string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.CreateOrderMessage(context.Background(), "order_123", tc.content)
			if err == nil {
				t.Error("Expected error for empty content, got nil")
			}
			if err != nil && err.Error() != "content is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateOrderMessageAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "order not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CreateOrderMessage(context.Background(), "order_invalid", "Test message")
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

func TestCreateShopMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/conversations/shop-message" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req ShopMessageCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerID != "cust_456" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}
		if req.Content != "Thank you for shopping with us!" {
			t.Errorf("Unexpected content: %s", req.Content)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CreateShopMessage(context.Background(), "cust_456", "Thank you for shopping with us!")
	if err != nil {
		t.Fatalf("CreateShopMessage failed: %v", err)
	}
}

func TestCreateShopMessageEmptyCustomerID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name       string
		customerID string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.CreateShopMessage(context.Background(), tc.customerID, "Test message")
			if err == nil {
				t.Error("Expected error for empty customer ID, got nil")
			}
			if err != nil && err.Error() != "customer id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateShopMessageEmptyContent(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name    string
		content string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.CreateShopMessage(context.Background(), "cust_123", tc.content)
			if err == nil {
				t.Error("Expected error for empty content, got nil")
			}
			if err != nil && err.Error() != "content is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestCreateShopMessageAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "customer not found"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.CreateShopMessage(context.Background(), "cust_invalid", "Test message")
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
}

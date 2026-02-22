package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestConversationsCommandSetup(t *testing.T) {
	if conversationsCmd.Use != "conversations" {
		t.Errorf("expected Use 'conversations', got %q", conversationsCmd.Use)
	}
	if conversationsCmd.Short != "Manage customer conversations/chat" {
		t.Errorf("expected Short 'Manage customer conversations/chat', got %q", conversationsCmd.Short)
	}
}

func TestConversationsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":         "List conversations",
		"get":          "Get conversation details",
		"create":       "Create a conversation",
		"delete":       "Delete a conversation",
		"messages":     "List messages in a conversation",
		"send":         "Send a message to a conversation",
		"shop-message": "Create shop message (documented endpoint; raw JSON body)",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range conversationsCmd.Commands() {
				if sub.Use == name || (len(sub.Use) > len(name) && sub.Use[:len(name)] == name) {
					found = true
					if sub.Short != short {
						t.Errorf("expected Short %q, got %q", short, sub.Short)
					}
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", name)
			}
		})
	}
}

func TestConversationsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

func TestConversationsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
		{"channel", ""},
		{"customer-id", ""},
		{"assignee-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := conversationsListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestConversationsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"customer-id", ""},
		{"channel", "chat"},
		{"subject", ""},
		{"message", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := conversationsCreateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestConversationsMessagesFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := conversationsMessagesCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestConversationsSendFlags(t *testing.T) {
	flag := conversationsSendCmd.Flags().Lookup("body")
	if flag == nil {
		t.Error("flag 'body' not found")
		return
	}
	if flag.DefValue != "" {
		t.Errorf("expected default '', got %q", flag.DefValue)
	}
}

func TestConversationsWithMockStore(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	store := &mockStore{
		names: []string{"teststore"},
		creds: map[string]*secrets.StoreCredentials{
			"teststore": {Handle: "test-handle", AccessToken: "test-token"},
		},
	}

	secretsStoreFactory = func() (CredentialStore, error) {
		return store, nil
	}

	cmd := newTestCmdWithFlags()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// conversationsMockAPIClient is a mock implementation of api.APIClient for conversations tests.
type conversationsMockAPIClient struct {
	api.MockClient
	listConversationsResp  *api.ConversationsListResponse
	listConversationsErr   error
	getConversationResp    *api.Conversation
	getConversationErr     error
	createConversationResp *api.Conversation
	createConversationErr  error
	deleteConversationErr  error
	listMessagesResp       *api.ConversationMessagesListResponse
	listMessagesErr        error
	sendMessageResp        *api.ConversationMessage
	sendMessageErr         error
	shopMessageResp        json.RawMessage
	shopMessageErr         error
}

func (m *conversationsMockAPIClient) ListConversations(ctx context.Context, opts *api.ConversationsListOptions) (*api.ConversationsListResponse, error) {
	return m.listConversationsResp, m.listConversationsErr
}

func (m *conversationsMockAPIClient) GetConversation(ctx context.Context, id string) (*api.Conversation, error) {
	return m.getConversationResp, m.getConversationErr
}

func (m *conversationsMockAPIClient) CreateConversation(ctx context.Context, req *api.ConversationCreateRequest) (*api.Conversation, error) {
	return m.createConversationResp, m.createConversationErr
}

func (m *conversationsMockAPIClient) DeleteConversation(ctx context.Context, id string) error {
	return m.deleteConversationErr
}

func (m *conversationsMockAPIClient) ListConversationMessages(ctx context.Context, conversationID string, page, pageSize int) (*api.ConversationMessagesListResponse, error) {
	return m.listMessagesResp, m.listMessagesErr
}

func (m *conversationsMockAPIClient) SendConversationMessage(ctx context.Context, conversationID string, req *api.ConversationMessageCreateRequest) (*api.ConversationMessage, error) {
	return m.sendMessageResp, m.sendMessageErr
}

func (m *conversationsMockAPIClient) CreateConversationShopMessage(ctx context.Context, body any) (json.RawMessage, error) {
	return m.shopMessageResp, m.shopMessageErr
}

// setupConversationsMockFactories sets up mock factories for conversations tests.
func setupConversationsMockFactories(mockClient *conversationsMockAPIClient) (func(), *bytes.Buffer) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	buf := new(bytes.Buffer)
	formatterWriter = buf

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cleanup := func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return cleanup, buf
}

// newConversationsTestCmd creates a test command with common flags for conversations tests.
func newConversationsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

// TestConversationsListRunE tests the conversations list command with mock API.
func TestConversationsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ConversationsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ConversationsListResponse{
				Items: []api.Conversation{
					{
						ID:            "conv_123",
						CustomerName:  "Alice Smith",
						Status:        "open",
						Channel:       "chat",
						MessageCount:  5,
						LastMessageAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "conv_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ConversationsListResponse{
				Items:      []api.Conversation{},
				TotalCount: 0,
			},
		},
		{
			name: "conversation with no last message",
			mockResp: &api.ConversationsListResponse{
				Items: []api.Conversation{
					{
						ID:           "conv_456",
						CustomerName: "Bob Jones",
						Status:       "pending",
						Channel:      "email",
						MessageCount: 0,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "conv_456",
		},
		{
			name: "multiple conversations",
			mockResp: &api.ConversationsListResponse{
				Items: []api.Conversation{
					{
						ID:           "conv_001",
						CustomerName: "Customer One",
						Status:       "open",
						Channel:      "chat",
						MessageCount: 3,
					},
					{
						ID:           "conv_002",
						CustomerName: "Customer Two",
						Status:       "closed",
						Channel:      "whatsapp",
						MessageCount: 10,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "conv_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &conversationsMockAPIClient{
				listConversationsResp: tt.mockResp,
				listConversationsErr:  tt.mockErr,
			}
			cleanup, buf := setupConversationsMockFactories(mockClient)
			defer cleanup()

			cmd := newConversationsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("channel", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("assignee-id", "", "")

			err := conversationsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestConversationsListRunEWithJSON tests JSON output format for list command.
func TestConversationsListRunEWithJSON(t *testing.T) {
	mockClient := &conversationsMockAPIClient{
		listConversationsResp: &api.ConversationsListResponse{
			Items: []api.Conversation{
				{ID: "conv_json", CustomerName: "JSON Customer", Status: "open", Channel: "chat"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("channel", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("assignee-id", "", "")

	err := conversationsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "conv_json") {
		t.Errorf("JSON output should contain conversation ID, got: %s", output)
	}
}

// TestConversationsGetRunE tests the conversations get command with mock API.
func TestConversationsGetRunE(t *testing.T) {
	tests := []struct {
		name           string
		conversationID string
		mockResp       *api.Conversation
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful get",
			conversationID: "conv_123",
			mockResp: &api.Conversation{
				ID:            "conv_123",
				CustomerID:    "cust_456",
				CustomerName:  "Alice Smith",
				CustomerEmail: "alice@example.com",
				Subject:       "Order inquiry",
				Status:        "open",
				Channel:       "chat",
				AssigneeID:    "staff_789",
				AssigneeName:  "Support Agent",
				MessageCount:  5,
				UnreadCount:   2,
				LastMessageAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				CreatedAt:     time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:           "conversation not found",
			conversationID: "conv_999",
			mockErr:        errors.New("conversation not found"),
			wantErr:        true,
		},
		{
			name:           "minimal conversation",
			conversationID: "conv_min",
			mockResp: &api.Conversation{
				ID:           "conv_min",
				CustomerName: "Minimal Customer",
				Status:       "pending",
				Channel:      "email",
				MessageCount: 0,
				CreatedAt:    time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "conversation with no email",
			conversationID: "conv_noemail",
			mockResp: &api.Conversation{
				ID:           "conv_noemail",
				CustomerName: "No Email Customer",
				Status:       "open",
				Channel:      "messenger",
				MessageCount: 3,
				CreatedAt:    time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "conversation with no assignee",
			conversationID: "conv_noassign",
			mockResp: &api.Conversation{
				ID:           "conv_noassign",
				CustomerName: "Unassigned Customer",
				Status:       "open",
				Channel:      "whatsapp",
				MessageCount: 1,
				CreatedAt:    time.Date(2024, 4, 1, 8, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 4, 1, 8, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "conversation with unread messages",
			conversationID: "conv_unread",
			mockResp: &api.Conversation{
				ID:           "conv_unread",
				CustomerName: "Customer with Unread",
				Status:       "open",
				Channel:      "chat",
				MessageCount: 10,
				UnreadCount:  5,
				CreatedAt:    time.Date(2024, 5, 1, 7, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 5, 1, 7, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &conversationsMockAPIClient{
				getConversationResp: tt.mockResp,
				getConversationErr:  tt.mockErr,
			}
			cleanup, _ := setupConversationsMockFactories(mockClient)
			defer cleanup()

			cmd := newConversationsTestCmd()

			err := conversationsGetCmd.RunE(cmd, []string{tt.conversationID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConversationsGetRunEWithJSON tests JSON output format for get command.
func TestConversationsGetRunEWithJSON(t *testing.T) {
	mockClient := &conversationsMockAPIClient{
		getConversationResp: &api.Conversation{
			ID:           "conv_json",
			CustomerName: "JSON Test Customer",
			Status:       "open",
			Channel:      "chat",
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := conversationsGetCmd.RunE(cmd, []string{"conv_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "conv_json") {
		t.Errorf("JSON output should contain conversation ID, got: %s", output)
	}
}

// TestConversationsGetArgs verifies get command requires exactly 1 argument.
func TestConversationsGetArgs(t *testing.T) {
	err := conversationsGetCmd.Args(conversationsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = conversationsGetCmd.Args(conversationsGetCmd, []string{"conv-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestConversationsCreateRunE tests the conversations create command with mock API.
func TestConversationsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.Conversation
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Conversation{
				ID:      "conv_new",
				Channel: "chat",
				Status:  "open",
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create conversation"),
			wantErr: true,
		},
		{
			name: "create with email channel",
			mockResp: &api.Conversation{
				ID:      "conv_email",
				Channel: "email",
				Status:  "open",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &conversationsMockAPIClient{
				createConversationResp: tt.mockResp,
				createConversationErr:  tt.mockErr,
			}
			cleanup, _ := setupConversationsMockFactories(mockClient)
			defer cleanup()

			cmd := newConversationsTestCmd()
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("channel", "chat", "")
			cmd.Flags().String("subject", "Test Subject", "")
			cmd.Flags().String("message", "Test Message", "")

			err := conversationsCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConversationsCreateRunEWithJSON tests JSON output format for create command.
func TestConversationsCreateRunEWithJSON(t *testing.T) {
	mockClient := &conversationsMockAPIClient{
		createConversationResp: &api.Conversation{
			ID:      "conv_json_create",
			Channel: "chat",
			Status:  "open",
		},
	}
	cleanup, buf := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("channel", "chat", "")
	cmd.Flags().String("subject", "", "")
	cmd.Flags().String("message", "", "")

	err := conversationsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "conv_json_create") {
		t.Errorf("JSON output should contain conversation ID, got: %s", output)
	}
}

// TestConversationsCreateDryRun tests the dry-run flag for create command.
func TestConversationsCreateDryRun(t *testing.T) {
	mockClient := &conversationsMockAPIClient{}
	cleanup, _ := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("channel", "chat", "")
	cmd.Flags().String("subject", "", "")
	cmd.Flags().String("message", "", "")

	err := conversationsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestConversationsDeleteRunE tests the conversations delete command with mock API.
func TestConversationsDeleteRunE(t *testing.T) {
	tests := []struct {
		name           string
		conversationID string
		mockErr        error
		wantErr        bool
		yes            bool
	}{
		{
			name:           "successful delete with confirmation",
			conversationID: "conv_123",
			yes:            true,
		},
		{
			name:           "API error",
			conversationID: "conv_err",
			mockErr:        errors.New("failed to delete"),
			wantErr:        true,
			yes:            true,
		},
		{
			name:           "delete without confirmation",
			conversationID: "conv_noconfirm",
			yes:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &conversationsMockAPIClient{
				deleteConversationErr: tt.mockErr,
			}
			cleanup, _ := setupConversationsMockFactories(mockClient)
			defer cleanup()

			cmd := newConversationsTestCmd()
			if tt.yes {
				_ = cmd.Flags().Set("yes", "true")
			} else {
				_ = cmd.Flags().Set("yes", "false")
			}

			err := conversationsDeleteCmd.RunE(cmd, []string{tt.conversationID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConversationsDeleteDryRun tests the dry-run flag for delete command.
func TestConversationsDeleteDryRun(t *testing.T) {
	mockClient := &conversationsMockAPIClient{}
	cleanup, _ := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := conversationsDeleteCmd.RunE(cmd, []string{"conv_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestConversationsDeleteArgs verifies delete command requires exactly 1 argument.
func TestConversationsDeleteArgs(t *testing.T) {
	err := conversationsDeleteCmd.Args(conversationsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = conversationsDeleteCmd.Args(conversationsDeleteCmd, []string{"conv-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestConversationsMessagesRunE tests the conversations messages command with mock API.
func TestConversationsMessagesRunE(t *testing.T) {
	tests := []struct {
		name           string
		conversationID string
		mockResp       *api.ConversationMessagesListResponse
		mockErr        error
		wantErr        bool
		wantOutput     string
	}{
		{
			name:           "successful list messages",
			conversationID: "conv_123",
			mockResp: &api.ConversationMessagesListResponse{
				Items: []api.ConversationMessage{
					{
						ID:         "msg_001",
						Body:       "Hello, I have a question",
						SenderID:   "cust_456",
						SenderType: "customer",
						SenderName: "Alice Smith",
						CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:         "msg_002",
						Body:       "Hi Alice, how can I help?",
						SenderID:   "staff_789",
						SenderType: "staff",
						SenderName: "Support Agent",
						CreatedAt:  time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "msg_001",
		},
		{
			name:           "API error",
			conversationID: "conv_err",
			mockErr:        errors.New("failed to list messages"),
			wantErr:        true,
		},
		{
			name:           "empty messages",
			conversationID: "conv_empty",
			mockResp: &api.ConversationMessagesListResponse{
				Items:      []api.ConversationMessage{},
				TotalCount: 0,
			},
		},
		{
			name:           "message with long body truncated",
			conversationID: "conv_long",
			mockResp: &api.ConversationMessagesListResponse{
				Items: []api.ConversationMessage{
					{
						ID:         "msg_long",
						Body:       "This is a very long message body that should be truncated when displayed in the table output format for better readability",
						SenderID:   "cust_123",
						SenderType: "customer",
						SenderName: "Long Message Customer",
						CreatedAt:  time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "msg_long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &conversationsMockAPIClient{
				listMessagesResp: tt.mockResp,
				listMessagesErr:  tt.mockErr,
			}
			cleanup, buf := setupConversationsMockFactories(mockClient)
			defer cleanup()

			cmd := newConversationsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := conversationsMessagesCmd.RunE(cmd, []string{tt.conversationID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			output := buf.String()
			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestConversationsMessagesRunEWithJSON tests JSON output format for messages command.
func TestConversationsMessagesRunEWithJSON(t *testing.T) {
	mockClient := &conversationsMockAPIClient{
		listMessagesResp: &api.ConversationMessagesListResponse{
			Items: []api.ConversationMessage{
				{
					ID:         "msg_json",
					Body:       "JSON test message",
					SenderType: "customer",
					SenderName: "JSON Customer",
					CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := conversationsMessagesCmd.RunE(cmd, []string{"conv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "msg_json") {
		t.Errorf("JSON output should contain message ID, got: %s", output)
	}
}

// TestConversationsMessagesArgs verifies messages command requires exactly 1 argument.
func TestConversationsMessagesArgs(t *testing.T) {
	err := conversationsMessagesCmd.Args(conversationsMessagesCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = conversationsMessagesCmd.Args(conversationsMessagesCmd, []string{"conv-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestConversationsSendRunE tests the conversations send command with mock API.
func TestConversationsSendRunE(t *testing.T) {
	tests := []struct {
		name           string
		conversationID string
		mockResp       *api.ConversationMessage
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful send",
			conversationID: "conv_123",
			mockResp: &api.ConversationMessage{
				ID:         "msg_new",
				Body:       "Thank you for your inquiry",
				SenderID:   "staff_123",
				SenderType: "staff",
				SenderName: "Support Agent",
				CreatedAt:  time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "API error",
			conversationID: "conv_err",
			mockErr:        errors.New("failed to send message"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &conversationsMockAPIClient{
				sendMessageResp: tt.mockResp,
				sendMessageErr:  tt.mockErr,
			}
			cleanup, _ := setupConversationsMockFactories(mockClient)
			defer cleanup()

			cmd := newConversationsTestCmd()
			cmd.Flags().String("body", "Test message body", "")

			err := conversationsSendCmd.RunE(cmd, []string{tt.conversationID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConversationsSendRunEWithJSON tests JSON output format for send command.
func TestConversationsSendRunEWithJSON(t *testing.T) {
	mockClient := &conversationsMockAPIClient{
		sendMessageResp: &api.ConversationMessage{
			ID:         "msg_json_send",
			Body:       "JSON test message",
			SenderType: "staff",
			SenderName: "JSON Staff",
			CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("body", "Test message", "")

	err := conversationsSendCmd.RunE(cmd, []string{"conv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "msg_json_send") {
		t.Errorf("JSON output should contain message ID, got: %s", output)
	}
}

// TestConversationsSendDryRun tests the dry-run flag for send command.
func TestConversationsSendDryRun(t *testing.T) {
	mockClient := &conversationsMockAPIClient{}
	cleanup, _ := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("body", "Test message", "")

	err := conversationsSendCmd.RunE(cmd, []string{"conv_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run: %v", err)
	}
}

// TestConversationsSendArgs verifies send command requires exactly 1 argument.
func TestConversationsSendArgs(t *testing.T) {
	err := conversationsSendCmd.Args(conversationsSendCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = conversationsSendCmd.Args(conversationsSendCmd, []string{"conv-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestConversationsGetClientError tests error handling when getClient fails in list command.
func TestConversationsListGetClientErrorRunE(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newConversationsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("channel", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("assignee-id", "", "")

	err := conversationsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestConversationsGetGetClientErrorRunE tests error handling when getClient fails in get command.
func TestConversationsGetGetClientErrorRunE(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newConversationsTestCmd()

	err := conversationsGetCmd.RunE(cmd, []string{"conv_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestConversationsMessagesGetClientErrorRunE tests error handling when getClient fails in messages command.
func TestConversationsMessagesGetClientErrorRunE(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newConversationsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := conversationsMessagesCmd.RunE(cmd, []string{"conv_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestConversationsSendGetClientErrorRunE tests error handling when getClient fails in send command.
func TestConversationsSendGetClientErrorRunE(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newConversationsTestCmd()
	cmd.Flags().String("body", "test", "")

	err := conversationsSendCmd.RunE(cmd, []string{"conv_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestConversationsCreateGetClientErrorRunE tests error handling when getClient fails in create command.
func TestConversationsCreateGetClientErrorRunE(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newConversationsTestCmd()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("channel", "chat", "")
	cmd.Flags().String("subject", "", "")
	cmd.Flags().String("message", "", "")

	err := conversationsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestConversationsDeleteGetClientErrorRunE tests error handling when getClient fails in delete command.
func TestConversationsDeleteGetClientErrorRunE(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newConversationsTestCmd()
	_ = cmd.Flags().Set("yes", "true")

	err := conversationsDeleteCmd.RunE(cmd, []string{"conv_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestConversationsListNoProfiles verifies list command error handling when no profiles exist.
func TestConversationsListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newConversationsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("channel", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("assignee-id", "", "")

	err := conversationsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestConversationsGetMultipleProfiles verifies get command error handling when multiple profiles exist.
func TestConversationsGetMultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newConversationsTestCmd()
	err := conversationsGetCmd.RunE(cmd, []string{"conv-id"})
	if err == nil {
		t.Error("expected error for multiple profiles")
	}
}

func TestConversationsShopMessageRunE(t *testing.T) {
	mockClient := &conversationsMockAPIClient{
		shopMessageResp: json.RawMessage(`{"id":"msg_1"}`),
	}
	cleanup, buf := setupConversationsMockFactories(mockClient)
	defer cleanup()

	cmd := newConversationsTestCmd()
	addJSONBodyFlags(cmd)
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"ok":true}`)

	if err := conversationsShopMessageCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"msg_1\"") {
		t.Fatalf("expected msg_1 in output, got %q", buf.String())
	}
}

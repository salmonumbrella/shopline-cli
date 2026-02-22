package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// TestWebhooksCommandSetup verifies webhooks command initialization
func TestWebhooksCommandSetup(t *testing.T) {
	if webhooksCmd.Use != "webhooks" {
		t.Errorf("expected Use 'webhooks', got %q", webhooksCmd.Use)
	}
	if webhooksCmd.Short != "Manage webhooks" {
		t.Errorf("expected Short 'Manage webhooks', got %q", webhooksCmd.Short)
	}
}

// TestWebhooksSubcommands verifies all subcommands are registered
func TestWebhooksSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List webhooks",
		"get":    "Get webhook details",
		"create": "Create a webhook",
		"update": "Update a webhook",
		"delete": "Delete a webhook",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range webhooksCmd.Commands() {
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

// TestWebhooksListFlags verifies list command flags exist
func TestWebhooksListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"topic", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := webhooksListCmd.Flags().Lookup(f.name)
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

// TestWebhooksCreateFlags verifies create command flags exist
func TestWebhooksCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"topic", ""},
		{"address", ""},
		{"format", "json"},
		{"api-version", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := webhooksCreateCmd.Flags().Lookup(f.name)
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

// TestWebhooksCreateRequiredFlags verifies required flags
func TestWebhooksCreateRequiredFlags(t *testing.T) {
	// Topic and address should be required
	topicFlag := webhooksCreateCmd.Flags().Lookup("topic")
	addressFlag := webhooksCreateCmd.Flags().Lookup("address")

	if topicFlag == nil {
		t.Error("topic flag not found")
	}
	if addressFlag == nil {
		t.Error("address flag not found")
	}
}

func TestWebhooksUpdateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"topic", ""},
		{"address", ""},
		{"format", ""},
		{"api-version", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := webhooksUpdateCmd.Flags().Lookup(f.name)
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

// webhooksMockAPIClient is a mock implementation of api.APIClient for webhooks tests.
type webhooksMockAPIClient struct {
	api.MockClient
	listWebhooksResp   *api.WebhooksListResponse
	listWebhooksErr    error
	getWebhookResp     *api.Webhook
	getWebhookErr      error
	createWebhookResp  *api.Webhook
	createWebhookErr   error
	deleteWebhookErr   error
	deleteWebhookID    string
	createWebhookCalls int
}

func (m *webhooksMockAPIClient) ListWebhooks(ctx context.Context, opts *api.WebhooksListOptions) (*api.WebhooksListResponse, error) {
	return m.listWebhooksResp, m.listWebhooksErr
}

func (m *webhooksMockAPIClient) GetWebhook(ctx context.Context, id string) (*api.Webhook, error) {
	return m.getWebhookResp, m.getWebhookErr
}

func (m *webhooksMockAPIClient) CreateWebhook(ctx context.Context, req *api.WebhookCreateRequest) (*api.Webhook, error) {
	m.createWebhookCalls++
	return m.createWebhookResp, m.createWebhookErr
}

func (m *webhooksMockAPIClient) DeleteWebhook(ctx context.Context, id string) error {
	m.deleteWebhookID = id
	return m.deleteWebhookErr
}

// setupWebhooksMockFactories sets up mock factories for webhooks tests.
func setupWebhooksMockFactories(mockClient *webhooksMockAPIClient) (func(), *bytes.Buffer) {
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

// newWebhooksTestCmd creates a test command with common flags for webhooks tests.
func newWebhooksTestCmd() *cobra.Command {
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

// TestWebhooksListRunE tests the webhooks list command with mock API.
func TestWebhooksListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.WebhooksListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.WebhooksListResponse{
				Items: []api.Webhook{
					{
						ID:         "wh_123",
						Topic:      "orders/create",
						Address:    "https://example.com/webhook",
						Format:     api.WebhookFormatJSON,
						APIVersion: "2024-01",
						CreatedAt:  testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "wh_123",
		},
		{
			name: "multiple webhooks",
			mockResp: &api.WebhooksListResponse{
				Items: []api.Webhook{
					{
						ID:         "wh_123",
						Topic:      "orders/create",
						Address:    "https://example.com/orders",
						Format:     api.WebhookFormatJSON,
						APIVersion: "2024-01",
						CreatedAt:  testTime,
					},
					{
						ID:         "wh_456",
						Topic:      "products/update",
						Address:    "https://example.com/products",
						Format:     api.WebhookFormatXML,
						APIVersion: "2024-01",
						CreatedAt:  testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "wh_456",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.WebhooksListResponse{
				Items:      []api.Webhook{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &webhooksMockAPIClient{
				listWebhooksResp: tt.mockResp,
				listWebhooksErr:  tt.mockErr,
			}
			cleanup, buf := setupWebhooksMockFactories(mockClient)
			defer cleanup()

			cmd := newWebhooksTestCmd()
			cmd.Flags().String("topic", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := webhooksListCmd.RunE(cmd, []string{})

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

// TestWebhooksListJSONOutput tests the webhooks list command with JSON output.
func TestWebhooksListJSONOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &webhooksMockAPIClient{
		listWebhooksResp: &api.WebhooksListResponse{
			Items: []api.Webhook{
				{
					ID:         "wh_123",
					Topic:      "orders/create",
					Address:    "https://example.com/webhook",
					Format:     api.WebhookFormatJSON,
					APIVersion: "2024-01",
					CreatedAt:  testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("topic", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := webhooksListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain the webhook ID
	if !strings.Contains(output, "wh_123") {
		t.Errorf("JSON output should contain webhook ID, got %q", output)
	}
}

// TestWebhooksGetRunE tests the webhooks get command with mock API.
func TestWebhooksGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		webhookID string
		mockResp  *api.Webhook
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			webhookID: "wh_123",
			mockResp: &api.Webhook{
				ID:         "wh_123",
				Topic:      "orders/create",
				Address:    "https://example.com/webhook",
				Format:     api.WebhookFormatJSON,
				APIVersion: "2024-01",
				CreatedAt:  testTime,
				UpdatedAt:  testTime,
			},
		},
		{
			name:      "webhook not found",
			webhookID: "wh_999",
			mockErr:   errors.New("webhook not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &webhooksMockAPIClient{
				getWebhookResp: tt.mockResp,
				getWebhookErr:  tt.mockErr,
			}
			cleanup, _ := setupWebhooksMockFactories(mockClient)
			defer cleanup()

			cmd := newWebhooksTestCmd()

			err := webhooksGetCmd.RunE(cmd, []string{tt.webhookID})

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

// TestWebhooksGetJSONOutput tests the webhooks get command with JSON output.
func TestWebhooksGetJSONOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &webhooksMockAPIClient{
		getWebhookResp: &api.Webhook{
			ID:         "wh_123",
			Topic:      "orders/create",
			Address:    "https://example.com/webhook",
			Format:     api.WebhookFormatJSON,
			APIVersion: "2024-01",
			CreatedAt:  testTime,
			UpdatedAt:  testTime,
		},
	}
	cleanup, buf := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := webhooksGetCmd.RunE(cmd, []string{"wh_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "wh_123") {
		t.Errorf("JSON output should contain webhook ID, got %q", output)
	}
}

// TestWebhooksCreateRunE tests the webhooks create command with mock API.
func TestWebhooksCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		topic    string
		address  string
		format   string
		mockResp *api.Webhook
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful create",
			topic:   "orders/create",
			address: "https://example.com/webhook",
			format:  "json",
			mockResp: &api.Webhook{
				ID:         "wh_new",
				Topic:      "orders/create",
				Address:    "https://example.com/webhook",
				Format:     api.WebhookFormatJSON,
				APIVersion: "2024-01",
				CreatedAt:  testTime,
			},
		},
		{
			name:    "create with XML format",
			topic:   "products/update",
			address: "https://example.com/products",
			format:  "xml",
			mockResp: &api.Webhook{
				ID:         "wh_xml",
				Topic:      "products/update",
				Address:    "https://example.com/products",
				Format:     api.WebhookFormatXML,
				APIVersion: "2024-01",
				CreatedAt:  testTime,
			},
		},
		{
			name:    "API error",
			topic:   "orders/create",
			address: "https://example.com/webhook",
			format:  "json",
			mockErr: errors.New("failed to create webhook"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &webhooksMockAPIClient{
				createWebhookResp: tt.mockResp,
				createWebhookErr:  tt.mockErr,
			}
			cleanup, _ := setupWebhooksMockFactories(mockClient)
			defer cleanup()

			cmd := newWebhooksTestCmd()
			cmd.Flags().String("topic", tt.topic, "")
			cmd.Flags().String("address", tt.address, "")
			cmd.Flags().String("format", tt.format, "")
			cmd.Flags().String("api-version", "", "")

			err := webhooksCreateCmd.RunE(cmd, []string{})

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

// TestWebhooksCreateJSONOutput tests the webhooks create command with JSON output.
func TestWebhooksCreateJSONOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &webhooksMockAPIClient{
		createWebhookResp: &api.Webhook{
			ID:         "wh_new",
			Topic:      "orders/create",
			Address:    "https://example.com/webhook",
			Format:     api.WebhookFormatJSON,
			APIVersion: "2024-01",
			CreatedAt:  testTime,
		},
	}
	cleanup, buf := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("topic", "orders/create", "")
	cmd.Flags().String("address", "https://example.com/webhook", "")
	cmd.Flags().String("format", "json", "")
	cmd.Flags().String("api-version", "", "")

	err := webhooksCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "wh_new") {
		t.Errorf("JSON output should contain webhook ID, got %q", output)
	}
}

// TestWebhooksCreateDryRun verifies dry-run mode works
func TestWebhooksCreateDryRun(t *testing.T) {
	mockClient := &webhooksMockAPIClient{}
	cleanup, _ := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("topic", "orders/create", "")
	cmd.Flags().String("address", "https://example.com/webhook", "")
	cmd.Flags().String("format", "json", "")
	cmd.Flags().String("api-version", "", "")

	err := webhooksCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify API was NOT called
	if mockClient.createWebhookCalls > 0 {
		t.Error("API should not be called in dry-run mode")
	}
}

// TestWebhooksCreateInvalidAddress verifies address validation
func TestWebhooksCreateInvalidAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"http rejected", "http://example.com/webhook", true},
		{"https accepted", "https://example.com/webhook", false},
		{"no scheme rejected", "example.com/webhook", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebhookAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
		})
	}
}

// TestWebhooksDeleteRunE tests the webhooks delete command with mock API.
func TestWebhooksDeleteRunE(t *testing.T) {
	tests := []struct {
		name      string
		webhookID string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful delete",
			webhookID: "wh_123",
			mockErr:   nil,
		},
		{
			name:      "webhook not found",
			webhookID: "wh_999",
			mockErr:   errors.New("webhook not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &webhooksMockAPIClient{
				deleteWebhookErr: tt.mockErr,
			}
			cleanup, _ := setupWebhooksMockFactories(mockClient)
			defer cleanup()

			cmd := newWebhooksTestCmd()

			err := webhooksDeleteCmd.RunE(cmd, []string{tt.webhookID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify the correct ID was passed
			if mockClient.deleteWebhookID != tt.webhookID {
				t.Errorf("expected delete ID %q, got %q", tt.webhookID, mockClient.deleteWebhookID)
			}
		})
	}
}

// TestWebhooksDeleteDryRun verifies dry-run mode on delete
func TestWebhooksDeleteDryRun(t *testing.T) {
	mockClient := &webhooksMockAPIClient{}
	cleanup, _ := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := webhooksDeleteCmd.RunE(cmd, []string{"wh_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify API was NOT called
	if mockClient.deleteWebhookID != "" {
		t.Error("API should not be called in dry-run mode")
	}
}

// TestWebhooksListGetClientError verifies error handling when getClient fails
func TestWebhooksListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newWebhooksTestCmd()
	cmd.Flags().String("topic", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := webhooksListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestWebhooksGetGetClientError verifies error handling when getClient fails on get
func TestWebhooksGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newWebhooksTestCmd()

	err := webhooksGetCmd.RunE(cmd, []string{"wh_123"})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestWebhooksCreateGetClientError verifies error handling when getClient fails on create
func TestWebhooksCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newWebhooksTestCmd()
	cmd.Flags().String("topic", "orders/create", "")
	cmd.Flags().String("address", "https://example.com/webhook", "")
	cmd.Flags().String("format", "json", "")
	cmd.Flags().String("api-version", "", "")

	err := webhooksCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestWebhooksDeleteGetClientError verifies error handling when getClient fails on delete
func TestWebhooksDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newWebhooksTestCmd()

	err := webhooksDeleteCmd.RunE(cmd, []string{"wh_123"})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestWebhooksWithMockStore tests webhook commands with a mock credential store
func TestWebhooksWithMockStore(t *testing.T) {
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

// TestValidateWebhookAddressHTTPS tests webhook address validation with various inputs.
func TestValidateWebhookAddressHTTPS(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid https url",
			address: "https://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "valid https with port",
			address: "https://example.com:8443/webhook",
			wantErr: false,
		},
		{
			name:    "valid https uppercase",
			address: "HTTPS://example.com/webhook",
			wantErr: false,
		},
		{
			name:    "http url rejected",
			address: "http://example.com/webhook",
			wantErr: true,
		},
		{
			name:    "no scheme rejected",
			address: "example.com/webhook",
			wantErr: true,
		},
		{
			name:    "empty string rejected",
			address: "",
			wantErr: true,
		},
		{
			name:    "ftp scheme rejected",
			address: "ftp://example.com/file",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebhookAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				expectedMsg := "webhook address must be a valid HTTPS URL"
				if err.Error() != expectedMsg {
					t.Errorf("validateWebhookAddress(%q) error message = %q, want %q", tt.address, err.Error(), expectedMsg)
				}
			}
		})
	}
}

// TestWebhooksCreateAddressValidationIntegration tests that address validation
// is called during create command execution.
func TestWebhooksCreateAddressValidationIntegration(t *testing.T) {
	mockClient := &webhooksMockAPIClient{}
	cleanup, _ := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	cmd.Flags().String("topic", "orders/create", "")
	cmd.Flags().String("address", "http://example.com/webhook", "") // HTTP, not HTTPS
	cmd.Flags().String("format", "json", "")
	cmd.Flags().String("api-version", "", "")

	err := webhooksCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for HTTP address")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("expected HTTPS validation error, got: %v", err)
	}

	// Verify API was NOT called due to validation failure
	if mockClient.createWebhookCalls > 0 {
		t.Error("API should not be called when address validation fails")
	}
}

// TestWebhooksCreateEmptyFormat tests create with empty format.
func TestWebhooksCreateEmptyFormat(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &webhooksMockAPIClient{
		createWebhookResp: &api.Webhook{
			ID:         "wh_new",
			Topic:      "orders/create",
			Address:    "https://example.com/webhook",
			Format:     api.WebhookFormatJSON,
			APIVersion: "2024-01",
			CreatedAt:  testTime,
		},
	}
	cleanup, _ := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	cmd.Flags().String("topic", "orders/create", "")
	cmd.Flags().String("address", "https://example.com/webhook", "")
	cmd.Flags().String("format", "", "") // Empty format
	cmd.Flags().String("api-version", "", "")

	err := webhooksCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWebhooksCreateWithAPIVersion tests create with custom API version.
func TestWebhooksCreateWithAPIVersion(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &webhooksMockAPIClient{
		createWebhookResp: &api.Webhook{
			ID:         "wh_new",
			Topic:      "orders/create",
			Address:    "https://example.com/webhook",
			Format:     api.WebhookFormatJSON,
			APIVersion: "2024-07",
			CreatedAt:  testTime,
		},
	}
	cleanup, _ := setupWebhooksMockFactories(mockClient)
	defer cleanup()

	cmd := newWebhooksTestCmd()
	cmd.Flags().String("topic", "orders/create", "")
	cmd.Flags().String("address", "https://example.com/webhook", "")
	cmd.Flags().String("format", "json", "")
	cmd.Flags().String("api-version", "2024-07", "")

	err := webhooksCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

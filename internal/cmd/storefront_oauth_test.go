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
)

func TestStorefrontOAuthCmdStructure(t *testing.T) {
	if storefrontOAuthCmd.Use != "storefront-oauth" {
		t.Errorf("Expected Use 'storefront-oauth', got %q", storefrontOAuthCmd.Use)
	}

	subcommands := storefrontOAuthCmd.Commands()
	expectedSubs := []string{"list", "get", "create", "update", "delete", "rotate-secret"}

	for _, exp := range expectedSubs {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == exp || strings.HasPrefix(cmd.Use, exp+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing subcommand: %s", exp)
		}
	}
}

func TestStorefrontOAuthListCmdFlags(t *testing.T) {
	pageFlag := storefrontOAuthListCmd.Flags().Lookup("page")
	if pageFlag == nil {
		t.Error("Missing --page flag")
	}

	pageSizeFlag := storefrontOAuthListCmd.Flags().Lookup("page-size")
	if pageSizeFlag == nil {
		t.Error("Missing --page-size flag")
	}
}

func TestStorefrontOAuthGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"oauth_123"}, wantErr: false},
		{name: "too many args", args: []string{"oauth_1", "oauth_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontOAuthGetCmd.Args(storefrontOAuthGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorefrontOAuthDeleteCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"oauth_123"}, wantErr: false},
		{name: "too many args", args: []string{"oauth_1", "oauth_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontOAuthDeleteCmd.Args(storefrontOAuthDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorefrontOAuthListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(storefrontOAuthListCmd)

	err := storefrontOAuthListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// storefrontOAuthMockAPIClient is a mock implementation of api.APIClient for storefront OAuth tests.
type storefrontOAuthMockAPIClient struct {
	api.MockClient
	listResp   *api.StorefrontOAuthClientsListResponse
	listErr    error
	getResp    *api.StorefrontOAuthClient
	getErr     error
	createResp *api.StorefrontOAuthClient
	createErr  error
	updateResp *api.StorefrontOAuthClient
	updateErr  error
	deleteErr  error
	rotateResp *api.StorefrontOAuthClient
	rotateErr  error
}

func (m *storefrontOAuthMockAPIClient) ListStorefrontOAuthClients(ctx context.Context, opts *api.StorefrontOAuthClientsListOptions) (*api.StorefrontOAuthClientsListResponse, error) {
	return m.listResp, m.listErr
}

func (m *storefrontOAuthMockAPIClient) GetStorefrontOAuthClient(ctx context.Context, id string) (*api.StorefrontOAuthClient, error) {
	return m.getResp, m.getErr
}

func (m *storefrontOAuthMockAPIClient) CreateStorefrontOAuthClient(ctx context.Context, req *api.StorefrontOAuthClientCreateRequest) (*api.StorefrontOAuthClient, error) {
	return m.createResp, m.createErr
}

func (m *storefrontOAuthMockAPIClient) UpdateStorefrontOAuthClient(ctx context.Context, id string, req *api.StorefrontOAuthClientUpdateRequest) (*api.StorefrontOAuthClient, error) {
	return m.updateResp, m.updateErr
}

func (m *storefrontOAuthMockAPIClient) DeleteStorefrontOAuthClient(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *storefrontOAuthMockAPIClient) RotateStorefrontOAuthClientSecret(ctx context.Context, id string) (*api.StorefrontOAuthClient, error) {
	return m.rotateResp, m.rotateErr
}

// setupStorefrontOAuthMockFactories sets up mock factories for storefront OAuth tests.
func setupStorefrontOAuthMockFactories(mockClient *storefrontOAuthMockAPIClient) (func(), *bytes.Buffer) {
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

// TestStorefrontOAuthListRunE tests the list command with mock API.
func TestStorefrontOAuthListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.StorefrontOAuthClientsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.StorefrontOAuthClientsListResponse{
				Items: []api.StorefrontOAuthClient{
					{
						ID:        "oauth_123",
						Name:      "Test Client",
						ClientID:  "client_abc",
						Scopes:    []string{"read", "write"},
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "oauth_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.StorefrontOAuthClientsListResponse{
				Items:      []api.StorefrontOAuthClient{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple clients",
			mockResp: &api.StorefrontOAuthClientsListResponse{
				Items: []api.StorefrontOAuthClient{
					{ID: "oauth_1", Name: "Client 1", ClientID: "client_1", CreatedAt: time.Now()},
					{ID: "oauth_2", Name: "Client 2", ClientID: "client_2", CreatedAt: time.Now()},
				},
				TotalCount: 2,
			},
			wantOutput: "oauth_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontOAuthMockAPIClient{
				listResp: tt.mockResp,
				listErr:  tt.mockErr,
			}
			cleanup, buf := setupStorefrontOAuthMockFactories(mockClient)
			defer cleanup()

			cmd := newTestCmdWithFlags()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := storefrontOAuthListCmd.RunE(cmd, []string{})

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

// TestStorefrontOAuthListRunEWithJSON tests JSON output format.
func TestStorefrontOAuthListRunEWithJSON(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		listResp: &api.StorefrontOAuthClientsListResponse{
			Items: []api.StorefrontOAuthClient{
				{ID: "oauth_json", Name: "JSON Client", ClientID: "client_json"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := storefrontOAuthListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "oauth_json") {
		t.Errorf("JSON output should contain client ID, got: %s", output)
	}
}

// TestStorefrontOAuthGetRunE tests the get command with mock API.
func TestStorefrontOAuthGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		mockResp *api.StorefrontOAuthClient
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			clientID: "oauth_123",
			mockResp: &api.StorefrontOAuthClient{
				ID:           "oauth_123",
				Name:         "Test Client",
				ClientID:     "client_abc",
				RedirectURIs: []string{"https://example.com/callback"},
				Scopes:       []string{"read", "write"},
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:     "client not found",
			clientID: "oauth_999",
			mockErr:  errors.New("not found"),
			wantErr:  true,
		},
		{
			name:     "get client with empty scopes",
			clientID: "oauth_456",
			mockResp: &api.StorefrontOAuthClient{
				ID:           "oauth_456",
				Name:         "Simple Client",
				ClientID:     "client_simple",
				RedirectURIs: []string{},
				Scopes:       []string{},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontOAuthMockAPIClient{
				getResp: tt.mockResp,
				getErr:  tt.mockErr,
			}
			cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
			defer cleanup()

			cmd := newTestCmdWithFlags()

			err := storefrontOAuthGetCmd.RunE(cmd, []string{tt.clientID})

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

// TestStorefrontOAuthGetRunEWithJSON tests JSON output format for get command.
func TestStorefrontOAuthGetRunEWithJSON(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		getResp: &api.StorefrontOAuthClient{
			ID:       "oauth_json_get",
			Name:     "JSON Get Client",
			ClientID: "client_json_get",
		},
	}
	cleanup, buf := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("output", "json")

	err := storefrontOAuthGetCmd.RunE(cmd, []string{"oauth_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "oauth_json_get") {
		t.Errorf("JSON output should contain client ID, got: %s", output)
	}
}

// TestStorefrontOAuthCreateRunE tests the create command with mock API.
func TestStorefrontOAuthCreateRunE(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		createResp: &api.StorefrontOAuthClient{
			ID:           "oauth_new",
			Name:         "New Client",
			ClientID:     "client_new",
			ClientSecret: "secret_new_123",
		},
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "New Client", "")
	cmd.Flags().String("redirect-uris", "https://example.com/callback", "")
	cmd.Flags().String("scopes", "read,write", "")

	err := storefrontOAuthCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Note: Text output goes to stdout via fmt.Printf, not the formatter buffer
	// JSON output is tested in TestStorefrontOAuthCreateWithJSON
}

// TestStorefrontOAuthCreateAPIError tests create command error handling.
func TestStorefrontOAuthCreateAPIError(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		createErr: errors.New("validation error"),
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("redirect-uris", "https://example.com", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestStorefrontOAuthCreateDryRun tests create command dry run mode.
func TestStorefrontOAuthCreateDryRun(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test Client", "")
	cmd.Flags().String("redirect-uris", "https://example.com", "")
	cmd.Flags().String("scopes", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontOAuthCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthCreateWithJSON tests JSON output format for create command.
func TestStorefrontOAuthCreateWithJSON(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		createResp: &api.StorefrontOAuthClient{
			ID:           "oauth_json_create",
			Name:         "JSON Create Client",
			ClientID:     "client_json_create",
			ClientSecret: "secret_json_create",
		},
	}
	cleanup, buf := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "JSON Create Client", "")
	cmd.Flags().String("redirect-uris", "https://example.com/callback", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "oauth_json_create") {
		t.Errorf("JSON output should contain client ID, got: %s", output)
	}
}

// TestStorefrontOAuthUpdateRunE tests the update command with mock API.
func TestStorefrontOAuthUpdateRunE(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		updateResp: &api.StorefrontOAuthClient{
			ID:   "oauth_123",
			Name: "Updated Client",
		},
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Updated Client", "")
	cmd.Flags().String("redirect-uris", "", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthUpdateCmd.RunE(cmd, []string{"oauth_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthUpdateAPIError tests update command error handling.
func TestStorefrontOAuthUpdateAPIError(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		updateErr: errors.New("update failed"),
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Updated Client", "")
	cmd.Flags().String("redirect-uris", "", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthUpdateCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestStorefrontOAuthUpdateDryRun tests update command dry run mode.
func TestStorefrontOAuthUpdateDryRun(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Updated Client", "")
	cmd.Flags().String("redirect-uris", "", "")
	cmd.Flags().String("scopes", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontOAuthUpdateCmd.RunE(cmd, []string{"oauth_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthUpdateWithJSON tests JSON output format for update command.
func TestStorefrontOAuthUpdateWithJSON(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		updateResp: &api.StorefrontOAuthClient{
			ID:   "oauth_json_update",
			Name: "JSON Update Client",
		},
	}
	cleanup, buf := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "JSON Update Client", "")
	cmd.Flags().String("redirect-uris", "", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthUpdateCmd.RunE(cmd, []string{"oauth_json_update"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "oauth_json_update") {
		t.Errorf("JSON output should contain client ID, got: %s", output)
	}
}

// TestStorefrontOAuthDeleteRunE tests the delete command with mock API.
func TestStorefrontOAuthDeleteRunE(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		deleteErr: nil,
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthDeleteCmd.RunE(cmd, []string{"oauth_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthDeleteAPIError tests delete command error handling.
func TestStorefrontOAuthDeleteAPIError(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		deleteErr: errors.New("cannot delete"),
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthDeleteCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestStorefrontOAuthDeleteDryRun tests delete command dry run mode.
func TestStorefrontOAuthDeleteDryRun(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontOAuthDeleteCmd.RunE(cmd, []string{"oauth_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthRotateRunE tests the rotate-secret command with mock API.
func TestStorefrontOAuthRotateRunE(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		rotateResp: &api.StorefrontOAuthClient{
			ID:           "oauth_123",
			Name:         "Test Client",
			ClientID:     "client_abc",
			ClientSecret: "new_secret_456",
		},
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthRotateCmd.RunE(cmd, []string{"oauth_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthRotateAPIError tests rotate-secret command error handling.
func TestStorefrontOAuthRotateAPIError(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		rotateErr: errors.New("rotation failed"),
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthRotateCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestStorefrontOAuthRotateDryRun tests rotate-secret command dry run mode.
func TestStorefrontOAuthRotateDryRun(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontOAuthRotateCmd.RunE(cmd, []string{"oauth_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthRotateWithJSON tests JSON output format for rotate-secret command.
func TestStorefrontOAuthRotateWithJSON(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		rotateResp: &api.StorefrontOAuthClient{
			ID:           "oauth_json_rotate",
			Name:         "JSON Rotate Client",
			ClientID:     "client_json_rotate",
			ClientSecret: "new_secret_json",
		},
	}
	cleanup, buf := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthRotateCmd.RunE(cmd, []string{"oauth_json_rotate"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "oauth_json_rotate") {
		t.Errorf("JSON output should contain client ID, got: %s", output)
	}
}

// TestStorefrontOAuthUpdateCmdArgs tests update command argument validation.
func TestStorefrontOAuthUpdateCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"oauth_123"}, wantErr: false},
		{name: "too many args", args: []string{"oauth_1", "oauth_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontOAuthUpdateCmd.Args(storefrontOAuthUpdateCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStorefrontOAuthRotateCmdArgs tests rotate-secret command argument validation.
func TestStorefrontOAuthRotateCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"oauth_123"}, wantErr: false},
		{name: "too many args", args: []string{"oauth_1", "oauth_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontOAuthRotateCmd.Args(storefrontOAuthRotateCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStorefrontOAuthGetClientError tests get command error handling when getClient fails.
func TestStorefrontOAuthGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(storefrontOAuthGetCmd)

	err := storefrontOAuthGetCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestStorefrontOAuthCreateGetClientError tests create command error handling when getClient fails.
func TestStorefrontOAuthCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("redirect-uris", "https://example.com", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestStorefrontOAuthUpdateGetClientError tests update command error handling when getClient fails.
func TestStorefrontOAuthUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("redirect-uris", "", "")
	cmd.Flags().String("scopes", "", "")

	err := storefrontOAuthUpdateCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestStorefrontOAuthDeleteGetClientError tests delete command error handling when getClient fails.
func TestStorefrontOAuthDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthDeleteCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestStorefrontOAuthRotateGetClientError tests rotate-secret command error handling when getClient fails.
func TestStorefrontOAuthRotateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := storefrontOAuthRotateCmd.RunE(cmd, []string{"oauth_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestStorefrontOAuthCreateWithMultipleRedirectURIs tests create command with multiple redirect URIs.
func TestStorefrontOAuthCreateWithMultipleRedirectURIs(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		createResp: &api.StorefrontOAuthClient{
			ID:           "oauth_multi",
			Name:         "Multi URI Client",
			ClientID:     "client_multi",
			RedirectURIs: []string{"https://example.com/callback", "https://app.example.com/auth"},
		},
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Multi URI Client", "")
	cmd.Flags().String("redirect-uris", "https://example.com/callback, https://app.example.com/auth", "")
	cmd.Flags().String("scopes", "read, write, admin", "")

	err := storefrontOAuthCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontOAuthUpdateWithAllFields tests update command with all fields.
func TestStorefrontOAuthUpdateWithAllFields(t *testing.T) {
	mockClient := &storefrontOAuthMockAPIClient{
		updateResp: &api.StorefrontOAuthClient{
			ID:           "oauth_full",
			Name:         "Fully Updated Client",
			ClientID:     "client_full",
			RedirectURIs: []string{"https://new.example.com/callback"},
			Scopes:       []string{"read", "write", "delete"},
		},
	}
	cleanup, _ := setupStorefrontOAuthMockFactories(mockClient)
	defer cleanup()

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Fully Updated Client", "")
	cmd.Flags().String("redirect-uris", "https://new.example.com/callback", "")
	cmd.Flags().String("scopes", "read,write,delete", "")

	err := storefrontOAuthUpdateCmd.RunE(cmd, []string{"oauth_full"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

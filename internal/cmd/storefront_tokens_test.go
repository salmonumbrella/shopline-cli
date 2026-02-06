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

// storefrontTokensMockAPIClient is a mock implementation of api.APIClient for storefront tokens tests.
type storefrontTokensMockAPIClient struct {
	api.MockClient

	listStorefrontTokensResp *api.StorefrontTokensListResponse
	listStorefrontTokensErr  error

	getStorefrontTokenResp *api.StorefrontToken
	getStorefrontTokenErr  error

	createStorefrontTokenResp *api.StorefrontToken
	createStorefrontTokenErr  error

	deleteStorefrontTokenErr error
}

func (m *storefrontTokensMockAPIClient) ListStorefrontTokens(ctx context.Context, opts *api.StorefrontTokensListOptions) (*api.StorefrontTokensListResponse, error) {
	return m.listStorefrontTokensResp, m.listStorefrontTokensErr
}

func (m *storefrontTokensMockAPIClient) GetStorefrontToken(ctx context.Context, id string) (*api.StorefrontToken, error) {
	return m.getStorefrontTokenResp, m.getStorefrontTokenErr
}

func (m *storefrontTokensMockAPIClient) CreateStorefrontToken(ctx context.Context, req *api.StorefrontTokenCreateRequest) (*api.StorefrontToken, error) {
	return m.createStorefrontTokenResp, m.createStorefrontTokenErr
}

func (m *storefrontTokensMockAPIClient) DeleteStorefrontToken(ctx context.Context, id string) error {
	return m.deleteStorefrontTokenErr
}

// setupStorefrontTokensMockFactories sets up mock factories for storefront tokens tests.
func setupStorefrontTokensMockFactories(mockClient *storefrontTokensMockAPIClient) (func(), *bytes.Buffer) {
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

// newStorefrontTokensTestCmd creates a test command with common flags for storefront tokens tests.
func newStorefrontTokensTestCmd() *cobra.Command {
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

// TestStorefrontTokensCmdStructure verifies storefront tokens command structure
func TestStorefrontTokensCmdStructure(t *testing.T) {
	if storefrontTokensCmd.Use != "storefront-tokens" {
		t.Errorf("Expected Use 'storefront-tokens', got %q", storefrontTokensCmd.Use)
	}

	subcommands := storefrontTokensCmd.Commands()
	expectedSubs := []string{"list", "get", "create", "delete"}

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

// TestStorefrontTokensSubcommands verifies all subcommands are registered
func TestStorefrontTokensSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List storefront access tokens",
		"get":    "Get storefront token details",
		"create": "Create a storefront access token",
		"delete": "Delete a storefront access token",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range storefrontTokensCmd.Commands() {
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

// TestStorefrontTokensListCmdFlags verifies list command flags exist with correct defaults
func TestStorefrontTokensListCmdFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := storefrontTokensListCmd.Flags().Lookup(f.name)
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

// TestStorefrontTokensGetCmdArgs verifies get command argument validation
func TestStorefrontTokensGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"token_123"}, wantErr: false},
		{name: "too many args", args: []string{"token_1", "token_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontTokensGetCmd.Args(storefrontTokensGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStorefrontTokensCreateCmdFlags verifies create command flags exist
func TestStorefrontTokensCreateCmdFlags(t *testing.T) {
	flag := storefrontTokensCreateCmd.Flags().Lookup("title")
	if flag == nil {
		t.Error("flag 'title' not found")
	}
}

// TestStorefrontTokensDeleteCmdArgs verifies delete command argument validation
func TestStorefrontTokensDeleteCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"token_123"}, wantErr: false},
		{name: "too many args", args: []string{"token_1", "token_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontTokensDeleteCmd.Args(storefrontTokensDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStorefrontTokensListGetClientError verifies error handling when getClient fails
func TestStorefrontTokensListGetClientError(t *testing.T) {
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
	cmd.AddCommand(storefrontTokensListCmd)

	err := storefrontTokensListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStorefrontTokensGetGetClientError verifies error handling when getClient fails
func TestStorefrontTokensGetGetClientError(t *testing.T) {
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
	err := storefrontTokensGetCmd.RunE(cmd, []string{"token_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStorefrontTokensCreateGetClientError verifies error handling when getClient fails
func TestStorefrontTokensCreateGetClientError(t *testing.T) {
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
	cmd.Flags().String("title", "Test Token", "")

	err := storefrontTokensCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStorefrontTokensDeleteGetClientError verifies error handling when getClient fails
func TestStorefrontTokensDeleteGetClientError(t *testing.T) {
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

	cmd := newStorefrontTokensTestCmd() // Use test cmd which has --yes flag set
	err := storefrontTokensDeleteCmd.RunE(cmd, []string{"token_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStorefrontTokensListRunE tests the storefront tokens list command execution with mock API.
func TestStorefrontTokensListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.StorefrontTokensListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.StorefrontTokensListResponse{
				Items: []api.StorefrontToken{
					{
						ID:        "sft_123",
						Title:     "Test Storefront Token",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sft_123",
		},
		{
			name: "multiple tokens",
			mockResp: &api.StorefrontTokensListResponse{
				Items: []api.StorefrontToken{
					{
						ID:        "sft_123",
						Title:     "Token 1",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:        "sft_456",
						Title:     "Token 2",
						CreatedAt: time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "sft_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.StorefrontTokensListResponse{
				Items:      []api.StorefrontToken{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontTokensMockAPIClient{
				listStorefrontTokensResp: tt.mockResp,
				listStorefrontTokensErr:  tt.mockErr,
			}
			cleanup, buf := setupStorefrontTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontTokensTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := storefrontTokensListCmd.RunE(cmd, []string{})

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

// TestStorefrontTokensListRunEJSON tests the storefront tokens list command with JSON output.
func TestStorefrontTokensListRunEJSON(t *testing.T) {
	mockClient := &storefrontTokensMockAPIClient{
		listStorefrontTokensResp: &api.StorefrontTokensListResponse{
			Items: []api.StorefrontToken{
				{
					ID:        "sft_json_123",
					Title:     "JSON Test Storefront Token",
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupStorefrontTokensMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontTokensTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := storefrontTokensListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "sft_json_123") {
		t.Errorf("JSON output %q should contain token ID", output)
	}
}

// TestStorefrontTokensGetRunE tests the storefront tokens get command execution with mock API.
func TestStorefrontTokensGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		tokenID  string
		mockResp *api.StorefrontToken
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			tokenID: "sft_123",
			mockResp: &api.StorefrontToken{
				ID:        "sft_123",
				Title:     "Test Storefront Token",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "token not found",
			tokenID: "sft_999",
			mockErr: errors.New("token not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontTokensMockAPIClient{
				getStorefrontTokenResp: tt.mockResp,
				getStorefrontTokenErr:  tt.mockErr,
			}
			cleanup, _ := setupStorefrontTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontTokensTestCmd()

			err := storefrontTokensGetCmd.RunE(cmd, []string{tt.tokenID})

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

// TestStorefrontTokensGetRunEJSON tests the storefront tokens get command with JSON output.
func TestStorefrontTokensGetRunEJSON(t *testing.T) {
	mockClient := &storefrontTokensMockAPIClient{
		getStorefrontTokenResp: &api.StorefrontToken{
			ID:        "sft_json_get",
			Title:     "JSON Get Storefront Token",
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupStorefrontTokensMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontTokensTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := storefrontTokensGetCmd.RunE(cmd, []string{"sft_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "sft_json_get") {
		t.Errorf("JSON output %q should contain token ID", output)
	}
}

// TestStorefrontTokensCreateRunE tests the storefront tokens create command execution with mock API.
func TestStorefrontTokensCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		mockResp *api.StorefrontToken
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful create",
			title: "New Storefront Token",
			mockResp: &api.StorefrontToken{
				ID:          "sft_new",
				Title:       "New Storefront Token",
				AccessToken: "sf_access_token_secret",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:  "create without access token in response",
			title: "Token No Access",
			mockResp: &api.StorefrontToken{
				ID:          "sft_no_access",
				Title:       "Token No Access",
				AccessToken: "",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "API error on create",
			title:   "Failed Token",
			mockErr: errors.New("failed to create token"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontTokensMockAPIClient{
				createStorefrontTokenResp: tt.mockResp,
				createStorefrontTokenErr:  tt.mockErr,
			}
			cleanup, _ := setupStorefrontTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontTokensTestCmd()
			cmd.Flags().String("title", tt.title, "")

			err := storefrontTokensCreateCmd.RunE(cmd, []string{})

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

// TestStorefrontTokensCreateDryRun tests the storefront tokens create command with dry-run flag.
func TestStorefrontTokensCreateDryRun(t *testing.T) {
	// No mock client needed for dry-run
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	cmd := newStorefrontTokensTestCmd()
	cmd.Flags().String("title", "Dry Run Storefront Token", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontTokensCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestStorefrontTokensCreateRunEJSON tests the storefront tokens create command with JSON output.
func TestStorefrontTokensCreateRunEJSON(t *testing.T) {
	mockClient := &storefrontTokensMockAPIClient{
		createStorefrontTokenResp: &api.StorefrontToken{
			ID:          "sft_json_create",
			Title:       "JSON Create Storefront Token",
			AccessToken: "secret_access_token",
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupStorefrontTokensMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontTokensTestCmd()
	cmd.Flags().String("title", "JSON Create Storefront Token", "")
	_ = cmd.Flags().Set("output", "json")

	err := storefrontTokensCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "sft_json_create") {
		t.Errorf("JSON output %q should contain token ID", output)
	}
}

// TestStorefrontTokensDeleteRunE tests the storefront tokens delete command execution with mock API.
func TestStorefrontTokensDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		tokenID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			tokenID: "sft_123",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			tokenID: "sft_456",
			mockErr: errors.New("token not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontTokensMockAPIClient{
				deleteStorefrontTokenErr: tt.mockErr,
			}
			cleanup, _ := setupStorefrontTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontTokensTestCmd()

			err := storefrontTokensDeleteCmd.RunE(cmd, []string{tt.tokenID})

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

// TestStorefrontTokensDeleteDryRun tests the storefront tokens delete command with dry-run flag.
func TestStorefrontTokensDeleteDryRun(t *testing.T) {
	// No mock client needed for dry-run
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	cmd := newStorefrontTokensTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontTokensDeleteCmd.RunE(cmd, []string{"sft_dry_run"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestStorefrontTokensListRunE_NoProfiles verifies error when no profiles are configured
func TestStorefrontTokensListRunE_NoProfiles(t *testing.T) {
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

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := storefrontTokensListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestStorefrontTokensGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestStorefrontTokensGetRunE_MultipleProfiles(t *testing.T) {
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

	cmd := newTestCmdWithFlags()
	err := storefrontTokensGetCmd.RunE(cmd, []string{"sft_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestStorefrontTokensCreateRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestStorefrontTokensCreateRunE_MultipleProfiles(t *testing.T) {
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

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("title", "Test Token", "")

	err := storefrontTokensCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestStorefrontTokensDeleteRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestStorefrontTokensDeleteRunE_MultipleProfiles(t *testing.T) {
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

	cmd := newStorefrontTokensTestCmd()
	err := storefrontTokensDeleteCmd.RunE(cmd, []string{"sft_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestStorefrontTokensWithMockStore tests storefront token commands with a mock credential store
func TestStorefrontTokensWithMockStore(t *testing.T) {
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

// Ensure unused imports don't cause errors
var _ = secrets.StoreCredentials{}

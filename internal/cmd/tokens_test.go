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

// TestTokensCommandSetup verifies tokens command initialization
func TestTokensCommandSetup(t *testing.T) {
	if tokensCmd.Use != "tokens" {
		t.Errorf("expected Use 'tokens', got %q", tokensCmd.Use)
	}
	if tokensCmd.Short != "Manage API tokens" {
		t.Errorf("expected Short 'Manage API tokens', got %q", tokensCmd.Short)
	}
}

// TestTokensSubcommands verifies all subcommands are registered
func TestTokensSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List API tokens",
		"get":    "Get token details",
		"create": "Create an API token",
		"delete": "Delete an API token",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range tokensCmd.Commands() {
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

// TestTokensListCmdFlags verifies list command flags exist with correct defaults
func TestTokensListCmdFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := tokensListCmd.Flags().Lookup(f.name)
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

// TestTokensGetCmdArgs verifies get command argument validation
func TestTokensGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"tok_123"}, wantErr: false},
		{name: "too many args", args: []string{"tok_1", "tok_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tokensGetCmd.Args(tokensGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTokensCreateCmdFlags verifies create command flags exist
func TestTokensCreateCmdFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"scopes", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := tokensCreateCmd.Flags().Lookup(f.name)
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

// TestTokensDeleteCmdArgs verifies delete command argument validation
func TestTokensDeleteCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"tok_123"}, wantErr: false},
		{name: "too many args", args: []string{"tok_1", "tok_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tokensDeleteCmd.Args(tokensDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// tokensMockAPIClient is a mock implementation of api.APIClient for tokens tests.
type tokensMockAPIClient struct {
	api.MockClient

	listTokensResp *api.TokensListResponse
	listTokensErr  error

	getTokenResp *api.Token
	getTokenErr  error

	createTokenResp *api.Token
	createTokenErr  error

	deleteTokenErr error
}

func (m *tokensMockAPIClient) ListTokens(ctx context.Context, opts *api.TokensListOptions) (*api.TokensListResponse, error) {
	return m.listTokensResp, m.listTokensErr
}

func (m *tokensMockAPIClient) GetToken(ctx context.Context, id string) (*api.Token, error) {
	return m.getTokenResp, m.getTokenErr
}

func (m *tokensMockAPIClient) CreateToken(ctx context.Context, req *api.TokenCreateRequest) (*api.Token, error) {
	return m.createTokenResp, m.createTokenErr
}

func (m *tokensMockAPIClient) DeleteToken(ctx context.Context, id string) error {
	return m.deleteTokenErr
}

// setupTokensMockFactories sets up mock factories for tokens tests.
func setupTokensMockFactories(mockClient *tokensMockAPIClient) (func(), *bytes.Buffer) {
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

// newTokensTestCmd creates a test command with common flags for tokens tests.
func newTokensTestCmd() *cobra.Command {
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

// TestTokensListGetClientError verifies error handling when getClient fails
func TestTokensListGetClientError(t *testing.T) {
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
	cmd.AddCommand(tokensListCmd)

	err := tokensListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestTokensGetGetClientError verifies error handling when getClient fails
func TestTokensGetGetClientError(t *testing.T) {
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
	err := tokensGetCmd.RunE(cmd, []string{"tok_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestTokensCreateGetClientError verifies error handling when getClient fails
func TestTokensCreateGetClientError(t *testing.T) {
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
	cmd.Flags().String("scopes", "", "")

	err := tokensCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestTokensDeleteGetClientError verifies error handling when getClient fails
func TestTokensDeleteGetClientError(t *testing.T) {
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

	cmd := newTokensTestCmd() // Use newTokensTestCmd which has --yes flag set
	err := tokensDeleteCmd.RunE(cmd, []string{"tok_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestTokensListRunE tests the tokens list command execution with mock API.
func TestTokensListRunE(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.TokensListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.TokensListResponse{
				Items: []api.Token{
					{
						ID:        "tok_123",
						Title:     "Test Token",
						Scopes:    []string{"read_products", "write_products"},
						ExpiresAt: &expiresAt,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tok_123",
		},
		{
			name: "list with never-expiring token",
			mockResp: &api.TokensListResponse{
				Items: []api.Token{
					{
						ID:        "tok_456",
						Title:     "Permanent Token",
						Scopes:    []string{"read_orders"},
						ExpiresAt: nil,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tok_456",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.TokensListResponse{
				Items:      []api.Token{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tokensMockAPIClient{
				listTokensResp: tt.mockResp,
				listTokensErr:  tt.mockErr,
			}
			cleanup, buf := setupTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newTokensTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := tokensListCmd.RunE(cmd, []string{})

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

// TestTokensListRunEJSON tests the tokens list command with JSON output.
func TestTokensListRunEJSON(t *testing.T) {
	mockClient := &tokensMockAPIClient{
		listTokensResp: &api.TokensListResponse{
			Items: []api.Token{
				{
					ID:        "tok_json_123",
					Title:     "JSON Test Token",
					Scopes:    []string{"read_products"},
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTokensMockFactories(mockClient)
	defer cleanup()

	cmd := newTokensTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := tokensListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "tok_json_123") {
		t.Errorf("JSON output %q should contain token ID", output)
	}
}

// TestTokensGetRunE tests the tokens get command execution with mock API.
func TestTokensGetRunE(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name      string
		tokenID   string
		mockResp  *api.Token
		mockErr   error
		wantErr   bool
		wantInOut string
	}{
		{
			name:    "successful get",
			tokenID: "tok_123",
			mockResp: &api.Token{
				ID:        "tok_123",
				Title:     "Test Token",
				Scopes:    []string{"read_products", "write_products"},
				ExpiresAt: &expiresAt,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "get token with no expiration",
			tokenID: "tok_456",
			mockResp: &api.Token{
				ID:        "tok_456",
				Title:     "Permanent Token",
				Scopes:    []string{"read_orders"},
				ExpiresAt: nil,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "token not found",
			tokenID: "tok_999",
			mockErr: errors.New("token not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tokensMockAPIClient{
				getTokenResp: tt.mockResp,
				getTokenErr:  tt.mockErr,
			}
			cleanup, _ := setupTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newTokensTestCmd()

			err := tokensGetCmd.RunE(cmd, []string{tt.tokenID})

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

// TestTokensGetRunEJSON tests the tokens get command with JSON output.
func TestTokensGetRunEJSON(t *testing.T) {
	mockClient := &tokensMockAPIClient{
		getTokenResp: &api.Token{
			ID:        "tok_json_get",
			Title:     "JSON Get Token",
			Scopes:    []string{"read_products"},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupTokensMockFactories(mockClient)
	defer cleanup()

	cmd := newTokensTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := tokensGetCmd.RunE(cmd, []string{"tok_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "tok_json_get") {
		t.Errorf("JSON output %q should contain token ID", output)
	}
}

// TestTokensCreateRunE tests the tokens create command execution with mock API.
func TestTokensCreateRunE(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		scopes    string
		mockResp  *api.Token
		mockErr   error
		wantErr   bool
		wantInOut string
	}{
		{
			name:   "successful create",
			title:  "New Token",
			scopes: "read_products, write_products",
			mockResp: &api.Token{
				ID:          "tok_new",
				Title:       "New Token",
				AccessToken: "access_token_secret",
				Scopes:      []string{"read_products", "write_products"},
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:   "create without scopes",
			title:  "Token No Scopes",
			scopes: "",
			mockResp: &api.Token{
				ID:          "tok_no_scopes",
				Title:       "Token No Scopes",
				AccessToken: "access_token_secret",
				Scopes:      []string{},
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:   "create without access token in response",
			title:  "Token No Access",
			scopes: "read_orders",
			mockResp: &api.Token{
				ID:          "tok_no_access",
				Title:       "Token No Access",
				AccessToken: "",
				Scopes:      []string{"read_orders"},
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "API error on create",
			title:   "Failed Token",
			scopes:  "read_products",
			mockErr: errors.New("failed to create token"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tokensMockAPIClient{
				createTokenResp: tt.mockResp,
				createTokenErr:  tt.mockErr,
			}
			cleanup, _ := setupTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newTokensTestCmd()
			cmd.Flags().String("title", tt.title, "")
			cmd.Flags().String("scopes", tt.scopes, "")

			err := tokensCreateCmd.RunE(cmd, []string{})

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

// TestTokensCreateDryRun tests the tokens create command with dry-run flag.
func TestTokensCreateDryRun(t *testing.T) {
	// No mock client needed for dry-run
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	cmd := newTokensTestCmd()
	cmd.Flags().String("title", "Dry Run Token", "")
	cmd.Flags().String("scopes", "read_products, write_products", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := tokensCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestTokensCreateRunEJSON tests the tokens create command with JSON output.
func TestTokensCreateRunEJSON(t *testing.T) {
	mockClient := &tokensMockAPIClient{
		createTokenResp: &api.Token{
			ID:          "tok_json_create",
			Title:       "JSON Create Token",
			AccessToken: "secret_access_token",
			Scopes:      []string{"read_products"},
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupTokensMockFactories(mockClient)
	defer cleanup()

	cmd := newTokensTestCmd()
	cmd.Flags().String("title", "JSON Create Token", "")
	cmd.Flags().String("scopes", "read_products", "")
	_ = cmd.Flags().Set("output", "json")

	err := tokensCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "tok_json_create") {
		t.Errorf("JSON output %q should contain token ID", output)
	}
}

// TestTokensDeleteRunE tests the tokens delete command execution with mock API.
func TestTokensDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		tokenID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			tokenID: "tok_123",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			tokenID: "tok_456",
			mockErr: errors.New("token not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tokensMockAPIClient{
				deleteTokenErr: tt.mockErr,
			}
			cleanup, _ := setupTokensMockFactories(mockClient)
			defer cleanup()

			cmd := newTokensTestCmd()

			err := tokensDeleteCmd.RunE(cmd, []string{tt.tokenID})

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

// TestTokensDeleteDryRun tests the tokens delete command with dry-run flag.
func TestTokensDeleteDryRun(t *testing.T) {
	// No mock client needed for dry-run
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	cmd := newTokensTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := tokensDeleteCmd.RunE(cmd, []string{"tok_dry_run"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestTokensWithMockStore tests token commands with a mock credential store
func TestTokensWithMockStore(t *testing.T) {
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

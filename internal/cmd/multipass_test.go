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

// TestMultipassCmdStructure verifies multipass command initialization.
func TestMultipassCmdStructure(t *testing.T) {
	if multipassCmd.Use != "multipass" {
		t.Errorf("expected Use 'multipass', got %q", multipassCmd.Use)
	}
	if multipassCmd.Short != "Manage multipass authentication" {
		t.Errorf("expected Short 'Manage multipass authentication', got %q", multipassCmd.Short)
	}
}

// TestMultipassSubcommands verifies all subcommands are registered.
func TestMultipassSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"status":    "Get multipass configuration status",
		"enable":    "Enable multipass authentication",
		"disable":   "Disable multipass authentication",
		"rotate":    "Rotate multipass secret",
		"token":     "Generate a multipass login token",
		"secret":    "Manage multipass secret (documented endpoints)",
		"linkings":  "Manage multipass linking records (documented endpoints)",
		"customers": "Manage multipass customer linkings (documented endpoints)",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range multipassCmd.Commands() {
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

// TestMultipassTokenCmdFlags verifies token command flags exist.
func TestMultipassTokenCmdFlags(t *testing.T) {
	emailFlag := multipassTokenCmd.Flags().Lookup("email")
	if emailFlag == nil {
		t.Error("Missing --email flag")
	}

	returnToFlag := multipassTokenCmd.Flags().Lookup("return-to")
	if returnToFlag == nil {
		t.Error("Missing --return-to flag")
	}
}

// TestMultipassStatusGetClientError verifies error handling when getClient fails.
func TestMultipassStatusGetClientError(t *testing.T) {
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
	cmd.AddCommand(multipassStatusCmd)

	err := multipassStatusCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestMultipassEnableGetClientError verifies error handling when getClient fails on enable.
func TestMultipassEnableGetClientError(t *testing.T) {
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
	cmd.AddCommand(multipassEnableCmd)

	err := multipassEnableCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMultipassDisableGetClientError verifies error handling when getClient fails on disable.
func TestMultipassDisableGetClientError(t *testing.T) {
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
	cmd.AddCommand(multipassDisableCmd)
	_ = cmd.Flags().Set("yes", "true")

	err := multipassDisableCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMultipassRotateGetClientError verifies error handling when getClient fails on rotate.
func TestMultipassRotateGetClientError(t *testing.T) {
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
	cmd.AddCommand(multipassRotateCmd)
	_ = cmd.Flags().Set("yes", "true")

	err := multipassRotateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMultipassTokenGetClientError verifies error handling when getClient fails on token.
func TestMultipassTokenGetClientError(t *testing.T) {
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
	cmd.AddCommand(multipassTokenCmd)
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("return-to", "", "")

	err := multipassTokenCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestMultipassEnableDryRun verifies dry-run mode works for enable.
func TestMultipassEnableDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.AddCommand(multipassEnableCmd)
	_ = cmd.Flags().Set("dry-run", "true")

	err := multipassEnableCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestMultipassDisableDryRun verifies dry-run mode works for disable.
func TestMultipassDisableDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.AddCommand(multipassDisableCmd)
	_ = cmd.Flags().Set("dry-run", "true")

	err := multipassDisableCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestMultipassRotateDryRun verifies dry-run mode works for rotate.
func TestMultipassRotateDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.AddCommand(multipassRotateCmd)
	_ = cmd.Flags().Set("dry-run", "true")

	err := multipassRotateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestMultipassTokenDryRun verifies dry-run mode works for token.
func TestMultipassTokenDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.AddCommand(multipassTokenCmd)
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("return-to", "", "")

	err := multipassTokenCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// multipassMockAPIClient is a mock implementation of api.APIClient for multipass tests.
type multipassMockAPIClient struct {
	api.MockClient
	getMultipassResp           *api.Multipass
	getMultipassErr            error
	enableMultipassResp        *api.Multipass
	enableMultipassErr         error
	disableMultipassErr        error
	rotateMultipassSecretResp  *api.Multipass
	rotateMultipassSecretErr   error
	generateMultipassTokenResp *api.MultipassToken
	generateMultipassTokenErr  error

	getMultipassSecretResp    json.RawMessage
	getMultipassSecretErr     error
	createMultipassSecretResp json.RawMessage
	createMultipassSecretErr  error
	listMultipassLinkingsResp json.RawMessage
	listMultipassLinkingsErr  error
	updateCustomerLinkingResp json.RawMessage
	updateCustomerLinkingErr  error
	deleteCustomerLinkingResp json.RawMessage
	deleteCustomerLinkingErr  error
}

func (m *multipassMockAPIClient) GetMultipass(ctx context.Context) (*api.Multipass, error) {
	return m.getMultipassResp, m.getMultipassErr
}

func (m *multipassMockAPIClient) EnableMultipass(ctx context.Context) (*api.Multipass, error) {
	return m.enableMultipassResp, m.enableMultipassErr
}

func (m *multipassMockAPIClient) DisableMultipass(ctx context.Context) error {
	return m.disableMultipassErr
}

func (m *multipassMockAPIClient) RotateMultipassSecret(ctx context.Context) (*api.Multipass, error) {
	return m.rotateMultipassSecretResp, m.rotateMultipassSecretErr
}

func (m *multipassMockAPIClient) GenerateMultipassToken(ctx context.Context, req *api.MultipassTokenRequest) (*api.MultipassToken, error) {
	return m.generateMultipassTokenResp, m.generateMultipassTokenErr
}

func (m *multipassMockAPIClient) GetMultipassSecret(ctx context.Context) (json.RawMessage, error) {
	return m.getMultipassSecretResp, m.getMultipassSecretErr
}

func (m *multipassMockAPIClient) CreateMultipassSecret(ctx context.Context, body any) (json.RawMessage, error) {
	return m.createMultipassSecretResp, m.createMultipassSecretErr
}

func (m *multipassMockAPIClient) ListMultipassLinkings(ctx context.Context, customerIDs []string) (json.RawMessage, error) {
	return m.listMultipassLinkingsResp, m.listMultipassLinkingsErr
}

func (m *multipassMockAPIClient) UpdateMultipassCustomerLinking(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return m.updateCustomerLinkingResp, m.updateCustomerLinkingErr
}

func (m *multipassMockAPIClient) DeleteMultipassCustomerLinking(ctx context.Context, customerID string) (json.RawMessage, error) {
	return m.deleteCustomerLinkingResp, m.deleteCustomerLinkingErr
}

// setupMultipassMockFactories sets up mock factories for multipass tests.
func setupMultipassMockFactories(mockClient *multipassMockAPIClient) (func(), *bytes.Buffer) {
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

// newMultipassTestCmd creates a test command with common flags for multipass tests.
func newMultipassTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

func newMultipassTestCmdWithBodyFlags() *cobra.Command {
	cmd := newMultipassTestCmd()
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	return cmd
}

// TestMultipassStatusRunE tests the multipass status command with mock API.
func TestMultipassStatusRunE(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		mockResp *api.Multipass
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful status enabled",
			mockResp: &api.Multipass{
				Enabled:   true,
				CreatedAt: baseTime,
				UpdatedAt: baseTime.Add(24 * time.Hour),
			},
		},
		{
			name: "successful status disabled",
			mockResp: &api.Multipass{
				Enabled:   false,
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &multipassMockAPIClient{
				getMultipassResp: tt.mockResp,
				getMultipassErr:  tt.mockErr,
			}
			cleanup, _ := setupMultipassMockFactories(mockClient)
			defer cleanup()

			cmd := newMultipassTestCmd()

			err := multipassStatusCmd.RunE(cmd, []string{})

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

// TestMultipassStatusRunEWithJSON tests JSON output format for status.
func TestMultipassStatusRunEWithJSON(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		getMultipassResp: &api.Multipass{
			Enabled:   true,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := multipassStatusCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "enabled") {
		t.Errorf("JSON output should contain 'enabled', got: %s", output)
	}
}

// TestMultipassEnableRunE tests the multipass enable command with mock API.
func TestMultipassEnableRunE(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		mockResp *api.Multipass
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful enable with secret",
			mockResp: &api.Multipass{
				Enabled:   true,
				Secret:    "secret_abc123xyz",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
		},
		{
			name: "successful enable without secret",
			mockResp: &api.Multipass{
				Enabled:   true,
				Secret:    "",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &multipassMockAPIClient{
				enableMultipassResp: tt.mockResp,
				enableMultipassErr:  tt.mockErr,
			}
			cleanup, _ := setupMultipassMockFactories(mockClient)
			defer cleanup()

			cmd := newMultipassTestCmd()

			err := multipassEnableCmd.RunE(cmd, []string{})

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

// TestMultipassEnableRunEWithJSON tests JSON output format for enable.
func TestMultipassEnableRunEWithJSON(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		enableMultipassResp: &api.Multipass{
			Enabled:   true,
			Secret:    "secret_xyz",
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := multipassEnableCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "secret") {
		t.Errorf("JSON output should contain 'secret', got: %s", output)
	}
}

func TestMultipassSecretGetRunE_WithMockAPI(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		getMultipassSecretResp: json.RawMessage(`{"secret":"s"}`),
	}
	cleanup, _ := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")

	if err := multipassSecretGetCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultipassSecretCreateRunE_NoBody_WithMockAPI(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		createMultipassSecretResp: json.RawMessage(`{"secret":"new"}`),
	}
	cleanup, _ := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmdWithBodyFlags()
	_ = cmd.Flags().Set("output", "json")

	if err := multipassSecretCreateCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultipassLinkingsListRunE_WithMockAPI(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		listMultipassLinkingsResp: json.RawMessage(`{"items":[]}`),
	}
	cleanup, _ := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().StringSlice("customer-id", nil, "")
	_ = cmd.Flags().Set("customer-id", "cust_1")

	if err := multipassLinkingsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultipassCustomersLinkRunE_WithMockAPI(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		updateCustomerLinkingResp: json.RawMessage(`{"updated":true}`),
	}
	cleanup, _ := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmdWithBodyFlags()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"ok":true}`)

	if err := multipassCustomersLinkCmd.RunE(cmd, []string{"cust_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMultipassCustomersUnlinkRunE_WithMockAPI(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		deleteCustomerLinkingResp: json.RawMessage(`{"deleted":true}`),
	}
	cleanup, _ := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("yes", "true")

	if err := multipassCustomersUnlinkCmd.RunE(cmd, []string{"cust_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestMultipassDisableRunE tests the multipass disable command with mock API.
func TestMultipassDisableRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful disable",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &multipassMockAPIClient{
				disableMultipassErr: tt.mockErr,
			}
			cleanup, _ := setupMultipassMockFactories(mockClient)
			defer cleanup()

			cmd := newMultipassTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := multipassDisableCmd.RunE(cmd, []string{})

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

// TestMultipassRotateRunE tests the multipass rotate command with mock API.
func TestMultipassRotateRunE(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		mockResp *api.Multipass
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful rotate with new secret",
			mockResp: &api.Multipass{
				Enabled:   true,
				Secret:    "new_secret_xyz789",
				CreatedAt: baseTime,
				UpdatedAt: baseTime.Add(24 * time.Hour),
			},
		},
		{
			name: "successful rotate without secret",
			mockResp: &api.Multipass{
				Enabled:   true,
				Secret:    "",
				CreatedAt: baseTime,
				UpdatedAt: baseTime,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &multipassMockAPIClient{
				rotateMultipassSecretResp: tt.mockResp,
				rotateMultipassSecretErr:  tt.mockErr,
			}
			cleanup, _ := setupMultipassMockFactories(mockClient)
			defer cleanup()

			cmd := newMultipassTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := multipassRotateCmd.RunE(cmd, []string{})

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

// TestMultipassRotateRunEWithJSON tests JSON output format for rotate.
func TestMultipassRotateRunEWithJSON(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		rotateMultipassSecretResp: &api.Multipass{
			Enabled:   true,
			Secret:    "rotated_secret_abc",
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("yes", "true")

	err := multipassRotateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "rotated_secret_abc") {
		t.Errorf("JSON output should contain secret, got: %s", output)
	}
}

// TestMultipassTokenRunE tests the multipass token command with mock API.
func TestMultipassTokenRunE(t *testing.T) {
	expiresAt := time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		email    string
		returnTo string
		mockResp *api.MultipassToken
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful token generation",
			email: "test@example.com",
			mockResp: &api.MultipassToken{
				Token:     "mp_token_abc123",
				URL:       "https://store.myshopline.com/account/login/multipass/mp_token_abc123",
				ExpiresAt: expiresAt,
			},
		},
		{
			name:     "successful token with return-to",
			email:    "user@example.com",
			returnTo: "/products",
			mockResp: &api.MultipassToken{
				Token:     "mp_token_xyz789",
				URL:       "https://store.myshopline.com/account/login/multipass/mp_token_xyz789?return_to=/products",
				ExpiresAt: expiresAt,
			},
		},
		{
			name:    "API error",
			email:   "error@example.com",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &multipassMockAPIClient{
				generateMultipassTokenResp: tt.mockResp,
				generateMultipassTokenErr:  tt.mockErr,
			}
			cleanup, _ := setupMultipassMockFactories(mockClient)
			defer cleanup()

			cmd := newMultipassTestCmd()
			cmd.Flags().String("email", tt.email, "")
			cmd.Flags().String("return-to", tt.returnTo, "")

			err := multipassTokenCmd.RunE(cmd, []string{})

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

// TestMultipassTokenRunEWithJSON tests JSON output format for token.
func TestMultipassTokenRunEWithJSON(t *testing.T) {
	mockClient := &multipassMockAPIClient{
		generateMultipassTokenResp: &api.MultipassToken{
			Token:     "json_token_abc",
			URL:       "https://store.myshopline.com/multipass/json_token_abc",
			ExpiresAt: time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupMultipassMockFactories(mockClient)
	defer cleanup()

	cmd := newMultipassTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("email", "json@example.com", "")
	cmd.Flags().String("return-to", "", "")

	err := multipassTokenCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "json_token_abc") {
		t.Errorf("JSON output should contain token, got: %s", output)
	}
}

// TestMultipassStatusNoProfiles verifies status command error handling when no profiles exist.
func TestMultipassStatusNoProfiles(t *testing.T) {
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
	err := multipassStatusCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestMultipassEnableNoProfiles verifies enable command error handling when no profiles exist.
func TestMultipassEnableNoProfiles(t *testing.T) {
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
	err := multipassEnableCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestMultipassDisableNoProfiles verifies disable command error handling when no profiles exist.
func TestMultipassDisableNoProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true")
	err := multipassDisableCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestMultipassRotateNoProfiles verifies rotate command error handling when no profiles exist.
func TestMultipassRotateNoProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true")
	err := multipassRotateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestMultipassTokenNoProfiles verifies token command error handling when no profiles exist.
func TestMultipassTokenNoProfiles(t *testing.T) {
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
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("return-to", "", "")
	err := multipassTokenCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestMultipassStatusMultipleProfiles verifies status command error handling when multiple profiles exist.
func TestMultipassStatusMultipleProfiles(t *testing.T) {
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
	err := multipassStatusCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for multiple profiles")
	}
}

// TestMultipassTokenMultipleProfiles verifies token command error handling when multiple profiles exist.
func TestMultipassTokenMultipleProfiles(t *testing.T) {
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
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("return-to", "", "")
	err := multipassTokenCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for multiple profiles")
	}
}

// TestMultipassWithMockStore tests multipass commands with a mock credential store.
func TestMultipassWithMockStore(t *testing.T) {
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

// TestMultipassErrorMessages verifies proper error wrapping for API failures.
func TestMultipassErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func() *multipassMockAPIClient
		runCmd        func(cmd *cobra.Command) error
		wantErrSubstr string
	}{
		{
			name: "get multipass failure",
			setupMock: func() *multipassMockAPIClient {
				return &multipassMockAPIClient{
					getMultipassErr: errors.New("connection refused"),
				}
			},
			runCmd: func(cmd *cobra.Command) error {
				return multipassStatusCmd.RunE(cmd, []string{})
			},
			wantErrSubstr: "failed to get multipass status",
		},
		{
			name: "enable multipass failure",
			setupMock: func() *multipassMockAPIClient {
				return &multipassMockAPIClient{
					enableMultipassErr: errors.New("permission denied"),
				}
			},
			runCmd: func(cmd *cobra.Command) error {
				return multipassEnableCmd.RunE(cmd, []string{})
			},
			wantErrSubstr: "failed to enable multipass",
		},
		{
			name: "disable multipass failure",
			setupMock: func() *multipassMockAPIClient {
				return &multipassMockAPIClient{
					disableMultipassErr: errors.New("not found"),
				}
			},
			runCmd: func(cmd *cobra.Command) error {
				_ = cmd.Flags().Set("yes", "true")
				return multipassDisableCmd.RunE(cmd, []string{})
			},
			wantErrSubstr: "failed to disable multipass",
		},
		{
			name: "rotate multipass failure",
			setupMock: func() *multipassMockAPIClient {
				return &multipassMockAPIClient{
					rotateMultipassSecretErr: errors.New("rate limited"),
				}
			},
			runCmd: func(cmd *cobra.Command) error {
				_ = cmd.Flags().Set("yes", "true")
				return multipassRotateCmd.RunE(cmd, []string{})
			},
			wantErrSubstr: "failed to rotate multipass secret",
		},
		{
			name: "generate token failure",
			setupMock: func() *multipassMockAPIClient {
				return &multipassMockAPIClient{
					generateMultipassTokenErr: errors.New("invalid email"),
				}
			},
			runCmd: func(cmd *cobra.Command) error {
				cmd.Flags().String("email", "bad@email", "")
				cmd.Flags().String("return-to", "", "")
				return multipassTokenCmd.RunE(cmd, []string{})
			},
			wantErrSubstr: "failed to generate multipass token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.setupMock()
			cleanup, _ := setupMultipassMockFactories(mockClient)
			defer cleanup()

			cmd := newMultipassTestCmd()
			err := tt.runCmd(cmd)

			if err == nil {
				t.Error("expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrSubstr) {
				t.Errorf("error %q should contain %q", err.Error(), tt.wantErrSubstr)
			}
		})
	}
}

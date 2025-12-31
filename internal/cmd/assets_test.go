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

func TestAssetsCommandSetup(t *testing.T) {
	if assetsCmd.Use != "assets" {
		t.Errorf("expected Use 'assets', got %q", assetsCmd.Use)
	}
	if assetsCmd.Short != "Manage theme assets" {
		t.Errorf("expected Short 'Manage theme assets', got %q", assetsCmd.Short)
	}
}

func TestAssetsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List assets for a theme",
		"get":    "Get asset details",
		"put":    "Create or update an asset",
		"delete": "Delete an asset",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range assetsCmd.Commands() {
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

func TestAssetsListGetClientError(t *testing.T) {
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

func TestAssetsListFlags(t *testing.T) {
	flag := assetsListCmd.Flags().Lookup("theme-id")
	if flag == nil {
		t.Error("flag 'theme-id' not found")
		return
	}
	if flag.DefValue != "" {
		t.Errorf("expected default '', got %q", flag.DefValue)
	}
}

func TestAssetsGetFlags(t *testing.T) {
	flags := []string{"theme-id", "key"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := assetsGetCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

func TestAssetsPutFlags(t *testing.T) {
	flags := []string{"theme-id", "key", "value"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := assetsPutCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

func TestAssetsDeleteFlags(t *testing.T) {
	flags := []string{"theme-id", "key"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := assetsDeleteCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

func TestAssetsWithMockStore(t *testing.T) {
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

// assetsMockAPIClient is a mock implementation of api.APIClient for assets tests.
type assetsMockAPIClient struct {
	api.MockClient
	listAssetsResp  *api.AssetsListResponse
	listAssetsErr   error
	getAssetResp    *api.Asset
	getAssetErr     error
	updateAssetResp *api.Asset
	updateAssetErr  error
	deleteAssetErr  error
}

func (m *assetsMockAPIClient) ListAssets(ctx context.Context, themeID string) (*api.AssetsListResponse, error) {
	return m.listAssetsResp, m.listAssetsErr
}

func (m *assetsMockAPIClient) GetAsset(ctx context.Context, themeID, key string) (*api.Asset, error) {
	return m.getAssetResp, m.getAssetErr
}

func (m *assetsMockAPIClient) UpdateAsset(ctx context.Context, themeID string, req *api.AssetUpdateRequest) (*api.Asset, error) {
	return m.updateAssetResp, m.updateAssetErr
}

func (m *assetsMockAPIClient) DeleteAsset(ctx context.Context, themeID, key string) error {
	return m.deleteAssetErr
}

// setupAssetsMockFactories sets up mock factories for assets tests.
func setupAssetsMockFactories(mockClient *assetsMockAPIClient) (func(), *bytes.Buffer) {
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

// newAssetsTestCmd creates a test command with common flags.
func newAssetsTestCmd() *cobra.Command {
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

// TestAssetsListRunE tests the assets list command with mock API.
func TestAssetsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		themeID    string
		mockResp   *api.AssetsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:    "successful list",
			themeID: "theme_123",
			mockResp: &api.AssetsListResponse{
				Items: []api.Asset{
					{
						Key:         "templates/index.liquid",
						ContentType: "text/x-liquid",
						Size:        1024,
						UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						Key:         "assets/style.css",
						ContentType: "text/css",
						Size:        512,
						UpdatedAt:   time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
					},
				},
			},
			wantOutput: "templates/index.liquid",
		},
		{
			name:    "API error",
			themeID: "theme_123",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name:    "empty list",
			themeID: "theme_123",
			mockResp: &api.AssetsListResponse{
				Items: []api.Asset{},
			},
		},
		{
			name:    "multiple assets with different types",
			themeID: "theme_456",
			mockResp: &api.AssetsListResponse{
				Items: []api.Asset{
					{
						Key:         "config/settings_schema.json",
						ContentType: "application/json",
						Size:        2048,
						UpdatedAt:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						Key:         "assets/logo.png",
						ContentType: "image/png",
						Size:        4096,
						UpdatedAt:   time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			wantOutput: "config/settings_schema.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &assetsMockAPIClient{
				listAssetsResp: tt.mockResp,
				listAssetsErr:  tt.mockErr,
			}
			cleanup, buf := setupAssetsMockFactories(mockClient)
			defer cleanup()

			cmd := newAssetsTestCmd()
			cmd.Flags().String("theme-id", tt.themeID, "")

			err := assetsListCmd.RunE(cmd, []string{})

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

// TestAssetsListRunEWithJSON tests JSON output format.
func TestAssetsListRunEWithJSON(t *testing.T) {
	mockClient := &assetsMockAPIClient{
		listAssetsResp: &api.AssetsListResponse{
			Items: []api.Asset{
				{Key: "templates/product.liquid", ContentType: "text/x-liquid", Size: 768},
			},
		},
	}
	cleanup, buf := setupAssetsMockFactories(mockClient)
	defer cleanup()

	cmd := newAssetsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("theme-id", "theme_json", "")

	err := assetsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "templates/product.liquid") {
		t.Errorf("JSON output should contain asset key, got: %s", output)
	}
}

// TestAssetsGetRunE tests the assets get command with mock API.
func TestAssetsGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		themeID    string
		key        string
		mockResp   *api.Asset
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:    "successful get",
			themeID: "theme_123",
			key:     "templates/index.liquid",
			mockResp: &api.Asset{
				Key:         "templates/index.liquid",
				ThemeID:     "theme_123",
				ContentType: "text/x-liquid",
				Size:        1024,
				Checksum:    "abc123",
				Value:       "<h1>Welcome</h1>",
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "asset not found",
			themeID: "theme_123",
			key:     "templates/nonexistent.liquid",
			mockErr: errors.New("asset not found"),
			wantErr: true,
		},
		{
			name:    "asset with public URL",
			themeID: "theme_456",
			key:     "assets/logo.png",
			mockResp: &api.Asset{
				Key:         "assets/logo.png",
				ThemeID:     "theme_456",
				ContentType: "image/png",
				Size:        4096,
				Checksum:    "def456",
				PublicURL:   "https://cdn.shopline.com/assets/logo.png",
				CreatedAt:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 2, 10, 8, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "asset without value",
			themeID: "theme_789",
			key:     "config/settings_data.json",
			mockResp: &api.Asset{
				Key:         "config/settings_data.json",
				ThemeID:     "theme_789",
				ContentType: "application/json",
				Size:        2048,
				Checksum:    "ghi789",
				CreatedAt:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 3, 5, 14, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &assetsMockAPIClient{
				getAssetResp: tt.mockResp,
				getAssetErr:  tt.mockErr,
			}
			cleanup, _ := setupAssetsMockFactories(mockClient)
			defer cleanup()

			cmd := newAssetsTestCmd()
			cmd.Flags().String("theme-id", tt.themeID, "")
			cmd.Flags().String("key", tt.key, "")

			err := assetsGetCmd.RunE(cmd, []string{})

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

// TestAssetsGetRunEWithJSON tests JSON output format for get command.
func TestAssetsGetRunEWithJSON(t *testing.T) {
	mockClient := &assetsMockAPIClient{
		getAssetResp: &api.Asset{
			Key:         "templates/collection.liquid",
			ThemeID:     "theme_json",
			ContentType: "text/x-liquid",
			Size:        512,
		},
	}
	cleanup, buf := setupAssetsMockFactories(mockClient)
	defer cleanup()

	cmd := newAssetsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("theme-id", "theme_json", "")
	cmd.Flags().String("key", "templates/collection.liquid", "")

	err := assetsGetCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "templates/collection.liquid") {
		t.Errorf("JSON output should contain asset key, got: %s", output)
	}
}

// TestAssetsPutRunE tests the assets put command with mock API.
func TestAssetsPutRunE(t *testing.T) {
	tests := []struct {
		name     string
		themeID  string
		key      string
		value    string
		mockResp *api.Asset
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful create",
			themeID: "theme_123",
			key:     "templates/new.liquid",
			value:   "<h1>New Template</h1>",
			mockResp: &api.Asset{
				Key:         "templates/new.liquid",
				ThemeID:     "theme_123",
				ContentType: "text/x-liquid",
				Size:        21,
			},
		},
		{
			name:    "successful update",
			themeID: "theme_123",
			key:     "templates/index.liquid",
			value:   "<h1>Updated</h1>",
			mockResp: &api.Asset{
				Key:         "templates/index.liquid",
				ThemeID:     "theme_123",
				ContentType: "text/x-liquid",
				Size:        16,
			},
		},
		{
			name:    "update fails",
			themeID: "theme_456",
			key:     "templates/protected.liquid",
			value:   "<h1>Test</h1>",
			mockErr: errors.New("asset is read-only"),
			wantErr: true,
		},
		{
			name:    "empty value",
			themeID: "theme_789",
			key:     "templates/empty.liquid",
			value:   "",
			mockResp: &api.Asset{
				Key:         "templates/empty.liquid",
				ThemeID:     "theme_789",
				ContentType: "text/x-liquid",
				Size:        0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &assetsMockAPIClient{
				updateAssetResp: tt.mockResp,
				updateAssetErr:  tt.mockErr,
			}
			cleanup, _ := setupAssetsMockFactories(mockClient)
			defer cleanup()

			cmd := newAssetsTestCmd()
			cmd.Flags().String("theme-id", tt.themeID, "")
			cmd.Flags().String("key", tt.key, "")
			cmd.Flags().String("value", tt.value, "")

			err := assetsPutCmd.RunE(cmd, []string{})

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

// TestAssetsPutRunEWithJSON tests JSON output format for put command.
func TestAssetsPutRunEWithJSON(t *testing.T) {
	mockClient := &assetsMockAPIClient{
		updateAssetResp: &api.Asset{
			Key:         "templates/json-test.liquid",
			ThemeID:     "theme_json",
			ContentType: "text/x-liquid",
			Size:        100,
		},
	}
	cleanup, buf := setupAssetsMockFactories(mockClient)
	defer cleanup()

	cmd := newAssetsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("theme-id", "theme_json", "")
	cmd.Flags().String("key", "templates/json-test.liquid", "")
	cmd.Flags().String("value", "<h1>Test</h1>", "")

	err := assetsPutCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "templates/json-test.liquid") {
		t.Errorf("JSON output should contain asset key, got: %s", output)
	}
}

// TestAssetsPutRunEDryRun tests the dry-run mode of put command.
func TestAssetsPutRunEDryRun(t *testing.T) {
	mockClient := &assetsMockAPIClient{}
	cleanup, _ := setupAssetsMockFactories(mockClient)
	defer cleanup()

	cmd := newAssetsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("theme-id", "theme_123", "")
	cmd.Flags().String("key", "templates/test.liquid", "")
	cmd.Flags().String("value", "<h1>Test</h1>", "")

	err := assetsPutCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAssetsDeleteRunE tests the assets delete command with mock API.
func TestAssetsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		themeID string
		key     string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			themeID: "theme_123",
			key:     "templates/old.liquid",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			themeID: "theme_456",
			key:     "templates/protected.liquid",
			mockErr: errors.New("asset is protected"),
			wantErr: true,
		},
		{
			name:    "asset not found",
			themeID: "theme_789",
			key:     "templates/nonexistent.liquid",
			mockErr: errors.New("asset not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &assetsMockAPIClient{
				deleteAssetErr: tt.mockErr,
			}
			cleanup, _ := setupAssetsMockFactories(mockClient)
			defer cleanup()

			cmd := newAssetsTestCmd()
			cmd.Flags().String("theme-id", tt.themeID, "")
			cmd.Flags().String("key", tt.key, "")

			err := assetsDeleteCmd.RunE(cmd, []string{})

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

// TestAssetsDeleteRunEDryRun tests the dry-run mode of delete command.
func TestAssetsDeleteRunEDryRun(t *testing.T) {
	mockClient := &assetsMockAPIClient{}
	cleanup, _ := setupAssetsMockFactories(mockClient)
	defer cleanup()

	cmd := newAssetsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")
	cmd.Flags().String("theme-id", "theme_123", "")
	cmd.Flags().String("key", "templates/test.liquid", "")

	err := assetsDeleteCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestAssetsDeleteRunEWithoutConfirmation tests delete command without --yes flag.
func TestAssetsDeleteRunEWithoutConfirmation(t *testing.T) {
	mockClient := &assetsMockAPIClient{}
	cleanup, _ := setupAssetsMockFactories(mockClient)
	defer cleanup()

	cmd := newAssetsTestCmd()
	_ = cmd.Flags().Set("yes", "false")
	cmd.Flags().String("theme-id", "theme_123", "")
	cmd.Flags().String("key", "templates/test.liquid", "")

	err := assetsDeleteCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Should return nil without actually deleting (prompts for confirmation)
}

// TestAssetsListRunEGetClientError tests error when getClient fails.
func TestAssetsListRunEGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() {
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store unavailable")
	}

	cmd := newAssetsTestCmd()
	cmd.Flags().String("theme-id", "theme_123", "")

	err := assetsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestAssetsGetRunEGetClientError tests error when getClient fails.
func TestAssetsGetRunEGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() {
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store unavailable")
	}

	cmd := newAssetsTestCmd()
	cmd.Flags().String("theme-id", "theme_123", "")
	cmd.Flags().String("key", "templates/index.liquid", "")

	err := assetsGetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestAssetsPutRunEGetClientError tests error when getClient fails.
func TestAssetsPutRunEGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() {
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store unavailable")
	}

	cmd := newAssetsTestCmd()
	cmd.Flags().String("theme-id", "theme_123", "")
	cmd.Flags().String("key", "templates/index.liquid", "")
	cmd.Flags().String("value", "<h1>Test</h1>", "")

	err := assetsPutCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestAssetsDeleteRunEGetClientError tests error when getClient fails.
func TestAssetsDeleteRunEGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() {
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("credential store unavailable")
	}

	cmd := newAssetsTestCmd()
	cmd.Flags().String("theme-id", "theme_123", "")
	cmd.Flags().String("key", "templates/index.liquid", "")

	err := assetsDeleteCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

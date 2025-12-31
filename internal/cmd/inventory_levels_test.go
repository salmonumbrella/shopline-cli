package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// inventoryLevelsAPIClient is a mock implementation of api.APIClient for inventory-levels tests.
type inventoryLevelsAPIClient struct {
	api.MockClient

	listInventoryLevelsResp  *api.InventoryListResponse
	listInventoryLevelsErr   error
	getInventoryLevelResp    *api.InventoryLevel
	getInventoryLevelErr     error
	adjustInventoryLevelResp *api.InventoryLevel
	adjustInventoryLevelErr  error
	setInventoryLevelResp    *api.InventoryLevel
	setInventoryLevelErr     error
}

func (m *inventoryLevelsAPIClient) ListInventoryLevels(ctx context.Context, opts *api.InventoryListOptions) (*api.InventoryListResponse, error) {
	return m.listInventoryLevelsResp, m.listInventoryLevelsErr
}

func (m *inventoryLevelsAPIClient) GetInventoryLevel(ctx context.Context, id string) (*api.InventoryLevel, error) {
	return m.getInventoryLevelResp, m.getInventoryLevelErr
}

func (m *inventoryLevelsAPIClient) AdjustInventoryLevel(ctx context.Context, req *api.InventoryLevelAdjustRequest) (*api.InventoryLevel, error) {
	return m.adjustInventoryLevelResp, m.adjustInventoryLevelErr
}

func (m *inventoryLevelsAPIClient) SetInventoryLevel(ctx context.Context, req *api.InventoryLevelSetRequest) (*api.InventoryLevel, error) {
	return m.setInventoryLevelResp, m.setInventoryLevelErr
}

// TestInventoryLevelsCommandSetup verifies inventory-levels command initialization
func TestInventoryLevelsCommandSetup(t *testing.T) {
	if inventoryLevelsCmd.Use != "inventory-levels" {
		t.Errorf("expected Use 'inventory-levels', got %q", inventoryLevelsCmd.Use)
	}
	if inventoryLevelsCmd.Short != "Manage inventory levels" {
		t.Errorf("expected Short 'Manage inventory levels', got %q", inventoryLevelsCmd.Short)
	}
}

// TestInventoryLevelsSubcommands verifies all subcommands are registered
func TestInventoryLevelsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List inventory levels",
		"get":    "Get inventory level details",
		"adjust": "Adjust inventory level",
		"set":    "Set inventory level",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range inventoryLevelsCmd.Commands() {
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

// TestInventoryLevelsListFlags verifies list command flags exist with correct defaults
func TestInventoryLevelsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"location-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := inventoryLevelsListCmd.Flags().Lookup(f.name)
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

// TestInventoryLevelsAdjustFlags verifies adjust command flags exist
func TestInventoryLevelsAdjustFlags(t *testing.T) {
	flags := []string{"inventory-item-id", "location-id", "adjustment"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := inventoryLevelsAdjustCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

// TestInventoryLevelsSetFlags verifies set command flags exist
func TestInventoryLevelsSetFlags(t *testing.T) {
	flags := []string{"inventory-item-id", "location-id", "available"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := inventoryLevelsSetCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

// TestInventoryLevelsGetCmdUse verifies the get command has correct use string
func TestInventoryLevelsGetCmdUse(t *testing.T) {
	if inventoryLevelsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", inventoryLevelsGetCmd.Use)
	}
}

// TestInventoryLevelsListRunE_Success tests the inventory-levels list command execution with mock API.
func TestInventoryLevelsListRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name     string
		mockResp *api.InventoryListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list",
			mockResp: &api.InventoryListResponse{
				Items: []api.InventoryLevel{
					{
						ID:              "inv_level_123",
						InventoryItemID: "item_456",
						LocationID:      "loc_789",
						Available:       100,
						Reserved:        10,
						Incoming:        5,
						OnHand:          110,
						UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.InventoryListResponse{
				Items:      []api.InventoryLevel{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &inventoryLevelsAPIClient{
				listInventoryLevelsResp: tt.mockResp,
				listInventoryLevelsErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("location-id", "", "")

			err := inventoryLevelsListCmd.RunE(cmd, []string{})

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

// TestInventoryLevelsGetRunE_Success tests the inventory-levels get command execution with mock API.
func TestInventoryLevelsGetRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name     string
		id       string
		mockResp *api.InventoryLevel
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "inv_level_123",
			mockResp: &api.InventoryLevel{
				ID:              "inv_level_123",
				InventoryItemID: "item_456",
				LocationID:      "loc_789",
				Available:       100,
				Reserved:        10,
				Incoming:        5,
				OnHand:          110,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "inv_level_999",
			mockErr: errors.New("inventory level not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &inventoryLevelsAPIClient{
				getInventoryLevelResp: tt.mockResp,
				getInventoryLevelErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := inventoryLevelsGetCmd.RunE(cmd, []string{tt.id})

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

// TestInventoryLevelsAdjustRunE_Success tests the inventory-levels adjust command execution with mock API.
func TestInventoryLevelsAdjustRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name     string
		mockResp *api.InventoryLevel
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful adjust",
			mockResp: &api.InventoryLevel{
				ID:              "inv_level_123",
				InventoryItemID: "item_456",
				LocationID:      "loc_789",
				Available:       105,
				Reserved:        10,
				Incoming:        5,
				OnHand:          115,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "adjust fails",
			mockErr: errors.New("insufficient inventory"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &inventoryLevelsAPIClient{
				adjustInventoryLevelResp: tt.mockResp,
				adjustInventoryLevelErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("inventory-item-id", "item_456", "")
			cmd.Flags().String("location-id", "loc_789", "")
			cmd.Flags().Int("adjustment", 5, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := inventoryLevelsAdjustCmd.RunE(cmd, []string{})

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

// TestInventoryLevelsSetRunE_Success tests the inventory-levels set command execution with mock API.
func TestInventoryLevelsSetRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name     string
		mockResp *api.InventoryLevel
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful set",
			mockResp: &api.InventoryLevel{
				ID:              "inv_level_123",
				InventoryItemID: "item_456",
				LocationID:      "loc_789",
				Available:       50,
				Reserved:        10,
				Incoming:        5,
				OnHand:          60,
				UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "set fails",
			mockErr: errors.New("invalid quantity"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &inventoryLevelsAPIClient{
				setInventoryLevelResp: tt.mockResp,
				setInventoryLevelErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("inventory-item-id", "item_456", "")
			cmd.Flags().String("location-id", "loc_789", "")
			cmd.Flags().Int("available", 50, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := inventoryLevelsSetCmd.RunE(cmd, []string{})

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

// TestInventoryLevelsListRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryLevelsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := inventoryLevelsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryLevelsGetRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryLevelsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := inventoryLevelsGetCmd.RunE(cmd, []string{"inv_level_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryLevelsAdjustRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryLevelsAdjustRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("inventory-item-id", "", "")
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().Int("adjustment", 0, "")
	_ = cmd.Flags().Set("inventory-item-id", "item_123")
	_ = cmd.Flags().Set("location-id", "loc_123")
	_ = cmd.Flags().Set("adjustment", "5")
	err := inventoryLevelsAdjustCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryLevelsSetRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryLevelsSetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("inventory-item-id", "", "")
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().Int("available", 0, "")
	_ = cmd.Flags().Set("inventory-item-id", "item_123")
	_ = cmd.Flags().Set("location-id", "loc_123")
	_ = cmd.Flags().Set("available", "10")
	err := inventoryLevelsSetCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryLevelsListRunE_NoProfiles verifies error when no profiles are configured
func TestInventoryLevelsListRunE_NoProfiles(t *testing.T) {
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
	err := inventoryLevelsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestInventoryLevelsGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestInventoryLevelsGetRunE_MultipleProfiles(t *testing.T) {
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
	err := inventoryLevelsGetCmd.RunE(cmd, []string{"inv_level_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

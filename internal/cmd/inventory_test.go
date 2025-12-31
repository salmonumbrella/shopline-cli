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

// inventoryAPIClient is a mock implementation of api.APIClient for inventory tests.
type inventoryAPIClient struct {
	api.MockClient

	listInventoryLevelsResp *api.InventoryListResponse
	listInventoryLevelsErr  error
	getInventoryLevelResp   *api.InventoryLevel
	getInventoryLevelErr    error
	adjustInventoryResp     *api.InventoryLevel
	adjustInventoryErr      error
}

func (m *inventoryAPIClient) ListInventoryLevels(ctx context.Context, opts *api.InventoryListOptions) (*api.InventoryListResponse, error) {
	return m.listInventoryLevelsResp, m.listInventoryLevelsErr
}

func (m *inventoryAPIClient) GetInventoryLevel(ctx context.Context, id string) (*api.InventoryLevel, error) {
	return m.getInventoryLevelResp, m.getInventoryLevelErr
}

func (m *inventoryAPIClient) AdjustInventory(ctx context.Context, id string, delta int) (*api.InventoryLevel, error) {
	return m.adjustInventoryResp, m.adjustInventoryErr
}

// TestInventoryCommandSetup verifies inventory command initialization
func TestInventoryCommandSetup(t *testing.T) {
	if inventoryCmd.Use != "inventory" {
		t.Errorf("expected Use 'inventory', got %q", inventoryCmd.Use)
	}
	if inventoryCmd.Short != "Manage inventory levels" {
		t.Errorf("expected Short 'Manage inventory levels', got %q", inventoryCmd.Short)
	}
}

// TestInventorySubcommands verifies all subcommands are registered
func TestInventorySubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List inventory levels",
		"get":    "Get inventory level details",
		"adjust": "Adjust inventory quantity",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range inventoryCmd.Commands() {
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

// TestInventoryListFlags verifies list command flags exist with correct defaults
func TestInventoryListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"location-id", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := inventoryListCmd.Flags().Lookup(f.name)
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

// TestInventoryAdjustFlags verifies adjust command flags exist with correct defaults
func TestInventoryAdjustFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
		required     bool
	}{
		{"delta", "0", true},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := inventoryAdjustCmd.Flags().Lookup(f.name)
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

// TestInventoryGetCmdUse verifies the get command has correct use string
func TestInventoryGetCmdUse(t *testing.T) {
	if inventoryGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", inventoryGetCmd.Use)
	}
}

// TestInventoryAdjustCmdUse verifies the adjust command has correct use string
func TestInventoryAdjustCmdUse(t *testing.T) {
	if inventoryAdjustCmd.Use != "adjust <id>" {
		t.Errorf("expected Use 'adjust <id>', got %q", inventoryAdjustCmd.Use)
	}
}

// TestInventoryListRunE_Success tests the inventory list command execution with mock API.
func TestInventoryListRunE_Success(t *testing.T) {
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
						ID:              "inv_123",
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
			mockClient := &inventoryAPIClient{
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
			cmd.Flags().String("location-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := inventoryListCmd.RunE(cmd, []string{})

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

// TestInventoryGetRunE_Success tests the inventory get command execution with mock API.
func TestInventoryGetRunE_Success(t *testing.T) {
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
			id:   "inv_123",
			mockResp: &api.InventoryLevel{
				ID:              "inv_123",
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
			id:      "inv_999",
			mockErr: errors.New("inventory level not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &inventoryAPIClient{
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

			err := inventoryGetCmd.RunE(cmd, []string{tt.id})

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

// TestInventoryAdjustRunE_Success tests the inventory adjust command execution with mock API.
func TestInventoryAdjustRunE_Success(t *testing.T) {
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
		delta    int
		mockResp *api.InventoryLevel
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful adjust",
			id:    "inv_123",
			delta: 5,
			mockResp: &api.InventoryLevel{
				ID:              "inv_123",
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
			id:      "inv_456",
			delta:   -200,
			mockErr: errors.New("insufficient inventory"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &inventoryAPIClient{
				adjustInventoryResp: tt.mockResp,
				adjustInventoryErr:  tt.mockErr,
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
			cmd.Flags().Int("delta", tt.delta, "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := inventoryAdjustCmd.RunE(cmd, []string{tt.id})

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

// TestInventoryListRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := inventoryListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryGetRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := inventoryGetCmd.RunE(cmd, []string{"inv_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryAdjustRunE_GetClientFails verifies error handling when getClient fails
func TestInventoryAdjustRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("delta", 5, "")
	_ = cmd.Flags().Set("delta", "5")
	err := inventoryAdjustCmd.RunE(cmd, []string{"inv_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestInventoryListRunE_NoProfiles verifies error when no profiles are configured
func TestInventoryListRunE_NoProfiles(t *testing.T) {
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
	err := inventoryListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestInventoryGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestInventoryGetRunE_MultipleProfiles(t *testing.T) {
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
	err := inventoryGetCmd.RunE(cmd, []string{"inv_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestInventoryListRunE_JSONOutput tests the inventory list command with JSON output format.
func TestInventoryListRunE_JSONOutput(t *testing.T) {
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

	mockClient := &inventoryAPIClient{
		listInventoryLevelsResp: &api.InventoryListResponse{
			Items: []api.InventoryLevel{
				{
					ID:              "inv_123",
					InventoryItemID: "item_456",
					LocationID:      "loc_789",
					Available:       100,
					UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := inventoryListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInventoryGetRunE_JSONOutput tests the inventory get command with JSON output format.
func TestInventoryGetRunE_JSONOutput(t *testing.T) {
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

	mockClient := &inventoryAPIClient{
		getInventoryLevelResp: &api.InventoryLevel{
			ID:              "inv_123",
			InventoryItemID: "item_456",
			LocationID:      "loc_789",
			Available:       100,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := inventoryGetCmd.RunE(cmd, []string{"inv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInventoryAdjustRunE_JSONOutput tests the inventory adjust command with JSON output format.
func TestInventoryAdjustRunE_JSONOutput(t *testing.T) {
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

	mockClient := &inventoryAPIClient{
		adjustInventoryResp: &api.InventoryLevel{
			ID:              "inv_123",
			InventoryItemID: "item_456",
			LocationID:      "loc_789",
			Available:       105,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("delta", 5, "")
	cmd.Flags().Bool("yes", true, "")
	_ = cmd.Flags().Set("output", "json")

	err := inventoryAdjustCmd.RunE(cmd, []string{"inv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInventoryAdjustRunE_ZeroDelta tests the inventory adjust command with zero delta.
func TestInventoryAdjustRunE_ZeroDelta(t *testing.T) {
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

	mockClient := &inventoryAPIClient{}
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
	cmd.Flags().Int("delta", 0, "")
	cmd.Flags().Bool("yes", true, "")

	err := inventoryAdjustCmd.RunE(cmd, []string{"inv_123"})
	if err == nil {
		t.Fatal("Expected error for zero delta, got nil")
	}
	if err.Error() != "--delta is required and must be non-zero" {
		t.Errorf("Expected delta error message, got: %v", err)
	}
}

// TestInventoryAdjustRunE_ConfirmationCancelled tests the inventory adjust command when confirmation is cancelled.
func TestInventoryAdjustRunE_ConfirmationCancelled(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	origStdin := os.Stdin
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
		os.Stdin = origStdin
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &inventoryAPIClient{}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	// Create a pipe to simulate stdin with "n\n" (cancel confirmation)
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("delta", 5, "")
	cmd.Flags().Bool("yes", false, "") // Not auto-confirmed

	err := inventoryAdjustCmd.RunE(cmd, []string{"inv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInventoryAdjustRunE_ConfirmationAccepted tests the inventory adjust command when confirmation is accepted.
func TestInventoryAdjustRunE_ConfirmationAccepted(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	origStdin := os.Stdin
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
		os.Stdin = origStdin
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &inventoryAPIClient{
		adjustInventoryResp: &api.InventoryLevel{
			ID:              "inv_123",
			InventoryItemID: "item_456",
			LocationID:      "loc_789",
			Available:       105,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	// Create a pipe to simulate stdin with "y\n" (accept confirmation)
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("y\n")
	_ = w.Close()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("delta", 5, "")
	cmd.Flags().Bool("yes", false, "") // Not auto-confirmed

	err := inventoryAdjustCmd.RunE(cmd, []string{"inv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInventoryAdjustRunE_NegativeDelta tests the inventory adjust command with negative delta (decrease).
func TestInventoryAdjustRunE_NegativeDelta(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	origStdin := os.Stdin
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
		os.Stdin = origStdin
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &inventoryAPIClient{
		adjustInventoryResp: &api.InventoryLevel{
			ID:              "inv_123",
			InventoryItemID: "item_456",
			LocationID:      "loc_789",
			Available:       95,
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	// Create a pipe to simulate stdin with "Y\n" (uppercase accept confirmation)
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("Y\n")
	_ = w.Close()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("delta", -5, "") // Negative delta for decrease
	cmd.Flags().Bool("yes", false, "")

	err := inventoryAdjustCmd.RunE(cmd, []string{"inv_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

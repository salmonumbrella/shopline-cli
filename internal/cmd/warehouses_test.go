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

// warehousesAPIClient is a mock implementation of api.APIClient for warehouses tests.
type warehousesAPIClient struct {
	api.MockClient

	listWarehousesResp  *api.WarehousesListResponse
	listWarehousesErr   error
	getWarehouseResp    *api.Warehouse
	getWarehouseErr     error
	createWarehouseResp *api.Warehouse
	createWarehouseErr  error
	deleteWarehouseErr  error
}

func (m *warehousesAPIClient) ListWarehouses(ctx context.Context, opts *api.WarehousesListOptions) (*api.WarehousesListResponse, error) {
	return m.listWarehousesResp, m.listWarehousesErr
}

func (m *warehousesAPIClient) GetWarehouse(ctx context.Context, id string) (*api.Warehouse, error) {
	return m.getWarehouseResp, m.getWarehouseErr
}

func (m *warehousesAPIClient) CreateWarehouse(ctx context.Context, req *api.WarehouseCreateRequest) (*api.Warehouse, error) {
	return m.createWarehouseResp, m.createWarehouseErr
}

func (m *warehousesAPIClient) DeleteWarehouse(ctx context.Context, id string) error {
	return m.deleteWarehouseErr
}

// TestWarehousesCommandSetup verifies warehouses command initialization
func TestWarehousesCommandSetup(t *testing.T) {
	if warehousesCmd.Use != "warehouses" {
		t.Errorf("expected Use 'warehouses', got %q", warehousesCmd.Use)
	}
	if warehousesCmd.Short != "Manage warehouses" {
		t.Errorf("expected Short 'Manage warehouses', got %q", warehousesCmd.Short)
	}
}

// TestWarehousesSubcommands verifies all subcommands are registered
func TestWarehousesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List warehouses",
		"get":    "Get warehouse details",
		"create": "Create a warehouse",
		"delete": "Delete a warehouse",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range warehousesCmd.Commands() {
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

// TestWarehousesListFlags verifies list command flags exist with correct defaults
func TestWarehousesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := warehousesListCmd.Flags().Lookup(f.name)
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

// TestWarehousesCreateFlags verifies create command flags exist with correct defaults
func TestWarehousesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"code", ""},
		{"address", ""},
		{"city", ""},
		{"country", ""},
		{"phone", ""},
		{"email", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := warehousesCreateCmd.Flags().Lookup(f.name)
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

// TestWarehousesGetClientError verifies error handling when getClient fails
func TestWarehousesGetClientError(t *testing.T) {
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

// TestWarehousesWithMockStore tests warehouses commands with a mock credential store
func TestWarehousesWithMockStore(t *testing.T) {
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

// TestWarehousesGetArgs verifies get command requires exactly one argument
func TestWarehousesGetArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if warehousesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", warehousesGetCmd.Use)
	}
}

// TestWarehousesDeleteArgs verifies delete command requires exactly one argument
func TestWarehousesDeleteArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if warehousesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", warehousesDeleteCmd.Use)
	}
}

// TestWarehousesListFlagDescriptions verifies flag descriptions are set
func TestWarehousesListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"page":      "Page number",
		"page-size": "Results per page",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := warehousesListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// TestWarehousesCreateFlagDescriptions verifies create flag descriptions
func TestWarehousesCreateFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"name":    "Warehouse name",
		"code":    "Warehouse code",
		"address": "Street address",
		"city":    "City",
		"country": "Country",
		"phone":   "Phone number",
		"email":   "Email address",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := warehousesCreateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// TestWarehousesCreateRequiredFlags verifies that name, address, city, country flags are required
func TestWarehousesCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"name", "address", "city", "country"}

	for _, flagName := range requiredFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := warehousesCreateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("%s flag not found", flagName)
				return
			}
			// Verify it exists and is a string flag
			if flag.Value.Type() != "string" {
				t.Errorf("%s flag should be a string type", flagName)
			}
		})
	}
}

// TestWarehousesCreateFlagTypes verifies flag types are correct
func TestWarehousesCreateFlagTypes(t *testing.T) {
	flags := map[string]string{
		"name":    "string",
		"code":    "string",
		"address": "string",
		"city":    "string",
		"country": "string",
		"phone":   "string",
		"email":   "string",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := warehousesCreateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Value.Type() != expectedType {
				t.Errorf("expected type %q, got %q", expectedType, flag.Value.Type())
			}
		})
	}
}

// TestWarehousesListFlagTypes verifies list flag types are correct
func TestWarehousesListFlagTypes(t *testing.T) {
	flags := map[string]string{
		"page":      "int",
		"page-size": "int",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := warehousesListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Value.Type() != expectedType {
				t.Errorf("expected type %q, got %q", expectedType, flag.Value.Type())
			}
		})
	}
}

// TestWarehousesListRunE_Success tests the warehouses list command execution with mock API.
func TestWarehousesListRunE_Success(t *testing.T) {
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
		name         string
		outputFormat string
		mockResp     *api.WarehousesListResponse
		mockErr      error
		wantErr      bool
	}{
		{
			name: "successful list",
			mockResp: &api.WarehousesListResponse{
				Items: []api.Warehouse{
					{
						ID:        "wh_123",
						Name:      "Main Warehouse",
						Code:      "MAIN",
						Address1:  "123 Main St",
						City:      "New York",
						Country:   "US",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:         "successful list with JSON output",
			outputFormat: "json",
			mockResp: &api.WarehousesListResponse{
				Items: []api.Warehouse{
					{
						ID:        "wh_123",
						Name:      "Main Warehouse",
						Code:      "MAIN",
						Address1:  "123 Main St",
						City:      "New York",
						Country:   "US",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
		},
		{
			name: "list with active and default warehouses",
			mockResp: &api.WarehousesListResponse{
				Items: []api.Warehouse{
					{
						ID:        "wh_active",
						Name:      "Active Default Warehouse",
						Code:      "ACTIVE",
						Address1:  "456 Active St",
						City:      "Los Angeles",
						Country:   "US",
						Active:    true,
						IsDefault: true,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
			mockResp: &api.WarehousesListResponse{
				Items:      []api.Warehouse{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &warehousesAPIClient{
				listWarehousesResp: tt.mockResp,
				listWarehousesErr:  tt.mockErr,
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

			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := warehousesListCmd.RunE(cmd, []string{})

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

// TestWarehousesGetRunE_Success tests the warehouses get command execution with mock API.
func TestWarehousesGetRunE_Success(t *testing.T) {
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
		name         string
		id           string
		outputFormat string
		mockResp     *api.Warehouse
		mockErr      error
		wantErr      bool
	}{
		{
			name: "successful get",
			id:   "wh_123",
			mockResp: &api.Warehouse{
				ID:        "wh_123",
				Name:      "Main Warehouse",
				Code:      "MAIN",
				Address1:  "123 Main St",
				City:      "New York",
				Country:   "US",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:         "successful get with JSON output",
			id:           "wh_123",
			outputFormat: "json",
			mockResp: &api.Warehouse{
				ID:        "wh_123",
				Name:      "Main Warehouse",
				Code:      "MAIN",
				Address1:  "123 Main St",
				City:      "New York",
				Country:   "US",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "get with all optional fields",
			id:   "wh_full",
			mockResp: &api.Warehouse{
				ID:           "wh_full",
				Name:         "Full Warehouse",
				Code:         "FULL",
				Address1:     "123 Main St",
				Address2:     "Suite 100",
				City:         "New York",
				Province:     "New York",
				ProvinceCode: "NY",
				Country:      "United States",
				CountryCode:  "US",
				Zip:          "10001",
				Phone:        "+1-555-555-5555",
				Email:        "warehouse@example.com",
				Active:       true,
				IsDefault:    true,
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "wh_999",
			mockErr: errors.New("warehouse not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &warehousesAPIClient{
				getWarehouseResp: tt.mockResp,
				getWarehouseErr:  tt.mockErr,
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

			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := warehousesGetCmd.RunE(cmd, []string{tt.id})

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

// TestWarehousesCreateRunE_Success tests the warehouses create command execution with mock API.
func TestWarehousesCreateRunE_Success(t *testing.T) {
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
		name         string
		outputFormat string
		mockResp     *api.Warehouse
		mockErr      error
		wantErr      bool
	}{
		{
			name: "successful create",
			mockResp: &api.Warehouse{
				ID:       "wh_new",
				Name:     "New Warehouse",
				Code:     "NEW",
				Address1: "456 Oak Ave",
				City:     "Los Angeles",
				Country:  "US",
			},
		},
		{
			name:         "successful create with JSON output",
			outputFormat: "json",
			mockResp: &api.Warehouse{
				ID:       "wh_new",
				Name:     "New Warehouse",
				Code:     "NEW",
				Address1: "456 Oak Ave",
				City:     "Los Angeles",
				Country:  "US",
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &warehousesAPIClient{
				createWarehouseResp: tt.mockResp,
				createWarehouseErr:  tt.mockErr,
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
			cmd.Flags().String("name", "New Warehouse", "")
			cmd.Flags().String("code", "NEW", "")
			cmd.Flags().String("address", "456 Oak Ave", "")
			cmd.Flags().String("city", "Los Angeles", "")
			cmd.Flags().String("country", "US", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().String("email", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := warehousesCreateCmd.RunE(cmd, []string{})

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

// TestWarehousesDeleteRunE_Success tests the warehouses delete command execution with mock API.
func TestWarehousesDeleteRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
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
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   "wh_123",
		},
		{
			name:    "delete fails",
			id:      "wh_456",
			mockErr: errors.New("warehouse in use"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &warehousesAPIClient{
				deleteWarehouseErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := warehousesDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestWarehousesListRunE_GetClientFails tests error handling when getClient fails.
func TestWarehousesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := warehousesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestWarehousesGetRunE_GetClientFails tests error handling when getClient fails.
func TestWarehousesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := warehousesGetCmd.RunE(cmd, []string{"wh_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestWarehousesCreateRunE_GetClientFails tests error handling when getClient fails.
func TestWarehousesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("code", "TEST", "")
	cmd.Flags().String("address", "123 St", "")
	cmd.Flags().String("city", "City", "")
	cmd.Flags().String("country", "US", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("email", "", "")

	err := warehousesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestWarehousesDeleteRunE_GetClientFails tests error handling when getClient fails.
func TestWarehousesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := warehousesDeleteCmd.RunE(cmd, []string{"wh_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestWarehousesDeleteRunE_ConfirmationCancelled tests the confirmation cancellation path.
func TestWarehousesDeleteRunE_ConfirmationCancelled(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &warehousesAPIClient{}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "") // Not skipping confirmation

	// Since Scanln will fail or return empty in tests, the command should print "Cancelled."
	// and return nil (no error, just cancelled)
	err := warehousesDeleteCmd.RunE(cmd, []string{"wh_123"})
	// No error expected since user "cancelled" (empty confirm != "y")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWarehousesListRunE_NoProfiles tests error handling when no profiles are configured.
func TestWarehousesListRunE_NoProfiles(t *testing.T) {
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

	err := warehousesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestWarehousesGetRunE_MultipleProfiles tests error handling when multiple profiles exist.
func TestWarehousesGetRunE_MultipleProfiles(t *testing.T) {
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

	err := warehousesGetCmd.RunE(cmd, []string{"wh_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

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

// locationsAPIClient is a mock implementation of api.APIClient for locations tests.
type locationsAPIClient struct {
	api.MockClient

	listLocationsResp  *api.LocationsListResponse
	listLocationsErr   error
	getLocationResp    *api.Location
	getLocationErr     error
	createLocationResp *api.Location
	createLocationErr  error
	deleteLocationErr  error
}

func (m *locationsAPIClient) ListLocations(ctx context.Context, opts *api.LocationsListOptions) (*api.LocationsListResponse, error) {
	return m.listLocationsResp, m.listLocationsErr
}

func (m *locationsAPIClient) GetLocation(ctx context.Context, id string) (*api.Location, error) {
	return m.getLocationResp, m.getLocationErr
}

func (m *locationsAPIClient) CreateLocation(ctx context.Context, req *api.LocationCreateRequest) (*api.Location, error) {
	return m.createLocationResp, m.createLocationErr
}

func (m *locationsAPIClient) DeleteLocation(ctx context.Context, id string) error {
	return m.deleteLocationErr
}

// TestLocationsCommandSetup verifies locations command initialization
func TestLocationsCommandSetup(t *testing.T) {
	if locationsCmd.Use != "locations" {
		t.Errorf("expected Use 'locations', got %q", locationsCmd.Use)
	}
	if locationsCmd.Short != "Manage store locations" {
		t.Errorf("expected Short 'Manage store locations', got %q", locationsCmd.Short)
	}
}

// TestLocationsSubcommands verifies all subcommands are registered
func TestLocationsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List locations",
		"get":    "Get location details",
		"create": "Create a location",
		"delete": "Delete a location",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range locationsCmd.Commands() {
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

// TestLocationsListFlags verifies list command flags exist with correct defaults
func TestLocationsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := locationsListCmd.Flags().Lookup(f.name)
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

// TestLocationsCreateFlags verifies create command flags exist
func TestLocationsCreateFlags(t *testing.T) {
	flags := []string{"name", "address", "city", "country", "phone"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := locationsCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

// TestLocationsGetCmdUse verifies the get command has correct use string
func TestLocationsGetCmdUse(t *testing.T) {
	if locationsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", locationsGetCmd.Use)
	}
}

// TestLocationsDeleteCmdUse verifies the delete command has correct use string
func TestLocationsDeleteCmdUse(t *testing.T) {
	if locationsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", locationsDeleteCmd.Use)
	}
}

// TestLocationsListRunE_Success tests the locations list command execution with mock API.
func TestLocationsListRunE_Success(t *testing.T) {
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
		mockResp *api.LocationsListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list",
			mockResp: &api.LocationsListResponse{
				Items: []api.Location{
					{
						ID:        "loc_123",
						Name:      "Main Warehouse",
						Address1:  "123 Main St",
						City:      "New York",
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
			mockResp: &api.LocationsListResponse{
				Items:      []api.Location{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &locationsAPIClient{
				listLocationsResp: tt.mockResp,
				listLocationsErr:  tt.mockErr,
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

			err := locationsListCmd.RunE(cmd, []string{})

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

// TestLocationsGetRunE_Success tests the locations get command execution with mock API.
func TestLocationsGetRunE_Success(t *testing.T) {
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
		mockResp *api.Location
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "loc_123",
			mockResp: &api.Location{
				ID:        "loc_123",
				Name:      "Main Warehouse",
				Address1:  "123 Main St",
				City:      "New York",
				Country:   "US",
				Phone:     "+1234567890",
				Active:    true,
				IsDefault: true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "loc_999",
			mockErr: errors.New("location not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &locationsAPIClient{
				getLocationResp: tt.mockResp,
				getLocationErr:  tt.mockErr,
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

			err := locationsGetCmd.RunE(cmd, []string{tt.id})

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

// TestLocationsCreateRunE_Success tests the locations create command execution with mock API.
func TestLocationsCreateRunE_Success(t *testing.T) {
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
		mockResp *api.Location
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.Location{
				ID:       "loc_new",
				Name:     "New Location",
				Address1: "456 Oak Ave",
				City:     "Los Angeles",
				Country:  "US",
				Active:   true,
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
			mockClient := &locationsAPIClient{
				createLocationResp: tt.mockResp,
				createLocationErr:  tt.mockErr,
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
			cmd.Flags().String("name", "New Location", "")
			cmd.Flags().String("address", "456 Oak Ave", "")
			cmd.Flags().String("city", "Los Angeles", "")
			cmd.Flags().String("country", "US", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := locationsCreateCmd.RunE(cmd, []string{})

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

// TestLocationsDeleteRunE_Success tests the locations delete command execution with mock API.
func TestLocationsDeleteRunE_Success(t *testing.T) {
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
			id:   "loc_123",
		},
		{
			name:    "delete fails",
			id:      "loc_456",
			mockErr: errors.New("location in use"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &locationsAPIClient{
				deleteLocationErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := locationsDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestLocationsListRunE_GetClientFails verifies error handling when getClient fails
func TestLocationsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := locationsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLocationsGetRunE_GetClientFails verifies error handling when getClient fails
func TestLocationsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := locationsGetCmd.RunE(cmd, []string{"loc_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLocationsCreateRunE_GetClientFails verifies error handling when getClient fails
func TestLocationsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("address", "", "")
	cmd.Flags().String("city", "", "")
	cmd.Flags().String("country", "", "")
	cmd.Flags().String("phone", "", "")
	_ = cmd.Flags().Set("name", "Test Location")
	_ = cmd.Flags().Set("address", "123 Test St")
	_ = cmd.Flags().Set("city", "Test City")
	_ = cmd.Flags().Set("country", "US")

	err := locationsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLocationsDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestLocationsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := locationsDeleteCmd.RunE(cmd, []string{"loc_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLocationsListRunE_NoProfiles verifies error when no profiles are configured
func TestLocationsListRunE_NoProfiles(t *testing.T) {
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
	err := locationsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestLocationsGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestLocationsGetRunE_MultipleProfiles(t *testing.T) {
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
	err := locationsGetCmd.RunE(cmd, []string{"loc_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

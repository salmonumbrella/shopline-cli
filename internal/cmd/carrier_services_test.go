package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// carrierServicesAPIClient is a mock implementation of api.APIClient for carrier services tests.
type carrierServicesAPIClient struct {
	api.MockClient

	listCarrierServicesResp  *api.CarrierServicesListResponse
	listCarrierServicesErr   error
	getCarrierServiceResp    *api.CarrierService
	getCarrierServiceErr     error
	createCarrierServiceResp *api.CarrierService
	createCarrierServiceErr  error
	deleteCarrierServiceErr  error
}

func (m *carrierServicesAPIClient) ListCarrierServices(ctx context.Context, opts *api.CarrierServicesListOptions) (*api.CarrierServicesListResponse, error) {
	return m.listCarrierServicesResp, m.listCarrierServicesErr
}

func (m *carrierServicesAPIClient) GetCarrierService(ctx context.Context, id string) (*api.CarrierService, error) {
	return m.getCarrierServiceResp, m.getCarrierServiceErr
}

func (m *carrierServicesAPIClient) CreateCarrierService(ctx context.Context, req *api.CarrierServiceCreateRequest) (*api.CarrierService, error) {
	return m.createCarrierServiceResp, m.createCarrierServiceErr
}

func (m *carrierServicesAPIClient) DeleteCarrierService(ctx context.Context, id string) error {
	return m.deleteCarrierServiceErr
}

// TestCarrierServicesCommandSetup verifies carrier-services command initialization
func TestCarrierServicesCommandSetup(t *testing.T) {
	if carrierServicesCmd.Use != "carrier-services" {
		t.Errorf("expected Use 'carrier-services', got %q", carrierServicesCmd.Use)
	}
	if carrierServicesCmd.Short != "Manage carrier services" {
		t.Errorf("expected Short 'Manage carrier services', got %q", carrierServicesCmd.Short)
	}
}

// TestCarrierServicesSubcommands verifies all subcommands are registered
func TestCarrierServicesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List carrier services",
		"get":    "Get carrier service details",
		"create": "Create a carrier service",
		"delete": "Delete a carrier service",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range carrierServicesCmd.Commands() {
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

// TestCarrierServicesListFlags verifies list command flags exist with correct defaults
func TestCarrierServicesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := carrierServicesListCmd.Flags().Lookup(f.name)
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

// TestCarrierServicesCreateFlags verifies create command flags exist with correct defaults
func TestCarrierServicesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
		required     bool
	}{
		{"name", "", true},
		{"callback-url", "", true},
		{"service-discovery", "false", false},
		{"type", "api", false},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := carrierServicesCreateCmd.Flags().Lookup(f.name)
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

// TestCarrierServicesDeleteFlags verifies delete command flags exist with correct defaults
func TestCarrierServicesDeleteFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"yes", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := carrierServicesDeleteCmd.Flags().Lookup(f.name)
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

// TestCarrierServicesGetCmd verifies get command configuration
func TestCarrierServicesGetCmd(t *testing.T) {
	if carrierServicesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", carrierServicesGetCmd.Use)
	}
	if carrierServicesGetCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestCarrierServicesDeleteCmd verifies delete command configuration
func TestCarrierServicesDeleteCmd(t *testing.T) {
	if carrierServicesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", carrierServicesDeleteCmd.Use)
	}
	if carrierServicesDeleteCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestCarrierServicesListRunE_Success tests the carrier-services list command execution with mock API.
func TestCarrierServicesListRunE_Success(t *testing.T) {
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
		name       string
		mockResp   *api.CarrierServicesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CarrierServicesListResponse{
				Items: []api.CarrierService{
					{
						ID:                 "cs_123",
						Name:               "Test Carrier",
						CarrierServiceType: "api",
						CallbackURL:        "https://example.com/rates",
						Active:             true,
						ServiceDiscovery:   false,
						CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cs_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CarrierServicesListResponse{
				Items:      []api.CarrierService{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &carrierServicesAPIClient{
				listCarrierServicesResp: tt.mockResp,
				listCarrierServicesErr:  tt.mockErr,
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

			err := carrierServicesListCmd.RunE(cmd, []string{})

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

// TestCarrierServicesGetRunE_Success tests the carrier-services get command execution with mock API.
func TestCarrierServicesGetRunE_Success(t *testing.T) {
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
		mockResp *api.CarrierService
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "cs_123",
			mockResp: &api.CarrierService{
				ID:                 "cs_123",
				Name:               "Test Carrier",
				CarrierServiceType: "api",
				CallbackURL:        "https://example.com/rates",
				Active:             true,
				ServiceDiscovery:   false,
				Format:             "json",
				CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:          time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "cs_999",
			mockErr: errors.New("carrier service not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &carrierServicesAPIClient{
				getCarrierServiceResp: tt.mockResp,
				getCarrierServiceErr:  tt.mockErr,
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

			err := carrierServicesGetCmd.RunE(cmd, []string{tt.id})

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

// TestCarrierServicesCreateRunE_Success tests the carrier-services create command execution with mock API.
func TestCarrierServicesCreateRunE_Success(t *testing.T) {
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
		mockResp *api.CarrierService
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.CarrierService{
				ID:                 "cs_new",
				Name:               "New Carrier",
				CarrierServiceType: "api",
				CallbackURL:        "https://example.com/rates",
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
			mockClient := &carrierServicesAPIClient{
				createCarrierServiceResp: tt.mockResp,
				createCarrierServiceErr:  tt.mockErr,
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
			cmd.Flags().String("name", "New Carrier", "")
			cmd.Flags().String("callback-url", "https://example.com/rates", "")
			cmd.Flags().Bool("service-discovery", false, "")
			cmd.Flags().String("type", "api", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := carrierServicesCreateCmd.RunE(cmd, []string{})

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

// TestCarrierServicesDeleteRunE_Success tests the carrier-services delete command execution with mock API.
func TestCarrierServicesDeleteRunE_Success(t *testing.T) {
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
			id:   "cs_123",
		},
		{
			name:    "delete fails",
			id:      "cs_456",
			mockErr: errors.New("carrier service in use"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &carrierServicesAPIClient{
				deleteCarrierServiceErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := carrierServicesDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestCarrierServicesListRunE_GetClientFails tests error handling when getClient fails.
func TestCarrierServicesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := carrierServicesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestCarrierServicesGetRunE_GetClientFails tests error handling when getClient fails.
func TestCarrierServicesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := carrierServicesGetCmd.RunE(cmd, []string{"cs_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestCarrierServicesCreateRunE_GetClientFails tests error handling when getClient fails.
func TestCarrierServicesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("callback-url", "https://example.com", "")
	cmd.Flags().Bool("service-discovery", false, "")
	cmd.Flags().String("type", "api", "")

	err := carrierServicesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestCarrierServicesDeleteRunE_GetClientFails tests error handling when getClient fails.
func TestCarrierServicesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := carrierServicesDeleteCmd.RunE(cmd, []string{"cs_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

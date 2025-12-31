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

// shippingZonesAPIClient is a mock implementation of api.APIClient for shipping zones tests.
type shippingZonesAPIClient struct {
	api.MockClient

	listShippingZonesResp  *api.ShippingZonesListResponse
	listShippingZonesErr   error
	getShippingZoneResp    *api.ShippingZone
	getShippingZoneErr     error
	createShippingZoneResp *api.ShippingZone
	createShippingZoneErr  error
	deleteShippingZoneErr  error
}

func (m *shippingZonesAPIClient) ListShippingZones(ctx context.Context, opts *api.ShippingZonesListOptions) (*api.ShippingZonesListResponse, error) {
	return m.listShippingZonesResp, m.listShippingZonesErr
}

func (m *shippingZonesAPIClient) GetShippingZone(ctx context.Context, id string) (*api.ShippingZone, error) {
	return m.getShippingZoneResp, m.getShippingZoneErr
}

func (m *shippingZonesAPIClient) CreateShippingZone(ctx context.Context, req *api.ShippingZoneCreateRequest) (*api.ShippingZone, error) {
	return m.createShippingZoneResp, m.createShippingZoneErr
}

func (m *shippingZonesAPIClient) DeleteShippingZone(ctx context.Context, id string) error {
	return m.deleteShippingZoneErr
}

func TestShippingZonesCommandStructure(t *testing.T) {
	subcommands := shippingZonesCmd.Commands()

	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		if _, exists := expectedCmds[cmd.Name()]; exists {
			expectedCmds[cmd.Name()] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestShippingZonesListFlags(t *testing.T) {
	flags := []string{"page", "page-size"}

	for _, flagName := range flags {
		if shippingZonesListCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on list command", flagName)
		}
	}
}

func TestShippingZonesGetArgs(t *testing.T) {
	err := shippingZonesGetCmd.Args(shippingZonesGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = shippingZonesGetCmd.Args(shippingZonesGetCmd, []string{"zone-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestShippingZonesDeleteArgs(t *testing.T) {
	err := shippingZonesDeleteCmd.Args(shippingZonesDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = shippingZonesDeleteCmd.Args(shippingZonesDeleteCmd, []string{"zone-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestShippingZonesCreateFlags(t *testing.T) {
	flags := []string{"name"}

	for _, flagName := range flags {
		if shippingZonesCreateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on create command", flagName)
		}
	}
}

func TestShippingZonesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := shippingZonesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestShippingZonesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := shippingZonesGetCmd.RunE(cmd, []string{"zone-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestShippingZonesListWithValidStore(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()

	err := shippingZonesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("shippingZonesListCmd succeeded (might be due to mock setup)")
	}
}

// TestShippingZonesListRunE_Success tests the shipping-zones list command execution with mock API.
func TestShippingZonesListRunE_Success(t *testing.T) {
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
		mockResp *api.ShippingZonesListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list",
			mockResp: &api.ShippingZonesListResponse{
				Items: []api.ShippingZone{
					{
						ID:        "sz_123",
						Name:      "US Shipping",
						Countries: []api.ZoneCountry{{Code: "US"}, {Code: "CA"}},
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
			mockResp: &api.ShippingZonesListResponse{
				Items:      []api.ShippingZone{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &shippingZonesAPIClient{
				listShippingZonesResp: tt.mockResp,
				listShippingZonesErr:  tt.mockErr,
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

			err := shippingZonesListCmd.RunE(cmd, []string{})

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

// TestShippingZonesGetRunE_Success tests the shipping-zones get command execution with mock API.
func TestShippingZonesGetRunE_Success(t *testing.T) {
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
		mockResp *api.ShippingZone
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "sz_123",
			mockResp: &api.ShippingZone{
				ID:        "sz_123",
				Name:      "US Shipping",
				Countries: []api.ZoneCountry{{Code: "US"}, {Code: "CA"}},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "sz_999",
			mockErr: errors.New("shipping zone not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &shippingZonesAPIClient{
				getShippingZoneResp: tt.mockResp,
				getShippingZoneErr:  tt.mockErr,
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

			err := shippingZonesGetCmd.RunE(cmd, []string{tt.id})

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

// TestShippingZonesCreateRunE_Success tests the shipping-zones create command execution with mock API.
func TestShippingZonesCreateRunE_Success(t *testing.T) {
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
		mockResp *api.ShippingZone
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.ShippingZone{
				ID:        "sz_new",
				Name:      "New Zone",
				Countries: []api.ZoneCountry{{Code: "US"}},
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
			mockClient := &shippingZonesAPIClient{
				createShippingZoneResp: tt.mockResp,
				createShippingZoneErr:  tt.mockErr,
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
			cmd.Flags().String("name", "New Zone", "")
			cmd.Flags().Bool("dry-run", false, "")

			err := shippingZonesCreateCmd.RunE(cmd, []string{})

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

// TestShippingZonesDeleteRunE_Success tests the shipping-zones delete command execution with mock API.
func TestShippingZonesDeleteRunE_Success(t *testing.T) {
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
			id:   "sz_123",
		},
		{
			name:    "delete fails",
			id:      "sz_456",
			mockErr: errors.New("zone in use"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &shippingZonesAPIClient{
				deleteShippingZoneErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "")
			cmd.Flags().Bool("dry-run", false, "")

			err := shippingZonesDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestShippingZonesCreateRunE_GetClientFails tests error handling when getClient fails.
func TestShippingZonesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")

	err := shippingZonesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestShippingZonesDeleteRunE_GetClientFails tests error handling when getClient fails.
func TestShippingZonesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := shippingZonesDeleteCmd.RunE(cmd, []string{"sz_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestShippingZonesListRunE_JSONOutput tests JSON output format for list command.
func TestShippingZonesListRunE_JSONOutput(t *testing.T) {
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

	mockClient := &shippingZonesAPIClient{
		listShippingZonesResp: &api.ShippingZonesListResponse{
			Items: []api.ShippingZone{
				{
					ID:        "sz_123",
					Name:      "US Shipping",
					Countries: []api.ZoneCountry{{Code: "US", Name: "United States"}},
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := shippingZonesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShippingZonesGetRunE_JSONOutput tests JSON output format for get command.
func TestShippingZonesGetRunE_JSONOutput(t *testing.T) {
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

	mockClient := &shippingZonesAPIClient{
		getShippingZoneResp: &api.ShippingZone{
			ID:        "sz_123",
			Name:      "US Shipping",
			Countries: []api.ZoneCountry{{Code: "US", Name: "United States"}},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
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

	err := shippingZonesGetCmd.RunE(cmd, []string{"sz_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShippingZonesGetRunE_WithRates tests get command with price and weight based rates.
func TestShippingZonesGetRunE_WithRates(t *testing.T) {
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

	mockClient := &shippingZonesAPIClient{
		getShippingZoneResp: &api.ShippingZone{
			ID:   "sz_123",
			Name: "US Shipping",
			Countries: []api.ZoneCountry{
				{Code: "US", Name: "United States"},
				{Code: "CA", Name: "Canada"},
			},
			PriceBasedRates: []api.ShippingRate{
				{
					ID:       "pr_1",
					Name:     "Standard Shipping",
					Price:    "5.99",
					MinValue: "0.00",
					MaxValue: "50.00",
				},
				{
					ID:       "pr_2",
					Name:     "Free Shipping",
					Price:    "0.00",
					MinValue: "50.01",
					MaxValue: "1000.00",
				},
			},
			WeightBasedRates: []api.ShippingRate{
				{
					ID:        "wr_1",
					Name:      "Light Package",
					Price:     "3.99",
					MinWeight: 0.0,
					MaxWeight: 1.0,
				},
				{
					ID:        "wr_2",
					Name:      "Heavy Package",
					Price:     "9.99",
					MinWeight: 1.01,
					MaxWeight: 10.0,
				},
			},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shippingZonesGetCmd.RunE(cmd, []string{"sz_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShippingZonesGetRunE_NoCountries tests get command with no countries.
func TestShippingZonesGetRunE_NoCountries(t *testing.T) {
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

	mockClient := &shippingZonesAPIClient{
		getShippingZoneResp: &api.ShippingZone{
			ID:        "sz_123",
			Name:      "Empty Zone",
			Countries: []api.ZoneCountry{},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shippingZonesGetCmd.RunE(cmd, []string{"sz_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShippingZonesCreateRunE_JSONOutput tests JSON output format for create command.
func TestShippingZonesCreateRunE_JSONOutput(t *testing.T) {
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

	mockClient := &shippingZonesAPIClient{
		createShippingZoneResp: &api.ShippingZone{
			ID:        "sz_new",
			Name:      "New Zone",
			Countries: []api.ZoneCountry{{Code: "US", Name: "United States"}},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("name", "New Zone", "")
	cmd.Flags().Bool("dry-run", false, "")
	_ = cmd.Flags().Set("output", "json")

	err := shippingZonesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShippingZonesListRunE_WithPriceAndWeightRates tests list with zones containing rates.
func TestShippingZonesListRunE_WithPriceAndWeightRates(t *testing.T) {
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

	mockClient := &shippingZonesAPIClient{
		listShippingZonesResp: &api.ShippingZonesListResponse{
			Items: []api.ShippingZone{
				{
					ID:   "sz_123",
					Name: "US Shipping",
					Countries: []api.ZoneCountry{
						{Code: "US", Name: "United States"},
					},
					PriceBasedRates: []api.ShippingRate{
						{ID: "pr_1", Name: "Standard", Price: "5.99"},
						{ID: "pr_2", Name: "Express", Price: "9.99"},
					},
					WeightBasedRates: []api.ShippingRate{
						{ID: "wr_1", Name: "Light", Price: "3.99"},
					},
					CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := shippingZonesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShippingZonesCommandSetup verifies shipping-zones command initialization.
func TestShippingZonesCommandSetup(t *testing.T) {
	if shippingZonesCmd.Use != "shipping-zones" {
		t.Errorf("expected Use 'shipping-zones', got %q", shippingZonesCmd.Use)
	}
	if shippingZonesCmd.Short != "Manage shipping zones" {
		t.Errorf("expected Short 'Manage shipping zones', got %q", shippingZonesCmd.Short)
	}
}

// TestShippingZonesSubcommands verifies all subcommands are registered.
func TestShippingZonesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List shipping zones",
		"get":    "Get shipping zone details",
		"create": "Create a shipping zone",
		"delete": "Delete a shipping zone",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range shippingZonesCmd.Commands() {
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

// TestShippingZonesGetCmd verifies get command configuration.
func TestShippingZonesGetCmd(t *testing.T) {
	if shippingZonesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", shippingZonesGetCmd.Use)
	}
	if shippingZonesGetCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestShippingZonesDeleteCmd verifies delete command configuration.
func TestShippingZonesDeleteCmd(t *testing.T) {
	if shippingZonesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", shippingZonesDeleteCmd.Use)
	}
	if shippingZonesDeleteCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestShippingZonesListFlagsWithDefaults verifies list command flags with correct defaults.
func TestShippingZonesListFlagsWithDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := shippingZonesListCmd.Flags().Lookup(f.name)
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

// TestShippingZonesCreateFlagsWithDefaults verifies create command flags with correct defaults.
func TestShippingZonesCreateFlagsWithDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := shippingZonesCreateCmd.Flags().Lookup(f.name)
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

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// deliveryOptionsAPIClient is a mock implementation of api.APIClient for delivery options tests.
type deliveryOptionsAPIClient struct {
	api.MockClient

	listDeliveryOptionsResp        *api.DeliveryOptionsListResponse
	listDeliveryOptionsErr         error
	getDeliveryOptionResp          *api.DeliveryOption
	getDeliveryOptionErr           error
	listDeliveryTimeSlotsResp      *api.DeliveryTimeSlotsListResponse
	listDeliveryTimeSlotsErr       error
	updateDeliveryOptionPickupResp *api.DeliveryOption
	updateDeliveryOptionPickupErr  error

	getDeliveryConfigResp        json.RawMessage
	getDeliveryConfigErr         error
	getDeliveryTimeSlotsOpenResp json.RawMessage
	getDeliveryTimeSlotsOpenErr  error
	updateDeliveryStoresInfoResp json.RawMessage
	updateDeliveryStoresInfoErr  error
	lastUpdateStoresInfoBody     json.RawMessage
}

func (m *deliveryOptionsAPIClient) ListDeliveryOptions(ctx context.Context, opts *api.DeliveryOptionsListOptions) (*api.DeliveryOptionsListResponse, error) {
	return m.listDeliveryOptionsResp, m.listDeliveryOptionsErr
}

func (m *deliveryOptionsAPIClient) GetDeliveryOption(ctx context.Context, id string) (*api.DeliveryOption, error) {
	return m.getDeliveryOptionResp, m.getDeliveryOptionErr
}

func (m *deliveryOptionsAPIClient) ListDeliveryTimeSlots(ctx context.Context, id string, opts *api.DeliveryTimeSlotsListOptions) (*api.DeliveryTimeSlotsListResponse, error) {
	return m.listDeliveryTimeSlotsResp, m.listDeliveryTimeSlotsErr
}

func (m *deliveryOptionsAPIClient) UpdateDeliveryOptionPickupStore(ctx context.Context, id string, req *api.PickupStoreUpdateRequest) (*api.DeliveryOption, error) {
	return m.updateDeliveryOptionPickupResp, m.updateDeliveryOptionPickupErr
}

func (m *deliveryOptionsAPIClient) GetDeliveryConfig(ctx context.Context, opts *api.DeliveryConfigOptions) (json.RawMessage, error) {
	return m.getDeliveryConfigResp, m.getDeliveryConfigErr
}

func (m *deliveryOptionsAPIClient) GetDeliveryTimeSlotsOpenAPI(ctx context.Context, id string) (json.RawMessage, error) {
	return m.getDeliveryTimeSlotsOpenResp, m.getDeliveryTimeSlotsOpenErr
}

func (m *deliveryOptionsAPIClient) UpdateDeliveryOptionStoresInfo(ctx context.Context, id string, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastUpdateStoresInfoBody = b
	}
	return m.updateDeliveryStoresInfoResp, m.updateDeliveryStoresInfoErr
}

func TestDeliveryOptionsCommandStructure(t *testing.T) {
	subcommands := deliveryOptionsCmd.Commands()

	expectedCmds := map[string]bool{
		"list":                false,
		"get":                 false,
		"time-slots":          false,
		"delivery-time-slots": false,
		"update-pickup":       false,
		"config":              false,
		"stores-info":         false,
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

func TestDeliveryOptionsListFlags(t *testing.T) {
	flags := []string{"page", "page-size", "status", "type"}

	for _, flagName := range flags {
		if deliveryOptionsListCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on list command", flagName)
		}
	}
}

func TestDeliveryOptionsGetArgs(t *testing.T) {
	err := deliveryOptionsGetCmd.Args(deliveryOptionsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = deliveryOptionsGetCmd.Args(deliveryOptionsGetCmd, []string{"opt-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDeliveryOptionsTimeSlotsArgs(t *testing.T) {
	err := deliveryOptionsTimeSlotsCmd.Args(deliveryOptionsTimeSlotsCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = deliveryOptionsTimeSlotsCmd.Args(deliveryOptionsTimeSlotsCmd, []string{"opt-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDeliveryOptionsUpdatePickupArgs(t *testing.T) {
	err := deliveryOptionsUpdatePickupCmd.Args(deliveryOptionsUpdatePickupCmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = deliveryOptionsUpdatePickupCmd.Args(deliveryOptionsUpdatePickupCmd, []string{"opt-id"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDeliveryOptionsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := deliveryOptionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

func TestDeliveryOptionsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()

	err := deliveryOptionsGetCmd.RunE(cmd, []string{"opt-id"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
}

// TestDeliveryOptionsListRunE_Success tests the delivery-options list command execution with mock API.
func TestDeliveryOptionsListRunE_Success(t *testing.T) {
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
		mockResp *api.DeliveryOptionsListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list",
			mockResp: &api.DeliveryOptionsListResponse{
				Items: []api.DeliveryOption{
					{
						ID:                 "do_123",
						Name:               "Standard Delivery",
						Type:               "shipping",
						Status:             "active",
						SupportedCountries: []string{"US", "CA"},
						CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
			mockResp: &api.DeliveryOptionsListResponse{
				Items:      []api.DeliveryOption{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &deliveryOptionsAPIClient{
				listDeliveryOptionsResp: tt.mockResp,
				listDeliveryOptionsErr:  tt.mockErr,
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
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("type", "", "")

			err := deliveryOptionsListCmd.RunE(cmd, []string{})

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

// TestDeliveryOptionsGetRunE_Success tests the delivery-options get command execution with mock API.
func TestDeliveryOptionsGetRunE_Success(t *testing.T) {
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
		mockResp *api.DeliveryOption
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "do_123",
			mockResp: &api.DeliveryOption{
				ID:                 "do_123",
				Name:               "Standard Delivery",
				Type:               "shipping",
				Status:             "active",
				Description:        "Standard delivery option",
				SupportedCountries: []string{"US", "CA"},
				CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:          time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "do_999",
			mockErr: errors.New("delivery option not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &deliveryOptionsAPIClient{
				getDeliveryOptionResp: tt.mockResp,
				getDeliveryOptionErr:  tt.mockErr,
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

			err := deliveryOptionsGetCmd.RunE(cmd, []string{tt.id})

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

// TestDeliveryOptionsTimeSlotsRunE_Success tests the delivery-options time-slots command execution.
func TestDeliveryOptionsTimeSlotsRunE_Success(t *testing.T) {
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
		mockResp *api.DeliveryTimeSlotsListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list",
			id:   "do_123",
			mockResp: &api.DeliveryTimeSlotsListResponse{
				Items: []api.DeliveryTimeSlot{
					{
						ID:        "ts_123",
						Date:      "2024-01-20",
						StartTime: "09:00",
						EndTime:   "12:00",
						Available: true,
						Capacity:  10,
						Booked:    3,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:    "API error",
			id:      "do_123",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &deliveryOptionsAPIClient{
				listDeliveryTimeSlotsResp: tt.mockResp,
				listDeliveryTimeSlotsErr:  tt.mockErr,
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
			cmd.Flags().String("start-date", "", "")
			cmd.Flags().String("end-date", "", "")

			err := deliveryOptionsTimeSlotsCmd.RunE(cmd, []string{tt.id})

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

// TestDeliveryOptionsUpdatePickupRunE_Success tests the update-pickup command execution.
func TestDeliveryOptionsUpdatePickupRunE_Success(t *testing.T) {
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
		mockResp *api.DeliveryOption
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful update",
			id:   "do_123",
			mockResp: &api.DeliveryOption{
				ID:     "do_123",
				Name:   "Pickup Option",
				Type:   "pickup",
				Status: "active",
			},
		},
		{
			name:    "update fails",
			id:      "do_123",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &deliveryOptionsAPIClient{
				updateDeliveryOptionPickupResp: tt.mockResp,
				updateDeliveryOptionPickupErr:  tt.mockErr,
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
			cmd.Flags().String("store-id", "store_123", "")
			cmd.Flags().String("store-name", "Test Store", "")
			cmd.Flags().String("address", "123 Main St", "")
			cmd.Flags().String("phone", "555-1234", "")

			err := deliveryOptionsUpdatePickupCmd.RunE(cmd, []string{tt.id})

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

// TestDeliveryOptionsListRunE_JSONOutput tests JSON output format for list command.
func TestDeliveryOptionsListRunE_JSONOutput(t *testing.T) {
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

	mockClient := &deliveryOptionsAPIClient{
		listDeliveryOptionsResp: &api.DeliveryOptionsListResponse{
			Items: []api.DeliveryOption{
				{
					ID:                 "do_123",
					Name:               "Standard Delivery",
					Type:               "shipping",
					Status:             "active",
					SupportedCountries: []string{"US"},
					CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("type", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := deliveryOptionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDeliveryOptionsCommandSetup verifies delivery-options command initialization.
func TestDeliveryOptionsCommandSetup(t *testing.T) {
	if deliveryOptionsCmd.Use != "delivery-options" {
		t.Errorf("expected Use 'delivery-options', got %q", deliveryOptionsCmd.Use)
	}
	if deliveryOptionsCmd.Short != "Manage delivery options" {
		t.Errorf("expected Short 'Manage delivery options', got %q", deliveryOptionsCmd.Short)
	}
}

func TestDeliveryOptionsStoresInfoUpdateFlags(t *testing.T) {
	flags := []string{"body", "body-file"}
	for _, flagName := range flags {
		if deliveryOptionsStoresInfoUpdateCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("Expected flag %q not found on stores-info update command", flagName)
		}
	}
}

func TestDeliveryOptionsConfigGetRunE_WithMockAPI(t *testing.T) {
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

	mockClient := &deliveryOptionsAPIClient{
		getDeliveryConfigResp: json.RawMessage(`{"ok":true}`),
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("type", "shipping", "")
	cmd.Flags().String("delivery-option-id", "", "")

	if err := deliveryOptionsConfigGetCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("expected ok=true in output, got %q", buf.String())
	}
}

func TestDeliveryOptionsDeliveryTimeSlotsRunE_WithMockAPI(t *testing.T) {
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

	mockClient := &deliveryOptionsAPIClient{
		getDeliveryTimeSlotsOpenResp: json.RawMessage(`{"items":[]}`),
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := deliveryOptionsDeliveryTimeSlotsCmd.RunE(cmd, []string{"do_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"items": []`)) {
		t.Fatalf("expected items in output, got %q", buf.String())
	}
}

func TestDeliveryOptionsStoresInfoUpdateRunE_WithMockAPI(t *testing.T) {
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

	mockClient := &deliveryOptionsAPIClient{
		updateDeliveryStoresInfoResp: json.RawMessage(`{"updated":true}`),
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("body", `{"stores":[]}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := deliveryOptionsStoresInfoUpdateCmd.RunE(cmd, []string{"do_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"updated": true`)) {
		t.Fatalf("expected updated=true in output, got %q", buf.String())
	}
	if len(mockClient.lastUpdateStoresInfoBody) == 0 {
		t.Fatalf("expected request body to be captured, got empty")
	}
}

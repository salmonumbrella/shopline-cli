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

func TestCustomerAddressesCmd(t *testing.T) {
	if customerAddressesCmd.Use != "customer-addresses" {
		t.Errorf("Expected Use to be 'customer-addresses', got %q", customerAddressesCmd.Use)
	}
}

func TestCustomerAddressesListCmd(t *testing.T) {
	if customerAddressesListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %q", customerAddressesListCmd.Use)
	}
}

func TestCustomerAddressesGetCmd(t *testing.T) {
	if customerAddressesGetCmd.Use != "get <address-id>" {
		t.Errorf("Expected Use to be 'get <address-id>', got %q", customerAddressesGetCmd.Use)
	}
}

func TestCustomerAddressesCreateCmd(t *testing.T) {
	if customerAddressesCreateCmd.Use != "create" {
		t.Errorf("Expected Use to be 'create', got %q", customerAddressesCreateCmd.Use)
	}
}

func TestCustomerAddressesSetDefaultCmd(t *testing.T) {
	if customerAddressesSetDefaultCmd.Use != "set-default <address-id>" {
		t.Errorf("Expected Use, got %q", customerAddressesSetDefaultCmd.Use)
	}
}

func TestCustomerAddressesDeleteCmd(t *testing.T) {
	if customerAddressesDeleteCmd.Use != "delete <address-id>" {
		t.Errorf("Expected Use, got %q", customerAddressesDeleteCmd.Use)
	}
}

func TestCustomerAddressesListFlags(t *testing.T) {
	flags := []string{"page", "page-size"}
	for _, flag := range flags {
		if customerAddressesListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerAddressesCreateFlags(t *testing.T) {
	flags := []string{"first-name", "last-name", "address", "city", "country", "phone", "default"}
	for _, flag := range flags {
		if customerAddressesCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerAddressesPersistentFlags(t *testing.T) {
	if customerAddressesCmd.PersistentFlags().Lookup("customer-id") == nil {
		t.Error("Expected persistent flag 'customer-id' to be defined")
	}
}

func TestCustomerAddressesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	err := customerAddressesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerAddressesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	err := customerAddressesGetCmd.RunE(cmd, []string{"addr_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerAddressesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("first-name", "John", "")
	cmd.Flags().String("last-name", "Doe", "")
	cmd.Flags().String("address", "123 Main St", "")
	cmd.Flags().String("city", "City", "")
	cmd.Flags().String("country", "US", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Bool("default", false, "")
	err := customerAddressesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerAddressesSetDefaultRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	err := customerAddressesSetDefaultCmd.RunE(cmd, []string{"addr_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerAddressesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "cust_123", "")
	err := customerAddressesDeleteCmd.RunE(cmd, []string{"addr_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerAddressesListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("customer-id", "cust_123", "")
	err := customerAddressesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// addressesMockClient is a mock implementation for customer addresses testing.
type addressesMockClient struct {
	api.MockClient
	listResp       *api.CustomerAddressesListResponse
	listErr        error
	getResp        *api.CustomerAddress
	getErr         error
	createResp     *api.CustomerAddress
	createErr      error
	setDefaultResp *api.CustomerAddress
	setDefaultErr  error
	deleteErr      error
}

func (m *addressesMockClient) ListCustomerAddresses(ctx context.Context, customerID string, opts *api.CustomerAddressesListOptions) (*api.CustomerAddressesListResponse, error) {
	return m.listResp, m.listErr
}

func (m *addressesMockClient) GetCustomerAddress(ctx context.Context, customerID, addressID string) (*api.CustomerAddress, error) {
	return m.getResp, m.getErr
}

func (m *addressesMockClient) CreateCustomerAddress(ctx context.Context, customerID string, req *api.CustomerAddressCreateRequest) (*api.CustomerAddress, error) {
	return m.createResp, m.createErr
}

func (m *addressesMockClient) SetDefaultCustomerAddress(ctx context.Context, customerID, addressID string) (*api.CustomerAddress, error) {
	return m.setDefaultResp, m.setDefaultErr
}

func (m *addressesMockClient) DeleteCustomerAddress(ctx context.Context, customerID, addressID string) error {
	return m.deleteErr
}

func TestCustomerAddressesListRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *addressesMockClient
		wantErr    bool
		errContain string
		checkOut   func(t *testing.T, out string)
	}{
		{
			name: "success with addresses",
			mockClient: &addressesMockClient{
				listResp: &api.CustomerAddressesListResponse{
					Items: []api.CustomerAddress{
						{
							ID:        "addr_123",
							FirstName: "John",
							LastName:  "Doe",
							Address1:  "123 Main St",
							City:      "NYC",
							Country:   "US",
							Default:   true,
							CreatedAt: now,
						},
						{
							ID:        "addr_456",
							FirstName: "Jane",
							LastName:  "Doe",
							Address1:  "456 Oak Ave",
							City:      "LA",
							Country:   "US",
							Default:   false,
							CreatedAt: now,
						},
					},
					TotalCount: 2,
				},
			},
		},
		{
			name: "API error",
			mockClient: &addressesMockClient{
				listErr: errors.New("API connection failed"),
			},
			wantErr:    true,
			errContain: "failed to list customer addresses",
		},
		{
			name: "empty list",
			mockClient: &addressesMockClient{
				listResp: &api.CustomerAddressesListResponse{
					Items:      []api.CustomerAddress{},
					TotalCount: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := customerAddressesListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkOut != nil {
				tt.checkOut(t, buf.String())
			}
		})
	}
}

func TestCustomerAddressesGetRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *addressesMockClient
		wantErr    bool
		errContain string
		checkOut   func(t *testing.T, out string)
	}{
		{
			name: "success",
			mockClient: &addressesMockClient{
				getResp: &api.CustomerAddress{
					ID:          "addr_123",
					CustomerID:  "cust_123",
					FirstName:   "John",
					LastName:    "Doe",
					Address1:    "123 Main St",
					City:        "NYC",
					Country:     "US",
					CountryCode: "US",
					Default:     true,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
		},
		{
			name: "not found",
			mockClient: &addressesMockClient{
				getErr: errors.New("address not found"),
			},
			wantErr:    true,
			errContain: "failed to get customer address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")

			err := customerAddressesGetCmd.RunE(cmd, []string{"addr_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkOut != nil {
				tt.checkOut(t, buf.String())
			}
		})
	}
}

func TestCustomerAddressesCreateRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *addressesMockClient
		wantErr    bool
		errContain string
		checkOut   func(t *testing.T, out string)
	}{
		{
			name: "success",
			mockClient: &addressesMockClient{
				createResp: &api.CustomerAddress{
					ID:        "addr_new",
					FirstName: "John",
					LastName:  "Doe",
					Address1:  "123 Main St",
					City:      "NYC",
					Country:   "US",
					Default:   false,
					CreatedAt: now,
				},
			},
		},
		{
			name: "API error",
			mockClient: &addressesMockClient{
				createErr: errors.New("validation failed"),
			},
			wantErr:    true,
			errContain: "failed to create customer address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("first-name", "John", "")
			cmd.Flags().String("last-name", "Doe", "")
			cmd.Flags().String("address", "123 Main St", "")
			cmd.Flags().String("city", "NYC", "")
			cmd.Flags().String("country", "US", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().Bool("default", false, "")

			err := customerAddressesCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkOut != nil {
				tt.checkOut(t, buf.String())
			}
		})
	}
}

func TestCustomerAddressesSetDefaultRunE_WithMockAPI(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockClient *addressesMockClient
		wantErr    bool
		errContain string
		checkOut   func(t *testing.T, out string)
	}{
		{
			name: "success",
			mockClient: &addressesMockClient{
				setDefaultResp: &api.CustomerAddress{
					ID:        "addr_123",
					Default:   true,
					CreatedAt: now,
				},
			},
		},
		{
			name: "API error",
			mockClient: &addressesMockClient{
				setDefaultErr: errors.New("address not found"),
			},
			wantErr:    true,
			errContain: "failed to set default address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")

			err := customerAddressesSetDefaultCmd.RunE(cmd, []string{"addr_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkOut != nil {
				tt.checkOut(t, buf.String())
			}
		})
	}
}

func TestCustomerAddressesDeleteRunE_WithMockAPI(t *testing.T) {
	tests := []struct {
		name       string
		mockClient *addressesMockClient
		wantErr    bool
		errContain string
		checkOut   func(t *testing.T, out string)
	}{
		{
			name: "success",
			mockClient: &addressesMockClient{
				deleteErr: nil,
			},
		},
		{
			name: "API error",
			mockClient: &addressesMockClient{
				deleteErr: errors.New("address not found"),
			},
			wantErr:    true,
			errContain: "failed to delete customer address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			clientFactory = func(handle, token string) api.APIClient {
				return tt.mockClient
			}
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := &cobra.Command{}
			cmd.Flags().StringP("store", "s", "", "")
			cmd.Flags().StringP("output", "o", "text", "")
			cmd.Flags().String("color", "auto", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().BoolP("yes", "y", true, "") // Skip confirmation

			err := customerAddressesDeleteCmd.RunE(cmd, []string{"addr_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("Expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkOut != nil {
				tt.checkOut(t, buf.String())
			}
		})
	}
}

// TestCustomerAddressesDeleteRunE_Cancelled tests delete cancellation when user declines.
func TestCustomerAddressesDeleteRunE_Cancelled(t *testing.T) {
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
	mockClient := &addressesMockClient{
		deleteErr: nil,
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	// Create a pipe to simulate user input
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "text", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().BoolP("yes", "y", false, "") // Don't skip confirmation

	err := customerAddressesDeleteCmd.RunE(cmd, []string{"addr_123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Cancelled") {
		t.Errorf("Expected output to contain 'Cancelled', got: %s", output)
	}
}

// TestCustomerAddressesDeleteRunE_ConfirmedWithY tests delete when user types 'Y'.
func TestCustomerAddressesDeleteRunE_ConfirmedWithY(t *testing.T) {
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
	mockClient := &addressesMockClient{
		deleteErr: nil,
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	// Create a pipe to simulate user input 'Y'
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("Y\n")
	_ = w.Close()

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "text", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().BoolP("yes", "y", false, "") // Don't skip confirmation

	err := customerAddressesDeleteCmd.RunE(cmd, []string{"addr_123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted") {
		t.Errorf("Expected output to contain 'Deleted', got: %s", output)
	}
}

// TestCustomerAddressesDeleteRunE_ConfirmedWithLowercaseY tests delete when user types 'y'.
func TestCustomerAddressesDeleteRunE_ConfirmedWithLowercaseY(t *testing.T) {
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
	mockClient := &addressesMockClient{
		deleteErr: nil,
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	// Create a pipe to simulate user input 'y'
	r, w, _ := os.Pipe()
	os.Stdin = r
	_, _ = w.WriteString("y\n")
	_ = w.Close()

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "text", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().BoolP("yes", "y", false, "") // Don't skip confirmation

	err := customerAddressesDeleteCmd.RunE(cmd, []string{"addr_123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted") {
		t.Errorf("Expected output to contain 'Deleted', got: %s", output)
	}
}

// TestCustomerAddressesListRunE_JSONOutput tests JSON output for list command.
func TestCustomerAddressesListRunE_JSONOutput(t *testing.T) {
	now := time.Now()
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
	mockClient := &addressesMockClient{
		listResp: &api.CustomerAddressesListResponse{
			Items: []api.CustomerAddress{
				{
					ID:        "addr_123",
					FirstName: "John",
					LastName:  "Doe",
					Address1:  "123 Main St",
					City:      "NYC",
					Country:   "US",
					Default:   true,
					CreatedAt: now,
				},
			},
			TotalCount: 1,
		},
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := customerAddressesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "addr_123") {
		t.Errorf("Expected JSON output to contain address ID, got: %s", output)
	}
}

// TestCustomerAddressesGetRunE_JSONOutput tests JSON output for get command.
func TestCustomerAddressesGetRunE_JSONOutput(t *testing.T) {
	now := time.Now()
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
	mockClient := &addressesMockClient{
		getResp: &api.CustomerAddress{
			ID:          "addr_123",
			CustomerID:  "cust_123",
			FirstName:   "John",
			LastName:    "Doe",
			Address1:    "123 Main St",
			City:        "NYC",
			Country:     "US",
			CountryCode: "US",
			Default:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	_ = cmd.Flags().Set("output", "json")

	err := customerAddressesGetCmd.RunE(cmd, []string{"addr_123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "addr_123") {
		t.Errorf("Expected JSON output to contain address ID, got: %s", output)
	}
}

// TestCustomerAddressesGetRunE_WithOptionalFields tests get command with all optional fields.
func TestCustomerAddressesGetRunE_WithOptionalFields(t *testing.T) {
	now := time.Now()
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
	mockClient := &addressesMockClient{
		getResp: &api.CustomerAddress{
			ID:           "addr_full",
			CustomerID:   "cust_123",
			FirstName:    "John",
			LastName:     "Doe",
			Company:      "Acme Corp",
			Address1:     "123 Main St",
			Address2:     "Suite 456",
			City:         "NYC",
			Province:     "New York",
			ProvinceCode: "NY",
			Country:      "United States",
			CountryCode:  "US",
			Zip:          "10001",
			Phone:        "+1-555-1234",
			Default:      false,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "text", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")

	err := customerAddressesGetCmd.RunE(cmd, []string{"addr_full"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	// Check all optional fields are displayed
	expectedStrings := []string{
		"Company:      Acme Corp",
		"Suite 456",
		"Province:     New York (NY)",
		"Phone:        +1-555-1234",
	}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got: %s", expected, output)
		}
	}
}

// TestCustomerAddressesCreateRunE_JSONOutput tests JSON output for create command.
func TestCustomerAddressesCreateRunE_JSONOutput(t *testing.T) {
	now := time.Now()
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
	mockClient := &addressesMockClient{
		createResp: &api.CustomerAddress{
			ID:        "addr_new",
			FirstName: "Jane",
			LastName:  "Smith",
			Address1:  "456 Oak Ave",
			City:      "LA",
			Country:   "US",
			Default:   true,
			CreatedAt: now,
		},
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}
	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.Flags().String("color", "auto", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("customer-id", "cust_123", "")
	cmd.Flags().String("first-name", "Jane", "")
	cmd.Flags().String("last-name", "Smith", "")
	cmd.Flags().String("address", "456 Oak Ave", "")
	cmd.Flags().String("city", "LA", "")
	cmd.Flags().String("country", "US", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Bool("default", true, "")
	_ = cmd.Flags().Set("output", "json")

	err := customerAddressesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "addr_new") {
		t.Errorf("Expected JSON output to contain address ID, got: %s", output)
	}
}

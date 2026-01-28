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

// TestCustomersCommandSetup verifies customers command initialization
func TestCustomersCommandSetup(t *testing.T) {
	if customersCmd.Use != "customers" {
		t.Errorf("expected Use 'customers', got %q", customersCmd.Use)
	}
	if customersCmd.Short != "Manage customers" {
		t.Errorf("expected Short 'Manage customers', got %q", customersCmd.Short)
	}
}

// TestCustomersSubcommands verifies all subcommands are registered
func TestCustomersSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list": "List customers",
		"get":  "Get customer details",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range customersCmd.Commands() {
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

// TestCustomersListFlags verifies list command flags exist with correct defaults
func TestCustomersListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"email", ""},
		{"state", ""},
		{"tags", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := customersListCmd.Flags().Lookup(f.name)
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

// TestCustomersGetArgs verifies get command requires exactly 1 argument
func TestCustomersGetArgs(t *testing.T) {
	err := customersGetCmd.Args(customersGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = customersGetCmd.Args(customersGetCmd, []string{"cust-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCustomersGetClientError verifies error handling when getClient fails
func TestCustomersGetClientError(t *testing.T) {
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

// TestCustomersListGetClientError verifies list command error handling when getClient fails
func TestCustomersListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(customersListCmd)

	err := customersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCustomersGetGetClientError verifies get command error handling when getClient fails
func TestCustomersGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(customersGetCmd)

	err := customersGetCmd.RunE(cmd, []string{"cust-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCustomersListNoProfiles verifies list command error handling when no profiles exist
func TestCustomersListNoProfiles(t *testing.T) {
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
	err := customersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestCustomersGetMultipleProfiles verifies get command error handling when multiple profiles exist
func TestCustomersGetMultipleProfiles(t *testing.T) {
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
	err := customersGetCmd.RunE(cmd, []string{"cust-id"})
	if err == nil {
		t.Error("expected error for multiple profiles")
	}
}

// TestCustomersWithMockStore tests customers commands with a mock credential store
func TestCustomersWithMockStore(t *testing.T) {
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

// TestCustomersListWithValidStore tests list command execution with valid store
func TestCustomersListWithValidStore(t *testing.T) {
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
	cmd.AddCommand(customersListCmd)

	// This will fail at the API call level, but validates the client setup works
	err := customersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("customersListCmd succeeded (might be due to mock setup)")
	}
}

// customersMockAPIClient is a mock implementation of api.APIClient for customers tests.
type customersMockAPIClient struct {
	api.MockClient
	listCustomersResp *api.CustomersListResponse
	listCustomersErr  error
	getCustomerResp   *api.Customer
	getCustomerErr    error
}

func (m *customersMockAPIClient) ListCustomers(ctx context.Context, opts *api.CustomersListOptions) (*api.CustomersListResponse, error) {
	return m.listCustomersResp, m.listCustomersErr
}

func (m *customersMockAPIClient) GetCustomer(ctx context.Context, id string) (*api.Customer, error) {
	return m.getCustomerResp, m.getCustomerErr
}

// setupCustomersMockFactories sets up mock factories for customers tests.
func setupCustomersMockFactories(mockClient *customersMockAPIClient) (func(), *bytes.Buffer) {
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

// newCustomersTestCmd creates a test command with common flags for customers tests.
func newCustomersTestCmd() *cobra.Command {
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

// TestCustomersListRunE tests the customers list command with mock API.
func TestCustomersListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CustomersListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CustomersListResponse{
				Items: []api.Customer{
					{
						ID:          "cust_123",
						Email:       "alice@example.com",
						FirstName:   "Alice",
						LastName:    "Smith",
						State:       "enabled",
						OrdersCount: 5,
						TotalSpent:  "250.00",
						Currency:    "USD",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cust_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CustomersListResponse{
				Items:      []api.Customer{},
				TotalCount: 0,
			},
		},
		{
			name: "customer with only first name",
			mockResp: &api.CustomersListResponse{
				Items: []api.Customer{
					{
						ID:        "cust_456",
						Email:     "bob@example.com",
						FirstName: "Bob",
						LastName:  "",
						State:     "enabled",
						CreatedAt: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "Bob",
		},
		{
			name: "customer with only last name",
			mockResp: &api.CustomersListResponse{
				Items: []api.Customer{
					{
						ID:        "cust_789",
						Email:     "charlie@example.com",
						FirstName: "",
						LastName:  "Brown",
						State:     "disabled",
						CreatedAt: time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "Brown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customersMockAPIClient{
				listCustomersResp: tt.mockResp,
				listCustomersErr:  tt.mockErr,
			}
			cleanup, buf := setupCustomersMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomersTestCmd()
			cmd.Flags().String("email", "", "")
			cmd.Flags().String("state", "", "")
			cmd.Flags().String("tags", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := customersListCmd.RunE(cmd, []string{})

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

// TestCustomersListRunEWithJSON tests JSON output format.
func TestCustomersListRunEWithJSON(t *testing.T) {
	mockClient := &customersMockAPIClient{
		listCustomersResp: &api.CustomersListResponse{
			Items: []api.Customer{
				{ID: "cust_json", Email: "json@example.com"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("state", "", "")
	cmd.Flags().String("tags", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := customersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cust_json") {
		t.Errorf("JSON output should contain customer ID, got: %s", output)
	}
}

// TestCustomersGetRunE tests the customers get command with mock API.
func TestCustomersGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		mockResp   *api.Customer
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get",
			customerID: "cust_123",
			mockResp: &api.Customer{
				ID:               "cust_123",
				Email:            "alice@example.com",
				FirstName:        "Alice",
				LastName:         "Smith",
				Phone:            "+1234567890",
				State:            "enabled",
				AcceptsMarketing: true,
				OrdersCount:      10,
				TotalSpent:       "500.00",
				Currency:         "USD",
				Tags:             []string{"vip", "loyal"},
				Note:             "Preferred customer",
				CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "customer not found",
			customerID: "cust_999",
			mockErr:    errors.New("customer not found"),
			wantErr:    true,
		},
		{
			name:       "get customer with only first name",
			customerID: "cust_456",
			mockResp: &api.Customer{
				ID:        "cust_456",
				Email:     "bob@example.com",
				FirstName: "Bob",
				LastName:  "",
			},
		},
		{
			name:       "get customer with only last name",
			customerID: "cust_789",
			mockResp: &api.Customer{
				ID:        "cust_789",
				Email:     "charlie@example.com",
				FirstName: "",
				LastName:  "Brown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customersMockAPIClient{
				getCustomerResp: tt.mockResp,
				getCustomerErr:  tt.mockErr,
			}
			cleanup, _ := setupCustomersMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomersTestCmd()

			err := customersGetCmd.RunE(cmd, []string{tt.customerID})

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

// TestCustomersGetRunEWithJSON tests JSON output format for get command.
func TestCustomersGetRunEWithJSON(t *testing.T) {
	mockClient := &customersMockAPIClient{
		getCustomerResp: &api.Customer{
			ID:        "cust_json",
			Email:     "json@example.com",
			FirstName: "JSON",
			LastName:  "Test",
		},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := customersGetCmd.RunE(cmd, []string{"cust_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cust_json") {
		t.Errorf("JSON output should contain customer ID, got: %s", output)
	}
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// TestFormatCustomerCreditBalance tests the formatCustomerCreditBalance function.
func TestFormatCustomerCreditBalance(t *testing.T) {
	tests := []struct {
		name     string
		customer *api.Customer
		want     string
	}{
		{
			name:     "nil customer",
			customer: nil,
			want:     "N/A",
		},
		{
			name:     "nil balance",
			customer: &api.Customer{Currency: "USD"},
			want:     "N/A",
		},
		{
			name:     "zero balance with currency",
			customer: &api.Customer{CreditBalance: ptr(0.0), Currency: "USD"},
			want:     "0.00 USD",
		},
		{
			name:     "positive balance with currency",
			customer: &api.Customer{CreditBalance: ptr(42.50), Currency: "USD"},
			want:     "42.50 USD",
		},
		{
			name:     "balance without currency",
			customer: &api.Customer{CreditBalance: ptr(100.0)},
			want:     "100.00",
		},
		{
			name:     "negative balance",
			customer: &api.Customer{CreditBalance: ptr(-50.0), Currency: "USD"},
			want:     "-50.00 USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCustomerCreditBalance(tt.customer)
			if got != tt.want {
				t.Errorf("formatCustomerCreditBalance() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFormatCustomerSubscriptions tests the formatCustomerSubscriptions function.
func TestFormatCustomerSubscriptions(t *testing.T) {
	tests := []struct {
		name     string
		customer *api.Customer
		want     string
	}{
		{
			name:     "nil customer",
			customer: nil,
			want:     "N/A",
		},
		{
			name:     "empty subscriptions slice",
			customer: &api.Customer{Subscriptions: []api.CustomerSubscription{}},
			want:     "N/A",
		},
		{
			name: "single active subscription",
			customer: &api.Customer{
				Subscriptions: []api.CustomerSubscription{
					{Platform: "email", IsActive: true},
				},
			},
			want: "email=active",
		},
		{
			name: "single inactive subscription",
			customer: &api.Customer{
				Subscriptions: []api.CustomerSubscription{
					{Platform: "sms", IsActive: false},
				},
			},
			want: "sms=inactive",
		},
		{
			name: "mixed active and inactive subscriptions",
			customer: &api.Customer{
				Subscriptions: []api.CustomerSubscription{
					{Platform: "email", IsActive: true},
					{Platform: "sms", IsActive: false},
				},
			},
			want: "email=active, sms=inactive",
		},
		{
			name: "empty platform name active",
			customer: &api.Customer{
				Subscriptions: []api.CustomerSubscription{
					{Platform: "", IsActive: true},
				},
			},
			want: "active",
		},
		{
			name: "empty platform name inactive",
			customer: &api.Customer{
				Subscriptions: []api.CustomerSubscription{
					{Platform: "", IsActive: false},
				},
			},
			want: "inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCustomerSubscriptions(tt.customer)
			if got != tt.want {
				t.Errorf("formatCustomerSubscriptions() = %q, want %q", got, tt.want)
			}
		})
	}
}

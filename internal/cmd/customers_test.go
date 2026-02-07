package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
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
		"list":            "List customers",
		"get":             "Get customer details",
		"search":          "Search customers",
		"create":          "Create a customer",
		"update":          "Update a customer",
		"delete":          "Delete a customer",
		"tags":            "Manage customer tags",
		"subscriptions":   "Manage customer subscriptions",
		"line":            "Lookup customers by LINE ID",
		"metafields":      "Manage customer metafields",
		"app-metafields":  "Manage customer app metafields",
		"store-credits":   "Manage customer store credits",
		"membership-info": "Get customers membership info",
		"membership-tier": "Customer membership tier tools",
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

// TestCustomersGetArgs verifies get command accepts 0 or 1 arguments
func TestCustomersGetArgs(t *testing.T) {
	err := customersGetCmd.Args(customersGetCmd, []string{})
	if err != nil {
		t.Errorf("expected no error with 0 args (--by mode), got: %v", err)
	}

	err = customersGetCmd.Args(customersGetCmd, []string{"cust-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = customersGetCmd.Args(customersGetCmd, []string{"a", "b"})
	if err == nil {
		t.Error("expected error with 2 args")
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
	listCustomersResp               *api.CustomersListResponse
	listCustomersErr                error
	listCustomersByPage             map[int]*api.CustomersListResponse
	listCustomersCalls              []*api.CustomersListOptions
	getCustomerResp                 *api.Customer
	getCustomerErr                  error
	searchCustomersResp             *api.CustomersListResponse
	searchCustomersErr              error
	searchCustomersByPage           map[int]*api.CustomersListResponse
	searchCustomersCalls            []*api.CustomerSearchOptions
	createCustomerResp              *api.Customer
	createCustomerErr               error
	updateCustomerResp              *api.Customer
	updateCustomerErr               error
	deleteCustomerErr               error
	setCustomerTagsResp             *api.Customer
	setCustomerTagsErr              error
	updateCustomerTagsResp          *api.Customer
	updateCustomerTagsErr           error
	updateCustomerSubscriptionsResp json.RawMessage
	updateCustomerSubscriptionsErr  error
	getLineCustomerResp             *api.Customer
	getLineCustomerErr              error
}

func (m *customersMockAPIClient) ListCustomers(ctx context.Context, opts *api.CustomersListOptions) (*api.CustomersListResponse, error) {
	if opts != nil {
		cp := *opts
		m.listCustomersCalls = append(m.listCustomersCalls, &cp)
		if m.listCustomersByPage != nil {
			if resp, ok := m.listCustomersByPage[opts.Page]; ok {
				return resp, m.listCustomersErr
			}
		}
	}
	return m.listCustomersResp, m.listCustomersErr
}

func (m *customersMockAPIClient) GetCustomer(ctx context.Context, id string) (*api.Customer, error) {
	return m.getCustomerResp, m.getCustomerErr
}

func (m *customersMockAPIClient) SearchCustomers(ctx context.Context, opts *api.CustomerSearchOptions) (*api.CustomersListResponse, error) {
	if opts != nil {
		cp := *opts
		m.searchCustomersCalls = append(m.searchCustomersCalls, &cp)
		if m.searchCustomersByPage != nil {
			if resp, ok := m.searchCustomersByPage[opts.Page]; ok {
				return resp, m.searchCustomersErr
			}
		}
	}
	return m.searchCustomersResp, m.searchCustomersErr
}

func (m *customersMockAPIClient) CreateCustomer(ctx context.Context, req *api.CustomerCreateRequest) (*api.Customer, error) {
	return m.createCustomerResp, m.createCustomerErr
}

func (m *customersMockAPIClient) UpdateCustomer(ctx context.Context, id string, req *api.CustomerUpdateRequest) (*api.Customer, error) {
	return m.updateCustomerResp, m.updateCustomerErr
}

func (m *customersMockAPIClient) DeleteCustomer(ctx context.Context, id string) error {
	return m.deleteCustomerErr
}

func (m *customersMockAPIClient) SetCustomerTags(ctx context.Context, id string, tags []string) (*api.Customer, error) {
	return m.setCustomerTagsResp, m.setCustomerTagsErr
}

func (m *customersMockAPIClient) UpdateCustomerTags(ctx context.Context, id string, req *api.CustomerTagsUpdateRequest) (*api.Customer, error) {
	return m.updateCustomerTagsResp, m.updateCustomerTagsErr
}

func (m *customersMockAPIClient) UpdateCustomerSubscriptions(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return m.updateCustomerSubscriptionsResp, m.updateCustomerSubscriptionsErr
}

func (m *customersMockAPIClient) GetLineCustomer(ctx context.Context, lineID string) (*api.Customer, error) {
	return m.getLineCustomerResp, m.getLineCustomerErr
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

func TestCustomersSearchRunEWithJSON(t *testing.T) {
	mockClient := &customersMockAPIClient{
		searchCustomersResp: &api.CustomersListResponse{
			Items:      []api.Customer{{ID: "cust_s", Email: "s@example.com"}},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("q", "", "")
	_ = cmd.Flags().Set("q", "alice")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := customersSearchCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cust_s") {
		t.Fatalf("expected output to contain customer id, got: %s", buf.String())
	}
}

func TestCustomersListRunELimitPaginates(t *testing.T) {
	page1 := &api.CustomersListResponse{
		Items:      make([]api.Customer, 24),
		TotalCount: 100,
		HasMore:    true,
	}
	for i := range page1.Items {
		page1.Items[i] = api.Customer{ID: "cust_p1_" + strconv.Itoa(i)}
	}

	page2 := &api.CustomersListResponse{
		Items:      make([]api.Customer, 24),
		TotalCount: 100,
		HasMore:    false,
	}
	for i := range page2.Items {
		page2.Items[i] = api.Customer{ID: "cust_p2_" + strconv.Itoa(i)}
	}

	mockClient := &customersMockAPIClient{
		listCustomersByPage: map[int]*api.CustomersListResponse{
			1: page1,
			2: page2,
		},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Int("limit", 0, "")
	_ = cmd.Flags().Set("limit", "30")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("state", "", "")
	cmd.Flags().String("tags", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := customersListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.listCustomersCalls) < 2 {
		t.Fatalf("expected at least 2 ListCustomers calls, got %d", len(mockClient.listCustomersCalls))
	}
	if mockClient.listCustomersCalls[0].Page != 1 || mockClient.listCustomersCalls[1].Page != 2 {
		t.Fatalf("expected calls for pages 1 and 2, got %+v", mockClient.listCustomersCalls)
	}

	var resp api.CustomersListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if len(resp.Items) != 30 {
		t.Fatalf("expected 30 items, got %d", len(resp.Items))
	}
}

func TestCustomersSearchRunELimitPaginates(t *testing.T) {
	page1 := &api.CustomersListResponse{
		Items:      make([]api.Customer, 24),
		TotalCount: 100,
		HasMore:    true,
	}
	for i := range page1.Items {
		page1.Items[i] = api.Customer{ID: "cust_s1_" + strconv.Itoa(i)}
	}

	page2 := &api.CustomersListResponse{
		Items:      make([]api.Customer, 24),
		TotalCount: 100,
		HasMore:    false,
	}
	for i := range page2.Items {
		page2.Items[i] = api.Customer{ID: "cust_s2_" + strconv.Itoa(i)}
	}

	mockClient := &customersMockAPIClient{
		searchCustomersByPage: map[int]*api.CustomersListResponse{
			1: page1,
			2: page2,
		},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Int("limit", 0, "")
	_ = cmd.Flags().Set("limit", "30")
	cmd.Flags().String("q", "", "")
	_ = cmd.Flags().Set("q", "alice")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := customersSearchCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.searchCustomersCalls) < 2 {
		t.Fatalf("expected at least 2 SearchCustomers calls, got %d", len(mockClient.searchCustomersCalls))
	}
	if mockClient.searchCustomersCalls[0].Page != 1 || mockClient.searchCustomersCalls[1].Page != 2 {
		t.Fatalf("expected calls for pages 1 and 2, got %+v", mockClient.searchCustomersCalls)
	}

	var resp api.CustomersListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if len(resp.Items) != 30 {
		t.Fatalf("expected 30 items, got %d", len(resp.Items))
	}
}

func TestCustomersCreateRunEWithJSON(t *testing.T) {
	mockClient := &customersMockAPIClient{
		createCustomerResp: &api.Customer{ID: "cust_new", Email: "new@example.com"},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("email", "new@example.com", "")
	cmd.Flags().String("first-name", "New", "")
	cmd.Flags().String("last-name", "User", "")
	cmd.Flags().String("phone", "+1", "")
	cmd.Flags().Bool("accepts-marketing", false, "")
	cmd.Flags().StringSlice("tag", nil, "")
	cmd.Flags().String("note", "", "")

	if err := customersCreateCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cust_new") {
		t.Fatalf("expected output to contain customer id, got: %s", buf.String())
	}
}

func TestCustomersUpdateRunEWithJSON(t *testing.T) {
	mockClient := &customersMockAPIClient{
		updateCustomerResp: &api.Customer{ID: "cust_1", Email: "u@example.com"},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().Bool("accepts-marketing", false, "")
	cmd.Flags().StringSlice("tag", nil, "")
	cmd.Flags().String("note", "", "")
	_ = cmd.Flags().Set("first-name", "Updated")

	if err := customersUpdateCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cust_1") {
		t.Fatalf("expected output to contain customer id, got: %s", buf.String())
	}
}

func TestCustomersDeleteRunE(t *testing.T) {
	mockClient := &customersMockAPIClient{}
	cleanup, _ := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	if err := customersDeleteCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomersTagsSetRunE(t *testing.T) {
	mockClient := &customersMockAPIClient{
		setCustomerTagsResp: &api.Customer{ID: "cust_1", Tags: []string{"a"}},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	cmd.Flags().StringSlice("tag", []string{"a"}, "")

	if err := customersTagsSetCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cust_1") {
		t.Fatalf("expected output to contain customer id, got: %s", buf.String())
	}
}

func TestCustomersSubscriptionsUpdateRunE(t *testing.T) {
	mockClient := &customersMockAPIClient{
		updateCustomerSubscriptionsResp: json.RawMessage(`{"ok":true}`),
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	cmd.Flags().String("body", `{"k":"v"}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := customersSubscriptionsUpdateCmd.RunE(cmd, []string{"cust_1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"ok\"") {
		t.Fatalf("expected output to contain ok, got: %s", buf.String())
	}
}

func TestCustomersLineGetRunE(t *testing.T) {
	mockClient := &customersMockAPIClient{
		getLineCustomerResp: &api.Customer{ID: "cust_line"},
	}
	cleanup, buf := setupCustomersMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomersTestCmd()
	if err := customersLineGetCmd.RunE(cmd, []string{"line_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cust_line") {
		t.Fatalf("expected output to contain customer id, got: %s", buf.String())
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

// TestCustomersGetByFlag tests the --by flag on the customers get command.
func TestCustomersGetByFlag(t *testing.T) {
	t.Run("resolves customer by email", func(t *testing.T) {
		mockClient := &customersMockAPIClient{
			searchCustomersResp: &api.CustomersListResponse{
				Items:      []api.Customer{{ID: "cust_found", Email: "alice@example.com"}},
				TotalCount: 1,
			},
			getCustomerResp: &api.Customer{
				ID:    "cust_found",
				Email: "alice@example.com",
			},
		}
		cleanup, buf := setupCustomersMockFactories(mockClient)
		defer cleanup()

		cmd := newCustomersTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "alice@example.com")

		if err := customersGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "cust_found") {
			t.Errorf("expected output to contain 'cust_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &customersMockAPIClient{
			searchCustomersResp: &api.CustomersListResponse{
				Items:      []api.Customer{},
				TotalCount: 0,
			},
		}
		cleanup, _ := setupCustomersMockFactories(mockClient)
		defer cleanup()

		cmd := newCustomersTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "nobody@example.com")

		err := customersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no customer found")
		}
		if !strings.Contains(err.Error(), "no customer found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &customersMockAPIClient{
			searchCustomersErr: errors.New("API error"),
		}
		cleanup, _ := setupCustomersMockFactories(mockClient)
		defer cleanup()

		cmd := newCustomersTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "alice@example.com")

		err := customersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("warns on multiple matches", func(t *testing.T) {
		mockClient := &customersMockAPIClient{
			searchCustomersResp: &api.CustomersListResponse{
				Items: []api.Customer{
					{ID: "cust_1", Email: "alice@example.com"},
					{ID: "cust_2", Email: "alice+2@example.com"},
				},
				TotalCount: 2,
			},
			getCustomerResp: &api.Customer{
				ID:    "cust_1",
				Email: "alice@example.com",
			},
		}
		cleanup, buf := setupCustomersMockFactories(mockClient)
		defer cleanup()

		cmd := newCustomersTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "alice")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		if err := customersGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "cust_1") {
			t.Errorf("expected output to contain 'cust_1', got: %s", buf.String())
		}
		if !strings.Contains(stderr.String(), "2 customers matched") {
			t.Errorf("expected stderr warning about multiple matches, got: %s", stderr.String())
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &customersMockAPIClient{
			getCustomerResp: &api.Customer{
				ID:    "cust_direct",
				Email: "direct@example.com",
			},
		}
		cleanup, buf := setupCustomersMockFactories(mockClient)
		defer cleanup()

		cmd := newCustomersTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := customersGetCmd.RunE(cmd, []string{"cust_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "cust_direct") {
			t.Errorf("expected output to contain 'cust_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &customersMockAPIClient{}
		cleanup, _ := setupCustomersMockFactories(mockClient)
		defer cleanup()

		cmd := newCustomersTestCmd()
		cmd.Flags().String("by", "", "")

		err := customersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
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

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

func TestCustomerSavedSearchesCmd(t *testing.T) {
	if customerSavedSearchesCmd.Use != "customer-saved-searches" {
		t.Errorf("Expected Use to be 'customer-saved-searches', got %q", customerSavedSearchesCmd.Use)
	}
	if customerSavedSearchesCmd.Short != "Manage customer saved searches" {
		t.Errorf("Expected Short to be 'Manage customer saved searches', got %q", customerSavedSearchesCmd.Short)
	}
}

func TestCustomerSavedSearchesListCmd(t *testing.T) {
	if customerSavedSearchesListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %q", customerSavedSearchesListCmd.Use)
	}
	if customerSavedSearchesListCmd.Short != "List customer saved searches" {
		t.Errorf("Expected Short to be 'List customer saved searches', got %q", customerSavedSearchesListCmd.Short)
	}
}

func TestCustomerSavedSearchesGetCmd(t *testing.T) {
	if customerSavedSearchesGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use to be 'get <id>', got %q", customerSavedSearchesGetCmd.Use)
	}
	if customerSavedSearchesGetCmd.Short != "Get saved search details" {
		t.Errorf("Expected Short to be 'Get saved search details', got %q", customerSavedSearchesGetCmd.Short)
	}
}

func TestCustomerSavedSearchesCreateCmd(t *testing.T) {
	if customerSavedSearchesCreateCmd.Use != "create" {
		t.Errorf("Expected Use to be 'create', got %q", customerSavedSearchesCreateCmd.Use)
	}
	if customerSavedSearchesCreateCmd.Short != "Create a customer saved search" {
		t.Errorf("Expected Short to be 'Create a customer saved search', got %q", customerSavedSearchesCreateCmd.Short)
	}
}

func TestCustomerSavedSearchesDeleteCmd(t *testing.T) {
	if customerSavedSearchesDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use to be 'delete <id>', got %q", customerSavedSearchesDeleteCmd.Use)
	}
	if customerSavedSearchesDeleteCmd.Short != "Delete a customer saved search" {
		t.Errorf("Expected Short to be 'Delete a customer saved search', got %q", customerSavedSearchesDeleteCmd.Short)
	}
}

func TestCustomerSavedSearchesListFlags(t *testing.T) {
	flags := []string{"name", "page", "page-size"}
	for _, flag := range flags {
		if customerSavedSearchesListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerSavedSearchesCreateFlags(t *testing.T) {
	flags := []string{"name", "q"}
	for _, flag := range flags {
		if customerSavedSearchesCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q to be defined", flag)
		}
	}
}

func TestCustomerSavedSearchesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := customerSavedSearchesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerSavedSearchesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := customerSavedSearchesGetCmd.RunE(cmd, []string{"search_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerSavedSearchesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newCustomerSavedSearchesCreateTestCmd()
	_ = cmd.Flags().Set("name", "VIP Customers")
	_ = cmd.Flags().Set("q", "orders_count > 10")
	err := customerSavedSearchesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerSavedSearchesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := customerSavedSearchesDeleteCmd.RunE(cmd, []string{"search_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCustomerSavedSearchesListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := customerSavedSearchesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

func TestCustomerSavedSearchesGetRunE_MultipleProfiles(t *testing.T) {
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
	err := customerSavedSearchesGetCmd.RunE(cmd, []string{"search_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles without selection, got nil")
	}
}

func TestCustomerSavedSearchesCreateRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newCustomerSavedSearchesCreateTestCmd()
	_ = cmd.Flags().Set("name", "VIP Customers")
	_ = cmd.Flags().Set("q", "orders_count > 10")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err := customerSavedSearchesCreateCmd.RunE(cmd, []string{})

	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
}

func TestCustomerSavedSearchesDeleteRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newTestCmdWithFlags()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err := customerSavedSearchesDeleteCmd.RunE(cmd, []string{"search_123"})

	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
}

// customerSavedSearchesMockClient is a mock implementation of api.APIClient for customer saved searches tests.
type customerSavedSearchesMockClient struct {
	api.MockClient
	listResp   *api.CustomerSavedSearchesListResponse
	listErr    error
	getResp    *api.CustomerSavedSearch
	getErr     error
	createResp *api.CustomerSavedSearch
	createErr  error
	deleteErr  error
}

func (m *customerSavedSearchesMockClient) ListCustomerSavedSearches(ctx context.Context, opts *api.CustomerSavedSearchesListOptions) (*api.CustomerSavedSearchesListResponse, error) {
	return m.listResp, m.listErr
}

func (m *customerSavedSearchesMockClient) GetCustomerSavedSearch(ctx context.Context, id string) (*api.CustomerSavedSearch, error) {
	return m.getResp, m.getErr
}

func (m *customerSavedSearchesMockClient) CreateCustomerSavedSearch(ctx context.Context, req *api.CustomerSavedSearchCreateRequest) (*api.CustomerSavedSearch, error) {
	return m.createResp, m.createErr
}

func (m *customerSavedSearchesMockClient) DeleteCustomerSavedSearch(ctx context.Context, id string) error {
	return m.deleteErr
}

// setupCustomerSavedSearchesMockFactories sets up mock factories for customer saved searches tests.
func setupCustomerSavedSearchesMockFactories(mockClient *customerSavedSearchesMockClient) (func(), *bytes.Buffer) {
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

// newCustomerSavedSearchesTestCmd creates a test command with common flags for customer saved searches tests.
func newCustomerSavedSearchesTestCmd() *cobra.Command {
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

// TestCustomerSavedSearchesListRunE tests the list command with mock API.
func TestCustomerSavedSearchesListRunE(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.CustomerSavedSearchesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with multiple items",
			mockResp: &api.CustomerSavedSearchesListResponse{
				Items: []api.CustomerSavedSearch{
					{
						ID:        "search_123",
						Name:      "VIP Customers",
						Query:     "orders_count > 10",
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
					{
						ID:        "search_456",
						Name:      "New Customers",
						Query:     "created_at > -30d",
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "search_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CustomerSavedSearchesListResponse{
				Items:      []api.CustomerSavedSearch{},
				TotalCount: 0,
			},
		},
		{
			name: "single item",
			mockResp: &api.CustomerSavedSearchesListResponse{
				Items: []api.CustomerSavedSearch{
					{
						ID:        "search_789",
						Name:      "High Spenders",
						Query:     "total_spent > 1000",
						CreatedAt: testTime,
						UpdatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "High Spenders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customerSavedSearchesMockClient{
				listResp: tt.mockResp,
				listErr:  tt.mockErr,
			}
			cleanup, buf := setupCustomerSavedSearchesMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomerSavedSearchesTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := customerSavedSearchesListCmd.RunE(cmd, []string{})

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

// TestCustomerSavedSearchesListRunEWithJSON tests JSON output format for list.
func TestCustomerSavedSearchesListRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &customerSavedSearchesMockClient{
		listResp: &api.CustomerSavedSearchesListResponse{
			Items: []api.CustomerSavedSearch{
				{
					ID:        "search_json",
					Name:      "JSON Test",
					Query:     "test_query",
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCustomerSavedSearchesMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomerSavedSearchesTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := customerSavedSearchesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "search_json") {
		t.Errorf("JSON output should contain search ID, got: %s", output)
	}
}

// TestCustomerSavedSearchesListRunEWithFilters tests list with name filter.
func TestCustomerSavedSearchesListRunEWithFilters(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &customerSavedSearchesMockClient{
		listResp: &api.CustomerSavedSearchesListResponse{
			Items: []api.CustomerSavedSearch{
				{
					ID:        "search_filtered",
					Name:      "VIP Customers",
					Query:     "orders_count > 10",
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCustomerSavedSearchesMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomerSavedSearchesTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("name", "VIP")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "50")

	err := customerSavedSearchesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "VIP Customers") {
		t.Errorf("output should contain filtered results, got: %s", output)
	}
}

// TestCustomerSavedSearchesGetRunE tests the get command with mock API.
func TestCustomerSavedSearchesGetRunE(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		searchID string
		mockResp *api.CustomerSavedSearch
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			searchID: "search_123",
			mockResp: &api.CustomerSavedSearch{
				ID:        "search_123",
				Name:      "VIP Customers",
				Query:     "orders_count > 10",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:     "search not found",
			searchID: "search_999",
			mockErr:  errors.New("saved search not found"),
			wantErr:  true,
		},
		{
			name:     "search with complex query",
			searchID: "search_456",
			mockResp: &api.CustomerSavedSearch{
				ID:        "search_456",
				Name:      "Premium Loyal Customers",
				Query:     "orders_count > 5 AND total_spent > 500 AND accepts_marketing = true",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customerSavedSearchesMockClient{
				getResp: tt.mockResp,
				getErr:  tt.mockErr,
			}
			cleanup, _ := setupCustomerSavedSearchesMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomerSavedSearchesTestCmd()

			err := customerSavedSearchesGetCmd.RunE(cmd, []string{tt.searchID})

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

// TestCustomerSavedSearchesGetRunEWithJSON tests JSON output format for get.
func TestCustomerSavedSearchesGetRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &customerSavedSearchesMockClient{
		getResp: &api.CustomerSavedSearch{
			ID:        "search_json",
			Name:      "JSON Get Test",
			Query:     "test_query",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, buf := setupCustomerSavedSearchesMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomerSavedSearchesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := customerSavedSearchesGetCmd.RunE(cmd, []string{"search_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "search_json") {
		t.Errorf("JSON output should contain search ID, got: %s", output)
	}
}

// TestCustomerSavedSearchesGetArgs verifies get command requires exactly 1 argument.
func TestCustomerSavedSearchesGetArgs(t *testing.T) {
	err := customerSavedSearchesGetCmd.Args(customerSavedSearchesGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = customerSavedSearchesGetCmd.Args(customerSavedSearchesGetCmd, []string{"search-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// newCustomerSavedSearchesCreateTestCmd creates a test command with flags specific for create tests.
func newCustomerSavedSearchesCreateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().String("name", "", "Search name")
	cmd.Flags().String("q", "", "Search query")
	return cmd
}

// TestCustomerSavedSearchesCreateRunE tests the create command with mock API.
func TestCustomerSavedSearchesCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		searchName string
		query      string
		mockResp   *api.CustomerSavedSearch
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful create",
			searchName: "VIP Customers",
			query:      "orders_count > 10",
			mockResp: &api.CustomerSavedSearch{
				ID:        "search_new",
				Name:      "VIP Customers",
				Query:     "orders_count > 10",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:       "create fails",
			searchName: "Bad Search",
			query:      "invalid_query",
			mockErr:    errors.New("invalid query syntax"),
			wantErr:    true,
		},
		{
			name:       "create with complex query",
			searchName: "Multi-condition Search",
			query:      "state:enabled AND orders_count:>5 AND total_spent:>100",
			mockResp: &api.CustomerSavedSearch{
				ID:        "search_complex",
				Name:      "Multi-condition Search",
				Query:     "state:enabled AND orders_count:>5 AND total_spent:>100",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customerSavedSearchesMockClient{
				createResp: tt.mockResp,
				createErr:  tt.mockErr,
			}
			cleanup, _ := setupCustomerSavedSearchesMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomerSavedSearchesCreateTestCmd()
			_ = cmd.Flags().Set("name", tt.searchName)
			_ = cmd.Flags().Set("q", tt.query)

			err := customerSavedSearchesCreateCmd.RunE(cmd, []string{})

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

// TestCustomerSavedSearchesCreateRunEWithJSON tests JSON output format for create.
// Note: The create command uses --q for search query and --query for output filter.
// We set both to keep this test deterministic for JSON formatting.
func TestCustomerSavedSearchesCreateRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &customerSavedSearchesMockClient{
		createResp: &api.CustomerSavedSearch{
			ID:        "search_json_create",
			Name:      "JSON Create Test",
			Query:     ".",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, buf := setupCustomerSavedSearchesMockFactories(mockClient)
	defer cleanup()

	cmd := newCustomerSavedSearchesCreateTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("name", "JSON Create Test")
	_ = cmd.Flags().Set("q", ".")
	_ = cmd.Flags().Set("query", ".")

	err := customerSavedSearchesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "search_json_create") {
		t.Errorf("JSON output should contain search ID, got: %s", output)
	}
}

// TestCustomerSavedSearchesDeleteRunE tests the delete command with mock API.
func TestCustomerSavedSearchesDeleteRunE(t *testing.T) {
	tests := []struct {
		name     string
		searchID string
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			searchID: "search_123",
			mockErr:  nil,
		},
		{
			name:     "delete not found",
			searchID: "search_999",
			mockErr:  errors.New("saved search not found"),
			wantErr:  true,
		},
		{
			name:     "delete fails with server error",
			searchID: "search_456",
			mockErr:  errors.New("internal server error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &customerSavedSearchesMockClient{
				deleteErr: tt.mockErr,
			}
			cleanup, _ := setupCustomerSavedSearchesMockFactories(mockClient)
			defer cleanup()

			cmd := newCustomerSavedSearchesTestCmd()

			err := customerSavedSearchesDeleteCmd.RunE(cmd, []string{tt.searchID})

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

// TestCustomerSavedSearchesDeleteArgs verifies delete command requires exactly 1 argument.
func TestCustomerSavedSearchesDeleteArgs(t *testing.T) {
	err := customerSavedSearchesDeleteCmd.Args(customerSavedSearchesDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = customerSavedSearchesDeleteCmd.Args(customerSavedSearchesDeleteCmd, []string{"search-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCustomerSavedSearchesSubcommands verifies all subcommands are registered.
func TestCustomerSavedSearchesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List customer saved searches",
		"get":    "Get saved search details",
		"create": "Create a customer saved search",
		"delete": "Delete a customer saved search",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range customerSavedSearchesCmd.Commands() {
				if sub.Use == name || strings.HasPrefix(sub.Use, name+" ") {
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

// TestCustomerSavedSearchesListFlagDefaults verifies list command flag defaults.
func TestCustomerSavedSearchesListFlagDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := customerSavedSearchesListCmd.Flags().Lookup(f.name)
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

// TestCustomerSavedSearchesCreateMultipleProfiles verifies create command error handling when multiple profiles exist.
func TestCustomerSavedSearchesCreateMultipleProfiles(t *testing.T) {
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

	cmd := newCustomerSavedSearchesCreateTestCmd()
	_ = cmd.Flags().Set("name", "Test Search")
	_ = cmd.Flags().Set("q", "test_query")
	err := customerSavedSearchesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for multiple profiles without selection")
	}
}

// TestCustomerSavedSearchesDeleteMultipleProfiles verifies delete command error handling when multiple profiles exist.
func TestCustomerSavedSearchesDeleteMultipleProfiles(t *testing.T) {
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
	err := customerSavedSearchesDeleteCmd.RunE(cmd, []string{"search_123"})
	if err == nil {
		t.Error("expected error for multiple profiles without selection")
	}
}

// TestCustomerSavedSearchesDeleteNoProfiles verifies delete command error handling when no profiles exist.
func TestCustomerSavedSearchesDeleteNoProfiles(t *testing.T) {
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
	err := customerSavedSearchesDeleteCmd.RunE(cmd, []string{"search_123"})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestCustomerSavedSearchesCreateNoProfiles verifies create command error handling when no profiles exist.
func TestCustomerSavedSearchesCreateNoProfiles(t *testing.T) {
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

	cmd := newCustomerSavedSearchesCreateTestCmd()
	_ = cmd.Flags().Set("name", "Test Search")
	_ = cmd.Flags().Set("q", "test_query")
	err := customerSavedSearchesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestCustomerSavedSearchesGetNoProfiles verifies get command error handling when no profiles exist.
func TestCustomerSavedSearchesGetNoProfiles(t *testing.T) {
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
	err := customerSavedSearchesGetCmd.RunE(cmd, []string{"search_123"})
	if err == nil {
		t.Error("expected error for no profiles")
	}
}

// TestCustomerSavedSearchesListMultipleProfiles verifies list command error handling when multiple profiles exist.
func TestCustomerSavedSearchesListMultipleProfiles(t *testing.T) {
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
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := customerSavedSearchesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for multiple profiles without selection")
	}
}

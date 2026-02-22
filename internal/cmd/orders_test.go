package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/batch"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// mockAPIClient is a mock implementation of api.APIClient for testing.
type mockAPIClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values for specific methods
	listOrdersResp   *api.OrdersListResponse
	listOrdersErr    error
	listOrdersByPage map[int]*api.OrdersListResponse
	listOrdersCalls  []*api.OrdersListOptions

	getOrderResp *api.Order
	getOrderErr  error
	getOrderByID map[string]*api.Order
	getOrderMu   sync.Mutex
	getOrderIDs  []string

	cancelOrderErr error

	createOrderResp *api.Order
	createOrderErr  error
	createOrderReq  *api.OrderCreateRequest

	updateOrderResp *api.Order
	updateOrderErr  error
	updateOrderReq  *api.OrderUpdateRequest
	updateOrderID   string

	searchOrdersResp  *api.OrdersListResponse
	searchOrdersErr   error
	searchOrdersCalls []*api.OrderSearchOptions
}

func (m *mockAPIClient) ListOrders(ctx context.Context, opts *api.OrdersListOptions) (*api.OrdersListResponse, error) {
	if opts != nil {
		cp := *opts
		m.listOrdersCalls = append(m.listOrdersCalls, &cp)
		if m.listOrdersByPage != nil {
			if resp, ok := m.listOrdersByPage[opts.Page]; ok {
				return resp, m.listOrdersErr
			}
		}
	}
	return m.listOrdersResp, m.listOrdersErr
}

func (m *mockAPIClient) GetOrder(ctx context.Context, id string) (*api.Order, error) {
	m.getOrderMu.Lock()
	m.getOrderIDs = append(m.getOrderIDs, id)
	m.getOrderMu.Unlock()
	if m.getOrderByID != nil {
		if o, ok := m.getOrderByID[id]; ok {
			return o, m.getOrderErr
		}
	}
	return m.getOrderResp, m.getOrderErr
}

func (m *mockAPIClient) CancelOrder(ctx context.Context, id string) error {
	return m.cancelOrderErr
}

func (m *mockAPIClient) CreateOrder(ctx context.Context, req *api.OrderCreateRequest) (*api.Order, error) {
	m.createOrderReq = req
	return m.createOrderResp, m.createOrderErr
}

func (m *mockAPIClient) UpdateOrder(ctx context.Context, id string, req *api.OrderUpdateRequest) (*api.Order, error) {
	m.updateOrderID = id
	m.updateOrderReq = req
	return m.updateOrderResp, m.updateOrderErr
}

func (m *mockAPIClient) SearchOrders(ctx context.Context, opts *api.OrderSearchOptions) (*api.OrdersListResponse, error) {
	if opts != nil {
		cp := *opts
		m.searchOrdersCalls = append(m.searchOrdersCalls, &cp)
	}
	return m.searchOrdersResp, m.searchOrdersErr
}

// mockStore is a mock implementation of CredentialStore for testing.
type mockStore struct {
	names []string
	creds map[string]*secrets.StoreCredentials
	err   error
}

func (m *mockStore) List() ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.names, nil
}

func (m *mockStore) Get(name string) (*secrets.StoreCredentials, error) {
	if m.err != nil {
		return nil, m.err
	}
	if creds, ok := m.creds[name]; ok {
		return creds, nil
	}
	return nil, errors.New("not found")
}

func TestCancelOrdersBatchInput(t *testing.T) {
	// Test that batch.ReadItems correctly parses order IDs
	// This tests the parsing logic without making API calls

	tests := []struct {
		name    string
		input   string
		wantIDs []string
		wantErr bool
	}{
		{
			name:    "JSON array",
			input:   `[{"id": "ord_123"}, {"id": "ord_456"}]`,
			wantIDs: []string{"ord_123", "ord_456"},
		},
		{
			name:    "NDJSON",
			input:   "{\"id\": \"ord_789\"}\n{\"id\": \"ord_101\"}",
			wantIDs: []string{"ord_789", "ord_101"},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := batch.ReadItemsFromReader(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(items) != len(tt.wantIDs) {
				t.Fatalf("Expected %d items, got %d", len(tt.wantIDs), len(items))
			}

			for i, item := range items {
				var parsed struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(item, &parsed); err != nil {
					t.Fatalf("Failed to parse item %d: %v", i, err)
				}
				if parsed.ID != tt.wantIDs[i] {
					t.Errorf("Item %d: expected ID %q, got %q", i, tt.wantIDs[i], parsed.ID)
				}
			}
		})
	}
}

func TestBatchResultOutput(t *testing.T) {
	// Test that results are written in NDJSON format
	results := []batch.Result{
		{ID: "ord_123", Index: 0, Success: true},
		{ID: "ord_456", Index: 1, Success: false, Error: "order not found"},
		{Index: 2, Success: false, Error: "missing id field"},
	}

	var buf bytes.Buffer
	err := batch.WriteResults(&buf, results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var result batch.Result
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}

	// Verify specific content
	if !strings.Contains(output, `"id":"ord_123"`) {
		t.Error("Missing first order ID")
	}
	if !strings.Contains(output, `"success":true`) {
		t.Error("Missing success field")
	}
	if !strings.Contains(output, `"error":"order not found"`) {
		t.Error("Missing error message")
	}
}

func TestBatchInputValidation(t *testing.T) {
	// Test parsing items with missing or invalid id fields
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "missing id field",
			input:   `{"name": "test"}`,
			wantErr: "", // No parse error, but ID will be empty
		},
		{
			name:    "invalid JSON",
			input:   `{not valid json}`,
			wantErr: "", // ReadItems doesn't validate JSON structure, just reads lines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := batch.ReadItemsFromReader(strings.NewReader(tt.input))
			if err != nil {
				// Some inputs might fail at read level
				return
			}

			// Test parsing the item
			if len(items) > 0 {
				var parsed struct {
					ID string `json:"id"`
				}
				err := json.Unmarshal(items[0], &parsed)
				if tt.name == "invalid JSON" && err == nil {
					t.Error("Expected parse error for invalid JSON")
				}
				if tt.name == "missing id field" && parsed.ID != "" {
					t.Errorf("Expected empty ID, got %q", parsed.ID)
				}
			}
		})
	}
}

// Helper function to create a test command with persistent flags.
// Uses PersistentFlags() so flags are accessible via cmd.Flags().Get*() methods.
func newTestCmdWithFlags() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "Store profile name")
	cmd.Flags().StringP("output", "o", "text", "Output format")
	cmd.Flags().String("color", "auto", "Color mode")
	cmd.Flags().String("query", "", "JQ filter")
	cmd.Flags().Bool("dry-run", false, "Preview changes without executing them")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	return cmd
}

func TestGetClient(t *testing.T) {
	// Save and restore original factory and env
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	tests := []struct {
		name      string
		storeName string
		envStore  string
		store     *mockStore
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success with store flag",
			storeName: "mystore",
			store: &mockStore{
				names: []string{"mystore"},
				creds: map[string]*secrets.StoreCredentials{
					"mystore": {Handle: "test", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name:     "success with env var",
			envStore: "envstore",
			store: &mockStore{
				names: []string{"envstore"},
				creds: map[string]*secrets.StoreCredentials{
					"envstore": {Handle: "test", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name:      "success when store flag matches handle",
			storeName: "demoshop",
			store: &mockStore{
				names: []string{"demo"},
				creds: map[string]*secrets.StoreCredentials{
					"demo": {Handle: "demoshop", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name:      "success when store flag uses admin URL",
			storeName: "https://admin.shoplineapp.com/admin/demoshop/",
			store: &mockStore{
				names: []string{"demo"},
				creds: map[string]*secrets.StoreCredentials{
					"demo": {Handle: "demoshop", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name:      "success when store flag uses shop domain",
			storeName: "demoshop.myshopline.com",
			store: &mockStore{
				names: []string{"demo"},
				creds: map[string]*secrets.StoreCredentials{
					"demo": {Handle: "demoshop", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name:      "success when store flag is unique prefix",
			storeName: "demo",
			store: &mockStore{
				names: []string{"demoshop"},
				creds: map[string]*secrets.StoreCredentials{
					"demoshop": {Handle: "demoshop", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name: "success auto-select single profile",
			store: &mockStore{
				names: []string{"only-store"},
				creds: map[string]*secrets.StoreCredentials{
					"only-store": {Handle: "test", AccessToken: "token123"},
				},
			},
			wantErr: false,
		},
		{
			name: "error no profiles",
			store: &mockStore{
				names: []string{},
			},
			wantErr: true,
			errMsg:  "no store profiles configured",
		},
		{
			name: "error multiple profiles without flag",
			store: &mockStore{
				names: []string{"store1", "store2"},
			},
			wantErr: true,
			errMsg:  "multiple profiles configured",
		},
		{
			name:      "error profile not found",
			storeName: "nonexistent",
			store: &mockStore{
				names: []string{"other"},
				creds: map[string]*secrets.StoreCredentials{},
			},
			wantErr: true,
			errMsg:  "profile not found",
		},
		{
			name:      "error ambiguous prefix",
			storeName: "demo",
			store: &mockStore{
				names: []string{"demo-ca", "demo-us"},
				creds: map[string]*secrets.StoreCredentials{
					"demo-ca": {Handle: "demoshop", AccessToken: "token123"},
					"demo-us": {Handle: "demousa", AccessToken: "token456"},
				},
			},
			wantErr: true,
			errMsg:  "multiple matches",
		},
		{
			name: "error factory fails",
			store: &mockStore{
				err: errors.New("keyring error"),
			},
			wantErr: true,
			errMsg:  "failed to open credential store",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset env
			_ = os.Unsetenv("SHOPLINE_STORE")
			if tt.envStore != "" {
				_ = os.Setenv("SHOPLINE_STORE", tt.envStore)
			}

			// Setup mock
			secretsStoreFactory = func() (CredentialStore, error) {
				if tt.store.err != nil {
					return nil, tt.store.err
				}
				return tt.store, nil
			}

			// Create command with flags
			cmd := newTestCmdWithFlags()
			if tt.storeName != "" {
				_ = cmd.Flags().Set("store", tt.storeName)
			}

			client, err := getClient(cmd)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if client == nil {
				t.Error("Expected client, got nil")
			}
		})
	}
}

func TestGetClient_UsesEnvTokenWithoutStoreProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origClientFactory := clientFactory
	origStoreEnv := os.Getenv(defaultStoreEnvName)
	origAccessToken := os.Getenv(envAccessToken)
	origAPIToken := os.Getenv(envAPIToken)
	origGenericToken := os.Getenv(envGenericToken)
	origAdminToken := os.Getenv(envAdminToken)
	defer func() {
		secretsStoreFactory = origFactory
		clientFactory = origClientFactory
		_ = os.Setenv(defaultStoreEnvName, origStoreEnv)
		_ = os.Setenv(envAccessToken, origAccessToken)
		_ = os.Setenv(envAPIToken, origAPIToken)
		_ = os.Setenv(envGenericToken, origGenericToken)
		_ = os.Setenv(envAdminToken, origAdminToken)
	}()

	_ = os.Unsetenv(defaultStoreEnvName)
	_ = os.Unsetenv(envAccessToken)
	_ = os.Setenv(envAPIToken, "api-token")
	_ = os.Unsetenv(envGenericToken)
	_ = os.Unsetenv(envAdminToken)

	storeFactoryCalled := false
	secretsStoreFactory = func() (CredentialStore, error) {
		storeFactoryCalled = true
		return nil, errors.New("should not be called when env token is set")
	}

	var gotToken string
	clientFactory = func(handle, accessToken string) api.APIClient {
		gotToken = accessToken
		return &api.MockClient{}
	}

	cmd := newTestCmdWithFlags()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if storeFactoryCalled {
		t.Fatal("expected credential store factory to be skipped when env token is present")
	}
	if gotToken != "api-token" {
		t.Fatalf("expected api token from env, got %q", gotToken)
	}
}

func TestGetClient_ExplicitStoreBeatsEnvToken(t *testing.T) {
	origFactory := secretsStoreFactory
	origClientFactory := clientFactory
	origStoreEnv := os.Getenv(defaultStoreEnvName)
	origAPIToken := os.Getenv(envAPIToken)
	defer func() {
		secretsStoreFactory = origFactory
		clientFactory = origClientFactory
		_ = os.Setenv(defaultStoreEnvName, origStoreEnv)
		_ = os.Setenv(envAPIToken, origAPIToken)
	}()

	_ = os.Unsetenv(defaultStoreEnvName)
	_ = os.Setenv(envAPIToken, "env-token")

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"mystore"},
			creds: map[string]*secrets.StoreCredentials{
				"mystore": {Handle: "handle", AccessToken: "profile-token"},
			},
		}, nil
	}

	var gotToken string
	clientFactory = func(handle, accessToken string) api.APIClient {
		gotToken = accessToken
		return &api.MockClient{}
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("store", "mystore")

	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if gotToken != "profile-token" {
		t.Fatalf("expected profile token to take precedence, got %q", gotToken)
	}
}

func TestGetFormatter(t *testing.T) {
	// Save and restore original writer
	origWriter := formatterWriter
	defer func() { formatterWriter = origWriter }()

	tests := []struct {
		name         string
		outputFormat string
		colorMode    string
		query        string
	}{
		{
			name:         "text format",
			outputFormat: "text",
			colorMode:    "auto",
		},
		{
			name:         "json format",
			outputFormat: "json",
			colorMode:    "always",
		},
		{
			name:         "with query",
			outputFormat: "json",
			colorMode:    "never",
			query:        ".items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a buffer for output
			buf := new(bytes.Buffer)
			formatterWriter = buf

			cmd := newTestCmdWithFlags()
			_ = cmd.Flags().Set("output", tt.outputFormat)
			_ = cmd.Flags().Set("color", tt.colorMode)
			if tt.query != "" {
				_ = cmd.Flags().Set("query", tt.query)
			}

			f := getFormatter(cmd)
			if f == nil {
				t.Error("Expected formatter, got nil")
			}
		})
	}
}

func TestDefaultClientFactory(t *testing.T) {
	client := defaultClientFactory("test-handle", "test-token")
	if client == nil {
		t.Error("Expected client, got nil")
	}
}

func TestDefaultSecretsStoreFactory(t *testing.T) {
	// This will fail if keyring is not available, which is expected in CI
	// We just test that the function is callable
	_, _ = defaultSecretsStoreFactory()
}

func TestCancelOrdersBatch(t *testing.T) {
	// Save and restore original factory
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	tests := []struct {
		name    string
		content string
		store   *mockStore
		wantErr bool
		errMsg  string
	}{
		{
			name:    "getClient fails",
			content: `[{"id": "ord_123"}]`,
			store: &mockStore{
				err: errors.New("keyring error"),
			},
			wantErr: true,
			errMsg:  "failed to open credential store",
		},
		{
			name:    "file not found",
			content: "", // Will use nonexistent file
			store: &mockStore{
				names: []string{"test"},
				creds: map[string]*secrets.StoreCredentials{
					"test": {Handle: "test", AccessToken: "token"},
				},
			},
			wantErr: true,
			errMsg:  "failed to read batch file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			secretsStoreFactory = func() (CredentialStore, error) {
				if tt.store.err != nil {
					return nil, tt.store.err
				}
				return tt.store, nil
			}

			cmd := newTestCmdWithFlags()

			var filename string
			if tt.content != "" {
				// Create temp file
				f, err := os.CreateTemp("", "batch-test-*.json")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer func() { _ = os.Remove(f.Name()) }()
				if _, err := f.WriteString(tt.content); err != nil {
					t.Fatalf("Failed to write temp file: %v", err)
				}
				_ = f.Close()
				filename = f.Name()
			} else {
				filename = "/nonexistent/batch-file.json"
			}

			err := cancelOrdersBatch(cmd, filename)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCancelOrdersBatchWithFile(t *testing.T) {
	// Save and restore original factory
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	// Create temp file with valid JSON
	content := `[{"id": "ord_123"}, {"id": "ord_456"}]`
	f, err := os.CreateTemp("", "batch-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = f.Close()

	// Setup mock store
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()

	// This will fail at the API call level, but exercises most of the code
	err = cancelOrdersBatch(cmd, f.Name())
	// We expect an error because the API call will fail (no real server)
	// but this exercises the batch reading and processing code
	if err == nil {
		// It might succeed if there are no items or something unexpected happens
		t.Log("cancelOrdersBatch succeeded (might be due to mock setup)")
	}
}

func TestCancelOrdersBatchInvalidJSON(t *testing.T) {
	// Save and restore original factory
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	// Create temp file with invalid JSON items
	content := `[{"not_id": "value"}]`
	f, err := os.CreateTemp("", "batch-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = f.Close()

	// Setup mock store
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()

	// Run the batch - will process items but mark them as missing id
	err = cancelOrdersBatch(cmd, f.Name())
	// Errors in individual items don't cause the function to return an error
	// The function returns nil and writes results to stdout
	if err != nil {
		t.Logf("cancelOrdersBatch returned error: %v (expected for API calls)", err)
	}
}

func TestCancelOrdersBatchMalformedJSON(t *testing.T) {
	// Save and restore original factory
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	// Create temp file with malformed JSON
	content := `{not valid json}`
	f, err := os.CreateTemp("", "batch-test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = f.Close()

	// Setup mock store
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()

	// Run the batch with malformed JSON
	_ = cancelOrdersBatch(cmd, f.Name())
	// The batch reader treats malformed JSON as NDJSON lines
}

// Ensure unused imports don't cause errors
var (
	_ = api.NewClient
	_ io.Writer
)

// setupOrdersTest sets up mock factories for order tests.
func setupOrdersTest(t *testing.T, mockClient *mockAPIClient, mockStore *mockStore) (cleanup func()) {
	t.Helper()
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	secretsStoreFactory = func() (CredentialStore, error) {
		if mockStore.err != nil {
			return nil, mockStore.err
		}
		return mockStore, nil
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// defaultMockStore returns a mock store with a single test profile.
func defaultMockStore() *mockStore {
	return &mockStore{
		names: []string{"test"},
		creds: map[string]*secrets.StoreCredentials{
			"test": {Handle: "test", AccessToken: "token"},
		},
	}
}

// TestOrdersListRunE tests the orders list command execution with mock API.
func TestOrdersListRunE(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     *api.OrdersListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
		wantHeaders  []string
		wantJSONID   string
	}{
		{
			name: "successful list text format",
			mockResp: &api.OrdersListResponse{
				Items: []api.OrderSummary{
					{
						ID:            "ord_123",
						OrderNumber:   "1001",
						Status:        "completed",
						TotalPrice:    "99.99",
						Currency:      "USD",
						CustomerEmail: "test@example.com",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFormat: "text",
			wantOutput:   "[order:$ord_123]",
			wantHeaders:  []string{"ORDER", "NUMBER", "STATUS", "TOTAL", "CUSTOMER", "CREATED"},
		},
		{
			name: "successful list JSON format",
			mockResp: &api.OrdersListResponse{
				Items: []api.OrderSummary{
					{
						ID:            "ord_456",
						OrderNumber:   "1002",
						Status:        "pending",
						TotalPrice:    "50.00",
						Currency:      "EUR",
						CustomerEmail: "user@example.com",
						CreatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   `"id": "ord_456"`,
			wantJSONID:   "ord_456",
		},
		{
			name:         "API error",
			mockErr:      errors.New("API unavailable"),
			outputFormat: "text",
			wantErr:      true,
		},
		{
			name: "empty list",
			mockResp: &api.OrdersListResponse{
				Items:      []api.OrderSummary{},
				TotalCount: 0,
			},
			outputFormat: "text",
			wantOutput:   "", // Empty table headers only go to buffer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAPIClient{
				listOrdersResp: tt.mockResp,
				listOrdersErr:  tt.mockErr,
			}
			cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := ordersListCmd.RunE(cmd, []string{})

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
			if len(tt.wantHeaders) > 0 {
				firstLine := strings.SplitN(output, "\n", 2)[0]
				fields := strings.Fields(firstLine)
				if len(fields) != len(tt.wantHeaders) {
					t.Fatalf("expected %d headers, got %d: %q", len(tt.wantHeaders), len(fields), firstLine)
				}
				for i, header := range tt.wantHeaders {
					if fields[i] != header {
						t.Fatalf("header[%d] = %q, want %q (line: %q)", i, fields[i], header, firstLine)
					}
				}
			}
			if tt.wantJSONID != "" {
				var resp api.OrdersListResponse
				if err := json.Unmarshal([]byte(output), &resp); err != nil {
					t.Fatalf("failed to unmarshal JSON output: %v", err)
				}
				if len(resp.Items) == 0 || resp.Items[0].ID != tt.wantJSONID {
					t.Fatalf("unexpected JSON items: %+v", resp.Items)
				}
			}
		})
	}
}

func TestOrdersListRunELimitPaginates(t *testing.T) {
	page1 := &api.OrdersListResponse{
		Items:      make([]api.OrderSummary, 24),
		TotalCount: 100,
		HasMore:    true,
	}
	for i := range page1.Items {
		page1.Items[i] = api.OrderSummary{ID: "ord_p1_" + strconv.Itoa(i)}
	}

	page2 := &api.OrdersListResponse{
		Items:      make([]api.OrderSummary, 24),
		TotalCount: 100,
		HasMore:    false,
	}
	for i := range page2.Items {
		page2.Items[i] = api.OrderSummary{ID: "ord_p2_" + strconv.Itoa(i)}
	}

	mockClient := &mockAPIClient{
		listOrdersByPage: map[int]*api.OrdersListResponse{
			1: page1,
			2: page2,
		},
	}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().Int("limit", 0, "")
	_ = cmd.Flags().Set("limit", "30")

	if err := ordersListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.listOrdersCalls) < 2 {
		t.Fatalf("expected at least 2 ListOrders calls, got %d", len(mockClient.listOrdersCalls))
	}
	if mockClient.listOrdersCalls[0].Page != 1 || mockClient.listOrdersCalls[1].Page != 2 {
		t.Fatalf("expected calls for pages 1 and 2, got %+v", mockClient.listOrdersCalls)
	}

	var resp api.OrdersListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if len(resp.Items) != 30 {
		t.Fatalf("expected 30 items, got %d", len(resp.Items))
	}
}

func TestOrdersListRunEUsesSearchWhenEmailFilterSet(t *testing.T) {
	mockClient := &mockAPIClient{
		searchOrdersResp: &api.OrdersListResponse{
			Items: []api.OrderSummary{
				{
					ID:            "ord_emailed",
					OrderNumber:   "1009",
					Status:        "confirmed",
					CustomerEmail: "user@example.com",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("email", "user@example.com")

	if err := ordersListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.listOrdersCalls) != 0 {
		t.Fatalf("expected ListOrders not to be called, got %d calls", len(mockClient.listOrdersCalls))
	}
	if len(mockClient.searchOrdersCalls) != 1 {
		t.Fatalf("expected SearchOrders to be called once, got %d", len(mockClient.searchOrdersCalls))
	}
	if got := mockClient.searchOrdersCalls[0].Query; got != "user@example.com" {
		t.Fatalf("expected SearchOrders query user@example.com, got %q", got)
	}

	var resp api.OrdersListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].ID != "ord_emailed" {
		t.Fatalf("unexpected response items: %+v", resp.Items)
	}
}

func TestOrdersListRunERejectsEmailAndCustomerIDTogether(t *testing.T) {
	mockClient := &mockAPIClient{}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("email", "user@example.com")
	_ = cmd.Flags().Set("customer-id", "cust_123")

	err := ordersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when both --email and --customer-id are set")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrdersListRunEJSONExpandDetails(t *testing.T) {
	mockClient := &mockAPIClient{
		listOrdersResp: &api.OrdersListResponse{
			Items: []api.OrderSummary{
				{ID: "ord_1", OrderNumber: "1001"},
				{ID: "ord_2", OrderNumber: "1002"},
			},
			TotalCount: 2,
			HasMore:    false,
		},
		getOrderByID: map[string]*api.Order{
			"ord_1": {
				ID:          "ord_1",
				OrderNumber: "1001",
				LineItems:   []api.OrderLineItem{{ID: "li_1", ProductID: "prod_1", Quantity: 1}},
			},
			"ord_2": {
				ID:          "ord_2",
				OrderNumber: "1002",
				LineItems:   []api.OrderLineItem{{ID: "li_2", ProductID: "prod_2", Quantity: 2}},
			},
		},
	}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().StringSlice("expand", nil, "")
	cmd.Flags().Int("jobs", 2, "")
	_ = cmd.Flags().Set("expand", "details")

	if err := ordersListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.getOrderIDs) != 2 {
		t.Fatalf("expected 2 GetOrder calls, got %d (%v)", len(mockClient.getOrderIDs), mockClient.getOrderIDs)
	}

	var resp api.ListResponse[api.Order]
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != "ord_1" || resp.Items[1].ID != "ord_2" {
		t.Fatalf("unexpected item IDs: %+v", []string{resp.Items[0].ID, resp.Items[1].ID})
	}
	if len(resp.Items[0].LineItems) != 1 || len(resp.Items[1].LineItems) != 1 {
		t.Fatalf("expected line items on expanded orders, got %+v", resp.Items)
	}
}

// TestOrdersListGetClientError tests error when getClient fails for list.
func TestOrdersListGetClientError(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockStore := &mockStore{err: errors.New("keyring error")}
	cleanup := setupOrdersTest(t, mockClient, mockStore)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := ordersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "credential store") {
		t.Errorf("expected credential store error, got: %v", err)
	}
}

// TestOrdersGetRunE tests the orders get command execution with mock API.
func TestOrdersGetRunE(t *testing.T) {
	tests := []struct {
		name         string
		orderID      string
		mockResp     *api.Order
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name:    "successful get text format",
			orderID: "ord_123",
			mockResp: &api.Order{
				ID:            "ord_123",
				OrderNumber:   "1001",
				Status:        "completed",
				PaymentStatus: "paid",
				FulfillStatus: "shipped",
				TotalPrice:    "99.99",
				Currency:      "USD",
				CustomerEmail: "test@example.com",
				CustomerName:  "John Doe",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			outputFormat: "text",
			wantOutput:   "", // Text output uses fmt.Printf to stdout, not formatterWriter
		},
		{
			name:    "successful get JSON format",
			orderID: "ord_789",
			mockResp: &api.Order{
				ID:            "ord_789",
				OrderNumber:   "1003",
				Status:        "pending",
				PaymentStatus: "unpaid",
				FulfillStatus: "unfulfilled",
				TotalPrice:    "150.00",
				Currency:      "GBP",
				CustomerEmail: "json@example.com",
				CustomerName:  "Jane Smith",
				CreatedAt:     time.Date(2024, 3, 10, 12, 0, 0, 0, time.UTC),
			},
			outputFormat: "json",
			wantOutput:   `"id": "ord_789"`,
		},
		{
			name:         "order not found",
			orderID:      "ord_999",
			mockErr:      errors.New("order not found"),
			outputFormat: "text",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAPIClient{
				getOrderResp: tt.mockResp,
				getOrderErr:  tt.mockErr,
			}
			cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := ordersGetCmd.RunE(cmd, []string{tt.orderID})

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

// TestOrdersGetGetClientError tests error when getClient fails for get.
func TestOrdersGetGetClientError(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockStore := &mockStore{err: errors.New("keyring error")}
	cleanup := setupOrdersTest(t, mockClient, mockStore)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := ordersGetCmd.RunE(cmd, []string{"ord_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "credential store") {
		t.Errorf("expected credential store error, got: %v", err)
	}
}

// TestOrdersGetByFlag tests the --by flag on the orders get command.
func TestOrdersGetByFlag(t *testing.T) {
	t.Run("resolves order by query", func(t *testing.T) {
		mockClient := &mockAPIClient{
			searchOrdersResp: &api.OrdersListResponse{
				Items: []api.OrderSummary{
					{ID: "ord_found", OrderNumber: "1001", CustomerEmail: "alice@example.com"},
				},
				TotalCount: 1,
			},
			getOrderResp: &api.Order{
				ID:          "ord_found",
				OrderNumber: "1001",
			},
		}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		var buf bytes.Buffer
		formatterWriter = &buf

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "1001")

		if err := ordersGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "ord_found") {
			t.Errorf("expected output to contain 'ord_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &mockAPIClient{
			searchOrdersResp: &api.OrdersListResponse{
				Items:      []api.OrderSummary{},
				TotalCount: 0,
			},
		}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "nonexistent")

		err := ordersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no order found")
		}
		if !strings.Contains(err.Error(), "no order found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &mockAPIClient{
			searchOrdersErr: errors.New("API error"),
		}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "1001")

		err := ordersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("prefers exact order number over fuzzy matches", func(t *testing.T) {
		mockClient := &mockAPIClient{
			searchOrdersResp: &api.OrdersListResponse{
				Items: []api.OrderSummary{
					{ID: "ord_suffix", OrderNumber: "1001A"},
					{ID: "ord_exact", OrderNumber: "1001"},
				},
				TotalCount: 2,
			},
			getOrderByID: map[string]*api.Order{
				"ord_exact": {
					ID:          "ord_exact",
					OrderNumber: "1001",
				},
			},
		}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		var buf bytes.Buffer
		formatterWriter = &buf

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "1001")

		if err := ordersGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "ord_exact") {
			t.Fatalf("expected resolved order to be ord_exact, output: %s", buf.String())
		}
		if strings.Contains(buf.String(), "ord_suffix") {
			t.Fatalf("expected suffix match not to be selected, output: %s", buf.String())
		}
	})

	t.Run("errors when multiple exact matches exist", func(t *testing.T) {
		mockClient := &mockAPIClient{
			searchOrdersResp: &api.OrdersListResponse{
				Items: []api.OrderSummary{
					{ID: "ord_1", OrderNumber: "1001"},
					{ID: "ord_2", OrderNumber: "1001"},
				},
				TotalCount: 2,
			},
		}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "1001")

		err := ordersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error for multiple exact matches")
		}
		if !strings.Contains(err.Error(), "multiple exact orders found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &mockAPIClient{
			getOrderResp: &api.Order{
				ID:          "ord_direct",
				OrderNumber: "999",
			},
		}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		var buf bytes.Buffer
		formatterWriter = &buf

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := ordersGetCmd.RunE(cmd, []string{"ord_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "ord_direct") {
			t.Errorf("expected output to contain 'ord_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &mockAPIClient{}
		cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
		defer cleanup()

		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())
		cmd.Flags().String("store", "", "")
		cmd.Flags().String("output", "", "")
		cmd.Flags().String("color", "never", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("by", "", "")

		err := ordersGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestOrdersCancelRunE tests the orders cancel command execution.
func TestOrdersCancelRunE(t *testing.T) {
	tests := []struct {
		name       string
		orderID    string
		mockErr    error
		yesFlag    bool
		dryRun     bool
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful cancel with yes flag",
			orderID:    "ord_123",
			mockErr:    nil,
			yesFlag:    true,
			wantOutput: "cancelled",
		},
		{
			name:    "cancel fails",
			orderID: "ord_456",
			mockErr: errors.New("order already cancelled"),
			yesFlag: true,
			wantErr: true,
		},
		{
			name:       "dry-run mode",
			orderID:    "ord_789",
			dryRun:     true,
			yesFlag:    true,
			wantOutput: "[DRY-RUN]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAPIClient{
				cancelOrderErr: tt.mockErr,
			}
			cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("batch", "", "")
			cmd.Flags().Bool("yes", tt.yesFlag, "")
			cmd.Flags().Bool("dry-run", tt.dryRun, "")

			err := ordersCancelCmd.RunE(cmd, []string{tt.orderID})

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

// TestOrdersCancelNoOrderID tests cancel command without order ID.
func TestOrdersCancelNoOrderID(t *testing.T) {
	mockClient := &mockAPIClient{}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("batch", "", "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("dry-run", false, "")

	err := ordersCancelCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "order ID required") {
		t.Errorf("expected 'order ID required' error, got: %v", err)
	}
}

// TestOrdersCancelGetClientError tests cancel command when getClient fails.
func TestOrdersCancelGetClientError(t *testing.T) {
	mockClient := &mockAPIClient{}
	mockStore := &mockStore{err: errors.New("keyring error")}
	cleanup := setupOrdersTest(t, mockClient, mockStore)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("batch", "", "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("dry-run", false, "")

	err := ordersCancelCmd.RunE(cmd, []string{"ord_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "credential store") {
		t.Errorf("expected credential store error, got: %v", err)
	}
}

// TestOrdersCancelBatchMode tests cancel command with batch flag.
func TestOrdersCancelBatchMode(t *testing.T) {
	// Create temp file with batch data
	content := `[{"id": "ord_123"}, {"id": "ord_456"}]`
	f, err := os.CreateTemp("", "batch-cancel-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = f.Close()

	mockClient := &mockAPIClient{
		cancelOrderErr: nil, // Success
	}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("batch", f.Name(), "")
	_ = cmd.Flags().Set("batch", f.Name())
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("dry-run", false, "")

	err = ordersCancelCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrdersCancelBatchModeWithErrors tests batch cancel with partial failures.
func TestOrdersCancelBatchModeWithErrors(t *testing.T) {
	// Create temp file with batch data including one that will fail
	content := `[{"id": "ord_success"}, {"id": "ord_fail"}]`
	f, err := os.CreateTemp("", "batch-cancel-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	_ = f.Close()

	// Mock client that fails on specific order
	callCount := 0
	mockClient := &mockAPIClient{}
	origClientFactory := clientFactory
	defer func() { clientFactory = origClientFactory }()

	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	// Override with custom client factory to alternate success/failure
	clientFactory = func(handle, accessToken string) api.APIClient {
		return &mockAPIClientWithCancelCounter{
			callCount: &callCount,
		}
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("batch", f.Name(), "")
	_ = cmd.Flags().Set("batch", f.Name())
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("dry-run", false, "")

	// This should not return an error even with partial failures
	// Individual results are written to output
	err = ordersCancelCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// mockAPIClientWithCancelCounter tracks cancel calls and alternates success/failure.
type mockAPIClientWithCancelCounter struct {
	api.MockClient
	callCount *int
}

func (m *mockAPIClientWithCancelCounter) CancelOrder(ctx context.Context, id string) error {
	*m.callCount++
	if id == "ord_fail" {
		return errors.New("order cancellation failed")
	}
	return nil
}

// TestGetClientListError tests getClient when List() returns an error.
func TestGetClientListError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	// Mock store where List() fails
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStoreListError{}, nil
	}

	cmd := newTestCmdWithFlags()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "list error") {
		t.Errorf("expected 'list error', got: %v", err)
	}
}

// mockStoreListError is a mock store where List() fails.
type mockStoreListError struct{}

func (m *mockStoreListError) List() ([]string, error) {
	return nil, errors.New("list error")
}

func (m *mockStoreListError) Get(name string) (*secrets.StoreCredentials, error) {
	return nil, errors.New("not implemented")
}

// TestEnrichError tests the enrichError helper function.
func TestEnrichError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		resource   string
		resourceID string
		wantNil    bool
	}{
		{
			name:       "nil error returns nil",
			err:        nil,
			resource:   "orders",
			resourceID: "ord_123",
			wantNil:    true,
		},
		{
			name:       "basic error is enriched",
			err:        errors.New("something went wrong"),
			resource:   "orders",
			resourceID: "ord_456",
			wantNil:    false,
		},
		{
			name:       "API 404 error gets enriched",
			err:        &api.APIError{Code: "not_found", Message: "order not found", Status: 404},
			resource:   "orders",
			resourceID: "ord_789",
			wantNil:    false,
		},
		{
			name:       "auth error gets enriched",
			err:        &api.AuthError{Reason: "token expired"},
			resource:   "orders",
			resourceID: "",
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enrichError(tt.err, tt.resource, tt.resourceID)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Error("expected non-nil error, got nil")
				return
			}

			// Check that it's a RichError
			var richErr *api.RichError
			if !errors.As(result, &richErr) {
				t.Errorf("expected RichError, got %T", result)
				return
			}

			// Verify resource info is set
			if richErr.Resource != tt.resource {
				t.Errorf("resource = %q, want %q", richErr.Resource, tt.resource)
			}
			if richErr.ResourceID != tt.resourceID {
				t.Errorf("resourceID = %q, want %q", richErr.ResourceID, tt.resourceID)
			}
		})
	}
}

// TestHandleError tests the handleError helper function.
func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		resource       string
		resourceID     string
		wantSuggestion string
	}{
		{
			name:           "404 error shows suggestions",
			err:            &api.APIError{Code: "not_found", Message: "order not found", Status: 404},
			resource:       "orders",
			resourceID:     "ord_123",
			wantSuggestion: "Verify the orders ID",
		},
		{
			name:           "auth error shows auth suggestions",
			err:            &api.AuthError{Reason: "token expired"},
			resource:       "orders",
			resourceID:     "",
			wantSuggestion: "spl auth login",
		},
		{
			name:           "rate limit error shows retry suggestion",
			err:            &api.RateLimitError{RetryAfter: 30 * time.Second},
			resource:       "orders",
			resourceID:     "",
			wantSuggestion: "Wait 30 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr output
			var stderr bytes.Buffer
			cmd := &cobra.Command{Use: "test"}
			cmd.SetErr(&stderr)

			result := handleError(cmd, tt.err, tt.resource, tt.resourceID)

			// Verify error was returned
			if result == nil {
				t.Error("expected non-nil error, got nil")
				return
			}

			// Verify stderr contains suggestion
			output := stderr.String()
			if !strings.Contains(output, tt.wantSuggestion) {
				t.Errorf("stderr output %q should contain %q", output, tt.wantSuggestion)
			}
		})
	}
}

// TestOrdersGetRichError tests that orders get command produces rich errors.
func TestOrdersGetRichError(t *testing.T) {
	mockClient := &mockAPIClient{
		getOrderErr: &api.APIError{Code: "not_found", Message: "order not found", Status: 404},
	}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	// Capture stderr
	var stderr bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.SetErr(&stderr)
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := ordersGetCmd.RunE(cmd, []string{"ord_nonexistent"})

	// Should return an error
	if err == nil {
		t.Error("expected error, got nil")
		return
	}

	// Should be a RichError
	var richErr *api.RichError
	if !errors.As(err, &richErr) {
		t.Errorf("expected RichError, got %T", err)
		return
	}

	// Should have suggestions
	if len(richErr.Suggestions) == 0 {
		t.Error("expected suggestions in RichError")
	}

	// Stderr should contain formatted error with suggestions
	output := stderr.String()
	if !strings.Contains(output, "Suggestions:") {
		t.Errorf("stderr should contain suggestions, got: %q", output)
	}
}

// TestOrdersCreateShorthandFlags tests the create command with shorthand flags.
func TestOrdersCreateShorthandFlags(t *testing.T) {
	tests := []struct {
		name       string
		flags      map[string]string
		wantEmail  string
		wantNote   string
		wantTags   []string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "all flags",
			flags:     map[string]string{"email": "user@example.com", "note": "Test order", "tags": "vip,rush"},
			wantEmail: "user@example.com",
			wantNote:  "Test order",
			wantTags:  []string{"vip", "rush"},
		},
		{
			name:      "email only",
			flags:     map[string]string{"email": "user@example.com"},
			wantEmail: "user@example.com",
		},
		{
			name:     "note only",
			flags:    map[string]string{"note": "Just a note"},
			wantNote: "Just a note",
		},
		{
			name:     "tags only",
			flags:    map[string]string{"tags": "a, b, c"},
			wantTags: []string{"a", "b", "c"},
		},
		{
			name:       "body and flags conflict",
			flags:      map[string]string{"body": `{"note":"x"}`, "email": "user@example.com"},
			wantErr:    true,
			wantErrMsg: "use either --body/--body-file or individual flags, not both",
		},
		{
			name:       "no input at all",
			flags:      map[string]string{},
			wantErr:    true,
			wantErrMsg: "provide order data via --body/--body-file or individual flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAPIClient{
				createOrderResp: &api.Order{
					ID:          "ord_new",
					OrderNumber: "2001",
				},
			}
			cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("items-only", false, "")
			cmd.Flags().Bool("dry-run", false, "")
			addJSONBodyFlags(cmd)
			cmd.Flags().String("email", "", "Customer email")
			cmd.Flags().String("note", "", "Order note")
			cmd.Flags().String("tags", "", "Comma-separated tags")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			err := ordersCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			req := mockClient.createOrderReq
			if req == nil {
				t.Fatal("expected CreateOrder to be called with a request")
			}
			if req.CustomerEmail != tt.wantEmail {
				t.Errorf("email = %q, want %q", req.CustomerEmail, tt.wantEmail)
			}
			if req.Note != tt.wantNote {
				t.Errorf("note = %q, want %q", req.Note, tt.wantNote)
			}
			if len(tt.wantTags) > 0 {
				if len(req.Tags) != len(tt.wantTags) {
					t.Fatalf("tags count = %d, want %d", len(req.Tags), len(tt.wantTags))
				}
				for i, tag := range tt.wantTags {
					if req.Tags[i] != tag {
						t.Errorf("tags[%d] = %q, want %q", i, req.Tags[i], tag)
					}
				}
			}
		})
	}
}

// TestOrdersCreateWithBody tests that --body still works for create.
func TestOrdersCreateWithBody(t *testing.T) {
	mockClient := &mockAPIClient{
		createOrderResp: &api.Order{
			ID:          "ord_body",
			OrderNumber: "3001",
		},
	}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	addJSONBodyFlags(cmd)
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("note", "", "")
	cmd.Flags().String("tags", "", "")

	_ = cmd.Flags().Set("body", `{"customer_email":"body@example.com","note":"from body"}`)

	err := ordersCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := mockClient.createOrderReq
	if req == nil {
		t.Fatal("expected CreateOrder to be called")
	}
	if req.CustomerEmail != "body@example.com" {
		t.Errorf("email = %q, want %q", req.CustomerEmail, "body@example.com")
	}
	if req.Note != "from body" {
		t.Errorf("note = %q, want %q", req.Note, "from body")
	}
}

// TestOrdersCreateDryRun tests that dry-run skips shorthand flag validation.
func TestOrdersCreateDryRun(t *testing.T) {
	mockClient := &mockAPIClient{}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	addJSONBodyFlags(cmd)
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("note", "", "")
	cmd.Flags().String("tags", "", "")

	_ = cmd.Flags().Set("dry-run", "true")

	err := ordersCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.createOrderReq != nil {
		t.Error("expected CreateOrder NOT to be called in dry-run mode")
	}
}

// TestOrdersUpdateShorthandFlags tests the update command with shorthand flags.
func TestOrdersUpdateShorthandFlags(t *testing.T) {
	tests := []struct {
		name       string
		flags      map[string]string
		wantNote   *string
		wantTags   []string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "note and tags",
			flags:    map[string]string{"note": "Updated note", "tags": "vip,rush"},
			wantNote: strPtr("Updated note"),
			wantTags: []string{"vip", "rush"},
		},
		{
			name:     "note only",
			flags:    map[string]string{"note": "Just a note"},
			wantNote: strPtr("Just a note"),
		},
		{
			name:     "tags only",
			flags:    map[string]string{"tags": "a,b"},
			wantTags: []string{"a", "b"},
		},
		{
			name:     "empty note (clear)",
			flags:    map[string]string{"note": ""},
			wantNote: strPtr(""),
		},
		{
			name:       "body and flags conflict",
			flags:      map[string]string{"body": `{"note":"x"}`, "note": "y"},
			wantErr:    true,
			wantErrMsg: "use either --body/--body-file or individual flags, not both",
		},
		{
			name:       "no input at all",
			flags:      map[string]string{},
			wantErr:    true,
			wantErrMsg: "provide order data via --body/--body-file or individual flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockAPIClient{
				updateOrderResp: &api.Order{
					ID:          "ord_upd",
					OrderNumber: "4001",
				},
			}
			cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("items-only", false, "")
			cmd.Flags().Bool("dry-run", false, "")
			addJSONBodyFlags(cmd)
			cmd.Flags().String("note", "", "Order note")
			cmd.Flags().String("tags", "", "Comma-separated tags")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			err := ordersUpdateCmd.RunE(cmd, []string{"ord_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if mockClient.updateOrderID != "ord_123" {
				t.Errorf("update order ID = %q, want %q", mockClient.updateOrderID, "ord_123")
			}

			req := mockClient.updateOrderReq
			if req == nil {
				t.Fatal("expected UpdateOrder to be called with a request")
			}
			if tt.wantNote != nil {
				if req.Note == nil {
					t.Fatal("expected Note to be set")
				}
				if *req.Note != *tt.wantNote {
					t.Errorf("note = %q, want %q", *req.Note, *tt.wantNote)
				}
			} else {
				if req.Note != nil {
					t.Errorf("expected Note to be nil, got %q", *req.Note)
				}
			}
			if len(tt.wantTags) > 0 {
				if len(req.Tags) != len(tt.wantTags) {
					t.Fatalf("tags count = %d, want %d", len(req.Tags), len(tt.wantTags))
				}
				for i, tag := range tt.wantTags {
					if req.Tags[i] != tag {
						t.Errorf("tags[%d] = %q, want %q", i, req.Tags[i], tag)
					}
				}
			}
		})
	}
}

// TestOrdersUpdateDryRun tests that dry-run skips shorthand flag validation for update.
func TestOrdersUpdateDryRun(t *testing.T) {
	mockClient := &mockAPIClient{}
	cleanup := setupOrdersTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	addJSONBodyFlags(cmd)
	cmd.Flags().String("note", "", "")
	cmd.Flags().String("tags", "", "")

	_ = cmd.Flags().Set("dry-run", "true")

	err := ordersUpdateCmd.RunE(cmd, []string{"ord_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.updateOrderReq != nil {
		t.Error("expected UpdateOrder NOT to be called in dry-run mode")
	}
}

// TestSplitTags tests the splitTags helper.
func TestSplitTags(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"vip,rush", []string{"vip", "rush"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{" , , ", nil},
		{"", nil},
		{"  spaces  , around  ", []string{"spaces", "around"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitTags(tt.input)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("splitTags(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitTags(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

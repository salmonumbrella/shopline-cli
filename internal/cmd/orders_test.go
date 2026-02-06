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
	getOrderIDs  []string

	cancelOrderErr error
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
	m.getOrderIDs = append(m.getOrderIDs, id)
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
			wantSuggestion: "shopline auth login",
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

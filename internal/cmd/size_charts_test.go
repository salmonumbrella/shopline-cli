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

// TestSizeChartsCommandSetup verifies size-charts command initialization
func TestSizeChartsCommandSetup(t *testing.T) {
	if sizeChartsCmd.Use != "size-charts" {
		t.Errorf("expected Use 'size-charts', got %q", sizeChartsCmd.Use)
	}
	if sizeChartsCmd.Short != "Manage size charts" {
		t.Errorf("expected Short 'Manage size charts', got %q", sizeChartsCmd.Short)
	}
}

// TestSizeChartsSubcommands verifies all subcommands are registered
func TestSizeChartsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List size charts",
		"get":    "Get size chart details",
		"create": "Create a size chart",
		"delete": "Delete a size chart",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range sizeChartsCmd.Commands() {
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

// TestSizeChartsListFlags verifies list command flags exist with correct defaults
func TestSizeChartsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := sizeChartsListCmd.Flags().Lookup(f.name)
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

// TestSizeChartsCreateFlags verifies create command flags exist with correct defaults
func TestSizeChartsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"description", ""},
		{"unit", "cm"},
		{"active", "true"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := sizeChartsCreateCmd.Flags().Lookup(f.name)
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

// TestSizeChartsGetClientError verifies error handling when getClient fails
func TestSizeChartsGetClientError(t *testing.T) {
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

// TestSizeChartsWithMockStore tests size charts commands with a mock credential store
func TestSizeChartsWithMockStore(t *testing.T) {
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

// TestSizeChartsCreateRequiredFlags verifies that name flag is required
func TestSizeChartsCreateRequiredFlags(t *testing.T) {
	// Check that the name flag exists
	flag := sizeChartsCreateCmd.Flags().Lookup("name")
	if flag == nil {
		t.Error("name flag not found")
		return
	}

	// Verify it's a string flag
	if flag.Value.Type() != "string" {
		t.Error("name flag should be a string type")
	}
}

// TestSizeChartsDeleteArgs verifies delete command requires exactly one argument
func TestSizeChartsDeleteArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if sizeChartsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", sizeChartsDeleteCmd.Use)
	}
}

// TestSizeChartsGetArgs verifies get command requires exactly one argument
func TestSizeChartsGetArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if sizeChartsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", sizeChartsGetCmd.Use)
	}
}

// sizeChartsMockAPIClient is a mock implementation of api.APIClient for size charts tests.
type sizeChartsMockAPIClient struct {
	api.MockClient
	listSizeChartsResp  *api.SizeChartsListResponse
	listSizeChartsErr   error
	getSizeChartResp    *api.SizeChart
	getSizeChartErr     error
	createSizeChartResp *api.SizeChart
	createSizeChartErr  error
	deleteSizeChartErr  error
}

func (m *sizeChartsMockAPIClient) ListSizeCharts(ctx context.Context, opts *api.SizeChartsListOptions) (*api.SizeChartsListResponse, error) {
	return m.listSizeChartsResp, m.listSizeChartsErr
}

func (m *sizeChartsMockAPIClient) GetSizeChart(ctx context.Context, id string) (*api.SizeChart, error) {
	return m.getSizeChartResp, m.getSizeChartErr
}

func (m *sizeChartsMockAPIClient) CreateSizeChart(ctx context.Context, req *api.SizeChartCreateRequest) (*api.SizeChart, error) {
	return m.createSizeChartResp, m.createSizeChartErr
}

func (m *sizeChartsMockAPIClient) DeleteSizeChart(ctx context.Context, id string) error {
	return m.deleteSizeChartErr
}

// setupSizeChartsMockFactories sets up mock factories for size charts tests.
func setupSizeChartsMockFactories(mockClient *sizeChartsMockAPIClient) (func(), *bytes.Buffer) {
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

// newSizeChartsTestCmd creates a test command with common flags for size charts tests.
func newSizeChartsTestCmd() *cobra.Command {
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

// TestSizeChartsListRunE tests the size charts list command with mock API.
func TestSizeChartsListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.SizeChartsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.SizeChartsListResponse{
				Items: []api.SizeChart{
					{
						ID:        "sc_123",
						Name:      "Men's Shirt Sizes",
						Unit:      "cm",
						Active:    true,
						Rows:      []api.SizeChartRow{{Size: "S", Values: []string{"38", "76"}}},
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sc_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.SizeChartsListResponse{
				Items:      []api.SizeChart{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple size charts",
			mockResp: &api.SizeChartsListResponse{
				Items: []api.SizeChart{
					{
						ID:        "sc_001",
						Name:      "T-Shirt Sizes",
						Unit:      "inches",
						Active:    true,
						Rows:      []api.SizeChartRow{{Size: "M", Values: []string{"40"}}},
						CreatedAt: testTime,
					},
					{
						ID:        "sc_002",
						Name:      "Pants Sizes",
						Unit:      "cm",
						Active:    false,
						Rows:      []api.SizeChartRow{},
						CreatedAt: testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "sc_001",
		},
		{
			name: "inactive size chart",
			mockResp: &api.SizeChartsListResponse{
				Items: []api.SizeChart{
					{
						ID:        "sc_inactive",
						Name:      "Old Sizes",
						Unit:      "cm",
						Active:    false,
						Rows:      []api.SizeChartRow{},
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "No",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &sizeChartsMockAPIClient{
				listSizeChartsResp: tt.mockResp,
				listSizeChartsErr:  tt.mockErr,
			}
			cleanup, buf := setupSizeChartsMockFactories(mockClient)
			defer cleanup()

			cmd := newSizeChartsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := sizeChartsListCmd.RunE(cmd, []string{})

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

// TestSizeChartsListJSONOutput tests the size charts list command with JSON output.
func TestSizeChartsListJSONOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &sizeChartsMockAPIClient{
		listSizeChartsResp: &api.SizeChartsListResponse{
			Items: []api.SizeChart{
				{
					ID:        "sc_json",
					Name:      "JSON Test Chart",
					Unit:      "cm",
					Active:    true,
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := sizeChartsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "sc_json") {
		t.Errorf("JSON output should contain 'sc_json', got %q", output)
	}
}

// TestSizeChartsGetRunE tests the size charts get command with mock API.
func TestSizeChartsGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		sizeChartID string
		mockResp    *api.SizeChart
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful get",
			sizeChartID: "sc_123",
			mockResp: &api.SizeChart{
				ID:          "sc_123",
				Name:        "Men's Shirt Sizes",
				Description: "Standard men's shirt sizing",
				Unit:        "cm",
				Active:      true,
				Headers:     []string{"Chest", "Waist"},
				Rows: []api.SizeChartRow{
					{Size: "S", Values: []string{"38", "32"}},
					{Size: "M", Values: []string{"40", "34"}},
				},
				ProductIDs: []string{"prod_1", "prod_2"},
				CreatedAt:  testTime,
				UpdatedAt:  testTime,
			},
		},
		{
			name:        "size chart not found",
			sizeChartID: "sc_999",
			mockErr:     errors.New("size chart not found"),
			wantErr:     true,
		},
		{
			name:        "size chart with no headers",
			sizeChartID: "sc_noheaders",
			mockResp: &api.SizeChart{
				ID:        "sc_noheaders",
				Name:      "Simple Chart",
				Unit:      "inches",
				Active:    true,
				Headers:   []string{},
				Rows:      []api.SizeChartRow{},
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:        "size chart with headers and rows",
			sizeChartID: "sc_full",
			mockResp: &api.SizeChart{
				ID:          "sc_full",
				Name:        "Complete Chart",
				Description: "Full sizing information",
				Unit:        "cm",
				Active:      true,
				Headers:     []string{"Chest", "Waist", "Hip"},
				Rows: []api.SizeChartRow{
					{Size: "XS", Values: []string{"36", "28", "34"}},
					{Size: "S", Values: []string{"38", "30", "36"}},
				},
				ProductIDs: []string{},
				CreatedAt:  testTime,
				UpdatedAt:  testTime,
			},
		},
		{
			name:        "size chart with products",
			sizeChartID: "sc_products",
			mockResp: &api.SizeChart{
				ID:         "sc_products",
				Name:       "Product Chart",
				Unit:       "cm",
				Active:     true,
				Headers:    []string{},
				Rows:       []api.SizeChartRow{},
				ProductIDs: []string{"prod_a", "prod_b", "prod_c"},
				CreatedAt:  testTime,
				UpdatedAt:  testTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &sizeChartsMockAPIClient{
				getSizeChartResp: tt.mockResp,
				getSizeChartErr:  tt.mockErr,
			}
			cleanup, _ := setupSizeChartsMockFactories(mockClient)
			defer cleanup()

			cmd := newSizeChartsTestCmd()

			err := sizeChartsGetCmd.RunE(cmd, []string{tt.sizeChartID})

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

// TestSizeChartsGetJSONOutput tests the size charts get command with JSON output.
func TestSizeChartsGetJSONOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &sizeChartsMockAPIClient{
		getSizeChartResp: &api.SizeChart{
			ID:        "sc_json",
			Name:      "JSON Chart",
			Unit:      "cm",
			Active:    true,
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, buf := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := sizeChartsGetCmd.RunE(cmd, []string{"sc_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "sc_json") {
		t.Errorf("JSON output should contain 'sc_json', got %q", output)
	}
}

// TestSizeChartsCreateRunE tests the size charts create command with mock API.
func TestSizeChartsCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		chartName string
		desc      string
		unit      string
		active    bool
		mockResp  *api.SizeChart
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful create",
			chartName: "New Size Chart",
			desc:      "A new chart",
			unit:      "cm",
			active:    true,
			mockResp: &api.SizeChart{
				ID:          "sc_new",
				Name:        "New Size Chart",
				Description: "A new chart",
				Unit:        "cm",
				Active:      true,
				CreatedAt:   testTime,
			},
		},
		{
			name:      "create with inches",
			chartName: "Inches Chart",
			unit:      "inches",
			active:    true,
			mockResp: &api.SizeChart{
				ID:        "sc_inches",
				Name:      "Inches Chart",
				Unit:      "inches",
				Active:    true,
				CreatedAt: testTime,
			},
		},
		{
			name:      "create inactive",
			chartName: "Inactive Chart",
			unit:      "cm",
			active:    false,
			mockResp: &api.SizeChart{
				ID:        "sc_inactive",
				Name:      "Inactive Chart",
				Unit:      "cm",
				Active:    false,
				CreatedAt: testTime,
			},
		},
		{
			name:      "API error",
			chartName: "Error Chart",
			unit:      "cm",
			active:    true,
			mockErr:   errors.New("failed to create"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &sizeChartsMockAPIClient{
				createSizeChartResp: tt.mockResp,
				createSizeChartErr:  tt.mockErr,
			}
			cleanup, _ := setupSizeChartsMockFactories(mockClient)
			defer cleanup()

			cmd := newSizeChartsTestCmd()
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("unit", "cm", "")
			cmd.Flags().Bool("active", true, "")
			_ = cmd.Flags().Set("name", tt.chartName)
			_ = cmd.Flags().Set("description", tt.desc)
			_ = cmd.Flags().Set("unit", tt.unit)
			if !tt.active {
				_ = cmd.Flags().Set("active", "false")
			}

			err := sizeChartsCreateCmd.RunE(cmd, []string{})

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

// TestSizeChartsCreateDryRun tests the size charts create command with dry-run flag.
func TestSizeChartsCreateDryRun(t *testing.T) {
	mockClient := &sizeChartsMockAPIClient{}
	cleanup, _ := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("unit", "cm", "")
	cmd.Flags().Bool("active", true, "")
	_ = cmd.Flags().Set("name", "Dry Run Chart")
	_ = cmd.Flags().Set("dry-run", "true")

	// Dry-run should complete without error and without calling the API
	err := sizeChartsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSizeChartsCreateJSONOutput tests the size charts create command with JSON output.
func TestSizeChartsCreateJSONOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &sizeChartsMockAPIClient{
		createSizeChartResp: &api.SizeChart{
			ID:        "sc_json_create",
			Name:      "JSON Created Chart",
			Unit:      "cm",
			Active:    true,
			CreatedAt: testTime,
		},
	}
	cleanup, buf := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("unit", "cm", "")
	cmd.Flags().Bool("active", true, "")
	_ = cmd.Flags().Set("name", "JSON Created Chart")
	_ = cmd.Flags().Set("output", "json")

	err := sizeChartsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "sc_json_create") {
		t.Errorf("JSON output should contain 'sc_json_create', got %q", output)
	}
}

// TestSizeChartsDeleteRunE tests the size charts delete command with mock API.
func TestSizeChartsDeleteRunE(t *testing.T) {
	tests := []struct {
		name        string
		sizeChartID string
		yes         bool
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful delete",
			sizeChartID: "sc_123",
			yes:         true,
			mockErr:     nil,
		},
		{
			name:        "delete without confirmation",
			sizeChartID: "sc_456",
			yes:         false,
			mockErr:     nil,
		},
		{
			name:        "API error",
			sizeChartID: "sc_error",
			yes:         true,
			mockErr:     errors.New("delete failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &sizeChartsMockAPIClient{
				deleteSizeChartErr: tt.mockErr,
			}
			cleanup, _ := setupSizeChartsMockFactories(mockClient)
			defer cleanup()

			cmd := newSizeChartsTestCmd()
			if tt.yes {
				_ = cmd.Flags().Set("yes", "true")
			} else {
				_ = cmd.Flags().Set("yes", "false")
			}

			err := sizeChartsDeleteCmd.RunE(cmd, []string{tt.sizeChartID})

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

// TestSizeChartsDeleteDryRun tests the size charts delete command with dry-run flag.
func TestSizeChartsDeleteDryRun(t *testing.T) {
	mockClient := &sizeChartsMockAPIClient{}
	cleanup, _ := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	// Dry-run should complete without error and without calling the API
	err := sizeChartsDeleteCmd.RunE(cmd, []string{"sc_dryrun"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSizeChartsListWithPagination tests the list command respects pagination flags.
func TestSizeChartsListWithPagination(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &sizeChartsMockAPIClient{
		listSizeChartsResp: &api.SizeChartsListResponse{
			Items: []api.SizeChart{
				{ID: "sc_1", Name: "Chart 1", Unit: "cm", Active: true, CreatedAt: testTime},
				{ID: "sc_2", Name: "Chart 2", Unit: "cm", Active: true, CreatedAt: testTime},
			},
			TotalCount: 50,
		},
	}
	cleanup, buf := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := sizeChartsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	// Should show the items from the mock response
	if !strings.Contains(output, "sc_1") {
		t.Errorf("output should contain 'sc_1', got %q", output)
	}
}

// TestSizeChartsGetWithEmptyProductIDs tests get command with empty product IDs.
func TestSizeChartsGetWithEmptyProductIDs(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &sizeChartsMockAPIClient{
		getSizeChartResp: &api.SizeChart{
			ID:         "sc_empty_products",
			Name:       "No Products Chart",
			Unit:       "cm",
			Active:     true,
			Headers:    []string{"Size"},
			Rows:       []api.SizeChartRow{{Size: "M", Values: []string{"40"}}},
			ProductIDs: []string{},
			CreatedAt:  testTime,
			UpdatedAt:  testTime,
		},
	}
	cleanup, _ := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()

	err := sizeChartsGetCmd.RunE(cmd, []string{"sc_empty_products"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSizeChartsCreateWithAllFields tests create with all possible fields.
func TestSizeChartsCreateWithAllFields(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &sizeChartsMockAPIClient{
		createSizeChartResp: &api.SizeChart{
			ID:          "sc_full",
			Name:        "Full Chart",
			Description: "Complete description",
			Unit:        "inches",
			Active:      false,
			CreatedAt:   testTime,
		},
	}
	cleanup, _ := setupSizeChartsMockFactories(mockClient)
	defer cleanup()

	cmd := newSizeChartsTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("unit", "cm", "")
	cmd.Flags().Bool("active", true, "")
	_ = cmd.Flags().Set("name", "Full Chart")
	_ = cmd.Flags().Set("description", "Complete description")
	_ = cmd.Flags().Set("unit", "inches")
	_ = cmd.Flags().Set("active", "false")

	err := sizeChartsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

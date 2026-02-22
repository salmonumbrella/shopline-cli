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

// TestSalesCommandSetup verifies sales command initialization
func TestSalesCommandSetup(t *testing.T) {
	if salesCmd.Use != "sales" {
		t.Errorf("expected Use 'sales', got %q", salesCmd.Use)
	}
	if salesCmd.Short != "Manage sale campaigns" {
		t.Errorf("expected Short 'Manage sale campaigns', got %q", salesCmd.Short)
	}
}

// TestSalesSubcommands verifies all subcommands are registered
func TestSalesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":            "List sale campaigns",
		"get":             "Get sale details",
		"create":          "Create a sale campaign",
		"activate":        "Activate a sale campaign",
		"deactivate":      "Deactivate a sale campaign",
		"delete":          "Delete a sale campaign",
		"delete-products": "Delete products from a sale",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range salesCmd.Commands() {
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

// TestSalesListFlags verifies list command flags exist with correct defaults
func TestSalesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := salesListCmd.Flags().Lookup(f.name)
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

// TestSalesCreateFlags verifies create command flags exist with correct defaults
func TestSalesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"description", ""},
		{"discount-type", ""},
		{"discount-value", "0"},
		{"applies-to", "all"},
		{"product-ids", "[]"},
		{"collection-ids", "[]"},
		{"starts-at", ""},
		{"ends-at", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := salesCreateCmd.Flags().Lookup(f.name)
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

// TestSalesDeleteFlags verifies delete command flags exist
func TestSalesDeleteFlags(t *testing.T) {
	flag := salesDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

func TestSalesDeleteProductsFlags(t *testing.T) {
	flag := salesDeleteProductsCmd.Flags().Lookup("product-ids")
	if flag == nil {
		t.Error("flag 'product-ids' not found")
		return
	}
	if flag.DefValue != "[]" {
		t.Errorf("expected default '[]', got %q", flag.DefValue)
	}
}

// TestSalesGetCmd verifies get command setup
func TestSalesGetCmd(t *testing.T) {
	if salesGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", salesGetCmd.Use)
	}
}

// TestSalesCreateCmd verifies create command setup
func TestSalesCreateCmd(t *testing.T) {
	if salesCreateCmd.Use != "create" {
		t.Errorf("expected Use 'create', got %q", salesCreateCmd.Use)
	}
}

// TestSalesActivateCmd verifies activate command setup
func TestSalesActivateCmd(t *testing.T) {
	if salesActivateCmd.Use != "activate <id>" {
		t.Errorf("expected Use 'activate <id>', got %q", salesActivateCmd.Use)
	}
}

// TestSalesDeactivateCmd verifies deactivate command setup
func TestSalesDeactivateCmd(t *testing.T) {
	if salesDeactivateCmd.Use != "deactivate <id>" {
		t.Errorf("expected Use 'deactivate <id>', got %q", salesDeactivateCmd.Use)
	}
}

// TestSalesDeleteCmd verifies delete command setup
func TestSalesDeleteCmd(t *testing.T) {
	if salesDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", salesDeleteCmd.Use)
	}
}

// TestSalesListRunE_GetClientFails verifies error handling when getClient fails
func TestSalesListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")

	err := salesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestSalesGetRunE_GetClientFails verifies error handling when getClient fails
func TestSalesGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := salesGetCmd.RunE(cmd, []string{"sale_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestSalesCreateRunE_GetClientFails verifies error handling when getClient fails
func TestSalesCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("title", "Test Sale", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Float64("discount-value", 10.0, "")
	cmd.Flags().String("applies-to", "all", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().StringSlice("collection-ids", nil, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")

	err := salesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestSalesActivateRunE_GetClientFails verifies error handling when getClient fails
func TestSalesActivateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := salesActivateCmd.RunE(cmd, []string{"sale_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestSalesDeactivateRunE_GetClientFails verifies error handling when getClient fails
func TestSalesDeactivateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := salesDeactivateCmd.RunE(cmd, []string{"sale_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestSalesDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestSalesDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := salesDeleteCmd.RunE(cmd, []string{"sale_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestSalesListRunE_NoProfiles verifies error handling when no profiles exist
func TestSalesListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")

	err := salesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// salesMockAPIClient is a mock implementation of api.APIClient for sales tests.
type salesMockAPIClient struct {
	api.MockClient
	listSalesResp            *api.SalesListResponse
	listSalesErr             error
	getSaleResp              *api.Sale
	getSaleErr               error
	createSaleResp           *api.Sale
	createSaleErr            error
	activateSaleResp         *api.Sale
	activateSaleErr          error
	deactivateSaleResp       *api.Sale
	deactivateSaleErr        error
	deleteSaleErr            error
	deleteSaleProductsSaleID string
	deleteSaleProductsReq    *api.SaleDeleteProductsRequest
	deleteSaleProductsErr    error
}

func (m *salesMockAPIClient) ListSales(ctx context.Context, opts *api.SalesListOptions) (*api.SalesListResponse, error) {
	return m.listSalesResp, m.listSalesErr
}

func (m *salesMockAPIClient) GetSale(ctx context.Context, id string) (*api.Sale, error) {
	return m.getSaleResp, m.getSaleErr
}

func (m *salesMockAPIClient) CreateSale(ctx context.Context, req *api.SaleCreateRequest) (*api.Sale, error) {
	return m.createSaleResp, m.createSaleErr
}

func (m *salesMockAPIClient) ActivateSale(ctx context.Context, id string) (*api.Sale, error) {
	return m.activateSaleResp, m.activateSaleErr
}

func (m *salesMockAPIClient) DeactivateSale(ctx context.Context, id string) (*api.Sale, error) {
	return m.deactivateSaleResp, m.deactivateSaleErr
}

func (m *salesMockAPIClient) DeleteSale(ctx context.Context, id string) error {
	return m.deleteSaleErr
}

func (m *salesMockAPIClient) DeleteSaleProducts(ctx context.Context, saleID string, req *api.SaleDeleteProductsRequest) error {
	m.deleteSaleProductsSaleID = saleID
	m.deleteSaleProductsReq = req
	return m.deleteSaleProductsErr
}

// setupSalesMockFactories sets up mock factories for sales tests.
func setupSalesMockFactories(mockClient *salesMockAPIClient) (func(), *bytes.Buffer) {
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

// newSalesTestCmd creates a test command with common flags for sales tests.
func newSalesTestCmd() *cobra.Command {
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

// TestSalesListRunE tests the sales list command with mock API.
func TestSalesListRunE(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		mockResp   *api.SalesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with percentage discount",
			mockResp: &api.SalesListResponse{
				Items: []api.Sale{
					{
						ID:            "sale_123",
						Title:         "Summer Sale",
						DiscountType:  "percentage",
						DiscountValue: 20,
						AppliesTo:     "all",
						Status:        "active",
						StartsAt:      now,
						EndsAt:        now.Add(24 * time.Hour),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sale_123",
		},
		{
			name: "successful list with fixed amount discount",
			mockResp: &api.SalesListResponse{
				Items: []api.Sale{
					{
						ID:            "sale_456",
						Title:         "Winter Sale",
						DiscountType:  "fixed_amount",
						DiscountValue: 50,
						AppliesTo:     "products",
						Status:        "scheduled",
						StartsAt:      now.Add(24 * time.Hour),
						EndsAt:        now.Add(48 * time.Hour),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sale_456",
		},
		{
			name: "successful list with zero dates",
			mockResp: &api.SalesListResponse{
				Items: []api.Sale{
					{
						ID:            "sale_789",
						Title:         "Ongoing Sale",
						DiscountType:  "percentage",
						DiscountValue: 10,
						AppliesTo:     "collections",
						Status:        "active",
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sale_789",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.SalesListResponse{
				Items:      []api.Sale{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &salesMockAPIClient{
				listSalesResp: tt.mockResp,
				listSalesErr:  tt.mockErr,
			}
			cleanup, buf := setupSalesMockFactories(mockClient)
			defer cleanup()

			cmd := newSalesTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")

			err := salesListCmd.RunE(cmd, []string{})

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

// TestSalesListRunE_JSONOutput tests the sales list command with JSON output.
func TestSalesListRunE_JSONOutput(t *testing.T) {
	mockClient := &salesMockAPIClient{
		listSalesResp: &api.SalesListResponse{
			Items: []api.Sale{
				{
					ID:            "sale_json",
					Title:         "JSON Sale",
					DiscountType:  "percentage",
					DiscountValue: 15,
					Status:        "active",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := salesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sale_json") {
		t.Errorf("JSON output should contain sale_json, got %q", output)
	}
}

// TestSalesGetRunE tests the sales get command with mock API.
func TestSalesGetRunE(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		saleID   string
		mockResp *api.Sale
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful get with all fields",
			saleID: "sale_123",
			mockResp: &api.Sale{
				ID:            "sale_123",
				Title:         "Test Sale",
				Description:   "A test sale",
				DiscountType:  "percentage",
				DiscountValue: 20,
				AppliesTo:     "products",
				ProductIDs:    []string{"prod_1", "prod_2"},
				CollectionIDs: []string{"col_1"},
				Status:        "active",
				StartsAt:      now,
				EndsAt:        now.Add(24 * time.Hour),
				CreatedAt:     now.Add(-48 * time.Hour),
			},
		},
		{
			name:   "successful get with minimal fields",
			saleID: "sale_456",
			mockResp: &api.Sale{
				ID:            "sale_456",
				Title:         "Minimal Sale",
				DiscountType:  "fixed_amount",
				DiscountValue: 100,
				AppliesTo:     "all",
				Status:        "inactive",
				CreatedAt:     now,
			},
		},
		{
			name:    "sale not found",
			saleID:  "sale_999",
			mockErr: errors.New("sale not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &salesMockAPIClient{
				getSaleResp: tt.mockResp,
				getSaleErr:  tt.mockErr,
			}
			cleanup, _ := setupSalesMockFactories(mockClient)
			defer cleanup()

			cmd := newSalesTestCmd()

			err := salesGetCmd.RunE(cmd, []string{tt.saleID})

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

// TestSalesGetRunE_JSONOutput tests the sales get command with JSON output.
func TestSalesGetRunE_JSONOutput(t *testing.T) {
	mockClient := &salesMockAPIClient{
		getSaleResp: &api.Sale{
			ID:            "sale_json",
			Title:         "JSON Sale",
			DiscountType:  "percentage",
			DiscountValue: 25,
			AppliesTo:     "all",
			Status:        "active",
			CreatedAt:     time.Now(),
		},
	}
	cleanup, buf := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := salesGetCmd.RunE(cmd, []string{"sale_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sale_json") {
		t.Errorf("JSON output should contain sale_json, got %q", output)
	}
}

// TestSalesCreateRunE tests the sales create command with mock API.
func TestSalesCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		mockResp *api.Sale
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful create",
			title: "New Sale",
			mockResp: &api.Sale{
				ID:            "sale_new",
				Title:         "New Sale",
				DiscountType:  "percentage",
				DiscountValue: 30,
				AppliesTo:     "all",
				Status:        "scheduled",
			},
		},
		{
			name:    "API error",
			title:   "Failed Sale",
			mockErr: errors.New("validation failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &salesMockAPIClient{
				createSaleResp: tt.mockResp,
				createSaleErr:  tt.mockErr,
			}
			cleanup, _ := setupSalesMockFactories(mockClient)
			defer cleanup()

			cmd := newSalesTestCmd()
			cmd.Flags().String("title", tt.title, "")
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("discount-type", "percentage", "")
			cmd.Flags().Float64("discount-value", 30, "")
			cmd.Flags().String("applies-to", "all", "")
			cmd.Flags().StringSlice("product-ids", nil, "")
			cmd.Flags().StringSlice("collection-ids", nil, "")
			cmd.Flags().String("starts-at", "", "")
			cmd.Flags().String("ends-at", "", "")

			err := salesCreateCmd.RunE(cmd, []string{})

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

// TestSalesCreateRunE_JSONOutput tests the sales create command with JSON output.
func TestSalesCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &salesMockAPIClient{
		createSaleResp: &api.Sale{
			ID:            "sale_created_json",
			Title:         "Created Sale",
			DiscountType:  "percentage",
			DiscountValue: 20,
			AppliesTo:     "all",
			Status:        "scheduled",
		},
	}
	cleanup, buf := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().String("title", "Created Sale", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Float64("discount-value", 20, "")
	cmd.Flags().String("applies-to", "all", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().StringSlice("collection-ids", nil, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := salesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sale_created_json") {
		t.Errorf("JSON output should contain sale_created_json, got %q", output)
	}
}

// TestSalesCreateRunE_WithDates tests the sales create command with start/end dates.
func TestSalesCreateRunE_WithDates(t *testing.T) {
	mockClient := &salesMockAPIClient{
		createSaleResp: &api.Sale{
			ID:            "sale_dated",
			Title:         "Dated Sale",
			DiscountType:  "percentage",
			DiscountValue: 15,
			AppliesTo:     "all",
			Status:        "scheduled",
		},
	}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().String("title", "Dated Sale", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Float64("discount-value", 15, "")
	cmd.Flags().String("applies-to", "all", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().StringSlice("collection-ids", nil, "")
	cmd.Flags().String("starts-at", "2024-06-01T00:00:00Z", "")
	cmd.Flags().String("ends-at", "2024-06-30T23:59:59Z", "")

	err := salesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSalesCreateRunE_InvalidStartsAt tests the sales create command with invalid starts-at.
func TestSalesCreateRunE_InvalidStartsAt(t *testing.T) {
	mockClient := &salesMockAPIClient{}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().String("title", "Invalid Date Sale", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Float64("discount-value", 10, "")
	cmd.Flags().String("applies-to", "all", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().StringSlice("collection-ids", nil, "")
	cmd.Flags().String("starts-at", "invalid-date", "")
	cmd.Flags().String("ends-at", "", "")

	err := salesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid starts-at, got nil")
	}
	if !strings.Contains(err.Error(), "starts-at") {
		t.Errorf("error should mention starts-at, got %q", err.Error())
	}
}

// TestSalesCreateRunE_InvalidEndsAt tests the sales create command with invalid ends-at.
func TestSalesCreateRunE_InvalidEndsAt(t *testing.T) {
	mockClient := &salesMockAPIClient{}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().String("title", "Invalid End Date Sale", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Float64("discount-value", 10, "")
	cmd.Flags().String("applies-to", "all", "")
	cmd.Flags().StringSlice("product-ids", nil, "")
	cmd.Flags().StringSlice("collection-ids", nil, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "not-a-date", "")

	err := salesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid ends-at, got nil")
	}
	if !strings.Contains(err.Error(), "ends-at") {
		t.Errorf("error should mention ends-at, got %q", err.Error())
	}
}

// TestSalesActivateRunE tests the sales activate command with mock API.
func TestSalesActivateRunE(t *testing.T) {
	tests := []struct {
		name     string
		saleID   string
		mockResp *api.Sale
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful activate",
			saleID: "sale_123",
			mockResp: &api.Sale{
				ID:     "sale_123",
				Title:  "Activated Sale",
				Status: "active",
			},
		},
		{
			name:    "API error",
			saleID:  "sale_999",
			mockErr: errors.New("sale not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &salesMockAPIClient{
				activateSaleResp: tt.mockResp,
				activateSaleErr:  tt.mockErr,
			}
			cleanup, _ := setupSalesMockFactories(mockClient)
			defer cleanup()

			cmd := newSalesTestCmd()

			err := salesActivateCmd.RunE(cmd, []string{tt.saleID})

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

// TestSalesDeactivateRunE tests the sales deactivate command with mock API.
func TestSalesDeactivateRunE(t *testing.T) {
	tests := []struct {
		name     string
		saleID   string
		mockResp *api.Sale
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "successful deactivate",
			saleID: "sale_123",
			mockResp: &api.Sale{
				ID:     "sale_123",
				Title:  "Deactivated Sale",
				Status: "inactive",
			},
		},
		{
			name:    "API error",
			saleID:  "sale_999",
			mockErr: errors.New("sale not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &salesMockAPIClient{
				deactivateSaleResp: tt.mockResp,
				deactivateSaleErr:  tt.mockErr,
			}
			cleanup, _ := setupSalesMockFactories(mockClient)
			defer cleanup()

			cmd := newSalesTestCmd()

			err := salesDeactivateCmd.RunE(cmd, []string{tt.saleID})

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

// TestSalesDeleteRunE tests the sales delete command with mock API.
func TestSalesDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		saleID  string
		mockErr error
		wantErr bool
	}{
		{
			name:   "successful delete with --yes flag",
			saleID: "sale_123",
		},
		{
			name:    "API error",
			saleID:  "sale_999",
			mockErr: errors.New("sale not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &salesMockAPIClient{
				deleteSaleErr: tt.mockErr,
			}
			cleanup, _ := setupSalesMockFactories(mockClient)
			defer cleanup()

			cmd := newSalesTestCmd()

			err := salesDeleteCmd.RunE(cmd, []string{tt.saleID})

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

// TestSalesDeleteRunE_WithoutYesFlag tests the delete command without --yes flag.
func TestSalesDeleteRunE_WithoutYesFlag(t *testing.T) {
	mockClient := &salesMockAPIClient{}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	_ = cmd.Flags().Set("yes", "false")

	// When --yes is false and no input is provided, the command should cancel
	// Since Scanln will fail or return empty, the command should print "Cancelled."
	err := salesDeleteCmd.RunE(cmd, []string{"sale_123"})
	// The command should succeed (cancellation is not an error)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSalesGetRunE_WithProductIDs tests output when sale has product IDs.
func TestSalesGetRunE_WithProductIDs(t *testing.T) {
	mockClient := &salesMockAPIClient{
		getSaleResp: &api.Sale{
			ID:            "sale_products",
			Title:         "Products Sale",
			DiscountType:  "percentage",
			DiscountValue: 10,
			AppliesTo:     "products",
			ProductIDs:    []string{"prod_1", "prod_2", "prod_3"},
			Status:        "active",
			CreatedAt:     time.Now(),
		},
	}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()

	err := salesGetCmd.RunE(cmd, []string{"sale_products"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSalesGetRunE_WithCollectionIDs tests output when sale has collection IDs.
func TestSalesGetRunE_WithCollectionIDs(t *testing.T) {
	mockClient := &salesMockAPIClient{
		getSaleResp: &api.Sale{
			ID:            "sale_collections",
			Title:         "Collections Sale",
			DiscountType:  "fixed_amount",
			DiscountValue: 50,
			AppliesTo:     "collections",
			CollectionIDs: []string{"col_1", "col_2"},
			Status:        "active",
			CreatedAt:     time.Now(),
		},
	}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()

	err := salesGetCmd.RunE(cmd, []string{"sale_collections"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSalesGetRunE_WithDates tests output when sale has start/end dates.
func TestSalesGetRunE_WithDates(t *testing.T) {
	now := time.Now()
	mockClient := &salesMockAPIClient{
		getSaleResp: &api.Sale{
			ID:            "sale_dated",
			Title:         "Dated Sale",
			DiscountType:  "percentage",
			DiscountValue: 25,
			AppliesTo:     "all",
			Status:        "scheduled",
			StartsAt:      now.Add(24 * time.Hour),
			EndsAt:        now.Add(48 * time.Hour),
			CreatedAt:     now,
		},
	}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()

	err := salesGetCmd.RunE(cmd, []string{"sale_dated"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSalesListRunE_WithStatusFilter tests the list command with status filter.
func TestSalesListRunE_WithStatusFilter(t *testing.T) {
	mockClient := &salesMockAPIClient{
		listSalesResp: &api.SalesListResponse{
			Items: []api.Sale{
				{
					ID:            "sale_active",
					Title:         "Active Sale",
					DiscountType:  "percentage",
					DiscountValue: 15,
					Status:        "active",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "active", "")

	err := salesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sale_active") {
		t.Errorf("output should contain sale_active, got %q", output)
	}
}

// TestSalesCreateRunE_WithProductsAndCollections tests create with product and collection IDs.
func TestSalesCreateRunE_WithProductsAndCollections(t *testing.T) {
	mockClient := &salesMockAPIClient{
		createSaleResp: &api.Sale{
			ID:            "sale_with_targets",
			Title:         "Targeted Sale",
			DiscountType:  "percentage",
			DiscountValue: 20,
			AppliesTo:     "products",
			ProductIDs:    []string{"prod_1", "prod_2"},
			CollectionIDs: []string{"col_1"},
			Status:        "scheduled",
		},
	}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().String("title", "Targeted Sale", "")
	cmd.Flags().String("description", "A sale with targets", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Float64("discount-value", 20, "")
	cmd.Flags().String("applies-to", "products", "")
	cmd.Flags().StringSlice("product-ids", []string{"prod_1", "prod_2"}, "")
	cmd.Flags().StringSlice("collection-ids", []string{"col_1"}, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")

	err := salesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSalesDeleteProductsRunE_JSON(t *testing.T) {
	mockClient := &salesMockAPIClient{}
	cleanup, buf := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().StringSlice("product-ids", nil, "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("product-ids", "prod_1,prod_2")

	err := salesDeleteProductsCmd.RunE(cmd, []string{"sale_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.deleteSaleProductsSaleID != "sale_123" {
		t.Fatalf("expected saleID sale_123, got %q", mockClient.deleteSaleProductsSaleID)
	}
	if mockClient.deleteSaleProductsReq == nil || len(mockClient.deleteSaleProductsReq.ProductIDs) != 2 {
		t.Fatalf("expected product ids in request, got %+v", mockClient.deleteSaleProductsReq)
	}

	out := buf.String()
	if !strings.Contains(out, "\"ok\"") {
		t.Fatalf("expected ok in JSON output, got %q", out)
	}
}

func TestSalesDeleteProductsRunE_DryRun(t *testing.T) {
	mockClient := &salesMockAPIClient{}
	cleanup, _ := setupSalesMockFactories(mockClient)
	defer cleanup()

	cmd := newSalesTestCmd()
	cmd.Flags().StringSlice("product-ids", nil, "")
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("product-ids", "prod_1")

	err := salesDeleteProductsCmd.RunE(cmd, []string{"sale_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockClient.deleteSaleProductsSaleID != "" {
		t.Fatalf("expected API not called on dry-run, got saleID=%q", mockClient.deleteSaleProductsSaleID)
	}
}

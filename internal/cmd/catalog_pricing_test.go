package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// catalogPricingMockAPIClient is a mock implementation of api.APIClient for catalog pricing tests.
type catalogPricingMockAPIClient struct {
	api.MockClient
	listCatalogPricingResp   *api.CatalogPricingListResponse
	listCatalogPricingErr    error
	getCatalogPricingResp    *api.CatalogPricing
	getCatalogPricingErr     error
	createCatalogPricingResp *api.CatalogPricing
	createCatalogPricingErr  error
	updateCatalogPricingResp *api.CatalogPricing
	updateCatalogPricingErr  error
	deleteCatalogPricingErr  error
}

func (m *catalogPricingMockAPIClient) ListCatalogPricing(ctx context.Context, opts *api.CatalogPricingListOptions) (*api.CatalogPricingListResponse, error) {
	return m.listCatalogPricingResp, m.listCatalogPricingErr
}

func (m *catalogPricingMockAPIClient) GetCatalogPricing(ctx context.Context, id string) (*api.CatalogPricing, error) {
	return m.getCatalogPricingResp, m.getCatalogPricingErr
}

func (m *catalogPricingMockAPIClient) CreateCatalogPricing(ctx context.Context, req *api.CatalogPricingCreateRequest) (*api.CatalogPricing, error) {
	return m.createCatalogPricingResp, m.createCatalogPricingErr
}

func (m *catalogPricingMockAPIClient) UpdateCatalogPricing(ctx context.Context, id string, req *api.CatalogPricingUpdateRequest) (*api.CatalogPricing, error) {
	return m.updateCatalogPricingResp, m.updateCatalogPricingErr
}

func (m *catalogPricingMockAPIClient) DeleteCatalogPricing(ctx context.Context, id string) error {
	return m.deleteCatalogPricingErr
}

// setupCatalogPricingMockFactories sets up mock factories for catalog pricing tests.
func setupCatalogPricingMockFactories(mockClient *catalogPricingMockAPIClient) (func(), *bytes.Buffer) {
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

// newCatalogPricingTestCmd creates a test command with common flags for catalog pricing tests.
func newCatalogPricingTestCmd() *cobra.Command {
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

func TestCatalogPricingListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := catalogPricingListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCatalogPricingGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := catalogPricingGetCmd.RunE(cmd, []string{"pricing-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCatalogPricingCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("catalog-id", "cat-123", "")
	cmd.Flags().String("product-id", "prod-123", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Float64("catalog-price", 99.99, "")
	cmd.Flags().Int("min-quantity", 0, "")
	cmd.Flags().Int("max-quantity", 0, "")

	err := catalogPricingCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCatalogPricingUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Float64("catalog-price", 0, "")
	cmd.Flags().Int("min-quantity", 0, "")
	cmd.Flags().Int("max-quantity", 0, "")

	err := catalogPricingUpdateCmd.RunE(cmd, []string{"pricing-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCatalogPricingDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	// yes flag already added by newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := catalogPricingDeleteCmd.RunE(cmd, []string{"pricing-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCatalogPricingListFlags(t *testing.T) {
	flags := catalogPricingListCmd.Flags()

	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
	if flags.Lookup("catalog-id") == nil {
		t.Error("Expected catalog-id flag")
	}
	if flags.Lookup("product-id") == nil {
		t.Error("Expected product-id flag")
	}
}

func TestCatalogPricingCommandStructure(t *testing.T) {
	if catalogPricingCmd.Use != "catalog-pricing" {
		t.Errorf("Expected Use 'catalog-pricing', got %s", catalogPricingCmd.Use)
	}

	subcommands := catalogPricingCmd.Commands()
	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %s not found", name)
		}
	}
}

// TestCatalogPricingListRunE tests the catalog pricing list command with mock API.
func TestCatalogPricingListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CatalogPricingListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CatalogPricingListResponse{
				Items: []api.CatalogPricing{
					{
						ID:            "pricing_123",
						CatalogID:     "catalog_abc",
						ProductID:     "product_xyz",
						OriginalPrice: 100.00,
						CatalogPrice:  85.00,
						DiscountPct:   15,
						MinQuantity:   10,
						MaxQuantity:   100,
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "pricing_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CatalogPricingListResponse{
				Items:      []api.CatalogPricing{},
				TotalCount: 0,
			},
		},
		{
			name: "list with quantity ranges - min only",
			mockResp: &api.CatalogPricingListResponse{
				Items: []api.CatalogPricing{
					{
						ID:            "pricing_456",
						CatalogID:     "catalog_def",
						ProductID:     "product_ghi",
						OriginalPrice: 50.00,
						CatalogPrice:  40.00,
						DiscountPct:   20,
						MinQuantity:   5,
						MaxQuantity:   0, // No max, should show "5+"
						CreatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "5+",
		},
		{
			name: "list with no quantity range",
			mockResp: &api.CatalogPricingListResponse{
				Items: []api.CatalogPricing{
					{
						ID:            "pricing_789",
						CatalogID:     "catalog_jkl",
						ProductID:     "product_mno",
						OriginalPrice: 200.00,
						CatalogPrice:  180.00,
						DiscountPct:   10,
						MinQuantity:   0,
						MaxQuantity:   0, // No range, should show "-"
						CreatedAt:     time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &catalogPricingMockAPIClient{
				listCatalogPricingResp: tt.mockResp,
				listCatalogPricingErr:  tt.mockErr,
			}
			cleanup, buf := setupCatalogPricingMockFactories(mockClient)
			defer cleanup()

			cmd := newCatalogPricingTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("catalog-id", "", "")
			cmd.Flags().String("product-id", "", "")

			err := catalogPricingListCmd.RunE(cmd, []string{})

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

// TestCatalogPricingListRunEJSON tests the catalog pricing list command with JSON output.
func TestCatalogPricingListRunEJSON(t *testing.T) {
	mockResp := &api.CatalogPricingListResponse{
		Items: []api.CatalogPricing{
			{
				ID:            "pricing_json",
				CatalogID:     "catalog_json",
				ProductID:     "product_json",
				OriginalPrice: 100.00,
				CatalogPrice:  75.00,
				DiscountPct:   25,
				MinQuantity:   1,
				MaxQuantity:   50,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		TotalCount: 1,
	}

	mockClient := &catalogPricingMockAPIClient{
		listCatalogPricingResp: mockResp,
	}
	cleanup, buf := setupCatalogPricingMockFactories(mockClient)
	defer cleanup()

	cmd := newCatalogPricingTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("catalog-id", "", "")
	cmd.Flags().String("product-id", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := catalogPricingListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "pricing_json") {
		t.Errorf("JSON output should contain pricing ID, got: %s", output)
	}
}

// TestCatalogPricingGetRunE tests the catalog pricing get command with mock API.
func TestCatalogPricingGetRunE(t *testing.T) {
	tests := []struct {
		name      string
		pricingID string
		mockResp  *api.CatalogPricing
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			pricingID: "pricing_123",
			mockResp: &api.CatalogPricing{
				ID:            "pricing_123",
				CatalogID:     "catalog_abc",
				ProductID:     "product_xyz",
				VariantID:     "variant_123",
				OriginalPrice: 100.00,
				CatalogPrice:  85.00,
				DiscountPct:   15,
				MinQuantity:   10,
				MaxQuantity:   100,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "pricing not found",
			pricingID: "pricing_999",
			mockErr:   errors.New("pricing not found"),
			wantErr:   true,
		},
		{
			name:      "get with no variant ID",
			pricingID: "pricing_456",
			mockResp: &api.CatalogPricing{
				ID:            "pricing_456",
				CatalogID:     "catalog_def",
				ProductID:     "product_ghi",
				VariantID:     "", // No variant
				OriginalPrice: 50.00,
				CatalogPrice:  40.00,
				DiscountPct:   20,
				MinQuantity:   0,
				MaxQuantity:   0,
				CreatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "get with min quantity only",
			pricingID: "pricing_789",
			mockResp: &api.CatalogPricing{
				ID:            "pricing_789",
				CatalogID:     "catalog_jkl",
				ProductID:     "product_mno",
				OriginalPrice: 200.00,
				CatalogPrice:  180.00,
				DiscountPct:   10,
				MinQuantity:   5,
				MaxQuantity:   0,
				CreatedAt:     time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "get with max quantity only",
			pricingID: "pricing_abc",
			mockResp: &api.CatalogPricing{
				ID:            "pricing_abc",
				CatalogID:     "catalog_pqr",
				ProductID:     "product_stu",
				OriginalPrice: 150.00,
				CatalogPrice:  120.00,
				DiscountPct:   20,
				MinQuantity:   0,
				MaxQuantity:   25,
				CreatedAt:     time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &catalogPricingMockAPIClient{
				getCatalogPricingResp: tt.mockResp,
				getCatalogPricingErr:  tt.mockErr,
			}
			cleanup, _ := setupCatalogPricingMockFactories(mockClient)
			defer cleanup()

			cmd := newCatalogPricingTestCmd()

			err := catalogPricingGetCmd.RunE(cmd, []string{tt.pricingID})

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

// TestCatalogPricingGetRunEJSON tests the catalog pricing get command with JSON output.
func TestCatalogPricingGetRunEJSON(t *testing.T) {
	mockResp := &api.CatalogPricing{
		ID:            "pricing_json",
		CatalogID:     "catalog_json",
		ProductID:     "product_json",
		VariantID:     "variant_json",
		OriginalPrice: 100.00,
		CatalogPrice:  75.00,
		DiscountPct:   25,
		MinQuantity:   1,
		MaxQuantity:   50,
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	mockClient := &catalogPricingMockAPIClient{
		getCatalogPricingResp: mockResp,
	}
	cleanup, buf := setupCatalogPricingMockFactories(mockClient)
	defer cleanup()

	cmd := newCatalogPricingTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := catalogPricingGetCmd.RunE(cmd, []string{"pricing_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "pricing_json") {
		t.Errorf("JSON output should contain pricing ID, got: %s", output)
	}
}

// TestCatalogPricingCreateRunE tests the catalog pricing create command with mock API.
func TestCatalogPricingCreateRunE(t *testing.T) {
	tests := []struct {
		name         string
		catalogID    string
		productID    string
		variantID    string
		catalogPrice float64
		minQuantity  int
		maxQuantity  int
		mockResp     *api.CatalogPricing
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful create",
			catalogID:    "catalog_123",
			productID:    "product_456",
			catalogPrice: 99.99,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_new",
				CatalogID:     "catalog_123",
				ProductID:     "product_456",
				OriginalPrice: 120.00,
				CatalogPrice:  99.99,
				DiscountPct:   16.67,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:         "create with variant",
			catalogID:    "catalog_123",
			productID:    "product_456",
			variantID:    "variant_789",
			catalogPrice: 49.99,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_variant",
				CatalogID:     "catalog_123",
				ProductID:     "product_456",
				VariantID:     "variant_789",
				OriginalPrice: 60.00,
				CatalogPrice:  49.99,
				DiscountPct:   16.68,
				CreatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:         "create with quantity range",
			catalogID:    "catalog_123",
			productID:    "product_456",
			catalogPrice: 79.99,
			minQuantity:  10,
			maxQuantity:  50,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_qty",
				CatalogID:     "catalog_123",
				ProductID:     "product_456",
				OriginalPrice: 100.00,
				CatalogPrice:  79.99,
				DiscountPct:   20.01,
				MinQuantity:   10,
				MaxQuantity:   50,
				CreatedAt:     time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name:         "API error",
			catalogID:    "catalog_123",
			productID:    "product_456",
			catalogPrice: 99.99,
			mockErr:      errors.New("failed to create catalog pricing"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &catalogPricingMockAPIClient{
				createCatalogPricingResp: tt.mockResp,
				createCatalogPricingErr:  tt.mockErr,
			}
			cleanup, _ := setupCatalogPricingMockFactories(mockClient)
			defer cleanup()

			cmd := newCatalogPricingTestCmd()
			cmd.Flags().String("catalog-id", tt.catalogID, "")
			cmd.Flags().String("product-id", tt.productID, "")
			cmd.Flags().String("variant-id", tt.variantID, "")
			cmd.Flags().Float64("catalog-price", tt.catalogPrice, "")
			cmd.Flags().Int("min-quantity", tt.minQuantity, "")
			cmd.Flags().Int("max-quantity", tt.maxQuantity, "")

			err := catalogPricingCreateCmd.RunE(cmd, []string{})

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

// TestCatalogPricingCreateRunEJSON tests the catalog pricing create command with JSON output.
func TestCatalogPricingCreateRunEJSON(t *testing.T) {
	mockResp := &api.CatalogPricing{
		ID:            "pricing_json_create",
		CatalogID:     "catalog_json",
		ProductID:     "product_json",
		OriginalPrice: 100.00,
		CatalogPrice:  75.00,
		DiscountPct:   25,
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	mockClient := &catalogPricingMockAPIClient{
		createCatalogPricingResp: mockResp,
	}
	cleanup, buf := setupCatalogPricingMockFactories(mockClient)
	defer cleanup()

	cmd := newCatalogPricingTestCmd()
	cmd.Flags().String("catalog-id", "catalog_json", "")
	cmd.Flags().String("product-id", "product_json", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Float64("catalog-price", 75.00, "")
	cmd.Flags().Int("min-quantity", 0, "")
	cmd.Flags().Int("max-quantity", 0, "")
	_ = cmd.Flags().Set("output", "json")

	err := catalogPricingCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "pricing_json_create") {
		t.Errorf("JSON output should contain pricing ID, got: %s", output)
	}
}

// TestCatalogPricingUpdateRunE tests the catalog pricing update command with mock API.
func TestCatalogPricingUpdateRunE(t *testing.T) {
	tests := []struct {
		name         string
		pricingID    string
		catalogPrice float64
		setPrice     bool
		minQuantity  int
		setMin       bool
		maxQuantity  int
		setMax       bool
		mockResp     *api.CatalogPricing
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful update price",
			pricingID:    "pricing_123",
			catalogPrice: 89.99,
			setPrice:     true,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_123",
				CatalogID:     "catalog_abc",
				ProductID:     "product_xyz",
				OriginalPrice: 100.00,
				CatalogPrice:  89.99,
				DiscountPct:   10.01,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful update min quantity",
			pricingID:   "pricing_123",
			minQuantity: 5,
			setMin:      true,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_123",
				CatalogID:     "catalog_abc",
				ProductID:     "product_xyz",
				OriginalPrice: 100.00,
				CatalogPrice:  85.00,
				DiscountPct:   15,
				MinQuantity:   5,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "successful update max quantity",
			pricingID:   "pricing_123",
			maxQuantity: 100,
			setMax:      true,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_123",
				CatalogID:     "catalog_abc",
				ProductID:     "product_xyz",
				OriginalPrice: 100.00,
				CatalogPrice:  85.00,
				DiscountPct:   15,
				MaxQuantity:   100,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:         "successful update all fields",
			pricingID:    "pricing_456",
			catalogPrice: 59.99,
			setPrice:     true,
			minQuantity:  10,
			setMin:       true,
			maxQuantity:  50,
			setMax:       true,
			mockResp: &api.CatalogPricing{
				ID:            "pricing_456",
				CatalogID:     "catalog_def",
				ProductID:     "product_ghi",
				OriginalPrice: 80.00,
				CatalogPrice:  59.99,
				DiscountPct:   25.01,
				MinQuantity:   10,
				MaxQuantity:   50,
				CreatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 2, 2, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "pricing not found",
			pricingID: "pricing_999",
			mockErr:   errors.New("pricing not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &catalogPricingMockAPIClient{
				updateCatalogPricingResp: tt.mockResp,
				updateCatalogPricingErr:  tt.mockErr,
			}
			cleanup, _ := setupCatalogPricingMockFactories(mockClient)
			defer cleanup()

			cmd := newCatalogPricingTestCmd()
			cmd.Flags().Float64("catalog-price", tt.catalogPrice, "")
			cmd.Flags().Int("min-quantity", tt.minQuantity, "")
			cmd.Flags().Int("max-quantity", tt.maxQuantity, "")

			// Simulate flag being changed
			if tt.setPrice {
				_ = cmd.Flags().Set("catalog-price", "89.99")
			}
			if tt.setMin {
				_ = cmd.Flags().Set("min-quantity", "5")
			}
			if tt.setMax {
				_ = cmd.Flags().Set("max-quantity", "100")
			}

			err := catalogPricingUpdateCmd.RunE(cmd, []string{tt.pricingID})

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

// TestCatalogPricingUpdateRunEJSON tests the catalog pricing update command with JSON output.
func TestCatalogPricingUpdateRunEJSON(t *testing.T) {
	mockResp := &api.CatalogPricing{
		ID:            "pricing_json_update",
		CatalogID:     "catalog_json",
		ProductID:     "product_json",
		OriginalPrice: 100.00,
		CatalogPrice:  65.00,
		DiscountPct:   35,
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
	}

	mockClient := &catalogPricingMockAPIClient{
		updateCatalogPricingResp: mockResp,
	}
	cleanup, buf := setupCatalogPricingMockFactories(mockClient)
	defer cleanup()

	cmd := newCatalogPricingTestCmd()
	cmd.Flags().Float64("catalog-price", 65.00, "")
	cmd.Flags().Int("min-quantity", 0, "")
	cmd.Flags().Int("max-quantity", 0, "")
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("catalog-price", "65.00")

	err := catalogPricingUpdateCmd.RunE(cmd, []string{"pricing_json_update"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "pricing_json_update") {
		t.Errorf("JSON output should contain pricing ID, got: %s", output)
	}
}

// TestCatalogPricingDeleteRunE tests the catalog pricing delete command with mock API.
func TestCatalogPricingDeleteRunE(t *testing.T) {
	tests := []struct {
		name      string
		pricingID string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful delete",
			pricingID: "pricing_123",
		},
		{
			name:      "pricing not found",
			pricingID: "pricing_999",
			mockErr:   errors.New("pricing not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &catalogPricingMockAPIClient{
				deleteCatalogPricingErr: tt.mockErr,
			}
			cleanup, _ := setupCatalogPricingMockFactories(mockClient)
			defer cleanup()

			cmd := newCatalogPricingTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := catalogPricingDeleteCmd.RunE(cmd, []string{tt.pricingID})

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

// TestCatalogPricingCreateFlags verifies create command flags exist
func TestCatalogPricingCreateFlags(t *testing.T) {
	flags := catalogPricingCreateCmd.Flags()

	expectedFlags := []string{
		"catalog-id",
		"product-id",
		"variant-id",
		"catalog-price",
		"min-quantity",
		"max-quantity",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("Expected flag %s not found", name)
		}
	}
}

// TestCatalogPricingUpdateFlags verifies update command flags exist
func TestCatalogPricingUpdateFlags(t *testing.T) {
	flags := catalogPricingUpdateCmd.Flags()

	expectedFlags := []string{
		"catalog-price",
		"min-quantity",
		"max-quantity",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("Expected flag %s not found", name)
		}
	}
}

// TestCatalogPricingDeleteFlags verifies delete command flags exist
func TestCatalogPricingDeleteFlags(t *testing.T) {
	flags := catalogPricingDeleteCmd.Flags()

	if flags.Lookup("yes") == nil {
		t.Error("Expected yes flag not found")
	}
}

// TestCatalogPricingSubcommandDescriptions verifies subcommand descriptions
func TestCatalogPricingSubcommandDescriptions(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List catalog pricing entries",
		"get":    "Get catalog pricing details",
		"create": "Create catalog pricing",
		"update": "Update catalog pricing",
		"delete": "Delete catalog pricing",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range catalogPricingCmd.Commands() {
				baseUse := getBaseUse(sub.Use)
				if baseUse == name {
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

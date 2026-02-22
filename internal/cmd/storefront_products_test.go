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

func TestStorefrontProductsCmdStructure(t *testing.T) {
	if storefrontProductsCmd.Use != "storefront-products" {
		t.Errorf("Expected Use 'storefront-products', got %q", storefrontProductsCmd.Use)
	}

	subcommands := storefrontProductsCmd.Commands()
	expectedSubs := []string{"list", "get"}

	for _, exp := range expectedSubs {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == exp || strings.HasPrefix(cmd.Use, exp+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing subcommand: %s", exp)
		}
	}
}

func TestStorefrontProductsListCmdFlags(t *testing.T) {
	flags := []string{"collection", "category", "vendor", "product-type", "tag", "q", "page", "page-size"}
	for _, flagName := range flags {
		flag := storefrontProductsListCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Missing --%s flag", flagName)
		}
	}
}

func TestStorefrontProductsGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"prod_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"prod_1", "prod_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontProductsGetCmd.Args(storefrontProductsGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorefrontProductsGetCmdFlags(t *testing.T) {
	byHandleFlag := storefrontProductsGetCmd.Flags().Lookup("by-handle")
	if byHandleFlag == nil {
		t.Error("Missing --by-handle flag")
	}
}

func TestStorefrontProductsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(storefrontProductsListCmd)

	err := storefrontProductsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStorefrontProductsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(storefrontProductsGetCmd)
	cmd.Flags().Bool("by-handle", false, "")

	err := storefrontProductsGetCmd.RunE(cmd, []string{"prod_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// Ensure unused imports don't cause errors
var _ = secrets.StoreCredentials{}

// storefrontProductsMockAPIClient is a mock implementation of api.APIClient for storefront products tests.
type storefrontProductsMockAPIClient struct {
	api.MockClient
	listStorefrontProductsResp       *api.StorefrontProductsListResponse
	listStorefrontProductsErr        error
	getStorefrontProductResp         *api.StorefrontProduct
	getStorefrontProductErr          error
	getStorefrontProductByHandleResp *api.StorefrontProduct
	getStorefrontProductByHandleErr  error
}

func (m *storefrontProductsMockAPIClient) ListStorefrontProducts(ctx context.Context, opts *api.StorefrontProductsListOptions) (*api.StorefrontProductsListResponse, error) {
	return m.listStorefrontProductsResp, m.listStorefrontProductsErr
}

func (m *storefrontProductsMockAPIClient) GetStorefrontProduct(ctx context.Context, id string) (*api.StorefrontProduct, error) {
	return m.getStorefrontProductResp, m.getStorefrontProductErr
}

func (m *storefrontProductsMockAPIClient) GetStorefrontProductByHandle(ctx context.Context, handle string) (*api.StorefrontProduct, error) {
	return m.getStorefrontProductByHandleResp, m.getStorefrontProductByHandleErr
}

// setupStorefrontProductsMockFactories sets up mock factories for storefront products tests.
func setupStorefrontProductsMockFactories(mockClient *storefrontProductsMockAPIClient) (func(), *bytes.Buffer) {
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

// newStorefrontProductsTestCmd creates a test command with common flags for storefront products tests.
func newStorefrontProductsTestCmd() *cobra.Command {
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

// TestStorefrontProductsListRunE tests the storefront products list command with mock API.
func TestStorefrontProductsListRunE(t *testing.T) {
	publishedAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		mockResp   *api.StorefrontProductsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.StorefrontProductsListResponse{
				Items: []api.StorefrontProduct{
					{
						ID:          "prod_123",
						Handle:      "test-product",
						Title:       "Test Product",
						Description: "A great product",
						Vendor:      "Test Vendor",
						ProductType: "Electronics",
						Tags:        []string{"sale", "featured"},
						Status:      "active",
						Available:   true,
						Price:       "99.99",
						Currency:    "USD",
						ViewCount:   1000,
						SalesCount:  50,
						ReviewCount: 25,
						PublishedAt: &publishedAt,
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "prod_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.StorefrontProductsListResponse{
				Items:      []api.StorefrontProduct{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontProductsMockAPIClient{
				listStorefrontProductsResp: tt.mockResp,
				listStorefrontProductsErr:  tt.mockErr,
			}
			cleanup, buf := setupStorefrontProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontProductsTestCmd()
			cmd.Flags().String("collection", "", "")
			cmd.Flags().String("category", "", "")
			cmd.Flags().String("vendor", "", "")
			cmd.Flags().String("product-type", "", "")
			cmd.Flags().String("tag", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := storefrontProductsListCmd.RunE(cmd, []string{})

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

// TestStorefrontProductsGetRunE tests the storefront products get command with mock API.
func TestStorefrontProductsGetRunE(t *testing.T) {
	publishedAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		id       string
		byHandle bool
		mockResp *api.StorefrontProduct
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get by ID",
			id:   "prod_123",
			mockResp: &api.StorefrontProduct{
				ID:             "prod_123",
				Handle:         "test-product",
				Title:          "Test Product",
				Description:    "A great product",
				Vendor:         "Test Vendor",
				ProductType:    "Electronics",
				Tags:           []string{"sale", "featured"},
				Status:         "active",
				Available:      true,
				Price:          "99.99",
				CompareAtPrice: "129.99",
				Currency:       "USD",
				ViewCount:      1000,
				SalesCount:     50,
				ReviewCount:    25,
				AverageRating:  4.5,
				PublishedAt:    &publishedAt,
				Variants: []api.StorefrontProductVariant{
					{
						ID:        "var_123",
						Title:     "Default",
						SKU:       "TEST-001",
						Price:     "99.99",
						Available: true,
						Inventory: 100,
					},
				},
				Images: []api.StorefrontProductImage{
					{
						ID:     "img_123",
						URL:    "https://example.com/image.jpg",
						Width:  800,
						Height: 600,
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:     "successful get by handle",
			id:       "test-product",
			byHandle: true,
			mockResp: &api.StorefrontProduct{
				ID:     "prod_123",
				Handle: "test-product",
				Title:  "Test Product",
			},
		},
		{
			name:    "not found",
			id:      "prod_999",
			mockErr: errors.New("storefront product not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontProductsMockAPIClient{
				getStorefrontProductResp:         tt.mockResp,
				getStorefrontProductErr:          tt.mockErr,
				getStorefrontProductByHandleResp: tt.mockResp,
				getStorefrontProductByHandleErr:  tt.mockErr,
			}
			cleanup, _ := setupStorefrontProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontProductsTestCmd()
			cmd.Flags().Bool("by-handle", tt.byHandle, "")

			err := storefrontProductsGetCmd.RunE(cmd, []string{tt.id})

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

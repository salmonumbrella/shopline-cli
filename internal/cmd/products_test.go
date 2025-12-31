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

// TestProductsCommandSetup verifies products command initialization
func TestProductsCommandSetup(t *testing.T) {
	if productsCmd.Use != "products" {
		t.Errorf("expected Use 'products', got %q", productsCmd.Use)
	}
	if productsCmd.Short != "Manage products" {
		t.Errorf("expected Short 'Manage products', got %q", productsCmd.Short)
	}
}

// TestProductsSubcommands verifies all subcommands are registered
func TestProductsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list": "List products",
		"get":  "Get product details",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range productsCmd.Commands() {
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

// TestProductsListFlags verifies list command flags exist with correct defaults
func TestProductsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status", ""},
		{"vendor", ""},
		{"product-type", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := productsListCmd.Flags().Lookup(f.name)
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

// TestProductsGetClientError verifies error handling when getClient fails
func TestProductsGetClientError(t *testing.T) {
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

// TestProductsWithMockStore tests products commands with a mock credential store
func TestProductsWithMockStore(t *testing.T) {
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

// productsMockAPIClient is a mock implementation of api.APIClient for products tests.
type productsMockAPIClient struct {
	api.MockClient
	listProductsResp *api.ProductsListResponse
	listProductsErr  error
	getProductResp   *api.Product
	getProductErr    error
}

func (m *productsMockAPIClient) ListProducts(ctx context.Context, opts *api.ProductsListOptions) (*api.ProductsListResponse, error) {
	return m.listProductsResp, m.listProductsErr
}

func (m *productsMockAPIClient) GetProduct(ctx context.Context, id string) (*api.Product, error) {
	return m.getProductResp, m.getProductErr
}

// setupProductsMockFactories sets up mock factories for products tests.
func setupProductsMockFactories(mockClient *productsMockAPIClient) (func(), *bytes.Buffer) {
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

// newProductsTestCmd creates a test command with common flags for products tests.
func newProductsTestCmd() *cobra.Command {
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

// TestProductsListRunE tests the products list command with mock API.
func TestProductsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ProductsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ProductsListResponse{
				Items: []api.Product{
					{
						ID:          "prod_123",
						Title:       "Test Product",
						Handle:      "test-product",
						Status:      "active",
						Vendor:      "Test Vendor",
						ProductType: "Electronics",
						Price:       &api.Price{Cents: 9999, Label: "99.99"},
						Currency:    "USD",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
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
			mockResp: &api.ProductsListResponse{
				Items:      []api.Product{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productsMockAPIClient{
				listProductsResp: tt.mockResp,
				listProductsErr:  tt.mockErr,
			}
			cleanup, buf := setupProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductsTestCmd()
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("vendor", "", "")
			cmd.Flags().String("product-type", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := productsListCmd.RunE(cmd, []string{})

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

// TestProductsGetRunE tests the products get command with mock API.
func TestProductsGetRunE(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		mockResp  *api.Product
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			productID: "prod_123",
			mockResp: &api.Product{
				ID:          "prod_123",
				Title:       "Test Product",
				Handle:      "test-product",
				Description: "A test product",
				Status:      "active",
				Vendor:      "Test Vendor",
				ProductType: "Electronics",
				Price:       &api.Price{Cents: 9999, Label: "99.99"},
				Currency:    "USD",
			},
		},
		{
			name:      "product not found",
			productID: "prod_999",
			mockErr:   errors.New("product not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productsMockAPIClient{
				getProductResp: tt.mockResp,
				getProductErr:  tt.mockErr,
			}
			cleanup, _ := setupProductsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductsTestCmd()

			err := productsGetCmd.RunE(cmd, []string{tt.productID})

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

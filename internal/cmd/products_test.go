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
	listProductsResp   *api.ProductsListResponse
	listProductsErr    error
	listProductsByPage map[int]*api.ProductsListResponse
	listProductsCalls  []*api.ProductsListOptions
	getProductResp     *api.Product
	getProductErr      error

	searchProductsResp *api.ProductsListResponse
	searchProductsErr  error
}

func (m *productsMockAPIClient) ListProducts(ctx context.Context, opts *api.ProductsListOptions) (*api.ProductsListResponse, error) {
	if opts != nil {
		cp := *opts
		m.listProductsCalls = append(m.listProductsCalls, &cp)
		if m.listProductsByPage != nil {
			if resp, ok := m.listProductsByPage[opts.Page]; ok {
				return resp, m.listProductsErr
			}
		}
	}
	return m.listProductsResp, m.listProductsErr
}

func (m *productsMockAPIClient) GetProduct(ctx context.Context, id string) (*api.Product, error) {
	return m.getProductResp, m.getProductErr
}

func (m *productsMockAPIClient) SearchProducts(ctx context.Context, opts *api.ProductSearchOptions) (*api.ProductsListResponse, error) {
	return m.searchProductsResp, m.searchProductsErr
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

func TestProductsListRunELimitPaginates(t *testing.T) {
	page1 := &api.ProductsListResponse{
		Items:      make([]api.Product, 24),
		TotalCount: 100,
		HasMore:    true,
	}
	for i := range page1.Items {
		page1.Items[i] = api.Product{ID: "prod_p1_" + strconv.Itoa(i)}
	}

	page2 := &api.ProductsListResponse{
		Items:      make([]api.Product, 24),
		TotalCount: 100,
		HasMore:    false,
	}
	for i := range page2.Items {
		page2.Items[i] = api.Product{ID: "prod_p2_" + strconv.Itoa(i)}
	}

	mockClient := &productsMockAPIClient{
		listProductsByPage: map[int]*api.ProductsListResponse{
			1: page1,
			2: page2,
		},
	}
	cleanup, buf := setupProductsMockFactories(mockClient)
	defer cleanup()

	cmd := newProductsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Int("limit", 0, "")
	_ = cmd.Flags().Set("limit", "30")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("vendor", "", "")
	cmd.Flags().String("product-type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := productsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockClient.listProductsCalls) < 2 {
		t.Fatalf("expected at least 2 ListProducts calls, got %d", len(mockClient.listProductsCalls))
	}
	if mockClient.listProductsCalls[0].Page != 1 || mockClient.listProductsCalls[1].Page != 2 {
		t.Fatalf("expected calls for pages 1 and 2, got %+v", mockClient.listProductsCalls)
	}

	var resp api.ProductsListResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}
	if len(resp.Items) != 30 {
		t.Fatalf("expected 30 items, got %d", len(resp.Items))
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

// TestProductsGetByFlag tests the --by flag on the products get command.
func TestProductsGetByFlag(t *testing.T) {
	t.Run("resolves product by title", func(t *testing.T) {
		mockClient := &productsMockAPIClient{
			searchProductsResp: &api.ProductsListResponse{
				Items:      []api.Product{{ID: "prod_found", Title: "Cool Widget"}},
				TotalCount: 1,
			},
			getProductResp: &api.Product{
				ID:    "prod_found",
				Title: "Cool Widget",
			},
		}
		cleanup, buf := setupProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newProductsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Cool Widget")

		if err := productsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "prod_found") {
			t.Errorf("expected output to contain 'prod_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &productsMockAPIClient{
			searchProductsResp: &api.ProductsListResponse{
				Items:      []api.Product{},
				TotalCount: 0,
			},
		}
		cleanup, _ := setupProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newProductsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Nonexistent Product")

		err := productsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no product found")
		}
		if !strings.Contains(err.Error(), "no product found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &productsMockAPIClient{
			searchProductsErr: errors.New("API error"),
		}
		cleanup, _ := setupProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newProductsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Widget")

		err := productsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("warns on multiple matches", func(t *testing.T) {
		mockClient := &productsMockAPIClient{
			searchProductsResp: &api.ProductsListResponse{
				Items: []api.Product{
					{ID: "prod_1", Title: "Widget A"},
					{ID: "prod_2", Title: "Widget B"},
				},
				TotalCount: 2,
			},
			getProductResp: &api.Product{
				ID:    "prod_1",
				Title: "Widget A",
			},
		}
		cleanup, buf := setupProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newProductsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Widget")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		if err := productsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "prod_1") {
			t.Errorf("expected output to contain 'prod_1', got: %s", buf.String())
		}
		if !strings.Contains(stderr.String(), "2 products matched") {
			t.Errorf("expected stderr warning about multiple matches, got: %s", stderr.String())
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &productsMockAPIClient{
			getProductResp: &api.Product{
				ID:    "prod_direct",
				Title: "Direct Product",
			},
		}
		cleanup, buf := setupProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newProductsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := productsGetCmd.RunE(cmd, []string{"prod_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "prod_direct") {
			t.Errorf("expected output to contain 'prod_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &productsMockAPIClient{}
		cleanup, _ := setupProductsMockFactories(mockClient)
		defer cleanup()

		cmd := newProductsTestCmd()
		cmd.Flags().String("by", "", "")

		err := productsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

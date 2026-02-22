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

// TestProductListingsCommandSetup verifies product-listings command initialization
func TestProductListingsCommandSetup(t *testing.T) {
	if productListingsCmd.Use != "product-listings" {
		t.Errorf("expected Use 'product-listings', got %q", productListingsCmd.Use)
	}
	if productListingsCmd.Short != "Manage product listings (products published to sales channels)" {
		t.Errorf("expected Short 'Manage product listings (products published to sales channels)', got %q", productListingsCmd.Short)
	}
}

// TestProductListingsSubcommands verifies all subcommands are registered
func TestProductListingsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List product listings",
		"get":    "Get product listing details",
		"create": "Publish a product to a sales channel",
		"delete": "Remove a product listing from a sales channel",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range productListingsCmd.Commands() {
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

// TestProductListingsListFlags verifies list command flags exist with correct defaults
func TestProductListingsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := productListingsListCmd.Flags().Lookup(f.name)
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

// TestProductListingsCreateFlags verifies create command flags exist
func TestProductListingsCreateFlags(t *testing.T) {
	flag := productListingsCreateCmd.Flags().Lookup("product-id")
	if flag == nil {
		t.Error("flag 'product-id' not found")
	}
}

// TestProductListingsDeleteFlags verifies delete command flags exist
func TestProductListingsDeleteFlags(t *testing.T) {
	flag := productListingsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
	}
}

// TestProductListingsGetCmdUse verifies the get command has correct use string
func TestProductListingsGetCmdUse(t *testing.T) {
	if productListingsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", productListingsGetCmd.Use)
	}
}

// TestProductListingsDeleteCmdUse verifies the delete command has correct use string
func TestProductListingsDeleteCmdUse(t *testing.T) {
	if productListingsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", productListingsDeleteCmd.Use)
	}
}

// TestProductListingsListRunE_GetClientFails verifies error handling when getClient fails
func TestProductListingsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := productListingsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestProductListingsGetRunE_GetClientFails verifies error handling when getClient fails
func TestProductListingsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := productListingsGetCmd.RunE(cmd, []string{"listing_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestProductListingsCreateRunE_DryRun verifies dry-run mode works
func TestProductListingsCreateRunE_DryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("product-id", "", "")
	_ = cmd.Flags().Set("product-id", "prod_123")
	_ = cmd.Flags().Set("dry-run", "true")

	err := productListingsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %v", err)
	}
}

// TestProductListingsCreateRunE_GetClientFails verifies error handling when getClient fails
func TestProductListingsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("product-id", "", "")
	_ = cmd.Flags().Set("product-id", "prod_123")

	err := productListingsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestProductListingsDeleteRunE_DryRun verifies dry-run mode works
func TestProductListingsDeleteRunE_DryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	_ = cmd.Flags().Set("dry-run", "true")

	err := productListingsDeleteCmd.RunE(cmd, []string{"listing_123"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %v", err)
	}
}

// TestProductListingsDeleteRunE_NoConfirmation verifies delete requires confirmation
func TestProductListingsDeleteRunE_NoConfirmation(t *testing.T) {
	cmd := newTestCmdWithFlags()

	err := productListingsDeleteCmd.RunE(cmd, []string{"listing_123"})
	// Should return nil but print a message (no confirmation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestProductListingsListRunE_NoProfiles verifies error when no profiles are configured
func TestProductListingsListRunE_NoProfiles(t *testing.T) {
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
	err := productListingsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestProductListingsGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestProductListingsGetRunE_MultipleProfiles(t *testing.T) {
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
	err := productListingsGetCmd.RunE(cmd, []string{"listing_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// productListingsMockAPIClient is a mock implementation of api.APIClient for product listings tests.
type productListingsMockAPIClient struct {
	api.MockClient
	listProductListingsResp  *api.ProductListingsListResponse
	listProductListingsErr   error
	getProductListingResp    *api.ProductListing
	getProductListingErr     error
	createProductListingResp *api.ProductListing
	createProductListingErr  error
	deleteProductListingErr  error
}

func (m *productListingsMockAPIClient) ListProductListings(ctx context.Context, opts *api.ProductListingsListOptions) (*api.ProductListingsListResponse, error) {
	return m.listProductListingsResp, m.listProductListingsErr
}

func (m *productListingsMockAPIClient) GetProductListing(ctx context.Context, id string) (*api.ProductListing, error) {
	return m.getProductListingResp, m.getProductListingErr
}

func (m *productListingsMockAPIClient) CreateProductListing(ctx context.Context, productID string) (*api.ProductListing, error) {
	return m.createProductListingResp, m.createProductListingErr
}

func (m *productListingsMockAPIClient) DeleteProductListing(ctx context.Context, id string) error {
	return m.deleteProductListingErr
}

// setupProductListingsMockFactories sets up mock factories for product listings tests.
func setupProductListingsMockFactories(mockClient *productListingsMockAPIClient) (func(), *bytes.Buffer) {
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

// newProductListingsTestCmd creates a test command with common flags for product listings tests.
func newProductListingsTestCmd() *cobra.Command {
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

// TestProductListingsListRunE tests the product listings list command with mock API.
func TestProductListingsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ProductListingsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ProductListingsListResponse{
				Items: []api.ProductListing{
					{
						ID:          "pl_123",
						ProductID:   "prod_123",
						Title:       "Test Product",
						Handle:      "test-product",
						Vendor:      "Test Vendor",
						ProductType: "Electronics",
						Available:   true,
						PublishedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "pl_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ProductListingsListResponse{
				Items:      []api.ProductListing{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productListingsMockAPIClient{
				listProductListingsResp: tt.mockResp,
				listProductListingsErr:  tt.mockErr,
			}
			cleanup, buf := setupProductListingsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductListingsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := productListingsListCmd.RunE(cmd, []string{})

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

// TestProductListingsGetRunE tests the product listings get command with mock API.
func TestProductListingsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.ProductListing
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "pl_123",
			mockResp: &api.ProductListing{
				ID:          "pl_123",
				ProductID:   "prod_123",
				Title:       "Test Product",
				Handle:      "test-product",
				BodyHTML:    "<p>Description</p>",
				Vendor:      "Test Vendor",
				ProductType: "Electronics",
				Available:   true,
				PublishedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "pl_999",
			mockErr: errors.New("product listing not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productListingsMockAPIClient{
				getProductListingResp: tt.mockResp,
				getProductListingErr:  tt.mockErr,
			}
			cleanup, _ := setupProductListingsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductListingsTestCmd()

			err := productListingsGetCmd.RunE(cmd, []string{tt.id})

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

// TestProductListingsCreateRunE tests the product listings create command with mock API.
func TestProductListingsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		dryRun   bool
		mockResp *api.ProductListing
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "dry run",
			dryRun: true,
		},
		{
			name: "successful create",
			mockResp: &api.ProductListing{
				ID:        "pl_new",
				ProductID: "prod_123",
				Title:     "Test Product",
				Available: true,
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productListingsMockAPIClient{
				createProductListingResp: tt.mockResp,
				createProductListingErr:  tt.mockErr,
			}
			cleanup, _ := setupProductListingsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductListingsTestCmd()
			cmd.Flags().String("product-id", "prod_123", "")
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := productListingsCreateCmd.RunE(cmd, []string{})

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

// TestProductListingsDeleteRunE tests the product listings delete command with mock API.
func TestProductListingsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dryRun  bool
		mockErr error
		wantErr bool
	}{
		{
			name:   "dry run",
			id:     "pl_123",
			dryRun: true,
		},
		{
			name: "successful delete",
			id:   "pl_123",
		},
		{
			name:    "delete fails",
			id:      "pl_123",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productListingsMockAPIClient{
				deleteProductListingErr: tt.mockErr,
			}
			cleanup, _ := setupProductListingsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductListingsTestCmd()
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := productListingsDeleteCmd.RunE(cmd, []string{tt.id})

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

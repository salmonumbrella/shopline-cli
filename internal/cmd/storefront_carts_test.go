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

func TestStorefrontCartsCmdStructure(t *testing.T) {
	if storefrontCartsCmd.Use != "storefront-carts" {
		t.Errorf("Expected Use 'storefront-carts', got %q", storefrontCartsCmd.Use)
	}

	subcommands := storefrontCartsCmd.Commands()
	expectedSubs := []string{"list", "get", "create", "delete"}

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

func TestStorefrontCartsListCmdFlags(t *testing.T) {
	flags := []string{"customer-id", "status", "page", "page-size"}
	for _, flagName := range flags {
		flag := storefrontCartsListCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Missing --%s flag", flagName)
		}
	}
}

func TestStorefrontCartsGetCmdArgs(t *testing.T) {
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
			args:    []string{"cart_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"cart_1", "cart_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontCartsGetCmd.Args(storefrontCartsGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorefrontCartsCreateCmdFlags(t *testing.T) {
	flags := []string{"customer-id", "email", "currency"}
	for _, flagName := range flags {
		flag := storefrontCartsCreateCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Missing --%s flag", flagName)
		}
	}
}

func TestStorefrontCartsDeleteCmdArgs(t *testing.T) {
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
			args:    []string{"cart_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"cart_1", "cart_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontCartsDeleteCmd.Args(storefrontCartsDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorefrontCartsListGetClientError(t *testing.T) {
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
	cmd.AddCommand(storefrontCartsListCmd)

	err := storefrontCartsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStorefrontCartsGetGetClientError(t *testing.T) {
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
	cmd.AddCommand(storefrontCartsGetCmd)

	err := storefrontCartsGetCmd.RunE(cmd, []string{"cart_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStorefrontCartsCreateGetClientError(t *testing.T) {
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
	cmd.AddCommand(storefrontCartsCreateCmd)
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("currency", "", "")

	err := storefrontCartsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStorefrontCartsDeleteGetClientError(t *testing.T) {
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
	cmd.AddCommand(storefrontCartsDeleteCmd)

	err := storefrontCartsDeleteCmd.RunE(cmd, []string{"cart_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// Ensure unused imports don't cause errors
var _ = secrets.StoreCredentials{}

// storefrontCartsMockAPIClient is a mock implementation of api.APIClient for storefront carts tests.
type storefrontCartsMockAPIClient struct {
	api.MockClient
	listStorefrontCartsResp  *api.StorefrontCartsListResponse
	listStorefrontCartsErr   error
	getStorefrontCartResp    *api.StorefrontCart
	getStorefrontCartErr     error
	createStorefrontCartResp *api.StorefrontCart
	createStorefrontCartErr  error
	deleteStorefrontCartErr  error
}

func (m *storefrontCartsMockAPIClient) ListStorefrontCarts(ctx context.Context, opts *api.StorefrontCartsListOptions) (*api.StorefrontCartsListResponse, error) {
	return m.listStorefrontCartsResp, m.listStorefrontCartsErr
}

func (m *storefrontCartsMockAPIClient) GetStorefrontCart(ctx context.Context, id string) (*api.StorefrontCart, error) {
	return m.getStorefrontCartResp, m.getStorefrontCartErr
}

func (m *storefrontCartsMockAPIClient) CreateStorefrontCart(ctx context.Context, req *api.StorefrontCartCreateRequest) (*api.StorefrontCart, error) {
	return m.createStorefrontCartResp, m.createStorefrontCartErr
}

func (m *storefrontCartsMockAPIClient) DeleteStorefrontCart(ctx context.Context, id string) error {
	return m.deleteStorefrontCartErr
}

// setupStorefrontCartsMockFactories sets up mock factories for storefront carts tests.
func setupStorefrontCartsMockFactories(mockClient *storefrontCartsMockAPIClient) (func(), *bytes.Buffer) {
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

// newStorefrontCartsTestCmd creates a test command with common flags for storefront carts tests.
func newStorefrontCartsTestCmd() *cobra.Command {
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

// TestStorefrontCartsListRunE tests the storefront carts list command with mock API.
func TestStorefrontCartsListRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		mockResp   *api.StorefrontCartsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.StorefrontCartsListResponse{
				Items: []api.StorefrontCart{
					{
						ID:         "cart_123",
						CustomerID: "cust_456",
						Email:      "customer@example.com",
						Currency:   "USD",
						Subtotal:   "99.99",
						TotalPrice: "109.99",
						TotalTax:   "10.00",
						ItemCount:  3,
						CreatedAt:  createdAt,
						UpdatedAt:  updatedAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "cart_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.StorefrontCartsListResponse{
				Items:      []api.StorefrontCart{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontCartsMockAPIClient{
				listStorefrontCartsResp: tt.mockResp,
				listStorefrontCartsErr:  tt.mockErr,
			}
			cleanup, buf := setupStorefrontCartsMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontCartsTestCmd()
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := storefrontCartsListCmd.RunE(cmd, []string{})

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

// TestStorefrontCartsListRunEJSON tests the storefront carts list command with JSON output.
func TestStorefrontCartsListRunEJSON(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	mockClient := &storefrontCartsMockAPIClient{
		listStorefrontCartsResp: &api.StorefrontCartsListResponse{
			Items: []api.StorefrontCart{
				{
					ID:         "cart_123",
					CustomerID: "cust_456",
					Email:      "customer@example.com",
					Currency:   "USD",
					Subtotal:   "99.99",
					TotalPrice: "109.99",
					TotalTax:   "10.00",
					ItemCount:  3,
					CreatedAt:  createdAt,
					UpdatedAt:  updatedAt,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := storefrontCartsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cart_123") {
		t.Errorf("JSON output should contain cart ID, got: %s", output)
	}
}

// TestStorefrontCartsGetRunE tests the storefront carts get command with mock API.
func TestStorefrontCartsGetRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		id       string
		mockResp *api.StorefrontCart
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "cart_123",
			mockResp: &api.StorefrontCart{
				ID:            "cart_123",
				CustomerID:    "cust_456",
				Email:         "customer@example.com",
				Currency:      "USD",
				Subtotal:      "99.99",
				TotalPrice:    "109.99",
				TotalTax:      "10.00",
				TotalDiscount: "5.00",
				ItemCount:     3,
				Items: []api.StorefrontCartItem{
					{
						ID:           "item_1",
						ProductID:    "prod_123",
						VariantID:    "var_456",
						Title:        "Test Product",
						VariantTitle: "Large",
						Quantity:     2,
						Price:        "49.99",
						LineTotal:    "99.98",
					},
				},
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		},
		{
			name:    "cart not found",
			id:      "cart_999",
			mockErr: errors.New("storefront cart not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontCartsMockAPIClient{
				getStorefrontCartResp: tt.mockResp,
				getStorefrontCartErr:  tt.mockErr,
			}
			cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontCartsTestCmd()

			err := storefrontCartsGetCmd.RunE(cmd, []string{tt.id})

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

// TestStorefrontCartsGetRunEJSON tests the storefront carts get command with JSON output.
func TestStorefrontCartsGetRunEJSON(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	mockClient := &storefrontCartsMockAPIClient{
		getStorefrontCartResp: &api.StorefrontCart{
			ID:            "cart_123",
			CustomerID:    "cust_456",
			Email:         "customer@example.com",
			Currency:      "USD",
			Subtotal:      "99.99",
			TotalPrice:    "109.99",
			TotalTax:      "10.00",
			TotalDiscount: "5.00",
			ItemCount:     3,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		},
	}
	cleanup, buf := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := storefrontCartsGetCmd.RunE(cmd, []string{"cart_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cart_123") {
		t.Errorf("JSON output should contain cart ID, got: %s", output)
	}
}

// TestStorefrontCartsGetRunEWithItems tests the get command displays items.
func TestStorefrontCartsGetRunEWithItems(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	mockClient := &storefrontCartsMockAPIClient{
		getStorefrontCartResp: &api.StorefrontCart{
			ID:            "cart_123",
			CustomerID:    "cust_456",
			Email:         "customer@example.com",
			Currency:      "USD",
			Subtotal:      "99.99",
			TotalPrice:    "109.99",
			TotalTax:      "10.00",
			TotalDiscount: "5.00",
			ItemCount:     3,
			Items: []api.StorefrontCartItem{
				{
					ID:           "item_1",
					ProductID:    "prod_123",
					VariantID:    "var_456",
					Title:        "Test Product",
					VariantTitle: "Large",
					Quantity:     2,
					Price:        "49.99",
					LineTotal:    "99.98",
				},
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
	}
	cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()

	err := storefrontCartsGetCmd.RunE(cmd, []string{"cart_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontCartsGetRunENoItems tests the get command handles carts without items.
func TestStorefrontCartsGetRunENoItems(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	mockClient := &storefrontCartsMockAPIClient{
		getStorefrontCartResp: &api.StorefrontCart{
			ID:            "cart_123",
			CustomerID:    "cust_456",
			Email:         "customer@example.com",
			Currency:      "USD",
			Subtotal:      "0.00",
			TotalPrice:    "0.00",
			TotalTax:      "0.00",
			TotalDiscount: "0.00",
			ItemCount:     0,
			Items:         []api.StorefrontCartItem{},
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		},
	}
	cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()

	err := storefrontCartsGetCmd.RunE(cmd, []string{"cart_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStorefrontCartsCreateRunE tests the storefront carts create command with mock API.
func TestStorefrontCartsCreateRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		customerID string
		email      string
		currency   string
		mockResp   *api.StorefrontCart
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful create with customer ID",
			customerID: "cust_456",
			email:      "customer@example.com",
			currency:   "USD",
			mockResp: &api.StorefrontCart{
				ID:         "cart_123",
				CustomerID: "cust_456",
				Email:      "customer@example.com",
				Currency:   "USD",
				Subtotal:   "0.00",
				TotalPrice: "0.00",
				CreatedAt:  createdAt,
				UpdatedAt:  updatedAt,
			},
		},
		{
			name:     "successful create without customer ID",
			email:    "guest@example.com",
			currency: "EUR",
			mockResp: &api.StorefrontCart{
				ID:         "cart_124",
				CustomerID: "",
				Email:      "guest@example.com",
				Currency:   "EUR",
				Subtotal:   "0.00",
				TotalPrice: "0.00",
				CreatedAt:  createdAt,
				UpdatedAt:  updatedAt,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create cart"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontCartsMockAPIClient{
				createStorefrontCartResp: tt.mockResp,
				createStorefrontCartErr:  tt.mockErr,
			}
			cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontCartsTestCmd()
			cmd.Flags().String("customer-id", tt.customerID, "")
			cmd.Flags().String("email", tt.email, "")
			cmd.Flags().String("currency", tt.currency, "")

			err := storefrontCartsCreateCmd.RunE(cmd, []string{})

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

// TestStorefrontCartsCreateRunEJSON tests the storefront carts create command with JSON output.
func TestStorefrontCartsCreateRunEJSON(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	mockClient := &storefrontCartsMockAPIClient{
		createStorefrontCartResp: &api.StorefrontCart{
			ID:         "cart_123",
			CustomerID: "cust_456",
			Email:      "customer@example.com",
			Currency:   "USD",
			Subtotal:   "0.00",
			TotalPrice: "0.00",
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		},
	}
	cleanup, buf := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	cmd.Flags().String("customer-id", "cust_456", "")
	cmd.Flags().String("email", "customer@example.com", "")
	cmd.Flags().String("currency", "USD", "")
	_ = cmd.Flags().Set("output", "json")

	err := storefrontCartsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cart_123") {
		t.Errorf("JSON output should contain cart ID, got: %s", output)
	}
}

// TestStorefrontCartsCreateRunEDryRun tests the storefront carts create command with dry-run.
func TestStorefrontCartsCreateRunEDryRun(t *testing.T) {
	mockClient := &storefrontCartsMockAPIClient{}
	cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("currency", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontCartsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestStorefrontCartsCreateRunEDryRunWithCustomer tests the storefront carts create command with dry-run and customer ID.
func TestStorefrontCartsCreateRunEDryRunWithCustomer(t *testing.T) {
	mockClient := &storefrontCartsMockAPIClient{}
	cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	cmd.Flags().String("customer-id", "cust_456", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("currency", "", "")
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("customer-id", "cust_456")

	err := storefrontCartsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestStorefrontCartsDeleteRunE tests the storefront carts delete command with mock API.
func TestStorefrontCartsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   "cart_123",
		},
		{
			name:    "cart not found",
			id:      "cart_999",
			mockErr: errors.New("storefront cart not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &storefrontCartsMockAPIClient{
				deleteStorefrontCartErr: tt.mockErr,
			}
			cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
			defer cleanup()

			cmd := newStorefrontCartsTestCmd()

			err := storefrontCartsDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestStorefrontCartsDeleteRunEDryRun tests the storefront carts delete command with dry-run.
func TestStorefrontCartsDeleteRunEDryRun(t *testing.T) {
	mockClient := &storefrontCartsMockAPIClient{}
	cleanup, _ := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := storefrontCartsDeleteCmd.RunE(cmd, []string{"cart_123"})
	if err != nil {
		t.Errorf("unexpected error in dry-run mode: %v", err)
	}
}

// TestStorefrontCartsListWithFilters tests list command with various filter combinations.
func TestStorefrontCartsListWithFilters(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	mockClient := &storefrontCartsMockAPIClient{
		listStorefrontCartsResp: &api.StorefrontCartsListResponse{
			Items: []api.StorefrontCart{
				{
					ID:         "cart_123",
					CustomerID: "cust_456",
					Email:      "customer@example.com",
					Currency:   "USD",
					Subtotal:   "99.99",
					TotalPrice: "109.99",
					TotalTax:   "10.00",
					ItemCount:  3,
					CreatedAt:  createdAt,
					UpdatedAt:  updatedAt,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupStorefrontCartsMockFactories(mockClient)
	defer cleanup()

	cmd := newStorefrontCartsTestCmd()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("customer-id", "cust_456")
	_ = cmd.Flags().Set("status", "active")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := storefrontCartsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cart_123") {
		t.Errorf("output should contain cart ID, got: %s", output)
	}
}

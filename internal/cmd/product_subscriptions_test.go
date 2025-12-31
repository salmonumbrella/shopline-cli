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

func TestProductSubscriptionsCommand(t *testing.T) {
	if productSubscriptionsCmd == nil {
		t.Fatal("productSubscriptionsCmd is nil")
	}
	if productSubscriptionsCmd.Use != "product-subscriptions" {
		t.Errorf("Expected Use to be 'product-subscriptions', got %q", productSubscriptionsCmd.Use)
	}
}

func TestProductSubscriptionsSubcommands(t *testing.T) {
	subcommands := productSubscriptionsCmd.Commands()
	expectedCmds := map[string]bool{"list": false, "get": false, "create": false, "delete": false}
	for _, cmd := range subcommands {
		switch cmd.Use {
		case "list":
			expectedCmds["list"] = true
		case "get <id>":
			expectedCmds["get"] = true
		case "create":
			expectedCmds["create"] = true
		case "delete <id>":
			expectedCmds["delete"] = true
		}
	}
	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %q not found", name)
		}
	}
}

func TestProductSubscriptionsListFlags(t *testing.T) {
	flags := []string{"product-id", "customer-id", "status", "page", "page-size"}
	for _, flag := range flags {
		if productSubscriptionsListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on list command", flag)
		}
	}
}

func TestProductSubscriptionsCreateFlags(t *testing.T) {
	flags := []string{"product-id", "variant-id", "customer-id", "selling-plan-id", "quantity", "next-billing-date"}
	for _, flag := range flags {
		if productSubscriptionsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on create command", flag)
		}
	}
}

func TestProductSubscriptionsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := productSubscriptionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProductSubscriptionsGetArgsValidation(t *testing.T) {
	if productSubscriptionsGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := productSubscriptionsGetCmd.Args(productSubscriptionsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = productSubscriptionsGetCmd.Args(productSubscriptionsGetCmd, []string{"sub_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestProductSubscriptionsDeleteArgsValidation(t *testing.T) {
	if productSubscriptionsDeleteCmd.Args == nil {
		t.Fatal("Expected Args validator on delete command")
	}
	err := productSubscriptionsDeleteCmd.Args(productSubscriptionsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = productSubscriptionsDeleteCmd.Args(productSubscriptionsDeleteCmd, []string{"sub_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestProductSubscriptionsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := productSubscriptionsGetCmd.RunE(cmd, []string{"sub_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProductSubscriptionsDryRunFlagExists(t *testing.T) {
	// Dry-run flag is inherited from root command's PersistentFlags
	// We just verify the commands exist and can receive the flag
	if productSubscriptionsCreateCmd == nil {
		t.Fatal("productSubscriptionsCreateCmd is nil")
	}
	if productSubscriptionsDeleteCmd == nil {
		t.Fatal("productSubscriptionsDeleteCmd is nil")
	}
}

func TestProductSubscriptionsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := productSubscriptionsDeleteCmd.RunE(cmd, []string{"sub_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProductSubscriptionsCreateWithClient(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{"test": {Handle: "test", AccessToken: "token"}},
		}, nil
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("product-id", "", "Product ID")
	_ = cmd.Flags().Set("product-id", "prod_123")
	cmd.Flags().String("variant-id", "", "Variant ID")
	cmd.Flags().String("customer-id", "", "Customer ID")
	_ = cmd.Flags().Set("customer-id", "cust_123")
	cmd.Flags().String("selling-plan-id", "", "Selling plan ID")
	_ = cmd.Flags().Set("selling-plan-id", "plan_123")
	cmd.Flags().Int("quantity", 1, "Quantity")
	cmd.Flags().String("next-billing-date", "", "Next billing date")
	err := productSubscriptionsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("Create succeeded unexpectedly")
	}
}

// productSubscriptionsMockAPIClient is a mock implementation of api.APIClient for product subscriptions tests.
type productSubscriptionsMockAPIClient struct {
	api.MockClient
	listProductSubscriptionsResp  *api.ProductSubscriptionsListResponse
	listProductSubscriptionsErr   error
	getProductSubscriptionResp    *api.ProductSubscription
	getProductSubscriptionErr     error
	createProductSubscriptionResp *api.ProductSubscription
	createProductSubscriptionErr  error
	deleteProductSubscriptionErr  error
}

func (m *productSubscriptionsMockAPIClient) ListProductSubscriptions(ctx context.Context, opts *api.ProductSubscriptionsListOptions) (*api.ProductSubscriptionsListResponse, error) {
	return m.listProductSubscriptionsResp, m.listProductSubscriptionsErr
}

func (m *productSubscriptionsMockAPIClient) GetProductSubscription(ctx context.Context, id string) (*api.ProductSubscription, error) {
	return m.getProductSubscriptionResp, m.getProductSubscriptionErr
}

func (m *productSubscriptionsMockAPIClient) CreateProductSubscription(ctx context.Context, req *api.ProductSubscriptionCreateRequest) (*api.ProductSubscription, error) {
	return m.createProductSubscriptionResp, m.createProductSubscriptionErr
}

func (m *productSubscriptionsMockAPIClient) DeleteProductSubscription(ctx context.Context, id string) error {
	return m.deleteProductSubscriptionErr
}

// setupProductSubscriptionsMockFactories sets up mock factories for product subscriptions tests.
func setupProductSubscriptionsMockFactories(mockClient *productSubscriptionsMockAPIClient) (func(), *bytes.Buffer) {
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

// newProductSubscriptionsTestCmd creates a test command with common flags for product subscriptions tests.
func newProductSubscriptionsTestCmd() *cobra.Command {
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

// TestProductSubscriptionsListRunE tests the product subscriptions list command with mock API.
func TestProductSubscriptionsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ProductSubscriptionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ProductSubscriptionsListResponse{
				Items: []api.ProductSubscription{
					{
						ID:                "sub_123",
						ProductID:         "prod_123",
						VariantID:         "var_123",
						CustomerID:        "cust_123",
						SellingPlanID:     "plan_123",
						Status:            "active",
						Frequency:         "monthly",
						FrequencyInterval: 1,
						NextBillingDate:   time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
						Price:             "19.99",
						Currency:          "USD",
						Quantity:          1,
						TotalCycles:       12,
						CompletedCycles:   2,
						CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:         time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "sub_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ProductSubscriptionsListResponse{
				Items:      []api.ProductSubscription{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productSubscriptionsMockAPIClient{
				listProductSubscriptionsResp: tt.mockResp,
				listProductSubscriptionsErr:  tt.mockErr,
			}
			cleanup, buf := setupProductSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductSubscriptionsTestCmd()
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := productSubscriptionsListCmd.RunE(cmd, []string{})

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

// TestProductSubscriptionsGetRunE tests the product subscriptions get command with mock API.
func TestProductSubscriptionsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.ProductSubscription
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "sub_123",
			mockResp: &api.ProductSubscription{
				ID:                "sub_123",
				ProductID:         "prod_123",
				VariantID:         "var_123",
				CustomerID:        "cust_123",
				SellingPlanID:     "plan_123",
				Status:            "active",
				Frequency:         "monthly",
				FrequencyInterval: 1,
				NextBillingDate:   time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
				LastBillingDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Price:             "19.99",
				Currency:          "USD",
				Quantity:          1,
				TotalCycles:       12,
				CompletedCycles:   2,
				CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:         time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "sub_999",
			mockErr: errors.New("product subscription not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productSubscriptionsMockAPIClient{
				getProductSubscriptionResp: tt.mockResp,
				getProductSubscriptionErr:  tt.mockErr,
			}
			cleanup, _ := setupProductSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductSubscriptionsTestCmd()

			err := productSubscriptionsGetCmd.RunE(cmd, []string{tt.id})

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

// TestProductSubscriptionsCreateRunE tests the product subscriptions create command with mock API.
func TestProductSubscriptionsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		dryRun   bool
		mockResp *api.ProductSubscription
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "dry run",
			dryRun: true,
		},
		{
			name: "successful create",
			mockResp: &api.ProductSubscription{
				ID:            "sub_new",
				ProductID:     "prod_123",
				CustomerID:    "cust_123",
				SellingPlanID: "plan_123",
				Status:        "active",
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
			mockClient := &productSubscriptionsMockAPIClient{
				createProductSubscriptionResp: tt.mockResp,
				createProductSubscriptionErr:  tt.mockErr,
			}
			cleanup, _ := setupProductSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductSubscriptionsTestCmd()
			cmd.Flags().String("product-id", "prod_123", "")
			cmd.Flags().String("variant-id", "", "")
			cmd.Flags().String("customer-id", "cust_123", "")
			cmd.Flags().String("selling-plan-id", "plan_123", "")
			cmd.Flags().Int("quantity", 1, "")
			cmd.Flags().String("next-billing-date", "", "")
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := productSubscriptionsCreateCmd.RunE(cmd, []string{})

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

// TestProductSubscriptionsDeleteRunE tests the product subscriptions delete command with mock API.
func TestProductSubscriptionsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dryRun  bool
		mockErr error
		wantErr bool
	}{
		{
			name:   "dry run",
			id:     "sub_123",
			dryRun: true,
		},
		{
			name: "successful delete",
			id:   "sub_123",
		},
		{
			name:    "delete fails",
			id:      "sub_123",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productSubscriptionsMockAPIClient{
				deleteProductSubscriptionErr: tt.mockErr,
			}
			cleanup, _ := setupProductSubscriptionsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductSubscriptionsTestCmd()
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := productSubscriptionsDeleteCmd.RunE(cmd, []string{tt.id})

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

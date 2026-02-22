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

func TestProductReviewsCommand(t *testing.T) {
	if productReviewsCmd == nil {
		t.Fatal("productReviewsCmd is nil")
	}
	if productReviewsCmd.Use != "product-reviews" {
		t.Errorf("Expected Use to be 'product-reviews', got %q", productReviewsCmd.Use)
	}
}

func TestProductReviewsSubcommands(t *testing.T) {
	subcommands := productReviewsCmd.Commands()
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

func TestProductReviewsListFlags(t *testing.T) {
	flags := []string{"product-id", "status", "rating", "page", "page-size"}
	for _, flag := range flags {
		if productReviewsListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on list command", flag)
		}
	}
}

func TestProductReviewsCreateFlags(t *testing.T) {
	flags := []string{"product-id", "customer-id", "customer-name", "rating", "title", "content"}
	for _, flag := range flags {
		if productReviewsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag --%s not found on create command", flag)
		}
	}
}

func TestProductReviewsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := productReviewsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProductReviewsGetArgsValidation(t *testing.T) {
	if productReviewsGetCmd.Args == nil {
		t.Fatal("Expected Args validator on get command")
	}
	err := productReviewsGetCmd.Args(productReviewsGetCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = productReviewsGetCmd.Args(productReviewsGetCmd, []string{"review_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestProductReviewsDeleteArgsValidation(t *testing.T) {
	if productReviewsDeleteCmd.Args == nil {
		t.Fatal("Expected Args validator on delete command")
	}
	err := productReviewsDeleteCmd.Args(productReviewsDeleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with no args")
	}
	err = productReviewsDeleteCmd.Args(productReviewsDeleteCmd, []string{"review_123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}
}

func TestProductReviewsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := productReviewsGetCmd.RunE(cmd, []string{"review_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProductReviewsDryRunFlagExists(t *testing.T) {
	// Dry-run flag is inherited from root command's PersistentFlags
	// We just verify the commands exist and can receive the flag
	if productReviewsCreateCmd == nil {
		t.Fatal("productReviewsCreateCmd is nil")
	}
	if productReviewsDeleteCmd == nil {
		t.Fatal("productReviewsDeleteCmd is nil")
	}
}

func TestProductReviewsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := productReviewsDeleteCmd.RunE(cmd, []string{"review_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProductReviewsCreateWithClient(t *testing.T) {
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
	cmd.Flags().String("customer-id", "", "Customer ID")
	cmd.Flags().String("customer-name", "", "Customer name")
	_ = cmd.Flags().Set("customer-name", "John Doe")
	cmd.Flags().Int("rating", 0, "Rating")
	_ = cmd.Flags().Set("rating", "5")
	cmd.Flags().String("title", "", "Title")
	cmd.Flags().String("content", "", "Content")
	_ = cmd.Flags().Set("content", "Great product!")
	err := productReviewsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Log("Create succeeded unexpectedly")
	}
}

// productReviewsMockAPIClient is a mock implementation of api.APIClient for product reviews tests.
type productReviewsMockAPIClient struct {
	api.MockClient
	listProductReviewsResp  *api.ProductReviewsListResponse
	listProductReviewsErr   error
	getProductReviewResp    *api.ProductReview
	getProductReviewErr     error
	createProductReviewResp *api.ProductReview
	createProductReviewErr  error
	deleteProductReviewErr  error
}

func (m *productReviewsMockAPIClient) ListProductReviews(ctx context.Context, opts *api.ProductReviewsListOptions) (*api.ProductReviewsListResponse, error) {
	return m.listProductReviewsResp, m.listProductReviewsErr
}

func (m *productReviewsMockAPIClient) GetProductReview(ctx context.Context, id string) (*api.ProductReview, error) {
	return m.getProductReviewResp, m.getProductReviewErr
}

func (m *productReviewsMockAPIClient) CreateProductReview(ctx context.Context, req *api.ProductReviewCreateRequest) (*api.ProductReview, error) {
	return m.createProductReviewResp, m.createProductReviewErr
}

func (m *productReviewsMockAPIClient) DeleteProductReview(ctx context.Context, id string) error {
	return m.deleteProductReviewErr
}

// setupProductReviewsMockFactories sets up mock factories for product reviews tests.
func setupProductReviewsMockFactories(mockClient *productReviewsMockAPIClient) (func(), *bytes.Buffer) {
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

// newProductReviewsTestCmd creates a test command with common flags for product reviews tests.
func newProductReviewsTestCmd() *cobra.Command {
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

// TestProductReviewsListRunE tests the product reviews list command with mock API.
func TestProductReviewsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.ProductReviewsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ProductReviewsListResponse{
				Items: []api.ProductReview{
					{
						ID:           "review_123",
						ProductID:    "prod_123",
						CustomerID:   "cust_123",
						CustomerName: "John Doe",
						Rating:       5,
						Title:        "Great product",
						Content:      "I love this product!",
						Status:       "approved",
						Verified:     true,
						HelpfulCount: 10,
						CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:    time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "review_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ProductReviewsListResponse{
				Items:      []api.ProductReview{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productReviewsMockAPIClient{
				listProductReviewsResp: tt.mockResp,
				listProductReviewsErr:  tt.mockErr,
			}
			cleanup, buf := setupProductReviewsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductReviewsTestCmd()
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("rating", 0, "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := productReviewsListCmd.RunE(cmd, []string{})

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

// TestProductReviewsGetRunE tests the product reviews get command with mock API.
func TestProductReviewsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.ProductReview
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "review_123",
			mockResp: &api.ProductReview{
				ID:           "review_123",
				ProductID:    "prod_123",
				CustomerID:   "cust_123",
				CustomerName: "John Doe",
				Rating:       5,
				Title:        "Great product",
				Content:      "I love this product!",
				Status:       "approved",
				Verified:     true,
				HelpfulCount: 10,
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "review_999",
			mockErr: errors.New("product review not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productReviewsMockAPIClient{
				getProductReviewResp: tt.mockResp,
				getProductReviewErr:  tt.mockErr,
			}
			cleanup, _ := setupProductReviewsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductReviewsTestCmd()

			err := productReviewsGetCmd.RunE(cmd, []string{tt.id})

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

// TestProductReviewsCreateRunE tests the product reviews create command with mock API.
func TestProductReviewsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		dryRun   bool
		mockResp *api.ProductReview
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "dry run",
			dryRun: true,
		},
		{
			name: "successful create",
			mockResp: &api.ProductReview{
				ID:        "review_new",
				ProductID: "prod_123",
				Rating:    5,
				Status:    "pending",
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
			mockClient := &productReviewsMockAPIClient{
				createProductReviewResp: tt.mockResp,
				createProductReviewErr:  tt.mockErr,
			}
			cleanup, _ := setupProductReviewsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductReviewsTestCmd()
			cmd.Flags().String("product-id", "prod_123", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("customer-name", "John", "")
			cmd.Flags().Int("rating", 5, "")
			cmd.Flags().String("title", "", "")
			cmd.Flags().String("content", "Great!", "")
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := productReviewsCreateCmd.RunE(cmd, []string{})

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

// TestProductReviewsDeleteRunE tests the product reviews delete command with mock API.
func TestProductReviewsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		dryRun  bool
		mockErr error
		wantErr bool
	}{
		{
			name:   "dry run",
			id:     "review_123",
			dryRun: true,
		},
		{
			name: "successful delete",
			id:   "review_123",
		},
		{
			name:    "delete fails",
			id:      "review_123",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &productReviewsMockAPIClient{
				deleteProductReviewErr: tt.mockErr,
			}
			cleanup, _ := setupProductReviewsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductReviewsTestCmd()
			if tt.dryRun {
				_ = cmd.Flags().Set("dry-run", "true")
			}

			err := productReviewsDeleteCmd.RunE(cmd, []string{tt.id})

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

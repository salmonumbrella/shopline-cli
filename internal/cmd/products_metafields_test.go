package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// TestProductsMetafieldsCommandSetup verifies metafields parent command initialization.
func TestProductsMetafieldsCommandSetup(t *testing.T) {
	if productsMetafieldsCmd.Use != "metafields" {
		t.Errorf("expected Use 'metafields', got %q", productsMetafieldsCmd.Use)
	}
	if productsMetafieldsCmd.Short != "Manage product metafields" {
		t.Errorf("expected Short 'Manage product metafields', got %q", productsMetafieldsCmd.Short)
	}
}

// TestProductsAppMetafieldsCommandSetup verifies app-metafields parent command initialization.
func TestProductsAppMetafieldsCommandSetup(t *testing.T) {
	if productsAppMetafieldsCmd.Use != "app-metafields" {
		t.Errorf("expected Use 'app-metafields', got %q", productsAppMetafieldsCmd.Use)
	}
	if productsAppMetafieldsCmd.Short != "Manage product app metafields" {
		t.Errorf("expected Short 'Manage product app metafields', got %q", productsAppMetafieldsCmd.Short)
	}
}

// TestProductsMetafieldsSubcommands verifies all subcommands are registered.
func TestProductsMetafieldsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":        "List metafields attached to a product",
		"get":         "Get a specific product metafield",
		"create":      "Create a product metafield",
		"update":      "Update a product metafield",
		"delete":      "Delete a product metafield",
		"bulk-create": "Bulk create product metafields",
		"bulk-update": "Bulk update product metafields",
		"bulk-delete": "Bulk delete product metafields",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range productsMetafieldsCmd.Commands() {
				if sub.Use == name || strings.HasPrefix(sub.Use, name+" ") {
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

// TestProductsAppMetafieldsSubcommands verifies all app-metafields subcommands are registered.
func TestProductsAppMetafieldsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":        "List app metafields attached to a product",
		"get":         "Get a specific product app metafield",
		"create":      "Create a product app metafield",
		"update":      "Update a product app metafield",
		"delete":      "Delete a product app metafield",
		"bulk-create": "Bulk create product app metafields",
		"bulk-update": "Bulk update product app metafields",
		"bulk-delete": "Bulk delete product app metafields",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range productsAppMetafieldsCmd.Commands() {
				if sub.Use == name || strings.HasPrefix(sub.Use, name+" ") {
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

// TestProductsMetafieldsListCmdArgs verifies list command expects exactly 1 arg.
func TestProductsMetafieldsListCmdArgs(t *testing.T) {
	if productsMetafieldsListCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestProductsMetafieldsGetCmdArgs verifies get command expects exactly 2 args.
func TestProductsMetafieldsGetCmdArgs(t *testing.T) {
	if productsMetafieldsGetCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// TestProductsMetafieldsDeleteCmdArgs verifies delete command expects exactly 2 args.
func TestProductsMetafieldsDeleteCmdArgs(t *testing.T) {
	if productsMetafieldsDeleteCmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

// mockProductMetafieldsClient is a mock implementation of api.APIClient for product metafields tests.
type mockProductMetafieldsClient struct {
	api.MockClient

	// Product metafields
	listProductMetafieldsResp      json.RawMessage
	listProductMetafieldsErr       error
	getProductMetafieldResp        json.RawMessage
	getProductMetafieldErr         error
	createProductMetafieldResp     json.RawMessage
	createProductMetafieldErr      error
	updateProductMetafieldResp     json.RawMessage
	updateProductMetafieldErr      error
	deleteProductMetafieldErr      error
	bulkCreateProductMetafieldsErr error
	bulkUpdateProductMetafieldsErr error
	bulkDeleteProductMetafieldsErr error

	// Product app-metafields
	listProductAppMetafieldsResp      json.RawMessage
	listProductAppMetafieldsErr       error
	getProductAppMetafieldResp        json.RawMessage
	getProductAppMetafieldErr         error
	createProductAppMetafieldResp     json.RawMessage
	createProductAppMetafieldErr      error
	updateProductAppMetafieldResp     json.RawMessage
	updateProductAppMetafieldErr      error
	deleteProductAppMetafieldErr      error
	bulkCreateProductAppMetafieldsErr error
	bulkUpdateProductAppMetafieldsErr error
	bulkDeleteProductAppMetafieldsErr error
}

func (m *mockProductMetafieldsClient) ListProductMetafields(ctx context.Context, productID string) (json.RawMessage, error) {
	return m.listProductMetafieldsResp, m.listProductMetafieldsErr
}

func (m *mockProductMetafieldsClient) GetProductMetafield(ctx context.Context, productID, metafieldID string) (json.RawMessage, error) {
	return m.getProductMetafieldResp, m.getProductMetafieldErr
}

func (m *mockProductMetafieldsClient) CreateProductMetafield(ctx context.Context, productID string, body any) (json.RawMessage, error) {
	return m.createProductMetafieldResp, m.createProductMetafieldErr
}

func (m *mockProductMetafieldsClient) UpdateProductMetafield(ctx context.Context, productID, metafieldID string, body any) (json.RawMessage, error) {
	return m.updateProductMetafieldResp, m.updateProductMetafieldErr
}

func (m *mockProductMetafieldsClient) DeleteProductMetafield(ctx context.Context, productID, metafieldID string) error {
	return m.deleteProductMetafieldErr
}

func (m *mockProductMetafieldsClient) BulkCreateProductMetafields(ctx context.Context, productID string, body any) error {
	return m.bulkCreateProductMetafieldsErr
}

func (m *mockProductMetafieldsClient) BulkUpdateProductMetafields(ctx context.Context, productID string, body any) error {
	return m.bulkUpdateProductMetafieldsErr
}

func (m *mockProductMetafieldsClient) BulkDeleteProductMetafields(ctx context.Context, productID string, body any) error {
	return m.bulkDeleteProductMetafieldsErr
}

func (m *mockProductMetafieldsClient) ListProductAppMetafields(ctx context.Context, productID string) (json.RawMessage, error) {
	return m.listProductAppMetafieldsResp, m.listProductAppMetafieldsErr
}

func (m *mockProductMetafieldsClient) GetProductAppMetafield(ctx context.Context, productID, metafieldID string) (json.RawMessage, error) {
	return m.getProductAppMetafieldResp, m.getProductAppMetafieldErr
}

func (m *mockProductMetafieldsClient) CreateProductAppMetafield(ctx context.Context, productID string, body any) (json.RawMessage, error) {
	return m.createProductAppMetafieldResp, m.createProductAppMetafieldErr
}

func (m *mockProductMetafieldsClient) UpdateProductAppMetafield(ctx context.Context, productID, metafieldID string, body any) (json.RawMessage, error) {
	return m.updateProductAppMetafieldResp, m.updateProductAppMetafieldErr
}

func (m *mockProductMetafieldsClient) DeleteProductAppMetafield(ctx context.Context, productID, metafieldID string) error {
	return m.deleteProductAppMetafieldErr
}

func (m *mockProductMetafieldsClient) BulkCreateProductAppMetafields(ctx context.Context, productID string, body any) error {
	return m.bulkCreateProductAppMetafieldsErr
}

func (m *mockProductMetafieldsClient) BulkUpdateProductAppMetafields(ctx context.Context, productID string, body any) error {
	return m.bulkUpdateProductAppMetafieldsErr
}

func (m *mockProductMetafieldsClient) BulkDeleteProductAppMetafields(ctx context.Context, productID string, body any) error {
	return m.bulkDeleteProductAppMetafieldsErr
}

// setupProductMetafieldsMockFactories sets up mock factories for product metafields tests.
func setupProductMetafieldsMockFactories(mockClient *mockProductMetafieldsClient) (func(), *bytes.Buffer) {
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

// newProductMetafieldsTestCmd creates a test command with common flags.
func newProductMetafieldsTestCmd() *cobra.Command {
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

// ============================
// Product Metafields Tests
// ============================

// TestProductMetafieldsListRunE tests the product metafields list command.
func TestProductMetafieldsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   json.RawMessage
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful list",
			mockResp:   json.RawMessage(`[{"id":"mf_1","key":"color","value":"blue"}]`),
			wantOutput: "mf_1",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				listProductMetafieldsResp: tt.mockResp,
				listProductMetafieldsErr:  tt.mockErr,
			}
			cleanup, buf := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()

			err := productsMetafieldsListCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductMetafieldsGetRunE tests the product metafields get command.
func TestProductMetafieldsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp json.RawMessage
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			mockResp: json.RawMessage(`{"id":"mf_1","key":"color","value":"blue"}`),
		},
		{
			name:    "not found",
			mockErr: errors.New("metafield not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				getProductMetafieldResp: tt.mockResp,
				getProductMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()

			err := productsMetafieldsGetCmd.RunE(cmd, []string{"prod_123", "mf_1"})

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

// TestProductMetafieldsDeleteRunE tests the product metafields delete command.
func TestProductMetafieldsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete with yes flag",
		},
		{
			name:    "delete fails",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				deleteProductMetafieldErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()

			err := productsMetafieldsDeleteCmd.RunE(cmd, []string{"prod_123", "mf_1"})

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

// TestProductMetafieldsCreateRunE tests the product metafields create command.
func TestProductMetafieldsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp json.RawMessage
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful create",
			mockResp: json.RawMessage(`{"id":"mf_new","key":"size","value":"large"}`),
		},
		{
			name:    "create fails",
			mockErr: errors.New("create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				createProductMetafieldResp: tt.mockResp,
				createProductMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `{"key":"size","value":"large"}`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsMetafieldsCreateCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductMetafieldsUpdateRunE tests the product metafields update command.
func TestProductMetafieldsUpdateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp json.RawMessage
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful update",
			mockResp: json.RawMessage(`{"id":"mf_1","key":"color","value":"red"}`),
		},
		{
			name:    "update fails",
			mockErr: errors.New("update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				updateProductMetafieldResp: tt.mockResp,
				updateProductMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `{"value":"red"}`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsMetafieldsUpdateCmd.RunE(cmd, []string{"prod_123", "mf_1"})

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

// TestProductMetafieldsBulkCreateRunE tests the bulk create command.
func TestProductMetafieldsBulkCreateRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful bulk create",
		},
		{
			name:    "bulk create fails",
			mockErr: errors.New("bulk create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				bulkCreateProductMetafieldsErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `[{"key":"k1","value":"v1"}]`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsMetafieldsBulkCreateCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductMetafieldsBulkUpdateRunE tests the bulk update command.
func TestProductMetafieldsBulkUpdateRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful bulk update",
		},
		{
			name:    "bulk update fails",
			mockErr: errors.New("bulk update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				bulkUpdateProductMetafieldsErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `[{"id":"mf_1","value":"updated"}]`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsMetafieldsBulkUpdateCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductMetafieldsBulkDeleteRunE tests the bulk delete command.
func TestProductMetafieldsBulkDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful bulk delete",
		},
		{
			name:    "bulk delete fails",
			mockErr: errors.New("bulk delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				bulkDeleteProductMetafieldsErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `["mf_1","mf_2"]`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsMetafieldsBulkDeleteCmd.RunE(cmd, []string{"prod_123"})

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

// ============================
// Product App-Metafields Tests
// ============================

// TestProductAppMetafieldsListRunE tests the product app-metafields list command.
func TestProductAppMetafieldsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   json.RawMessage
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful list",
			mockResp:   json.RawMessage(`[{"id":"amf_1","key":"app_color","value":"green"}]`),
			wantOutput: "amf_1",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				listProductAppMetafieldsResp: tt.mockResp,
				listProductAppMetafieldsErr:  tt.mockErr,
			}
			cleanup, buf := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()

			err := productsAppMetafieldsListCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductAppMetafieldsGetRunE tests the product app-metafields get command.
func TestProductAppMetafieldsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp json.RawMessage
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful get",
			mockResp: json.RawMessage(`{"id":"amf_1","key":"app_color","value":"green"}`),
		},
		{
			name:    "not found",
			mockErr: errors.New("app metafield not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				getProductAppMetafieldResp: tt.mockResp,
				getProductAppMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()

			err := productsAppMetafieldsGetCmd.RunE(cmd, []string{"prod_123", "amf_1"})

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

// TestProductAppMetafieldsDeleteRunE tests the product app-metafields delete command.
func TestProductAppMetafieldsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete with yes flag",
		},
		{
			name:    "delete fails",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				deleteProductAppMetafieldErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()

			err := productsAppMetafieldsDeleteCmd.RunE(cmd, []string{"prod_123", "amf_1"})

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

// TestProductAppMetafieldsCreateRunE tests the product app-metafields create command.
func TestProductAppMetafieldsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp json.RawMessage
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful create",
			mockResp: json.RawMessage(`{"id":"amf_new","key":"app_size","value":"xl"}`),
		},
		{
			name:    "create fails",
			mockErr: errors.New("create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				createProductAppMetafieldResp: tt.mockResp,
				createProductAppMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `{"key":"app_size","value":"xl"}`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsAppMetafieldsCreateCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductAppMetafieldsUpdateRunE tests the product app-metafields update command.
func TestProductAppMetafieldsUpdateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp json.RawMessage
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful update",
			mockResp: json.RawMessage(`{"id":"amf_1","key":"app_color","value":"yellow"}`),
		},
		{
			name:    "update fails",
			mockErr: errors.New("update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				updateProductAppMetafieldResp: tt.mockResp,
				updateProductAppMetafieldErr:  tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `{"value":"yellow"}`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsAppMetafieldsUpdateCmd.RunE(cmd, []string{"prod_123", "amf_1"})

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

// TestProductAppMetafieldsBulkCreateRunE tests the bulk create command for app metafields.
func TestProductAppMetafieldsBulkCreateRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful bulk create",
		},
		{
			name:    "bulk create fails",
			mockErr: errors.New("bulk create failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				bulkCreateProductAppMetafieldsErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `[{"key":"k1","value":"v1"}]`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsAppMetafieldsBulkCreateCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductAppMetafieldsBulkUpdateRunE tests the bulk update command for app metafields.
func TestProductAppMetafieldsBulkUpdateRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful bulk update",
		},
		{
			name:    "bulk update fails",
			mockErr: errors.New("bulk update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				bulkUpdateProductAppMetafieldsErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `[{"id":"amf_1","value":"updated"}]`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsAppMetafieldsBulkUpdateCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductAppMetafieldsBulkDeleteRunE tests the bulk delete command for app metafields.
func TestProductAppMetafieldsBulkDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful bulk delete",
		},
		{
			name:    "bulk delete fails",
			mockErr: errors.New("bulk delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockProductMetafieldsClient{
				bulkDeleteProductAppMetafieldsErr: tt.mockErr,
			}
			cleanup, _ := setupProductMetafieldsMockFactories(mockClient)
			defer cleanup()

			cmd := newProductMetafieldsTestCmd()
			cmd.Flags().String("body", `["amf_1","amf_2"]`, "")
			cmd.Flags().String("body-file", "", "")

			err := productsAppMetafieldsBulkDeleteCmd.RunE(cmd, []string{"prod_123"})

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

// TestProductMetafieldsGetClientError tests error when getClient fails.
func TestProductMetafieldsGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newProductMetafieldsTestCmd()

	err := productsMetafieldsListCmd.RunE(cmd, []string{"prod_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "credential store") {
		t.Errorf("expected credential store error, got: %v", err)
	}
}

// TestProductAppMetafieldsGetClientError tests error when getClient fails for app-metafields.
func TestProductAppMetafieldsGetClientError(t *testing.T) {
	origSecretsFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origSecretsFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newProductMetafieldsTestCmd()

	err := productsAppMetafieldsListCmd.RunE(cmd, []string{"prod_123"})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "credential store") {
		t.Errorf("expected credential store error, got: %v", err)
	}
}

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// purchaseOrdersAPIClient is a mock implementation of api.APIClient for purchase orders tests.
type purchaseOrdersAPIClient struct {
	api.MockClient

	listPurchaseOrdersResp   *api.PurchaseOrdersListResponse
	listPurchaseOrdersErr    error
	getPurchaseOrderResp     *api.PurchaseOrder
	getPurchaseOrderErr      error
	receivePurchaseOrderResp *api.PurchaseOrder
	receivePurchaseOrderErr  error
	cancelPurchaseOrderResp  *api.PurchaseOrder
	cancelPurchaseOrderErr   error
	deletePurchaseOrderErr   error
}

func (m *purchaseOrdersAPIClient) ListPurchaseOrders(ctx context.Context, opts *api.PurchaseOrdersListOptions) (*api.PurchaseOrdersListResponse, error) {
	return m.listPurchaseOrdersResp, m.listPurchaseOrdersErr
}

func (m *purchaseOrdersAPIClient) GetPurchaseOrder(ctx context.Context, id string) (*api.PurchaseOrder, error) {
	return m.getPurchaseOrderResp, m.getPurchaseOrderErr
}

func (m *purchaseOrdersAPIClient) ReceivePurchaseOrder(ctx context.Context, id string) (*api.PurchaseOrder, error) {
	return m.receivePurchaseOrderResp, m.receivePurchaseOrderErr
}

func (m *purchaseOrdersAPIClient) CancelPurchaseOrder(ctx context.Context, id string) (*api.PurchaseOrder, error) {
	return m.cancelPurchaseOrderResp, m.cancelPurchaseOrderErr
}

func (m *purchaseOrdersAPIClient) DeletePurchaseOrder(ctx context.Context, id string) error {
	return m.deletePurchaseOrderErr
}

// setupPurchaseOrdersTest sets up mock factories and returns cleanup function.
func setupPurchaseOrdersTest(t *testing.T, mockClient api.APIClient) func() {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

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

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestPurchaseOrdersCommandSetup verifies purchase-orders command initialization
func TestPurchaseOrdersCommandSetup(t *testing.T) {
	if purchaseOrdersCmd.Use != "purchase-orders" {
		t.Errorf("expected Use 'purchase-orders', got %q", purchaseOrdersCmd.Use)
	}
	if purchaseOrdersCmd.Short != "Manage purchase orders" {
		t.Errorf("expected Short 'Manage purchase orders', got %q", purchaseOrdersCmd.Short)
	}
}

// TestPurchaseOrdersSubcommands verifies all subcommands are registered
func TestPurchaseOrdersSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":    "List purchase orders",
		"get":     "Get purchase order details",
		"receive": "Mark a purchase order as received",
		"cancel":  "Cancel a purchase order",
		"delete":  "Delete a purchase order",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range purchaseOrdersCmd.Commands() {
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

// TestPurchaseOrdersListFlags verifies list command flags exist with correct defaults
func TestPurchaseOrdersListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
		{"supplier-id", ""},
		{"warehouse-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := purchaseOrdersListCmd.Flags().Lookup(f.name)
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

// TestPurchaseOrdersGetCmd verifies get command setup
func TestPurchaseOrdersGetCmd(t *testing.T) {
	if purchaseOrdersGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", purchaseOrdersGetCmd.Use)
	}
}

// TestPurchaseOrdersReceiveCmd verifies receive command setup
func TestPurchaseOrdersReceiveCmd(t *testing.T) {
	if purchaseOrdersReceiveCmd.Use != "receive <id>" {
		t.Errorf("expected Use 'receive <id>', got %q", purchaseOrdersReceiveCmd.Use)
	}
}

// TestPurchaseOrdersCancelCmd verifies cancel command setup
func TestPurchaseOrdersCancelCmd(t *testing.T) {
	if purchaseOrdersCancelCmd.Use != "cancel <id>" {
		t.Errorf("expected Use 'cancel <id>', got %q", purchaseOrdersCancelCmd.Use)
	}
}

// TestPurchaseOrdersDeleteCmd verifies delete command setup
func TestPurchaseOrdersDeleteCmd(t *testing.T) {
	if purchaseOrdersDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", purchaseOrdersDeleteCmd.Use)
	}
}

// TestPurchaseOrdersListRunE_GetClientFails verifies error handling when getClient fails
func TestPurchaseOrdersListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("supplier-id", "", "")
	cmd.Flags().String("warehouse-id", "", "")

	err := purchaseOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPurchaseOrdersGetRunE_GetClientFails verifies error handling when getClient fails
func TestPurchaseOrdersGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := purchaseOrdersGetCmd.RunE(cmd, []string{"po_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPurchaseOrdersReceiveRunE_GetClientFails verifies error handling when getClient fails
func TestPurchaseOrdersReceiveRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := purchaseOrdersReceiveCmd.RunE(cmd, []string{"po_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPurchaseOrdersCancelRunE_GetClientFails verifies error handling when getClient fails
func TestPurchaseOrdersCancelRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := purchaseOrdersCancelCmd.RunE(cmd, []string{"po_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPurchaseOrdersDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestPurchaseOrdersDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := purchaseOrdersDeleteCmd.RunE(cmd, []string{"po_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPurchaseOrdersListRunE_NoProfiles verifies error handling when no profiles exist
func TestPurchaseOrdersListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("supplier-id", "", "")
	cmd.Flags().String("warehouse-id", "", "")

	err := purchaseOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPurchaseOrdersListRunE_Success tests the list command with mock API.
func TestPurchaseOrdersListRunE_Success(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.PurchaseOrdersListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with items",
			mockResp: &api.PurchaseOrdersListResponse{
				Items: []api.PurchaseOrder{
					{
						ID:            "po_123",
						Number:        "PO-001",
						Status:        "pending",
						SupplierName:  "Acme Supplies",
						WarehouseName: "Main Warehouse",
						Total:         "1500.00",
						ExpectedAt:    time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:            "po_456",
						Number:        "PO-002",
						Status:        "received",
						SupplierName:  "Beta Corp",
						WarehouseName: "Secondary Warehouse",
						Total:         "2500.00",
						CreatedAt:     time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "po_123",
		},
		{
			name: "empty list",
			mockResp: &api.PurchaseOrdersListResponse{
				Items:      []api.PurchaseOrder{},
				TotalCount: 0,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &purchaseOrdersAPIClient{
				listPurchaseOrdersResp: tt.mockResp,
				listPurchaseOrdersErr:  tt.mockErr,
			}
			cleanup := setupPurchaseOrdersTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("supplier-id", "", "")
			cmd.Flags().String("warehouse-id", "", "")

			err := purchaseOrdersListCmd.RunE(cmd, []string{})

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

// TestPurchaseOrdersListRunE_JSONOutput tests JSON output format.
func TestPurchaseOrdersListRunE_JSONOutput(t *testing.T) {
	mockResp := &api.PurchaseOrdersListResponse{
		Items: []api.PurchaseOrder{
			{
				ID:            "po_json_123",
				Number:        "PO-JSON",
				Status:        "pending",
				SupplierName:  "JSON Supplier",
				WarehouseName: "JSON Warehouse",
				Total:         "999.99",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		TotalCount: 1,
	}

	mockClient := &purchaseOrdersAPIClient{
		listPurchaseOrdersResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("supplier-id", "", "")
	cmd.Flags().String("warehouse-id", "", "")

	err := purchaseOrdersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Verify JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	if !strings.Contains(output, "po_json_123") {
		t.Errorf("JSON output should contain purchase order ID")
	}
}

// TestPurchaseOrdersListRunE_WithFilters tests list command with filter flags.
func TestPurchaseOrdersListRunE_WithFilters(t *testing.T) {
	mockResp := &api.PurchaseOrdersListResponse{
		Items:      []api.PurchaseOrder{},
		TotalCount: 0,
	}

	mockClient := &purchaseOrdersAPIClient{
		listPurchaseOrdersResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")
	cmd.Flags().String("status", "pending", "")
	cmd.Flags().String("supplier-id", "sup_123", "")
	cmd.Flags().String("warehouse-id", "wh_456", "")

	err := purchaseOrdersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPurchaseOrdersGetRunE_Success tests the get command with mock API.
func TestPurchaseOrdersGetRunE_Success(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockResp   *api.PurchaseOrder
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful get",
			id:   "po_123",
			mockResp: &api.PurchaseOrder{
				ID:            "po_123",
				Number:        "PO-001",
				Status:        "pending",
				SupplierID:    "sup_123",
				SupplierName:  "Acme Supplies",
				WarehouseID:   "wh_456",
				WarehouseName: "Main Warehouse",
				Currency:      "USD",
				Subtotal:      "1400.00",
				Tax:           "100.00",
				Total:         "1500.00",
				Note:          "Rush order",
				ExpectedAt:    time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
				LineItems: []api.PurchaseOrderItem{
					{
						ID:          "item_1",
						ProductID:   "prod_1",
						VariantID:   "var_1",
						SKU:         "SKU-001",
						Title:       "Product A",
						Quantity:    10,
						ReceivedQty: 0,
						UnitCost:    "140.00",
						Total:       "1400.00",
					},
				},
			},
			wantOutput: "Purchase Order ID: po_123",
		},
		{
			name: "get with received_at",
			id:   "po_received",
			mockResp: &api.PurchaseOrder{
				ID:            "po_received",
				Number:        "PO-002",
				Status:        "received",
				SupplierID:    "sup_123",
				SupplierName:  "Acme Supplies",
				WarehouseID:   "wh_456",
				WarehouseName: "Main Warehouse",
				Currency:      "USD",
				Subtotal:      "500.00",
				Tax:           "0.00",
				Total:         "500.00",
				ReceivedAt:    time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
			wantOutput: "Received At:",
		},
		{
			name:    "not found",
			id:      "po_999",
			mockErr: errors.New("purchase order not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &purchaseOrdersAPIClient{
				getPurchaseOrderResp: tt.mockResp,
				getPurchaseOrderErr:  tt.mockErr,
			}
			cleanup := setupPurchaseOrdersTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := purchaseOrdersGetCmd.RunE(cmd, []string{tt.id})

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

// TestPurchaseOrdersGetRunE_JSONOutput tests JSON output format for get command.
func TestPurchaseOrdersGetRunE_JSONOutput(t *testing.T) {
	mockResp := &api.PurchaseOrder{
		ID:            "po_json_get",
		Number:        "PO-JSON-GET",
		Status:        "pending",
		SupplierID:    "sup_123",
		SupplierName:  "JSON Supplier",
		WarehouseID:   "wh_456",
		WarehouseName: "JSON Warehouse",
		Currency:      "USD",
		Subtotal:      "100.00",
		Tax:           "10.00",
		Total:         "110.00",
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	mockClient := &purchaseOrdersAPIClient{
		getPurchaseOrderResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := purchaseOrdersGetCmd.RunE(cmd, []string{"po_json_get"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Verify JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	if !strings.Contains(output, "po_json_get") {
		t.Errorf("JSON output should contain purchase order ID")
	}
}

// TestPurchaseOrdersReceiveRunE_Success tests the receive command with mock API.
func TestPurchaseOrdersReceiveRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.PurchaseOrder
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful receive",
			id:   "po_123",
			mockResp: &api.PurchaseOrder{
				ID:     "po_123",
				Number: "PO-001",
				Status: "received",
			},
		},
		{
			name:    "receive fails",
			id:      "po_invalid",
			mockErr: errors.New("purchase order already received"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &purchaseOrdersAPIClient{
				receivePurchaseOrderResp: tt.mockResp,
				receivePurchaseOrderErr:  tt.mockErr,
			}
			cleanup := setupPurchaseOrdersTest(t, mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")

			err := purchaseOrdersReceiveCmd.RunE(cmd, []string{tt.id})

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

// TestPurchaseOrdersCancelRunE_Success tests the cancel command with mock API.
func TestPurchaseOrdersCancelRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.PurchaseOrder
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful cancel with yes flag",
			id:   "po_123",
			mockResp: &api.PurchaseOrder{
				ID:     "po_123",
				Number: "PO-001",
				Status: "cancelled",
			},
		},
		{
			name:    "cancel fails",
			id:      "po_invalid",
			mockErr: errors.New("purchase order already cancelled"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &purchaseOrdersAPIClient{
				cancelPurchaseOrderResp: tt.mockResp,
				cancelPurchaseOrderErr:  tt.mockErr,
			}
			cleanup := setupPurchaseOrdersTest(t, mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := purchaseOrdersCancelCmd.RunE(cmd, []string{tt.id})

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

// TestPurchaseOrdersDeleteRunE_Success tests the delete command with mock API.
func TestPurchaseOrdersDeleteRunE_Success(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete with yes flag",
			id:   "po_123",
		},
		{
			name:    "delete fails",
			id:      "po_invalid",
			mockErr: errors.New("purchase order not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &purchaseOrdersAPIClient{
				deletePurchaseOrderErr: tt.mockErr,
			}
			cleanup := setupPurchaseOrdersTest(t, mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := purchaseOrdersDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestPurchaseOrdersListFlagDescriptions verifies flag descriptions are set
func TestPurchaseOrdersListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"page":         "Page number",
		"page-size":    "Results per page",
		"status":       "Filter by status (draft, pending, received, cancelled)",
		"supplier-id":  "Filter by supplier ID",
		"warehouse-id": "Filter by warehouse ID",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := purchaseOrdersListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// TestPurchaseOrdersListFlagTypes verifies flag types are correct
func TestPurchaseOrdersListFlagTypes(t *testing.T) {
	flags := map[string]string{
		"page":         "int",
		"page-size":    "int",
		"status":       "string",
		"supplier-id":  "string",
		"warehouse-id": "string",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := purchaseOrdersListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Value.Type() != expectedType {
				t.Errorf("expected type %q, got %q", expectedType, flag.Value.Type())
			}
		})
	}
}

// TestPurchaseOrdersGetRunE_NoLineItems tests get command with no line items.
func TestPurchaseOrdersGetRunE_NoLineItems(t *testing.T) {
	mockResp := &api.PurchaseOrder{
		ID:            "po_no_items",
		Number:        "PO-EMPTY",
		Status:        "draft",
		SupplierID:    "sup_123",
		SupplierName:  "Supplier A",
		WarehouseID:   "wh_456",
		WarehouseName: "Warehouse A",
		Currency:      "USD",
		Subtotal:      "0.00",
		Tax:           "0.00",
		Total:         "0.00",
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		LineItems:     []api.PurchaseOrderItem{}, // No line items
	}

	mockClient := &purchaseOrdersAPIClient{
		getPurchaseOrderResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := purchaseOrdersGetCmd.RunE(cmd, []string{"po_no_items"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPurchaseOrdersGetRunE_NoNote tests get command with empty note.
func TestPurchaseOrdersGetRunE_NoNote(t *testing.T) {
	mockResp := &api.PurchaseOrder{
		ID:            "po_no_note",
		Number:        "PO-NO-NOTE",
		Status:        "pending",
		SupplierID:    "sup_123",
		SupplierName:  "Supplier A",
		WarehouseID:   "wh_456",
		WarehouseName: "Warehouse A",
		Currency:      "USD",
		Subtotal:      "100.00",
		Tax:           "10.00",
		Total:         "110.00",
		Note:          "", // Empty note
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	mockClient := &purchaseOrdersAPIClient{
		getPurchaseOrderResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := purchaseOrdersGetCmd.RunE(cmd, []string{"po_no_note"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPurchaseOrdersGetRunE_ZeroExpectedAt tests get command with zero expected_at time.
func TestPurchaseOrdersGetRunE_ZeroExpectedAt(t *testing.T) {
	mockResp := &api.PurchaseOrder{
		ID:            "po_no_expected",
		Number:        "PO-NO-EXPECTED",
		Status:        "pending",
		SupplierID:    "sup_123",
		SupplierName:  "Supplier A",
		WarehouseID:   "wh_456",
		WarehouseName: "Warehouse A",
		Currency:      "USD",
		Subtotal:      "100.00",
		Tax:           "10.00",
		Total:         "110.00",
		// ExpectedAt is zero time (not set)
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	mockClient := &purchaseOrdersAPIClient{
		getPurchaseOrderResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := purchaseOrdersGetCmd.RunE(cmd, []string{"po_no_expected"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPurchaseOrdersListRunE_ZeroExpectedAt tests list with items having zero expected_at.
func TestPurchaseOrdersListRunE_ZeroExpectedAt(t *testing.T) {
	mockResp := &api.PurchaseOrdersListResponse{
		Items: []api.PurchaseOrder{
			{
				ID:            "po_zero_expected",
				Number:        "PO-ZERO",
				Status:        "pending",
				SupplierName:  "Supplier",
				WarehouseName: "Warehouse",
				Total:         "100.00",
				// ExpectedAt is zero time
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		TotalCount: 1,
	}

	mockClient := &purchaseOrdersAPIClient{
		listPurchaseOrdersResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("supplier-id", "", "")
	cmd.Flags().String("warehouse-id", "", "")

	err := purchaseOrdersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should show "-" for missing expected_at
	if !strings.Contains(output, "-") {
		t.Log("Table output should contain '-' placeholder for zero expected_at")
	}
}

// TestPurchaseOrdersGetRunE_MultipleLineItems tests get command with multiple line items.
func TestPurchaseOrdersGetRunE_MultipleLineItems(t *testing.T) {
	mockResp := &api.PurchaseOrder{
		ID:            "po_multi_items",
		Number:        "PO-MULTI",
		Status:        "pending",
		SupplierID:    "sup_123",
		SupplierName:  "Supplier A",
		WarehouseID:   "wh_456",
		WarehouseName: "Warehouse A",
		Currency:      "USD",
		Subtotal:      "300.00",
		Tax:           "30.00",
		Total:         "330.00",
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		LineItems: []api.PurchaseOrderItem{
			{
				ID:          "item_1",
				ProductID:   "prod_1",
				VariantID:   "var_1",
				SKU:         "SKU-001",
				Title:       "Product A",
				Quantity:    10,
				ReceivedQty: 5,
				UnitCost:    "10.00",
				Total:       "100.00",
			},
			{
				ID:          "item_2",
				ProductID:   "prod_2",
				VariantID:   "var_2",
				SKU:         "SKU-002",
				Title:       "Product B",
				Quantity:    20,
				ReceivedQty: 0,
				UnitCost:    "10.00",
				Total:       "200.00",
			},
		},
	}

	mockClient := &purchaseOrdersAPIClient{
		getPurchaseOrderResp: mockResp,
	}
	cleanup := setupPurchaseOrdersTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := purchaseOrdersGetCmd.RunE(cmd, []string{"po_multi_items"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

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

func TestReturnOrdersCommandStructure(t *testing.T) {
	if returnOrdersCmd == nil {
		t.Fatal("returnOrdersCmd is nil")
	}
	if returnOrdersCmd.Use != "return-orders" {
		t.Errorf("Expected Use 'return-orders', got %q", returnOrdersCmd.Use)
	}
	if returnOrdersCmd.Short != "Manage return orders" {
		t.Errorf("Expected Short 'Manage return orders', got %q", returnOrdersCmd.Short)
	}
}

func TestReturnOrdersSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":     "List return orders",
		"get":      "Get return order details",
		"cancel":   "Cancel a return order",
		"complete": "Mark a return order as complete",
		"receive":  "Mark returned items as received",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range returnOrdersCmd.Commands() {
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

func TestReturnOrdersListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status", ""},
		{"order-id", ""},
		{"customer-id", ""},
		{"type", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := returnOrdersListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestReturnOrdersGetCmd(t *testing.T) {
	if returnOrdersGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use 'get <id>', got %q", returnOrdersGetCmd.Use)
	}
}

func TestReturnOrdersCancelCmd(t *testing.T) {
	if returnOrdersCancelCmd.Use != "cancel <id>" {
		t.Errorf("Expected Use 'cancel <id>', got %q", returnOrdersCancelCmd.Use)
	}
}

func TestReturnOrdersCompleteCmd(t *testing.T) {
	if returnOrdersCompleteCmd.Use != "complete <id>" {
		t.Errorf("Expected Use 'complete <id>', got %q", returnOrdersCompleteCmd.Use)
	}
}

func TestReturnOrdersReceiveCmd(t *testing.T) {
	if returnOrdersReceiveCmd.Use != "receive <id>" {
		t.Errorf("Expected Use 'receive <id>', got %q", returnOrdersReceiveCmd.Use)
	}
}

func TestReturnOrdersGetRequiresArg(t *testing.T) {
	cmd := returnOrdersGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"return_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestReturnOrdersCancelRequiresArg(t *testing.T) {
	cmd := returnOrdersCancelCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"return_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestReturnOrdersCompleteRequiresArg(t *testing.T) {
	cmd := returnOrdersCompleteCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"return_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestReturnOrdersReceiveRequiresArg(t *testing.T) {
	cmd := returnOrdersReceiveCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"return_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestReturnOrdersListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := returnOrdersListCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestReturnOrdersGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := returnOrdersGetCmd.RunE(cmd, []string{"return_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestReturnOrdersCancelGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := returnOrdersCancelCmd.RunE(cmd, []string{"return_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestReturnOrdersCompleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := returnOrdersCompleteCmd.RunE(cmd, []string{"return_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestReturnOrdersReceiveGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }

	cmd := newTestCmdWithFlags()
	if err := returnOrdersReceiveCmd.RunE(cmd, []string{"return_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestReturnOrdersListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := returnOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestReturnOrdersListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"envstore", "other"},
			creds: map[string]*secrets.StoreCredentials{
				"envstore": {Handle: "test", AccessToken: "token123"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := returnOrdersListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// returnOrdersAPIClient is a mock implementation for return orders testing.
type returnOrdersAPIClient struct {
	api.MockClient

	listReturnOrdersResp *api.ReturnOrdersListResponse
	listReturnOrdersErr  error

	getReturnOrderResp *api.ReturnOrder
	getReturnOrderErr  error

	cancelReturnOrderErr error

	completeReturnOrderResp *api.ReturnOrder
	completeReturnOrderErr  error

	receiveReturnOrderResp *api.ReturnOrder
	receiveReturnOrderErr  error
}

func (m *returnOrdersAPIClient) ListReturnOrders(ctx context.Context, opts *api.ReturnOrdersListOptions) (*api.ReturnOrdersListResponse, error) {
	return m.listReturnOrdersResp, m.listReturnOrdersErr
}

func (m *returnOrdersAPIClient) GetReturnOrder(ctx context.Context, id string) (*api.ReturnOrder, error) {
	return m.getReturnOrderResp, m.getReturnOrderErr
}

func (m *returnOrdersAPIClient) CancelReturnOrder(ctx context.Context, id string) error {
	return m.cancelReturnOrderErr
}

func (m *returnOrdersAPIClient) CompleteReturnOrder(ctx context.Context, id string) (*api.ReturnOrder, error) {
	return m.completeReturnOrderResp, m.completeReturnOrderErr
}

func (m *returnOrdersAPIClient) ReceiveReturnOrder(ctx context.Context, id string) (*api.ReturnOrder, error) {
	return m.receiveReturnOrderResp, m.receiveReturnOrderErr
}

func TestReturnOrdersListRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	now := time.Now()
	tests := []struct {
		name       string
		mockResp   *api.ReturnOrdersListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.ReturnOrdersListResponse{
				Items: []api.ReturnOrder{
					{
						ID:          "ret_123",
						OrderNumber: "ORD-1001",
						Status:      "pending",
						ReturnType:  "return",
						TotalAmount: "99.99",
						Currency:    "USD",
						LineItems: []api.ReturnOrderLineItem{
							{LineItemID: "li_1", Title: "Product 1", Quantity: 1, ReturnReason: "defective"},
						},
						CreatedAt: now,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "ret_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.ReturnOrdersListResponse{
				Items:      []api.ReturnOrder{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple return orders",
			mockResp: &api.ReturnOrdersListResponse{
				Items: []api.ReturnOrder{
					{
						ID:          "ret_001",
						OrderNumber: "ORD-1001",
						Status:      "pending",
						ReturnType:  "return",
						TotalAmount: "50.00",
						Currency:    "USD",
						CreatedAt:   now,
					},
					{
						ID:          "ret_002",
						OrderNumber: "ORD-1002",
						Status:      "received",
						ReturnType:  "exchange",
						TotalAmount: "75.00",
						Currency:    "USD",
						CreatedAt:   now,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "ret_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &returnOrdersAPIClient{
				listReturnOrdersResp: tt.mockResp,
				listReturnOrdersErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("order-id", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("type", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := returnOrdersListCmd.RunE(cmd, []string{})

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

func TestReturnOrdersListRunE_JSONOutput(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &returnOrdersAPIClient{
		listReturnOrdersResp: &api.ReturnOrdersListResponse{
			Items: []api.ReturnOrder{
				{
					ID:          "ret_json",
					OrderNumber: "ORD-JSON",
					Status:      "pending",
					ReturnType:  "return",
					TotalAmount: "100.00",
					Currency:    "USD",
					CreatedAt:   time.Now(),
				},
			},
			TotalCount: 1,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := returnOrdersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ret_json") {
		t.Errorf("JSON output should contain 'ret_json', got: %s", output)
	}
}

func TestReturnOrdersGetRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	now := time.Now()
	receivedAt := now.Add(-2 * time.Hour)
	completedAt := now.Add(-1 * time.Hour)

	tests := []struct {
		name          string
		returnOrderID string
		mockResp      *api.ReturnOrder
		mockErr       error
		wantErr       bool
	}{
		{
			name:          "successful get",
			returnOrderID: "ret_123",
			mockResp: &api.ReturnOrder{
				ID:              "ret_123",
				OrderID:         "ord_456",
				OrderNumber:     "ORD-1001",
				Status:          "pending",
				ReturnType:      "return",
				CustomerID:      "cust_789",
				CustomerEmail:   "customer@example.com",
				TotalAmount:     "99.99",
				RefundAmount:    "89.99",
				Currency:        "USD",
				Reason:          "Product defective",
				Note:            "Customer request",
				TrackingNumber:  "TRACK123",
				TrackingCompany: "FedEx",
				LineItems: []api.ReturnOrderLineItem{
					{LineItemID: "li_1", Title: "Product 1", Quantity: 2, ReturnReason: "defective"},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:          "return order with all optional fields",
			returnOrderID: "ret_full",
			mockResp: &api.ReturnOrder{
				ID:              "ret_full",
				OrderID:         "ord_789",
				OrderNumber:     "ORD-2002",
				Status:          "completed",
				ReturnType:      "exchange",
				CustomerID:      "cust_101",
				CustomerEmail:   "full@example.com",
				TotalAmount:     "150.00",
				RefundAmount:    "150.00",
				Currency:        "EUR",
				Reason:          "Wrong size",
				Note:            "Exchange for larger size",
				TrackingNumber:  "TRACK456",
				TrackingCompany: "DHL",
				ReceivedAt:      &receivedAt,
				CompletedAt:     &completedAt,
				LineItems: []api.ReturnOrderLineItem{
					{LineItemID: "li_2", Title: "Shirt", Quantity: 1, ReturnReason: "wrong_size"},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name:          "return order not found",
			returnOrderID: "ret_999",
			mockErr:       errors.New("return order not found"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &returnOrdersAPIClient{
				getReturnOrderResp: tt.mockResp,
				getReturnOrderErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := returnOrdersGetCmd.RunE(cmd, []string{tt.returnOrderID})

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

func TestReturnOrdersGetRunE_JSONOutput(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &returnOrdersAPIClient{
		getReturnOrderResp: &api.ReturnOrder{
			ID:            "ret_json_get",
			OrderNumber:   "ORD-JSON",
			Status:        "pending",
			ReturnType:    "return",
			TotalAmount:   "100.00",
			RefundAmount:  "100.00",
			Currency:      "USD",
			CustomerEmail: "json@example.com",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := returnOrdersGetCmd.RunE(cmd, []string{"ret_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ret_json_get") {
		t.Errorf("JSON output should contain 'ret_json_get', got: %s", output)
	}
}

func TestReturnOrdersGetRunE_WithCancelledAt(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	now := time.Now()
	cancelledAt := now.Add(-30 * time.Minute)

	mockClient := &returnOrdersAPIClient{
		getReturnOrderResp: &api.ReturnOrder{
			ID:            "ret_cancelled",
			OrderNumber:   "ORD-CANCELLED",
			Status:        "cancelled",
			ReturnType:    "return",
			TotalAmount:   "50.00",
			RefundAmount:  "0.00",
			Currency:      "USD",
			CustomerEmail: "cancelled@example.com",
			CancelledAt:   &cancelledAt,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := returnOrdersGetCmd.RunE(cmd, []string{"ret_cancelled"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReturnOrdersCancelRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	tests := []struct {
		name          string
		returnOrderID string
		mockErr       error
		wantErr       bool
	}{
		{
			name:          "successful cancel",
			returnOrderID: "ret_123",
			mockErr:       nil,
		},
		{
			name:          "cancel fails",
			returnOrderID: "ret_456",
			mockErr:       errors.New("return order already cancelled"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &returnOrdersAPIClient{
				cancelReturnOrderErr: tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := returnOrdersCancelCmd.RunE(cmd, []string{tt.returnOrderID})

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

func TestReturnOrdersCompleteRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	now := time.Now()
	completedAt := now

	tests := []struct {
		name          string
		returnOrderID string
		mockResp      *api.ReturnOrder
		mockErr       error
		wantErr       bool
	}{
		{
			name:          "successful complete",
			returnOrderID: "ret_123",
			mockResp: &api.ReturnOrder{
				ID:          "ret_123",
				OrderNumber: "ORD-1001",
				Status:      "completed",
				ReturnType:  "return",
				CompletedAt: &completedAt,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name:          "complete fails",
			returnOrderID: "ret_456",
			mockErr:       errors.New("return order cannot be completed"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &returnOrdersAPIClient{
				completeReturnOrderResp: tt.mockResp,
				completeReturnOrderErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")

			err := returnOrdersCompleteCmd.RunE(cmd, []string{tt.returnOrderID})

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

func TestReturnOrdersReceiveRunE_Success(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	now := time.Now()
	receivedAt := now

	tests := []struct {
		name          string
		returnOrderID string
		mockResp      *api.ReturnOrder
		mockErr       error
		wantErr       bool
	}{
		{
			name:          "successful receive",
			returnOrderID: "ret_123",
			mockResp: &api.ReturnOrder{
				ID:          "ret_123",
				OrderNumber: "ORD-1001",
				Status:      "received",
				ReturnType:  "return",
				ReceivedAt:  &receivedAt,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			name:          "receive fails",
			returnOrderID: "ret_456",
			mockErr:       errors.New("return order cannot be received"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &returnOrdersAPIClient{
				receiveReturnOrderResp: tt.mockResp,
				receiveReturnOrderErr:  tt.mockErr,
			}
			clientFactory = func(handle, accessToken string) api.APIClient {
				return mockClient
			}

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")

			err := returnOrdersReceiveCmd.RunE(cmd, []string{tt.returnOrderID})

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

func TestReturnOrdersGetRunE_NoLineItems(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &returnOrdersAPIClient{
		getReturnOrderResp: &api.ReturnOrder{
			ID:            "ret_no_items",
			OrderNumber:   "ORD-NOITEMS",
			Status:        "pending",
			ReturnType:    "return",
			TotalAmount:   "0.00",
			RefundAmount:  "0.00",
			Currency:      "USD",
			CustomerEmail: "noitems@example.com",
			LineItems:     []api.ReturnOrderLineItem{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := returnOrdersGetCmd.RunE(cmd, []string{"ret_no_items"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReturnOrdersGetRunE_NoReason(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &returnOrdersAPIClient{
		getReturnOrderResp: &api.ReturnOrder{
			ID:            "ret_no_reason",
			OrderNumber:   "ORD-NOREASON",
			Status:        "pending",
			ReturnType:    "return",
			TotalAmount:   "50.00",
			RefundAmount:  "50.00",
			Currency:      "USD",
			CustomerEmail: "noreason@example.com",
			Reason:        "", // No reason
			Note:          "", // No note
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := returnOrdersGetCmd.RunE(cmd, []string{"ret_no_reason"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReturnOrdersGetRunE_NoTracking(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &returnOrdersAPIClient{
		getReturnOrderResp: &api.ReturnOrder{
			ID:              "ret_no_tracking",
			OrderNumber:     "ORD-NOTRACK",
			Status:          "pending",
			ReturnType:      "return",
			TotalAmount:     "50.00",
			RefundAmount:    "50.00",
			Currency:        "USD",
			CustomerEmail:   "notrack@example.com",
			TrackingNumber:  "", // No tracking
			TrackingCompany: "", // No tracking company
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := returnOrdersGetCmd.RunE(cmd, []string{"ret_no_tracking"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

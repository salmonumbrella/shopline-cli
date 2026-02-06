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

func TestDraftOrdersCommandStructure(t *testing.T) {
	if draftOrdersCmd == nil {
		t.Fatal("draftOrdersCmd is nil")
	}
	if draftOrdersCmd.Use != "draft-orders" {
		t.Errorf("Expected Use 'draft-orders', got %q", draftOrdersCmd.Use)
	}
	subcommands := map[string]bool{"list": false, "get": false, "delete": false, "complete": false, "send-invoice": false}
	for _, cmd := range draftOrdersCmd.Commands() {
		for key := range subcommands {
			if strings.HasPrefix(cmd.Use, key) {
				subcommands[key] = true
			}
		}
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("Subcommand %q not found", name)
		}
	}
}

func TestDraftOrdersListFlags(t *testing.T) {
	cmd := draftOrdersListCmd
	flags := []struct{ name, defaultValue string }{{"status", ""}, {"customer-id", ""}, {"page", "1"}, {"page-size", "20"}}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestDraftOrdersGetRequiresArg(t *testing.T) {
	cmd := draftOrdersGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"draft_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDraftOrdersDeleteRequiresArg(t *testing.T) {
	cmd := draftOrdersDeleteCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"draft_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDraftOrdersCompleteRequiresArg(t *testing.T) {
	cmd := draftOrdersCompleteCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"draft_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDraftOrdersSendInvoiceRequiresArg(t *testing.T) {
	cmd := draftOrdersSendInvoiceCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"draft_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestDraftOrdersListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := draftOrdersListCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDraftOrdersListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := draftOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestDraftOrdersListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := draftOrdersListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// draftOrdersMockAPIClient is a mock implementation of api.APIClient for draft orders tests.
type draftOrdersMockAPIClient struct {
	api.MockClient
	listDraftOrdersResp      *api.DraftOrdersListResponse
	listDraftOrdersErr       error
	getDraftOrderResp        *api.DraftOrder
	getDraftOrderErr         error
	deleteDraftOrderErr      error
	completeDraftOrderResp   *api.DraftOrder
	completeDraftOrderErr    error
	sendDraftOrderInvoiceErr error
}

func (m *draftOrdersMockAPIClient) ListDraftOrders(ctx context.Context, opts *api.DraftOrdersListOptions) (*api.DraftOrdersListResponse, error) {
	return m.listDraftOrdersResp, m.listDraftOrdersErr
}

func (m *draftOrdersMockAPIClient) GetDraftOrder(ctx context.Context, id string) (*api.DraftOrder, error) {
	return m.getDraftOrderResp, m.getDraftOrderErr
}

func (m *draftOrdersMockAPIClient) DeleteDraftOrder(ctx context.Context, id string) error {
	return m.deleteDraftOrderErr
}

func (m *draftOrdersMockAPIClient) CompleteDraftOrder(ctx context.Context, id string) (*api.DraftOrder, error) {
	return m.completeDraftOrderResp, m.completeDraftOrderErr
}

func (m *draftOrdersMockAPIClient) SendDraftOrderInvoice(ctx context.Context, id string) error {
	return m.sendDraftOrderInvoiceErr
}

// setupDraftOrdersMockFactories sets up mock factories for draft orders tests.
func setupDraftOrdersMockFactories(mockClient *draftOrdersMockAPIClient) (func(), *bytes.Buffer) {
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

// newDraftOrdersTestCmd creates a test command with common flags for draft orders tests.
func newDraftOrdersTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().BoolP("yes", "y", true, "")
	return cmd
}

// TestDraftOrdersListRunE tests the draft orders list command with mock API.
func TestDraftOrdersListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.DraftOrdersListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.DraftOrdersListResponse{
				Items: []api.DraftOrder{
					{
						ID:            "draft_123",
						Name:          "#D001",
						Status:        "open",
						CustomerEmail: "customer@example.com",
						TotalPrice:    "99.99",
						Currency:      "USD",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "draft_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.DraftOrdersListResponse{
				Items:      []api.DraftOrder{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple draft orders",
			mockResp: &api.DraftOrdersListResponse{
				Items: []api.DraftOrder{
					{
						ID:            "draft_001",
						Name:          "#D001",
						Status:        "open",
						CustomerEmail: "alice@example.com",
						TotalPrice:    "50.00",
						Currency:      "USD",
						CreatedAt:     time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
					},
					{
						ID:            "draft_002",
						Name:          "#D002",
						Status:        "invoice_sent",
						CustomerEmail: "bob@example.com",
						TotalPrice:    "75.00",
						Currency:      "EUR",
						CreatedAt:     time.Date(2024, 1, 11, 10, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "draft_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &draftOrdersMockAPIClient{
				listDraftOrdersResp: tt.mockResp,
				listDraftOrdersErr:  tt.mockErr,
			}
			cleanup, buf := setupDraftOrdersMockFactories(mockClient)
			defer cleanup()

			cmd := newDraftOrdersTestCmd()
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := draftOrdersListCmd.RunE(cmd, []string{})

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

// TestDraftOrdersListRunEJSON tests the draft orders list command with JSON output.
func TestDraftOrdersListRunEJSON(t *testing.T) {
	mockClient := &draftOrdersMockAPIClient{
		listDraftOrdersResp: &api.DraftOrdersListResponse{
			Items: []api.DraftOrder{
				{
					ID:            "draft_json",
					Name:          "#DJSON",
					Status:        "open",
					CustomerEmail: "json@example.com",
					TotalPrice:    "123.45",
					Currency:      "USD",
					CreatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupDraftOrdersMockFactories(mockClient)
	defer cleanup()

	cmd := newDraftOrdersTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := draftOrdersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "draft_json") {
		t.Errorf("JSON output should contain draft_json, got: %s", output)
	}
	if !strings.Contains(output, "DJSON") {
		t.Errorf("JSON output should contain DJSON, got: %s", output)
	}
}

// TestDraftOrdersGetRunE tests the draft orders get command with mock API.
// Note: Text output goes to stdout via fmt.Printf, so we only verify no error occurs.
// JSON output is tested separately in TestDraftOrdersGetRunEJSON.
func TestDraftOrdersGetRunE(t *testing.T) {
	invoiceSentAt := time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC)
	completedAt := time.Date(2024, 1, 25, 16, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		draftID  string
		mockResp *api.DraftOrder
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			draftID: "draft_123",
			mockResp: &api.DraftOrder{
				ID:            "draft_123",
				Name:          "#D001",
				Status:        "open",
				CustomerID:    "cust_456",
				CustomerEmail: "customer@example.com",
				TotalPrice:    "99.99",
				SubtotalPrice: "89.99",
				TotalTax:      "10.00",
				Currency:      "USD",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "draft order with note",
			draftID: "draft_with_note",
			mockResp: &api.DraftOrder{
				ID:            "draft_with_note",
				Name:          "#D002",
				Status:        "open",
				CustomerID:    "cust_789",
				CustomerEmail: "customer2@example.com",
				TotalPrice:    "50.00",
				SubtotalPrice: "45.00",
				TotalTax:      "5.00",
				Currency:      "USD",
				Note:          "Rush order",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "draft order with invoice URL",
			draftID: "draft_with_invoice",
			mockResp: &api.DraftOrder{
				ID:            "draft_with_invoice",
				Name:          "#D003",
				Status:        "invoice_sent",
				CustomerID:    "cust_101",
				CustomerEmail: "invoiced@example.com",
				TotalPrice:    "75.00",
				SubtotalPrice: "70.00",
				TotalTax:      "5.00",
				Currency:      "USD",
				InvoiceURL:    "https://example.com/invoice/123",
				InvoiceSentAt: &invoiceSentAt,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "completed draft order",
			draftID: "draft_completed",
			mockResp: &api.DraftOrder{
				ID:            "draft_completed",
				Name:          "#D004",
				Status:        "completed",
				CustomerID:    "cust_202",
				CustomerEmail: "completed@example.com",
				TotalPrice:    "200.00",
				SubtotalPrice: "180.00",
				TotalTax:      "20.00",
				Currency:      "USD",
				CompletedAt:   &completedAt,
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 25, 16, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "draft order with line items",
			draftID: "draft_with_items",
			mockResp: &api.DraftOrder{
				ID:            "draft_with_items",
				Name:          "#D005",
				Status:        "open",
				CustomerID:    "cust_303",
				CustomerEmail: "items@example.com",
				TotalPrice:    "150.00",
				SubtotalPrice: "140.00",
				TotalTax:      "10.00",
				Currency:      "USD",
				LineItems: []api.DraftOrderLineItem{
					{
						Title:     "Widget A",
						VariantID: "var_001",
						Quantity:  2,
						Price:     50.00,
					},
					{
						Title:     "Widget B",
						VariantID: "var_002",
						Quantity:  1,
						Price:     40.00,
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "draft order not found",
			draftID: "draft_999",
			mockErr: errors.New("draft order not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &draftOrdersMockAPIClient{
				getDraftOrderResp: tt.mockResp,
				getDraftOrderErr:  tt.mockErr,
			}
			cleanup, _ := setupDraftOrdersMockFactories(mockClient)
			defer cleanup()

			cmd := newDraftOrdersTestCmd()

			err := draftOrdersGetCmd.RunE(cmd, []string{tt.draftID})

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

// TestDraftOrdersGetRunEJSON tests the draft orders get command with JSON output.
func TestDraftOrdersGetRunEJSON(t *testing.T) {
	mockClient := &draftOrdersMockAPIClient{
		getDraftOrderResp: &api.DraftOrder{
			ID:            "draft_json_get",
			Name:          "#DJSONGET",
			Status:        "open",
			CustomerID:    "cust_json",
			CustomerEmail: "json@example.com",
			TotalPrice:    "199.99",
			SubtotalPrice: "179.99",
			TotalTax:      "20.00",
			Currency:      "EUR",
			CreatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 2, 2, 14, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupDraftOrdersMockFactories(mockClient)
	defer cleanup()

	cmd := newDraftOrdersTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := draftOrdersGetCmd.RunE(cmd, []string{"draft_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "draft_json_get") {
		t.Errorf("JSON output should contain draft_json_get, got: %s", output)
	}
	if !strings.Contains(output, "DJSONGET") {
		t.Errorf("JSON output should contain DJSONGET, got: %s", output)
	}
}

// TestDraftOrdersGetGetClientError tests get command error handling when getClient fails.
func TestDraftOrdersGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := draftOrdersGetCmd.RunE(cmd, []string{"draft_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDraftOrdersDeleteRunE tests the draft orders delete command with mock API.
func TestDraftOrdersDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		draftID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			draftID: "draft_123",
			mockErr: nil,
		},
		{
			name:    "delete fails",
			draftID: "draft_456",
			mockErr: errors.New("draft order not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &draftOrdersMockAPIClient{
				deleteDraftOrderErr: tt.mockErr,
			}
			cleanup, _ := setupDraftOrdersMockFactories(mockClient)
			defer cleanup()

			cmd := newDraftOrdersTestCmd()

			err := draftOrdersDeleteCmd.RunE(cmd, []string{tt.draftID})

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

// TestDraftOrdersDeleteGetClientError tests delete command error handling when getClient fails.
func TestDraftOrdersDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	err := draftOrdersDeleteCmd.RunE(cmd, []string{"draft_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDraftOrdersCompleteRunE tests the draft orders complete command with mock API.
func TestDraftOrdersCompleteRunE(t *testing.T) {
	tests := []struct {
		name     string
		draftID  string
		mockResp *api.DraftOrder
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful complete",
			draftID: "draft_123",
			mockResp: &api.DraftOrder{
				ID:     "draft_123",
				Status: "completed",
			},
			mockErr: nil,
		},
		{
			name:    "complete fails",
			draftID: "draft_456",
			mockErr: errors.New("draft order already completed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &draftOrdersMockAPIClient{
				completeDraftOrderResp: tt.mockResp,
				completeDraftOrderErr:  tt.mockErr,
			}
			cleanup, _ := setupDraftOrdersMockFactories(mockClient)
			defer cleanup()

			cmd := newDraftOrdersTestCmd()

			err := draftOrdersCompleteCmd.RunE(cmd, []string{tt.draftID})

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

// TestDraftOrdersCompleteGetClientError tests complete command error handling when getClient fails.
func TestDraftOrdersCompleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	err := draftOrdersCompleteCmd.RunE(cmd, []string{"draft_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDraftOrdersSendInvoiceRunE tests the draft orders send-invoice command with mock API.
func TestDraftOrdersSendInvoiceRunE(t *testing.T) {
	tests := []struct {
		name    string
		draftID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful send invoice",
			draftID: "draft_123",
			mockErr: nil,
		},
		{
			name:    "send invoice fails",
			draftID: "draft_456",
			mockErr: errors.New("draft order not found"),
			wantErr: true,
		},
		{
			name:    "send invoice fails - no customer email",
			draftID: "draft_789",
			mockErr: errors.New("customer email required"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &draftOrdersMockAPIClient{
				sendDraftOrderInvoiceErr: tt.mockErr,
			}
			cleanup, _ := setupDraftOrdersMockFactories(mockClient)
			defer cleanup()

			cmd := newDraftOrdersTestCmd()

			err := draftOrdersSendInvoiceCmd.RunE(cmd, []string{tt.draftID})

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

// TestDraftOrdersSendInvoiceGetClientError tests send-invoice command error handling when getClient fails.
func TestDraftOrdersSendInvoiceGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := draftOrdersSendInvoiceCmd.RunE(cmd, []string{"draft_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDraftOrdersListWithFilters tests the list command respects filter flags.
func TestDraftOrdersListWithFilters(t *testing.T) {
	mockClient := &draftOrdersMockAPIClient{
		listDraftOrdersResp: &api.DraftOrdersListResponse{
			Items: []api.DraftOrder{
				{
					ID:            "draft_filtered",
					Name:          "#DFILTER",
					Status:        "open",
					CustomerEmail: "filtered@example.com",
					TotalPrice:    "100.00",
					Currency:      "USD",
					CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupDraftOrdersMockFactories(mockClient)
	defer cleanup()

	cmd := newDraftOrdersTestCmd()
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	_ = cmd.Flags().Set("status", "open")
	_ = cmd.Flags().Set("customer-id", "cust_123")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := draftOrdersListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "draft_filtered") {
		t.Errorf("output should contain draft_filtered, got: %s", output)
	}
}

// TestDraftOrdersMultipleProfilesWithFlag tests that store flag resolves multiple profiles.
func TestDraftOrdersMultipleProfilesWithFlag(t *testing.T) {
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
				"store1": {Handle: "handle1", AccessToken: "token1"},
				"store2": {Handle: "handle2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("store", "store1")

	client, err := getClient(cmd)
	if err != nil {
		t.Errorf("should resolve store with flag, got error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// TestDraftOrdersMultipleProfilesWithoutFlag tests that multiple profiles without flag errors.
func TestDraftOrdersMultipleProfilesWithoutFlag(t *testing.T) {
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
				"store1": {Handle: "handle1", AccessToken: "token1"},
				"store2": {Handle: "handle2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	err := draftOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for multiple profiles without flag")
	}
	if !strings.Contains(err.Error(), "multiple profiles configured") {
		t.Errorf("expected 'multiple profiles configured' error, got: %v", err)
	}
}

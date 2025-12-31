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

// TestRefundsCommandSetup verifies refunds command initialization
func TestRefundsCommandSetup(t *testing.T) {
	if refundsCmd.Use != "refunds" {
		t.Errorf("expected Use 'refunds', got %q", refundsCmd.Use)
	}
	if refundsCmd.Short != "Manage order refunds" {
		t.Errorf("expected Short 'Manage order refunds', got %q", refundsCmd.Short)
	}
}

// TestRefundsSubcommands verifies all subcommands are registered
func TestRefundsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":  "List refunds",
		"get":   "Get refund details",
		"order": "List refunds for an order",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range refundsCmd.Commands() {
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

// TestRefundsListFlags verifies list command flags exist with correct defaults
func TestRefundsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := refundsListCmd.Flags().Lookup(f.name)
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

// TestRefundsGetCmd verifies get command setup
func TestRefundsGetCmd(t *testing.T) {
	if refundsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", refundsGetCmd.Use)
	}
}

// TestRefundsOrderCmd verifies order command setup
func TestRefundsOrderCmd(t *testing.T) {
	if refundsOrderCmd.Use != "order <order-id>" {
		t.Errorf("expected Use 'order <order-id>', got %q", refundsOrderCmd.Use)
	}
}

// TestRefundsGetRequiresArg verifies get command requires exactly one argument
func TestRefundsGetRequiresArg(t *testing.T) {
	cmd := refundsGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"ref_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
	if cmd.Args(cmd, []string{"ref_123", "extra"}) == nil {
		t.Error("Expected error with 2 args")
	}
}

// TestRefundsOrderRequiresArg verifies order command requires exactly one argument
func TestRefundsOrderRequiresArg(t *testing.T) {
	cmd := refundsOrderCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"order_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
	if cmd.Args(cmd, []string{"order_123", "extra"}) == nil {
		t.Error("Expected error with 2 args")
	}
}

// TestRefundsListRunE_GetClientFails verifies error handling when getClient fails
func TestRefundsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := refundsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestRefundsGetRunE_GetClientFails verifies error handling when getClient fails
func TestRefundsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := refundsGetCmd.RunE(cmd, []string{"ref_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestRefundsOrderRunE_GetClientFails verifies error handling when getClient fails
func TestRefundsOrderRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := refundsOrderCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestRefundsListRunE_NoProfiles verifies error handling when no profiles exist
func TestRefundsListRunE_NoProfiles(t *testing.T) {
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

	err := refundsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// refundsTestClient is a mock implementation for refunds testing.
type refundsTestClient struct {
	api.MockClient

	listRefundsResp      *api.RefundsListResponse
	listRefundsErr       error
	getRefundResp        *api.Refund
	getRefundErr         error
	listOrderRefundsResp *api.RefundsListResponse
	listOrderRefundsErr  error
}

func (m *refundsTestClient) ListRefunds(ctx context.Context, opts *api.RefundsListOptions) (*api.RefundsListResponse, error) {
	return m.listRefundsResp, m.listRefundsErr
}

func (m *refundsTestClient) GetRefund(ctx context.Context, id string) (*api.Refund, error) {
	return m.getRefundResp, m.getRefundErr
}

func (m *refundsTestClient) ListOrderRefunds(ctx context.Context, orderID string) (*api.RefundsListResponse, error) {
	return m.listOrderRefundsResp, m.listOrderRefundsErr
}

// setupRefundsTest sets up mocks for refunds testing and returns cleanup function.
func setupRefundsTest(mockClient *refundsTestClient) func() {
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

// TestRefundsListRunE tests the refunds list command execution with mock API.
func TestRefundsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.RefundsListResponse
		mockErr    error
		outputFmt  string
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with text output",
			mockResp: &api.RefundsListResponse{
				Items: []api.Refund{
					{
						ID:        "ref_123",
						OrderID:   "ord_456",
						Status:    "completed",
						Amount:    "50.00",
						Currency:  "USD",
						Note:      "Customer request",
						Restock:   true,
						LineItems: []api.RefundLineItem{{LineItemID: "li_1", Quantity: 1}},
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "text",
			wantOutput: "ref_123",
		},
		{
			name: "successful list with JSON output",
			mockResp: &api.RefundsListResponse{
				Items: []api.Refund{
					{
						ID:        "ref_789",
						OrderID:   "ord_012",
						Status:    "pending",
						Amount:    "100.00",
						Currency:  "EUR",
						CreatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "json",
			wantOutput: "ref_789",
		},
		{
			name: "list with long note truncated",
			mockResp: &api.RefundsListResponse{
				Items: []api.Refund{
					{
						ID:        "ref_long",
						OrderID:   "ord_long",
						Status:    "completed",
						Amount:    "25.00",
						Currency:  "USD",
						Note:      "This is a very long note that exceeds twenty characters",
						CreatedAt: time.Date(2024, 3, 10, 8, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "text",
			wantOutput: "ref_long",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.RefundsListResponse{
				Items:      []api.Refund{},
				TotalCount: 0,
			},
			outputFmt: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &refundsTestClient{
				listRefundsResp: tt.mockResp,
				listRefundsErr:  tt.mockErr,
			}

			cleanup := setupRefundsTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFmt, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := refundsListCmd.RunE(cmd, []string{})

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

// TestRefundsGetRunE tests the refunds get command execution with mock API.
func TestRefundsGetRunE(t *testing.T) {
	tests := []struct {
		name      string
		refundID  string
		mockResp  *api.Refund
		mockErr   error
		outputFmt string
		wantErr   bool
	}{
		{
			name:     "successful get with text output",
			refundID: "ref_123",
			mockResp: &api.Refund{
				ID:          "ref_123",
				OrderID:     "ord_456",
				Status:      "completed",
				Amount:      "50.00",
				Currency:    "USD",
				Note:        "Customer request",
				Restock:     true,
				ProcessedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				LineItems: []api.RefundLineItem{
					{LineItemID: "li_1", Quantity: 2, Subtotal: 25.00},
					{LineItemID: "li_2", Quantity: 1, Subtotal: 25.00},
				},
			},
			outputFmt: "text",
		},
		{
			name:     "successful get with JSON output",
			refundID: "ref_789",
			mockResp: &api.Refund{
				ID:        "ref_789",
				OrderID:   "ord_012",
				Status:    "pending",
				Amount:    "100.00",
				Currency:  "EUR",
				CreatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
			},
			outputFmt: "json",
		},
		{
			name:     "get without note",
			refundID: "ref_no_note",
			mockResp: &api.Refund{
				ID:        "ref_no_note",
				OrderID:   "ord_no_note",
				Status:    "completed",
				Amount:    "10.00",
				Currency:  "USD",
				Restock:   false,
				CreatedAt: time.Date(2024, 3, 5, 9, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:     "get without processed at",
			refundID: "ref_not_processed",
			mockResp: &api.Refund{
				ID:        "ref_not_processed",
				OrderID:   "ord_pending",
				Status:    "pending",
				Amount:    "75.00",
				Currency:  "GBP",
				Restock:   true,
				CreatedAt: time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:     "get with empty line items",
			refundID: "ref_no_items",
			mockResp: &api.Refund{
				ID:        "ref_no_items",
				OrderID:   "ord_no_items",
				Status:    "completed",
				Amount:    "5.00",
				Currency:  "CAD",
				LineItems: []api.RefundLineItem{},
				CreatedAt: time.Date(2024, 5, 10, 16, 0, 0, 0, time.UTC),
			},
			outputFmt: "text",
		},
		{
			name:     "refund not found",
			refundID: "ref_999",
			mockErr:  errors.New("refund not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &refundsTestClient{
				getRefundResp: tt.mockResp,
				getRefundErr:  tt.mockErr,
			}

			cleanup := setupRefundsTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFmt, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := refundsGetCmd.RunE(cmd, []string{tt.refundID})

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

// TestRefundsOrderRunE tests the refunds order command execution with mock API.
func TestRefundsOrderRunE(t *testing.T) {
	tests := []struct {
		name       string
		orderID    string
		mockResp   *api.RefundsListResponse
		mockErr    error
		outputFmt  string
		wantErr    bool
		wantOutput string
	}{
		{
			name:    "successful order refunds with text output",
			orderID: "ord_456",
			mockResp: &api.RefundsListResponse{
				Items: []api.Refund{
					{
						ID:        "ref_123",
						OrderID:   "ord_456",
						Status:    "completed",
						Amount:    "50.00",
						Currency:  "USD",
						Note:      "Partial refund",
						LineItems: []api.RefundLineItem{{LineItemID: "li_1", Quantity: 1}},
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:        "ref_124",
						OrderID:   "ord_456",
						Status:    "pending",
						Amount:    "25.00",
						Currency:  "USD",
						CreatedAt: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			outputFmt:  "text",
			wantOutput: "ref_123",
		},
		{
			name:    "successful order refunds with JSON output",
			orderID: "ord_789",
			mockResp: &api.RefundsListResponse{
				Items: []api.Refund{
					{
						ID:        "ref_456",
						OrderID:   "ord_789",
						Status:    "completed",
						Amount:    "100.00",
						Currency:  "EUR",
						CreatedAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "json",
			wantOutput: "ref_456",
		},
		{
			name:    "order refunds with long note truncated",
			orderID: "ord_long",
			mockResp: &api.RefundsListResponse{
				Items: []api.Refund{
					{
						ID:        "ref_long_note",
						OrderID:   "ord_long",
						Status:    "completed",
						Amount:    "30.00",
						Currency:  "USD",
						Note:      "This is an exceptionally long note that will be truncated",
						CreatedAt: time.Date(2024, 3, 10, 8, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			outputFmt:  "text",
			wantOutput: "ref_long_note",
		},
		{
			name:    "API error",
			orderID: "ord_error",
			mockErr: errors.New("failed to fetch order refunds"),
			wantErr: true,
		},
		{
			name:    "empty refunds list",
			orderID: "ord_empty",
			mockResp: &api.RefundsListResponse{
				Items:      []api.Refund{},
				TotalCount: 0,
			},
			outputFmt: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &refundsTestClient{
				listOrderRefundsResp: tt.mockResp,
				listOrderRefundsErr:  tt.mockErr,
			}

			cleanup := setupRefundsTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFmt, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := refundsOrderCmd.RunE(cmd, []string{tt.orderID})

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

// TestRefundsListWithEnvVar verifies store selection via environment variable
func TestRefundsListWithEnvVar(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := refundsListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// TestRefundsGetNoProfiles verifies error handling when no profiles exist for get command
func TestRefundsGetNoProfiles(t *testing.T) {
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
	err := refundsGetCmd.RunE(cmd, []string{"ref_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestRefundsOrderNoProfiles verifies error handling when no profiles exist for order command
func TestRefundsOrderNoProfiles(t *testing.T) {
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
	err := refundsOrderCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestRefundsListAPIErrorMessage verifies error wrapping in list command
func TestRefundsListAPIErrorMessage(t *testing.T) {
	mockClient := &refundsTestClient{
		listRefundsErr: errors.New("connection refused"),
	}

	cleanup := setupRefundsTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := refundsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to list refunds") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestRefundsGetAPIErrorMessage verifies error wrapping in get command
func TestRefundsGetAPIErrorMessage(t *testing.T) {
	mockClient := &refundsTestClient{
		getRefundErr: errors.New("not found"),
	}

	cleanup := setupRefundsTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := refundsGetCmd.RunE(cmd, []string{"ref_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to get refund") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestRefundsOrderAPIErrorMessage verifies error wrapping in order command
func TestRefundsOrderAPIErrorMessage(t *testing.T) {
	mockClient := &refundsTestClient{
		listOrderRefundsErr: errors.New("order not found"),
	}

	cleanup := setupRefundsTest(mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := refundsOrderCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "failed to list order refunds") {
		t.Errorf("Expected wrapped error message, got: %v", err)
	}
}

// TestRefundsListMultipleItems verifies handling of multiple refunds in list
func TestRefundsListMultipleItems(t *testing.T) {
	mockClient := &refundsTestClient{
		listRefundsResp: &api.RefundsListResponse{
			Items: []api.Refund{
				{
					ID:        "ref_1",
					OrderID:   "ord_1",
					Status:    "completed",
					Amount:    "10.00",
					Currency:  "USD",
					LineItems: []api.RefundLineItem{},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "ref_2",
					OrderID:   "ord_2",
					Status:    "pending",
					Amount:    "20.00",
					Currency:  "EUR",
					LineItems: []api.RefundLineItem{{LineItemID: "li_1", Quantity: 1}},
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "ref_3",
					OrderID:   "ord_3",
					Status:    "completed",
					Amount:    "30.00",
					Currency:  "GBP",
					LineItems: []api.RefundLineItem{
						{LineItemID: "li_1", Quantity: 1},
						{LineItemID: "li_2", Quantity: 2},
					},
					CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 3,
		},
	}

	cleanup := setupRefundsTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := refundsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ref_1") || !strings.Contains(output, "ref_2") || !strings.Contains(output, "ref_3") {
		t.Errorf("output should contain all three refund IDs, got: %q", output)
	}
}

// TestRefundsGetWithAllFields verifies complete refund output with all fields populated
func TestRefundsGetWithAllFields(t *testing.T) {
	mockClient := &refundsTestClient{
		getRefundResp: &api.Refund{
			ID:          "ref_full",
			OrderID:     "ord_full",
			Status:      "completed",
			Amount:      "150.00",
			Currency:    "USD",
			Note:        "Full refund with all fields",
			Restock:     true,
			ProcessedAt: time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
			CreatedAt:   time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			LineItems: []api.RefundLineItem{
				{LineItemID: "li_1", Quantity: 3, Subtotal: 75.00, RestockType: "return"},
				{LineItemID: "li_2", Quantity: 2, Subtotal: 50.00, RestockType: "cancel"},
				{LineItemID: "li_3", Quantity: 1, Subtotal: 25.00, RestockType: "no_restock"},
			},
		},
	}

	cleanup := setupRefundsTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := refundsGetCmd.RunE(cmd, []string{"ref_full"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

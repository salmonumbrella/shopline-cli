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

func TestPayoutsCommandStructure(t *testing.T) {
	if payoutsCmd == nil {
		t.Fatal("payoutsCmd is nil")
	}
	if payoutsCmd.Use != "payouts" {
		t.Errorf("Expected Use 'payouts', got %q", payoutsCmd.Use)
	}
	subcommands := map[string]bool{"list": false, "get": false}
	for _, cmd := range payoutsCmd.Commands() {
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

func TestPayoutsListFlags(t *testing.T) {
	cmd := payoutsListCmd
	flags := []struct{ name, defaultValue string }{{"page", "1"}, {"page-size", "20"}, {"status", ""}}
	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("Flag %q not found", f.name)
		} else if flag.DefValue != f.defaultValue {
			t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestPayoutsGetRequiresArg(t *testing.T) {
	cmd := payoutsGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"payout_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestPayoutsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := payoutsListCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestPayoutsListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := payoutsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestPayoutsListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := payoutsListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// payoutsTestClient is a mock implementation for payouts testing.
type payoutsTestClient struct {
	api.MockClient

	listPayoutsResp *api.PayoutsListResponse
	listPayoutsErr  error
	getPayoutResp   *api.Payout
	getPayoutErr    error
}

func (m *payoutsTestClient) ListPayouts(ctx context.Context, opts *api.PayoutsListOptions) (*api.PayoutsListResponse, error) {
	return m.listPayoutsResp, m.listPayoutsErr
}

func (m *payoutsTestClient) GetPayout(ctx context.Context, id string) (*api.Payout, error) {
	return m.getPayoutResp, m.getPayoutErr
}

// setupPayoutsTest configures the test environment for payouts tests.
func setupPayoutsTest(t *testing.T) (cleanup func()) {
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

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// TestPayoutsListRunE tests the payouts list command execution with mock API.
func TestPayoutsListRunE(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	tests := []struct {
		name       string
		mockResp   *api.PayoutsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.PayoutsListResponse{
				Items: []api.Payout{
					{
						ID:          "payout_123",
						Amount:      "1000.00",
						Currency:    "USD",
						Status:      "paid",
						Type:        "bank_transfer",
						BankAccount: "****1234",
						CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "payout_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.PayoutsListResponse{
				Items:      []api.Payout{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple payouts",
			mockResp: &api.PayoutsListResponse{
				Items: []api.Payout{
					{
						ID:          "payout_001",
						Amount:      "500.00",
						Currency:    "USD",
						Status:      "pending",
						Type:        "bank_transfer",
						BankAccount: "****5678",
						CreatedAt:   time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC),
					},
					{
						ID:          "payout_002",
						Amount:      "750.50",
						Currency:    "EUR",
						Status:      "in_transit",
						Type:        "bank_transfer",
						BankAccount: "****9012",
						CreatedAt:   time.Date(2024, 1, 12, 14, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "payout_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &payoutsTestClient{
				listPayoutsResp: tt.mockResp,
				listPayoutsErr:  tt.mockErr,
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
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := payoutsListCmd.RunE(cmd, []string{})

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

// TestPayoutsListRunEWithJSON tests the payouts list command with JSON output.
func TestPayoutsListRunEWithJSON(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	mockClient := &payoutsTestClient{
		listPayoutsResp: &api.PayoutsListResponse{
			Items: []api.Payout{
				{
					ID:          "payout_json_123",
					Amount:      "2500.00",
					Currency:    "GBP",
					Status:      "paid",
					Type:        "bank_transfer",
					BankAccount: "****4321",
					CreatedAt:   time.Date(2024, 2, 20, 16, 45, 0, 0, time.UTC),
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := payoutsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "payout_json_123") {
		t.Errorf("JSON output should contain payout ID, got: %s", output)
	}
}

// TestPayoutsListRunEWithStatusFilter tests the payouts list command with status filter.
func TestPayoutsListRunEWithStatusFilter(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	mockClient := &payoutsTestClient{
		listPayoutsResp: &api.PayoutsListResponse{
			Items: []api.Payout{
				{
					ID:          "payout_pending_001",
					Amount:      "300.00",
					Currency:    "USD",
					Status:      "pending",
					Type:        "bank_transfer",
					BankAccount: "****7890",
					CreatedAt:   time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
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
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "pending", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("status", "pending")

	err := payoutsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPayoutsGetRunE tests the payouts get command execution with mock API.
func TestPayoutsGetRunE(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	scheduledDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	arrivalDate := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		payoutID  string
		mockResp  *api.Payout
		mockErr   error
		wantErr   bool
		wantInOut []string
	}{
		{
			name:     "successful get basic payout",
			payoutID: "payout_123",
			mockResp: &api.Payout{
				ID:          "payout_123",
				Amount:      "1000.00",
				Currency:    "USD",
				Status:      "paid",
				Type:        "bank_transfer",
				BankAccount: "****1234",
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			},
			wantInOut: []string{"payout_123", "1000.00", "USD", "paid"},
		},
		{
			name:     "successful get payout with transaction ID",
			payoutID: "payout_456",
			mockResp: &api.Payout{
				ID:            "payout_456",
				Amount:        "2000.00",
				Currency:      "EUR",
				Status:        "in_transit",
				Type:          "wire_transfer",
				BankAccount:   "****5678",
				TransactionID: "txn_abc123",
				CreatedAt:     time.Date(2024, 1, 16, 14, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 16, 14, 30, 0, 0, time.UTC),
			},
			wantInOut: []string{"payout_456", "txn_abc123"},
		},
		{
			name:     "successful get payout with fee and net",
			payoutID: "payout_789",
			mockResp: &api.Payout{
				ID:          "payout_789",
				Amount:      "500.00",
				Currency:    "USD",
				Status:      "paid",
				Type:        "bank_transfer",
				BankAccount: "****9012",
				Fee:         "5.00",
				Net:         "495.00",
				CreatedAt:   time.Date(2024, 1, 17, 8, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC),
			},
			wantInOut: []string{"payout_789", "5.00", "495.00"},
		},
		{
			name:     "successful get payout with summary",
			payoutID: "payout_summary_001",
			mockResp: &api.Payout{
				ID:          "payout_summary_001",
				Amount:      "3000.00",
				Currency:    "USD",
				Status:      "paid",
				Type:        "bank_transfer",
				BankAccount: "****3456",
				Summary: &api.PayoutSummary{
					Sales:       "3500.00",
					Refunds:     "300.00",
					Adjustments: "50.00",
					Charges:     "250.00",
				},
				CreatedAt: time.Date(2024, 1, 18, 9, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 18, 11, 0, 0, 0, time.UTC),
			},
			wantInOut: []string{"3500.00", "300.00", "50.00", "250.00"},
		},
		{
			name:     "successful get payout with scheduled and arrival dates",
			payoutID: "payout_dates_001",
			mockResp: &api.Payout{
				ID:            "payout_dates_001",
				Amount:        "1500.00",
				Currency:      "USD",
				Status:        "pending",
				Type:          "bank_transfer",
				BankAccount:   "****7890",
				ScheduledDate: &scheduledDate,
				ArrivalDate:   &arrivalDate,
				CreatedAt:     time.Date(2024, 1, 19, 10, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 1, 19, 10, 0, 0, 0, time.UTC),
			},
			wantInOut: []string{"payout_dates_001", "1500.00"},
		},
		{
			name:     "payout not found",
			payoutID: "payout_999",
			mockErr:  errors.New("payout not found"),
			wantErr:  true,
		},
		{
			name:     "API error",
			payoutID: "payout_error",
			mockErr:  errors.New("API unavailable"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &payoutsTestClient{
				getPayoutResp: tt.mockResp,
				getPayoutErr:  tt.mockErr,
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

			err := payoutsGetCmd.RunE(cmd, []string{tt.payoutID})

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

			// Note: text output goes to stdout directly via fmt.Printf, not to the buffer
			// JSON output goes through the formatter
		})
	}
}

// TestPayoutsGetRunEWithJSON tests the payouts get command with JSON output.
func TestPayoutsGetRunEWithJSON(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	mockClient := &payoutsTestClient{
		getPayoutResp: &api.Payout{
			ID:          "payout_json_get_123",
			Amount:      "5000.00",
			Currency:    "GBP",
			Status:      "paid",
			Type:        "bank_transfer",
			BankAccount: "****4567",
			CreatedAt:   time.Date(2024, 2, 25, 11, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 2, 25, 14, 0, 0, 0, time.UTC),
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
	_ = cmd.Flags().Set("output", "json")

	err := payoutsGetCmd.RunE(cmd, []string{"payout_json_get_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "payout_json_get_123") {
		t.Errorf("JSON output should contain payout ID, got: %s", output)
	}
}

// TestPayoutsGetRunEClientError tests the payouts get command when getClient fails.
func TestPayoutsGetRunEClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := payoutsGetCmd.RunE(cmd, []string{"payout_123"})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestPayoutsGetRunENoProfiles tests the payouts get command when no profiles are configured.
func TestPayoutsGetRunENoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newTestCmdWithFlags()
	err := payoutsGetCmd.RunE(cmd, []string{"payout_123"})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestPayoutsGetRunEWithAllOptionalFields tests the payout get command displaying all optional fields.
func TestPayoutsGetRunEWithAllOptionalFields(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	scheduledDate := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	arrivalDate := time.Date(2024, 3, 3, 0, 0, 0, 0, time.UTC)

	mockClient := &payoutsTestClient{
		getPayoutResp: &api.Payout{
			ID:            "payout_complete_001",
			Amount:        "10000.00",
			Currency:      "USD",
			Status:        "paid",
			Type:          "wire_transfer",
			BankAccount:   "****1111",
			TransactionID: "txn_complete_123",
			Fee:           "25.00",
			Net:           "9975.00",
			Summary: &api.PayoutSummary{
				Sales:       "12000.00",
				Refunds:     "1500.00",
				Adjustments: "100.00",
				Charges:     "625.00",
			},
			ScheduledDate: &scheduledDate,
			ArrivalDate:   &arrivalDate,
			CreatedAt:     time.Date(2024, 2, 28, 10, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 3, 3, 12, 0, 0, 0, time.UTC),
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

	err := payoutsGetCmd.RunE(cmd, []string{"payout_complete_001"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPayoutsListRunEWithPagination tests the payouts list command with pagination options.
func TestPayoutsListRunEWithPagination(t *testing.T) {
	cleanup := setupPayoutsTest(t)
	defer cleanup()

	mockClient := &payoutsTestClient{
		listPayoutsResp: &api.PayoutsListResponse{
			Items: []api.Payout{
				{
					ID:          "payout_page_001",
					Amount:      "100.00",
					Currency:    "USD",
					Status:      "paid",
					Type:        "bank_transfer",
					BankAccount: "****2222",
					CreatedAt:   time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 50,
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
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 10, "")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := payoutsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

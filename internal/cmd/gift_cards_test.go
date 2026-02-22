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

func TestGiftCardsCmd(t *testing.T) {
	if giftCardsCmd.Use != "gift-cards" {
		t.Errorf("Expected Use, got %q", giftCardsCmd.Use)
	}
}

func TestGiftCardsListCmd(t *testing.T) {
	if giftCardsListCmd.Use != "list" {
		t.Errorf("Expected Use, got %q", giftCardsListCmd.Use)
	}
}

func TestGiftCardsGetCmd(t *testing.T) {
	if giftCardsGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use, got %q", giftCardsGetCmd.Use)
	}
}

func TestGiftCardsCreateCmd(t *testing.T) {
	if giftCardsCreateCmd.Use != "create" {
		t.Errorf("Expected Use, got %q", giftCardsCreateCmd.Use)
	}
}

func TestGiftCardsDeleteCmd(t *testing.T) {
	if giftCardsDeleteCmd.Use != "delete <id>" {
		t.Errorf("Expected Use, got %q", giftCardsDeleteCmd.Use)
	}
}

func TestGiftCardsListFlags(t *testing.T) {
	flags := []string{"status", "customer-id", "page", "page-size"}
	for _, flag := range flags {
		if giftCardsListCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestGiftCardsCreateFlags(t *testing.T) {
	flags := []string{"initial-value", "currency", "code", "customer-id", "note"}
	for _, flag := range flags {
		if giftCardsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

func TestGiftCardsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := giftCardsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftCardsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := giftCardsGetCmd.RunE(cmd, []string{"gc_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftCardsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("initial-value", "100.00", "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("note", "", "")
	err := giftCardsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftCardsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := giftCardsDeleteCmd.RunE(cmd, []string{"gc_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftCardsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := giftCardsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestGiftCardsCreateRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("initial-value", "100.00", "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("note", "", "")
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := giftCardsCreateCmd.RunE(cmd, []string{})
	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run, got: %s", output)
	}
}

func TestGiftCardsDeleteRunE_DryRun(t *testing.T) {
	var buf bytes.Buffer
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd := newTestCmdWithFlags()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}
	err := giftCardsDeleteCmd.RunE(cmd, []string{"gc_123"})
	_ = w.Close()
	os.Stdout = origStdout
	_, _ = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run, got: %s", output)
	}
}

// giftCardsTestClient is a mock implementation for gift cards testing.
type giftCardsTestClient struct {
	api.MockClient

	listGiftCardsResp  *api.GiftCardsListResponse
	listGiftCardsErr   error
	getGiftCardResp    *api.GiftCard
	getGiftCardErr     error
	createGiftCardResp *api.GiftCard
	createGiftCardErr  error
	deleteGiftCardErr  error
}

func (m *giftCardsTestClient) ListGiftCards(ctx context.Context, opts *api.GiftCardsListOptions) (*api.GiftCardsListResponse, error) {
	return m.listGiftCardsResp, m.listGiftCardsErr
}

func (m *giftCardsTestClient) GetGiftCard(ctx context.Context, id string) (*api.GiftCard, error) {
	return m.getGiftCardResp, m.getGiftCardErr
}

func (m *giftCardsTestClient) CreateGiftCard(ctx context.Context, req *api.GiftCardCreateRequest) (*api.GiftCard, error) {
	return m.createGiftCardResp, m.createGiftCardErr
}

func (m *giftCardsTestClient) DeleteGiftCard(ctx context.Context, id string) error {
	return m.deleteGiftCardErr
}

// setupGiftCardsMockFactories sets up mock factories for gift cards tests.
func setupGiftCardsMockFactories(mockClient *giftCardsTestClient) (func(), *bytes.Buffer) {
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

// newGiftCardsTestCmd creates a test command with common flags for gift cards tests.
func newGiftCardsTestCmd() *cobra.Command {
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

// TestGiftCardsListRunE tests the gift cards list command with mock API.
func TestGiftCardsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.GiftCardsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.GiftCardsListResponse{
				Items: []api.GiftCard{
					{
						ID:           "gc_123",
						MaskedCode:   "****1234",
						InitialValue: "100.00",
						Balance:      "75.00",
						Currency:     "USD",
						Status:       api.GiftCardStatusEnabled,
						CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "gc_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.GiftCardsListResponse{
				Items:      []api.GiftCard{},
				TotalCount: 0,
			},
		},
		{
			name: "gift card with expires at",
			mockResp: &api.GiftCardsListResponse{
				Items: []api.GiftCard{
					{
						ID:           "gc_456",
						MaskedCode:   "****5678",
						InitialValue: "50.00",
						Balance:      "50.00",
						Currency:     "USD",
						Status:       api.GiftCardStatusEnabled,
						ExpiresAt:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
						CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "2025-12-31",
		},
		{
			name: "gift card with zero expires at",
			mockResp: &api.GiftCardsListResponse{
				Items: []api.GiftCard{
					{
						ID:           "gc_789",
						MaskedCode:   "****9012",
						InitialValue: "200.00",
						Balance:      "100.00",
						Currency:     "EUR",
						Status:       api.GiftCardStatusDisabled,
						ExpiresAt:    time.Time{}, // Zero time
						CreatedAt:    time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "gc_789",
		},
		{
			name: "multiple gift cards",
			mockResp: &api.GiftCardsListResponse{
				Items: []api.GiftCard{
					{
						ID:           "gc_001",
						MaskedCode:   "****0001",
						InitialValue: "25.00",
						Balance:      "25.00",
						Currency:     "USD",
						Status:       api.GiftCardStatusEnabled,
						CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:           "gc_002",
						MaskedCode:   "****0002",
						InitialValue: "50.00",
						Balance:      "0.00",
						Currency:     "USD",
						Status:       api.GiftCardStatusDisabled,
						CreatedAt:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "gc_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftCardsTestClient{
				listGiftCardsResp: tt.mockResp,
				listGiftCardsErr:  tt.mockErr,
			}
			cleanup, buf := setupGiftCardsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftCardsTestCmd()
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := giftCardsListCmd.RunE(cmd, []string{})

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

// TestGiftCardsListRunEWithJSON tests JSON output format for list command.
func TestGiftCardsListRunEWithJSON(t *testing.T) {
	mockClient := &giftCardsTestClient{
		listGiftCardsResp: &api.GiftCardsListResponse{
			Items: []api.GiftCard{
				{
					ID:           "gc_json",
					MaskedCode:   "****JSON",
					InitialValue: "100.00",
					Balance:      "50.00",
					Currency:     "USD",
					Status:       api.GiftCardStatusEnabled,
					CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupGiftCardsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftCardsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := giftCardsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gc_json") {
		t.Errorf("JSON output should contain gift card ID, got: %s", output)
	}
}

// TestGiftCardsGetRunE tests the gift cards get command with mock API.
func TestGiftCardsGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		giftCardID string
		mockResp   *api.GiftCard
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful get",
			giftCardID: "gc_123",
			mockResp: &api.GiftCard{
				ID:           "gc_123",
				Code:         "GIFT-1234-5678-9012",
				MaskedCode:   "****9012",
				InitialValue: "100.00",
				Balance:      "75.00",
				Currency:     "USD",
				Status:       api.GiftCardStatusEnabled,
				CustomerID:   "cust_456",
				Note:         "Birthday gift",
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "gift card not found",
			giftCardID: "gc_999",
			mockErr:    errors.New("gift card not found"),
			wantErr:    true,
		},
		{
			name:       "gift card with expires at",
			giftCardID: "gc_456",
			mockResp: &api.GiftCard{
				ID:           "gc_456",
				Code:         "GIFT-ABCD-EFGH-IJKL",
				MaskedCode:   "****IJKL",
				InitialValue: "50.00",
				Balance:      "50.00",
				Currency:     "USD",
				Status:       api.GiftCardStatusEnabled,
				ExpiresAt:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:       "gift card with disabled at",
			giftCardID: "gc_789",
			mockResp: &api.GiftCard{
				ID:           "gc_789",
				Code:         "GIFT-MNOP-QRST-UVWX",
				MaskedCode:   "****UVWX",
				InitialValue: "200.00",
				Balance:      "0.00",
				Currency:     "EUR",
				Status:       api.GiftCardStatusDisabled,
				DisabledAt:   time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "gift card with both expires at and disabled at",
			giftCardID: "gc_full",
			mockResp: &api.GiftCard{
				ID:           "gc_full",
				Code:         "GIFT-FULL-TEST-DATA",
				MaskedCode:   "****DATA",
				InitialValue: "500.00",
				Balance:      "250.00",
				Currency:     "GBP",
				Status:       api.GiftCardStatusDisabled,
				CustomerID:   "cust_premium",
				Note:         "VIP customer gift",
				ExpiresAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DisabledAt:   time.Date(2024, 12, 15, 9, 30, 0, 0, time.UTC),
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 12, 15, 9, 30, 0, 0, time.UTC),
			},
		},
		{
			name:       "gift card with zero times",
			giftCardID: "gc_zero",
			mockResp: &api.GiftCard{
				ID:           "gc_zero",
				Code:         "GIFT-ZERO-TIME-TEST",
				MaskedCode:   "****TEST",
				InitialValue: "25.00",
				Balance:      "25.00",
				Currency:     "USD",
				Status:       api.GiftCardStatusEnabled,
				ExpiresAt:    time.Time{}, // Zero time - should show "N/A"
				DisabledAt:   time.Time{}, // Zero time - should show "N/A"
				CreatedAt:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftCardsTestClient{
				getGiftCardResp: tt.mockResp,
				getGiftCardErr:  tt.mockErr,
			}
			cleanup, _ := setupGiftCardsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftCardsTestCmd()

			err := giftCardsGetCmd.RunE(cmd, []string{tt.giftCardID})

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

// TestGiftCardsGetRunEWithJSON tests JSON output format for get command.
func TestGiftCardsGetRunEWithJSON(t *testing.T) {
	mockClient := &giftCardsTestClient{
		getGiftCardResp: &api.GiftCard{
			ID:           "gc_json",
			Code:         "GIFT-JSON-TEST-CODE",
			MaskedCode:   "****CODE",
			InitialValue: "100.00",
			Balance:      "100.00",
			Currency:     "USD",
			Status:       api.GiftCardStatusEnabled,
			CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupGiftCardsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftCardsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := giftCardsGetCmd.RunE(cmd, []string{"gc_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gc_json") {
		t.Errorf("JSON output should contain gift card ID, got: %s", output)
	}
}

// TestGiftCardsCreateRunE tests the gift cards create command with mock API.
func TestGiftCardsCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.GiftCard
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.GiftCard{
				ID:           "gc_new",
				Code:         "GIFT-NEW1-TEST-CODE",
				MaskedCode:   "****CODE",
				InitialValue: "100.00",
				Balance:      "100.00",
				Currency:     "USD",
				Status:       api.GiftCardStatusEnabled,
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("failed to create gift card"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftCardsTestClient{
				createGiftCardResp: tt.mockResp,
				createGiftCardErr:  tt.mockErr,
			}
			cleanup, _ := setupGiftCardsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftCardsTestCmd()
			cmd.Flags().String("initial-value", "100.00", "")
			cmd.Flags().String("currency", "USD", "")
			cmd.Flags().String("code", "", "")
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("note", "", "")

			err := giftCardsCreateCmd.RunE(cmd, []string{})

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

// TestGiftCardsCreateRunEWithJSON tests JSON output format for create command.
func TestGiftCardsCreateRunEWithJSON(t *testing.T) {
	mockClient := &giftCardsTestClient{
		createGiftCardResp: &api.GiftCard{
			ID:           "gc_created_json",
			Code:         "GIFT-CRJN-TEST-CODE",
			MaskedCode:   "****CODE",
			InitialValue: "200.00",
			Balance:      "200.00",
			Currency:     "EUR",
			Status:       api.GiftCardStatusEnabled,
			CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupGiftCardsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftCardsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("initial-value", "200.00", "")
	cmd.Flags().String("currency", "EUR", "")
	cmd.Flags().String("code", "GIFT-CRJN-TEST-CODE", "")
	cmd.Flags().String("customer-id", "cust_test", "")
	cmd.Flags().String("note", "Test note", "")

	err := giftCardsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gc_created_json") {
		t.Errorf("JSON output should contain gift card ID, got: %s", output)
	}
}

// TestGiftCardsDeleteRunE tests the gift cards delete command with mock API.
func TestGiftCardsDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		giftCardID string
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful delete",
			giftCardID: "gc_123",
			mockErr:    nil,
		},
		{
			name:       "delete fails",
			giftCardID: "gc_456",
			mockErr:    errors.New("gift card not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &giftCardsTestClient{
				deleteGiftCardErr: tt.mockErr,
			}
			cleanup, _ := setupGiftCardsMockFactories(mockClient)
			defer cleanup()

			cmd := newGiftCardsTestCmd()

			err := giftCardsDeleteCmd.RunE(cmd, []string{tt.giftCardID})

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

// TestGiftCardsGetArgs verifies get command requires exactly 1 argument.
func TestGiftCardsGetArgs(t *testing.T) {
	err := giftCardsGetCmd.Args(giftCardsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = giftCardsGetCmd.Args(giftCardsGetCmd, []string{"gc_id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestGiftCardsDeleteArgs verifies delete command requires exactly 1 argument.
func TestGiftCardsDeleteArgs(t *testing.T) {
	err := giftCardsDeleteCmd.Args(giftCardsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = giftCardsDeleteCmd.Args(giftCardsDeleteCmd, []string{"gc_id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestGiftCardsCommandStructure verifies the command hierarchy.
func TestGiftCardsCommandStructure(t *testing.T) {
	subcommands := giftCardsCmd.Commands()
	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		baseUse := strings.Split(cmd.Use, " ")[0]
		if _, ok := expectedCmds[baseUse]; ok {
			expectedCmds[baseUse] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %s not found", name)
		}
	}
}

// TestGiftCardsListWithFilters tests list command with various filter options.
func TestGiftCardsListWithFilters(t *testing.T) {
	mockClient := &giftCardsTestClient{
		listGiftCardsResp: &api.GiftCardsListResponse{
			Items: []api.GiftCard{
				{
					ID:           "gc_filtered",
					MaskedCode:   "****FILT",
					InitialValue: "100.00",
					Balance:      "100.00",
					Currency:     "USD",
					Status:       api.GiftCardStatusEnabled,
					CustomerID:   "cust_specific",
					CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupGiftCardsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftCardsTestCmd()
	cmd.Flags().String("status", "enabled", "")
	cmd.Flags().String("customer-id", "cust_specific", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 10, "")
	_ = cmd.Flags().Set("status", "enabled")
	_ = cmd.Flags().Set("customer-id", "cust_specific")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := giftCardsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gc_filtered") {
		t.Errorf("output should contain filtered gift card ID, got: %s", output)
	}
}

// TestGiftCardsCreateWithAllFlags tests create command with all optional flags.
func TestGiftCardsCreateWithAllFlags(t *testing.T) {
	mockClient := &giftCardsTestClient{
		createGiftCardResp: &api.GiftCard{
			ID:           "gc_all_flags",
			Code:         "CUSTOM-CODE-1234",
			MaskedCode:   "****1234",
			InitialValue: "500.00",
			Balance:      "500.00",
			Currency:     "GBP",
			Status:       api.GiftCardStatusEnabled,
			CustomerID:   "cust_vip",
			Note:         "VIP customer reward",
			CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupGiftCardsMockFactories(mockClient)
	defer cleanup()

	cmd := newGiftCardsTestCmd()
	cmd.Flags().String("initial-value", "", "")
	cmd.Flags().String("currency", "", "")
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("note", "", "")
	_ = cmd.Flags().Set("initial-value", "500.00")
	_ = cmd.Flags().Set("currency", "GBP")
	_ = cmd.Flags().Set("code", "CUSTOM-CODE-1234")
	_ = cmd.Flags().Set("customer-id", "cust_vip")
	_ = cmd.Flags().Set("note", "VIP customer reward")

	err := giftCardsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

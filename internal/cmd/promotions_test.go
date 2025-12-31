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

// mockPromotionsClient is a mock implementation of api.APIClient for promotions testing.
type mockPromotionsClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values for specific methods
	listPromotionsResp *api.PromotionsListResponse
	listPromotionsErr  error

	getPromotionResp *api.Promotion
	getPromotionErr  error

	activatePromotionResp *api.Promotion
	activatePromotionErr  error

	deactivatePromotionResp *api.Promotion
	deactivatePromotionErr  error

	deletePromotionErr error
}

func (m *mockPromotionsClient) ListPromotions(ctx context.Context, opts *api.PromotionsListOptions) (*api.PromotionsListResponse, error) {
	return m.listPromotionsResp, m.listPromotionsErr
}

func (m *mockPromotionsClient) GetPromotion(ctx context.Context, id string) (*api.Promotion, error) {
	return m.getPromotionResp, m.getPromotionErr
}

func (m *mockPromotionsClient) ActivatePromotion(ctx context.Context, id string) (*api.Promotion, error) {
	return m.activatePromotionResp, m.activatePromotionErr
}

func (m *mockPromotionsClient) DeactivatePromotion(ctx context.Context, id string) (*api.Promotion, error) {
	return m.deactivatePromotionResp, m.deactivatePromotionErr
}

func (m *mockPromotionsClient) DeletePromotion(ctx context.Context, id string) error {
	return m.deletePromotionErr
}

// setupPromotionsTest sets up the test environment with mocked factories.
func setupPromotionsTest(t *testing.T, mockClient *mockPromotionsClient) (restore func()) {
	t.Helper()
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	// Setup mock credential store
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	// Setup mock client factory
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// newPromotionsTestCmd creates a new command with the required flags for promotions testing.
func newPromotionsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	return cmd
}

// TestPromotionsCommandSetup verifies promotions command initialization
func TestPromotionsCommandSetup(t *testing.T) {
	if promotionsCmd.Use != "promotions" {
		t.Errorf("expected Use 'promotions', got %q", promotionsCmd.Use)
	}
	if promotionsCmd.Short != "Manage promotions" {
		t.Errorf("expected Short 'Manage promotions', got %q", promotionsCmd.Short)
	}
}

// TestPromotionsSubcommands verifies all subcommands are registered
func TestPromotionsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":       "List promotions",
		"get":        "Get promotion details",
		"activate":   "Activate a promotion",
		"deactivate": "Deactivate a promotion",
		"delete":     "Delete a promotion",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range promotionsCmd.Commands() {
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

// TestPromotionsListFlags verifies list command flags exist with correct defaults
func TestPromotionsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status", ""},
		{"type", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := promotionsListCmd.Flags().Lookup(f.name)
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

// TestPromotionsGetCmd verifies get command setup
func TestPromotionsGetCmd(t *testing.T) {
	if promotionsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", promotionsGetCmd.Use)
	}
}

// TestPromotionsActivateCmd verifies activate command setup
func TestPromotionsActivateCmd(t *testing.T) {
	if promotionsActivateCmd.Use != "activate <id>" {
		t.Errorf("expected Use 'activate <id>', got %q", promotionsActivateCmd.Use)
	}
}

// TestPromotionsDeactivateCmd verifies deactivate command setup
func TestPromotionsDeactivateCmd(t *testing.T) {
	if promotionsDeactivateCmd.Use != "deactivate <id>" {
		t.Errorf("expected Use 'deactivate <id>', got %q", promotionsDeactivateCmd.Use)
	}
}

// TestPromotionsDeleteCmd verifies delete command setup
func TestPromotionsDeleteCmd(t *testing.T) {
	if promotionsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", promotionsDeleteCmd.Use)
	}
}

// TestPromotionsListRunE_GetClientFails verifies error handling when getClient fails
func TestPromotionsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := promotionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPromotionsGetRunE_GetClientFails verifies error handling when getClient fails
func TestPromotionsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := promotionsGetCmd.RunE(cmd, []string{"promo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPromotionsActivateRunE_GetClientFails verifies error handling when getClient fails
func TestPromotionsActivateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := promotionsActivateCmd.RunE(cmd, []string{"promo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPromotionsDeactivateRunE_GetClientFails verifies error handling when getClient fails
func TestPromotionsDeactivateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := promotionsDeactivateCmd.RunE(cmd, []string{"promo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPromotionsDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestPromotionsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := promotionsDeleteCmd.RunE(cmd, []string{"promo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPromotionsListRunE_NoProfiles verifies error handling when no profiles exist
func TestPromotionsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := promotionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestPromotionsListRunE tests the promotions list command execution with mock API.
func TestPromotionsListRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name         string
		mockResp     *api.PromotionsListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   []string
	}{
		{
			name: "successful list with percentage discount",
			mockResp: &api.PromotionsListResponse{
				Items: []api.Promotion{
					{
						ID:            "promo_123",
						Title:         "Summer Sale",
						Type:          "discount",
						Status:        "active",
						DiscountType:  "percentage",
						DiscountValue: 20,
						UsageCount:    50,
						UsageLimit:    100,
						StartsAt:      startsAt,
						EndsAt:        endsAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: []string{"promo_123", "Summer Sale", "discount", "active", "20%", "50/100"},
		},
		{
			name: "successful list with fixed discount",
			mockResp: &api.PromotionsListResponse{
				Items: []api.Promotion{
					{
						ID:            "promo_456",
						Title:         "Flat Discount",
						Type:          "coupon",
						Status:        "scheduled",
						DiscountType:  "fixed",
						DiscountValue: 10,
						UsageCount:    0,
						UsageLimit:    0,
						StartsAt:      startsAt,
						EndsAt:        time.Time{},
					},
				},
				TotalCount: 1,
			},
			wantOutput: []string{"promo_456", "Flat Discount", "coupon", "scheduled", "10"},
		},
		{
			name: "successful list JSON output",
			mockResp: &api.PromotionsListResponse{
				Items: []api.Promotion{
					{
						ID:            "promo_789",
						Title:         "JSON Test",
						Type:          "bundle",
						Status:        "active",
						DiscountType:  "percentage",
						DiscountValue: 15,
						StartsAt:      startsAt,
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   []string{"promo_789", "JSON Test"},
		},
		{
			name: "empty list",
			mockResp: &api.PromotionsListResponse{
				Items:      []api.Promotion{},
				TotalCount: 0,
			},
			wantOutput: []string{"ID", "TITLE", "TYPE", "STATUS"},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				listPromotionsResp: tt.mockResp,
				listPromotionsErr:  tt.mockErr,
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newPromotionsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := promotionsListCmd.RunE(cmd, []string{})

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
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output %q should contain %q", output, want)
				}
			}
		})
	}
}

// TestPromotionsListRunE_WithFilters tests the promotions list command with filters.
func TestPromotionsListRunE_WithFilters(t *testing.T) {
	mockClient := &mockPromotionsClient{
		listPromotionsResp: &api.PromotionsListResponse{
			Items:      []api.Promotion{},
			TotalCount: 0,
		},
	}

	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newPromotionsTestCmd()
	_ = cmd.Flags().Set("status", "active")
	_ = cmd.Flags().Set("type", "discount")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "50")

	err := promotionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPromotionsGetRunE tests the promotions get command execution with mock API.
func TestPromotionsGetRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		promotionID  string
		mockResp     *api.Promotion
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   []string
	}{
		{
			name:        "successful get with usage limit",
			promotionID: "promo_123",
			mockResp: &api.Promotion{
				ID:            "promo_123",
				Title:         "Summer Sale",
				Description:   "Get 20% off all items",
				Type:          "discount",
				Status:        "active",
				DiscountType:  "percentage",
				DiscountValue: 20,
				MinPurchase:   50.00,
				UsageLimit:    100,
				UsageCount:    50,
				StartsAt:      startsAt,
				EndsAt:        endsAt,
				CreatedAt:     createdAt,
			},
		},
		{
			name:        "successful get without usage limit",
			promotionID: "promo_456",
			mockResp: &api.Promotion{
				ID:            "promo_456",
				Title:         "Flash Sale",
				Description:   "Limited time offer",
				Type:          "flash",
				Status:        "active",
				DiscountType:  "fixed",
				DiscountValue: 10,
				MinPurchase:   0,
				UsageLimit:    0,
				UsageCount:    25,
				StartsAt:      startsAt,
				EndsAt:        time.Time{},
				CreatedAt:     createdAt,
			},
		},
		{
			name:         "successful get JSON output",
			promotionID:  "promo_789",
			outputFormat: "json",
			mockResp: &api.Promotion{
				ID:            "promo_789",
				Title:         "JSON Promo",
				Type:          "bundle",
				Status:        "scheduled",
				DiscountType:  "percentage",
				DiscountValue: 15,
				StartsAt:      startsAt,
				CreatedAt:     createdAt,
			},
			wantOutput: []string{"promo_789", "JSON Promo"},
		},
		{
			name:        "promotion not found",
			promotionID: "promo_999",
			mockErr:     errors.New("promotion not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				getPromotionResp: tt.mockResp,
				getPromotionErr:  tt.mockErr,
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newPromotionsTestCmd()
			if tt.outputFormat != "" {
				_ = cmd.Flags().Set("output", tt.outputFormat)
			}

			err := promotionsGetCmd.RunE(cmd, []string{tt.promotionID})

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

			if tt.outputFormat == "json" {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output %q should contain %q", output, want)
					}
				}
			}
		})
	}
}

// TestPromotionsActivateRunE tests the promotions activate command execution.
func TestPromotionsActivateRunE(t *testing.T) {
	tests := []struct {
		name        string
		promotionID string
		mockResp    *api.Promotion
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful activate",
			promotionID: "promo_123",
			mockResp: &api.Promotion{
				ID:     "promo_123",
				Status: "active",
			},
		},
		{
			name:        "activation fails",
			promotionID: "promo_456",
			mockErr:     errors.New("promotion already active"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				activatePromotionResp: tt.mockResp,
				activatePromotionErr:  tt.mockErr,
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			cmd := newPromotionsTestCmd()

			err := promotionsActivateCmd.RunE(cmd, []string{tt.promotionID})

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

// TestPromotionsDeactivateRunE tests the promotions deactivate command execution.
func TestPromotionsDeactivateRunE(t *testing.T) {
	tests := []struct {
		name        string
		promotionID string
		mockResp    *api.Promotion
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful deactivate",
			promotionID: "promo_123",
			mockResp: &api.Promotion{
				ID:     "promo_123",
				Status: "inactive",
			},
		},
		{
			name:        "deactivation fails",
			promotionID: "promo_456",
			mockErr:     errors.New("promotion already inactive"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				deactivatePromotionResp: tt.mockResp,
				deactivatePromotionErr:  tt.mockErr,
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			cmd := newPromotionsTestCmd()

			err := promotionsDeactivateCmd.RunE(cmd, []string{tt.promotionID})

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

// TestPromotionsDeleteRunE tests the promotions delete command execution.
func TestPromotionsDeleteRunE(t *testing.T) {
	tests := []struct {
		name        string
		promotionID string
		yesFlag     bool
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful delete with yes flag",
			promotionID: "promo_123",
			yesFlag:     true,
			mockErr:     nil,
		},
		{
			name:        "delete fails",
			promotionID: "promo_456",
			yesFlag:     true,
			mockErr:     errors.New("promotion not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				deletePromotionErr: tt.mockErr,
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			cmd := newPromotionsTestCmd()
			if tt.yesFlag {
				_ = cmd.Flags().Set("yes", "true")
			}

			err := promotionsDeleteCmd.RunE(cmd, []string{tt.promotionID})

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

// TestPromotionsDeleteRunE_CancelConfirmation tests delete command cancellation.
func TestPromotionsDeleteRunE_CancelConfirmation(t *testing.T) {
	mockClient := &mockPromotionsClient{}

	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("n\n")
		_ = w.Close()
	}()

	cmd := newPromotionsTestCmd()

	err := promotionsDeleteCmd.RunE(cmd, []string{"promo_123"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPromotionsDeleteRunE_ConfirmYes tests delete command confirmation with "y".
func TestPromotionsDeleteRunE_ConfirmYes(t *testing.T) {
	mockClient := &mockPromotionsClient{}

	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("y\n")
		_ = w.Close()
	}()

	cmd := newPromotionsTestCmd()

	err := promotionsDeleteCmd.RunE(cmd, []string{"promo_123"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPromotionsDeleteRunE_ConfirmUpperY tests delete command confirmation with "Y".
func TestPromotionsDeleteRunE_ConfirmUpperY(t *testing.T) {
	mockClient := &mockPromotionsClient{}

	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}
	os.Stdin = r

	go func() {
		_, _ = w.WriteString("Y\n")
		_ = w.Close()
	}()

	cmd := newPromotionsTestCmd()

	err := promotionsDeleteCmd.RunE(cmd, []string{"promo_123"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

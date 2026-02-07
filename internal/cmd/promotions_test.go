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

// mockPromotionsClient is a mock implementation of api.APIClient for promotions testing.
type mockPromotionsClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values for specific methods
	listPromotionsResp *api.PromotionsListResponse
	listPromotionsErr  error

	searchPromotionsResp *api.PromotionsListResponse
	searchPromotionsErr  error

	getPromotionResp *api.Promotion
	getPromotionErr  error

	createPromotionResp *api.Promotion
	createPromotionErr  error
	createPromotionReq  *api.PromotionCreateRequest

	updatePromotionResp *api.Promotion
	updatePromotionErr  error
	updatePromotionReq  *api.PromotionUpdateRequest

	activatePromotionResp *api.Promotion
	activatePromotionErr  error

	deactivatePromotionResp *api.Promotion
	deactivatePromotionErr  error

	deletePromotionErr error

	getCouponCenterResp json.RawMessage
	getCouponCenterErr  error
}

func (m *mockPromotionsClient) ListPromotions(ctx context.Context, opts *api.PromotionsListOptions) (*api.PromotionsListResponse, error) {
	return m.listPromotionsResp, m.listPromotionsErr
}

func (m *mockPromotionsClient) SearchPromotions(ctx context.Context, opts *api.PromotionSearchOptions) (*api.PromotionsListResponse, error) {
	return m.searchPromotionsResp, m.searchPromotionsErr
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

func (m *mockPromotionsClient) CreatePromotion(ctx context.Context, req *api.PromotionCreateRequest) (*api.Promotion, error) {
	m.createPromotionReq = req
	return m.createPromotionResp, m.createPromotionErr
}

func (m *mockPromotionsClient) UpdatePromotion(ctx context.Context, id string, req *api.PromotionUpdateRequest) (*api.Promotion, error) {
	m.updatePromotionReq = req
	return m.updatePromotionResp, m.updatePromotionErr
}

func (m *mockPromotionsClient) GetPromotionsCouponCenter(ctx context.Context) (json.RawMessage, error) {
	return m.getCouponCenterResp, m.getCouponCenterErr
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
		"list":          "List promotions",
		"get":           "Get promotion details",
		"activate":      "Activate a promotion",
		"deactivate":    "Deactivate a promotion",
		"delete":        "Delete a promotion",
		"coupon-center": "Get coupon center promotions (documented endpoint; raw JSON)",
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

func TestPromotionsCouponCenterRunE(t *testing.T) {
	mockClient := &mockPromotionsClient{
		getCouponCenterResp: json.RawMessage(`{"items":[]}`),
	}
	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newPromotionsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	if err := promotionsCouponCenterCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"items\"") {
		t.Fatalf("expected items in output, got %q", buf.String())
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
	if promotionsGetCmd.Use != "get [id]" {
		t.Errorf("expected Use 'get [id]', got %q", promotionsGetCmd.Use)
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

// --- Create shorthand flags tests ---

// newPromotionsCreateTestCmd creates a test command with all flags needed for promotions create.
func newPromotionsCreateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	addJSONBodyFlags(cmd)
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("discount-type", "", "")
	cmd.Flags().Float64("discount-value", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	cmd.Flags().Int("usage-limit", 0, "")
	cmd.Flags().String("status", "", "")
	return cmd
}

// newPromotionsUpdateTestCmd creates a test command with all flags needed for promotions update.
func newPromotionsUpdateTestCmd() *cobra.Command {
	cmd := newPromotionsCreateTestCmd() // same flags
	return cmd
}

// TestPromotionsCreateShorthandFlags tests the create command with shorthand flags.
func TestPromotionsCreateShorthandFlags(t *testing.T) {
	tests := []struct {
		name              string
		flags             map[string]string
		wantTitle         string
		wantDiscountType  string
		wantDiscountValue float64
		wantUsageLimit    int
		wantErr           bool
		wantErrMsg        string
	}{
		{
			name: "all flags",
			flags: map[string]string{
				"title":          "Summer Sale",
				"discount-type":  "percentage",
				"discount-value": "20",
				"starts-at":      "2026-03-01",
				"ends-at":        "2026-06-01",
				"usage-limit":    "100",
				"status":         "active",
			},
			wantTitle:         "Summer Sale",
			wantDiscountType:  "percentage",
			wantDiscountValue: 20,
			wantUsageLimit:    100,
		},
		{
			name:             "title only",
			flags:            map[string]string{"title": "Flash Sale"},
			wantTitle:        "Flash Sale",
			wantDiscountType: "",
		},
		{
			name:              "discount flags only",
			flags:             map[string]string{"discount-type": "fixed_amount", "discount-value": "10.5"},
			wantDiscountType:  "fixed_amount",
			wantDiscountValue: 10.5,
		},
		{
			name:       "body and flags conflict",
			flags:      map[string]string{"body": `{"title":"x"}`, "title": "y"},
			wantErr:    true,
			wantErrMsg: "use either --body/--body-file or individual flags, not both",
		},
		{
			name:       "no input at all",
			flags:      map[string]string{},
			wantErr:    true,
			wantErrMsg: "provide promotion data via --body/--body-file or individual flags",
		},
		{
			name:       "invalid starts-at date",
			flags:      map[string]string{"title": "Bad Date", "starts-at": "not-a-date"},
			wantErr:    true,
			wantErrMsg: "invalid --starts-at format",
		},
		{
			name:       "invalid ends-at date",
			flags:      map[string]string{"title": "Bad Date", "ends-at": "not-a-date"},
			wantErr:    true,
			wantErrMsg: "invalid --ends-at format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				createPromotionResp: &api.Promotion{
					ID:     "promo_new",
					Status: "active",
				},
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newPromotionsCreateTestCmd()
			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			err := promotionsCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			req := mockClient.createPromotionReq
			if req == nil {
				t.Fatal("expected CreatePromotion to be called")
			}
			if tt.wantTitle != "" && req.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", req.Title, tt.wantTitle)
			}
			if tt.wantDiscountType != "" && req.DiscountType != tt.wantDiscountType {
				t.Errorf("discount_type = %q, want %q", req.DiscountType, tt.wantDiscountType)
			}
			if tt.wantDiscountValue != 0 && req.DiscountValue != tt.wantDiscountValue {
				t.Errorf("discount_value = %f, want %f", req.DiscountValue, tt.wantDiscountValue)
			}
			if tt.wantUsageLimit != 0 && req.UsageLimit != tt.wantUsageLimit {
				t.Errorf("usage_limit = %d, want %d", req.UsageLimit, tt.wantUsageLimit)
			}
		})
	}
}

// TestPromotionsCreateWithBody tests that --body still works for create.
func TestPromotionsCreateWithBody(t *testing.T) {
	mockClient := &mockPromotionsClient{
		createPromotionResp: &api.Promotion{
			ID:     "promo_body",
			Status: "active",
		},
	}
	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newPromotionsCreateTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"title":"Body Promo","discount_type":"percentage","discount_value":15,"starts_at":"2026-03-01T00:00:00Z"}`)

	err := promotionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := mockClient.createPromotionReq
	if req == nil {
		t.Fatal("expected CreatePromotion to be called")
	}
	if req.Title != "Body Promo" {
		t.Errorf("title = %q, want %q", req.Title, "Body Promo")
	}
	if req.DiscountType != "percentage" {
		t.Errorf("discount_type = %q, want %q", req.DiscountType, "percentage")
	}
}

// TestPromotionsCreateDryRun tests that dry-run skips shorthand flag validation.
func TestPromotionsCreateDryRun(t *testing.T) {
	mockClient := &mockPromotionsClient{}
	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	cmd := newPromotionsCreateTestCmd()
	cmd.SetOut(&buf)
	_ = cmd.Flags().Set("dry-run", "true")

	err := promotionsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.createPromotionReq != nil {
		t.Error("expected CreatePromotion NOT to be called in dry-run mode")
	}
	if !strings.Contains(buf.String(), "[DRY-RUN]") {
		t.Errorf("expected dry-run output, got %q", buf.String())
	}
}

// TestPromotionsCreateWithDateFormats tests both RFC3339 and YYYY-MM-DD date parsing.
func TestPromotionsCreateWithDateFormats(t *testing.T) {
	tests := []struct {
		name     string
		startsAt string
		endsAt   string
		wantYear int
	}{
		{
			name:     "YYYY-MM-DD format",
			startsAt: "2026-03-01",
			endsAt:   "2026-06-01",
			wantYear: 2026,
		},
		{
			name:     "RFC3339 format",
			startsAt: "2026-03-01T10:00:00Z",
			endsAt:   "2026-06-01T23:59:59Z",
			wantYear: 2026,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				createPromotionResp: &api.Promotion{
					ID:     "promo_date",
					Status: "active",
				},
			}
			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newPromotionsCreateTestCmd()
			_ = cmd.Flags().Set("title", "Date Test")
			_ = cmd.Flags().Set("starts-at", tt.startsAt)
			_ = cmd.Flags().Set("ends-at", tt.endsAt)

			err := promotionsCreateCmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			req := mockClient.createPromotionReq
			if req == nil {
				t.Fatal("expected CreatePromotion to be called")
			}
			if req.StartsAt.Year() != tt.wantYear {
				t.Errorf("starts_at year = %d, want %d", req.StartsAt.Year(), tt.wantYear)
			}
			if req.EndsAt.Year() != tt.wantYear {
				t.Errorf("ends_at year = %d, want %d", req.EndsAt.Year(), tt.wantYear)
			}
		})
	}
}

// --- Update shorthand flags tests ---

// TestPromotionsUpdateShorthandFlags tests the update command with shorthand flags.
func TestPromotionsUpdateShorthandFlags(t *testing.T) {
	tests := []struct {
		name              string
		flags             map[string]string
		wantTitle         *string
		wantDiscountType  *string
		wantDiscountValue *float64
		wantUsageLimit    *int
		wantErr           bool
		wantErrMsg        string
	}{
		{
			name: "all flags",
			flags: map[string]string{
				"title":          "Updated Sale",
				"discount-type":  "fixed_amount",
				"discount-value": "25",
				"starts-at":      "2026-04-01",
				"ends-at":        "2026-07-01",
				"usage-limit":    "200",
				"status":         "inactive",
			},
			wantTitle:         ptrString("Updated Sale"),
			wantDiscountType:  ptrString("fixed_amount"),
			wantDiscountValue: ptrFloat64(25),
			wantUsageLimit:    ptrInt(200),
		},
		{
			name:      "title only (partial update)",
			flags:     map[string]string{"title": "New Title"},
			wantTitle: ptrString("New Title"),
		},
		{
			name:              "discount-value only (partial update)",
			flags:             map[string]string{"discount-value": "30.5"},
			wantDiscountValue: ptrFloat64(30.5),
		},
		{
			name:       "body and flags conflict",
			flags:      map[string]string{"body": `{"title":"x"}`, "title": "y"},
			wantErr:    true,
			wantErrMsg: "use either --body/--body-file or individual flags, not both",
		},
		{
			name:       "no input at all",
			flags:      map[string]string{},
			wantErr:    true,
			wantErrMsg: "provide promotion data via --body/--body-file or individual flags",
		},
		{
			name:       "invalid starts-at date",
			flags:      map[string]string{"starts-at": "bad-date"},
			wantErr:    true,
			wantErrMsg: "invalid --starts-at format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPromotionsClient{
				updatePromotionResp: &api.Promotion{
					ID:     "promo_upd",
					Status: "active",
				},
			}

			restore := setupPromotionsTest(t, mockClient)
			defer restore()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newPromotionsUpdateTestCmd()
			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			err := promotionsUpdateCmd.RunE(cmd, []string{"promo_123"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			req := mockClient.updatePromotionReq
			if req == nil {
				t.Fatal("expected UpdatePromotion to be called")
			}

			if tt.wantTitle != nil {
				if req.Title == nil {
					t.Error("expected Title to be set")
				} else if *req.Title != *tt.wantTitle {
					t.Errorf("title = %q, want %q", *req.Title, *tt.wantTitle)
				}
			}
			if tt.wantDiscountType != nil {
				if req.DiscountType == nil {
					t.Error("expected DiscountType to be set")
				} else if *req.DiscountType != *tt.wantDiscountType {
					t.Errorf("discount_type = %q, want %q", *req.DiscountType, *tt.wantDiscountType)
				}
			}
			if tt.wantDiscountValue != nil {
				if req.DiscountValue == nil {
					t.Error("expected DiscountValue to be set")
				} else if *req.DiscountValue != *tt.wantDiscountValue {
					t.Errorf("discount_value = %f, want %f", *req.DiscountValue, *tt.wantDiscountValue)
				}
			}
			if tt.wantUsageLimit != nil {
				if req.UsageLimit == nil {
					t.Error("expected UsageLimit to be set")
				} else if *req.UsageLimit != *tt.wantUsageLimit {
					t.Errorf("usage_limit = %d, want %d", *req.UsageLimit, *tt.wantUsageLimit)
				}
			}
		})
	}
}

// TestPromotionsUpdateWithBody tests that --body still works for update.
func TestPromotionsUpdateWithBody(t *testing.T) {
	mockClient := &mockPromotionsClient{
		updatePromotionResp: &api.Promotion{
			ID:     "promo_body_upd",
			Status: "active",
		},
	}
	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newPromotionsUpdateTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"title":"Body Update"}`)

	err := promotionsUpdateCmd.RunE(cmd, []string{"promo_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := mockClient.updatePromotionReq
	if req == nil {
		t.Fatal("expected UpdatePromotion to be called")
	}
	if req.Title == nil || *req.Title != "Body Update" {
		t.Errorf("title = %v, want %q", req.Title, "Body Update")
	}
}

// TestPromotionsUpdateDryRun tests that dry-run skips shorthand flag validation for update.
func TestPromotionsUpdateDryRun(t *testing.T) {
	mockClient := &mockPromotionsClient{}
	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	cmd := newPromotionsUpdateTestCmd()
	cmd.SetOut(&buf)
	_ = cmd.Flags().Set("dry-run", "true")

	err := promotionsUpdateCmd.RunE(cmd, []string{"promo_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockClient.updatePromotionReq != nil {
		t.Error("expected UpdatePromotion NOT to be called in dry-run mode")
	}
	if !strings.Contains(buf.String(), "[DRY-RUN]") {
		t.Errorf("expected dry-run output, got %q", buf.String())
	}
}

// TestPromotionsUpdatePartialFlags tests that only Changed() flags are set in the update request.
func TestPromotionsUpdatePartialFlags(t *testing.T) {
	mockClient := &mockPromotionsClient{
		updatePromotionResp: &api.Promotion{
			ID:     "promo_partial",
			Status: "active",
		},
	}
	restore := setupPromotionsTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newPromotionsUpdateTestCmd()
	// Only set title -- other fields should remain nil
	_ = cmd.Flags().Set("title", "Partial Update")

	err := promotionsUpdateCmd.RunE(cmd, []string{"promo_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := mockClient.updatePromotionReq
	if req == nil {
		t.Fatal("expected UpdatePromotion to be called")
	}
	if req.Title == nil || *req.Title != "Partial Update" {
		t.Errorf("title = %v, want %q", req.Title, "Partial Update")
	}
	if req.DiscountType != nil {
		t.Errorf("discount_type should be nil, got %v", req.DiscountType)
	}
	if req.DiscountValue != nil {
		t.Errorf("discount_value should be nil, got %v", req.DiscountValue)
	}
	if req.UsageLimit != nil {
		t.Errorf("usage_limit should be nil, got %v", req.UsageLimit)
	}
	if req.StartsAt != nil {
		t.Errorf("starts_at should be nil, got %v", req.StartsAt)
	}
	if req.EndsAt != nil {
		t.Errorf("ends_at should be nil, got %v", req.EndsAt)
	}
}

// TestParsePromotionTime tests the time parsing helper.
func TestParsePromotionTime(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		wantY   int
		wantM   time.Month
		wantD   int
	}{
		{"2026-03-01", false, 2026, time.March, 1},
		{"2026-03-01T10:00:00Z", false, 2026, time.March, 1},
		{"not-a-date", true, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parsePromotionTime(tt.input, "test")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Year() != tt.wantY || got.Month() != tt.wantM || got.Day() != tt.wantD {
				t.Errorf("got %v, want %d-%02d-%02d", got, tt.wantY, tt.wantM, tt.wantD)
			}
		})
	}
}

// TestPromotionsGetByFlag tests the --by flag on the promotions get command.
func TestPromotionsGetByFlag(t *testing.T) {
	t.Run("resolves promotion by title", func(t *testing.T) {
		mockClient := &mockPromotionsClient{
			searchPromotionsResp: &api.PromotionsListResponse{
				Items:      []api.Promotion{{ID: "promo_found", Title: "Summer Sale"}},
				TotalCount: 1,
			},
			getPromotionResp: &api.Promotion{
				ID:        "promo_found",
				Title:     "Summer Sale",
				StartsAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		restore := setupPromotionsTest(t, mockClient)
		defer restore()

		var buf bytes.Buffer
		formatterWriter = &buf

		cmd := newPromotionsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Summer Sale")

		if err := promotionsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "promo_found") {
			t.Errorf("expected output to contain 'promo_found', got: %s", buf.String())
		}
	})

	t.Run("errors when no match", func(t *testing.T) {
		mockClient := &mockPromotionsClient{
			searchPromotionsResp: &api.PromotionsListResponse{
				Items:      []api.Promotion{},
				TotalCount: 0,
			},
		}
		restore := setupPromotionsTest(t, mockClient)
		defer restore()

		cmd := newPromotionsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "nonexistent")

		err := promotionsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when no promotion found")
		}
		if !strings.Contains(err.Error(), "no promotion found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("errors when search fails", func(t *testing.T) {
		mockClient := &mockPromotionsClient{
			searchPromotionsErr: errors.New("API error"),
		}
		restore := setupPromotionsTest(t, mockClient)
		defer restore()

		cmd := newPromotionsTestCmd()
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Summer Sale")

		err := promotionsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error when search fails")
		}
		if !strings.Contains(err.Error(), "search failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("warns on multiple matches", func(t *testing.T) {
		mockClient := &mockPromotionsClient{
			searchPromotionsResp: &api.PromotionsListResponse{
				Items: []api.Promotion{
					{ID: "promo_1", Title: "Summer Sale A"},
					{ID: "promo_2", Title: "Summer Sale B"},
				},
				TotalCount: 2,
			},
			getPromotionResp: &api.Promotion{
				ID:        "promo_1",
				Title:     "Summer Sale A",
				StartsAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		restore := setupPromotionsTest(t, mockClient)
		defer restore()

		var buf bytes.Buffer
		formatterWriter = &buf

		cmd := newPromotionsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "Summer Sale")

		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)

		if err := promotionsGetCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "promo_1") {
			t.Errorf("expected output to contain 'promo_1', got: %s", buf.String())
		}
		if !strings.Contains(stderr.String(), "2 matches found") {
			t.Errorf("expected stderr warning about multiple matches, got: %s", stderr.String())
		}
	})

	t.Run("positional arg takes precedence over --by", func(t *testing.T) {
		mockClient := &mockPromotionsClient{
			getPromotionResp: &api.Promotion{
				ID:        "promo_direct",
				Title:     "Direct Promo",
				StartsAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		restore := setupPromotionsTest(t, mockClient)
		defer restore()

		var buf bytes.Buffer
		formatterWriter = &buf

		cmd := newPromotionsTestCmd()
		_ = cmd.Flags().Set("output", "json")
		cmd.Flags().String("by", "", "")
		_ = cmd.Flags().Set("by", "should-not-be-used")

		if err := promotionsGetCmd.RunE(cmd, []string{"promo_direct"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(buf.String(), "promo_direct") {
			t.Errorf("expected output to contain 'promo_direct', got: %s", buf.String())
		}
	})

	t.Run("errors with no arg and no --by", func(t *testing.T) {
		mockClient := &mockPromotionsClient{}
		restore := setupPromotionsTest(t, mockClient)
		defer restore()

		cmd := newPromotionsTestCmd()
		cmd.Flags().String("by", "", "")

		err := promotionsGetCmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error with no arg and no --by")
		}
		if !strings.Contains(err.Error(), "provide a resource ID") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Pointer helper functions for test assertions.
func ptrString(s string) *string    { return &s }
func ptrFloat64(f float64) *float64 { return &f }
func ptrInt(i int) *int             { return &i }

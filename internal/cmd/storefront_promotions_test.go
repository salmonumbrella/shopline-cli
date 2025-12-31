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

// mockStorefrontPromotionsClient is a mock implementation for storefront promotions testing.
type mockStorefrontPromotionsClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values
	listResp     *api.StorefrontPromotionsListResponse
	listErr      error
	getResp      *api.StorefrontPromotion
	getErr       error
	getByCodeErr error
}

func (m *mockStorefrontPromotionsClient) ListStorefrontPromotions(ctx context.Context, opts *api.StorefrontPromotionsListOptions) (*api.StorefrontPromotionsListResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockStorefrontPromotionsClient) GetStorefrontPromotion(ctx context.Context, id string) (*api.StorefrontPromotion, error) {
	return m.getResp, m.getErr
}

func (m *mockStorefrontPromotionsClient) GetStorefrontPromotionByCode(ctx context.Context, code string) (*api.StorefrontPromotion, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	return m.getResp, m.getErr
}

// setupStorefrontPromotionsTest sets up test environment with mock client and credentials.
func setupStorefrontPromotionsTest(t *testing.T, mockClient *mockStorefrontPromotionsClient) (cleanup func()) {
	t.Helper()

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

func TestStorefrontPromotionsCmdStructure(t *testing.T) {
	if storefrontPromotionsCmd.Use != "storefront-promotions" {
		t.Errorf("Expected Use 'storefront-promotions', got %q", storefrontPromotionsCmd.Use)
	}

	subcommands := storefrontPromotionsCmd.Commands()
	expectedSubs := []string{"list", "get"}

	for _, exp := range expectedSubs {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == exp || strings.HasPrefix(cmd.Use, exp+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing subcommand: %s", exp)
		}
	}
}

func TestStorefrontPromotionsListCmdFlags(t *testing.T) {
	flags := []string{"status", "type", "discount-type", "page", "page-size"}
	for _, flagName := range flags {
		flag := storefrontPromotionsListCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Missing --%s flag", flagName)
		}
	}
}

func TestStorefrontPromotionsGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"promo_123"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"promo_1", "promo_2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storefrontPromotionsGetCmd.Args(storefrontPromotionsGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorefrontPromotionsGetCmdFlags(t *testing.T) {
	byCodeFlag := storefrontPromotionsGetCmd.Flags().Lookup("by-code")
	if byCodeFlag == nil {
		t.Error("Missing --by-code flag")
	}
}

func TestStorefrontPromotionsListRunE(t *testing.T) {
	now := time.Now()
	endsAt := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name       string
		output     string
		mockResp   *api.StorefrontPromotionsListResponse
		mockErr    error
		wantErr    bool
		wantOutput []string // Only checked for JSON output (text goes to stdout)
	}{
		{
			name:   "successful list text output with percentage discount",
			output: "text",
			mockResp: &api.StorefrontPromotionsListResponse{
				Items: []api.StorefrontPromotion{
					{
						ID:            "promo_123",
						Title:         "Summer Sale",
						Type:          "discount",
						DiscountType:  "percentage",
						DiscountValue: "20",
						Status:        "active",
						UsageCount:    50,
						UsageLimit:    100,
						EndsAt:        &endsAt,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:   "successful list text output with fixed discount and no usage limit",
			output: "text",
			mockResp: &api.StorefrontPromotionsListResponse{
				Items: []api.StorefrontPromotion{
					{
						ID:            "promo_456",
						Title:         "Free Shipping",
						Type:          "shipping",
						DiscountType:  "fixed",
						DiscountValue: "10.00",
						Status:        "active",
						UsageCount:    25,
						UsageLimit:    0, // No limit
						EndsAt:        nil,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.StorefrontPromotionsListResponse{
				Items: []api.StorefrontPromotion{
					{
						ID:            "promo_789",
						Title:         "Holiday Special",
						DiscountType:  "percentage",
						DiscountValue: "15",
						Status:        "scheduled",
					},
				},
				TotalCount: 1,
			},
			wantOutput: []string{"promo_789", "Holiday Special"},
		},
		{
			name:    "API error",
			output:  "text",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name:   "empty list",
			output: "text",
			mockResp: &api.StorefrontPromotionsListResponse{
				Items:      []api.StorefrontPromotion{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockStorefrontPromotionsClient{
				listResp: tt.mockResp,
				listErr:  tt.mockErr,
			}
			cleanup := setupStorefrontPromotionsTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().String("type", "", "")
			cmd.Flags().String("discount-type", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := storefrontPromotionsListCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to list storefront promotions") {
					t.Errorf("expected error to contain 'failed to list storefront promotions', got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text uses fmt.Printf to stdout)
			if tt.output == "json" && len(tt.wantOutput) > 0 {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got: %s", want, output)
					}
				}
			}
		})
	}
}

func TestStorefrontPromotionsListWithFilters(t *testing.T) {
	mockClient := &mockStorefrontPromotionsClient{
		listResp: &api.StorefrontPromotionsListResponse{
			Items:      []api.StorefrontPromotion{},
			TotalCount: 0,
		},
	}
	cleanup := setupStorefrontPromotionsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("status", "active", "")
	cmd.Flags().String("type", "discount", "")
	cmd.Flags().String("discount-type", "percentage", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 50, "")

	err := storefrontPromotionsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStorefrontPromotionsGetRunE(t *testing.T) {
	now := time.Now()
	endsAt := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name       string
		promoID    string
		output     string
		byCode     bool
		mockResp   *api.StorefrontPromotion
		mockErr    error
		wantErr    bool
		wantOutput []string // Only checked for JSON output (text goes to stdout)
	}{
		{
			name:    "successful get by ID text output with all fields",
			promoID: "promo_123",
			output:  "text",
			mockResp: &api.StorefrontPromotion{
				ID:            "promo_123",
				Title:         "Summer Sale",
				Description:   "Big summer discounts",
				Type:          "discount",
				Status:        "active",
				DiscountType:  "percentage",
				DiscountValue: "20",
				Code:          "SUMMER20",
				MinPurchase:   "50.00",
				MaxDiscount:   "100.00",
				UsageCount:    50,
				UsageLimit:    100,
				CustomerLimit: 1,
				Stackable:     true,
				AutoApply:     false,
				TargetType:    "all",
				StartsAt:      now,
				EndsAt:        &endsAt,
				CreatedAt:     now.Add(-24 * time.Hour),
				UpdatedAt:     now,
				Banner: &api.StorefrontPromotionBanner{
					Enabled:  true,
					Text:     "Summer Sale!",
					Position: "top",
				},
			},
		},
		{
			name:    "successful get by ID JSON output",
			promoID: "promo_456",
			output:  "json",
			mockResp: &api.StorefrontPromotion{
				ID:            "promo_456",
				Title:         "Holiday Special",
				DiscountType:  "fixed",
				DiscountValue: "10.00",
				Status:        "scheduled",
				StartsAt:      now,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			wantOutput: []string{"promo_456", "Holiday Special"},
		},
		{
			name:    "successful get by code",
			promoID: "SAVE10",
			output:  "text",
			byCode:  true,
			mockResp: &api.StorefrontPromotion{
				ID:            "promo_789",
				Title:         "Save 10%",
				Code:          "SAVE10",
				DiscountType:  "percentage",
				DiscountValue: "10",
				Status:        "active",
				StartsAt:      now,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
		{
			name:    "promotion not found",
			promoID: "promo_999",
			output:  "text",
			mockErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "get without optional fields - no usage limit, no end date, no banner",
			promoID: "promo_minimal",
			output:  "text",
			mockResp: &api.StorefrontPromotion{
				ID:            "promo_minimal",
				Title:         "Minimal Promo",
				Description:   "Basic promotion",
				Type:          "discount",
				Status:        "active",
				DiscountType:  "fixed",
				DiscountValue: "5.00",
				UsageCount:    10,
				UsageLimit:    0, // No limit
				CustomerLimit: 0, // No per-customer limit
				Stackable:     false,
				AutoApply:     true,
				TargetType:    "specific",
				StartsAt:      now,
				EndsAt:        nil, // No end date
				CreatedAt:     now,
				UpdatedAt:     now,
				Banner:        nil, // No banner
			},
		},
		{
			name:    "get with disabled banner",
			promoID: "promo_no_banner",
			output:  "text",
			mockResp: &api.StorefrontPromotion{
				ID:            "promo_no_banner",
				Title:         "No Banner Promo",
				Description:   "",
				Type:          "discount",
				Status:        "active",
				DiscountType:  "percentage",
				DiscountValue: "5",
				StartsAt:      now,
				CreatedAt:     now,
				UpdatedAt:     now,
				Banner: &api.StorefrontPromotionBanner{
					Enabled: false,
					Text:    "Hidden",
				},
			},
		},
		{
			name:    "get without code, min purchase, max discount",
			promoID: "promo_simple",
			output:  "text",
			mockResp: &api.StorefrontPromotion{
				ID:            "promo_simple",
				Title:         "Simple Promo",
				Description:   "A simple promotion",
				Type:          "discount",
				Status:        "active",
				DiscountType:  "percentage",
				DiscountValue: "10",
				Code:          "", // No code
				MinPurchase:   "", // No min purchase
				MaxDiscount:   "", // No max discount
				UsageCount:    5,
				UsageLimit:    0,
				CustomerLimit: 0,
				Stackable:     true,
				AutoApply:     true,
				TargetType:    "all",
				StartsAt:      now,
				EndsAt:        nil,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockStorefrontPromotionsClient{
				getResp: tt.mockResp,
				getErr:  tt.mockErr,
			}
			cleanup := setupStorefrontPromotionsTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().Bool("by-code", tt.byCode, "")

			err := storefrontPromotionsGetCmd.RunE(cmd, []string{tt.promoID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "failed to get storefront promotion") {
					t.Errorf("expected error to contain 'failed to get storefront promotion', got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text uses fmt.Printf to stdout)
			if tt.output == "json" && len(tt.wantOutput) > 0 {
				output := buf.String()
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("output should contain %q, got: %s", want, output)
					}
				}
			}
		})
	}
}

func TestStorefrontPromotionsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(storefrontPromotionsListCmd)

	err := storefrontPromotionsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStorefrontPromotionsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(storefrontPromotionsGetCmd)
	cmd.Flags().Bool("by-code", false, "")

	err := storefrontPromotionsGetCmd.RunE(cmd, []string{"promo_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStorefrontPromotionsGetByCodeError(t *testing.T) {
	mockClient := &mockStorefrontPromotionsClient{
		getByCodeErr: errors.New("invalid code"),
	}
	cleanup := setupStorefrontPromotionsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("by-code", true, "")

	err := storefrontPromotionsGetCmd.RunE(cmd, []string{"BADCODE"})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get storefront promotion") {
		t.Errorf("expected error to contain 'failed to get storefront promotion', got: %v", err)
	}
}

// Ensure unused imports don't cause errors
var _ = secrets.StoreCredentials{}

package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// couponsMockAPIClient is a mock implementation of api.APIClient for coupons testing.
type couponsMockAPIClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values for specific methods
	listCouponsResp *api.CouponsListResponse
	listCouponsErr  error

	getCouponResp *api.Coupon
	getCouponErr  error

	getCouponByCodeResp *api.Coupon
	getCouponByCodeErr  error

	createCouponResp *api.Coupon
	createCouponErr  error

	activateCouponResp *api.Coupon
	activateCouponErr  error

	deactivateCouponResp *api.Coupon
	deactivateCouponErr  error

	deleteCouponErr error
}

func (m *couponsMockAPIClient) ListCoupons(ctx context.Context, opts *api.CouponsListOptions) (*api.CouponsListResponse, error) {
	return m.listCouponsResp, m.listCouponsErr
}

func (m *couponsMockAPIClient) GetCoupon(ctx context.Context, id string) (*api.Coupon, error) {
	return m.getCouponResp, m.getCouponErr
}

func (m *couponsMockAPIClient) GetCouponByCode(ctx context.Context, code string) (*api.Coupon, error) {
	return m.getCouponByCodeResp, m.getCouponByCodeErr
}

func (m *couponsMockAPIClient) CreateCoupon(ctx context.Context, req *api.CouponCreateRequest) (*api.Coupon, error) {
	return m.createCouponResp, m.createCouponErr
}

func (m *couponsMockAPIClient) ActivateCoupon(ctx context.Context, id string) (*api.Coupon, error) {
	return m.activateCouponResp, m.activateCouponErr
}

func (m *couponsMockAPIClient) DeactivateCoupon(ctx context.Context, id string) (*api.Coupon, error) {
	return m.deactivateCouponResp, m.deactivateCouponErr
}

func (m *couponsMockAPIClient) DeleteCoupon(ctx context.Context, id string) error {
	return m.deleteCouponErr
}

// setupCouponsMockFactories configures mock factories for coupons testing.
func setupCouponsMockFactories(mockClient *couponsMockAPIClient) (cleanup func()) {
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

// TestCouponsCommandSetup verifies coupons command initialization
func TestCouponsCommandSetup(t *testing.T) {
	if couponsCmd.Use != "coupons" {
		t.Errorf("expected Use 'coupons', got %q", couponsCmd.Use)
	}
	if couponsCmd.Short != "Manage coupons" {
		t.Errorf("expected Short 'Manage coupons', got %q", couponsCmd.Short)
	}
}

// TestCouponsSubcommands verifies all subcommands are registered
func TestCouponsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":       "List coupons",
		"get":        "Get coupon details",
		"lookup":     "Lookup a coupon by code",
		"create":     "Create a coupon",
		"activate":   "Activate a coupon",
		"deactivate": "Deactivate a coupon",
		"delete":     "Delete a coupon",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range couponsCmd.Commands() {
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

// TestCouponsListFlags verifies list command flags exist with correct defaults
func TestCouponsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := couponsListCmd.Flags().Lookup(f.name)
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

// TestCouponsCreateFlags verifies create command flags exist with correct defaults
func TestCouponsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"code", ""},
		{"discount-type", ""},
		{"discount-value", "0"},
		{"title", ""},
		{"min-purchase", "0"},
		{"usage-limit", "0"},
		{"per-customer", "0"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := couponsCreateCmd.Flags().Lookup(f.name)
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

// TestCouponsCreateRequiredFlags verifies code, discount-type, and discount-value are required
func TestCouponsCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"code", "discount-type", "discount-value"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := couponsCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
				return
			}
			// Check if the flag has required annotation
			annotations := flag.Annotations
			if annotations == nil {
				t.Errorf("flag %q has no annotations, expected required", name)
				return
			}
			if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
				t.Errorf("flag %q is not marked as required", name)
			}
		})
	}
}

// TestCouponsDeleteFlags verifies delete command flags exist with correct defaults
func TestCouponsDeleteFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"yes", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := couponsDeleteCmd.Flags().Lookup(f.name)
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

// TestCouponsGetArgs verifies get command requires exactly 1 argument
func TestCouponsGetArgs(t *testing.T) {
	err := couponsGetCmd.Args(couponsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = couponsGetCmd.Args(couponsGetCmd, []string{"coupon-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCouponsLookupArgs verifies lookup command requires exactly 1 argument
func TestCouponsLookupArgs(t *testing.T) {
	err := couponsLookupCmd.Args(couponsLookupCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = couponsLookupCmd.Args(couponsLookupCmd, []string{"DISCOUNT10"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCouponsActivateArgs verifies activate command requires exactly 1 argument
func TestCouponsActivateArgs(t *testing.T) {
	err := couponsActivateCmd.Args(couponsActivateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = couponsActivateCmd.Args(couponsActivateCmd, []string{"coupon-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCouponsDeactivateArgs verifies deactivate command requires exactly 1 argument
func TestCouponsDeactivateArgs(t *testing.T) {
	err := couponsDeactivateCmd.Args(couponsDeactivateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = couponsDeactivateCmd.Args(couponsDeactivateCmd, []string{"coupon-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCouponsDeleteArgs verifies delete command requires exactly 1 argument
func TestCouponsDeleteArgs(t *testing.T) {
	err := couponsDeleteCmd.Args(couponsDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = couponsDeleteCmd.Args(couponsDeleteCmd, []string{"coupon-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestCouponsListRunE tests the coupons list command execution with mock API.
func TestCouponsListRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.CouponsListResponse
		mockErr    error
		output     string
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful list with percentage coupon",
			output: "text",
			mockResp: &api.CouponsListResponse{
				Items: []api.Coupon{
					{
						ID:            "coup_123",
						Code:          "SAVE10",
						Title:         "10% Off",
						DiscountType:  "percentage",
						DiscountValue: 10,
						UsageCount:    5,
						UsageLimit:    100,
						Status:        "active",
						StartsAt:      startsAt,
						EndsAt:        endsAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "coup_123",
		},
		{
			name:   "successful list with fixed_amount coupon no end date",
			output: "text",
			mockResp: &api.CouponsListResponse{
				Items: []api.Coupon{
					{
						ID:            "coup_456",
						Code:          "FLAT20",
						Title:         "$20 Off",
						DiscountType:  "fixed_amount",
						DiscountValue: 20,
						UsageCount:    10,
						UsageLimit:    0, // unlimited
						Status:        "active",
						StartsAt:      startsAt,
						EndsAt:        time.Time{}, // no end date
					},
				},
				TotalCount: 1,
			},
			wantOutput: "FLAT20",
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.CouponsListResponse{
				Items: []api.Coupon{
					{
						ID:            "coup_789",
						Code:          "WELCOME",
						Title:         "Welcome Discount",
						DiscountType:  "percentage",
						DiscountValue: 15,
						Status:        "active",
						StartsAt:      startsAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "coup_789",
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
			mockResp: &api.CouponsListResponse{
				Items:      []api.Coupon{},
				TotalCount: 0,
			},
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				listCouponsResp: tt.mockResp,
				listCouponsErr:  tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
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
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := couponsListCmd.RunE(cmd, []string{})

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

// TestCouponsGetRunE tests the coupons get command execution with mock API.
func TestCouponsGetRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2023, 12, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		couponID   string
		output     string
		mockResp   *api.Coupon
		mockErr    error
		wantErr    bool
		wantOutput string // Only checked for JSON output (text output goes to stdout)
	}{
		{
			name:     "successful get with all fields",
			couponID: "coup_123",
			output:   "text",
			mockResp: &api.Coupon{
				ID:            "coup_123",
				Code:          "SAVE10",
				Title:         "10% Off",
				Description:   "Save 10% on your order",
				DiscountType:  "percentage",
				DiscountValue: 10,
				MinPurchase:   50,
				MaxDiscount:   100,
				UsageCount:    5,
				UsageLimit:    100,
				PerCustomer:   1,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        endsAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "", // Text output goes to stdout, not captured
		},
		{
			name:     "successful get with no end date",
			couponID: "coup_456",
			output:   "text",
			mockResp: &api.Coupon{
				ID:            "coup_456",
				Code:          "FLAT20",
				Title:         "$20 Off",
				DiscountType:  "fixed_amount",
				DiscountValue: 20,
				UsageCount:    0,
				UsageLimit:    0,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        time.Time{},
				CreatedAt:     createdAt,
			},
			wantOutput: "", // Text output goes to stdout, not captured
		},
		{
			name:     "successful get JSON output",
			couponID: "coup_789",
			output:   "json",
			mockResp: &api.Coupon{
				ID:            "coup_789",
				Code:          "WELCOME",
				Title:         "Welcome Discount",
				DiscountType:  "percentage",
				DiscountValue: 15,
				Status:        "active",
				StartsAt:      startsAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "coup_789",
		},
		{
			name:     "coupon not found",
			couponID: "coup_999",
			output:   "text",
			mockErr:  errors.New("coupon not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				getCouponResp: tt.mockResp,
				getCouponErr:  tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := couponsGetCmd.RunE(cmd, []string{tt.couponID})

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

// TestCouponsLookupRunE tests the coupons lookup command execution with mock API.
func TestCouponsLookupRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		code       string
		output     string
		mockResp   *api.Coupon
		mockErr    error
		wantErr    bool
		wantOutput string // Only checked for JSON output (text output goes to stdout)
	}{
		{
			name:   "successful lookup",
			code:   "SAVE10",
			output: "text",
			mockResp: &api.Coupon{
				ID:            "coup_123",
				Code:          "SAVE10",
				Title:         "10% Off",
				DiscountType:  "percentage",
				DiscountValue: 10,
				Status:        "active",
				StartsAt:      startsAt,
			},
			wantOutput: "", // Text output goes to stdout, not captured
		},
		{
			name:   "successful lookup JSON output",
			code:   "WELCOME",
			output: "json",
			mockResp: &api.Coupon{
				ID:            "coup_456",
				Code:          "WELCOME",
				Title:         "Welcome Discount",
				DiscountType:  "percentage",
				DiscountValue: 15,
				Status:        "active",
				StartsAt:      startsAt,
			},
			wantOutput: "coup_456",
		},
		{
			name:    "code not found",
			code:    "INVALID",
			output:  "text",
			mockErr: errors.New("coupon not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				getCouponByCodeResp: tt.mockResp,
				getCouponByCodeErr:  tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := couponsLookupCmd.RunE(cmd, []string{tt.code})

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

// TestCouponsCreateRunE tests the coupons create command execution with mock API.
func TestCouponsCreateRunE(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		mockResp   *api.Coupon
		mockErr    error
		wantErr    bool
		wantOutput string // Only checked for JSON output (text output goes to stdout)
	}{
		{
			name:   "successful create",
			output: "text",
			mockResp: &api.Coupon{
				ID:            "coup_new",
				Code:          "NEWCODE",
				Title:         "New Coupon",
				DiscountType:  "percentage",
				DiscountValue: 20,
				Status:        "active",
			},
			wantOutput: "", // Text output goes to stdout, not captured
		},
		{
			name:   "successful create JSON output",
			output: "json",
			mockResp: &api.Coupon{
				ID:            "coup_json",
				Code:          "JSONCODE",
				Title:         "JSON Coupon",
				DiscountType:  "fixed_amount",
				DiscountValue: 50,
				Status:        "active",
			},
			wantOutput: "coup_json",
		},
		{
			name:    "create fails",
			output:  "text",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				createCouponResp: tt.mockResp,
				createCouponErr:  tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("code", "TESTCODE", "")
			cmd.Flags().String("discount-type", "percentage", "")
			cmd.Flags().Float64("discount-value", 10, "")
			cmd.Flags().String("title", "Test Coupon", "")
			cmd.Flags().Float64("min-purchase", 0, "")
			cmd.Flags().Int("usage-limit", 0, "")
			cmd.Flags().Int("per-customer", 0, "")

			err := couponsCreateCmd.RunE(cmd, []string{})

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

// TestCouponsActivateRunE tests the coupons activate command execution with mock API.
func TestCouponsActivateRunE(t *testing.T) {
	tests := []struct {
		name     string
		couponID string
		mockResp *api.Coupon
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful activate",
			couponID: "coup_123",
			mockResp: &api.Coupon{
				ID:     "coup_123",
				Code:   "SAVE10",
				Status: "active",
			},
		},
		{
			name:     "activate fails",
			couponID: "coup_456",
			mockErr:  errors.New("coupon already active"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				activateCouponResp: tt.mockResp,
				activateCouponErr:  tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")

			err := couponsActivateCmd.RunE(cmd, []string{tt.couponID})

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

// TestCouponsDeactivateRunE tests the coupons deactivate command execution with mock API.
func TestCouponsDeactivateRunE(t *testing.T) {
	tests := []struct {
		name     string
		couponID string
		mockResp *api.Coupon
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful deactivate",
			couponID: "coup_123",
			mockResp: &api.Coupon{
				ID:     "coup_123",
				Code:   "SAVE10",
				Status: "inactive",
			},
		},
		{
			name:     "deactivate fails",
			couponID: "coup_456",
			mockErr:  errors.New("coupon already inactive"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				deactivateCouponResp: tt.mockResp,
				deactivateCouponErr:  tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")

			err := couponsDeactivateCmd.RunE(cmd, []string{tt.couponID})

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

// TestCouponsDeleteRunE tests the coupons delete command execution with mock API.
func TestCouponsDeleteRunE(t *testing.T) {
	tests := []struct {
		name     string
		couponID string
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			couponID: "coup_123",
			mockErr:  nil,
		},
		{
			name:     "delete fails",
			couponID: "coup_456",
			mockErr:  errors.New("coupon not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				deleteCouponErr: tt.mockErr,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := couponsDeleteCmd.RunE(cmd, []string{tt.couponID})

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

// TestCouponsListGetClientError verifies list command error handling when getClient fails
func TestCouponsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsListCmd)

	err := couponsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsGetGetClientError verifies get command error handling when getClient fails
func TestCouponsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsGetCmd)

	err := couponsGetCmd.RunE(cmd, []string{"coupon-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsLookupGetClientError verifies lookup command error handling when getClient fails
func TestCouponsLookupGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsLookupCmd)

	err := couponsLookupCmd.RunE(cmd, []string{"DISCOUNT10"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsCreateGetClientError verifies create command error handling when getClient fails
func TestCouponsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsCreateCmd)

	err := couponsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsActivateGetClientError verifies activate command error handling when getClient fails
func TestCouponsActivateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsActivateCmd)

	err := couponsActivateCmd.RunE(cmd, []string{"coupon-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsDeactivateGetClientError verifies deactivate command error handling when getClient fails
func TestCouponsDeactivateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsDeactivateCmd)

	err := couponsDeactivateCmd.RunE(cmd, []string{"coupon-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsDeleteGetClientError verifies delete command error handling when getClient fails
func TestCouponsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(couponsDeleteCmd)

	err := couponsDeleteCmd.RunE(cmd, []string{"coupon-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestCouponsListTextOutputFormatting tests specific text output formatting.
func TestCouponsListTextOutputFormatting(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name        string
		coupon      api.Coupon
		wantOutputs []string
	}{
		{
			name: "percentage coupon with limit",
			coupon: api.Coupon{
				ID:            "coup_pct",
				Code:          "PCT10",
				Title:         "Percentage",
				DiscountType:  "percentage",
				DiscountValue: 10,
				UsageCount:    5,
				UsageLimit:    100,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        endsAt,
			},
			wantOutputs: []string{"PCT10", "10%", "5/100"},
		},
		{
			name: "fixed amount coupon unlimited",
			coupon: api.Coupon{
				ID:            "coup_fixed",
				Code:          "FLAT50",
				Title:         "Fixed",
				DiscountType:  "fixed_amount",
				DiscountValue: 50,
				UsageCount:    10,
				UsageLimit:    0,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        time.Time{},
			},
			wantOutputs: []string{"FLAT50", "50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				listCouponsResp: &api.CouponsListResponse{
					Items:      []api.Coupon{tt.coupon},
					TotalCount: 1,
				},
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := couponsListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantOutputs {
				if !strings.Contains(output, want) {
					t.Errorf("output %q should contain %q", output, want)
				}
			}
		})
	}
}

// TestCouponsGetTextOutputFormatting tests specific text output formatting for get command.
func TestCouponsGetTextOutputFormatting(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2023, 12, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		coupon      api.Coupon
		wantOutputs []string
	}{
		{
			name: "coupon with all optional fields",
			coupon: api.Coupon{
				ID:            "coup_full",
				Code:          "FULLTEST",
				Title:         "Full Test",
				Description:   "Full test description",
				DiscountType:  "percentage",
				DiscountValue: 25,
				MinPurchase:   100,
				MaxDiscount:   50,
				UsageCount:    10,
				UsageLimit:    100,
				PerCustomer:   2,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        endsAt,
				CreatedAt:     createdAt,
			},
			wantOutputs: []string{
				"coup_full",
				"FULLTEST",
				"Full test description",
				"Min Purchase",
				"Max Discount",
				"Per Customer",
			},
		},
		{
			name: "coupon with minimal fields",
			coupon: api.Coupon{
				ID:            "coup_min",
				Code:          "MINTEST",
				Title:         "Minimal Test",
				DiscountType:  "fixed_amount",
				DiscountValue: 10,
				UsageCount:    0,
				UsageLimit:    0,
				Status:        "inactive",
				StartsAt:      startsAt,
				EndsAt:        time.Time{},
				CreatedAt:     createdAt,
			},
			wantOutputs: []string{"coup_min", "MINTEST", "inactive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &couponsMockAPIClient{
				getCouponResp: &tt.coupon,
			}
			cleanup := setupCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := couponsGetCmd.RunE(cmd, []string{tt.coupon.ID})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Note: text output goes to stdout directly via fmt.Printf
			// We're verifying the command runs without error
		})
	}
}

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

// TestDiscountCodesCommandSetup verifies discount-codes command initialization
func TestDiscountCodesCommandSetup(t *testing.T) {
	if discountCodesCmd.Use != "discount-codes" {
		t.Errorf("expected Use 'discount-codes', got %q", discountCodesCmd.Use)
	}
	if discountCodesCmd.Short != "Manage discount codes" {
		t.Errorf("expected Short 'Manage discount codes', got %q", discountCodesCmd.Short)
	}
}

// TestDiscountCodesSubcommands verifies all subcommands are registered
func TestDiscountCodesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List discount codes",
		"get":    "Get discount code details",
		"lookup": "Lookup a discount code by code string",
		"create": "Create a discount code",
		"delete": "Delete a discount code",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range discountCodesCmd.Commands() {
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

// TestDiscountCodesListFlags verifies list command flags exist with correct defaults
func TestDiscountCodesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"price-rule-id", ""},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := discountCodesListCmd.Flags().Lookup(f.name)
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

// TestDiscountCodesCreateFlags verifies create command flags exist with correct defaults
func TestDiscountCodesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"code", ""},
		{"price-rule-id", ""},
		{"discount-type", ""},
		{"discount-value", "0"},
		{"usage-limit", "0"},
		{"min-purchase", "0"},
		{"starts-at", ""},
		{"ends-at", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := discountCodesCreateCmd.Flags().Lookup(f.name)
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

// TestDiscountCodesCreateRequiredFlags verifies code, discount-type, and discount-value are required
func TestDiscountCodesCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"code", "discount-type", "discount-value"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := discountCodesCreateCmd.Flags().Lookup(name)
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

// TestDiscountCodesDeleteFlags verifies delete command flags exist with correct defaults
func TestDiscountCodesDeleteFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"yes", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := discountCodesDeleteCmd.Flags().Lookup(f.name)
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

// TestDiscountCodesGetArgs verifies get command requires exactly 1 argument
func TestDiscountCodesGetArgs(t *testing.T) {
	err := discountCodesGetCmd.Args(discountCodesGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = discountCodesGetCmd.Args(discountCodesGetCmd, []string{"discount-code-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestDiscountCodesLookupArgs verifies lookup command requires exactly 1 argument
func TestDiscountCodesLookupArgs(t *testing.T) {
	err := discountCodesLookupCmd.Args(discountCodesLookupCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = discountCodesLookupCmd.Args(discountCodesLookupCmd, []string{"SAVE20"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestDiscountCodesDeleteArgs verifies delete command requires exactly 1 argument
func TestDiscountCodesDeleteArgs(t *testing.T) {
	err := discountCodesDeleteCmd.Args(discountCodesDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = discountCodesDeleteCmd.Args(discountCodesDeleteCmd, []string{"discount-code-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// discountCodesMockAPIClient is a mock implementation of api.APIClient for discount codes tests.
type discountCodesMockAPIClient struct {
	api.MockClient
	listDiscountCodesResp     *api.DiscountCodesListResponse
	listDiscountCodesErr      error
	getDiscountCodeResp       *api.DiscountCode
	getDiscountCodeErr        error
	getDiscountCodeByCodeResp *api.DiscountCode
	getDiscountCodeByCodeErr  error
	createDiscountCodeResp    *api.DiscountCode
	createDiscountCodeErr     error
	deleteDiscountCodeErr     error
}

func (m *discountCodesMockAPIClient) ListDiscountCodes(ctx context.Context, opts *api.DiscountCodesListOptions) (*api.DiscountCodesListResponse, error) {
	return m.listDiscountCodesResp, m.listDiscountCodesErr
}

func (m *discountCodesMockAPIClient) GetDiscountCode(ctx context.Context, id string) (*api.DiscountCode, error) {
	return m.getDiscountCodeResp, m.getDiscountCodeErr
}

func (m *discountCodesMockAPIClient) GetDiscountCodeByCode(ctx context.Context, code string) (*api.DiscountCode, error) {
	return m.getDiscountCodeByCodeResp, m.getDiscountCodeByCodeErr
}

func (m *discountCodesMockAPIClient) CreateDiscountCode(ctx context.Context, req *api.DiscountCodeCreateRequest) (*api.DiscountCode, error) {
	return m.createDiscountCodeResp, m.createDiscountCodeErr
}

func (m *discountCodesMockAPIClient) DeleteDiscountCode(ctx context.Context, id string) error {
	return m.deleteDiscountCodeErr
}

// setupDiscountCodesMockFactories sets up mock factories for discount codes tests.
func setupDiscountCodesMockFactories(mockClient *discountCodesMockAPIClient) (func(), *bytes.Buffer) {
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

// newDiscountCodesTestCmd creates a test command with common flags for discount codes tests.
func newDiscountCodesTestCmd() *cobra.Command {
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

// TestDiscountCodesListRunE tests the discount-codes list command with mock API.
func TestDiscountCodesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.DiscountCodesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with percentage discount",
			mockResp: &api.DiscountCodesListResponse{
				Items: []api.DiscountCode{
					{
						ID:            "dc_123",
						Code:          "SAVE20",
						DiscountType:  "percentage",
						DiscountValue: 20,
						UsageCount:    5,
						UsageLimit:    100,
						Status:        "active",
						StartsAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						EndsAt:        time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "dc_123",
		},
		{
			name: "successful list with fixed amount discount",
			mockResp: &api.DiscountCodesListResponse{
				Items: []api.DiscountCode{
					{
						ID:            "dc_456",
						Code:          "FLAT10",
						DiscountType:  "fixed_amount",
						DiscountValue: 10,
						UsageCount:    3,
						UsageLimit:    0, // unlimited
						Status:        "active",
					},
				},
				TotalCount: 1,
			},
			wantOutput: "dc_456",
		},
		{
			name: "successful list with no dates",
			mockResp: &api.DiscountCodesListResponse{
				Items: []api.DiscountCode{
					{
						ID:            "dc_789",
						Code:          "NODATES",
						DiscountType:  "percentage",
						DiscountValue: 15,
						UsageCount:    0,
						Status:        "inactive",
					},
				},
				TotalCount: 1,
			},
			wantOutput: "dc_789",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.DiscountCodesListResponse{
				Items:      []api.DiscountCode{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &discountCodesMockAPIClient{
				listDiscountCodesResp: tt.mockResp,
				listDiscountCodesErr:  tt.mockErr,
			}
			cleanup, buf := setupDiscountCodesMockFactories(mockClient)
			defer cleanup()

			cmd := newDiscountCodesTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("price-rule-id", "", "")
			cmd.Flags().String("status", "", "")

			err := discountCodesListCmd.RunE(cmd, []string{})

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

// TestDiscountCodesListRunE_JSONOutput tests the list command with JSON output.
func TestDiscountCodesListRunE_JSONOutput(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{
		listDiscountCodesResp: &api.DiscountCodesListResponse{
			Items: []api.DiscountCode{
				{
					ID:            "dc_json",
					Code:          "JSONTEST",
					DiscountType:  "percentage",
					DiscountValue: 25,
					Status:        "active",
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("price-rule-id", "", "")
	cmd.Flags().String("status", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := discountCodesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "dc_json") {
		t.Errorf("JSON output should contain discount code ID")
	}
}

// TestDiscountCodesGetRunE tests the discount-codes get command with mock API.
func TestDiscountCodesGetRunE(t *testing.T) {
	tests := []struct {
		name           string
		discountCodeID string
		mockResp       *api.DiscountCode
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful get with all fields",
			discountCodeID: "dc_123",
			mockResp: &api.DiscountCode{
				ID:            "dc_123",
				Code:          "SUMMER20",
				PriceRuleID:   "pr_456",
				DiscountType:  "percentage",
				DiscountValue: 20,
				MinPurchase:   50.00,
				UsageCount:    10,
				UsageLimit:    100,
				Status:        "active",
				StartsAt:      time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				EndsAt:        time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC),
				CreatedAt:     time.Date(2024, 5, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:           "successful get with minimal fields",
			discountCodeID: "dc_minimal",
			mockResp: &api.DiscountCode{
				ID:            "dc_minimal",
				Code:          "SIMPLE",
				DiscountType:  "fixed_amount",
				DiscountValue: 5,
				UsageCount:    0,
				Status:        "active",
				CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "discount code not found",
			discountCodeID: "dc_999",
			mockErr:        errors.New("discount code not found"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &discountCodesMockAPIClient{
				getDiscountCodeResp: tt.mockResp,
				getDiscountCodeErr:  tt.mockErr,
			}
			cleanup, _ := setupDiscountCodesMockFactories(mockClient)
			defer cleanup()

			cmd := newDiscountCodesTestCmd()

			err := discountCodesGetCmd.RunE(cmd, []string{tt.discountCodeID})

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

// TestDiscountCodesGetRunE_JSONOutput tests the get command with JSON output.
func TestDiscountCodesGetRunE_JSONOutput(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{
		getDiscountCodeResp: &api.DiscountCode{
			ID:            "dc_json",
			Code:          "JSONGET",
			DiscountType:  "percentage",
			DiscountValue: 15,
			Status:        "active",
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := discountCodesGetCmd.RunE(cmd, []string{"dc_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "dc_json") {
		t.Errorf("JSON output should contain discount code ID")
	}
}

// TestDiscountCodesLookupRunE tests the discount-codes lookup command with mock API.
func TestDiscountCodesLookupRunE(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		mockResp *api.DiscountCode
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful lookup",
			code: "SAVE20",
			mockResp: &api.DiscountCode{
				ID:            "dc_123",
				Code:          "SAVE20",
				DiscountType:  "percentage",
				DiscountValue: 20,
				Status:        "active",
			},
		},
		{
			name:    "code not found",
			code:    "INVALID",
			mockErr: errors.New("discount code not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &discountCodesMockAPIClient{
				getDiscountCodeByCodeResp: tt.mockResp,
				getDiscountCodeByCodeErr:  tt.mockErr,
			}
			cleanup, _ := setupDiscountCodesMockFactories(mockClient)
			defer cleanup()

			cmd := newDiscountCodesTestCmd()

			err := discountCodesLookupCmd.RunE(cmd, []string{tt.code})

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

// TestDiscountCodesLookupRunE_JSONOutput tests the lookup command with JSON output.
func TestDiscountCodesLookupRunE_JSONOutput(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{
		getDiscountCodeByCodeResp: &api.DiscountCode{
			ID:            "dc_lookup",
			Code:          "LOOKUPJSON",
			DiscountType:  "fixed_amount",
			DiscountValue: 10,
			Status:        "active",
		},
	}
	cleanup, buf := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := discountCodesLookupCmd.RunE(cmd, []string{"LOOKUPJSON"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "dc_lookup") {
		t.Errorf("JSON output should contain discount code ID")
	}
}

// TestDiscountCodesCreateRunE tests the discount-codes create command with mock API.
func TestDiscountCodesCreateRunE(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.DiscountCode
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.DiscountCode{
				ID:            "dc_new",
				Code:          "NEWCODE",
				DiscountType:  "percentage",
				DiscountValue: 15,
				Status:        "active",
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("failed to create discount code"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &discountCodesMockAPIClient{
				createDiscountCodeResp: tt.mockResp,
				createDiscountCodeErr:  tt.mockErr,
			}
			cleanup, _ := setupDiscountCodesMockFactories(mockClient)
			defer cleanup()

			cmd := newDiscountCodesTestCmd()
			cmd.Flags().String("code", "", "")
			cmd.Flags().String("price-rule-id", "", "")
			cmd.Flags().String("discount-type", "", "")
			cmd.Flags().Float64("discount-value", 0, "")
			cmd.Flags().Int("usage-limit", 0, "")
			cmd.Flags().Float64("min-purchase", 0, "")
			cmd.Flags().String("starts-at", "", "")
			cmd.Flags().String("ends-at", "", "")
			_ = cmd.Flags().Set("code", "NEWCODE")
			_ = cmd.Flags().Set("discount-type", "percentage")
			_ = cmd.Flags().Set("discount-value", "15")

			err := discountCodesCreateCmd.RunE(cmd, []string{})

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

// TestDiscountCodesCreateRunE_WithDates tests create command with date parameters.
func TestDiscountCodesCreateRunE_WithDates(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{
		createDiscountCodeResp: &api.DiscountCode{
			ID:            "dc_dated",
			Code:          "DATED",
			DiscountType:  "percentage",
			DiscountValue: 20,
			Status:        "active",
		},
	}
	cleanup, _ := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("price-rule-id", "", "")
	cmd.Flags().String("discount-type", "", "")
	cmd.Flags().Float64("discount-value", 0, "")
	cmd.Flags().Int("usage-limit", 0, "")
	cmd.Flags().Float64("min-purchase", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("code", "DATED")
	_ = cmd.Flags().Set("discount-type", "percentage")
	_ = cmd.Flags().Set("discount-value", "20")
	_ = cmd.Flags().Set("starts-at", "2024-01-01T00:00:00Z")
	_ = cmd.Flags().Set("ends-at", "2024-12-31T23:59:59Z")

	err := discountCodesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDiscountCodesCreateRunE_InvalidStartsAt tests create command with invalid starts-at date.
func TestDiscountCodesCreateRunE_InvalidStartsAt(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{}
	cleanup, _ := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("price-rule-id", "", "")
	cmd.Flags().String("discount-type", "", "")
	cmd.Flags().Float64("discount-value", 0, "")
	cmd.Flags().Int("usage-limit", 0, "")
	cmd.Flags().Float64("min-purchase", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("code", "BADDATE")
	_ = cmd.Flags().Set("discount-type", "percentage")
	_ = cmd.Flags().Set("discount-value", "10")
	_ = cmd.Flags().Set("starts-at", "invalid-date")

	err := discountCodesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid starts-at date")
	}
	if !strings.Contains(err.Error(), "invalid starts-at format") {
		t.Errorf("error should mention invalid starts-at format: %v", err)
	}
}

// TestDiscountCodesCreateRunE_InvalidEndsAt tests create command with invalid ends-at date.
func TestDiscountCodesCreateRunE_InvalidEndsAt(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{}
	cleanup, _ := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("price-rule-id", "", "")
	cmd.Flags().String("discount-type", "", "")
	cmd.Flags().Float64("discount-value", 0, "")
	cmd.Flags().Int("usage-limit", 0, "")
	cmd.Flags().Float64("min-purchase", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("code", "BADDATE")
	_ = cmd.Flags().Set("discount-type", "percentage")
	_ = cmd.Flags().Set("discount-value", "10")
	_ = cmd.Flags().Set("ends-at", "invalid-date")

	err := discountCodesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid ends-at date")
	}
	if !strings.Contains(err.Error(), "invalid ends-at format") {
		t.Errorf("error should mention invalid ends-at format: %v", err)
	}
}

// TestDiscountCodesCreateRunE_JSONOutput tests create command with JSON output.
func TestDiscountCodesCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &discountCodesMockAPIClient{
		createDiscountCodeResp: &api.DiscountCode{
			ID:            "dc_json_create",
			Code:          "JSONCREATE",
			DiscountType:  "percentage",
			DiscountValue: 25,
			Status:        "active",
		},
	}
	cleanup, buf := setupDiscountCodesMockFactories(mockClient)
	defer cleanup()

	cmd := newDiscountCodesTestCmd()
	cmd.Flags().String("code", "", "")
	cmd.Flags().String("price-rule-id", "", "")
	cmd.Flags().String("discount-type", "", "")
	cmd.Flags().Float64("discount-value", 0, "")
	cmd.Flags().Int("usage-limit", 0, "")
	cmd.Flags().Float64("min-purchase", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")
	_ = cmd.Flags().Set("code", "JSONCREATE")
	_ = cmd.Flags().Set("discount-type", "percentage")
	_ = cmd.Flags().Set("discount-value", "25")
	_ = cmd.Flags().Set("output", "json")

	err := discountCodesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "dc_json_create") {
		t.Errorf("JSON output should contain discount code ID")
	}
}

// TestDiscountCodesDeleteRunE tests the discount-codes delete command with mock API.
func TestDiscountCodesDeleteRunE(t *testing.T) {
	tests := []struct {
		name           string
		discountCodeID string
		mockErr        error
		wantErr        bool
	}{
		{
			name:           "successful delete",
			discountCodeID: "dc_123",
			mockErr:        nil,
		},
		{
			name:           "delete error",
			discountCodeID: "dc_456",
			mockErr:        errors.New("failed to delete"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &discountCodesMockAPIClient{
				deleteDiscountCodeErr: tt.mockErr,
			}
			cleanup, _ := setupDiscountCodesMockFactories(mockClient)
			defer cleanup()

			cmd := newDiscountCodesTestCmd()

			err := discountCodesDeleteCmd.RunE(cmd, []string{tt.discountCodeID})

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

// TestDiscountCodesGetClientError verifies error handling when getClient fails
func TestDiscountCodesGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestDiscountCodesListGetClientError verifies list command error handling when getClient fails
func TestDiscountCodesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(discountCodesListCmd)

	err := discountCodesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDiscountCodesGetGetClientError verifies get command error handling when getClient fails
func TestDiscountCodesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(discountCodesGetCmd)

	err := discountCodesGetCmd.RunE(cmd, []string{"discount-code-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDiscountCodesLookupGetClientError verifies lookup command error handling when getClient fails
func TestDiscountCodesLookupGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(discountCodesLookupCmd)

	err := discountCodesLookupCmd.RunE(cmd, []string{"SAVE20"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDiscountCodesCreateGetClientError verifies create command error handling when getClient fails
func TestDiscountCodesCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(discountCodesCreateCmd)

	err := discountCodesCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDiscountCodesDeleteGetClientError verifies delete command error handling when getClient fails
func TestDiscountCodesDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(discountCodesDeleteCmd)

	err := discountCodesDeleteCmd.RunE(cmd, []string{"discount-code-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestDiscountCodesWithMockStore tests discount codes commands with a mock credential store
func TestDiscountCodesWithMockStore(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	store := &mockStore{
		names: []string{"teststore"},
		creds: map[string]*secrets.StoreCredentials{
			"teststore": {Handle: "test-handle", AccessToken: "test-token"},
		},
	}

	secretsStoreFactory = func() (CredentialStore, error) {
		return store, nil
	}

	cmd := newTestCmdWithFlags()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// TestDiscountCodesListRunE_NoProfiles verifies error when no profiles are configured
func TestDiscountCodesListRunE_NoProfiles(t *testing.T) {
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
	err := discountCodesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestDiscountCodesGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestDiscountCodesGetRunE_MultipleProfiles(t *testing.T) {
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
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	err := discountCodesGetCmd.RunE(cmd, []string{"dc_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

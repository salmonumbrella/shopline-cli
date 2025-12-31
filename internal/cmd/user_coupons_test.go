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

// userCouponsMockAPIClient is a mock implementation of api.APIClient for user coupons testing.
type userCouponsMockAPIClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values for specific methods
	listUserCouponsResp *api.UserCouponsListResponse
	listUserCouponsErr  error

	getUserCouponResp *api.UserCoupon
	getUserCouponErr  error

	assignUserCouponResp *api.UserCoupon
	assignUserCouponErr  error

	revokeUserCouponErr error
}

func (m *userCouponsMockAPIClient) ListUserCoupons(ctx context.Context, opts *api.UserCouponsListOptions) (*api.UserCouponsListResponse, error) {
	return m.listUserCouponsResp, m.listUserCouponsErr
}

func (m *userCouponsMockAPIClient) GetUserCoupon(ctx context.Context, id string) (*api.UserCoupon, error) {
	return m.getUserCouponResp, m.getUserCouponErr
}

func (m *userCouponsMockAPIClient) AssignUserCoupon(ctx context.Context, req *api.UserCouponAssignRequest) (*api.UserCoupon, error) {
	return m.assignUserCouponResp, m.assignUserCouponErr
}

func (m *userCouponsMockAPIClient) RevokeUserCoupon(ctx context.Context, id string) error {
	return m.revokeUserCouponErr
}

// setupUserCouponsMockFactories configures mock factories for user coupons testing.
func setupUserCouponsMockFactories(mockClient *userCouponsMockAPIClient) (cleanup func()) {
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

// TestUserCouponsCommandSetup verifies user-coupons command initialization
func TestUserCouponsCommandSetup(t *testing.T) {
	if userCouponsCmd.Use != "user-coupons" {
		t.Errorf("expected Use 'user-coupons', got %q", userCouponsCmd.Use)
	}
	if userCouponsCmd.Short != "Manage user-assigned coupons" {
		t.Errorf("expected Short 'Manage user-assigned coupons', got %q", userCouponsCmd.Short)
	}
}

// TestUserCouponsSubcommands verifies all subcommands are registered
func TestUserCouponsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List user coupons",
		"get":    "Get user coupon details",
		"assign": "Assign a coupon to a user",
		"revoke": "Revoke a user's coupon",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range userCouponsCmd.Commands() {
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

// TestUserCouponsListFlags verifies list command flags exist with correct defaults
func TestUserCouponsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"user-id", ""},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := userCouponsListCmd.Flags().Lookup(f.name)
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

// TestUserCouponsAssignFlags verifies assign command flags exist with correct defaults
func TestUserCouponsAssignFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"user-id", ""},
		{"coupon-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := userCouponsAssignCmd.Flags().Lookup(f.name)
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

// TestUserCouponsRevokeFlags verifies revoke command flags exist with correct defaults
func TestUserCouponsRevokeFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"yes", "false"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := userCouponsRevokeCmd.Flags().Lookup(f.name)
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

// TestUserCouponsGetArgs verifies get command requires exactly one argument
func TestUserCouponsGetArgs(t *testing.T) {
	err := userCouponsGetCmd.Args(userCouponsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = userCouponsGetCmd.Args(userCouponsGetCmd, []string{"uc_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestUserCouponsRevokeArgs verifies revoke command requires exactly one argument
func TestUserCouponsRevokeArgs(t *testing.T) {
	err := userCouponsRevokeCmd.Args(userCouponsRevokeCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = userCouponsRevokeCmd.Args(userCouponsRevokeCmd, []string{"uc_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestUserCouponsListFlagDescriptions verifies flag descriptions are set
func TestUserCouponsListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"page":      "Page number",
		"page-size": "Results per page",
		"user-id":   "Filter by user ID",
		"status":    "Filter by status (active, used, expired, revoked)",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := userCouponsListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// TestUserCouponsAssignFlagDescriptions verifies assign flag descriptions
func TestUserCouponsAssignFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"user-id":   "User ID (required)",
		"coupon-id": "Coupon ID (required)",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := userCouponsAssignCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Usage != expectedUsage {
				t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
			}
		})
	}
}

// TestUserCouponsRevokeFlagDescriptions verifies revoke flag descriptions
func TestUserCouponsRevokeFlagDescriptions(t *testing.T) {
	flag := userCouponsRevokeCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("yes flag not found")
		return
	}
	expectedUsage := "Skip confirmation prompt"
	if flag.Usage != expectedUsage {
		t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
	}
}

// TestUserCouponsAssignRequiredFlags verifies that user-id and coupon-id flags are required
func TestUserCouponsAssignRequiredFlags(t *testing.T) {
	requiredFlags := []string{"user-id", "coupon-id"}

	for _, flagName := range requiredFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := userCouponsAssignCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("%s flag not found", flagName)
				return
			}
			// Verify it exists and is a string flag
			if flag.Value.Type() != "string" {
				t.Errorf("%s flag should be a string type", flagName)
			}
			// Check if the flag has required annotation
			annotations := flag.Annotations
			if annotations == nil {
				t.Errorf("flag %q has no annotations, expected required", flagName)
				return
			}
			if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
				t.Errorf("flag %q is not marked as required", flagName)
			}
		})
	}
}

// TestUserCouponsListFlagTypes verifies flag types are correct
func TestUserCouponsListFlagTypes(t *testing.T) {
	flags := map[string]string{
		"page":      "int",
		"page-size": "int",
		"user-id":   "string",
		"status":    "string",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := userCouponsListCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("flag %q not found", flagName)
				return
			}
			if flag.Value.Type() != expectedType {
				t.Errorf("expected type %q, got %q", expectedType, flag.Value.Type())
			}
		})
	}
}

// TestUserCouponsListRunE tests the user-coupons list command execution with mock API.
func TestUserCouponsListRunE(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.UserCouponsListResponse
		mockErr    error
		output     string
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful list with percentage coupon",
			output: "text",
			mockResp: &api.UserCouponsListResponse{
				Items: []api.UserCoupon{
					{
						ID:            "uc_123",
						UserID:        "user_456",
						CouponID:      "coup_789",
						CouponCode:    "SAVE10",
						Title:         "10% Off",
						DiscountType:  "percentage",
						DiscountValue: 10,
						Status:        "active",
						ExpiresAt:     expiresAt,
						CreatedAt:     createdAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "uc_123",
		},
		{
			name:   "successful list with fixed_amount coupon no expiry",
			output: "text",
			mockResp: &api.UserCouponsListResponse{
				Items: []api.UserCoupon{
					{
						ID:            "uc_456",
						UserID:        "user_789",
						CouponID:      "coup_101",
						CouponCode:    "FLAT20",
						Title:         "$20 Off",
						DiscountType:  "fixed_amount",
						DiscountValue: 20,
						Status:        "active",
						ExpiresAt:     time.Time{}, // no expiry
						CreatedAt:     createdAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "FLAT20",
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.UserCouponsListResponse{
				Items: []api.UserCoupon{
					{
						ID:            "uc_789",
						UserID:        "user_123",
						CouponID:      "coup_456",
						CouponCode:    "WELCOME",
						Title:         "Welcome Discount",
						DiscountType:  "percentage",
						DiscountValue: 15,
						Status:        "active",
						ExpiresAt:     expiresAt,
						CreatedAt:     createdAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "uc_789",
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
			mockResp: &api.UserCouponsListResponse{
				Items:      []api.UserCoupon{},
				TotalCount: 0,
			},
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				listUserCouponsResp: tt.mockResp,
				listUserCouponsErr:  tt.mockErr,
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("user-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := userCouponsListCmd.RunE(cmd, []string{})

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

// TestUserCouponsGetRunE tests the user-coupons get command execution with mock API.
func TestUserCouponsGetRunE(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	usedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		userCouponID string
		output       string
		mockResp     *api.UserCoupon
		mockErr      error
		wantErr      bool
		wantOutput   string // Only checked for JSON output
	}{
		{
			name:         "successful get with all fields",
			userCouponID: "uc_123",
			output:       "text",
			mockResp: &api.UserCoupon{
				ID:            "uc_123",
				UserID:        "user_456",
				CouponID:      "coup_789",
				CouponCode:    "SAVE10",
				Title:         "10% Off",
				DiscountType:  "percentage",
				DiscountValue: 10,
				Status:        "used",
				UsedAt:        usedAt,
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "", // Text output goes to stdout
		},
		{
			name:         "successful get with no used_at and no expires_at",
			userCouponID: "uc_456",
			output:       "text",
			mockResp: &api.UserCoupon{
				ID:            "uc_456",
				UserID:        "user_789",
				CouponID:      "coup_101",
				CouponCode:    "FLAT20",
				Title:         "$20 Off",
				DiscountType:  "fixed_amount",
				DiscountValue: 20,
				Status:        "active",
				UsedAt:        time.Time{}, // not used
				ExpiresAt:     time.Time{}, // no expiry
				CreatedAt:     createdAt,
			},
			wantOutput: "", // Text output goes to stdout
		},
		{
			name:         "successful get JSON output",
			userCouponID: "uc_789",
			output:       "json",
			mockResp: &api.UserCoupon{
				ID:            "uc_789",
				UserID:        "user_123",
				CouponID:      "coup_456",
				CouponCode:    "WELCOME",
				Title:         "Welcome Discount",
				DiscountType:  "percentage",
				DiscountValue: 15,
				Status:        "active",
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "uc_789",
		},
		{
			name:         "user coupon not found",
			userCouponID: "uc_999",
			output:       "text",
			mockErr:      errors.New("user coupon not found"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				getUserCouponResp: tt.mockResp,
				getUserCouponErr:  tt.mockErr,
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := userCouponsGetCmd.RunE(cmd, []string{tt.userCouponID})

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

// TestUserCouponsAssignRunE tests the user-coupons assign command execution with mock API.
func TestUserCouponsAssignRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		output     string
		mockResp   *api.UserCoupon
		mockErr    error
		wantErr    bool
		wantOutput string // Only checked for JSON output
	}{
		{
			name:   "successful assign",
			output: "text",
			mockResp: &api.UserCoupon{
				ID:            "uc_new",
				UserID:        "user_123",
				CouponID:      "coup_456",
				CouponCode:    "NEWCODE",
				Title:         "New Coupon",
				DiscountType:  "percentage",
				DiscountValue: 20,
				Status:        "active",
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "", // Text output goes to stdout
		},
		{
			name:   "successful assign JSON output",
			output: "json",
			mockResp: &api.UserCoupon{
				ID:            "uc_json",
				UserID:        "user_456",
				CouponID:      "coup_789",
				CouponCode:    "JSONCODE",
				Title:         "JSON Coupon",
				DiscountType:  "fixed_amount",
				DiscountValue: 50,
				Status:        "active",
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "uc_json",
		},
		{
			name:    "assign fails - user not found",
			output:  "text",
			mockErr: errors.New("user not found"),
			wantErr: true,
		},
		{
			name:    "assign fails - coupon not found",
			output:  "text",
			mockErr: errors.New("coupon not found"),
			wantErr: true,
		},
		{
			name:    "assign fails - coupon already assigned",
			output:  "text",
			mockErr: errors.New("coupon already assigned to user"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				assignUserCouponResp: tt.mockResp,
				assignUserCouponErr:  tt.mockErr,
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.output, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("user-id", "user_123", "")
			cmd.Flags().String("coupon-id", "coup_456", "")

			err := userCouponsAssignCmd.RunE(cmd, []string{})

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

// TestUserCouponsRevokeRunE tests the user-coupons revoke command execution with mock API.
func TestUserCouponsRevokeRunE(t *testing.T) {
	tests := []struct {
		name         string
		userCouponID string
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful revoke",
			userCouponID: "uc_123",
			mockErr:      nil,
		},
		{
			name:         "revoke fails - not found",
			userCouponID: "uc_456",
			mockErr:      errors.New("user coupon not found"),
			wantErr:      true,
		},
		{
			name:         "revoke fails - already used",
			userCouponID: "uc_789",
			mockErr:      errors.New("cannot revoke used coupon"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				revokeUserCouponErr: tt.mockErr,
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := userCouponsRevokeCmd.RunE(cmd, []string{tt.userCouponID})

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

// TestUserCouponsListGetClientError verifies list command error handling when getClient fails
func TestUserCouponsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(userCouponsListCmd)

	err := userCouponsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestUserCouponsGetGetClientError verifies get command error handling when getClient fails
func TestUserCouponsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(userCouponsGetCmd)

	err := userCouponsGetCmd.RunE(cmd, []string{"uc_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestUserCouponsAssignGetClientError verifies assign command error handling when getClient fails
func TestUserCouponsAssignGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(userCouponsAssignCmd)

	err := userCouponsAssignCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestUserCouponsRevokeGetClientError verifies revoke command error handling when getClient fails
func TestUserCouponsRevokeGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(userCouponsRevokeCmd)

	err := userCouponsRevokeCmd.RunE(cmd, []string{"uc_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestUserCouponsListTextOutputFormatting tests specific text output formatting.
func TestUserCouponsListTextOutputFormatting(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		userCoupon  api.UserCoupon
		wantOutputs []string
	}{
		{
			name: "percentage coupon with expiry",
			userCoupon: api.UserCoupon{
				ID:            "uc_pct",
				UserID:        "user_123",
				CouponID:      "coup_456",
				CouponCode:    "PCT10",
				Title:         "Percentage",
				DiscountType:  "percentage",
				DiscountValue: 10,
				Status:        "active",
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
			wantOutputs: []string{"PCT10", "10%", "2024-12-31"},
		},
		{
			name: "fixed amount coupon no expiry",
			userCoupon: api.UserCoupon{
				ID:            "uc_fixed",
				UserID:        "user_456",
				CouponID:      "coup_789",
				CouponCode:    "FLAT50",
				Title:         "Fixed",
				DiscountType:  "fixed_amount",
				DiscountValue: 50,
				Status:        "active",
				ExpiresAt:     time.Time{},
				CreatedAt:     createdAt,
			},
			wantOutputs: []string{"FLAT50", "50"},
		},
		{
			name: "used coupon",
			userCoupon: api.UserCoupon{
				ID:            "uc_used",
				UserID:        "user_789",
				CouponID:      "coup_101",
				CouponCode:    "USED20",
				Title:         "Used Coupon",
				DiscountType:  "percentage",
				DiscountValue: 20,
				Status:        "used",
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
			wantOutputs: []string{"USED20", "used", "20%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				listUserCouponsResp: &api.UserCouponsListResponse{
					Items:      []api.UserCoupon{tt.userCoupon},
					TotalCount: 1,
				},
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("user-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := userCouponsListCmd.RunE(cmd, []string{})
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

// TestUserCouponsGetTextOutputFormatting tests specific text output formatting for get command.
func TestUserCouponsGetTextOutputFormatting(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	usedAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		userCoupon api.UserCoupon
	}{
		{
			name: "coupon with all optional fields including used_at",
			userCoupon: api.UserCoupon{
				ID:            "uc_full",
				UserID:        "user_full",
				CouponID:      "coup_full",
				CouponCode:    "FULLTEST",
				Title:         "Full Test",
				DiscountType:  "percentage",
				DiscountValue: 25,
				Status:        "used",
				UsedAt:        usedAt,
				ExpiresAt:     expiresAt,
				CreatedAt:     createdAt,
			},
		},
		{
			name: "coupon with minimal fields (no used_at, no expires_at)",
			userCoupon: api.UserCoupon{
				ID:            "uc_min",
				UserID:        "user_min",
				CouponID:      "coup_min",
				CouponCode:    "MINTEST",
				Title:         "Minimal Test",
				DiscountType:  "fixed_amount",
				DiscountValue: 10,
				Status:        "active",
				UsedAt:        time.Time{},
				ExpiresAt:     time.Time{},
				CreatedAt:     createdAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				getUserCouponResp: &tt.userCoupon,
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := userCouponsGetCmd.RunE(cmd, []string{tt.userCoupon.ID})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Note: text output goes to stdout directly via fmt.Printf
			// We're verifying the command runs without error
		})
	}
}

// TestUserCouponsListWithFilters tests that list command respects filter flags.
func TestUserCouponsListWithFilters(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		userID   string
		status   string
		page     int
		pageSize int
	}{
		{
			name:     "filter by user ID",
			userID:   "user_specific",
			status:   "",
			page:     1,
			pageSize: 20,
		},
		{
			name:     "filter by status",
			userID:   "",
			status:   "active",
			page:     1,
			pageSize: 20,
		},
		{
			name:     "filter by both user ID and status",
			userID:   "user_123",
			status:   "used",
			page:     1,
			pageSize: 20,
		},
		{
			name:     "custom pagination",
			userID:   "",
			status:   "",
			page:     2,
			pageSize: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &userCouponsMockAPIClient{
				listUserCouponsResp: &api.UserCouponsListResponse{
					Items: []api.UserCoupon{
						{
							ID:            "uc_test",
							UserID:        "user_test",
							CouponCode:    "TEST",
							Title:         "Test",
							DiscountType:  "percentage",
							DiscountValue: 10,
							Status:        "active",
							CreatedAt:     createdAt,
						},
					},
					TotalCount: 1,
				},
			}
			cleanup := setupUserCouponsMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "text", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("user-id", tt.userID, "")
			cmd.Flags().String("status", tt.status, "")
			cmd.Flags().Int("page", tt.page, "")
			cmd.Flags().Int("page-size", tt.pageSize, "")

			err := userCouponsListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestUserCouponsMultipleCouponsInList tests list command with multiple coupons.
func TestUserCouponsMultipleCouponsInList(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	mockClient := &userCouponsMockAPIClient{
		listUserCouponsResp: &api.UserCouponsListResponse{
			Items: []api.UserCoupon{
				{
					ID:            "uc_1",
					UserID:        "user_1",
					CouponCode:    "FIRST10",
					Title:         "First Coupon",
					DiscountType:  "percentage",
					DiscountValue: 10,
					Status:        "active",
					ExpiresAt:     expiresAt,
					CreatedAt:     createdAt,
				},
				{
					ID:            "uc_2",
					UserID:        "user_2",
					CouponCode:    "SECOND20",
					Title:         "Second Coupon",
					DiscountType:  "fixed_amount",
					DiscountValue: 20,
					Status:        "used",
					ExpiresAt:     time.Time{},
					CreatedAt:     createdAt,
				},
				{
					ID:            "uc_3",
					UserID:        "user_3",
					CouponCode:    "THIRD30",
					Title:         "Third Coupon",
					DiscountType:  "percentage",
					DiscountValue: 30,
					Status:        "expired",
					ExpiresAt:     expiresAt,
					CreatedAt:     createdAt,
				},
			},
			TotalCount: 3,
		},
	}
	cleanup := setupUserCouponsMockFactories(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("user-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := userCouponsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expectedCoupons := []string{"FIRST10", "SECOND20", "THIRD30"}
	for _, coupon := range expectedCoupons {
		if !strings.Contains(output, coupon) {
			t.Errorf("output should contain %q", coupon)
		}
	}
}

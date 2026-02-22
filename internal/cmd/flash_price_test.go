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

// flashPriceMockAPIClient is a mock implementation of api.APIClient for flash price testing.
type flashPriceMockAPIClient struct {
	api.MockClient // embed base mock for unimplemented methods

	// Configurable return values for specific methods
	listFlashPricesResp *api.FlashPriceListResponse
	listFlashPricesErr  error

	getFlashPriceResp *api.FlashPrice
	getFlashPriceErr  error

	createFlashPriceResp *api.FlashPrice
	createFlashPriceErr  error

	activateFlashPriceResp *api.FlashPrice
	activateFlashPriceErr  error

	deactivateFlashPriceResp *api.FlashPrice
	deactivateFlashPriceErr  error

	updateFlashPriceResp *api.FlashPrice
	updateFlashPriceErr  error

	deleteFlashPriceErr error
}

func (m *flashPriceMockAPIClient) ListFlashPrices(ctx context.Context, opts *api.FlashPriceListOptions) (*api.FlashPriceListResponse, error) {
	return m.listFlashPricesResp, m.listFlashPricesErr
}

func (m *flashPriceMockAPIClient) GetFlashPrice(ctx context.Context, id string) (*api.FlashPrice, error) {
	return m.getFlashPriceResp, m.getFlashPriceErr
}

func (m *flashPriceMockAPIClient) CreateFlashPrice(ctx context.Context, req *api.FlashPriceCreateRequest) (*api.FlashPrice, error) {
	return m.createFlashPriceResp, m.createFlashPriceErr
}

func (m *flashPriceMockAPIClient) ActivateFlashPrice(ctx context.Context, id string) (*api.FlashPrice, error) {
	return m.activateFlashPriceResp, m.activateFlashPriceErr
}

func (m *flashPriceMockAPIClient) DeactivateFlashPrice(ctx context.Context, id string) (*api.FlashPrice, error) {
	return m.deactivateFlashPriceResp, m.deactivateFlashPriceErr
}

func (m *flashPriceMockAPIClient) UpdateFlashPrice(ctx context.Context, id string, req *api.FlashPriceUpdateRequest) (*api.FlashPrice, error) {
	return m.updateFlashPriceResp, m.updateFlashPriceErr
}

func (m *flashPriceMockAPIClient) DeleteFlashPrice(ctx context.Context, id string) error {
	return m.deleteFlashPriceErr
}

// setupFlashPriceMockFactories configures mock factories for flash price testing.
func setupFlashPriceMockFactories(mockClient *flashPriceMockAPIClient) (cleanup func()) {
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

// newFlashPriceTestCmd creates a test command with common flags for flash price tests.
func newFlashPriceTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	return cmd
}

// TestFlashPriceCommandSetup verifies flash-price command initialization
func TestFlashPriceCommandSetup(t *testing.T) {
	if flashPriceCmd.Use != "flash-price" {
		t.Errorf("expected Use 'flash-price', got %q", flashPriceCmd.Use)
	}
	if flashPriceCmd.Short != "Manage flash sale pricing" {
		t.Errorf("expected Short 'Manage flash sale pricing', got %q", flashPriceCmd.Short)
	}
}

// TestFlashPriceSubcommands verifies all subcommands are registered
func TestFlashPriceSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":       "List flash prices",
		"get":        "Get flash price details",
		"create":     "Create a flash price",
		"update":     "Update a flash price campaign",
		"activate":   "Activate a flash price",
		"deactivate": "Deactivate a flash price",
		"delete":     "Delete a flash price",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range flashPriceCmd.Commands() {
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

// TestFlashPriceListFlags verifies list command flags exist with correct defaults
func TestFlashPriceListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"product-id", ""},
		{"status", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := flashPriceListCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Fatalf("flag %q not found", f.name)
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}
}

// TestFlashPriceCreateFlags verifies create command flags exist with correct defaults
func TestFlashPriceCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"product-id", ""},
		{"variant-id", ""},
		{"flash-price", "0"},
		{"quantity", "0"},
		{"limit-per-user", "0"},
		{"starts-at", ""},
		{"ends-at", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := flashPriceCreateCmd.Flags().Lookup(f.name)
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

// TestFlashPriceCreateRequiredFlags verifies product-id and flash-price are required
func TestFlashPriceCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"product-id", "flash-price"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := flashPriceCreateCmd.Flags().Lookup(name)
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

// TestFlashPriceDeleteFlags verifies delete command flags exist with correct defaults
func TestFlashPriceDeleteFlags(t *testing.T) {
	flag := flashPriceDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected flag 'yes'")
	}
	if flag.DefValue != "false" {
		t.Errorf("Expected default 'false', got %q", flag.DefValue)
	}
}

// TestFlashPriceGetArgs verifies get command requires exactly 1 argument
func TestFlashPriceGetArgs(t *testing.T) {
	err := flashPriceGetCmd.Args(flashPriceGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = flashPriceGetCmd.Args(flashPriceGetCmd, []string{"fp_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestFlashPriceActivateArgs verifies activate command requires exactly 1 argument
func TestFlashPriceActivateArgs(t *testing.T) {
	err := flashPriceActivateCmd.Args(flashPriceActivateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = flashPriceActivateCmd.Args(flashPriceActivateCmd, []string{"fp_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestFlashPriceDeactivateArgs verifies deactivate command requires exactly 1 argument
func TestFlashPriceDeactivateArgs(t *testing.T) {
	err := flashPriceDeactivateCmd.Args(flashPriceDeactivateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = flashPriceDeactivateCmd.Args(flashPriceDeactivateCmd, []string{"fp_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestFlashPriceDeleteArgs verifies delete command requires exactly 1 argument
func TestFlashPriceDeleteArgs(t *testing.T) {
	err := flashPriceDeleteCmd.Args(flashPriceDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = flashPriceDeleteCmd.Args(flashPriceDeleteCmd, []string{"fp_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestFlashPriceListRunE tests the flash-price list command execution with mock API.
func TestFlashPriceListRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.FlashPriceListResponse
		mockErr    error
		output     string
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful list with quantity limit",
			output: "text",
			mockResp: &api.FlashPriceListResponse{
				Items: []api.FlashPrice{
					{
						ID:            "fp_123",
						ProductID:     "prod_456",
						OriginalPrice: 99.99,
						FlashPrice:    79.99,
						DiscountPct:   20,
						Quantity:      100,
						QuantitySold:  25,
						Status:        "active",
						StartsAt:      startsAt,
						EndsAt:        endsAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "fp_123",
		},
		{
			name:   "successful list unlimited quantity",
			output: "text",
			mockResp: &api.FlashPriceListResponse{
				Items: []api.FlashPrice{
					{
						ID:            "fp_789",
						ProductID:     "prod_012",
						OriginalPrice: 199.99,
						FlashPrice:    149.99,
						DiscountPct:   25,
						Quantity:      0,
						QuantitySold:  50,
						Status:        "active",
						StartsAt:      startsAt,
						EndsAt:        time.Time{},
					},
				},
				TotalCount: 1,
			},
			wantOutput: "fp_789",
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.FlashPriceListResponse{
				Items: []api.FlashPrice{
					{
						ID:            "fp_json",
						ProductID:     "prod_json",
						OriginalPrice: 49.99,
						FlashPrice:    39.99,
						DiscountPct:   20,
						Status:        "scheduled",
						StartsAt:      startsAt,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "fp_json",
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
			mockResp: &api.FlashPriceListResponse{
				Items:      []api.FlashPrice{},
				TotalCount: 0,
			},
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				listFlashPricesResp: tt.mockResp,
				listFlashPricesErr:  tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newFlashPriceTestCmd()
			_ = cmd.Flags().Set("output", tt.output)
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("status", "", "")

			err := flashPriceListCmd.RunE(cmd, []string{})

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

// TestFlashPriceGetRunE tests the flash-price get command execution with mock API.
func TestFlashPriceGetRunE(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2023, 12, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		flashPriceID string
		output       string
		mockResp     *api.FlashPrice
		mockErr      error
		wantErr      bool
		wantOutput   string
	}{
		{
			name:         "successful get with all fields",
			flashPriceID: "fp_123",
			output:       "text",
			mockResp: &api.FlashPrice{
				ID:            "fp_123",
				ProductID:     "prod_456",
				VariantID:     "var_789",
				OriginalPrice: 99.99,
				FlashPrice:    79.99,
				DiscountPct:   20,
				Quantity:      100,
				QuantitySold:  25,
				LimitPerUser:  2,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        endsAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "",
		},
		{
			name:         "successful get without optional fields",
			flashPriceID: "fp_456",
			output:       "text",
			mockResp: &api.FlashPrice{
				ID:            "fp_456",
				ProductID:     "prod_789",
				OriginalPrice: 49.99,
				FlashPrice:    39.99,
				DiscountPct:   20,
				Quantity:      0,
				QuantitySold:  10,
				LimitPerUser:  0,
				Status:        "active",
				StartsAt:      time.Time{},
				EndsAt:        time.Time{},
				CreatedAt:     createdAt,
			},
			wantOutput: "",
		},
		{
			name:         "successful get JSON output",
			flashPriceID: "fp_json",
			output:       "json",
			mockResp: &api.FlashPrice{
				ID:            "fp_json",
				ProductID:     "prod_json",
				OriginalPrice: 29.99,
				FlashPrice:    24.99,
				DiscountPct:   17,
				Status:        "scheduled",
				StartsAt:      startsAt,
				CreatedAt:     createdAt,
			},
			wantOutput: "fp_json",
		},
		{
			name:         "flash price not found",
			flashPriceID: "fp_999",
			output:       "text",
			mockErr:      errors.New("flash price not found"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				getFlashPriceResp: tt.mockResp,
				getFlashPriceErr:  tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newFlashPriceTestCmd()
			_ = cmd.Flags().Set("output", tt.output)

			err := flashPriceGetCmd.RunE(cmd, []string{tt.flashPriceID})

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

// TestFlashPriceCreateRunE tests the flash-price create command execution with mock API.
func TestFlashPriceCreateRunE(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		startsAt   string
		endsAt     string
		mockResp   *api.FlashPrice
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful create",
			output: "text",
			mockResp: &api.FlashPrice{
				ID:            "fp_new",
				ProductID:     "prod_123",
				OriginalPrice: 99.99,
				FlashPrice:    79.99,
				DiscountPct:   20,
				Status:        "scheduled",
			},
			wantOutput: "",
		},
		{
			name:     "successful create with dates",
			output:   "text",
			startsAt: "2024-01-01T00:00:00Z",
			endsAt:   "2024-12-31T23:59:59Z",
			mockResp: &api.FlashPrice{
				ID:            "fp_dated",
				ProductID:     "prod_456",
				OriginalPrice: 49.99,
				FlashPrice:    39.99,
				DiscountPct:   20,
				Status:        "scheduled",
			},
			wantOutput: "",
		},
		{
			name:   "successful create JSON output",
			output: "json",
			mockResp: &api.FlashPrice{
				ID:            "fp_json",
				ProductID:     "prod_json",
				OriginalPrice: 29.99,
				FlashPrice:    24.99,
				DiscountPct:   17,
				Status:        "active",
			},
			wantOutput: "fp_json",
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
			mockClient := &flashPriceMockAPIClient{
				createFlashPriceResp: tt.mockResp,
				createFlashPriceErr:  tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newFlashPriceTestCmd()
			_ = cmd.Flags().Set("output", tt.output)
			cmd.Flags().String("product-id", "prod_123", "")
			cmd.Flags().String("variant-id", "", "")
			cmd.Flags().Float64("flash-price", 79.99, "")
			cmd.Flags().Int("quantity", 0, "")
			cmd.Flags().Int("limit-per-user", 0, "")
			cmd.Flags().String("starts-at", tt.startsAt, "")
			cmd.Flags().String("ends-at", tt.endsAt, "")

			err := flashPriceCreateCmd.RunE(cmd, []string{})

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

// TestFlashPriceCreateRunE_InvalidStartsAt tests invalid starts-at date format
func TestFlashPriceCreateRunE_InvalidStartsAt(t *testing.T) {
	mockClient := &flashPriceMockAPIClient{}
	cleanup := setupFlashPriceMockFactories(mockClient)
	defer cleanup()

	cmd := newFlashPriceTestCmd()
	cmd.Flags().String("product-id", "prod_123", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Float64("flash-price", 9.99, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "invalid-date", "")
	cmd.Flags().String("ends-at", "", "")

	err := flashPriceCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for invalid starts-at format")
	}
	if !strings.Contains(err.Error(), "starts-at") {
		t.Errorf("Expected error message to mention starts-at, got: %v", err)
	}
}

// TestFlashPriceCreateRunE_InvalidEndsAt tests invalid ends-at date format
func TestFlashPriceCreateRunE_InvalidEndsAt(t *testing.T) {
	mockClient := &flashPriceMockAPIClient{}
	cleanup := setupFlashPriceMockFactories(mockClient)
	defer cleanup()

	cmd := newFlashPriceTestCmd()
	cmd.Flags().String("product-id", "prod_123", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Float64("flash-price", 9.99, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "invalid-date", "")

	err := flashPriceCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for invalid ends-at format")
	}
	if !strings.Contains(err.Error(), "ends-at") {
		t.Errorf("Expected error message to mention ends-at, got: %v", err)
	}
}

// TestFlashPriceActivateRunE tests the flash-price activate command execution with mock API.
func TestFlashPriceActivateRunE(t *testing.T) {
	tests := []struct {
		name         string
		flashPriceID string
		mockResp     *api.FlashPrice
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful activate",
			flashPriceID: "fp_123",
			mockResp: &api.FlashPrice{
				ID:     "fp_123",
				Status: "active",
			},
		},
		{
			name:         "activate fails",
			flashPriceID: "fp_456",
			mockErr:      errors.New("flash price already active"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				activateFlashPriceResp: tt.mockResp,
				activateFlashPriceErr:  tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			cmd := newFlashPriceTestCmd()

			err := flashPriceActivateCmd.RunE(cmd, []string{tt.flashPriceID})

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

// TestFlashPriceDeactivateRunE tests the flash-price deactivate command execution with mock API.
func TestFlashPriceDeactivateRunE(t *testing.T) {
	tests := []struct {
		name         string
		flashPriceID string
		mockResp     *api.FlashPrice
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful deactivate",
			flashPriceID: "fp_123",
			mockResp: &api.FlashPrice{
				ID:     "fp_123",
				Status: "inactive",
			},
		},
		{
			name:         "deactivate fails",
			flashPriceID: "fp_456",
			mockErr:      errors.New("flash price already inactive"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				deactivateFlashPriceResp: tt.mockResp,
				deactivateFlashPriceErr:  tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			cmd := newFlashPriceTestCmd()

			err := flashPriceDeactivateCmd.RunE(cmd, []string{tt.flashPriceID})

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

// TestFlashPriceDeleteRunE tests the flash-price delete command execution with mock API.
func TestFlashPriceDeleteRunE(t *testing.T) {
	tests := []struct {
		name         string
		flashPriceID string
		mockErr      error
		wantErr      bool
	}{
		{
			name:         "successful delete",
			flashPriceID: "fp_123",
			mockErr:      nil,
		},
		{
			name:         "delete fails",
			flashPriceID: "fp_456",
			mockErr:      errors.New("flash price not found"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				deleteFlashPriceErr: tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			cmd := newFlashPriceTestCmd()
			cmd.Flags().Bool("yes", true, "") // Skip confirmation

			err := flashPriceDeleteCmd.RunE(cmd, []string{tt.flashPriceID})

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

// TestFlashPriceListGetClientError verifies list command error handling when getClient fails
func TestFlashPriceListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("product-id", "", "")
	cmd.Flags().String("status", "", "")

	err := flashPriceListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestFlashPriceGetGetClientError verifies get command error handling when getClient fails
func TestFlashPriceGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := flashPriceGetCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestFlashPriceCreateGetClientError verifies create command error handling when getClient fails
func TestFlashPriceCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("product-id", "prod_123", "")
	cmd.Flags().String("variant-id", "", "")
	cmd.Flags().Float64("flash-price", 9.99, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")

	err := flashPriceCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestFlashPriceActivateGetClientError verifies activate command error handling when getClient fails
func TestFlashPriceActivateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := flashPriceActivateCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestFlashPriceDeactivateGetClientError verifies deactivate command error handling when getClient fails
func TestFlashPriceDeactivateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := flashPriceDeactivateCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestFlashPriceDeleteGetClientError verifies delete command error handling when getClient fails
func TestFlashPriceDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true") // yes flag is already defined in newTestCmdWithFlags
	err := flashPriceDeleteCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestFlashPriceListTextOutputFormatting tests specific text output formatting.
func TestFlashPriceListTextOutputFormatting(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name        string
		flashPrice  api.FlashPrice
		wantOutputs []string
	}{
		{
			name: "flash price with quantity limit",
			flashPrice: api.FlashPrice{
				ID:            "fp_qty",
				ProductID:     "prod_qty",
				OriginalPrice: 99.99,
				FlashPrice:    79.99,
				DiscountPct:   20,
				Quantity:      100,
				QuantitySold:  25,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        endsAt,
			},
			wantOutputs: []string{"fp_qty", "prod_qty", "20%", "25/100"},
		},
		{
			name: "flash price unlimited quantity",
			flashPrice: api.FlashPrice{
				ID:            "fp_unlim",
				ProductID:     "prod_unlim",
				OriginalPrice: 49.99,
				FlashPrice:    39.99,
				DiscountPct:   20,
				Quantity:      0,
				QuantitySold:  50,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        time.Time{},
			},
			wantOutputs: []string{"fp_unlim", "prod_unlim"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				listFlashPricesResp: &api.FlashPriceListResponse{
					Items:      []api.FlashPrice{tt.flashPrice},
					TotalCount: 1,
				},
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newFlashPriceTestCmd()
			_ = cmd.Flags().Set("output", "text")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("status", "", "")

			err := flashPriceListCmd.RunE(cmd, []string{})
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

// TestFlashPriceGetTextOutputFormatting tests specific text output formatting for get command.
func TestFlashPriceGetTextOutputFormatting(t *testing.T) {
	startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2023, 12, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		flashPrice api.FlashPrice
	}{
		{
			name: "flash price with all optional fields",
			flashPrice: api.FlashPrice{
				ID:            "fp_full",
				ProductID:     "prod_full",
				VariantID:     "var_full",
				OriginalPrice: 99.99,
				FlashPrice:    79.99,
				DiscountPct:   20,
				Quantity:      100,
				QuantitySold:  25,
				LimitPerUser:  5,
				Status:        "active",
				StartsAt:      startsAt,
				EndsAt:        endsAt,
				CreatedAt:     createdAt,
			},
		},
		{
			name: "flash price with minimal fields",
			flashPrice: api.FlashPrice{
				ID:            "fp_min",
				ProductID:     "prod_min",
				OriginalPrice: 49.99,
				FlashPrice:    39.99,
				DiscountPct:   20,
				Quantity:      0,
				QuantitySold:  0,
				LimitPerUser:  0,
				Status:        "inactive",
				StartsAt:      time.Time{},
				EndsAt:        time.Time{},
				CreatedAt:     createdAt,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				getFlashPriceResp: &tt.flashPrice,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newFlashPriceTestCmd()
			_ = cmd.Flags().Set("output", "text")

			err := flashPriceGetCmd.RunE(cmd, []string{tt.flashPrice.ID})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Note: text output goes to stdout directly via fmt.Printf
			// We're verifying the command runs without error
		})
	}
}

// TestFlashPriceUpdateFlags verifies update command flags exist with correct defaults
func TestFlashPriceUpdateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"flash-price", "0"},
		{"quantity", "0"},
		{"limit-per-user", "0"},
		{"starts-at", ""},
		{"ends-at", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := flashPriceUpdateCmd.Flags().Lookup(f.name)
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

// TestFlashPriceUpdateArgs verifies update command requires exactly 1 argument
func TestFlashPriceUpdateArgs(t *testing.T) {
	err := flashPriceUpdateCmd.Args(flashPriceUpdateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = flashPriceUpdateCmd.Args(flashPriceUpdateCmd, []string{"fp_123"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

// TestFlashPriceUpdateRunE tests the flash-price update command execution with mock API.
func TestFlashPriceUpdateRunE(t *testing.T) {
	tests := []struct {
		name         string
		flashPriceID string
		output       string
		startsAt     string
		endsAt       string
		mockResp     *api.FlashPrice
		mockErr      error
		wantErr      bool
		wantOutput   string
	}{
		{
			name:         "successful update",
			flashPriceID: "fp_123",
			output:       "text",
			mockResp: &api.FlashPrice{
				ID:            "fp_123",
				ProductID:     "prod_123",
				OriginalPrice: 99.99,
				FlashPrice:    89.99,
				DiscountPct:   10,
				Status:        "active",
			},
			wantOutput: "",
		},
		{
			name:         "successful update with dates",
			flashPriceID: "fp_456",
			output:       "text",
			startsAt:     "2024-01-01T00:00:00Z",
			endsAt:       "2024-12-31T23:59:59Z",
			mockResp: &api.FlashPrice{
				ID:            "fp_456",
				ProductID:     "prod_456",
				OriginalPrice: 49.99,
				FlashPrice:    39.99,
				DiscountPct:   20,
				Status:        "scheduled",
			},
			wantOutput: "",
		},
		{
			name:         "successful update JSON output",
			flashPriceID: "fp_json",
			output:       "json",
			mockResp: &api.FlashPrice{
				ID:            "fp_json",
				ProductID:     "prod_json",
				OriginalPrice: 29.99,
				FlashPrice:    24.99,
				DiscountPct:   17,
				Status:        "active",
			},
			wantOutput: "fp_json",
		},
		{
			name:         "update fails",
			flashPriceID: "fp_999",
			output:       "text",
			mockErr:      errors.New("flash price not found"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &flashPriceMockAPIClient{
				updateFlashPriceResp: tt.mockResp,
				updateFlashPriceErr:  tt.mockErr,
			}
			cleanup := setupFlashPriceMockFactories(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := newFlashPriceTestCmd()
			_ = cmd.Flags().Set("output", tt.output)
			cmd.Flags().Float64("flash-price", 89.99, "")
			cmd.Flags().Int("quantity", 0, "")
			cmd.Flags().Int("limit-per-user", 0, "")
			cmd.Flags().String("starts-at", tt.startsAt, "")
			cmd.Flags().String("ends-at", tt.endsAt, "")

			err := flashPriceUpdateCmd.RunE(cmd, []string{tt.flashPriceID})

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

// TestFlashPriceUpdateRunE_InvalidStartsAt tests invalid starts-at date format
func TestFlashPriceUpdateRunE_InvalidStartsAt(t *testing.T) {
	mockClient := &flashPriceMockAPIClient{}
	cleanup := setupFlashPriceMockFactories(mockClient)
	defer cleanup()

	cmd := newFlashPriceTestCmd()
	cmd.Flags().Float64("flash-price", 0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "invalid-date", "")
	cmd.Flags().String("ends-at", "", "")

	err := flashPriceUpdateCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Fatal("Expected error for invalid starts-at format")
	}
	if !strings.Contains(err.Error(), "starts-at") {
		t.Errorf("Expected error message to mention starts-at, got: %v", err)
	}
}

// TestFlashPriceUpdateRunE_InvalidEndsAt tests invalid ends-at date format
func TestFlashPriceUpdateRunE_InvalidEndsAt(t *testing.T) {
	mockClient := &flashPriceMockAPIClient{}
	cleanup := setupFlashPriceMockFactories(mockClient)
	defer cleanup()

	cmd := newFlashPriceTestCmd()
	cmd.Flags().Float64("flash-price", 0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "invalid-date", "")

	err := flashPriceUpdateCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Fatal("Expected error for invalid ends-at format")
	}
	if !strings.Contains(err.Error(), "ends-at") {
		t.Errorf("Expected error message to mention ends-at, got: %v", err)
	}
}

// TestFlashPriceUpdateGetClientError verifies update command error handling when getClient fails
func TestFlashPriceUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().Float64("flash-price", 0, "")
	cmd.Flags().Int("quantity", 0, "")
	cmd.Flags().Int("limit-per-user", 0, "")
	cmd.Flags().String("starts-at", "", "")
	cmd.Flags().String("ends-at", "", "")

	err := flashPriceUpdateCmd.RunE(cmd, []string{"fp_123"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

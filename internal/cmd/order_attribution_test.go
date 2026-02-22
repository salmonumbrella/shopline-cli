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

// mockOrderAttributionClient is a mock implementation for order attribution API methods.
type mockOrderAttributionClient struct {
	api.MockClient

	listOrderAttributionsResp *api.OrderAttributionListResponse
	listOrderAttributionsErr  error

	getOrderAttributionResp *api.OrderAttribution
	getOrderAttributionErr  error
}

func (m *mockOrderAttributionClient) ListOrderAttributions(ctx context.Context, opts *api.OrderAttributionListOptions) (*api.OrderAttributionListResponse, error) {
	return m.listOrderAttributionsResp, m.listOrderAttributionsErr
}

func (m *mockOrderAttributionClient) GetOrderAttribution(ctx context.Context, orderID string) (*api.OrderAttribution, error) {
	return m.getOrderAttributionResp, m.getOrderAttributionErr
}

// setupOrderAttributionTest configures mocks for order attribution command tests.
func setupOrderAttributionTest(mockClient *mockOrderAttributionClient) (cleanup func()) {
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

func TestOrderAttributionCommandStructure(t *testing.T) {
	if orderAttributionCmd == nil {
		t.Fatal("orderAttributionCmd is nil")
	}
	if orderAttributionCmd.Use != "order-attribution" {
		t.Errorf("Expected Use 'order-attribution', got %q", orderAttributionCmd.Use)
	}
	if orderAttributionCmd.Short != "Manage order attribution tracking" {
		t.Errorf("Expected Short 'Manage order attribution tracking', got %q", orderAttributionCmd.Short)
	}
}

func TestOrderAttributionSubcommands(t *testing.T) {
	subcommands := map[string]bool{"list": false, "get": false}
	for _, cmd := range orderAttributionCmd.Commands() {
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

func TestOrderAttributionListFlags(t *testing.T) {
	cmd := orderAttributionListCmd
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"source", ""},
		{"medium", ""},
		{"campaign", ""},
		{"page", "1"},
		{"page-size", "20"},
	}
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestOrderAttributionGetRequiresArg(t *testing.T) {
	cmd := orderAttributionGetCmd
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

func TestOrderAttributionGetCmdUse(t *testing.T) {
	if orderAttributionGetCmd.Use != "get <order-id>" {
		t.Errorf("Expected Use 'get <order-id>', got %q", orderAttributionGetCmd.Use)
	}
}

func TestOrderAttributionListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := orderAttributionListCmd.RunE(cmd, []string{}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOrderAttributionGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) { return nil, errors.New("keyring error") }
	cmd := newTestCmdWithFlags()
	if err := orderAttributionGetCmd.RunE(cmd, []string{"order_123"}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOrderAttributionListNoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) { return &mockStore{names: []string{}}, nil }
	cmd := newTestCmdWithFlags()
	err := orderAttributionListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

func TestOrderAttributionListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() { secretsStoreFactory = origFactory; _ = os.Setenv("SHOPLINE_STORE", origEnv) }()
	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{"envstore", "other"}, creds: map[string]*secrets.StoreCredentials{"envstore": {Handle: "test", AccessToken: "token123"}}}, nil
	}
	cmd := newTestCmdWithFlags()
	err := orderAttributionListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

func TestOrderAttributionListRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		mockResp     *api.OrderAttributionListResponse
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string
	}{
		{
			name: "successful list with text output",
			mockResp: &api.OrderAttributionListResponse{
				Items: []api.OrderAttribution{
					{
						ID:              "attr_123",
						OrderID:         "order_456",
						Source:          "google",
						Medium:          "cpc",
						Campaign:        "summer_sale",
						TouchpointCount: 3,
						CreatedAt:       createdAt,
					},
				},
				TotalCount: 1,
			},
			outputFormat: "text",
			wantOutput:   "attr_123",
		},
		{
			name: "successful list with JSON output",
			mockResp: &api.OrderAttributionListResponse{
				Items: []api.OrderAttribution{
					{
						ID:              "attr_789",
						OrderID:         "order_101",
						Source:          "facebook",
						Medium:          "social",
						Campaign:        "winter_promo",
						TouchpointCount: 5,
						CreatedAt:       createdAt,
					},
				},
				TotalCount: 1,
			},
			outputFormat: "json",
			wantOutput:   "attr_789",
		},
		{
			name: "empty list",
			mockResp: &api.OrderAttributionListResponse{
				Items:      []api.OrderAttribution{},
				TotalCount: 0,
			},
			outputFormat: "text",
		},
		{
			name: "multiple attributions",
			mockResp: &api.OrderAttributionListResponse{
				Items: []api.OrderAttribution{
					{
						ID:              "attr_001",
						OrderID:         "order_001",
						Source:          "google",
						Medium:          "organic",
						Campaign:        "",
						TouchpointCount: 1,
						CreatedAt:       createdAt,
					},
					{
						ID:              "attr_002",
						OrderID:         "order_002",
						Source:          "direct",
						Medium:          "none",
						Campaign:        "",
						TouchpointCount: 2,
						CreatedAt:       createdAt,
					},
				},
				TotalCount: 2,
			},
			outputFormat: "text",
			wantOutput:   "attr_001",
		},
		{
			name:         "API error",
			mockErr:      errors.New("API unavailable"),
			outputFormat: "text",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOrderAttributionClient{
				listOrderAttributionsResp: tt.mockResp,
				listOrderAttributionsErr:  tt.mockErr,
			}
			cleanup := setupOrderAttributionTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("source", "", "")
			cmd.Flags().String("medium", "", "")
			cmd.Flags().String("campaign", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := orderAttributionListCmd.RunE(cmd, []string{})

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

func TestOrderAttributionListWithFilters(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockOrderAttributionClient{
		listOrderAttributionsResp: &api.OrderAttributionListResponse{
			Items: []api.OrderAttribution{
				{
					ID:              "attr_filtered",
					OrderID:         "order_filtered",
					Source:          "google",
					Medium:          "cpc",
					Campaign:        "test_campaign",
					TouchpointCount: 2,
					CreatedAt:       createdAt,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup := setupOrderAttributionTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("source", "google", "")
	cmd.Flags().String("medium", "cpc", "")
	cmd.Flags().String("campaign", "test_campaign", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 10, "")

	err := orderAttributionListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "attr_filtered") {
		t.Errorf("output should contain 'attr_filtered', got: %s", output)
	}
}

func TestOrderAttributionGetRunE(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	firstVisit := time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2024, 1, 14, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		orderID      string
		mockResp     *api.OrderAttribution
		mockErr      error
		outputFormat string
		wantErr      bool
		wantOutput   string // Only checked for JSON output
	}{
		{
			name:    "successful get with text output - all fields",
			orderID: "order_123",
			mockResp: &api.OrderAttribution{
				ID:              "attr_123",
				OrderID:         "order_123",
				Source:          "google",
				Medium:          "cpc",
				Campaign:        "summer_sale",
				Content:         "banner_ad",
				Term:            "shoes",
				ReferrerURL:     "https://google.com/search",
				LandingPage:     "https://shop.com/products/shoes",
				UtmSource:       "google",
				UtmMedium:       "cpc",
				UtmCampaign:     "summer_sale",
				UtmContent:      "banner_ad",
				UtmTerm:         "shoes",
				FirstVisitAt:    &firstVisit,
				LastVisitAt:     &lastVisit,
				TouchpointCount: 5,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			},
			outputFormat: "text",
			// Text output goes to stdout via fmt.Printf, not to formatterWriter
		},
		{
			name:    "successful get with text output - minimal fields",
			orderID: "order_456",
			mockResp: &api.OrderAttribution{
				ID:              "attr_456",
				OrderID:         "order_456",
				Source:          "direct",
				Medium:          "none",
				TouchpointCount: 1,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			},
			outputFormat: "text",
			// Text output goes to stdout via fmt.Printf, not to formatterWriter
		},
		{
			name:    "successful get with JSON output",
			orderID: "order_789",
			mockResp: &api.OrderAttribution{
				ID:              "attr_789",
				OrderID:         "order_789",
				Source:          "facebook",
				Medium:          "social",
				Campaign:        "winter_promo",
				TouchpointCount: 3,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			},
			outputFormat: "json",
			wantOutput:   "attr_789",
		},
		{
			name:         "order not found",
			orderID:      "order_nonexistent",
			mockErr:      errors.New("order not found"),
			outputFormat: "text",
			wantErr:      true,
		},
		{
			name:         "API error",
			orderID:      "order_error",
			mockErr:      errors.New("API unavailable"),
			outputFormat: "text",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOrderAttributionClient{
				getOrderAttributionResp: tt.mockResp,
				getOrderAttributionErr:  tt.mockErr,
			}
			cleanup := setupOrderAttributionTest(mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", tt.outputFormat, "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")

			err := orderAttributionGetCmd.RunE(cmd, []string{tt.orderID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.mockErr != nil && !strings.Contains(err.Error(), "failed to get order attribution") {
					t.Errorf("expected error to wrap original error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Only check output for JSON format (text goes to stdout via fmt.Printf)
			if tt.outputFormat == "json" && tt.wantOutput != "" {
				output := buf.String()
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("output %q should contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestOrderAttributionGetRunE_WithOptionalFields(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	firstVisit := time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2024, 1, 14, 18, 0, 0, 0, time.UTC)

	// Test with all optional fields populated
	mockClient := &mockOrderAttributionClient{
		getOrderAttributionResp: &api.OrderAttribution{
			ID:              "attr_full",
			OrderID:         "order_full",
			Source:          "google",
			Medium:          "cpc",
			Campaign:        "holiday_sale",
			Content:         "sidebar_banner",
			Term:            "winter boots",
			ReferrerURL:     "https://www.google.com/search?q=winter+boots",
			LandingPage:     "https://shop.example.com/products/boots",
			UtmSource:       "google",
			UtmMedium:       "cpc",
			UtmCampaign:     "holiday_sale",
			UtmContent:      "sidebar_banner",
			UtmTerm:         "winter boots",
			FirstVisitAt:    &firstVisit,
			LastVisitAt:     &lastVisit,
			TouchpointCount: 7,
			CreatedAt:       createdAt,
		},
	}
	cleanup := setupOrderAttributionTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	// Text output goes to stdout via fmt.Printf, so we just verify no error
	err := orderAttributionGetCmd.RunE(cmd, []string{"order_full"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// The command prints to stdout via fmt.Printf. The test output shows
	// the correct fields are printed. We can't capture stdout easily in tests
	// without modifying the command to use a writer, so we just verify success.
}

func TestOrderAttributionGetRunE_WithNilOptionalFields(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// Test with nil optional time fields
	mockClient := &mockOrderAttributionClient{
		getOrderAttributionResp: &api.OrderAttribution{
			ID:              "attr_minimal",
			OrderID:         "order_minimal",
			Source:          "direct",
			Medium:          "none",
			TouchpointCount: 1,
			CreatedAt:       createdAt,
		},
	}
	cleanup := setupOrderAttributionTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	// Text output goes to stdout via fmt.Printf, so we just verify no error
	err := orderAttributionGetCmd.RunE(cmd, []string{"order_minimal"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// The command prints to stdout via fmt.Printf. We verify the command
	// handles nil optional fields without panicking.
}

func TestOrderAttributionListMultipleProfiles(t *testing.T) {
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
	err := orderAttributionListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
	if !strings.Contains(err.Error(), "multiple profiles configured") {
		t.Errorf("Expected 'multiple profiles' error, got: %v", err)
	}
}

func TestOrderAttributionGetMultipleProfiles(t *testing.T) {
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
	err := orderAttributionGetCmd.RunE(cmd, []string{"order_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

func TestOrderAttributionListAPIError(t *testing.T) {
	mockClient := &mockOrderAttributionClient{
		listOrderAttributionsErr: errors.New("connection refused"),
	}
	cleanup := setupOrderAttributionTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().String("medium", "", "")
	cmd.Flags().String("campaign", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := orderAttributionListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list order attributions") {
		t.Errorf("expected error to mention 'failed to list order attributions', got: %v", err)
	}
}

func TestOrderAttributionListJSONOutput(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockOrderAttributionClient{
		listOrderAttributionsResp: &api.OrderAttributionListResponse{
			Items: []api.OrderAttribution{
				{
					ID:              "attr_json_test",
					OrderID:         "order_json_test",
					Source:          "email",
					Medium:          "newsletter",
					Campaign:        "weekly_digest",
					TouchpointCount: 2,
					CreatedAt:       createdAt,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup := setupOrderAttributionTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().String("medium", "", "")
	cmd.Flags().String("campaign", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := orderAttributionListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	// JSON output should contain the ID
	if !strings.Contains(output, "attr_json_test") {
		t.Errorf("JSON output should contain 'attr_json_test', got: %s", output)
	}
}

func TestOrderAttributionGetJSONOutput(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &mockOrderAttributionClient{
		getOrderAttributionResp: &api.OrderAttribution{
			ID:              "attr_json_get",
			OrderID:         "order_json_get",
			Source:          "affiliate",
			Medium:          "partner",
			Campaign:        "partner_promo",
			TouchpointCount: 4,
			CreatedAt:       createdAt,
		},
	}
	cleanup := setupOrderAttributionTest(mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := orderAttributionGetCmd.RunE(cmd, []string{"order_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	// JSON output should contain the ID
	if !strings.Contains(output, "attr_json_get") {
		t.Errorf("JSON output should contain 'attr_json_get', got: %s", output)
	}
}

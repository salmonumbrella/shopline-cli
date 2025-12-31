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

// mockFulfillmentsClient is a mock implementation of api.APIClient for fulfillments testing.
type mockFulfillmentsClient struct {
	api.MockClient // embed base mock for unimplemented methods

	listFulfillmentsResp *api.FulfillmentsListResponse
	listFulfillmentsErr  error

	getFulfillmentResp *api.Fulfillment
	getFulfillmentErr  error
}

func (m *mockFulfillmentsClient) ListFulfillments(ctx context.Context, opts *api.FulfillmentsListOptions) (*api.FulfillmentsListResponse, error) {
	return m.listFulfillmentsResp, m.listFulfillmentsErr
}

func (m *mockFulfillmentsClient) GetFulfillment(ctx context.Context, id string) (*api.Fulfillment, error) {
	return m.getFulfillmentResp, m.getFulfillmentErr
}

// setupFulfillmentsTest sets up the test environment for fulfillments commands.
func setupFulfillmentsTest(t *testing.T, mockClient *mockFulfillmentsClient) (cleanup func()) {
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

func TestFulfillmentsCmd(t *testing.T) {
	if fulfillmentsCmd.Use != "fulfillments" {
		t.Errorf("Expected Use 'fulfillments', got %q", fulfillmentsCmd.Use)
	}
	if fulfillmentsCmd.Short != "Manage fulfillments" {
		t.Errorf("Expected Short 'Manage fulfillments', got %q", fulfillmentsCmd.Short)
	}
}

func TestFulfillmentsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list": "List fulfillments",
		"get":  "Get fulfillment details",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range fulfillmentsCmd.Commands() {
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

func TestFulfillmentsListCmd(t *testing.T) {
	if fulfillmentsListCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got %q", fulfillmentsListCmd.Use)
	}
}

func TestFulfillmentsGetCmd(t *testing.T) {
	if fulfillmentsGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use 'get <id>', got %q", fulfillmentsGetCmd.Use)
	}
}

func TestFulfillmentsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"order-id", ""},
		{"status", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := fulfillmentsListCmd.Flags().Lookup(f.name)
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

func TestFulfillmentsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentsListRunE_Success(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		listFulfillmentsResp: &api.FulfillmentsListResponse{
			Items: []api.Fulfillment{
				{
					ID:              "ff_123",
					OrderID:         "ord_456",
					Status:          api.FulfillmentStatusSuccess,
					TrackingCompany: "FedEx",
					TrackingNumber:  "1234567890",
					CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				},
				{
					ID:              "ff_789",
					OrderID:         "ord_111",
					Status:          api.FulfillmentStatusPending,
					TrackingCompany: "UPS",
					TrackingNumber:  "9876543210",
					CreatedAt:       time.Date(2024, 1, 16, 9, 0, 0, 0, time.UTC),
					UpdatedAt:       time.Date(2024, 1, 16, 9, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 2,
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ff_123") {
		t.Errorf("output should contain 'ff_123', got: %s", output)
	}
	if !strings.Contains(output, "ord_456") {
		t.Errorf("output should contain 'ord_456', got: %s", output)
	}
	if !strings.Contains(output, "FedEx") {
		t.Errorf("output should contain 'FedEx', got: %s", output)
	}
}

func TestFulfillmentsListRunE_WithFilters(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		listFulfillmentsResp: &api.FulfillmentsListResponse{
			Items: []api.Fulfillment{
				{
					ID:              "ff_123",
					OrderID:         "ord_456",
					Status:          api.FulfillmentStatusSuccess,
					TrackingCompany: "FedEx",
					TrackingNumber:  "1234567890",
					CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "ord_456", "")
	cmd.Flags().String("status", "success", "")
	cmd.Flags().Int("page", 2, "")
	cmd.Flags().Int("page-size", 10, "")
	_ = cmd.Flags().Set("order-id", "ord_456")
	_ = cmd.Flags().Set("status", "success")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFulfillmentsListRunE_JSONOutput(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		listFulfillmentsResp: &api.FulfillmentsListResponse{
			Items: []api.Fulfillment{
				{
					ID:              "ff_123",
					OrderID:         "ord_456",
					Status:          api.FulfillmentStatusSuccess,
					TrackingCompany: "FedEx",
					TrackingNumber:  "1234567890",
					CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"ff_123"`) && !strings.Contains(output, `"id": "ff_123"`) {
		t.Errorf("JSON output should contain fulfillment ID, got: %s", output)
	}
}

func TestFulfillmentsListRunE_EmptyList(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		listFulfillmentsResp: &api.FulfillmentsListResponse{
			Items:      []api.Fulfillment{},
			TotalCount: 0,
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFulfillmentsListRunE_APIError(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		listFulfillmentsErr: errors.New("API unavailable"),
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := fulfillmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list fulfillments") {
		t.Errorf("error should contain 'failed to list fulfillments', got: %v", err)
	}
}

func TestFulfillmentsGetRunE_Success(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		getFulfillmentResp: &api.Fulfillment{
			ID:              "ff_123",
			OrderID:         "ord_456",
			Status:          api.FulfillmentStatusSuccess,
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			TrackingURL:     "https://fedex.com/track/1234567890",
			LineItems: []api.FulfillmentLineItem{
				{
					ID:        "li_001",
					ProductID: "prod_001",
					VariantID: "var_001",
					Title:     "Test Product",
					Quantity:  2,
					SKU:       "SKU001",
				},
			},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFulfillmentsGetRunE_WithoutTrackingURL(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		getFulfillmentResp: &api.Fulfillment{
			ID:              "ff_123",
			OrderID:         "ord_456",
			Status:          api.FulfillmentStatusPending,
			TrackingCompany: "USPS",
			TrackingNumber:  "9999888877776666",
			TrackingURL:     "", // No tracking URL
			LineItems:       []api.FulfillmentLineItem{},
			CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFulfillmentsGetRunE_WithLineItems(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		getFulfillmentResp: &api.Fulfillment{
			ID:              "ff_123",
			OrderID:         "ord_456",
			Status:          api.FulfillmentStatusSuccess,
			TrackingCompany: "DHL",
			TrackingNumber:  "DHL123456",
			TrackingURL:     "https://dhl.com/track/DHL123456",
			LineItems: []api.FulfillmentLineItem{
				{
					ID:        "li_001",
					ProductID: "prod_001",
					VariantID: "var_001",
					Title:     "Widget A",
					Quantity:  3,
					SKU:       "WGT-A-001",
				},
				{
					ID:        "li_002",
					ProductID: "prod_002",
					VariantID: "var_002",
					Title:     "Gadget B",
					Quantity:  1,
					SKU:       "GDG-B-002",
				},
			},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFulfillmentsGetRunE_JSONOutput(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		getFulfillmentResp: &api.Fulfillment{
			ID:              "ff_123",
			OrderID:         "ord_456",
			Status:          api.FulfillmentStatusSuccess,
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			TrackingURL:     "https://fedex.com/track/1234567890",
			LineItems:       []api.FulfillmentLineItem{},
			CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"ff_123"`) && !strings.Contains(output, `"id": "ff_123"`) {
		t.Errorf("JSON output should contain fulfillment ID, got: %s", output)
	}
}

func TestFulfillmentsGetRunE_APIError(t *testing.T) {
	mockClient := &mockFulfillmentsClient{
		getFulfillmentErr: errors.New("fulfillment not found"),
	}

	cleanup := setupFulfillmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_nonexistent"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get fulfillment") {
		t.Errorf("error should contain 'failed to get fulfillment', got: %v", err)
	}
}

func TestFulfillmentsGetRunE_NoProfiles(t *testing.T) {
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
	err := fulfillmentsGetCmd.RunE(cmd, []string{"ff_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentsListRunE_AllStatuses(t *testing.T) {
	statuses := []api.FulfillmentStatus{
		api.FulfillmentStatusPending,
		api.FulfillmentStatusOpen,
		api.FulfillmentStatusSuccess,
		api.FulfillmentStatusCancelled,
		api.FulfillmentStatusFailure,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			mockClient := &mockFulfillmentsClient{
				listFulfillmentsResp: &api.FulfillmentsListResponse{
					Items: []api.Fulfillment{
						{
							ID:              "ff_123",
							OrderID:         "ord_456",
							Status:          status,
							TrackingCompany: "FedEx",
							TrackingNumber:  "1234567890",
							CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
							UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
						},
					},
					TotalCount: 1,
				},
			}

			cleanup := setupFulfillmentsTest(t, mockClient)
			defer cleanup()

			var buf bytes.Buffer
			formatterWriter = &buf

			cmd := &cobra.Command{Use: "test"}
			cmd.SetContext(context.Background())
			cmd.Flags().String("store", "", "")
			cmd.Flags().String("output", "", "")
			cmd.Flags().String("color", "never", "")
			cmd.Flags().String("query", "", "")
			cmd.Flags().String("order-id", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := fulfillmentsListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, string(status)) {
				t.Errorf("output should contain status '%s', got: %s", status, output)
			}
		})
	}
}

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

// TestShipmentsCommandSetup verifies shipments command initialization
func TestShipmentsCommandSetup(t *testing.T) {
	if shipmentsCmd.Use != "shipments" {
		t.Errorf("expected Use 'shipments', got %q", shipmentsCmd.Use)
	}
	if shipmentsCmd.Short != "Manage shipments" {
		t.Errorf("expected Short 'Manage shipments', got %q", shipmentsCmd.Short)
	}
}

// TestShipmentsSubcommands verifies all subcommands are registered
func TestShipmentsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List shipments",
		"get":    "Get shipment details",
		"create": "Create a shipment",
		"update": "Update a shipment",
		"delete": "Delete a shipment",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range shipmentsCmd.Commands() {
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

// TestShipmentsListFlags verifies list command flags exist with correct defaults
func TestShipmentsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"order-id", ""},
		{"fulfillment-id", ""},
		{"status", ""},
		{"tracking-number", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := shipmentsListCmd.Flags().Lookup(f.name)
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

// TestShipmentsCreateFlags verifies create command flags exist
func TestShipmentsCreateFlags(t *testing.T) {
	flags := []string{"order-id", "fulfillment-id", "tracking-company", "tracking-number", "tracking-url"}
	for _, flag := range flags {
		if shipmentsCreateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

// TestShipmentsUpdateFlags verifies update command flags exist
func TestShipmentsUpdateFlags(t *testing.T) {
	flags := []string{"tracking-company", "tracking-number", "tracking-url", "status"}
	for _, flag := range flags {
		if shipmentsUpdateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %q", flag)
		}
	}
}

// TestShipmentsDeleteFlags verifies delete command flags exist
func TestShipmentsDeleteFlags(t *testing.T) {
	flag := shipmentsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

// TestShipmentsGetCmd verifies get command setup
func TestShipmentsGetCmd(t *testing.T) {
	if shipmentsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", shipmentsGetCmd.Use)
	}
}

// TestShipmentsCreateCmd verifies create command setup
func TestShipmentsCreateCmd(t *testing.T) {
	if shipmentsCreateCmd.Use != "create" {
		t.Errorf("expected Use 'create', got %q", shipmentsCreateCmd.Use)
	}
}

// TestShipmentsUpdateCmd verifies update command setup
func TestShipmentsUpdateCmd(t *testing.T) {
	if shipmentsUpdateCmd.Use != "update <id>" {
		t.Errorf("expected Use 'update <id>', got %q", shipmentsUpdateCmd.Use)
	}
}

// TestShipmentsDeleteCmd verifies delete command setup
func TestShipmentsDeleteCmd(t *testing.T) {
	if shipmentsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", shipmentsDeleteCmd.Use)
	}
}

// TestShipmentsListRunE_GetClientFails verifies error handling when getClient fails
func TestShipmentsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("fulfillment-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := shipmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestShipmentsGetRunE_GetClientFails verifies error handling when getClient fails
func TestShipmentsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestShipmentsCreateRunE_GetClientFails verifies error handling when getClient fails
func TestShipmentsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("order-id", "ord_123", "")
	cmd.Flags().String("fulfillment-id", "ful_123", "")
	cmd.Flags().String("tracking-company", "UPS", "")
	cmd.Flags().String("tracking-number", "1Z999AA10123456784", "")
	cmd.Flags().String("tracking-url", "", "")

	err := shipmentsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestShipmentsUpdateRunE_GetClientFails verifies error handling when getClient fails
func TestShipmentsUpdateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("tracking-company", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().String("tracking-url", "", "")
	cmd.Flags().String("status", "", "")

	err := shipmentsUpdateCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestShipmentsDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestShipmentsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// TestShipmentsListRunE_NoProfiles verifies error handling when no profiles exist
func TestShipmentsListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("fulfillment-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := shipmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// shipmentsAPIClient is a mock implementation for shipments testing.
type shipmentsAPIClient struct {
	api.MockClient

	listShipmentsResp *api.ShipmentsListResponse
	listShipmentsErr  error

	getShipmentResp *api.Shipment
	getShipmentErr  error

	createShipmentResp *api.Shipment
	createShipmentErr  error

	updateShipmentResp *api.Shipment
	updateShipmentErr  error

	deleteShipmentErr error
}

func (m *shipmentsAPIClient) ListShipments(ctx context.Context, opts *api.ShipmentsListOptions) (*api.ShipmentsListResponse, error) {
	return m.listShipmentsResp, m.listShipmentsErr
}

func (m *shipmentsAPIClient) GetShipment(ctx context.Context, id string) (*api.Shipment, error) {
	return m.getShipmentResp, m.getShipmentErr
}

func (m *shipmentsAPIClient) CreateShipment(ctx context.Context, req *api.ShipmentCreateRequest) (*api.Shipment, error) {
	return m.createShipmentResp, m.createShipmentErr
}

func (m *shipmentsAPIClient) UpdateShipment(ctx context.Context, id string, req *api.ShipmentUpdateRequest) (*api.Shipment, error) {
	return m.updateShipmentResp, m.updateShipmentErr
}

func (m *shipmentsAPIClient) DeleteShipment(ctx context.Context, id string) error {
	return m.deleteShipmentErr
}

// setupShipmentsTest sets up the test environment for shipments commands.
func setupShipmentsTest(t *testing.T, mockClient *shipmentsAPIClient) (cleanup func()) {
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

// newShipmentsListTestCmd creates a command with standard flags for list tests.
func newShipmentsListTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "", "")
	cmd.Flags().String("fulfillment-id", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	return cmd
}

func TestShipmentsListRunE_Success(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		listShipmentsResp: &api.ShipmentsListResponse{
			Items: []api.Shipment{
				{
					ID:              "shp_123",
					OrderID:         "ord_456",
					FulfillmentID:   "ff_789",
					TrackingCompany: "FedEx",
					TrackingNumber:  "1234567890",
					Status:          "in_transit",
					CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newShipmentsListTestCmd()
	err := shipmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "shp_123") {
		t.Errorf("output should contain 'shp_123', got: %s", output)
	}
}

func TestShipmentsListRunE_APIError(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		listShipmentsErr: errors.New("API unavailable"),
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	cmd := newShipmentsListTestCmd()
	err := shipmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list shipments") {
		t.Errorf("error should contain 'failed to list shipments', got: %v", err)
	}
}

func TestShipmentsListRunE_EmptyList(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		listShipmentsResp: &api.ShipmentsListResponse{
			Items:      []api.Shipment{},
			TotalCount: 0,
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newShipmentsListTestCmd()
	err := shipmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShipmentsListRunE_JSONOutput(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		listShipmentsResp: &api.ShipmentsListResponse{
			Items: []api.Shipment{
				{
					ID:              "shp_123",
					OrderID:         "ord_456",
					FulfillmentID:   "ff_789",
					TrackingCompany: "FedEx",
					TrackingNumber:  "1234567890",
					Status:          "in_transit",
					CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newShipmentsListTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := shipmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"shp_123"`) && !strings.Contains(output, `"id": "shp_123"`) {
		t.Errorf("JSON output should contain shipment ID, got: %s", output)
	}
}

func TestShipmentsListRunE_WithFilters(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		listShipmentsResp: &api.ShipmentsListResponse{
			Items: []api.Shipment{
				{
					ID:              "shp_123",
					OrderID:         "ord_456",
					FulfillmentID:   "ff_789",
					TrackingCompany: "FedEx",
					TrackingNumber:  "TRACK123",
					Status:          "in_transit",
					CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newShipmentsListTestCmd()
	_ = cmd.Flags().Set("order-id", "ord_456")
	_ = cmd.Flags().Set("fulfillment-id", "ff_789")
	_ = cmd.Flags().Set("status", "in_transit")
	_ = cmd.Flags().Set("tracking-number", "TRACK123")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := shipmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "shp_123") {
		t.Errorf("output should contain 'shp_123', got: %s", output)
	}
}

func TestShipmentsGetRunE_Success(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		getShipmentResp: &api.Shipment{
			ID:              "shp_123",
			OrderID:         "ord_456",
			FulfillmentID:   "ff_789",
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			Status:          "delivered",
			CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShipmentsGetRunE_APIError(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		getShipmentErr: errors.New("shipment not found"),
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get shipment") {
		t.Errorf("error should contain 'failed to get shipment', got: %v", err)
	}
}

func TestShipmentsGetRunE_JSONOutput(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		getShipmentResp: &api.Shipment{
			ID:              "shp_123",
			OrderID:         "ord_456",
			FulfillmentID:   "ff_789",
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			Status:          "delivered",
			CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
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

	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"shp_123"`) && !strings.Contains(output, `"id": "shp_123"`) {
		t.Errorf("JSON output should contain shipment ID, got: %s", output)
	}
}

func TestShipmentsGetRunE_WithTrackingURL(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		getShipmentResp: &api.Shipment{
			ID:              "shp_123",
			OrderID:         "ord_456",
			FulfillmentID:   "ff_789",
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			TrackingURL:     "https://fedex.com/track/1234567890",
			Status:          "in_transit",
			CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Tracking URL:") {
		t.Errorf("output should contain 'Tracking URL:', got: %s", output)
	}
}

func TestShipmentsGetRunE_WithEstimatedDelivery(t *testing.T) {
	estimatedDelivery := time.Date(2024, 1, 20, 14, 0, 0, 0, time.UTC)
	mockClient := &shipmentsAPIClient{
		getShipmentResp: &api.Shipment{
			ID:                "shp_123",
			OrderID:           "ord_456",
			FulfillmentID:     "ff_789",
			TrackingCompany:   "FedEx",
			TrackingNumber:    "1234567890",
			Status:            "in_transit",
			EstimatedDelivery: estimatedDelivery,
			CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:         time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Estimated Delivery:") {
		t.Errorf("output should contain 'Estimated Delivery:', got: %s", output)
	}
}

func TestShipmentsGetRunE_WithDeliveredAt(t *testing.T) {
	deliveredAt := time.Date(2024, 1, 18, 11, 30, 0, 0, time.UTC)
	mockClient := &shipmentsAPIClient{
		getShipmentResp: &api.Shipment{
			ID:              "shp_123",
			OrderID:         "ord_456",
			FulfillmentID:   "ff_789",
			TrackingCompany: "FedEx",
			TrackingNumber:  "1234567890",
			Status:          "delivered",
			DeliveredAt:     deliveredAt,
			CreatedAt:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC),
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")

	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Delivered At:") {
		t.Errorf("output should contain 'Delivered At:', got: %s", output)
	}
}

func TestShipmentsCreateRunE_Success(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		createShipmentResp: &api.Shipment{
			ID:              "shp_new",
			OrderID:         "ord_123",
			FulfillmentID:   "ff_456",
			TrackingCompany: "UPS",
			TrackingNumber:  "9999999999",
			Status:          "pending",
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "ord_123", "")
	cmd.Flags().String("fulfillment-id", "ff_456", "")
	cmd.Flags().String("tracking-company", "UPS", "")
	cmd.Flags().String("tracking-number", "9999999999", "")
	cmd.Flags().String("tracking-url", "", "")

	err := shipmentsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShipmentsCreateRunE_APIError(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		createShipmentErr: errors.New("invalid request"),
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "ord_123", "")
	cmd.Flags().String("fulfillment-id", "ff_456", "")
	cmd.Flags().String("tracking-company", "UPS", "")
	cmd.Flags().String("tracking-number", "9999999999", "")
	cmd.Flags().String("tracking-url", "", "")

	err := shipmentsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create shipment") {
		t.Errorf("error should contain 'failed to create shipment', got: %v", err)
	}
}

func TestShipmentsCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		createShipmentResp: &api.Shipment{
			ID:              "shp_new",
			OrderID:         "ord_123",
			FulfillmentID:   "ff_456",
			TrackingCompany: "UPS",
			TrackingNumber:  "9999999999",
			Status:          "pending",
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("order-id", "ord_123", "")
	cmd.Flags().String("fulfillment-id", "ff_456", "")
	cmd.Flags().String("tracking-company", "UPS", "")
	cmd.Flags().String("tracking-number", "9999999999", "")
	cmd.Flags().String("tracking-url", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := shipmentsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"shp_new"`) && !strings.Contains(output, `"id": "shp_new"`) {
		t.Errorf("JSON output should contain shipment ID, got: %s", output)
	}
}

func TestShipmentsUpdateRunE_Success(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		updateShipmentResp: &api.Shipment{
			ID:              "shp_123",
			TrackingCompany: "DHL",
			TrackingNumber:  "NEW12345",
			Status:          "in_transit",
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("tracking-company", "DHL", "")
	cmd.Flags().String("tracking-number", "NEW12345", "")
	cmd.Flags().String("tracking-url", "", "")
	cmd.Flags().String("status", "in_transit", "")

	err := shipmentsUpdateCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShipmentsUpdateRunE_APIError(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		updateShipmentErr: errors.New("shipment not found"),
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("tracking-company", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().String("tracking-url", "", "")
	cmd.Flags().String("status", "", "")

	err := shipmentsUpdateCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update shipment") {
		t.Errorf("error should contain 'failed to update shipment', got: %v", err)
	}
}

func TestShipmentsUpdateRunE_JSONOutput(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		updateShipmentResp: &api.Shipment{
			ID:              "shp_123",
			OrderID:         "ord_456",
			FulfillmentID:   "ff_789",
			TrackingCompany: "DHL",
			TrackingNumber:  "NEW12345",
			Status:          "in_transit",
		},
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("tracking-company", "DHL", "")
	cmd.Flags().String("tracking-number", "NEW12345", "")
	cmd.Flags().String("tracking-url", "", "")
	cmd.Flags().String("status", "in_transit", "")
	_ = cmd.Flags().Set("output", "json")

	err := shipmentsUpdateCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"shp_123"`) && !strings.Contains(output, `"id": "shp_123"`) {
		t.Errorf("JSON output should contain shipment ID, got: %s", output)
	}
}

func TestShipmentsDeleteRunE_Success(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		deleteShipmentErr: nil,
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", true, "")

	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShipmentsDeleteRunE_APIError(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		deleteShipmentErr: errors.New("shipment not found"),
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", true, "")

	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to delete shipment") {
		t.Errorf("error should contain 'failed to delete shipment', got: %v", err)
	}
}

func TestShipmentsDeleteRunE_Cancelled(t *testing.T) {
	mockClient := &shipmentsAPIClient{}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	// Create a pipe to simulate "n" input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write "n" to simulate user declining
	go func() {
		_, _ = w.WriteString("n\n")
		_ = w.Close()
	}()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "")

	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	os.Stdin = oldStdin

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Cancelled") {
		t.Errorf("output should contain 'Cancelled', got: %s", output)
	}
}

func TestShipmentsDeleteRunE_ConfirmedWithInput(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		deleteShipmentErr: nil,
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	// Create a pipe to simulate "y" input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write "y" to simulate user confirming
	go func() {
		_, _ = w.WriteString("y\n")
		_ = w.Close()
	}()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "")

	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	os.Stdin = oldStdin

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted shipment shp_123") {
		t.Errorf("output should contain 'Deleted shipment shp_123', got: %s", output)
	}
}

func TestShipmentsDeleteRunE_ConfirmedWithUpperY(t *testing.T) {
	mockClient := &shipmentsAPIClient{
		deleteShipmentErr: nil,
	}

	cleanup := setupShipmentsTest(t, mockClient)
	defer cleanup()

	var buf bytes.Buffer
	formatterWriter = &buf

	// Create a pipe to simulate "Y" input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write "Y" to simulate user confirming
	go func() {
		_, _ = w.WriteString("Y\n")
		_ = w.Close()
	}()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", false, "")

	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	os.Stdin = oldStdin

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Deleted shipment shp_123") {
		t.Errorf("output should contain 'Deleted shipment shp_123', got: %s", output)
	}
}

func TestShipmentsGetRunE_NoProfiles(t *testing.T) {
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
	err := shipmentsGetCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestShipmentsCreateRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("order-id", "ord_123", "")
	cmd.Flags().String("fulfillment-id", "ff_456", "")
	cmd.Flags().String("tracking-company", "UPS", "")
	cmd.Flags().String("tracking-number", "9999999999", "")
	cmd.Flags().String("tracking-url", "", "")

	err := shipmentsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestShipmentsUpdateRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("tracking-company", "", "")
	cmd.Flags().String("tracking-number", "", "")
	cmd.Flags().String("tracking-url", "", "")
	cmd.Flags().String("status", "", "")

	err := shipmentsUpdateCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestShipmentsDeleteRunE_NoProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true")

	err := shipmentsDeleteCmd.RunE(cmd, []string{"shp_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

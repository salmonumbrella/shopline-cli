package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

func TestFulfillmentOrdersCmd(t *testing.T) {
	if fulfillmentOrdersCmd.Use != "fulfillment-orders" {
		t.Errorf("Expected Use 'fulfillment-orders', got %q", fulfillmentOrdersCmd.Use)
	}
	if fulfillmentOrdersCmd.Short != "Manage fulfillment orders" {
		t.Errorf("Expected Short 'Manage fulfillment orders', got %q", fulfillmentOrdersCmd.Short)
	}
}

func TestFulfillmentOrdersSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List fulfillment orders",
		"get":    "Get fulfillment order details",
		"move":   "Move a fulfillment order to a new location",
		"cancel": "Cancel a fulfillment order",
		"close":  "Close a fulfillment order",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range fulfillmentOrdersCmd.Commands() {
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

func TestFulfillmentOrdersListCmd(t *testing.T) {
	if fulfillmentOrdersListCmd.Use != "list" {
		t.Errorf("Expected Use 'list', got %q", fulfillmentOrdersListCmd.Use)
	}
}

func TestFulfillmentOrdersGetCmd(t *testing.T) {
	if fulfillmentOrdersGetCmd.Use != "get <id>" {
		t.Errorf("Expected Use 'get <id>', got %q", fulfillmentOrdersGetCmd.Use)
	}
}

func TestFulfillmentOrdersMoveCmd(t *testing.T) {
	if fulfillmentOrdersMoveCmd.Use != "move <id>" {
		t.Errorf("Expected Use 'move <id>', got %q", fulfillmentOrdersMoveCmd.Use)
	}
}

func TestFulfillmentOrdersCancelCmd(t *testing.T) {
	if fulfillmentOrdersCancelCmd.Use != "cancel <id>" {
		t.Errorf("Expected Use 'cancel <id>', got %q", fulfillmentOrdersCancelCmd.Use)
	}
}

func TestFulfillmentOrdersCloseCmd(t *testing.T) {
	if fulfillmentOrdersCloseCmd.Use != "close <id>" {
		t.Errorf("Expected Use 'close <id>', got %q", fulfillmentOrdersCloseCmd.Use)
	}
}

func TestFulfillmentOrdersListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"status", ""},
		{"order-id", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := fulfillmentOrdersListCmd.Flags().Lookup(f.name)
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

func TestFulfillmentOrdersMoveFlags(t *testing.T) {
	flag := fulfillmentOrdersMoveCmd.Flags().Lookup("location-id")
	if flag == nil {
		t.Errorf("Expected flag 'location-id'")
	}
}

func TestFulfillmentOrdersCancelFlags(t *testing.T) {
	flag := fulfillmentOrdersCancelCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected flag 'yes'")
	}
	if flag.DefValue != "false" {
		t.Errorf("Expected default 'false', got %q", flag.DefValue)
	}
}

func TestFulfillmentOrdersCloseFlags(t *testing.T) {
	flag := fulfillmentOrdersCloseCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("Expected flag 'yes'")
	}
	if flag.DefValue != "false" {
		t.Errorf("Expected default 'false', got %q", flag.DefValue)
	}
}

func TestFulfillmentOrdersListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	err := fulfillmentOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentOrdersGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	err := fulfillmentOrdersGetCmd.RunE(cmd, []string{"fo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentOrdersMoveRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("location-id", "loc_123", "")
	err := fulfillmentOrdersMoveCmd.RunE(cmd, []string{"fo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentOrdersCancelRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	err := fulfillmentOrdersCancelCmd.RunE(cmd, []string{"fo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentOrdersCloseRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()
	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	err := fulfillmentOrdersCloseCmd.RunE(cmd, []string{"fo_123"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestFulfillmentOrdersListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	err := fulfillmentOrdersListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

// fulfillmentOrdersAPIClient is a mock implementation for fulfillment orders testing.
type fulfillmentOrdersAPIClient struct {
	api.MockClient

	listFulfillmentOrdersResp *api.FulfillmentOrdersListResponse
	listFulfillmentOrdersErr  error

	getFulfillmentOrderResp *api.FulfillmentOrder
	getFulfillmentOrderErr  error

	moveFulfillmentOrderResp *api.FulfillmentOrder
	moveFulfillmentOrderErr  error

	cancelFulfillmentOrderResp *api.FulfillmentOrder
	cancelFulfillmentOrderErr  error

	closeFulfillmentOrderResp *api.FulfillmentOrder
	closeFulfillmentOrderErr  error
}

func (m *fulfillmentOrdersAPIClient) ListFulfillmentOrders(ctx context.Context, opts *api.FulfillmentOrdersListOptions) (*api.FulfillmentOrdersListResponse, error) {
	return m.listFulfillmentOrdersResp, m.listFulfillmentOrdersErr
}

func (m *fulfillmentOrdersAPIClient) GetFulfillmentOrder(ctx context.Context, id string) (*api.FulfillmentOrder, error) {
	return m.getFulfillmentOrderResp, m.getFulfillmentOrderErr
}

func (m *fulfillmentOrdersAPIClient) MoveFulfillmentOrder(ctx context.Context, id string, newLocationID string) (*api.FulfillmentOrder, error) {
	return m.moveFulfillmentOrderResp, m.moveFulfillmentOrderErr
}

func (m *fulfillmentOrdersAPIClient) CancelFulfillmentOrder(ctx context.Context, id string) (*api.FulfillmentOrder, error) {
	return m.cancelFulfillmentOrderResp, m.cancelFulfillmentOrderErr
}

func (m *fulfillmentOrdersAPIClient) CloseFulfillmentOrder(ctx context.Context, id string) (*api.FulfillmentOrder, error) {
	return m.closeFulfillmentOrderResp, m.closeFulfillmentOrderErr
}

// setupFulfillmentOrdersTest is a helper function to setup mock factories for fulfillment orders tests.
func setupFulfillmentOrdersTest(mockClient *fulfillmentOrdersAPIClient) (cleanup func(), buf *bytes.Buffer) {
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

	buf = &bytes.Buffer{}
	formatterWriter = buf

	cleanup = func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}

	return cleanup, buf
}

// newFulfillmentOrdersListTestCmd creates a test command with flags for list.
func newFulfillmentOrdersListTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("order-id", "", "")
	return cmd
}

// newFulfillmentOrdersGetTestCmd creates a test command with flags for get.
func newFulfillmentOrdersGetTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	return cmd
}

// newFulfillmentOrdersMoveTestCmd creates a test command with flags for move.
func newFulfillmentOrdersMoveTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("location-id", "loc_new", "")
	return cmd
}

// newFulfillmentOrdersCancelTestCmd creates a test command with flags for cancel.
func newFulfillmentOrdersCancelTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

// newFulfillmentOrdersCloseTestCmd creates a test command with flags for close.
func newFulfillmentOrdersCloseTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

func TestFulfillmentOrdersListRunE_Success(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		mockResp   *api.FulfillmentOrdersListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful list table output",
			output: "",
			mockResp: &api.FulfillmentOrdersListResponse{
				Items: []api.FulfillmentOrder{
					{
						ID:                 "fo_123",
						OrderID:            "ord_456",
						Status:             "open",
						FulfillmentStatus:  "unfulfilled",
						AssignedLocationID: "loc_789",
						LineItems: []api.FulfillmentOrderItem{
							{ID: "item_1", Quantity: 2},
							{ID: "item_2", Quantity: 1},
						},
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "fo_123",
		},
		{
			name:   "successful list JSON output",
			output: "json",
			mockResp: &api.FulfillmentOrdersListResponse{
				Items: []api.FulfillmentOrder{
					{
						ID:                 "fo_123",
						OrderID:            "ord_456",
						Status:             "open",
						FulfillmentStatus:  "unfulfilled",
						AssignedLocationID: "loc_789",
						CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "fo_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.FulfillmentOrdersListResponse{
				Items:      []api.FulfillmentOrder{},
				TotalCount: 0,
			},
		},
		{
			name:   "multiple fulfillment orders",
			output: "",
			mockResp: &api.FulfillmentOrdersListResponse{
				Items: []api.FulfillmentOrder{
					{
						ID:                 "fo_001",
						OrderID:            "ord_001",
						Status:             "open",
						FulfillmentStatus:  "unfulfilled",
						AssignedLocationID: "loc_001",
						CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:                 "fo_002",
						OrderID:            "ord_002",
						Status:             "closed",
						FulfillmentStatus:  "fulfilled",
						AssignedLocationID: "loc_002",
						CreatedAt:          time.Date(2024, 1, 16, 11, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "fo_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentOrdersAPIClient{
				listFulfillmentOrdersResp: tt.mockResp,
				listFulfillmentOrdersErr:  tt.mockErr,
			}
			cleanup, buf := setupFulfillmentOrdersTest(mockClient)
			defer cleanup()

			cmd := newFulfillmentOrdersListTestCmd()
			if tt.output != "" {
				_ = cmd.Flags().Set("output", tt.output)
			}

			err := fulfillmentOrdersListCmd.RunE(cmd, []string{})

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

func TestFulfillmentOrdersGetRunE_Success(t *testing.T) {
	tests := []struct {
		name       string
		foID       string
		output     string
		mockResp   *api.FulfillmentOrder
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful get table output",
			foID:   "fo_123",
			output: "",
			mockResp: &api.FulfillmentOrder{
				ID:                 "fo_123",
				OrderID:            "ord_456",
				Status:             "open",
				FulfillmentStatus:  "unfulfilled",
				AssignedLocationID: "loc_789",
				RequestStatus:      "unsubmitted",
				DeliveryMethod: api.FulfillmentDeliveryMethod{
					MethodType:  "shipping",
					ServiceCode: "standard",
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
			},
			// Text output goes to stdout directly via fmt.Printf, not to the formatter
		},
		{
			name:   "successful get JSON output",
			foID:   "fo_123",
			output: "json",
			mockResp: &api.FulfillmentOrder{
				ID:                 "fo_123",
				OrderID:            "ord_456",
				Status:             "open",
				FulfillmentStatus:  "unfulfilled",
				AssignedLocationID: "loc_789",
				CreatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:          time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
			},
			wantOutput: "fo_123",
		},
		{
			name:   "successful get with line items",
			foID:   "fo_123",
			output: "",
			mockResp: &api.FulfillmentOrder{
				ID:                 "fo_123",
				OrderID:            "ord_456",
				Status:             "open",
				FulfillmentStatus:  "unfulfilled",
				AssignedLocationID: "loc_789",
				RequestStatus:      "unsubmitted",
				DeliveryMethod: api.FulfillmentDeliveryMethod{
					MethodType:  "shipping",
					ServiceCode: "express",
				},
				LineItems: []api.FulfillmentOrderItem{
					{
						ID:                  "item_001",
						LineItemID:          "li_001",
						VariantID:           "var_001",
						Quantity:            3,
						FulfillableQuantity: 2,
						FulfilledQuantity:   1,
					},
					{
						ID:                  "item_002",
						LineItemID:          "li_002",
						VariantID:           "var_002",
						Quantity:            5,
						FulfillableQuantity: 5,
						FulfilledQuantity:   0,
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
			},
			// Text output with line items goes to stdout directly
		},
		{
			name:    "fulfillment order not found",
			foID:    "fo_999",
			mockErr: errors.New("fulfillment order not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentOrdersAPIClient{
				getFulfillmentOrderResp: tt.mockResp,
				getFulfillmentOrderErr:  tt.mockErr,
			}
			cleanup, buf := setupFulfillmentOrdersTest(mockClient)
			defer cleanup()

			cmd := newFulfillmentOrdersGetTestCmd()
			if tt.output != "" {
				_ = cmd.Flags().Set("output", tt.output)
			}

			err := fulfillmentOrdersGetCmd.RunE(cmd, []string{tt.foID})

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

			// Only check buffer output for JSON format (formatter output)
			if tt.output == "json" && tt.wantOutput != "" {
				output := buf.String()
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("output %q should contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestFulfillmentOrdersMoveRunE_Success(t *testing.T) {
	tests := []struct {
		name       string
		foID       string
		locationID string
		mockResp   *api.FulfillmentOrder
		mockErr    error
		wantErr    bool
	}{
		{
			name:       "successful move",
			foID:       "fo_123",
			locationID: "loc_new",
			mockResp: &api.FulfillmentOrder{
				ID:                 "fo_123",
				OrderID:            "ord_456",
				Status:             "open",
				AssignedLocationID: "loc_new",
			},
		},
		{
			name:       "move fails",
			foID:       "fo_999",
			locationID: "loc_invalid",
			mockErr:    errors.New("cannot move fulfillment order"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentOrdersAPIClient{
				moveFulfillmentOrderResp: tt.mockResp,
				moveFulfillmentOrderErr:  tt.mockErr,
			}
			cleanup, _ := setupFulfillmentOrdersTest(mockClient)
			defer cleanup()

			cmd := newFulfillmentOrdersMoveTestCmd()
			if tt.locationID != "" {
				_ = cmd.Flags().Set("location-id", tt.locationID)
			}

			err := fulfillmentOrdersMoveCmd.RunE(cmd, []string{tt.foID})

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

func TestFulfillmentOrdersCancelRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		foID     string
		mockResp *api.FulfillmentOrder
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful cancel",
			foID: "fo_123",
			mockResp: &api.FulfillmentOrder{
				ID:     "fo_123",
				Status: "cancelled",
			},
		},
		{
			name:    "cancel fails",
			foID:    "fo_999",
			mockErr: errors.New("cannot cancel fulfillment order"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentOrdersAPIClient{
				cancelFulfillmentOrderResp: tt.mockResp,
				cancelFulfillmentOrderErr:  tt.mockErr,
			}
			cleanup, _ := setupFulfillmentOrdersTest(mockClient)
			defer cleanup()

			cmd := newFulfillmentOrdersCancelTestCmd()

			err := fulfillmentOrdersCancelCmd.RunE(cmd, []string{tt.foID})

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

func TestFulfillmentOrdersCloseRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		foID     string
		mockResp *api.FulfillmentOrder
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful close",
			foID: "fo_123",
			mockResp: &api.FulfillmentOrder{
				ID:     "fo_123",
				Status: "closed",
			},
		},
		{
			name:    "close fails",
			foID:    "fo_999",
			mockErr: errors.New("cannot close fulfillment order"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentOrdersAPIClient{
				closeFulfillmentOrderResp: tt.mockResp,
				closeFulfillmentOrderErr:  tt.mockErr,
			}
			cleanup, _ := setupFulfillmentOrdersTest(mockClient)
			defer cleanup()

			cmd := newFulfillmentOrdersCloseTestCmd()

			err := fulfillmentOrdersCloseCmd.RunE(cmd, []string{tt.foID})

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

// TestFulfillmentOrdersListWithFilters tests list command with various filter combinations.
func TestFulfillmentOrdersListWithFilters(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		orderID  string
		page     int
		pageSize int
	}{
		{
			name:   "filter by status",
			status: "open",
		},
		{
			name:    "filter by order ID",
			orderID: "ord_123",
		},
		{
			name:     "with pagination",
			page:     2,
			pageSize: 50,
		},
		{
			name:     "all filters combined",
			status:   "closed",
			orderID:  "ord_456",
			page:     3,
			pageSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &fulfillmentOrdersAPIClient{
				listFulfillmentOrdersResp: &api.FulfillmentOrdersListResponse{
					Items:      []api.FulfillmentOrder{},
					TotalCount: 0,
				},
			}
			cleanup, _ := setupFulfillmentOrdersTest(mockClient)
			defer cleanup()

			cmd := newFulfillmentOrdersListTestCmd()
			if tt.status != "" {
				_ = cmd.Flags().Set("status", tt.status)
			}
			if tt.orderID != "" {
				_ = cmd.Flags().Set("order-id", tt.orderID)
			}
			if tt.page > 0 {
				_ = cmd.Flags().Set("page", fmt.Sprintf("%d", tt.page))
			}
			if tt.pageSize > 0 {
				_ = cmd.Flags().Set("page-size", fmt.Sprintf("%d", tt.pageSize))
			}

			err := fulfillmentOrdersListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestFulfillmentOrdersGetWithEmptyLineItems tests get command with no line items.
func TestFulfillmentOrdersGetWithEmptyLineItems(t *testing.T) {
	mockClient := &fulfillmentOrdersAPIClient{
		getFulfillmentOrderResp: &api.FulfillmentOrder{
			ID:                 "fo_empty",
			OrderID:            "ord_empty",
			Status:             "open",
			FulfillmentStatus:  "unfulfilled",
			AssignedLocationID: "loc_empty",
			RequestStatus:      "unsubmitted",
			DeliveryMethod: api.FulfillmentDeliveryMethod{
				MethodType:  "pickup",
				ServiceCode: "in_store",
			},
			LineItems: []api.FulfillmentOrderItem{},
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupFulfillmentOrdersTest(mockClient)
	defer cleanup()

	cmd := newFulfillmentOrdersGetTestCmd()
	err := fulfillmentOrdersGetCmd.RunE(cmd, []string{"fo_empty"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Output goes to stdout via fmt.Printf, not to the formatter buffer
	// We verify that no error occurred
}

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

// pickupMockAPIClient is a mock implementation of api.APIClient for pickup tests.
type pickupMockAPIClient struct {
	api.MockClient

	listPickupLocationsResp  *api.PickupListResponse
	listPickupLocationsErr   error
	getPickupLocationResp    *api.PickupLocation
	getPickupLocationErr     error
	createPickupLocationResp *api.PickupLocation
	createPickupLocationErr  error
	deletePickupLocationErr  error
}

func (m *pickupMockAPIClient) ListPickupLocations(ctx context.Context, opts *api.PickupListOptions) (*api.PickupListResponse, error) {
	return m.listPickupLocationsResp, m.listPickupLocationsErr
}

func (m *pickupMockAPIClient) GetPickupLocation(ctx context.Context, id string) (*api.PickupLocation, error) {
	return m.getPickupLocationResp, m.getPickupLocationErr
}

func (m *pickupMockAPIClient) CreatePickupLocation(ctx context.Context, req *api.PickupCreateRequest) (*api.PickupLocation, error) {
	return m.createPickupLocationResp, m.createPickupLocationErr
}

func (m *pickupMockAPIClient) DeletePickupLocation(ctx context.Context, id string) error {
	return m.deletePickupLocationErr
}

// setupPickupMockFactories sets up the mock factories for pickup tests and returns a cleanup function.
func setupPickupMockFactories(mockClient *pickupMockAPIClient) func() {
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

	var buf bytes.Buffer
	formatterWriter = &buf

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}
}

// newPickupTestCmd creates a test command with standard flags for pickup tests.
func newPickupTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	return cmd
}

func TestPickupCommandStructure(t *testing.T) {
	if pickupCmd == nil {
		t.Fatal("pickupCmd is nil")
	}
	if pickupCmd.Use != "pickup" {
		t.Errorf("Expected Use 'pickup', got %q", pickupCmd.Use)
	}
	if pickupCmd.Short != "Manage store pickup locations" {
		t.Errorf("Expected Short 'Manage store pickup locations', got %q", pickupCmd.Short)
	}
}

func TestPickupSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List pickup locations",
		"get":    "Get pickup location details",
		"create": "Create a pickup location",
		"delete": "Delete a pickup location",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range pickupCmd.Commands() {
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

func TestPickupListFlags(t *testing.T) {
	cmd := pickupListCmd
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"location-id", ""},
		{"active", ""},
		{"page", "1"},
		{"page-size", "20"},
	}
	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("Flag %q not found", f.name)
			} else if flag.DefValue != f.defaultValue {
				t.Errorf("Flag %q default: expected %q, got %q", f.name, f.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestPickupCreateFlags(t *testing.T) {
	flags := []string{"name", "address1", "address2", "city", "province", "country", "zip-code", "phone", "email", "instructions", "active", "location-id"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := pickupCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

func TestPickupGetCmdUse(t *testing.T) {
	if pickupGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", pickupGetCmd.Use)
	}
}

func TestPickupDeleteCmdUse(t *testing.T) {
	if pickupDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", pickupDeleteCmd.Use)
	}
}

func TestPickupGetRequiresArg(t *testing.T) {
	cmd := pickupGetCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"pickup_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestPickupDeleteRequiresArg(t *testing.T) {
	cmd := pickupDeleteCmd
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("Expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"pickup_123"}); err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

// TestPickupListRunE_Success tests the pickup list command execution with mock API.
func TestPickupListRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.PickupListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list",
			mockResp: &api.PickupListResponse{
				Items: []api.PickupLocation{
					{
						ID:        "pickup_123",
						Name:      "Main Store",
						Address1:  "123 Main St",
						Address2:  "Suite 100",
						City:      "New York",
						Country:   "US",
						Active:    true,
						Phone:     "+1234567890",
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.PickupListResponse{
				Items:      []api.PickupLocation{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple locations",
			mockResp: &api.PickupListResponse{
				Items: []api.PickupLocation{
					{
						ID:       "pickup_1",
						Name:     "Store 1",
						Address1: "100 First Ave",
						City:     "Boston",
						Country:  "US",
						Active:   true,
					},
					{
						ID:       "pickup_2",
						Name:     "Store 2",
						Address1: "200 Second Ave",
						Address2: "Floor 2",
						City:     "Chicago",
						Country:  "US",
						Active:   false,
					},
				},
				TotalCount: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pickupMockAPIClient{
				listPickupLocationsResp: tt.mockResp,
				listPickupLocationsErr:  tt.mockErr,
			}
			cleanup := setupPickupMockFactories(mockClient)
			defer cleanup()

			cmd := newPickupTestCmd()
			cmd.Flags().String("location-id", "", "")
			cmd.Flags().String("active", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := pickupListCmd.RunE(cmd, []string{})

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

// TestPickupListRunE_JSONOutput tests the pickup list command with JSON output format.
func TestPickupListRunE_JSONOutput(t *testing.T) {
	mockClient := &pickupMockAPIClient{
		listPickupLocationsResp: &api.PickupListResponse{
			Items: []api.PickupLocation{
				{
					ID:       "pickup_json",
					Name:     "JSON Test Store",
					Address1: "456 JSON St",
					City:     "San Francisco",
					Country:  "US",
					Active:   true,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup := setupPickupMockFactories(mockClient)
	defer cleanup()

	cmd := newPickupTestCmd()
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().String("active", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := pickupListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPickupListRunE_WithFilters tests the pickup list command with various filters.
func TestPickupListRunE_WithFilters(t *testing.T) {
	tests := []struct {
		name       string
		locationID string
		active     string
		page       string
		pageSize   string
	}{
		{
			name:       "with location-id filter",
			locationID: "loc_123",
		},
		{
			name:   "with active=true filter",
			active: "true",
		},
		{
			name:   "with active=false filter",
			active: "false",
		},
		{
			name:     "with pagination",
			page:     "2",
			pageSize: "50",
		},
		{
			name:       "with all filters",
			locationID: "loc_456",
			active:     "true",
			page:       "3",
			pageSize:   "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pickupMockAPIClient{
				listPickupLocationsResp: &api.PickupListResponse{
					Items:      []api.PickupLocation{},
					TotalCount: 0,
				},
			}
			cleanup := setupPickupMockFactories(mockClient)
			defer cleanup()

			cmd := newPickupTestCmd()
			cmd.Flags().String("location-id", "", "")
			cmd.Flags().String("active", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			if tt.locationID != "" {
				_ = cmd.Flags().Set("location-id", tt.locationID)
			}
			if tt.active != "" {
				_ = cmd.Flags().Set("active", tt.active)
			}
			if tt.page != "" {
				_ = cmd.Flags().Set("page", tt.page)
			}
			if tt.pageSize != "" {
				_ = cmd.Flags().Set("page-size", tt.pageSize)
			}

			err := pickupListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestPickupGetRunE_Success tests the pickup get command execution with mock API.
func TestPickupGetRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockResp *api.PickupLocation
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful get",
			id:   "pickup_123",
			mockResp: &api.PickupLocation{
				ID:           "pickup_123",
				Name:         "Main Store",
				Address1:     "123 Main St",
				Address2:     "Suite 100",
				City:         "New York",
				Province:     "NY",
				Country:      "US",
				ZipCode:      "10001",
				Phone:        "+1234567890",
				Email:        "store@example.com",
				Active:       true,
				Instructions: "Enter through the main entrance",
				LocationID:   "loc_456",
				Hours: []api.PickupHours{
					{Day: "Monday", OpenTime: "09:00", CloseTime: "17:00", Closed: false},
					{Day: "Sunday", Closed: true},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "not found",
			id:      "pickup_999",
			mockErr: errors.New("pickup location not found"),
			wantErr: true,
		},
		{
			name: "minimal data",
			id:   "pickup_min",
			mockResp: &api.PickupLocation{
				ID:        "pickup_min",
				Name:      "Minimal Store",
				Address1:  "456 Minimal St",
				City:      "Boston",
				Country:   "US",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pickupMockAPIClient{
				getPickupLocationResp: tt.mockResp,
				getPickupLocationErr:  tt.mockErr,
			}
			cleanup := setupPickupMockFactories(mockClient)
			defer cleanup()

			cmd := newPickupTestCmd()

			err := pickupGetCmd.RunE(cmd, []string{tt.id})

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

// TestPickupGetRunE_JSONOutput tests the pickup get command with JSON output format.
func TestPickupGetRunE_JSONOutput(t *testing.T) {
	mockClient := &pickupMockAPIClient{
		getPickupLocationResp: &api.PickupLocation{
			ID:        "pickup_json",
			Name:      "JSON Store",
			Address1:  "789 JSON Ave",
			City:      "Seattle",
			Country:   "US",
			Active:    true,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup := setupPickupMockFactories(mockClient)
	defer cleanup()

	cmd := newPickupTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := pickupGetCmd.RunE(cmd, []string{"pickup_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPickupGetRunE_WithOptionalFields tests the pickup get command with various optional fields.
func TestPickupGetRunE_WithOptionalFields(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.PickupLocation
	}{
		{
			name: "with address2",
			mockResp: &api.PickupLocation{
				ID:        "pickup_addr2",
				Name:      "Address2 Store",
				Address1:  "100 Main St",
				Address2:  "Floor 2",
				City:      "Chicago",
				Country:   "US",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with province",
			mockResp: &api.PickupLocation{
				ID:        "pickup_prov",
				Name:      "Province Store",
				Address1:  "200 Main St",
				City:      "Los Angeles",
				Province:  "CA",
				Country:   "US",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with zip code",
			mockResp: &api.PickupLocation{
				ID:        "pickup_zip",
				Name:      "ZIP Store",
				Address1:  "300 Main St",
				City:      "Miami",
				Country:   "US",
				ZipCode:   "33101",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with phone",
			mockResp: &api.PickupLocation{
				ID:        "pickup_phone",
				Name:      "Phone Store",
				Address1:  "400 Main St",
				City:      "Denver",
				Country:   "US",
				Phone:     "+1-555-123-4567",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with email",
			mockResp: &api.PickupLocation{
				ID:        "pickup_email",
				Name:      "Email Store",
				Address1:  "500 Main St",
				City:      "Austin",
				Country:   "US",
				Email:     "store@example.com",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with instructions",
			mockResp: &api.PickupLocation{
				ID:           "pickup_inst",
				Name:         "Instructions Store",
				Address1:     "600 Main St",
				City:         "Portland",
				Country:      "US",
				Instructions: "Park in the back lot and enter through door #3",
				Active:       true,
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with location ID",
			mockResp: &api.PickupLocation{
				ID:         "pickup_loc",
				Name:       "Location Store",
				Address1:   "700 Main St",
				City:       "Phoenix",
				Country:    "US",
				LocationID: "loc_linked_123",
				Active:     true,
				CreatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "with operating hours",
			mockResp: &api.PickupLocation{
				ID:       "pickup_hours",
				Name:     "Hours Store",
				Address1: "800 Main St",
				City:     "Philadelphia",
				Country:  "US",
				Hours: []api.PickupHours{
					{Day: "Monday", OpenTime: "09:00", CloseTime: "18:00", Closed: false},
					{Day: "Tuesday", OpenTime: "09:00", CloseTime: "18:00", Closed: false},
					{Day: "Sunday", Closed: true},
				},
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pickupMockAPIClient{
				getPickupLocationResp: tt.mockResp,
			}
			cleanup := setupPickupMockFactories(mockClient)
			defer cleanup()

			cmd := newPickupTestCmd()

			err := pickupGetCmd.RunE(cmd, []string{tt.mockResp.ID})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestPickupCreateRunE_Success tests the pickup create command execution with mock API.
func TestPickupCreateRunE_Success(t *testing.T) {
	tests := []struct {
		name     string
		mockResp *api.PickupLocation
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.PickupLocation{
				ID:        "pickup_new",
				Name:      "New Store",
				Address1:  "123 New St",
				City:      "New York",
				Country:   "US",
				Active:    true,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "create fails",
			mockErr: errors.New("validation error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pickupMockAPIClient{
				createPickupLocationResp: tt.mockResp,
				createPickupLocationErr:  tt.mockErr,
			}
			cleanup := setupPickupMockFactories(mockClient)
			defer cleanup()

			cmd := newPickupTestCmd()
			cmd.Flags().String("name", "New Store", "")
			cmd.Flags().String("address1", "123 New St", "")
			cmd.Flags().String("address2", "", "")
			cmd.Flags().String("city", "New York", "")
			cmd.Flags().String("province", "", "")
			cmd.Flags().String("country", "US", "")
			cmd.Flags().String("zip-code", "", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().String("email", "", "")
			cmd.Flags().String("instructions", "", "")
			cmd.Flags().Bool("active", true, "")
			cmd.Flags().String("location-id", "", "")

			err := pickupCreateCmd.RunE(cmd, []string{})

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

// TestPickupCreateRunE_JSONOutput tests the pickup create command with JSON output format.
func TestPickupCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &pickupMockAPIClient{
		createPickupLocationResp: &api.PickupLocation{
			ID:        "pickup_json_new",
			Name:      "JSON New Store",
			Address1:  "456 JSON St",
			City:      "Boston",
			Country:   "US",
			Active:    true,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup := setupPickupMockFactories(mockClient)
	defer cleanup()

	cmd := newPickupTestCmd()
	cmd.Flags().String("name", "JSON New Store", "")
	cmd.Flags().String("address1", "456 JSON St", "")
	cmd.Flags().String("address2", "", "")
	cmd.Flags().String("city", "Boston", "")
	cmd.Flags().String("province", "", "")
	cmd.Flags().String("country", "US", "")
	cmd.Flags().String("zip-code", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("location-id", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := pickupCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPickupCreateRunE_WithAllOptions tests the pickup create command with all options.
func TestPickupCreateRunE_WithAllOptions(t *testing.T) {
	mockClient := &pickupMockAPIClient{
		createPickupLocationResp: &api.PickupLocation{
			ID:           "pickup_full",
			Name:         "Full Store",
			Address1:     "123 Full St",
			Address2:     "Suite 500",
			City:         "San Francisco",
			Province:     "CA",
			Country:      "US",
			ZipCode:      "94102",
			Phone:        "+1-555-987-6543",
			Email:        "full@example.com",
			Instructions: "Ring the bell",
			LocationID:   "loc_789",
			Active:       true,
			CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup := setupPickupMockFactories(mockClient)
	defer cleanup()

	cmd := newPickupTestCmd()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("address1", "", "")
	cmd.Flags().String("address2", "", "")
	cmd.Flags().String("city", "", "")
	cmd.Flags().String("province", "", "")
	cmd.Flags().String("country", "", "")
	cmd.Flags().String("zip-code", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("location-id", "", "")

	_ = cmd.Flags().Set("name", "Full Store")
	_ = cmd.Flags().Set("address1", "123 Full St")
	_ = cmd.Flags().Set("address2", "Suite 500")
	_ = cmd.Flags().Set("city", "San Francisco")
	_ = cmd.Flags().Set("province", "CA")
	_ = cmd.Flags().Set("country", "US")
	_ = cmd.Flags().Set("zip-code", "94102")
	_ = cmd.Flags().Set("phone", "+1-555-987-6543")
	_ = cmd.Flags().Set("email", "full@example.com")
	_ = cmd.Flags().Set("instructions", "Ring the bell")
	_ = cmd.Flags().Set("location-id", "loc_789")

	err := pickupCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPickupDeleteRunE_Success tests the pickup delete command execution with mock API.
func TestPickupDeleteRunE_Success(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name: "successful delete",
			id:   "pickup_123",
		},
		{
			name:    "delete fails",
			id:      "pickup_456",
			mockErr: errors.New("pickup location in use"),
			wantErr: true,
		},
		{
			name:    "not found",
			id:      "pickup_999",
			mockErr: errors.New("pickup location not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &pickupMockAPIClient{
				deletePickupLocationErr: tt.mockErr,
			}
			cleanup := setupPickupMockFactories(mockClient)
			defer cleanup()

			cmd := newPickupTestCmd()
			cmd.Flags().Bool("yes", true, "")

			err := pickupDeleteCmd.RunE(cmd, []string{tt.id})

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

// TestPickupListRunE_GetClientFails verifies error handling when getClient fails
func TestPickupListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().String("active", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := pickupListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPickupGetRunE_GetClientFails verifies error handling when getClient fails
func TestPickupGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := pickupGetCmd.RunE(cmd, []string{"pickup_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPickupCreateRunE_GetClientFails verifies error handling when getClient fails
func TestPickupCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "Test", "")
	cmd.Flags().String("address1", "123 Test", "")
	cmd.Flags().String("address2", "", "")
	cmd.Flags().String("city", "Test City", "")
	cmd.Flags().String("province", "", "")
	cmd.Flags().String("country", "US", "")
	cmd.Flags().String("zip-code", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("location-id", "", "")

	err := pickupCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPickupDeleteRunE_GetClientFails verifies error handling when getClient fails
func TestPickupDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := pickupDeleteCmd.RunE(cmd, []string{"pickup_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestPickupListRunE_NoProfiles verifies error when no profiles are configured
func TestPickupListRunE_NoProfiles(t *testing.T) {
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
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().String("active", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := pickupListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
	if !strings.Contains(err.Error(), "no store profiles configured") {
		t.Errorf("Expected 'no store profiles' error, got: %v", err)
	}
}

// TestPickupGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestPickupGetRunE_MultipleProfiles(t *testing.T) {
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
	err := pickupGetCmd.RunE(cmd, []string{"pickup_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestPickupListWithEnvVar verifies store selection from environment variable
func TestPickupListWithEnvVar(t *testing.T) {
	origFactory := secretsStoreFactory
	origClientFactory := clientFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	origWriter := formatterWriter
	defer func() {
		secretsStoreFactory = origFactory
		clientFactory = origClientFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
		formatterWriter = origWriter
	}()

	_ = os.Setenv("SHOPLINE_STORE", "envstore")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"envstore", "other"},
			creds: map[string]*secrets.StoreCredentials{
				"envstore": {Handle: "test", AccessToken: "token123"},
				"other":    {Handle: "other", AccessToken: "token456"},
			},
		}, nil
	}

	mockClient := &pickupMockAPIClient{
		listPickupLocationsResp: &api.PickupListResponse{
			Items:      []api.PickupLocation{},
			TotalCount: 0,
		},
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("location-id", "", "")
	cmd.Flags().String("active", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := pickupListCmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "multiple profiles") {
		t.Errorf("Should have selected store from env var, got: %v", err)
	}
}

// TestPickupCreateRunE_WithInactiveStatus tests creating an inactive pickup location.
func TestPickupCreateRunE_WithInactiveStatus(t *testing.T) {
	mockClient := &pickupMockAPIClient{
		createPickupLocationResp: &api.PickupLocation{
			ID:        "pickup_inactive",
			Name:      "Inactive Store",
			Address1:  "123 Inactive St",
			City:      "Chicago",
			Country:   "US",
			Active:    false,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	cleanup := setupPickupMockFactories(mockClient)
	defer cleanup()

	cmd := newPickupTestCmd()
	cmd.Flags().String("name", "Inactive Store", "")
	cmd.Flags().String("address1", "123 Inactive St", "")
	cmd.Flags().String("address2", "", "")
	cmd.Flags().String("city", "Chicago", "")
	cmd.Flags().String("province", "", "")
	cmd.Flags().String("country", "US", "")
	cmd.Flags().String("zip-code", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("instructions", "", "")
	cmd.Flags().Bool("active", true, "")
	cmd.Flags().String("location-id", "", "")
	_ = cmd.Flags().Set("active", "false")

	err := pickupCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

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

// marketingEventsMockAPIClient is a mock implementation of api.APIClient for marketing events tests.
type marketingEventsMockAPIClient struct {
	api.MockClient
	listMarketingEventsResp  *api.MarketingEventsListResponse
	listMarketingEventsErr   error
	getMarketingEventResp    *api.MarketingEvent
	getMarketingEventErr     error
	createMarketingEventResp *api.MarketingEvent
	createMarketingEventErr  error
	deleteMarketingEventErr  error
}

func (m *marketingEventsMockAPIClient) ListMarketingEvents(ctx context.Context, opts *api.MarketingEventsListOptions) (*api.MarketingEventsListResponse, error) {
	return m.listMarketingEventsResp, m.listMarketingEventsErr
}

func (m *marketingEventsMockAPIClient) GetMarketingEvent(ctx context.Context, id string) (*api.MarketingEvent, error) {
	return m.getMarketingEventResp, m.getMarketingEventErr
}

func (m *marketingEventsMockAPIClient) CreateMarketingEvent(ctx context.Context, req *api.MarketingEventCreateRequest) (*api.MarketingEvent, error) {
	return m.createMarketingEventResp, m.createMarketingEventErr
}

func (m *marketingEventsMockAPIClient) DeleteMarketingEvent(ctx context.Context, id string) error {
	return m.deleteMarketingEventErr
}

// setupMarketingEventsMockFactories sets up mock factories for marketing events tests.
func setupMarketingEventsMockFactories(mockClient *marketingEventsMockAPIClient) (func(), *bytes.Buffer) {
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

// newMarketingEventsTestCmd creates a test command with common flags for marketing events tests.
func newMarketingEventsTestCmd() *cobra.Command {
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

// TestMarketingEventsListRunE tests the marketing events list command with mock API.
func TestMarketingEventsListRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.MarketingEventsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list with budget",
			mockResp: &api.MarketingEventsListResponse{
				Items: []api.MarketingEvent{
					{
						ID:            "event_123",
						EventType:     "campaign",
						MarketingType: "email",
						UTMCampaign:   "summer_sale",
						Budget:        1000.00,
						Currency:      "USD",
						CreatedAt:     fixedTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "event_123",
		},
		{
			name: "successful list without budget",
			mockResp: &api.MarketingEventsListResponse{
				Items: []api.MarketingEvent{
					{
						ID:            "event_456",
						EventType:     "ad",
						MarketingType: "cpc",
						UTMCampaign:   "winter_promo",
						Budget:        0,
						Currency:      "",
						CreatedAt:     fixedTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "event_456",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.MarketingEventsListResponse{
				Items:      []api.MarketingEvent{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketingEventsMockAPIClient{
				listMarketingEventsResp: tt.mockResp,
				listMarketingEventsErr:  tt.mockErr,
			}
			cleanup, buf := setupMarketingEventsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketingEventsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")
			cmd.Flags().String("event-type", "", "")
			cmd.Flags().String("marketing-type", "", "")

			err := marketingEventsListCmd.RunE(cmd, []string{})

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

// TestMarketingEventsListRunEJSON tests the list command with JSON output.
func TestMarketingEventsListRunEJSON(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &marketingEventsMockAPIClient{
		listMarketingEventsResp: &api.MarketingEventsListResponse{
			Items: []api.MarketingEvent{
				{
					ID:            "event_json",
					EventType:     "social",
					MarketingType: "display",
					UTMCampaign:   "spring_promo",
					Budget:        500.00,
					Currency:      "EUR",
					CreatedAt:     fixedTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupMarketingEventsMockFactories(mockClient)
	defer cleanup()

	cmd := newMarketingEventsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("event-type", "", "")
	cmd.Flags().String("marketing-type", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := marketingEventsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "event_json") {
		t.Errorf("JSON output should contain event ID, got: %s", output)
	}
}

// TestMarketingEventsGetRunE tests the marketing events get command with mock API.
func TestMarketingEventsGetRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		eventID  string
		mockResp *api.MarketingEvent
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get with all fields",
			eventID: "event_123",
			mockResp: &api.MarketingEvent{
				ID:            "event_123",
				EventType:     "campaign",
				MarketingType: "email",
				RemoteID:      "remote_abc",
				Description:   "Summer sale campaign",
				Budget:        1000.00,
				Currency:      "USD",
				UTMCampaign:   "summer_sale",
				UTMSource:     "newsletter",
				UTMMedium:     "email",
				ManageURL:     "https://example.com/manage",
				PreviewURL:    "https://example.com/preview",
				StartedAt:     fixedTime,
				EndedAt:       fixedTime.Add(24 * time.Hour),
				MarketedResources: []api.MarketedResource{
					{Type: "product", ID: "prod_1"},
					{Type: "collection", ID: "col_1"},
				},
				CreatedAt: fixedTime,
				UpdatedAt: fixedTime,
			},
		},
		{
			name:    "successful get with minimal fields",
			eventID: "event_456",
			mockResp: &api.MarketingEvent{
				ID:            "event_456",
				EventType:     "ad",
				MarketingType: "cpc",
				CreatedAt:     fixedTime,
				UpdatedAt:     fixedTime,
			},
		},
		{
			name:    "event not found",
			eventID: "event_999",
			mockErr: errors.New("event not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketingEventsMockAPIClient{
				getMarketingEventResp: tt.mockResp,
				getMarketingEventErr:  tt.mockErr,
			}
			cleanup, _ := setupMarketingEventsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketingEventsTestCmd()

			err := marketingEventsGetCmd.RunE(cmd, []string{tt.eventID})

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

// TestMarketingEventsGetRunEJSON tests the get command with JSON output.
func TestMarketingEventsGetRunEJSON(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &marketingEventsMockAPIClient{
		getMarketingEventResp: &api.MarketingEvent{
			ID:            "event_json",
			EventType:     "social",
			MarketingType: "display",
			CreatedAt:     fixedTime,
			UpdatedAt:     fixedTime,
		},
	}
	cleanup, buf := setupMarketingEventsMockFactories(mockClient)
	defer cleanup()

	cmd := newMarketingEventsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := marketingEventsGetCmd.RunE(cmd, []string{"event_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "event_json") {
		t.Errorf("JSON output should contain event ID, got: %s", output)
	}
}

// TestMarketingEventsCreateRunE tests the marketing events create command with mock API.
func TestMarketingEventsCreateRunE(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		mockResp *api.MarketingEvent
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful create",
			mockResp: &api.MarketingEvent{
				ID:            "event_new",
				EventType:     "campaign",
				MarketingType: "email",
				UTMCampaign:   "new_campaign",
				CreatedAt:     fixedTime,
				UpdatedAt:     fixedTime,
			},
		},
		{
			name:    "API error",
			mockErr: errors.New("creation failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketingEventsMockAPIClient{
				createMarketingEventResp: tt.mockResp,
				createMarketingEventErr:  tt.mockErr,
			}
			cleanup, _ := setupMarketingEventsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketingEventsTestCmd()
			cmd.Flags().String("event-type", "campaign", "")
			cmd.Flags().String("marketing-type", "email", "")
			cmd.Flags().String("utm-campaign", "test_campaign", "")
			cmd.Flags().String("utm-source", "test_source", "")
			cmd.Flags().String("utm-medium", "test_medium", "")
			cmd.Flags().Float64("budget", 500.0, "")
			cmd.Flags().String("currency", "USD", "")
			cmd.Flags().String("description", "Test description", "")

			err := marketingEventsCreateCmd.RunE(cmd, []string{})

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

// TestMarketingEventsCreateRunEJSON tests the create command with JSON output.
func TestMarketingEventsCreateRunEJSON(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	mockClient := &marketingEventsMockAPIClient{
		createMarketingEventResp: &api.MarketingEvent{
			ID:            "event_created_json",
			EventType:     "ad",
			MarketingType: "cpc",
			CreatedAt:     fixedTime,
			UpdatedAt:     fixedTime,
		},
	}
	cleanup, buf := setupMarketingEventsMockFactories(mockClient)
	defer cleanup()

	cmd := newMarketingEventsTestCmd()
	cmd.Flags().String("event-type", "ad", "")
	cmd.Flags().String("marketing-type", "cpc", "")
	cmd.Flags().String("utm-campaign", "", "")
	cmd.Flags().String("utm-source", "", "")
	cmd.Flags().String("utm-medium", "", "")
	cmd.Flags().Float64("budget", 0, "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("description", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := marketingEventsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "event_created_json") {
		t.Errorf("JSON output should contain event ID, got: %s", output)
	}
}

// TestMarketingEventsCreateDryRun tests the create command dry-run mode.
func TestMarketingEventsCreateDryRun(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("event-type", "campaign", "")
	cmd.Flags().String("marketing-type", "email", "")
	cmd.Flags().String("utm-campaign", "", "")
	cmd.Flags().String("utm-source", "", "")
	cmd.Flags().String("utm-medium", "", "")
	cmd.Flags().Float64("budget", 0, "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("description", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := marketingEventsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}
}

// TestMarketingEventsCreateGetClientError tests error handling when getClient fails.
func TestMarketingEventsCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("event-type", "campaign", "")
	cmd.Flags().String("marketing-type", "email", "")
	cmd.Flags().String("utm-campaign", "", "")
	cmd.Flags().String("utm-source", "", "")
	cmd.Flags().String("utm-medium", "", "")
	cmd.Flags().Float64("budget", 0, "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("description", "", "")

	err := marketingEventsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMarketingEventsDeleteRunE tests the marketing events delete command with mock API.
func TestMarketingEventsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		eventID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			eventID: "event_123",
		},
		{
			name:    "delete error",
			eventID: "event_456",
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &marketingEventsMockAPIClient{
				deleteMarketingEventErr: tt.mockErr,
			}
			cleanup, _ := setupMarketingEventsMockFactories(mockClient)
			defer cleanup()

			cmd := newMarketingEventsTestCmd()
			_ = cmd.Flags().Set("yes", "true")

			err := marketingEventsDeleteCmd.RunE(cmd, []string{tt.eventID})

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

// TestMarketingEventsDeleteDryRun tests the delete command dry-run mode.
func TestMarketingEventsDeleteDryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")

	err := marketingEventsDeleteCmd.RunE(cmd, []string{"event-123"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run: %v", err)
	}
}

// TestMarketingEventsDeleteNoConfirmation tests that delete requires confirmation.
func TestMarketingEventsDeleteNoConfirmation(t *testing.T) {
	cmd := newTestCmdWithFlags()
	// Without --yes flag, should print message and return nil

	err := marketingEventsDeleteCmd.RunE(cmd, []string{"event-123"})
	if err != nil {
		t.Errorf("Unexpected error without confirmation: %v", err)
	}
}

// TestMarketingEventsDeleteGetClientError tests error handling when getClient fails.
func TestMarketingEventsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")

	err := marketingEventsDeleteCmd.RunE(cmd, []string{"event-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMarketingEventsListGetClientError tests error handling when getClient fails.
func TestMarketingEventsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := marketingEventsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMarketingEventsGetGetClientError tests error handling when getClient fails.
func TestMarketingEventsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := marketingEventsGetCmd.RunE(cmd, []string{"event-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// TestMarketingEventsListFlags tests that list command has correct flags.
func TestMarketingEventsListFlags(t *testing.T) {
	flags := marketingEventsListCmd.Flags()

	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
	if flags.Lookup("event-type") == nil {
		t.Error("Expected event-type flag")
	}
	if flags.Lookup("marketing-type") == nil {
		t.Error("Expected marketing-type flag")
	}
}

// TestMarketingEventsListFlagsDefaults tests that list command flags have correct defaults.
func TestMarketingEventsListFlagsDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"event-type", ""},
		{"marketing-type", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := marketingEventsListCmd.Flags().Lookup(f.name)
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

// TestMarketingEventsCreateFlags tests that create command has correct flags.
func TestMarketingEventsCreateFlags(t *testing.T) {
	flags := []string{"event-type", "marketing-type", "utm-campaign", "utm-source", "utm-medium", "budget", "currency", "description"}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := marketingEventsCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("flag %q not found", name)
			}
		})
	}
}

// TestMarketingEventsDeleteFlags tests that delete command has correct flags.
func TestMarketingEventsDeleteFlags(t *testing.T) {
	flag := marketingEventsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("Expected yes flag")
	}
}

// TestMarketingEventsCommandSetup verifies marketing-events command initialization.
func TestMarketingEventsCommandSetup(t *testing.T) {
	if marketingEventsCmd.Use != "marketing-events" {
		t.Errorf("expected Use 'marketing-events', got %q", marketingEventsCmd.Use)
	}
	if marketingEventsCmd.Short != "Manage marketing event tracking" {
		t.Errorf("expected Short 'Manage marketing event tracking', got %q", marketingEventsCmd.Short)
	}
}

// TestMarketingEventsSubcommands verifies all subcommands are registered.
func TestMarketingEventsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List marketing events",
		"get":    "Get marketing event details",
		"create": "Create a marketing event",
		"delete": "Delete a marketing event",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range marketingEventsCmd.Commands() {
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

// TestMarketingEventsCommandStructure tests the command structure.
func TestMarketingEventsCommandStructure(t *testing.T) {
	if marketingEventsCmd.Use != "marketing-events" {
		t.Errorf("Expected Use 'marketing-events', got %s", marketingEventsCmd.Use)
	}

	subcommands := marketingEventsCmd.Commands()
	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected subcommand %s not found", name)
		}
	}
}

// TestMarketingEventsGetCmdUse verifies the get command has correct use string.
func TestMarketingEventsGetCmdUse(t *testing.T) {
	if marketingEventsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", marketingEventsGetCmd.Use)
	}
}

// TestMarketingEventsDeleteCmdUse verifies the delete command has correct use string.
func TestMarketingEventsDeleteCmdUse(t *testing.T) {
	if marketingEventsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", marketingEventsDeleteCmd.Use)
	}
}

// TestMarketingEventsListRunENoProfiles verifies error when no profiles are configured.
func TestMarketingEventsListRunENoProfiles(t *testing.T) {
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
	err := marketingEventsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestMarketingEventsGetRunEMultipleProfiles verifies error when multiple profiles exist without selection.
func TestMarketingEventsGetRunEMultipleProfiles(t *testing.T) {
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
	err := marketingEventsGetCmd.RunE(cmd, []string{"event_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestMarketingEventsCreateRunEMultipleProfiles verifies error when multiple profiles exist without selection.
func TestMarketingEventsCreateRunEMultipleProfiles(t *testing.T) {
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
	cmd.Flags().String("event-type", "campaign", "")
	cmd.Flags().String("marketing-type", "email", "")
	cmd.Flags().String("utm-campaign", "", "")
	cmd.Flags().String("utm-source", "", "")
	cmd.Flags().String("utm-medium", "", "")
	cmd.Flags().Float64("budget", 0, "")
	cmd.Flags().String("currency", "USD", "")
	cmd.Flags().String("description", "", "")

	err := marketingEventsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestMarketingEventsDeleteRunEMultipleProfiles verifies error when multiple profiles exist without selection.
func TestMarketingEventsDeleteRunEMultipleProfiles(t *testing.T) {
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
	_ = cmd.Flags().Set("yes", "true")

	err := marketingEventsDeleteCmd.RunE(cmd, []string{"event_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestMarketingEventsListFlagDescriptions verifies flag descriptions are set.
func TestMarketingEventsListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"page":           "Page number",
		"page-size":      "Results per page",
		"event-type":     "Filter by event type (ad, campaign, email, social)",
		"marketing-type": "Filter by marketing type (cpc, display, social, search, email)",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := marketingEventsListCmd.Flags().Lookup(flagName)
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

// TestMarketingEventsListFlagTypes verifies flag types are correct.
func TestMarketingEventsListFlagTypes(t *testing.T) {
	flags := map[string]string{
		"page":           "int",
		"page-size":      "int",
		"event-type":     "string",
		"marketing-type": "string",
	}

	for flagName, expectedType := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := marketingEventsListCmd.Flags().Lookup(flagName)
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

// TestMarketingEventsCreateFlagDescriptions verifies create flag descriptions are set.
func TestMarketingEventsCreateFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"event-type":     "Event type (ad, campaign, email, social) (required)",
		"marketing-type": "Marketing type (cpc, display, social, search, email) (required)",
		"utm-campaign":   "UTM campaign parameter",
		"utm-source":     "UTM source parameter",
		"utm-medium":     "UTM medium parameter",
		"budget":         "Campaign budget",
		"currency":       "Budget currency",
		"description":    "Event description",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := marketingEventsCreateCmd.Flags().Lookup(flagName)
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

// TestMarketingEventsCreateFlagDefaults tests that create command flags have correct defaults.
func TestMarketingEventsCreateFlagDefaults(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"event-type", ""},
		{"marketing-type", ""},
		{"utm-campaign", ""},
		{"utm-source", ""},
		{"utm-medium", ""},
		{"budget", "0"},
		{"currency", "USD"},
		{"description", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := marketingEventsCreateCmd.Flags().Lookup(f.name)
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

// TestMarketingEventsDeleteFlagDefaults tests that delete command flags have correct defaults.
func TestMarketingEventsDeleteFlagDefaults(t *testing.T) {
	flag := marketingEventsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default 'false', got %q", flag.DefValue)
	}
}

// TestMarketingEventsDeleteFlagDescription tests delete flag description.
func TestMarketingEventsDeleteFlagDescription(t *testing.T) {
	flag := marketingEventsDeleteCmd.Flags().Lookup("yes")
	if flag == nil {
		t.Error("flag 'yes' not found")
		return
	}
	expectedUsage := "Skip confirmation prompt"
	if flag.Usage != expectedUsage {
		t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
	}
}

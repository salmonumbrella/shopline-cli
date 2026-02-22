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

// cdpMockAPIClient is a mock implementation of api.APIClient for CDP tests.
type cdpMockAPIClient struct {
	api.MockClient
	listProfilesResp *api.CDPProfilesListResponse
	listProfilesErr  error
	getProfileResp   *api.CDPCustomerProfile
	getProfileErr    error
	listEventsResp   *api.CDPEventsListResponse
	listEventsErr    error
	getEventResp     *api.CDPEvent
	getEventErr      error
	listSegmentsResp *api.CDPSegmentsListResponse
	listSegmentsErr  error
	getSegmentResp   *api.CDPSegment
	getSegmentErr    error
}

func (m *cdpMockAPIClient) ListCDPProfiles(ctx context.Context, opts *api.CDPProfilesListOptions) (*api.CDPProfilesListResponse, error) {
	return m.listProfilesResp, m.listProfilesErr
}

func (m *cdpMockAPIClient) GetCDPProfile(ctx context.Context, id string) (*api.CDPCustomerProfile, error) {
	return m.getProfileResp, m.getProfileErr
}

func (m *cdpMockAPIClient) ListCDPEvents(ctx context.Context, opts *api.CDPEventsListOptions) (*api.CDPEventsListResponse, error) {
	return m.listEventsResp, m.listEventsErr
}

func (m *cdpMockAPIClient) GetCDPEvent(ctx context.Context, id string) (*api.CDPEvent, error) {
	return m.getEventResp, m.getEventErr
}

func (m *cdpMockAPIClient) ListCDPSegments(ctx context.Context, opts *api.CDPSegmentsListOptions) (*api.CDPSegmentsListResponse, error) {
	return m.listSegmentsResp, m.listSegmentsErr
}

func (m *cdpMockAPIClient) GetCDPSegment(ctx context.Context, id string) (*api.CDPSegment, error) {
	return m.getSegmentResp, m.getSegmentErr
}

// setupCDPMockFactories sets up mock factories for CDP tests.
func setupCDPMockFactories(mockClient *cdpMockAPIClient) (func(), *bytes.Buffer) {
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

// newCDPTestCmd creates a test command with common flags for CDP tests.
func newCDPTestCmd() *cobra.Command {
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

// ============================================================================
// CDP Profiles Tests
// ============================================================================

func TestCDPProfilesListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("segment", "", "")
	cmd.Flags().String("tag", "", "")
	cmd.Flags().String("churn-risk", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := cdpProfilesListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCDPProfilesGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := cdpProfilesGetCmd.RunE(cmd, []string{"profile-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCDPProfilesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CDPProfilesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CDPProfilesListResponse{
				Items: []api.CDPCustomerProfile{
					{
						ID:            "prof_123",
						Email:         "alice@example.com",
						TotalOrders:   10,
						TotalSpent:    "500.00",
						LifetimeValue: "750.00",
						ChurnRisk:     "low",
						Segments:      []string{"high_value", "repeat_buyer"},
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "prof_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CDPProfilesListResponse{
				Items:      []api.CDPCustomerProfile{},
				TotalCount: 0,
			},
		},
		{
			name: "profile with long segments list",
			mockResp: &api.CDPProfilesListResponse{
				Items: []api.CDPCustomerProfile{
					{
						ID:            "prof_456",
						Email:         "bob@example.com",
						TotalOrders:   5,
						TotalSpent:    "250.00",
						LifetimeValue: "300.00",
						ChurnRisk:     "medium",
						Segments:      []string{"segment_one", "segment_two", "segment_three", "segment_four"},
						CreatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "prof_456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &cdpMockAPIClient{
				listProfilesResp: tt.mockResp,
				listProfilesErr:  tt.mockErr,
			}
			cleanup, buf := setupCDPMockFactories(mockClient)
			defer cleanup()

			cmd := newCDPTestCmd()
			cmd.Flags().String("segment", "", "")
			cmd.Flags().String("tag", "", "")
			cmd.Flags().String("churn-risk", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := cdpProfilesListCmd.RunE(cmd, []string{})

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
				t.Errorf("output should contain %q, got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestCDPProfilesListRunEWithJSON(t *testing.T) {
	mockClient := &cdpMockAPIClient{
		listProfilesResp: &api.CDPProfilesListResponse{
			Items: []api.CDPCustomerProfile{
				{ID: "prof_json", Email: "json@example.com"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCDPMockFactories(mockClient)
	defer cleanup()

	cmd := newCDPTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("segment", "", "")
	cmd.Flags().String("tag", "", "")
	cmd.Flags().String("churn-risk", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := cdpProfilesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "prof_json") {
		t.Errorf("JSON output should contain profile ID, got: %s", output)
	}
}

func TestCDPProfilesGetRunE(t *testing.T) {
	now := time.Now()
	firstOrder := now.Add(-365 * 24 * time.Hour)
	lastOrder := now.Add(-7 * 24 * time.Hour)

	tests := []struct {
		name      string
		profileID string
		mockResp  *api.CDPCustomerProfile
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			profileID: "prof_123",
			mockResp: &api.CDPCustomerProfile{
				ID:                "prof_123",
				CustomerID:        "cust_456",
				Email:             "alice@example.com",
				Phone:             "+1234567890",
				FirstName:         "Alice",
				LastName:          "Smith",
				TotalOrders:       15,
				TotalSpent:        "1500.00",
				AverageOrderValue: "100.00",
				LifetimeValue:     "2000.00",
				PredictedLTV:      "2500.00",
				ChurnRisk:         "low",
				Segments:          []string{"high_value", "repeat_buyer"},
				Tags:              []string{"vip", "loyal"},
				FirstOrderAt:      &firstOrder,
				LastOrderAt:       &lastOrder,
				CreatedAt:         time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:         time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "profile not found",
			profileID: "prof_999",
			mockErr:   errors.New("profile not found"),
			wantErr:   true,
		},
		{
			name:      "profile with RFM score",
			profileID: "prof_rfm",
			mockResp: &api.CDPCustomerProfile{
				ID:          "prof_rfm",
				CustomerID:  "cust_rfm",
				Email:       "rfm@example.com",
				TotalOrders: 20,
				TotalSpent:  "2000.00",
				RFMScore: &api.CDPRFMScore{
					Recency:   5,
					Frequency: 4,
					Monetary:  5,
					Total:     14,
					Segment:   "Champion",
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "profile with preferences",
			profileID: "prof_pref",
			mockResp: &api.CDPCustomerProfile{
				ID:          "prof_pref",
				CustomerID:  "cust_pref",
				Email:       "pref@example.com",
				TotalOrders: 8,
				TotalSpent:  "800.00",
				Preferences: &api.CDPCustomerPreferences{
					EmailMarketing:    true,
					SMSMarketing:      false,
					PushNotifications: true,
					PreferredChannel:  "email",
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &cdpMockAPIClient{
				getProfileResp: tt.mockResp,
				getProfileErr:  tt.mockErr,
			}
			cleanup, _ := setupCDPMockFactories(mockClient)
			defer cleanup()

			cmd := newCDPTestCmd()

			err := cdpProfilesGetCmd.RunE(cmd, []string{tt.profileID})

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

func TestCDPProfilesGetRunEWithJSON(t *testing.T) {
	mockClient := &cdpMockAPIClient{
		getProfileResp: &api.CDPCustomerProfile{
			ID:          "prof_json",
			CustomerID:  "cust_json",
			Email:       "json@example.com",
			TotalOrders: 5,
			TotalSpent:  "500.00",
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupCDPMockFactories(mockClient)
	defer cleanup()

	cmd := newCDPTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := cdpProfilesGetCmd.RunE(cmd, []string{"prof_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "prof_json") {
		t.Errorf("JSON output should contain profile ID, got: %s", output)
	}
}

// ============================================================================
// CDP Events Tests
// ============================================================================

func TestCDPEventsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("event-type", "", "")
	cmd.Flags().String("event-name", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := cdpEventsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCDPEventsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := cdpEventsGetCmd.RunE(cmd, []string{"event-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCDPEventsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CDPEventsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CDPEventsListResponse{
				Items: []api.CDPEvent{
					{
						ID:         "evt_123",
						CustomerID: "cust_456",
						SessionID:  "sess_789",
						EventType:  "page_view",
						EventName:  "product_viewed",
						Source:     "web",
						Channel:    "desktop",
						Timestamp:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
						CreatedAt:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "evt_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CDPEventsListResponse{
				Items:      []api.CDPEvent{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple events",
			mockResp: &api.CDPEventsListResponse{
				Items: []api.CDPEvent{
					{
						ID:         "evt_1",
						CustomerID: "cust_456",
						EventType:  "page_view",
						EventName:  "homepage_visited",
						Source:     "web",
						Channel:    "desktop",
						Timestamp:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
						CreatedAt:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
					},
					{
						ID:         "evt_2",
						CustomerID: "cust_456",
						EventType:  "purchase",
						EventName:  "order_placed",
						Source:     "web",
						Channel:    "mobile",
						Timestamp:  time.Date(2024, 6, 15, 15, 0, 0, 0, time.UTC),
						CreatedAt:  time.Date(2024, 6, 15, 15, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "evt_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &cdpMockAPIClient{
				listEventsResp: tt.mockResp,
				listEventsErr:  tt.mockErr,
			}
			cleanup, buf := setupCDPMockFactories(mockClient)
			defer cleanup()

			cmd := newCDPTestCmd()
			cmd.Flags().String("customer-id", "", "")
			cmd.Flags().String("event-type", "", "")
			cmd.Flags().String("event-name", "", "")
			cmd.Flags().String("source", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := cdpEventsListCmd.RunE(cmd, []string{})

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
				t.Errorf("output should contain %q, got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestCDPEventsListRunEWithJSON(t *testing.T) {
	mockClient := &cdpMockAPIClient{
		listEventsResp: &api.CDPEventsListResponse{
			Items: []api.CDPEvent{
				{
					ID:         "evt_json",
					CustomerID: "cust_456",
					EventType:  "page_view",
					EventName:  "product_viewed",
					Source:     "web",
					Channel:    "desktop",
					Timestamp:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
					CreatedAt:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCDPMockFactories(mockClient)
	defer cleanup()

	cmd := newCDPTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("customer-id", "", "")
	cmd.Flags().String("event-type", "", "")
	cmd.Flags().String("event-name", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := cdpEventsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "evt_json") {
		t.Errorf("JSON output should contain event ID, got: %s", output)
	}
}

func TestCDPEventsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		eventID  string
		mockResp *api.CDPEvent
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			eventID: "evt_123",
			mockResp: &api.CDPEvent{
				ID:         "evt_123",
				CustomerID: "cust_456",
				SessionID:  "sess_789",
				EventType:  "purchase",
				EventName:  "order_placed",
				Source:     "web",
				Channel:    "desktop",
				Properties: map[string]interface{}{
					"order_id":    "ord_123",
					"total":       150.00,
					"items_count": 3,
				},
				Timestamp: time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
				CreatedAt: time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
			},
		},
		{
			name:    "event not found",
			eventID: "evt_999",
			mockErr: errors.New("event not found"),
			wantErr: true,
		},
		{
			name:    "event without properties",
			eventID: "evt_simple",
			mockResp: &api.CDPEvent{
				ID:         "evt_simple",
				CustomerID: "cust_456",
				SessionID:  "sess_789",
				EventType:  "page_view",
				EventName:  "homepage_visited",
				Source:     "web",
				Channel:    "mobile",
				Timestamp:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
				CreatedAt:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &cdpMockAPIClient{
				getEventResp: tt.mockResp,
				getEventErr:  tt.mockErr,
			}
			cleanup, _ := setupCDPMockFactories(mockClient)
			defer cleanup()

			cmd := newCDPTestCmd()

			err := cdpEventsGetCmd.RunE(cmd, []string{tt.eventID})

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

func TestCDPEventsGetRunEWithJSON(t *testing.T) {
	mockClient := &cdpMockAPIClient{
		getEventResp: &api.CDPEvent{
			ID:         "evt_json",
			CustomerID: "cust_456",
			EventType:  "purchase",
			EventName:  "order_placed",
			Source:     "web",
			Channel:    "desktop",
			Timestamp:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
			CreatedAt:  time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupCDPMockFactories(mockClient)
	defer cleanup()

	cmd := newCDPTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := cdpEventsGetCmd.RunE(cmd, []string{"evt_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "evt_json") {
		t.Errorf("JSON output should contain event ID, got: %s", output)
	}
}

// ============================================================================
// CDP Segments Tests
// ============================================================================

func TestCDPSegmentsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("type", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := cdpSegmentsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCDPSegmentsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := cdpSegmentsGetCmd.RunE(cmd, []string{"segment-123"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCDPSegmentsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.CDPSegmentsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.CDPSegmentsListResponse{
				Items: []api.CDPSegment{
					{
						ID:            "seg_123",
						Name:          "High Value Customers",
						Description:   "Customers with lifetime value > $1000",
						Type:          "dynamic",
						CustomerCount: 1500,
						Status:        "active",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "seg_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.CDPSegmentsListResponse{
				Items:      []api.CDPSegment{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple segments",
			mockResp: &api.CDPSegmentsListResponse{
				Items: []api.CDPSegment{
					{
						ID:            "seg_1",
						Name:          "High Value",
						Type:          "dynamic",
						CustomerCount: 500,
						Status:        "active",
						CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
					},
					{
						ID:            "seg_2",
						Name:          "At Risk",
						Type:          "dynamic",
						CustomerCount: 200,
						Status:        "active",
						CreatedAt:     time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt:     time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "seg_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &cdpMockAPIClient{
				listSegmentsResp: tt.mockResp,
				listSegmentsErr:  tt.mockErr,
			}
			cleanup, buf := setupCDPMockFactories(mockClient)
			defer cleanup()

			cmd := newCDPTestCmd()
			cmd.Flags().String("type", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := cdpSegmentsListCmd.RunE(cmd, []string{})

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
				t.Errorf("output should contain %q, got: %s", tt.wantOutput, output)
			}
		})
	}
}

func TestCDPSegmentsListRunEWithJSON(t *testing.T) {
	mockClient := &cdpMockAPIClient{
		listSegmentsResp: &api.CDPSegmentsListResponse{
			Items: []api.CDPSegment{
				{
					ID:            "seg_json",
					Name:          "Test Segment",
					Type:          "dynamic",
					CustomerCount: 100,
					Status:        "active",
					CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:     time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupCDPMockFactories(mockClient)
	defer cleanup()

	cmd := newCDPTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("type", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := cdpSegmentsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "seg_json") {
		t.Errorf("JSON output should contain segment ID, got: %s", output)
	}
}

func TestCDPSegmentsGetRunE(t *testing.T) {
	tests := []struct {
		name      string
		segmentID string
		mockResp  *api.CDPSegment
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful get",
			segmentID: "seg_123",
			mockResp: &api.CDPSegment{
				ID:            "seg_123",
				Name:          "High Value Customers",
				Description:   "Customers with lifetime value > $1000",
				Type:          "dynamic",
				CustomerCount: 1500,
				Status:        "active",
				Conditions: []api.CDPSegmentCondition{
					{
						Field:    "lifetime_value",
						Operator: "greater_than",
						Value:    1000,
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "segment not found",
			segmentID: "seg_999",
			mockErr:   errors.New("segment not found"),
			wantErr:   true,
		},
		{
			name:      "segment without conditions",
			segmentID: "seg_static",
			mockResp: &api.CDPSegment{
				ID:            "seg_static",
				Name:          "Manual Segment",
				Description:   "Manually curated customer list",
				Type:          "static",
				CustomerCount: 50,
				Status:        "active",
				CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "segment with multiple conditions",
			segmentID: "seg_multi",
			mockResp: &api.CDPSegment{
				ID:            "seg_multi",
				Name:          "High Value At Risk",
				Description:   "High value customers at risk of churning",
				Type:          "dynamic",
				CustomerCount: 75,
				Status:        "active",
				Conditions: []api.CDPSegmentCondition{
					{
						Field:    "lifetime_value",
						Operator: "greater_than",
						Value:    500,
					},
					{
						Field:    "churn_risk",
						Operator: "equals",
						Value:    "high",
					},
				},
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &cdpMockAPIClient{
				getSegmentResp: tt.mockResp,
				getSegmentErr:  tt.mockErr,
			}
			cleanup, _ := setupCDPMockFactories(mockClient)
			defer cleanup()

			cmd := newCDPTestCmd()

			err := cdpSegmentsGetCmd.RunE(cmd, []string{tt.segmentID})

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

func TestCDPSegmentsGetRunEWithJSON(t *testing.T) {
	mockClient := &cdpMockAPIClient{
		getSegmentResp: &api.CDPSegment{
			ID:            "seg_json",
			Name:          "Test Segment",
			Description:   "Test description",
			Type:          "dynamic",
			CustomerCount: 100,
			Status:        "active",
			CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
		},
	}
	cleanup, buf := setupCDPMockFactories(mockClient)
	defer cleanup()

	cmd := newCDPTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := cdpSegmentsGetCmd.RunE(cmd, []string{"seg_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "seg_json") {
		t.Errorf("JSON output should contain segment ID, got: %s", output)
	}
}

// ============================================================================
// CDP Command Structure Tests
// ============================================================================

func TestCDPProfilesListFlags(t *testing.T) {
	flags := cdpProfilesListCmd.Flags()

	if flags.Lookup("segment") == nil {
		t.Error("Expected segment flag")
	}
	if flags.Lookup("tag") == nil {
		t.Error("Expected tag flag")
	}
	if flags.Lookup("churn-risk") == nil {
		t.Error("Expected churn-risk flag")
	}
	if flags.Lookup("page") == nil {
		t.Error("Expected page flag")
	}
	if flags.Lookup("page-size") == nil {
		t.Error("Expected page-size flag")
	}
}

func TestCDPCommandStructure(t *testing.T) {
	if cdpCmd.Use != "cdp" {
		t.Errorf("Expected Use 'cdp', got %s", cdpCmd.Use)
	}

	subcommands := cdpCmd.Commands()
	expectedCmds := map[string]bool{
		"profiles": false,
		"events":   false,
		"segments": false,
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

func TestCDPProfilesCommandStructure(t *testing.T) {
	subcommands := cdpProfilesCmd.Commands()
	expectedCmds := map[string]bool{
		"list": false,
		"get":  false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected profiles subcommand %s not found", name)
		}
	}
}

func TestCDPEventsCommandStructure(t *testing.T) {
	subcommands := cdpEventsCmd.Commands()
	expectedCmds := map[string]bool{
		"list": false,
		"get":  false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected events subcommand %s not found", name)
		}
	}
}

func TestCDPSegmentsCommandStructure(t *testing.T) {
	subcommands := cdpSegmentsCmd.Commands()
	expectedCmds := map[string]bool{
		"list": false,
		"get":  false,
	}

	for _, cmd := range subcommands {
		if startsWithUse(cmd.Use, expectedCmds) {
			expectedCmds[getBaseUse(cmd.Use)] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("Expected segments subcommand %s not found", name)
		}
	}
}

func TestCDPEventsListFlags(t *testing.T) {
	flags := cdpEventsListCmd.Flags()

	expectedFlags := []string{"customer-id", "event-type", "event-name", "source", "page", "page-size"}
	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Expected %s flag", flagName)
		}
	}
}

func TestCDPSegmentsListFlags(t *testing.T) {
	flags := cdpSegmentsListCmd.Flags()

	expectedFlags := []string{"type", "status", "page", "page-size"}
	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Expected %s flag", flagName)
		}
	}
}

func TestCDPProfilesGetArgs(t *testing.T) {
	err := cdpProfilesGetCmd.Args(cdpProfilesGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cdpProfilesGetCmd.Args(cdpProfilesGetCmd, []string{"profile-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

func TestCDPEventsGetArgs(t *testing.T) {
	err := cdpEventsGetCmd.Args(cdpEventsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cdpEventsGetCmd.Args(cdpEventsGetCmd, []string{"event-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

func TestCDPSegmentsGetArgs(t *testing.T) {
	err := cdpSegmentsGetCmd.Args(cdpSegmentsGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cdpSegmentsGetCmd.Args(cdpSegmentsGetCmd, []string{"segment-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

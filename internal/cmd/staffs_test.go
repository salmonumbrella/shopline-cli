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

func TestStaffsCmdStructure(t *testing.T) {
	if staffsCmd.Use != "staffs" {
		t.Errorf("Expected Use 'staffs', got %q", staffsCmd.Use)
	}

	subcommands := staffsCmd.Commands()
	expectedSubs := []string{"list", "get", "invite", "update", "delete"}

	for _, exp := range expectedSubs {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == exp || strings.HasPrefix(cmd.Use, exp+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing subcommand: %s", exp)
		}
	}
}

func TestStaffsListCmdFlags(t *testing.T) {
	pageFlag := staffsListCmd.Flags().Lookup("page")
	if pageFlag == nil {
		t.Error("Missing --page flag")
	}

	pageSizeFlag := staffsListCmd.Flags().Lookup("page-size")
	if pageSizeFlag == nil {
		t.Error("Missing --page-size flag")
	}
}

func TestStaffsGetCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"staff_123"}, wantErr: false},
		{name: "too many args", args: []string{"staff_1", "staff_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := staffsGetCmd.Args(staffsGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStaffsUpdateCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"staff_123"}, wantErr: false},
		{name: "too many args", args: []string{"staff_1", "staff_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := staffsUpdateCmd.Args(staffsUpdateCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStaffsDeleteCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"staff_123"}, wantErr: false},
		{name: "too many args", args: []string{"staff_1", "staff_2"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := staffsDeleteCmd.Args(staffsDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStaffsInviteCmdFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"email", ""},
		{"first-name", ""},
		{"last-name", ""},
		{"permissions", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := staffsInviteCmd.Flags().Lookup(f.name)
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

func TestStaffsUpdateCmdFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"first-name", ""},
		{"last-name", ""},
		{"phone", ""},
		{"locale", ""},
		{"permissions", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := staffsUpdateCmd.Flags().Lookup(f.name)
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

func TestStaffsListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(staffsListCmd)

	err := staffsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

func TestStaffsGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(staffsGetCmd)

	err := staffsGetCmd.RunE(cmd, []string{"staff_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// staffsMockAPIClient is a mock implementation of api.APIClient for staffs tests.
type staffsMockAPIClient struct {
	api.MockClient
	listStaffsResp  *api.StaffsListResponse
	listStaffsErr   error
	getStaffResp    *api.Staff
	getStaffErr     error
	inviteStaffResp *api.Staff
	inviteStaffErr  error
	updateStaffResp *api.Staff
	updateStaffErr  error
	deleteStaffErr  error
}

func (m *staffsMockAPIClient) ListStaffs(ctx context.Context, opts *api.StaffsListOptions) (*api.StaffsListResponse, error) {
	return m.listStaffsResp, m.listStaffsErr
}

func (m *staffsMockAPIClient) GetStaff(ctx context.Context, id string) (*api.Staff, error) {
	return m.getStaffResp, m.getStaffErr
}

func (m *staffsMockAPIClient) InviteStaff(ctx context.Context, req *api.StaffInviteRequest) (*api.Staff, error) {
	return m.inviteStaffResp, m.inviteStaffErr
}

func (m *staffsMockAPIClient) UpdateStaff(ctx context.Context, id string, req *api.StaffUpdateRequest) (*api.Staff, error) {
	return m.updateStaffResp, m.updateStaffErr
}

func (m *staffsMockAPIClient) DeleteStaff(ctx context.Context, id string) error {
	return m.deleteStaffErr
}

// setupStaffsMockFactories sets up mock factories for staffs tests.
func setupStaffsMockFactories(mockClient *staffsMockAPIClient) (func(), *bytes.Buffer) {
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

// newStaffsTestCmd creates a test command with common flags for staffs tests.
func newStaffsTestCmd() *cobra.Command {
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

// TestStaffsListRunE tests the staffs list command with mock API.
func TestStaffsListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.StaffsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.StaffsListResponse{
				Items: []api.Staff{
					{
						ID:           "staff_123",
						Email:        "alice@example.com",
						FirstName:    "Alice",
						LastName:     "Smith",
						AccountOwner: true,
						Permissions:  []string{"orders", "products"},
						CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "staff_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.StaffsListResponse{
				Items:      []api.Staff{},
				TotalCount: 0,
			},
		},
		{
			name: "staff with no permissions",
			mockResp: &api.StaffsListResponse{
				Items: []api.Staff{
					{
						ID:           "staff_456",
						Email:        "bob@example.com",
						FirstName:    "Bob",
						LastName:     "Jones",
						AccountOwner: false,
						Permissions:  []string{},
						CreatedAt:    time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "Bob Jones",
		},
		{
			name: "staff with multiple permissions",
			mockResp: &api.StaffsListResponse{
				Items: []api.Staff{
					{
						ID:           "staff_789",
						Email:        "charlie@example.com",
						FirstName:    "Charlie",
						LastName:     "Brown",
						AccountOwner: false,
						Permissions:  []string{"orders", "products", "customers", "settings"},
						CreatedAt:    time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "orders, products, customers, settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &staffsMockAPIClient{
				listStaffsResp: tt.mockResp,
				listStaffsErr:  tt.mockErr,
			}
			cleanup, buf := setupStaffsMockFactories(mockClient)
			defer cleanup()

			cmd := newStaffsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := staffsListCmd.RunE(cmd, []string{})

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

// TestStaffsListRunEWithJSON tests JSON output format.
func TestStaffsListRunEWithJSON(t *testing.T) {
	mockClient := &staffsMockAPIClient{
		listStaffsResp: &api.StaffsListResponse{
			Items: []api.Staff{
				{ID: "staff_json", Email: "json@example.com", FirstName: "JSON", LastName: "Test"},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := staffsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "staff_json") {
		t.Errorf("JSON output should contain staff ID, got: %s", output)
	}
}

// TestStaffsGetRunE tests the staffs get command with mock API.
func TestStaffsGetRunE(t *testing.T) {
	tests := []struct {
		name     string
		staffID  string
		mockResp *api.Staff
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get",
			staffID: "staff_123",
			mockResp: &api.Staff{
				ID:           "staff_123",
				Email:        "alice@example.com",
				FirstName:    "Alice",
				LastName:     "Smith",
				Phone:        "+1234567890",
				AccountOwner: true,
				Locale:       "en",
				Permissions:  []string{"orders", "products"},
				CreatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name:    "staff not found",
			staffID: "staff_999",
			mockErr: errors.New("staff not found"),
			wantErr: true,
		},
		{
			name:    "get staff without phone",
			staffID: "staff_456",
			mockResp: &api.Staff{
				ID:           "staff_456",
				Email:        "bob@example.com",
				FirstName:    "Bob",
				LastName:     "Jones",
				Phone:        "",
				AccountOwner: false,
				Locale:       "zh",
				Permissions:  []string{"orders"},
				CreatedAt:    time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &staffsMockAPIClient{
				getStaffResp: tt.mockResp,
				getStaffErr:  tt.mockErr,
			}
			cleanup, _ := setupStaffsMockFactories(mockClient)
			defer cleanup()

			cmd := newStaffsTestCmd()

			err := staffsGetCmd.RunE(cmd, []string{tt.staffID})

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

// TestStaffsGetRunEWithJSON tests JSON output format for get command.
func TestStaffsGetRunEWithJSON(t *testing.T) {
	mockClient := &staffsMockAPIClient{
		getStaffResp: &api.Staff{
			ID:        "staff_json",
			Email:     "json@example.com",
			FirstName: "JSON",
			LastName:  "Test",
		},
	}
	cleanup, buf := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := staffsGetCmd.RunE(cmd, []string{"staff_json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "staff_json") {
		t.Errorf("JSON output should contain staff ID, got: %s", output)
	}
}

// TestStaffsInviteRunE tests the staffs invite command with mock API.
func TestStaffsInviteRunE(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		firstName   string
		lastName    string
		permissions string
		mockResp    *api.Staff
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "successful invite",
			email:       "newstaff@example.com",
			firstName:   "New",
			lastName:    "Staff",
			permissions: "orders, products",
			mockResp: &api.Staff{
				ID:          "staff_new",
				Email:       "newstaff@example.com",
				FirstName:   "New",
				LastName:    "Staff",
				Permissions: []string{"orders", "products"},
			},
		},
		{
			name:      "invite with minimal data",
			email:     "minimal@example.com",
			firstName: "",
			lastName:  "",
			mockResp: &api.Staff{
				ID:    "staff_minimal",
				Email: "minimal@example.com",
			},
		},
		{
			name:    "invite API error",
			email:   "error@example.com",
			mockErr: errors.New("email already exists"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &staffsMockAPIClient{
				inviteStaffResp: tt.mockResp,
				inviteStaffErr:  tt.mockErr,
			}
			cleanup, _ := setupStaffsMockFactories(mockClient)
			defer cleanup()

			cmd := newStaffsTestCmd()
			cmd.Flags().String("email", "", "")
			cmd.Flags().String("first-name", "", "")
			cmd.Flags().String("last-name", "", "")
			cmd.Flags().String("permissions", "", "")
			_ = cmd.Flags().Set("email", tt.email)
			if tt.firstName != "" {
				_ = cmd.Flags().Set("first-name", tt.firstName)
			}
			if tt.lastName != "" {
				_ = cmd.Flags().Set("last-name", tt.lastName)
			}
			if tt.permissions != "" {
				_ = cmd.Flags().Set("permissions", tt.permissions)
			}

			err := staffsInviteCmd.RunE(cmd, []string{})

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

// TestStaffsInviteRunEWithJSON tests JSON output format for invite command.
func TestStaffsInviteRunEWithJSON(t *testing.T) {
	mockClient := &staffsMockAPIClient{
		inviteStaffResp: &api.Staff{
			ID:        "staff_invited",
			Email:     "invited@example.com",
			FirstName: "Invited",
			LastName:  "User",
		},
	}
	cleanup, buf := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("permissions", "", "")
	_ = cmd.Flags().Set("email", "invited@example.com")

	err := staffsInviteCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "staff_invited") {
		t.Errorf("JSON output should contain staff ID, got: %s", output)
	}
}

// TestStaffsInviteDryRun tests the staffs invite command with dry-run flag.
func TestStaffsInviteDryRun(t *testing.T) {
	mockClient := &staffsMockAPIClient{}
	cleanup, _ := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("permissions", "", "")
	_ = cmd.Flags().Set("email", "dryrun@example.com")
	_ = cmd.Flags().Set("dry-run", "true")

	err := staffsInviteCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStaffsUpdateRunE tests the staffs update command with mock API.
func TestStaffsUpdateRunE(t *testing.T) {
	tests := []struct {
		name        string
		staffID     string
		firstName   string
		lastName    string
		phone       string
		locale      string
		permissions string
		mockResp    *api.Staff
		mockErr     error
		wantErr     bool
	}{
		{
			name:      "successful update",
			staffID:   "staff_123",
			firstName: "Updated",
			lastName:  "Name",
			phone:     "+9876543210",
			locale:    "zh",
			mockResp: &api.Staff{
				ID:        "staff_123",
				Email:     "alice@example.com",
				FirstName: "Updated",
				LastName:  "Name",
				Phone:     "+9876543210",
				Locale:    "zh",
			},
		},
		{
			name:        "update permissions only",
			staffID:     "staff_456",
			permissions: "orders, products, settings",
			mockResp: &api.Staff{
				ID:          "staff_456",
				Permissions: []string{"orders", "products", "settings"},
			},
		},
		{
			name:    "update API error",
			staffID: "staff_999",
			mockErr: errors.New("staff not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &staffsMockAPIClient{
				updateStaffResp: tt.mockResp,
				updateStaffErr:  tt.mockErr,
			}
			cleanup, _ := setupStaffsMockFactories(mockClient)
			defer cleanup()

			cmd := newStaffsTestCmd()
			cmd.Flags().String("first-name", "", "")
			cmd.Flags().String("last-name", "", "")
			cmd.Flags().String("phone", "", "")
			cmd.Flags().String("locale", "", "")
			cmd.Flags().String("permissions", "", "")
			if tt.firstName != "" {
				_ = cmd.Flags().Set("first-name", tt.firstName)
			}
			if tt.lastName != "" {
				_ = cmd.Flags().Set("last-name", tt.lastName)
			}
			if tt.phone != "" {
				_ = cmd.Flags().Set("phone", tt.phone)
			}
			if tt.locale != "" {
				_ = cmd.Flags().Set("locale", tt.locale)
			}
			if tt.permissions != "" {
				_ = cmd.Flags().Set("permissions", tt.permissions)
			}

			err := staffsUpdateCmd.RunE(cmd, []string{tt.staffID})

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

// TestStaffsUpdateRunEWithJSON tests JSON output format for update command.
func TestStaffsUpdateRunEWithJSON(t *testing.T) {
	mockClient := &staffsMockAPIClient{
		updateStaffResp: &api.Staff{
			ID:        "staff_updated",
			Email:     "updated@example.com",
			FirstName: "Updated",
			LastName:  "User",
		},
	}
	cleanup, buf := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	_ = cmd.Flags().Set("output", "json")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("locale", "", "")
	cmd.Flags().String("permissions", "", "")
	_ = cmd.Flags().Set("first-name", "Updated")

	err := staffsUpdateCmd.RunE(cmd, []string{"staff_updated"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "staff_updated") {
		t.Errorf("JSON output should contain staff ID, got: %s", output)
	}
}

// TestStaffsUpdateDryRun tests the staffs update command with dry-run flag.
func TestStaffsUpdateDryRun(t *testing.T) {
	mockClient := &staffsMockAPIClient{}
	cleanup, _ := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("locale", "", "")
	cmd.Flags().String("permissions", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := staffsUpdateCmd.RunE(cmd, []string{"staff_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStaffsDeleteRunE tests the staffs delete command with mock API.
func TestStaffsDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		staffID string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete",
			staffID: "staff_123",
			mockErr: nil,
		},
		{
			name:    "delete API error",
			staffID: "staff_999",
			mockErr: errors.New("staff not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &staffsMockAPIClient{
				deleteStaffErr: tt.mockErr,
			}
			cleanup, _ := setupStaffsMockFactories(mockClient)
			defer cleanup()

			cmd := newStaffsTestCmd()

			err := staffsDeleteCmd.RunE(cmd, []string{tt.staffID})

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

// TestStaffsDeleteDryRun tests the staffs delete command with dry-run flag.
func TestStaffsDeleteDryRun(t *testing.T) {
	mockClient := &staffsMockAPIClient{}
	cleanup, _ := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := staffsDeleteCmd.RunE(cmd, []string{"staff_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStaffsInviteGetClientError verifies invite command error handling when getClient fails.
func TestStaffsInviteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("permissions", "", "")
	_ = cmd.Flags().Set("email", "test@example.com")
	cmd.AddCommand(staffsInviteCmd)

	err := staffsInviteCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStaffsUpdateGetClientError verifies update command error handling when getClient fails.
func TestStaffsUpdateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("first-name", "", "")
	cmd.Flags().String("last-name", "", "")
	cmd.Flags().String("phone", "", "")
	cmd.Flags().String("locale", "", "")
	cmd.Flags().String("permissions", "", "")
	cmd.AddCommand(staffsUpdateCmd)

	err := staffsUpdateCmd.RunE(cmd, []string{"staff_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStaffsDeleteGetClientError verifies delete command error handling when getClient fails.
func TestStaffsDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()
	_ = os.Unsetenv("SHOPLINE_STORE")

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true") // Skip confirmation prompt
	cmd.AddCommand(staffsDeleteCmd)

	err := staffsDeleteCmd.RunE(cmd, []string{"staff_123"})
	if err == nil {
		t.Error("Expected error when getClient fails")
	}
	if !strings.Contains(err.Error(), "failed to open credential store") {
		t.Errorf("Expected 'failed to open credential store' error, got: %v", err)
	}
}

// TestStaffsListOwnerDisplay tests that owner status is displayed correctly.
func TestStaffsListOwnerDisplay(t *testing.T) {
	tests := []struct {
		name         string
		accountOwner bool
		wantOutput   string
	}{
		{
			name:         "account owner shows Yes",
			accountOwner: true,
			wantOutput:   "Yes",
		},
		{
			name:         "non-owner shows No",
			accountOwner: false,
			wantOutput:   "No",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &staffsMockAPIClient{
				listStaffsResp: &api.StaffsListResponse{
					Items: []api.Staff{
						{
							ID:           "staff_owner",
							Email:        "owner@example.com",
							FirstName:    "Owner",
							LastName:     "Test",
							AccountOwner: tt.accountOwner,
							Permissions:  []string{"all"},
							CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
					TotalCount: 1,
				},
			}
			cleanup, buf := setupStaffsMockFactories(mockClient)
			defer cleanup()

			cmd := newStaffsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := staffsListCmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestStaffsGetWithPhone tests that phone is displayed when present.
func TestStaffsGetWithPhone(t *testing.T) {
	mockClient := &staffsMockAPIClient{
		getStaffResp: &api.Staff{
			ID:           "staff_phone",
			Email:        "phone@example.com",
			FirstName:    "Phone",
			LastName:     "Test",
			Phone:        "+1234567890",
			AccountOwner: false,
			Locale:       "en",
			Permissions:  []string{"orders"},
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	cleanup, _ := setupStaffsMockFactories(mockClient)
	defer cleanup()

	cmd := newStaffsTestCmd()

	err := staffsGetCmd.RunE(cmd, []string{"staff_phone"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

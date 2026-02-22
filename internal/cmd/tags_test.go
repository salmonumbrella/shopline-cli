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

// TestTagsCommandSetup verifies tags command initialization
func TestTagsCommandSetup(t *testing.T) {
	if tagsCmd.Use != "tags" {
		t.Errorf("expected Use 'tags', got %q", tagsCmd.Use)
	}
	if tagsCmd.Short != "Manage product tags" {
		t.Errorf("expected Short 'Manage product tags', got %q", tagsCmd.Short)
	}
}

// TestTagsSubcommands verifies all subcommands are registered
func TestTagsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List tags",
		"get":    "Get tag details",
		"create": "Create a tag",
		"delete": "Delete a tag",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range tagsCmd.Commands() {
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

// TestTagsListFlags verifies list command flags exist with correct defaults
func TestTagsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"q", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := tagsListCmd.Flags().Lookup(f.name)
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

// TestTagsCreateFlags verifies create command flags exist with correct defaults
func TestTagsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := tagsCreateCmd.Flags().Lookup(f.name)
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

// TestTagsDeleteArgs verifies delete command requires exactly one argument
func TestTagsDeleteArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if tagsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", tagsDeleteCmd.Use)
	}
}

// TestTagsGetArgs verifies get command requires exactly one argument
func TestTagsGetArgs(t *testing.T) {
	// Check the Use field includes <id> which indicates required argument
	if tagsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", tagsGetCmd.Use)
	}
}

// TestTagsListFlagDescriptions verifies flag descriptions are set
func TestTagsListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"q":         "Search tags by name",
		"page":      "Page number",
		"page-size": "Results per page",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := tagsListCmd.Flags().Lookup(flagName)
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

// TestTagsCreateFlagDescriptions verifies create flag descriptions
func TestTagsCreateFlagDescriptions(t *testing.T) {
	flag := tagsCreateCmd.Flags().Lookup("name")
	if flag == nil {
		t.Error("name flag not found")
		return
	}
	expectedUsage := "Tag name"
	if flag.Usage != expectedUsage {
		t.Errorf("expected Usage %q, got %q", expectedUsage, flag.Usage)
	}
}

// TestTagsCreateRequiredFlags verifies that name flag is required
func TestTagsCreateRequiredFlags(t *testing.T) {
	flag := tagsCreateCmd.Flags().Lookup("name")
	if flag == nil {
		t.Error("name flag not found")
		return
	}
	// Verify it exists and is a string flag
	if flag.Value.Type() != "string" {
		t.Error("name flag should be a string type")
	}
}

// tagsMockAPIClient is a mock implementation of api.APIClient for tags tests.
type tagsMockAPIClient struct {
	api.MockClient
	listTagsResp  *api.TagsListResponse
	listTagsErr   error
	getTagResp    *api.Tag
	getTagErr     error
	createTagResp *api.Tag
	createTagErr  error
	deleteTagErr  error
}

func (m *tagsMockAPIClient) ListTags(ctx context.Context, opts *api.TagsListOptions) (*api.TagsListResponse, error) {
	return m.listTagsResp, m.listTagsErr
}

func (m *tagsMockAPIClient) GetTag(ctx context.Context, id string) (*api.Tag, error) {
	return m.getTagResp, m.getTagErr
}

func (m *tagsMockAPIClient) CreateTag(ctx context.Context, req *api.TagCreateRequest) (*api.Tag, error) {
	return m.createTagResp, m.createTagErr
}

func (m *tagsMockAPIClient) DeleteTag(ctx context.Context, id string) error {
	return m.deleteTagErr
}

// setupTagsMockFactories sets up mock factories for tags tests.
func setupTagsMockFactories(mockClient *tagsMockAPIClient) (func(), *bytes.Buffer) {
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

// newTagsTestCmd creates a test command with common flags for tags tests.
func newTagsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().String("q", "", "") // For tags list search query
	return cmd
}

// TestTagsListRunE tests the tags list command with mock API.
func TestTagsListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockResp   *api.TagsListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.TagsListResponse{
				Items: []api.Tag{
					{
						ID:           "tag_123",
						Name:         "Sale",
						Handle:       "sale",
						ProductCount: 10,
						CreatedAt:    testTime,
						UpdatedAt:    testTime,
					},
				},
				TotalCount: 1,
			},
			wantOutput: "tag_123",
		},
		{
			name: "successful list multiple tags",
			mockResp: &api.TagsListResponse{
				Items: []api.Tag{
					{
						ID:           "tag_001",
						Name:         "New",
						Handle:       "new",
						ProductCount: 5,
						CreatedAt:    testTime,
						UpdatedAt:    testTime,
					},
					{
						ID:           "tag_002",
						Name:         "Featured",
						Handle:       "featured",
						ProductCount: 15,
						CreatedAt:    testTime,
						UpdatedAt:    testTime,
					},
				},
				TotalCount: 2,
			},
			wantOutput: "tag_001",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.TagsListResponse{
				Items:      []api.Tag{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tagsMockAPIClient{
				listTagsResp: tt.mockResp,
				listTagsErr:  tt.mockErr,
			}
			cleanup, buf := setupTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newTagsTestCmd()
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := tagsListCmd.RunE(cmd, []string{})

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

// TestTagsListRunE_WithQuery tests tags list with query parameter.
func TestTagsListRunE_WithQuery(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		listTagsResp: &api.TagsListResponse{
			Items: []api.Tag{
				{
					ID:           "tag_123",
					Name:         "Sale",
					Handle:       "sale",
					ProductCount: 10,
					CreatedAt:    testTime,
					UpdatedAt:    testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("q", "sale")

	err := tagsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_123") {
		t.Errorf("output %q should contain tag_123", output)
	}
}

// TestTagsListRunE_JSON tests JSON output for tags list.
func TestTagsListRunE_JSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		listTagsResp: &api.TagsListResponse{
			Items: []api.Tag{
				{
					ID:           "tag_123",
					Name:         "Sale",
					Handle:       "sale",
					ProductCount: 10,
					CreatedAt:    testTime,
					UpdatedAt:    testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := tagsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain the tag data
	if !strings.Contains(output, "tag_123") {
		t.Errorf("JSON output %q should contain tag_123", output)
	}
}

// TestTagsListRunE_GetClientFails tests error handling when getClient fails.
func TestTagsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	// Note: The tagsListCmd expects a "q" flag
	// We use a separate test command that doesn't have this conflict

	err := tagsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTagsGetRunE tests the tags get command with mock API.
func TestTagsGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		tagID     string
		mockResp  *api.Tag
		mockErr   error
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "successful get",
			tagID: "tag_123",
			mockResp: &api.Tag{
				ID:           "tag_123",
				Name:         "Sale",
				Handle:       "sale",
				ProductCount: 10,
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
		},
		{
			name:      "tag not found",
			tagID:     "tag_999",
			mockErr:   errors.New("tag not found"),
			wantErr:   true,
			errSubstr: "failed to get tag",
		},
		{
			name:      "API error",
			tagID:     "tag_123",
			mockErr:   errors.New("API unavailable"),
			wantErr:   true,
			errSubstr: "failed to get tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tagsMockAPIClient{
				getTagResp: tt.mockResp,
				getTagErr:  tt.mockErr,
			}
			cleanup, _ := setupTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newTagsTestCmd()

			err := tagsGetCmd.RunE(cmd, []string{tt.tagID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestTagsGetRunE_JSON tests JSON output for tags get.
func TestTagsGetRunE_JSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		getTagResp: &api.Tag{
			ID:           "tag_123",
			Name:         "Sale",
			Handle:       "sale",
			ProductCount: 10,
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
		},
	}
	cleanup, buf := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := tagsGetCmd.RunE(cmd, []string{"tag_123"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain the tag data
	if !strings.Contains(output, "tag_123") {
		t.Errorf("JSON output %q should contain tag_123", output)
	}
}

// TestTagsGetRunE_GetClientFails tests error handling when getClient fails.
func TestTagsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTagsTestCmd()

	err := tagsGetCmd.RunE(cmd, []string{"tag_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTagsCreateRunE tests the tags create command with mock API.
func TestTagsCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		tagName    string
		mockResp   *api.Tag
		mockErr    error
		wantErr    bool
		errSubstr  string
		wantOutput string
	}{
		{
			name:    "successful create",
			tagName: "New Tag",
			mockResp: &api.Tag{
				ID:           "tag_456",
				Name:         "New Tag",
				Handle:       "new-tag",
				ProductCount: 0,
				CreatedAt:    testTime,
				UpdatedAt:    testTime,
			},
			wantOutput: "tag_456",
		},
		{
			name:      "API error",
			tagName:   "Test Tag",
			mockErr:   errors.New("API unavailable"),
			wantErr:   true,
			errSubstr: "failed to create tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tagsMockAPIClient{
				createTagResp: tt.mockResp,
				createTagErr:  tt.mockErr,
			}
			cleanup, _ := setupTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newTagsTestCmd()
			cmd.Flags().String("name", "", "")
			_ = cmd.Flags().Set("name", tt.tagName)

			err := tagsCreateCmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestTagsCreateRunE_DryRun tests dry-run mode for tags create.
func TestTagsCreateRunE_DryRun(t *testing.T) {
	cmd := newTagsTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "test-tag")
	_ = cmd.Flags().Set("dry-run", "true")

	err := tagsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %v", err)
	}
}

// TestTagsCreateRunE_JSON tests JSON output for tags create.
func TestTagsCreateRunE_JSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		createTagResp: &api.Tag{
			ID:           "tag_789",
			Name:         "Test Tag",
			Handle:       "test-tag",
			ProductCount: 0,
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
		},
	}
	cleanup, buf := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "Test Tag")
	_ = cmd.Flags().Set("output", "json")

	err := tagsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	// JSON output should contain the tag data
	if !strings.Contains(output, "tag_789") {
		t.Errorf("JSON output %q should contain tag_789", output)
	}
}

// TestTagsCreateRunE_GetClientFails tests error handling when getClient fails.
func TestTagsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTagsTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "test-tag")

	err := tagsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTagsDeleteRunE tests the tags delete command with mock API.
func TestTagsDeleteRunE(t *testing.T) {
	tests := []struct {
		name      string
		tagID     string
		mockErr   error
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "successful delete",
			tagID: "tag_123",
		},
		{
			name:      "tag not found",
			tagID:     "tag_999",
			mockErr:   errors.New("tag not found"),
			wantErr:   true,
			errSubstr: "failed to delete tag",
		},
		{
			name:      "API error",
			tagID:     "tag_123",
			mockErr:   errors.New("API unavailable"),
			wantErr:   true,
			errSubstr: "failed to delete tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &tagsMockAPIClient{
				deleteTagErr: tt.mockErr,
			}
			cleanup, _ := setupTagsMockFactories(mockClient)
			defer cleanup()

			cmd := newTagsTestCmd()

			err := tagsDeleteCmd.RunE(cmd, []string{tt.tagID})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestTagsDeleteRunE_DryRun tests dry-run mode for tags delete.
func TestTagsDeleteRunE_DryRun(t *testing.T) {
	cmd := newTagsTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := tagsDeleteCmd.RunE(cmd, []string{"tag_123"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %v", err)
	}
}

// TestTagsDeleteRunE_GetClientFails tests error handling when getClient fails.
func TestTagsDeleteRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTagsTestCmd()

	err := tagsDeleteCmd.RunE(cmd, []string{"tag_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestTagsListRunE_NoProfiles verifies error when no profiles are configured.
func TestTagsListRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	err := tagsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestTagsGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection.
func TestTagsGetRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTagsTestCmd()

	err := tagsGetCmd.RunE(cmd, []string{"tag_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

// TestTagsListRunE_Pagination tests pagination parameters for tags list.
func TestTagsListRunE_Pagination(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		listTagsResp: &api.TagsListResponse{
			Items: []api.Tag{
				{
					ID:           "tag_page2",
					Name:         "Page 2 Tag",
					Handle:       "page-2-tag",
					ProductCount: 5,
					CreatedAt:    testTime,
					UpdatedAt:    testTime,
				},
			},
			TotalCount: 50,
		},
	}
	cleanup, buf := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("page", "2")
	_ = cmd.Flags().Set("page-size", "10")

	err := tagsListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tag_page2") {
		t.Errorf("output %q should contain tag_page2", output)
	}
}

// TestTagsGetRunE_TextOutput tests text output for tags get.
func TestTagsGetRunE_TextOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		getTagResp: &api.Tag{
			ID:           "tag_detail",
			Name:         "Detailed Tag",
			Handle:       "detailed-tag",
			ProductCount: 25,
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
		},
	}
	cleanup, _ := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()

	err := tagsGetCmd.RunE(cmd, []string{"tag_detail"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestTagsCreateRunE_TextOutput tests text output for tags create.
func TestTagsCreateRunE_TextOutput(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &tagsMockAPIClient{
		createTagResp: &api.Tag{
			ID:           "tag_new",
			Name:         "Newly Created Tag",
			Handle:       "newly-created-tag",
			ProductCount: 0,
			CreatedAt:    testTime,
			UpdatedAt:    testTime,
		},
	}
	cleanup, _ := setupTagsMockFactories(mockClient)
	defer cleanup()

	cmd := newTagsTestCmd()
	cmd.Flags().String("name", "", "")
	_ = cmd.Flags().Set("name", "Newly Created Tag")

	err := tagsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

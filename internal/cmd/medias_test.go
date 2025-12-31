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

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string equal to max",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "string longer than max",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "long text truncated",
			input:    "This is a very long description text",
			maxLen:   15,
			expected: "This is a ve...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestMediasCommandSetup verifies medias command initialization
func TestMediasCommandSetup(t *testing.T) {
	if mediasCmd.Use != "medias" {
		t.Errorf("expected Use 'medias', got %q", mediasCmd.Use)
	}
	if mediasCmd.Short != "Manage product media files" {
		t.Errorf("expected Short 'Manage product media files', got %q", mediasCmd.Short)
	}
}

// TestMediasSubcommands verifies all subcommands are registered
func TestMediasSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List medias",
		"get":    "Get media details",
		"create": "Create a media",
		"delete": "Delete a media",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range mediasCmd.Commands() {
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

// TestMediasListFlags verifies list command flags exist with correct defaults
func TestMediasListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"product-id", ""},
		{"type", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := mediasListCmd.Flags().Lookup(f.name)
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

// TestMediasListFlagDescriptions verifies flag descriptions are set
func TestMediasListFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"product-id": "Filter by product ID",
		"type":       "Filter by media type (image, video, model_3d, external_video)",
		"page":       "Page number",
		"page-size":  "Results per page",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := mediasListCmd.Flags().Lookup(flagName)
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

// TestMediasCreateFlags verifies create command flags exist with correct defaults
func TestMediasCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"product-id", ""},
		{"type", "image"},
		{"src", ""},
		{"alt", ""},
		{"position", "0"},
		{"external-url", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := mediasCreateCmd.Flags().Lookup(f.name)
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

// TestMediasCreateFlagDescriptions verifies create flag descriptions
func TestMediasCreateFlagDescriptions(t *testing.T) {
	flags := map[string]string{
		"product-id":   "Product ID (required)",
		"type":         "Media type (image, video, model_3d, external_video)",
		"src":          "Media source URL",
		"alt":          "Alt text for the media",
		"position":     "Position in the media list",
		"external-url": "External URL for external_video type",
	}

	for flagName, expectedUsage := range flags {
		t.Run(flagName, func(t *testing.T) {
			flag := mediasCreateCmd.Flags().Lookup(flagName)
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

// TestMediasCreateRequiredFlags verifies product-id is required
func TestMediasCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"product-id"}

	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := mediasCreateCmd.Flags().Lookup(name)
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

// TestMediasGetArgs verifies get command requires exactly 1 argument
func TestMediasGetArgs(t *testing.T) {
	err := mediasGetCmd.Args(mediasGetCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = mediasGetCmd.Args(mediasGetCmd, []string{"media-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = mediasGetCmd.Args(mediasGetCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when multiple args provided")
	}
}

// TestMediasDeleteArgs verifies delete command requires exactly 1 argument
func TestMediasDeleteArgs(t *testing.T) {
	err := mediasDeleteCmd.Args(mediasDeleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = mediasDeleteCmd.Args(mediasDeleteCmd, []string{"media-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = mediasDeleteCmd.Args(mediasDeleteCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when multiple args provided")
	}
}

// TestMediasGetClientError verifies error handling when getClient fails
func TestMediasGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestMediasListGetClientError verifies list command error handling when getClient fails
func TestMediasListGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(mediasListCmd)

	err := mediasListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMediasGetGetClientError verifies get command error handling when getClient fails
func TestMediasGetGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.AddCommand(mediasGetCmd)

	err := mediasGetCmd.RunE(cmd, []string{"media-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMediasDeleteGetClientError verifies delete command error handling when getClient fails
func TestMediasDeleteGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("yes", "true")
	cmd.AddCommand(mediasDeleteCmd)

	err := mediasDeleteCmd.RunE(cmd, []string{"media-id"})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMediasWithMockStore tests medias commands with a mock credential store
func TestMediasWithMockStore(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")

	store := &mockStore{
		names: []string{"teststore"},
		creds: map[string]*secrets.StoreCredentials{
			"teststore": {Handle: "test-handle", AccessToken: "test-token"},
		},
	}

	secretsStoreFactory = func() (CredentialStore, error) {
		return store, nil
	}

	cmd := newTestCmdWithFlags()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// mediasMockAPIClient is a mock implementation of api.APIClient for medias tests.
type mediasMockAPIClient struct {
	api.MockClient
	listMediasResp  *api.MediasListResponse
	listMediasErr   error
	getMediaResp    *api.Media
	getMediaErr     error
	createMediaResp *api.Media
	createMediaErr  error
	deleteMediaErr  error
}

func (m *mediasMockAPIClient) ListMedias(ctx context.Context, opts *api.MediasListOptions) (*api.MediasListResponse, error) {
	return m.listMediasResp, m.listMediasErr
}

func (m *mediasMockAPIClient) GetMedia(ctx context.Context, id string) (*api.Media, error) {
	return m.getMediaResp, m.getMediaErr
}

func (m *mediasMockAPIClient) CreateMedia(ctx context.Context, req *api.MediaCreateRequest) (*api.Media, error) {
	return m.createMediaResp, m.createMediaErr
}

func (m *mediasMockAPIClient) DeleteMedia(ctx context.Context, id string) error {
	return m.deleteMediaErr
}

// setupMediasMockFactories sets up mock factories for medias tests.
func setupMediasMockFactories(mockClient *mediasMockAPIClient) (func(), *bytes.Buffer) {
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

// newMediasTestCmd creates a test command with common flags for medias tests.
func newMediasTestCmd() *cobra.Command {
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

// TestMediasListRunE tests the medias list command with mock API.
func TestMediasListRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		mockResp *api.MediasListResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "successful list with image",
			mockResp: &api.MediasListResponse{
				Items: []api.Media{
					{
						ID:        "media_123",
						ProductID: "prod_456",
						MediaType: api.MediaTypeImage,
						Alt:       "Product image",
						Width:     800,
						Height:    600,
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name: "successful list with video",
			mockResp: &api.MediasListResponse{
				Items: []api.Media{
					{
						ID:        "media_789",
						ProductID: "prod_456",
						MediaType: api.MediaTypeVideo,
						Alt:       "Product video",
						Width:     1920,
						Height:    1080,
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name: "successful list without dimensions",
			mockResp: &api.MediasListResponse{
				Items: []api.Media{
					{
						ID:        "media_abc",
						ProductID: "prod_123",
						MediaType: api.MediaTypeExternal,
						Alt:       "External video",
						CreatedAt: testTime,
					},
				},
				TotalCount: 1,
			},
		},
		{
			name: "successful list with long alt text truncation",
			mockResp: &api.MediasListResponse{
				Items: []api.Media{
					{
						ID:        "media_long",
						ProductID: "prod_789",
						MediaType: api.MediaTypeImage,
						Alt:       "This is a very long alt text that should be truncated",
						CreatedAt: testTime,
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
			mockResp: &api.MediasListResponse{
				Items:      []api.Media{},
				TotalCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mediasMockAPIClient{
				listMediasResp: tt.mockResp,
				listMediasErr:  tt.mockErr,
			}
			cleanup, _ := setupMediasMockFactories(mockClient)
			defer cleanup()

			cmd := newMediasTestCmd()
			cmd.Flags().String("product-id", "", "")
			cmd.Flags().String("type", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			err := mediasListCmd.RunE(cmd, []string{})

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

// TestMediasListRunEWithJSON tests the medias list command with JSON output.
func TestMediasListRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &mediasMockAPIClient{
		listMediasResp: &api.MediasListResponse{
			Items: []api.Media{
				{
					ID:        "media_json",
					ProductID: "prod_123",
					MediaType: api.MediaTypeImage,
					Alt:       "JSON test",
					CreatedAt: testTime,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupMediasMockFactories(mockClient)
	defer cleanup()

	cmd := newMediasTestCmd()
	cmd.Flags().String("product-id", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := mediasListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "media_json") {
		t.Errorf("JSON output should contain media ID, got: %s", output)
	}
}

// TestMediasGetRunE tests the medias get command with mock API.
func TestMediasGetRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name     string
		mediaID  string
		mockResp *api.Media
		mockErr  error
		wantErr  bool
	}{
		{
			name:    "successful get with full details",
			mediaID: "media_123",
			mockResp: &api.Media{
				ID:          "media_123",
				ProductID:   "prod_456",
				MediaType:   api.MediaTypeImage,
				Position:    1,
				Alt:         "Product image",
				Src:         "https://example.com/image.jpg",
				Width:       800,
				Height:      600,
				MimeType:    "image/jpeg",
				FileSize:    102400,
				Duration:    0,
				PreviewURL:  "",
				ExternalURL: "",
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
		},
		{
			name:    "successful get with video details",
			mediaID: "media_video",
			mockResp: &api.Media{
				ID:         "media_video",
				ProductID:  "prod_789",
				MediaType:  api.MediaTypeVideo,
				Position:   2,
				Alt:        "Product video",
				Src:        "https://example.com/video.mp4",
				Width:      1920,
				Height:     1080,
				MimeType:   "video/mp4",
				FileSize:   10485760,
				Duration:   120,
				PreviewURL: "https://example.com/preview.jpg",
				CreatedAt:  testTime,
				UpdatedAt:  testTime,
			},
		},
		{
			name:    "successful get with external video",
			mediaID: "media_external",
			mockResp: &api.Media{
				ID:          "media_external",
				ProductID:   "prod_ext",
				MediaType:   api.MediaTypeExternal,
				Position:    3,
				Alt:         "External video",
				ExternalURL: "https://youtube.com/watch?v=abc123",
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
		},
		{
			name:    "successful get without optional fields",
			mediaID: "media_minimal",
			mockResp: &api.Media{
				ID:        "media_minimal",
				ProductID: "prod_min",
				MediaType: api.MediaTypeImage,
				Position:  1,
				Alt:       "Minimal image",
				Src:       "https://example.com/minimal.jpg",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name:    "media not found",
			mediaID: "media_999",
			mockErr: errors.New("media not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mediasMockAPIClient{
				getMediaResp: tt.mockResp,
				getMediaErr:  tt.mockErr,
			}
			cleanup, _ := setupMediasMockFactories(mockClient)
			defer cleanup()

			cmd := newMediasTestCmd()

			err := mediasGetCmd.RunE(cmd, []string{tt.mediaID})

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

// TestMediasGetRunEWithJSON tests the medias get command with JSON output.
func TestMediasGetRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &mediasMockAPIClient{
		getMediaResp: &api.Media{
			ID:        "media_json_get",
			ProductID: "prod_123",
			MediaType: api.MediaTypeImage,
			Alt:       "JSON test",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}
	cleanup, buf := setupMediasMockFactories(mockClient)
	defer cleanup()

	cmd := newMediasTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := mediasGetCmd.RunE(cmd, []string{"media_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "media_json_get") {
		t.Errorf("JSON output should contain media ID, got: %s", output)
	}
}

// TestMediasCreateRunE tests the medias create command with mock API.
func TestMediasCreateRunE(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		productID string
		mediaType string
		src       string
		alt       string
		position  int
		extURL    string
		mockResp  *api.Media
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful create image",
			productID: "prod_123",
			mediaType: "image",
			src:       "https://example.com/new-image.jpg",
			alt:       "New product image",
			position:  1,
			mockResp: &api.Media{
				ID:        "media_new",
				ProductID: "prod_123",
				MediaType: api.MediaTypeImage,
				Src:       "https://example.com/new-image.jpg",
				Alt:       "New product image",
				CreatedAt: testTime,
			},
		},
		{
			name:      "successful create video",
			productID: "prod_456",
			mediaType: "video",
			src:       "https://example.com/new-video.mp4",
			mockResp: &api.Media{
				ID:        "media_video_new",
				ProductID: "prod_456",
				MediaType: api.MediaTypeVideo,
				Src:       "https://example.com/new-video.mp4",
				CreatedAt: testTime,
			},
		},
		{
			name:      "successful create external video",
			productID: "prod_789",
			mediaType: "external_video",
			extURL:    "https://youtube.com/watch?v=xyz",
			mockResp: &api.Media{
				ID:          "media_ext_new",
				ProductID:   "prod_789",
				MediaType:   api.MediaTypeExternal,
				ExternalURL: "https://youtube.com/watch?v=xyz",
				CreatedAt:   testTime,
			},
		},
		{
			name:      "create error",
			productID: "prod_fail",
			mediaType: "image",
			src:       "https://example.com/fail.jpg",
			mockErr:   errors.New("creation failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mediasMockAPIClient{
				createMediaResp: tt.mockResp,
				createMediaErr:  tt.mockErr,
			}
			cleanup, _ := setupMediasMockFactories(mockClient)
			defer cleanup()

			cmd := newMediasTestCmd()
			cmd.Flags().String("product-id", tt.productID, "")
			cmd.Flags().String("type", tt.mediaType, "")
			cmd.Flags().String("src", tt.src, "")
			cmd.Flags().String("alt", tt.alt, "")
			cmd.Flags().Int("position", tt.position, "")
			cmd.Flags().String("external-url", tt.extURL, "")

			err := mediasCreateCmd.RunE(cmd, []string{})

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

// TestMediasCreateRunEWithJSON tests the medias create command with JSON output.
func TestMediasCreateRunEWithJSON(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockClient := &mediasMockAPIClient{
		createMediaResp: &api.Media{
			ID:        "media_json_create",
			ProductID: "prod_json",
			MediaType: api.MediaTypeImage,
			Src:       "https://example.com/json.jpg",
			CreatedAt: testTime,
		},
	}
	cleanup, buf := setupMediasMockFactories(mockClient)
	defer cleanup()

	cmd := newMediasTestCmd()
	cmd.Flags().String("product-id", "prod_json", "")
	cmd.Flags().String("type", "image", "")
	cmd.Flags().String("src", "https://example.com/json.jpg", "")
	cmd.Flags().String("alt", "", "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().String("external-url", "", "")
	_ = cmd.Flags().Set("output", "json")

	err := mediasCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "media_json_create") {
		t.Errorf("JSON output should contain media ID, got: %s", output)
	}
}

// TestMediasCreateDryRun tests the medias create command with dry-run flag.
func TestMediasCreateDryRun(t *testing.T) {
	cleanup, _ := setupMediasMockFactories(&mediasMockAPIClient{})
	defer cleanup()

	cmd := newMediasTestCmd()
	cmd.Flags().String("product-id", "prod_dry", "")
	cmd.Flags().String("type", "image", "")
	cmd.Flags().String("src", "https://example.com/dry.jpg", "")
	cmd.Flags().String("alt", "", "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().String("external-url", "", "")
	_ = cmd.Flags().Set("dry-run", "true")

	err := mediasCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMediasCreateGetClientError verifies create command error handling when getClient fails
func TestMediasCreateGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newMediasTestCmd()
	cmd.Flags().String("product-id", "prod_123", "")
	cmd.Flags().String("type", "image", "")
	cmd.Flags().String("src", "", "")
	cmd.Flags().String("alt", "", "")
	cmd.Flags().Int("position", 0, "")
	cmd.Flags().String("external-url", "", "")

	err := mediasCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when getClient fails")
	}
}

// TestMediasDeleteRunE tests the medias delete command with mock API.
func TestMediasDeleteRunE(t *testing.T) {
	tests := []struct {
		name    string
		mediaID string
		yes     bool
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful delete with confirmation",
			mediaID: "media_del",
			yes:     true,
		},
		{
			name:    "delete without confirmation",
			mediaID: "media_no_confirm",
			yes:     false,
		},
		{
			name:    "delete error",
			mediaID: "media_fail",
			yes:     true,
			mockErr: errors.New("deletion failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mediasMockAPIClient{
				deleteMediaErr: tt.mockErr,
			}
			cleanup, _ := setupMediasMockFactories(mockClient)
			defer cleanup()

			cmd := newMediasTestCmd()
			if tt.yes {
				_ = cmd.Flags().Set("yes", "true")
			} else {
				_ = cmd.Flags().Set("yes", "false")
			}

			err := mediasDeleteCmd.RunE(cmd, []string{tt.mediaID})

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

// TestMediasDeleteDryRun tests the medias delete command with dry-run flag.
func TestMediasDeleteDryRun(t *testing.T) {
	cleanup, _ := setupMediasMockFactories(&mediasMockAPIClient{})
	defer cleanup()

	cmd := newMediasTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	err := mediasDeleteCmd.RunE(cmd, []string{"media_dry_del"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

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

// TestFilesCommandSetup verifies files command initialization.
func TestFilesCommandSetup(t *testing.T) {
	if filesCmd.Use != "files" {
		t.Errorf("expected Use 'files', got %q", filesCmd.Use)
	}
	if filesCmd.Short != "Manage files" {
		t.Errorf("expected Short 'Manage files', got %q", filesCmd.Short)
	}
}

// TestFilesSubcommands verifies all subcommands are registered.
func TestFilesSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List files",
		"get":    "Get file details",
		"create": "Create a file",
		"delete": "Delete a file",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range filesCmd.Commands() {
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

// TestFilesListFlags verifies list command flags exist with correct defaults.
func TestFilesListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"content-type", ""},
		{"status", ""},
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := filesListCmd.Flags().Lookup(f.name)
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

// TestFilesCreateFlags verifies create command flags exist with correct defaults.
func TestFilesCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"filename", ""},
		{"url", ""},
		{"alt", ""},
		{"content-type", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := filesCreateCmd.Flags().Lookup(f.name)
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

// TestFilesGetArgs verifies get command requires exactly 1 argument.
func TestFilesGetArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"file_123"}, wantErr: false},
		{name: "too many args", args: []string{"file_123", "file_456"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filesGetCmd.Args(filesGetCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFilesDeleteArgs verifies delete command requires exactly 1 argument.
func TestFilesDeleteArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"file_123"}, wantErr: false},
		{name: "too many args", args: []string{"file_123", "file_456"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filesDeleteCmd.Args(filesDeleteCmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFilesGetClientError verifies error handling when getClient fails.
func TestFilesGetClientError(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newFilesTestCmd()
	_, err := getClient(cmd)
	if err == nil {
		t.Error("expected error when credential store fails")
	}
}

// TestFilesWithMockStore tests files commands with a mock credential store.
func TestFilesWithMockStore(t *testing.T) {
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

	cmd := newFilesTestCmd()
	client, err := getClient(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

// TestFormatFileSize verifies file size formatting.
func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "zero bytes",
			size:     0,
			expected: "0 B",
		},
		{
			name:     "bytes below unit",
			size:     500,
			expected: "500 B",
		},
		{
			name:     "exactly 1 KB",
			size:     1024,
			expected: "1.0 KB",
		},
		{
			name:     "1.5 KB",
			size:     1536,
			expected: "1.5 KB",
		},
		{
			name:     "exactly 1 MB",
			size:     1048576,
			expected: "1.0 MB",
		},
		{
			name:     "exactly 1 GB",
			size:     1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "exactly 1 TB",
			size:     1099511627776,
			expected: "1.0 TB",
		},
		{
			name:     "1 byte",
			size:     1,
			expected: "1 B",
		},
		{
			name:     "1023 bytes",
			size:     1023,
			expected: "1023 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %q, want %q", tt.size, result, tt.expected)
			}
		})
	}
}

// filesMockAPIClient is a mock implementation of api.APIClient for files tests.
type filesMockAPIClient struct {
	api.MockClient
	listFilesResp  *api.FilesListResponse
	listFilesErr   error
	getFileResp    *api.File
	getFileErr     error
	createFileResp *api.File
	createFileErr  error
	deleteFileErr  error
}

func (m *filesMockAPIClient) ListFiles(ctx context.Context, opts *api.FilesListOptions) (*api.FilesListResponse, error) {
	return m.listFilesResp, m.listFilesErr
}

func (m *filesMockAPIClient) GetFile(ctx context.Context, id string) (*api.File, error) {
	return m.getFileResp, m.getFileErr
}

func (m *filesMockAPIClient) CreateFile(ctx context.Context, req *api.FileCreateRequest) (*api.File, error) {
	return m.createFileResp, m.createFileErr
}

func (m *filesMockAPIClient) DeleteFile(ctx context.Context, id string) error {
	return m.deleteFileErr
}

// setupFilesMockFactories sets up mock factories for files tests.
func setupFilesMockFactories(mockClient *filesMockAPIClient) (func(), *bytes.Buffer) {
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

// newFilesTestCmd creates a test command with common flags for files tests.
func newFilesTestCmd() *cobra.Command {
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

// captureStdout captures output during execution via formatterWriter and returns it.
func captureStdout(fn func() error) (string, error) {
	var buf bytes.Buffer
	formatterWriter = &buf
	err := fn()
	return buf.String(), err
}

// TestFilesListRunE tests the files list command with mock API.
func TestFilesListRunE(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   *api.FilesListResponse
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name: "successful list",
			mockResp: &api.FilesListResponse{
				Items: []api.File{
					{
						ID:        "file_123",
						Filename:  "image.png",
						MimeType:  "image/png",
						FileSize:  1024,
						Status:    api.FileStatusReady,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
				},
				TotalCount: 1,
			},
			wantOutput: "file_123",
		},
		{
			name:    "API error",
			mockErr: errors.New("API unavailable"),
			wantErr: true,
		},
		{
			name: "empty list",
			mockResp: &api.FilesListResponse{
				Items:      []api.File{},
				TotalCount: 0,
			},
		},
		{
			name: "multiple files",
			mockResp: &api.FilesListResponse{
				Items: []api.File{
					{
						ID:        "file_1",
						Filename:  "photo1.jpg",
						MimeType:  "image/jpeg",
						FileSize:  2048,
						Status:    api.FileStatusReady,
						CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					},
					{
						ID:        "file_2",
						Filename:  "document.pdf",
						MimeType:  "application/pdf",
						FileSize:  10240,
						Status:    api.FileStatusPending,
						CreatedAt: time.Date(2024, 1, 16, 11, 45, 0, 0, time.UTC),
					},
				},
				TotalCount: 2,
			},
			wantOutput: "file_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &filesMockAPIClient{
				listFilesResp: tt.mockResp,
				listFilesErr:  tt.mockErr,
			}
			cleanup, buf := setupFilesMockFactories(mockClient)
			defer cleanup()

			cmd := newFilesTestCmd()
			cmd.Flags().String("content-type", "", "")
			cmd.Flags().String("status", "", "")
			cmd.Flags().Int("page", 1, "")
			cmd.Flags().Int("page-size", 20, "")

			output, err := captureStdout(func() error {
				return filesListCmd.RunE(cmd, []string{})
			})

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

			// Check both formatter buffer and stdout capture
			combinedOutput := buf.String() + output
			if tt.wantOutput != "" && !strings.Contains(combinedOutput, tt.wantOutput) {
				t.Errorf("output %q should contain %q", combinedOutput, tt.wantOutput)
			}
		})
	}
}

// TestFilesListRunE_JSONOutput tests JSON output for files list command.
func TestFilesListRunE_JSONOutput(t *testing.T) {
	mockClient := &filesMockAPIClient{
		listFilesResp: &api.FilesListResponse{
			Items: []api.File{
				{
					ID:       "file_json",
					Filename: "test.png",
					MimeType: "image/png",
					FileSize: 512,
					Status:   api.FileStatusReady,
				},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupFilesMockFactories(mockClient)
	defer cleanup()

	cmd := newFilesTestCmd()
	cmd.Flags().String("content-type", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	_ = cmd.Flags().Set("output", "json")

	err := filesListCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "file_json") {
		t.Errorf("JSON output should contain file_json, got: %s", output)
	}
}

// TestFilesGetRunE tests the files get command with mock API.
func TestFilesGetRunE(t *testing.T) {
	tests := []struct {
		name       string
		fileID     string
		mockResp   *api.File
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:   "successful get",
			fileID: "file_123",
			mockResp: &api.File{
				ID:          "file_123",
				Filename:    "product.jpg",
				MimeType:    "image/jpeg",
				FileSize:    2048,
				URL:         "https://cdn.shopline.com/files/file_123.jpg",
				Alt:         "Product image",
				Status:      api.FileStatusReady,
				ContentType: "image",
				Width:       800,
				Height:      600,
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "file_123",
		},
		{
			name:    "file not found",
			fileID:  "file_999",
			mockErr: errors.New("file not found"),
			wantErr: true,
		},
		{
			name:   "file without dimensions",
			fileID: "file_doc",
			mockResp: &api.File{
				ID:        "file_doc",
				Filename:  "document.pdf",
				MimeType:  "application/pdf",
				FileSize:  10240,
				URL:       "https://cdn.shopline.com/files/file_doc.pdf",
				Alt:       "",
				Status:    api.FileStatusReady,
				Width:     0,
				Height:    0,
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantOutput: "file_doc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &filesMockAPIClient{
				getFileResp: tt.mockResp,
				getFileErr:  tt.mockErr,
			}
			cleanup, _ := setupFilesMockFactories(mockClient)
			defer cleanup()

			cmd := newFilesTestCmd()

			output, err := captureStdout(func() error {
				return filesGetCmd.RunE(cmd, []string{tt.fileID})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestFilesGetRunE_JSONOutput tests JSON output for files get command.
func TestFilesGetRunE_JSONOutput(t *testing.T) {
	mockClient := &filesMockAPIClient{
		getFileResp: &api.File{
			ID:        "file_json_get",
			Filename:  "test.png",
			MimeType:  "image/png",
			FileSize:  512,
			Status:    api.FileStatusReady,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	cleanup, buf := setupFilesMockFactories(mockClient)
	defer cleanup()

	cmd := newFilesTestCmd()
	_ = cmd.Flags().Set("output", "json")

	err := filesGetCmd.RunE(cmd, []string{"file_json_get"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "file_json_get") {
		t.Errorf("JSON output should contain file_json_get, got: %s", output)
	}
}

// TestFilesCreateRunE tests the files create command with mock API.
func TestFilesCreateRunE(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		url        string
		alt        string
		mockResp   *api.File
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:     "successful create",
			filename: "new-product.jpg",
			url:      "https://example.com/image.jpg",
			alt:      "New product image",
			mockResp: &api.File{
				ID:       "file_new",
				Filename: "new-product.jpg",
				Status:   api.FileStatusPending,
			},
			wantOutput: "file_new",
		},
		{
			name:     "create error",
			filename: "bad.jpg",
			mockErr:  errors.New("failed to create file"),
			wantErr:  true,
		},
		{
			name:     "create with minimal fields",
			filename: "minimal.png",
			mockResp: &api.File{
				ID:       "file_min",
				Filename: "minimal.png",
				Status:   api.FileStatusProcessing,
			},
			wantOutput: "file_min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &filesMockAPIClient{
				createFileResp: tt.mockResp,
				createFileErr:  tt.mockErr,
			}
			cleanup, _ := setupFilesMockFactories(mockClient)
			defer cleanup()

			cmd := newFilesTestCmd()
			cmd.Flags().String("filename", "", "")
			cmd.Flags().String("url", "", "")
			cmd.Flags().String("alt", "", "")
			cmd.Flags().String("content-type", "", "")
			_ = cmd.Flags().Set("filename", tt.filename)
			if tt.url != "" {
				_ = cmd.Flags().Set("url", tt.url)
			}
			if tt.alt != "" {
				_ = cmd.Flags().Set("alt", tt.alt)
			}

			output, err := captureStdout(func() error {
				return filesCreateCmd.RunE(cmd, []string{})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestFilesCreateRunE_DryRun tests dry-run mode for files create command.
func TestFilesCreateRunE_DryRun(t *testing.T) {
	mockClient := &filesMockAPIClient{}
	cleanup, _ := setupFilesMockFactories(mockClient)
	defer cleanup()

	cmd := newFilesTestCmd()
	cmd.Flags().String("filename", "", "")
	cmd.Flags().String("url", "", "")
	cmd.Flags().String("alt", "", "")
	cmd.Flags().String("content-type", "", "")
	_ = cmd.Flags().Set("filename", "dryrun.jpg")
	_ = cmd.Flags().Set("dry-run", "true")

	output, err := captureStdout(func() error {
		return filesCreateCmd.RunE(cmd, []string{})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "DRY-RUN") {
		t.Errorf("output should contain DRY-RUN, got: %s", output)
	}
}

// TestFilesCreateRunE_JSONOutput tests JSON output for files create command.
func TestFilesCreateRunE_JSONOutput(t *testing.T) {
	mockClient := &filesMockAPIClient{
		createFileResp: &api.File{
			ID:       "file_json_create",
			Filename: "test.png",
			Status:   api.FileStatusPending,
		},
	}
	cleanup, buf := setupFilesMockFactories(mockClient)
	defer cleanup()

	cmd := newFilesTestCmd()
	cmd.Flags().String("filename", "", "")
	cmd.Flags().String("url", "", "")
	cmd.Flags().String("alt", "", "")
	cmd.Flags().String("content-type", "", "")
	_ = cmd.Flags().Set("filename", "test.png")
	_ = cmd.Flags().Set("output", "json")

	err := filesCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "file_json_create") {
		t.Errorf("JSON output should contain file_json_create, got: %s", output)
	}
}

// TestFilesDeleteRunE tests the files delete command with mock API.
func TestFilesDeleteRunE(t *testing.T) {
	tests := []struct {
		name       string
		fileID     string
		yes        bool
		mockErr    error
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "successful delete",
			fileID:     "file_123",
			yes:        true,
			wantOutput: "Deleted file file_123",
		},
		{
			name:    "delete error",
			fileID:  "file_999",
			yes:     true,
			mockErr: errors.New("file not found"),
			wantErr: true,
		},
		{
			name:       "no confirmation",
			fileID:     "file_123",
			yes:        false,
			wantOutput: "Are you sure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &filesMockAPIClient{
				deleteFileErr: tt.mockErr,
			}
			cleanup, _ := setupFilesMockFactories(mockClient)
			defer cleanup()

			cmd := newFilesTestCmd()
			_ = cmd.Flags().Set("yes", "false")
			if tt.yes {
				_ = cmd.Flags().Set("yes", "true")
			}

			output, err := captureStdout(func() error {
				return filesDeleteCmd.RunE(cmd, []string{tt.fileID})
			})

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

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q should contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestFilesDeleteRunE_DryRun tests dry-run mode for files delete command.
func TestFilesDeleteRunE_DryRun(t *testing.T) {
	mockClient := &filesMockAPIClient{}
	cleanup, _ := setupFilesMockFactories(mockClient)
	defer cleanup()

	cmd := newFilesTestCmd()
	_ = cmd.Flags().Set("dry-run", "true")

	output, err := captureStdout(func() error {
		return filesDeleteCmd.RunE(cmd, []string{"file_dryrun"})
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "DRY-RUN") {
		t.Errorf("output should contain DRY-RUN, got: %s", output)
	}
}

// TestFilesCmdRegisteredToRoot verifies filesCmd is registered to rootCmd.
func TestFilesCmdRegisteredToRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "files" {
			found = true
			break
		}
	}
	if !found {
		t.Error("filesCmd not found in rootCmd.Commands()")
	}
}

// TestFilesCommandsHaveRunE verifies all files subcommands have RunE functions.
func TestFilesCommandsHaveRunE(t *testing.T) {
	if filesListCmd.RunE == nil {
		t.Error("filesListCmd should have RunE function")
	}
	if filesGetCmd.RunE == nil {
		t.Error("filesGetCmd should have RunE function")
	}
	if filesCreateCmd.RunE == nil {
		t.Error("filesCreateCmd should have RunE function")
	}
	if filesDeleteCmd.RunE == nil {
		t.Error("filesDeleteCmd should have RunE function")
	}
}

// TestFilesSubcommandsRegisteredToParent verifies subcommands are attached to files parent.
func TestFilesSubcommandsRegisteredToParent(t *testing.T) {
	subcommandNames := []string{"list", "get", "create", "delete"}
	registeredCmds := filesCmd.Commands()

	if len(registeredCmds) < len(subcommandNames) {
		t.Errorf("expected at least %d subcommands, got %d", len(subcommandNames), len(registeredCmds))
	}

	for _, expectedName := range subcommandNames {
		found := false
		for _, cmd := range registeredCmds {
			if len(cmd.Use) >= len(expectedName) && cmd.Use[:len(expectedName)] == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found in filesCmd.Commands()", expectedName)
		}
	}
}

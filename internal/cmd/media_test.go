package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

type mockMediaClient struct {
	api.MockClient

	createResp json.RawMessage
	createErr  error
}

func (m *mockMediaClient) CreateMediaImage(ctx context.Context, body any) (json.RawMessage, error) {
	return m.createResp, m.createErr
}

func setupMediaTest(t *testing.T, mockClient *mockMediaClient) (restore func()) {
	t.Helper()
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

func newMediaTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	addJSONBodyFlags(cmd)
	return cmd
}

func TestMediaCommandSetup(t *testing.T) {
	if mediaCmd.Use != "media" {
		t.Errorf("expected Use 'media', got %q", mediaCmd.Use)
	}
	if mediaCmd.Short != "Manage media uploads (documented endpoints)" {
		t.Errorf("expected Short 'Manage media uploads (documented endpoints)', got %q", mediaCmd.Short)
	}
}

func TestMediaSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"create-image": "Create image (documented endpoint; raw JSON body)",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range mediaCmd.Commands() {
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

func TestMediaCreateImageRunE(t *testing.T) {
	mockClient := &mockMediaClient{
		createResp: json.RawMessage(`{"id":"img_1"}`),
	}
	restore := setupMediaTest(t, mockClient)
	defer restore()

	var buf bytes.Buffer
	formatterWriter = &buf

	cmd := newMediaTestCmd()
	_ = cmd.Flags().Set("output", "json")
	_ = cmd.Flags().Set("body", `{"ok":true}`)

	if err := mediaCreateImageCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "\"img_1\"") {
		t.Fatalf("expected img_1 in output, got %q", buf.String())
	}
}

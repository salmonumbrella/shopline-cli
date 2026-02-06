package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type tokenInfoMockClient struct {
	api.MockClient
	resp json.RawMessage
	err  error
}

func (m *tokenInfoMockClient) GetTokenInfo(ctx context.Context) (json.RawMessage, error) {
	return m.resp, m.err
}

func TestTokenInfoRunEJSON(t *testing.T) {
	mockClient := &tokenInfoMockClient{resp: json.RawMessage(`{"ok":true}`)}

	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	t.Cleanup(func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	})

	buf := new(bytes.Buffer)
	formatterWriter = buf
	secretsStoreFactory = func() (CredentialStore, error) { return defaultMockStore(), nil }
	clientFactory = func(handle, accessToken string) api.APIClient { return mockClient }

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := tokenInfoCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); !bytes.Contains([]byte(got), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

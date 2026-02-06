package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type merchantsMetafieldsMockClient struct {
	api.MockClient

	listResp json.RawMessage
	listErr  error
}

func (m *merchantsMetafieldsMockClient) ListMerchantMetafields(ctx context.Context) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func setupMerchantsMetafieldsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
	t.Helper()
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter

	buf = new(bytes.Buffer)
	formatterWriter = buf

	secretsStoreFactory = func() (CredentialStore, error) {
		if store.err != nil {
			return nil, store.err
		}
		return store, nil
	}
	clientFactory = func(handle, accessToken string) api.APIClient {
		return mockClient
	}

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}, buf
}

func newMerchantsMetafieldsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Bool("yes", true, "")
	return cmd
}

func TestMerchantsMetafieldsCommandWired(t *testing.T) {
	names := map[string]bool{}
	for _, c := range merchantsCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"metafields", "app-metafields"} {
		if !names[want] {
			t.Fatalf("merchants subcommand %q not registered", want)
		}
	}
}

func TestMerchantsMetafieldsListRunEJSON(t *testing.T) {
	mockClient := &merchantsMetafieldsMockClient{
		listResp: json.RawMessage(`{"items":[{"id":"mf_1"}]}`),
	}
	cleanup, buf := setupMerchantsMetafieldsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newMerchantsMetafieldsTestCmd()

	if err := merchantsMetafieldsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if out := buf.String(); !bytes.Contains([]byte(out), []byte(`"id": "mf_1"`)) {
		t.Fatalf("unexpected output: %s", out)
	}
}

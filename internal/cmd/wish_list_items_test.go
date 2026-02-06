package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type wishListItemsMockClient struct {
	api.MockClient
	listResp json.RawMessage
	listErr  error
}

func (m *wishListItemsMockClient) ListWishListItems(ctx context.Context, opts *api.WishListItemsListOptions) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func TestWishListItemsCmdWired(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "wish-list-items" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("root subcommand %q not registered", "wish-list-items")
	}
}

func TestWishListItemsListRunEJSON(t *testing.T) {
	mockClient := &wishListItemsMockClient{
		listResp: json.RawMessage(`{"items":[{"id":"wli_1"}]}`),
	}

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
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := wishListItemsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if out := buf.String(); !bytes.Contains([]byte(out), []byte(`"id": "wli_1"`)) {
		t.Fatalf("unexpected output: %s", out)
	}
}

package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type storefrontOAuthAppsMockClient struct {
	api.MockClient
	listResp *api.StorefrontOAuthApplicationsListResponse
	listErr  error
}

func (m *storefrontOAuthAppsMockClient) ListStorefrontOAuthApplications(ctx context.Context, opts *api.StorefrontOAuthApplicationsListOptions) (*api.StorefrontOAuthApplicationsListResponse, error) {
	return m.listResp, m.listErr
}

func setupStorefrontOAuthAppsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
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
	clientFactory = func(handle, accessToken string) api.APIClient { return mockClient }

	return func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}, buf
}

func newStorefrontOAuthAppsTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	return cmd
}

func TestStorefrontOAuthApplicationsCmdWired(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "storefront-oauth-applications" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("root subcommand %q not registered", "storefront-oauth-applications")
	}
}

func TestStorefrontOAuthApplicationsListRunEJSON(t *testing.T) {
	mockClient := &storefrontOAuthAppsMockClient{
		listResp: &api.StorefrontOAuthApplicationsListResponse{
			Items: []api.StorefrontOAuthApplication{
				{ID: "app_1", Name: "A", ClientID: "cid", CreatedAt: time.Now()},
			},
			TotalCount: 1,
		},
	}
	cleanup, buf := setupStorefrontOAuthAppsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := newStorefrontOAuthAppsTestCmd()

	if err := storefrontOAuthApplicationsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if out := buf.String(); !bytes.Contains([]byte(out), []byte(`"id": "app_1"`)) {
		t.Fatalf("unexpected output: %s", out)
	}
}

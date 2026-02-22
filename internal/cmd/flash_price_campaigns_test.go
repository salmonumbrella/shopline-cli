package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

type flashPriceCampaignsMockClient struct {
	api.MockClient
	listResp   json.RawMessage
	listErr    error
	getResp    json.RawMessage
	getErr     error
	createResp json.RawMessage
	createErr  error
}

func (m *flashPriceCampaignsMockClient) ListFlashPriceCampaigns(ctx context.Context, opts *api.FlashPriceCampaignsListOptions) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *flashPriceCampaignsMockClient) GetFlashPriceCampaign(ctx context.Context, id string) (json.RawMessage, error) {
	return m.getResp, m.getErr
}

func (m *flashPriceCampaignsMockClient) CreateFlashPriceCampaign(ctx context.Context, body any) (json.RawMessage, error) {
	return m.createResp, m.createErr
}

func TestFlashPriceCampaignsListCmd_RunE(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &flashPriceCampaignsMockClient{
		listResp: json.RawMessage(`{"items":[]}`),
	}
	clientFactory = func(handle, token string) api.APIClient { return mockClient }

	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := flashPriceCampaignsListCmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"items"`)) {
		t.Fatalf("expected items in output, got %q", buf.String())
	}
}

func TestFlashPriceCampaignsGetCmd_RunE(t *testing.T) {
	origClientFactory := clientFactory
	origSecretsFactory := secretsStoreFactory
	origWriter := formatterWriter
	defer func() {
		clientFactory = origClientFactory
		secretsStoreFactory = origSecretsFactory
		formatterWriter = origWriter
	}()

	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"test"},
			creds: map[string]*secrets.StoreCredentials{
				"test": {Handle: "test", AccessToken: "token"},
			},
		}, nil
	}

	mockClient := &flashPriceCampaignsMockClient{
		getResp: json.RawMessage(`{"id":"fpc_1"}`),
	}
	clientFactory = func(handle, token string) api.APIClient { return mockClient }

	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := flashPriceCampaignsGetCmd.RunE(cmd, []string{"fpc_1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"fpc_1"`)) {
		t.Fatalf("expected id in output, got %q", buf.String())
	}
}

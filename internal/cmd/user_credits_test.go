package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/spf13/cobra"
)

type userCreditsMockClient struct {
	api.MockClient

	listResp json.RawMessage
	listErr  error

	bulkResp json.RawMessage
	bulkErr  error

	lastBody json.RawMessage
}

func (m *userCreditsMockClient) ListUserCredits(ctx context.Context, opts *api.UserCreditsListOptions) (json.RawMessage, error) {
	return m.listResp, m.listErr
}

func (m *userCreditsMockClient) BulkUpdateUserCredits(ctx context.Context, body any) (json.RawMessage, error) {
	if b, ok := body.(json.RawMessage); ok {
		m.lastBody = b
	}
	return m.bulkResp, m.bulkErr
}

func setupUserCreditsTest(t *testing.T, mockClient api.APIClient, store *mockStore) (cleanup func(), buf *bytes.Buffer) {
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

func TestUserCreditsCommandWired(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "user-credits" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("root subcommand %q not registered", "user-credits")
	}
}

func TestUserCreditsListRunE_Success(t *testing.T) {
	mockClient := &userCreditsMockClient{
		listResp: json.RawMessage(`{"items":[{"id":"uc_1"}]}`),
	}
	cleanup, buf := setupUserCreditsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")

	if err := userCreditsListCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains([]byte(got), []byte(`"id": "uc_1"`)) {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestUserCreditsBulkUpdateRunE_PassesBody(t *testing.T) {
	mockClient := &userCreditsMockClient{
		bulkResp: json.RawMessage(`{"ok":true}`),
	}
	cleanup, buf := setupUserCreditsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", `{"credits":[{"id":"uc_1","amount":1}]}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := userCreditsBulkUpdateCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE failed: %v", err)
	}
	if string(bytes.TrimSpace(mockClient.lastBody)) == "" {
		t.Fatalf("expected body, got empty")
	}
	if !bytes.Contains(buf.Bytes(), []byte(`"ok": true`)) {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestUserCreditsBulkUpdateRunE_Error(t *testing.T) {
	mockClient := &userCreditsMockClient{
		bulkErr: errors.New("boom"),
	}
	cleanup, _ := setupUserCreditsTest(t, mockClient, defaultMockStore())
	defer cleanup()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.Background())
	cmd.Flags().String("store", "", "")
	cmd.Flags().String("output", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.Flags().String("body", `{"credits":[]}`, "")
	cmd.Flags().String("body-file", "", "")

	if err := userCreditsBulkUpdateCmd.RunE(cmd, nil); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

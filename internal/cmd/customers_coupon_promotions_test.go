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

type customersCouponPromotionsMockClient struct {
	api.MockClient
	resp json.RawMessage
	err  error
}

func (m *customersCouponPromotionsMockClient) GetCustomerCouponPromotions(ctx context.Context, id string) (json.RawMessage, error) {
	return m.resp, m.err
}

func TestCustomersCouponPromotionsCmd_RunE(t *testing.T) {
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

	mockClient := &customersCouponPromotionsMockClient{
		resp: json.RawMessage(`{"items":[{"id":"promo_coupon_1"}]}`),
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}

	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := customersCouponPromotionsCmd.RunE(cmd, []string{"cust_123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains(buf.Bytes(), []byte("promo_coupon_1")) {
		t.Fatalf("expected output to contain promotion id, got %q", got)
	}
}

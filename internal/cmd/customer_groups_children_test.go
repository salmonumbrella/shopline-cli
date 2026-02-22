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

type customerGroupsChildrenMockClient struct {
	api.MockClient
	childrenResp json.RawMessage
	childrenErr  error

	childIDsResp *api.CustomerGroupIDsResponse
	childIDsErr  error
}

func (m *customerGroupsChildrenMockClient) GetCustomerGroupChildren(ctx context.Context, parentGroupID string) (json.RawMessage, error) {
	return m.childrenResp, m.childrenErr
}

func (m *customerGroupsChildrenMockClient) GetCustomerGroupChildCustomerIDs(ctx context.Context, parentGroupID, childGroupID string) (*api.CustomerGroupIDsResponse, error) {
	return m.childIDsResp, m.childIDsErr
}

func TestCustomerGroupsChildrenListCmd_RunE(t *testing.T) {
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

	mockClient := &customerGroupsChildrenMockClient{
		childrenResp: json.RawMessage(`{"items":[{"id":"grp_child_1"}]}`),
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

	if err := customerGroupsChildrenListCmd.RunE(cmd, []string{"grp_parent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains(buf.Bytes(), []byte("grp_child_1")) {
		t.Fatalf("expected output to contain child id, got %q", got)
	}
}

func TestCustomerGroupsChildrenCustomerIDsCmd_RunE_Text(t *testing.T) {
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

	mockClient := &customerGroupsChildrenMockClient{
		childIDsResp: &api.CustomerGroupIDsResponse{
			CustomerIDs: []string{"cust_1"},
			TotalCount:  1,
		},
	}
	clientFactory = func(handle, token string) api.APIClient {
		return mockClient
	}

	buf := new(bytes.Buffer)
	formatterWriter = buf

	cmd := &cobra.Command{}
	cmd.Flags().StringP("store", "s", "", "")
	cmd.Flags().StringP("output", "o", "text", "")
	cmd.Flags().String("color", "never", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().Bool("items-only", false, "")

	if err := customerGroupsChildrenCustomerIDsCmd.RunE(cmd, []string{"grp_parent", "grp_child"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got == "" || !bytes.Contains(buf.Bytes(), []byte("[customer:$cust_1]")) {
		t.Fatalf("expected formatted customer id in output, got %q", got)
	}
}

package cmd

import (
	"errors"
	"os"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/secrets"
)

// TestLabelsCommandSetup verifies labels command initialization
func TestLabelsCommandSetup(t *testing.T) {
	if labelsCmd.Use != "labels" {
		t.Errorf("expected Use 'labels', got %q", labelsCmd.Use)
	}
	if labelsCmd.Short != "Manage product labels" {
		t.Errorf("expected Short 'Manage product labels', got %q", labelsCmd.Short)
	}
}

// TestLabelsSubcommands verifies all subcommands are registered
func TestLabelsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":   "List labels",
		"get":    "Get label details",
		"create": "Create a label",
		"delete": "Delete a label",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range labelsCmd.Commands() {
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

// TestLabelsListFlags verifies list command flags exist with correct defaults
func TestLabelsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := labelsListCmd.Flags().Lookup(f.name)
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

// TestLabelsCreateFlags verifies create command flags exist
func TestLabelsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"name", ""},
		{"description", ""},
		{"color", ""},
		{"icon", ""},
		{"active", "true"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := labelsCreateCmd.Flags().Lookup(f.name)
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

// TestLabelsGetCmdUse verifies the get command has correct use string
func TestLabelsGetCmdUse(t *testing.T) {
	if labelsGetCmd.Use != "get <id>" {
		t.Errorf("expected Use 'get <id>', got %q", labelsGetCmd.Use)
	}
}

// TestLabelsDeleteCmdUse verifies the delete command has correct use string
func TestLabelsDeleteCmdUse(t *testing.T) {
	if labelsDeleteCmd.Use != "delete <id>" {
		t.Errorf("expected Use 'delete <id>', got %q", labelsDeleteCmd.Use)
	}
}

// TestLabelsListRunE_GetClientFails verifies error handling when getClient fails
func TestLabelsListRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := labelsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLabelsGetRunE_GetClientFails verifies error handling when getClient fails
func TestLabelsGetRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	err := labelsGetCmd.RunE(cmd, []string{"label_123"})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLabelsCreateRunE_DryRun verifies dry-run mode works
func TestLabelsCreateRunE_DryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("label-color", "", "") // Avoid conflict with root "color" flag
	cmd.Flags().String("icon", "", "")
	cmd.Flags().Bool("active", true, "")
	_ = cmd.Flags().Set("name", "test-label")
	_ = cmd.Flags().Set("dry-run", "true")

	err := labelsCreateCmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %v", err)
	}
}

// TestLabelsCreateRunE_GetClientFails verifies error handling when getClient fails
func TestLabelsCreateRunE_GetClientFails(t *testing.T) {
	origFactory := secretsStoreFactory
	defer func() { secretsStoreFactory = origFactory }()

	secretsStoreFactory = func() (CredentialStore, error) {
		return nil, errors.New("keyring error")
	}

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("label-color", "", "") // Avoid conflict with root "color" flag
	cmd.Flags().String("icon", "", "")
	cmd.Flags().Bool("active", true, "")
	_ = cmd.Flags().Set("name", "test-label")

	err := labelsCreateCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestLabelsDeleteRunE_DryRun verifies dry-run mode works
func TestLabelsDeleteRunE_DryRun(t *testing.T) {
	cmd := newTestCmdWithFlags()
	_ = cmd.Flags().Set("dry-run", "true")

	err := labelsDeleteCmd.RunE(cmd, []string{"label_123"})
	if err != nil {
		t.Errorf("Unexpected error in dry-run mode: %v", err)
	}
}

// TestLabelsDeleteRunE_NoConfirmation verifies delete requires confirmation
func TestLabelsDeleteRunE_NoConfirmation(t *testing.T) {
	cmd := newTestCmdWithFlags()

	err := labelsDeleteCmd.RunE(cmd, []string{"label_123"})
	// Should return nil but print a message (no confirmation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestLabelsListRunE_NoProfiles verifies error when no profiles are configured
func TestLabelsListRunE_NoProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{names: []string{}}, nil
	}

	cmd := newTestCmdWithFlags()
	err := labelsListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error for no profiles, got nil")
	}
}

// TestLabelsGetRunE_MultipleProfiles verifies error when multiple profiles exist without selection
func TestLabelsGetRunE_MultipleProfiles(t *testing.T) {
	origFactory := secretsStoreFactory
	origEnv := os.Getenv("SHOPLINE_STORE")
	defer func() {
		secretsStoreFactory = origFactory
		_ = os.Setenv("SHOPLINE_STORE", origEnv)
	}()

	_ = os.Unsetenv("SHOPLINE_STORE")
	secretsStoreFactory = func() (CredentialStore, error) {
		return &mockStore{
			names: []string{"store1", "store2"},
			creds: map[string]*secrets.StoreCredentials{
				"store1": {Handle: "test1", AccessToken: "token1"},
				"store2": {Handle: "test2", AccessToken: "token2"},
			},
		}, nil
	}

	cmd := newTestCmdWithFlags()
	err := labelsGetCmd.RunE(cmd, []string{"label_123"})
	if err == nil {
		t.Fatal("Expected error for multiple profiles, got nil")
	}
}

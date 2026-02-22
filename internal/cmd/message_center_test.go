package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// TestMessageCenterCommandSetup verifies message-center command initialization.
func TestMessageCenterCommandSetup(t *testing.T) {
	if messageCenterCmd.Use != "message-center" {
		t.Errorf("expected Use 'message-center', got %q", messageCenterCmd.Use)
	}
	if messageCenterCmd.Short != "Manage Shopline message center conversations (via Admin API)" {
		t.Errorf("expected Short 'Manage Shopline message center conversations (via Admin API)', got %q", messageCenterCmd.Short)
	}
	expectedAliases := []string{"mc", "messages"}
	if len(messageCenterCmd.Aliases) != len(expectedAliases) {
		t.Fatalf("expected %d aliases, got %d", len(expectedAliases), len(messageCenterCmd.Aliases))
	}
	for i, alias := range expectedAliases {
		if messageCenterCmd.Aliases[i] != alias {
			t.Errorf("expected alias[%d] %q, got %q", i, alias, messageCenterCmd.Aliases[i])
		}
	}
}

// TestMessageCenterSubcommands verifies list and send subcommands are registered.
func TestMessageCenterSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list": "List message center conversations",
		"send": "Send a shop/order message reply",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range messageCenterCmd.Commands() {
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

// TestMessageCenterListFlags verifies all 7 flags exist with correct defaults.
func TestMessageCenterListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"platform", ""},
		{"page", "1"},
		{"page-size", "24"},
		{"state", ""},
		{"archived", "false"},
		{"search-type", ""},
		{"search-query", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := messageCenterListCmd.Flags().Lookup(f.name)
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

// TestMessageCenterSendFlags verifies platform, type, content flags.
func TestMessageCenterSendFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"platform", ""},
		{"content", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := messageCenterSendCmd.Flags().Lookup(f.name)
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

// TestMessageCenterSendRequiredFlags verifies platform and content are required.
func TestMessageCenterSendRequiredFlags(t *testing.T) {
	requiredFlags := []string{"platform", "content"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := messageCenterSendCmd.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("required flag %q not found", name)
				return
			}
			annotations := flag.Annotations
			if annotations == nil {
				t.Errorf("flag %q has no annotations (expected required)", name)
				return
			}
			if _, ok := annotations[cobra.BashCompOneRequiredFlag]; !ok {
				t.Errorf("flag %q is not marked as required", name)
			}
		})
	}
}

// TestMessageCenterListRunE_NoAdminToken verifies error when no admin token is set.
func TestMessageCenterListRunE_NoAdminToken(t *testing.T) {
	t.Setenv("SHOPLINE_ADMIN_BASE_URL", "https://test.example.com")
	origToken := os.Getenv("SHOPLINE_ADMIN_TOKEN")
	origMerchant := os.Getenv("SHOPLINE_ADMIN_MERCHANT_ID")
	defer func() {
		_ = os.Setenv("SHOPLINE_ADMIN_TOKEN", origToken)
		_ = os.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", origMerchant)
	}()
	_ = os.Unsetenv("SHOPLINE_ADMIN_TOKEN")
	_ = os.Unsetenv("SHOPLINE_ADMIN_MERCHANT_ID")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("platform", "", "")
	cmd.Flags().Int("page", 1, "")
	cmd.Flags().Int("page-size", 20, "")
	cmd.Flags().String("state", "", "")
	cmd.Flags().Bool("archived", false, "")
	cmd.Flags().String("search-type", "", "")
	cmd.Flags().String("search-query", "", "")
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := messageCenterListCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "admin API token required: set --admin-token or SHOPLINE_ADMIN_TOKEN env var" {
		t.Errorf("unexpected error: %v", err)
	}
}

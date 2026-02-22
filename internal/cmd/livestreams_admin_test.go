package cmd

import (
	"os"
	"testing"
)

func TestLivestreamsCommandSetup(t *testing.T) {
	if livestreamsCmd.Use != "livestreams" {
		t.Errorf("expected Use 'livestreams', got %q", livestreamsCmd.Use)
	}
	if livestreamsCmd.Short != "Manage Shopline livestream sales (via Admin API)" {
		t.Errorf("expected Short 'Manage Shopline livestream sales (via Admin API)', got %q", livestreamsCmd.Short)
	}
	expectedAliases := []string{"live", "livestream", "streams", "lv"}
	if len(livestreamsCmd.Aliases) != len(expectedAliases) {
		t.Fatalf("expected %d aliases, got %d", len(expectedAliases), len(livestreamsCmd.Aliases))
	}
	// Check that all expected aliases are present (order-agnostic)
	for _, expected := range expectedAliases {
		found := false
		for _, actual := range livestreamsCmd.Aliases {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected alias %q not found in %v", expected, livestreamsCmd.Aliases)
		}
	}
}

func TestLivestreamsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":            "List livestreams",
		"get":             "Get livestream details",
		"create":          "Create a new livestream",
		"update":          "Update a livestream",
		"delete":          "Delete a livestream",
		"add-products":    "Add products to a livestream",
		"remove-products": "Remove products from a livestream",
		"start":           "Start a livestream",
		"end":             "End a livestream",
		"comments":        "Get comments for a livestream",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range livestreamsCmd.Commands() {
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

func TestLivestreamsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"type", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := livestreamsListCmd.Flags().Lookup(f.name)
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

func TestLivestreamsCreateFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"title", ""},
		{"owner", ""},
		{"description", ""},
		{"start-date", ""},
		{"end-date", ""},
		{"lock-inventory-time", ""},
		{"checkout-time", ""},
		{"checkout-message", ""},
		{"platform", ""},
		{"image", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := livestreamsCreateCmd.Flags().Lookup(f.name)
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

func TestLivestreamsCreateRequiredFlags(t *testing.T) {
	requiredFlags := []string{"title", "platform"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := livestreamsCreateCmd.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("flag %q not found", name)
			}
			annotations := flag.Annotations
			if annotations == nil {
				t.Errorf("flag %q has no annotations (expected required)", name)
				return
			}
			if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
				t.Errorf("flag %q is not marked as required", name)
			}
		})
	}
}

func TestLivestreamsExecuteFlags(t *testing.T) {
	// Verify update command flags
	updateFlags := []struct {
		name         string
		defaultValue string
	}{
		{"post-title", ""},
		{"post-owner", ""},
		{"post-description", ""},
		{"checkout-time", ""},
		{"lock-inventory-time", ""},
		{"archive-visible-time", ""},
	}

	for _, f := range updateFlags {
		t.Run("update-"+f.name, func(t *testing.T) {
			flag := livestreamsUpdateCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found on update command", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}

	// Verify comments command page flag
	t.Run("comments-page", func(t *testing.T) {
		flag := livestreamsCommentsCmd.Flags().Lookup("page")
		if flag == nil {
			t.Error("flag 'page' not found on comments command")
			return
		}
		if flag.DefValue != "1" {
			t.Errorf("expected default '1', got %q", flag.DefValue)
		}
	})

	// Verify add-products has body flags
	t.Run("add-products-body", func(t *testing.T) {
		flag := livestreamsAddProductsCmd.Flags().Lookup("body")
		if flag == nil {
			t.Error("flag 'body' not found on add-products command")
		}
	})
	t.Run("add-products-body-file", func(t *testing.T) {
		flag := livestreamsAddProductsCmd.Flags().Lookup("body-file")
		if flag == nil {
			t.Error("flag 'body-file' not found on add-products command")
		}
	})
}

func TestLivestreamsRemoveProductsFlags(t *testing.T) {
	flag := livestreamsRemoveProductsCmd.Flags().Lookup("product-ids")
	if flag == nil {
		t.Fatal("flag 'product-ids' not found")
	}
	if flag.DefValue != "[]" {
		t.Errorf("expected default '[]', got %q", flag.DefValue)
	}
	// Verify it's marked required
	annotations := flag.Annotations
	if annotations == nil {
		t.Error("flag 'product-ids' has no annotations (expected required)")
		return
	}
	if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Error("flag 'product-ids' is not marked as required")
	}
}

func TestLivestreamsStartFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"platform", ""},
		{"video-data", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := livestreamsStartCmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Errorf("flag %q not found", f.name)
				return
			}
			if flag.DefValue != f.defaultValue {
				t.Errorf("expected default %q, got %q", f.defaultValue, flag.DefValue)
			}
		})
	}

	// Verify platform is required
	t.Run("platform-required", func(t *testing.T) {
		flag := livestreamsStartCmd.Flags().Lookup("platform")
		if flag == nil {
			t.Fatal("flag 'platform' not found")
		}
		annotations := flag.Annotations
		if annotations == nil {
			t.Error("flag 'platform' has no annotations (expected required)")
			return
		}
		if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
			t.Error("flag 'platform' is not marked as required")
		}
	})
}

func TestLivestreamsDeleteRunE_NoAdminToken(t *testing.T) {
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
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")

	err := livestreamsDeleteCmd.RunE(cmd, []string{"stream_123"})
	if err == nil {
		t.Fatal("Expected error when no admin token, got nil")
	}
	expected := "admin API token required"
	if got := err.Error(); len(got) < len(expected) || got[:len(expected)] != expected {
		contains := false
		for i := 0; i <= len(got)-len(expected); i++ {
			if got[i:i+len(expected)] == expected {
				contains = true
				break
			}
		}
		if !contains {
			t.Errorf("expected error containing %q, got %q", expected, got)
		}
	}
}

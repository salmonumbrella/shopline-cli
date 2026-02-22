package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSocialPostsCommandSetup(t *testing.T) {
	if socialPostsCmd.Use != "social-posts" {
		t.Errorf("expected Use 'social-posts', got %q", socialPostsCmd.Use)
	}
	if socialPostsCmd.Short != "Manage social media sales events and channels (via Admin API)" {
		t.Errorf("expected Short 'Manage social media sales events and channels (via Admin API)', got %q", socialPostsCmd.Short)
	}
	// "social" is set directly; "sop" is applied by applyDesirePathAliases at root init time
	expectedAliases := []string{"social", "sop"}
	for _, expected := range expectedAliases {
		found := false
		for _, actual := range socialPostsCmd.Aliases {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected alias %q not found in %v", expected, socialPostsCmd.Aliases)
		}
	}
}

func TestSocialPostsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"channels":      "List social channels",
		"channel-posts": "List posts for a social channel",
		"categories":    "List social post categories",
		"products":      "Social posts product operations",
		"events":        "Sales event operations",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range socialPostsCmd.Commands() {
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

func TestSocialPostsEventsSubcommands(t *testing.T) {
	subcommands := map[string]string{
		"list":           "List sales events",
		"get":            "Get sales event details",
		"create":         "Create a sales event",
		"schedule":       "Schedule a sales event",
		"delete":         "Delete a sales event",
		"publish":        "Publish a sales event",
		"add-products":   "Add products to a sales event",
		"update-keys":    "Update product keywords in a sales event",
		"link-facebook":  "Link a Facebook post to a sales event",
		"link-instagram": "Link an Instagram post to a sales event",
		"link-fb-group":  "Link a Facebook Group post to a sales event",
	}

	for name, short := range subcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, sub := range socialPostsEventsCmd.Commands() {
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

func TestSocialPostsEventsListFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"page-size", "20"},
		{"type", "POST"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := socialPostsEventsListCmd.Flags().Lookup(f.name)
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

func TestSocialPostsEventsCreateFlags(t *testing.T) {
	requiredFlags := []string{"platform", "title"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := socialPostsEventsCreateCmd.Flags().Lookup(name)
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

func TestSocialPostsEventsScheduleFlags(t *testing.T) {
	requiredFlags := []string{"start-time", "end-time"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := socialPostsEventsScheduleCmd.Flags().Lookup(name)
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

func TestSocialPostsChannelPostsFlags(t *testing.T) {
	flag := socialPostsChannelPostsCmd.Flags().Lookup("channel-id")
	if flag == nil {
		t.Fatal("flag 'channel-id' not found")
	}
	annotations := flag.Annotations
	if annotations == nil {
		t.Error("flag 'channel-id' has no annotations (expected required)")
		return
	}
	if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Error("flag 'channel-id' is not marked as required")
	}
}

func TestSocialPostsProductsSearchFlags(t *testing.T) {
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"q", ""},
		{"page", "1"},
		{"page-size", "100"},
		{"search-type", ""},
		{"category-ids", "[]"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := socialPostsProductsSearchCmd.Flags().Lookup(f.name)
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

func TestSocialPostsEventsLinkFBGroupFlags(t *testing.T) {
	requiredFlags := []string{"page-id", "relation-url"}
	for _, name := range requiredFlags {
		t.Run(name, func(t *testing.T) {
			flag := socialPostsEventsLinkFBGroupCmd.Flags().Lookup(name)
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

func TestSocialPostsChannelsRunE_Success(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/social-posts/channels") {
			t.Errorf("expected path containing /social-posts/channels, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"ch1","name":"test-channel"}]`))
	}))
	defer ts.Close()

	t.Setenv("SHOPLINE_ADMIN_BASE_URL", ts.URL)
	t.Setenv("SHOPLINE_ADMIN_TOKEN", "test-token")
	t.Setenv("SHOPLINE_ADMIN_MERCHANT_ID", "test-merchant")

	cmd := newTestCmdWithFlags()
	cmd.Flags().String("admin-token", "", "")
	cmd.Flags().String("admin-merchant-id", "", "")
	cmd.Flags().Bool("items-only", false, "")
	cmd.SetContext(context.Background())
	var buf bytes.Buffer

	oldFW := formatterWriter
	formatterWriter = &buf
	defer func() { formatterWriter = oldFW }()

	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := socialPostsChannelsCmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected HTTP server to be called")
	}
	if !strings.Contains(buf.String(), "ch1") {
		t.Errorf("expected output to contain channel id, got %q", buf.String())
	}
}

func TestSocialPostsDeleteRunE_NoAdminToken(t *testing.T) {
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

	err := socialPostsEventsDeleteCmd.RunE(cmd, []string{"evt_123"})
	if err == nil {
		t.Fatal("Expected error when no admin token, got nil")
	}
	if !strings.Contains(err.Error(), "admin API token required") {
		t.Errorf("expected error containing 'admin API token required', got %q", err.Error())
	}
}

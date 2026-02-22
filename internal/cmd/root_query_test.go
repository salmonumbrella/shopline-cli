package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPreRunSetupQuery(t *testing.T) {
	newQueryCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("output", "text", "")
		cmd.Flags().String(outputModeFlagName, "text", "")
		_ = cmd.Flags().MarkHidden(outputModeFlagName)
		cmd.Flags().Bool("json", false, "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("jq", "", "")
		cmd.Flags().String("query-file", "", "")
		cmd.Flags().String("fields", "", "")
		return cmd
	}

	t.Run("fields implies json output and compiles to query", func(t *testing.T) {
		cmd := newQueryCmd()

		_ = cmd.Flags().Set("fields", "id,order_number")

		if err := preRunSetupQuery(cmd, nil); err != nil {
			t.Fatalf("preRunSetupQuery returned error: %v", err)
		}

		out, _ := cmd.Flags().GetString("output")
		if out != "json" {
			t.Fatalf("expected output=json, got %q", out)
		}
		q, _ := cmd.Flags().GetString("query")
		if q == "" || !strings.Contains(q, `"order_number"`) {
			t.Fatalf("expected compiled query, got %q", q)
		}
	})

	t.Run("jq and query conflict", func(t *testing.T) {
		cmd := newQueryCmd()
		_ = cmd.Flags().Set("output", "json")

		_ = cmd.Flags().Set("query", ".id")
		_ = cmd.Flags().Set("jq", ".order_number")

		if err := preRunSetupQuery(cmd, nil); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("fields and query conflict", func(t *testing.T) {
		cmd := newQueryCmd()
		_ = cmd.Flags().Set("output", "json")

		_ = cmd.Flags().Set("query", ".id")
		_ = cmd.Flags().Set("fields", "id")

		if err := preRunSetupQuery(cmd, nil); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("query-file implies json output and loads query", func(t *testing.T) {
		cmd := newQueryCmd()
		path := filepath.Join(t.TempDir(), "filter.jq")
		if err := os.WriteFile(path, []byte(".items[] | .id"), 0o600); err != nil {
			t.Fatalf("write query file: %v", err)
		}

		_ = cmd.Flags().Set("query-file", path)
		if err := preRunSetupQuery(cmd, nil); err != nil {
			t.Fatalf("preRunSetupQuery returned error: %v", err)
		}

		out, _ := cmd.Flags().GetString("output")
		if out != "json" {
			t.Fatalf("expected output=json, got %q", out)
		}
		q, _ := cmd.Flags().GetString("query")
		if q != ".items[] | .id" {
			t.Fatalf("expected loaded query, got %q", q)
		}
	})

	t.Run("query-file from stdin", func(t *testing.T) {
		cmd := newQueryCmd()
		cmd.SetIn(strings.NewReader(".id"))
		_ = cmd.Flags().Set("query-file", "-")

		if err := preRunSetupQuery(cmd, nil); err != nil {
			t.Fatalf("preRunSetupQuery returned error: %v", err)
		}

		q, _ := cmd.Flags().GetString("query")
		if q != ".id" {
			t.Fatalf("expected loaded stdin query, got %q", q)
		}
	})

	t.Run("query-file conflicts with query", func(t *testing.T) {
		cmd := newQueryCmd()
		path := filepath.Join(t.TempDir(), "filter.jq")
		if err := os.WriteFile(path, []byte(".id"), 0o600); err != nil {
			t.Fatalf("write query file: %v", err)
		}
		_ = cmd.Flags().Set("query-file", path)
		_ = cmd.Flags().Set("query", ".name")
		if err := preRunSetupQuery(cmd, nil); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("query-file conflicts with fields", func(t *testing.T) {
		cmd := newQueryCmd()
		path := filepath.Join(t.TempDir(), "filter.jq")
		if err := os.WriteFile(path, []byte(".id"), 0o600); err != nil {
			t.Fatalf("write query file: %v", err)
		}
		_ = cmd.Flags().Set("query-file", path)
		_ = cmd.Flags().Set("fields", "id")
		if err := preRunSetupQuery(cmd, nil); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("normalizes output aliases jsonl and ndjson", func(t *testing.T) {
		for _, v := range []string{"jsonl", "ndjson"} {
			cmd := newQueryCmd()
			_ = cmd.Flags().Set("output", v)
			if err := preRunSetupQuery(cmd, nil); err != nil {
				t.Fatalf("preRunSetupQuery(%s) returned error: %v", v, err)
			}
			out, _ := cmd.Flags().GetString("output")
			if out != "json" {
				t.Fatalf("expected normalized output=json for %s, got %q", v, out)
			}
			mode, _ := cmd.Flags().GetString(outputModeFlagName)
			if mode != v {
				t.Fatalf("expected output mode=%s, got %q", v, mode)
			}
		}
	})

	t.Run("query flags forced json mode sets requested mode json", func(t *testing.T) {
		cmd := newQueryCmd()
		_ = cmd.Flags().Set("query", ".id")

		if err := preRunSetupQuery(cmd, nil); err != nil {
			t.Fatalf("preRunSetupQuery returned error: %v", err)
		}
		mode, _ := cmd.Flags().GetString(outputModeFlagName)
		if mode != "json" {
			t.Fatalf("expected output mode json, got %q", mode)
		}
	})
}

func TestPreRunApplyNonInteractive(t *testing.T) {
	t.Run("force enables yes", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("yes", false, "")
		cmd.Flags().Bool("force", false, "")
		_ = cmd.Flags().Set("force", "true")

		if err := preRunApplyNonInteractive(cmd, nil); err != nil {
			t.Fatalf("preRunApplyNonInteractive returned error: %v", err)
		}
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			t.Fatalf("expected --force to set --yes")
		}
	})

	t.Run("no force leaves yes unchanged", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("yes", false, "")
		cmd.Flags().Bool("force", false, "")

		if err := preRunApplyNonInteractive(cmd, nil); err != nil {
			t.Fatalf("preRunApplyNonInteractive returned error: %v", err)
		}
		yes, _ := cmd.Flags().GetBool("yes")
		if yes {
			t.Fatalf("expected --yes to remain false when --force is not set")
		}
	})
}

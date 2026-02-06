package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPreRunSetupQuery(t *testing.T) {
	t.Run("fields implies json output and compiles to query", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("output", "text", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("jq", "", "")
		cmd.Flags().String("fields", "", "")

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
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("jq", "", "")
		cmd.Flags().String("fields", "", "")

		_ = cmd.Flags().Set("query", ".id")
		_ = cmd.Flags().Set("jq", ".order_number")

		if err := preRunSetupQuery(cmd, nil); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("fields and query conflict", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("output", "json", "")
		cmd.Flags().String("query", "", "")
		cmd.Flags().String("jq", "", "")
		cmd.Flags().String("fields", "", "")

		_ = cmd.Flags().Set("query", ".id")
		_ = cmd.Flags().Set("fields", "id")

		if err := preRunSetupQuery(cmd, nil); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

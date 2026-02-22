package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestParseFieldsWithPresets(t *testing.T) {
	t.Run("orders minimal expands", func(t *testing.T) {
		cmd := &cobra.Command{Use: "orders"}

		fields, err := parseFieldsWithPresets(cmd, "minimal")
		if err != nil {
			t.Fatalf("parseFieldsWithPresets returned error: %v", err)
		}
		if len(fields) < 3 || fields[0] != "id" {
			t.Fatalf("unexpected fields: %v", fields)
		}
	})

	t.Run("non-preset returns literal", func(t *testing.T) {
		cmd := &cobra.Command{Use: "orders"}

		fields, err := parseFieldsWithPresets(cmd, "id,order_number")
		if err != nil {
			t.Fatalf("parseFieldsWithPresets returned error: %v", err)
		}
		if len(fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(fields))
		}
	})
}

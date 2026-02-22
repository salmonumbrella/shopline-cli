package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfirmActionForceSkipsPrompt(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("no-input", false, "")
	_ = cmd.Flags().Set("force", "true")

	if !confirmAction(cmd, "confirm? ") {
		t.Fatalf("expected confirmAction=true when --force is set")
	}
}

func TestConfirmActionNoInputDeclinesWithoutPrompt(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("no-input", false, "")
	_ = cmd.Flags().Set("no-input", "true")

	out := new(bytes.Buffer)
	cmd.SetOut(out)

	if confirmAction(cmd, "confirm? ") {
		t.Fatalf("expected confirmAction=false when --no-input is set")
	}
	if out.Len() != 0 {
		t.Fatalf("expected no prompt output when --no-input is set, got %q", out.String())
	}
}

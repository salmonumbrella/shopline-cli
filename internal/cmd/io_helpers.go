package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

// outWriter returns the writer to use for normal command output.
//
// In production, cobra's cmd.OutOrStdout() is the right target.
// In unit tests, many tests override formatterWriter without setting cmd.SetOut,
// so we keep formatterWriter as the test hook when it is not os.Stdout.
func outWriter(cmd *cobra.Command) io.Writer {
	// Prefer cobra's configured writer when available (works with cmd.SetOut and
	// also with tests that capture by temporarily swapping os.Stdout).
	if cmd != nil && cmd.OutOrStdout() != nil {
		return cmd.OutOrStdout()
	}
	if formatterWriter != os.Stdout {
		return formatterWriter
	}
	return os.Stdout
}

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
	if formatterWriter != os.Stdout {
		return formatterWriter
	}
	if cmd != nil && cmd.OutOrStdout() != nil {
		return cmd.OutOrStdout()
	}
	return os.Stdout
}

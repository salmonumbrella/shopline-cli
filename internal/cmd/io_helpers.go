package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// outWriter returns the writer to use for normal command output.
//
// When a test overrides formatterWriter (to a buffer), that takes priority.
// Otherwise we fall through to cobra's cmd.OutOrStdout(), which honours both
// cmd.SetOut and os.Stdout swaps used by some tests.
func outWriter(cmd *cobra.Command) io.Writer {
	if formatterWriter != formatterWriterDefault {
		return formatterWriter
	}
	if cmd != nil {
		return cmd.OutOrStdout()
	}
	return os.Stdout
}

// inReader returns the reader to use for command input/prompts.
func inReader(cmd *cobra.Command) io.Reader {
	if cmd != nil && cmd.InOrStdin() != nil {
		return cmd.InOrStdin()
	}
	return os.Stdin
}

// confirmAction prompts for confirmation unless --yes is set. Returns true if confirmed.
func confirmAction(cmd *cobra.Command, message string) bool {
	yes, _ := cmd.Flags().GetBool("yes")
	force, _ := cmd.Flags().GetBool("force")
	if yes || force {
		return true
	}
	noInput, _ := cmd.Flags().GetBool("no-input")
	if noInput {
		return false
	}
	_, _ = fmt.Fprint(outWriter(cmd), message)
	var confirm string
	scanConfirmation(cmd, &confirm)
	return confirm == "y" || confirm == "Y"
}

// checkDryRun checks the --dry-run flag and prints a message if set. Returns true if dry-run mode.
func checkDryRun(cmd *cobra.Command, message string) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		_, _ = fmt.Fprintln(outWriter(cmd), message)
	}
	return dryRun
}

// scanConfirmation reads one token from command input for yes/no prompts.
// It mirrors previous Scanln behavior while honoring cmd.InOrStdin().
func scanConfirmation(cmd *cobra.Command, dest *string) {
	if dest == nil {
		return
	}
	*dest = ""
	reader := bufio.NewReader(inReader(cmd))
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "warning: error reading input: %v\n", err)
			}
			*dest = ""
			return
		}
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		*dest = string(r)
		return
	}
}

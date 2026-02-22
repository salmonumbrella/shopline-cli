package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// addJSONBodyFlags adds the standard agent-friendly JSON body flags.
//
// Conventions:
// - --body '<json>' for inline JSON
// - --body - to read JSON from stdin
// - --body-file path/to/file.json
func addJSONBodyFlags(c *cobra.Command) {
	c.Flags().String("body", "", "JSON request body (use '-' to read from stdin)")
	c.Flags().String("body-file", "", "Path to JSON file for request body")
}

func readJSONBodyFlags(cmd *cobra.Command) (json.RawMessage, error) {
	body, _ := cmd.Flags().GetString("body")
	bodyFile, _ := cmd.Flags().GetString("body-file")
	if strings.TrimSpace(body) != "" && strings.TrimSpace(bodyFile) != "" {
		return nil, fmt.Errorf("only one of --body or --body-file may be set")
	}
	if strings.TrimSpace(body) == "" && strings.TrimSpace(bodyFile) == "" {
		return nil, fmt.Errorf("request body required (use --body, --body -, or --body-file)")
	}

	var b []byte
	var err error
	if strings.TrimSpace(bodyFile) != "" {
		b, err = os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read --body-file: %w", err)
		}
	} else if strings.TrimSpace(body) == "-" {
		b, err = io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
	} else {
		b = []byte(body)
	}

	// Validate JSON early so errors are local and obvious.
	var tmp any
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, fmt.Errorf("invalid JSON body: %w", err)
	}
	return json.RawMessage(b), nil
}

func readJSONBodyFlagsInto(cmd *cobra.Command, v any) error {
	raw, err := readJSONBodyFlags(cmd)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("invalid JSON body for request: %w", err)
	}
	return nil
}

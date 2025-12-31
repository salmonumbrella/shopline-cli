package cmd

import (
	"fmt"
	"os"
	"testing"
)

// TestMain runs the tests sequentially to avoid race conditions
// with global cobra command state.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// boolToString converts a bool to its string representation for flag setting.
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// floatToString converts a float64 to its string representation for flag setting.
func floatToString(f float64) string {
	return fmt.Sprintf("%g", f)
}

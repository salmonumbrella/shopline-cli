package debug

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)

	logger.Printf("test message: %s", "hello")

	output := buf.String()
	if !strings.Contains(output, "test message: hello") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected output to contain [DEBUG], got: %s", output)
	}
}

func TestNopLogger(t *testing.T) {
	logger := Nop()
	// Should not panic
	logger.Printf("this should be discarded")
}

func TestNilWriter(t *testing.T) {
	logger := New(nil)
	// Should not panic when writer is nil
	logger.Printf("this should be silently ignored")
}

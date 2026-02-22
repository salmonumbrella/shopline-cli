package batch

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadItemsJSONArray(t *testing.T) {
	input := `[{"id": "1"}, {"id": "2"}]`
	items, err := ReadItemsFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to read items: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestReadItemsNDJSON(t *testing.T) {
	input := `{"id": "1"}
{"id": "2"}
{"id": "3"}`
	items, err := ReadItemsFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to read items: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestReadItemsNDJSONWithBlankLines(t *testing.T) {
	input := `{"id": "1"}

{"id": "2"}
`
	items, err := ReadItemsFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to read items: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestReadItemsEmptyInput(t *testing.T) {
	input := ``
	_, err := ReadItemsFromReader(strings.NewReader(input))
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestWriteResults(t *testing.T) {
	results := []Result{
		{ID: "1", Index: 0, Success: true},
		{ID: "2", Index: 1, Success: false, Error: "failed"},
	}

	var buf bytes.Buffer
	err := WriteResults(&buf, results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"id":"1"`) {
		t.Error("Missing first result ID")
	}
	if !strings.Contains(output, `"success":true`) {
		t.Error("Missing success field")
	}
	if !strings.Contains(output, `"error":"failed"`) {
		t.Error("Missing error field in second result")
	}

	// Verify NDJSON format (one JSON object per line)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines of NDJSON, got %d", len(lines))
	}
}

func TestReadItems_FromFile(t *testing.T) {
	// Create temp file with JSON content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "items.json")
	content := `[{"id": "file1"}, {"id": "file2"}]`
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	items, err := ReadItems(tmpFile)
	if err != nil {
		t.Fatalf("ReadItems failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestReadItems_FileNotFound(t *testing.T) {
	_, err := ReadItems("/nonexistent/path/to/file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected 'failed to open file' error, got: %v", err)
	}
}

func TestReadItems_EmptyFilename(t *testing.T) {
	// Save original stdin and restore after test
	oldStdin := os.Stdin

	// Create a pipe to simulate stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write test data to stdin
	go func() {
		_, _ = w.Write([]byte(`[{"id": "stdin1"}]`))
		_ = w.Close()
	}()

	items, err := ReadItems("")
	os.Stdin = oldStdin

	if err != nil {
		t.Fatalf("ReadItems with empty filename failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 item from stdin, got %d", len(items))
	}
}

func TestReadItems_Dash(t *testing.T) {
	// Save original stdin and restore after test
	oldStdin := os.Stdin

	// Create a pipe to simulate stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write test data to stdin
	go func() {
		_, _ = w.Write([]byte(`[{"id": "dash1"}, {"id": "dash2"}]`))
		_ = w.Close()
	}()

	items, err := ReadItems("-")
	os.Stdin = oldStdin

	if err != nil {
		t.Fatalf("ReadItems with '-' failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items from stdin, got %d", len(items))
	}
}

func TestReadItemsFromReader_TooManyItemsJSON(t *testing.T) {
	// Build a JSON array with MaxItems + 1 items
	var items []string
	for i := 0; i <= MaxItems; i++ {
		items = append(items, `{"id": "x"}`)
	}
	input := "[" + strings.Join(items, ",") + "]"

	_, err := ReadItemsFromReader(strings.NewReader(input))
	if err == nil {
		t.Error("Expected error for too many items")
	}
	if !strings.Contains(err.Error(), "too many items") {
		t.Errorf("Expected 'too many items' error, got: %v", err)
	}
}

func TestReadItemsFromReader_TooManyItemsNDJSON(t *testing.T) {
	// Build NDJSON with MaxItems + 1 items
	var lines []string
	for i := 0; i <= MaxItems; i++ {
		lines = append(lines, `{"id": "x"}`)
	}
	input := strings.Join(lines, "\n")

	_, err := ReadItemsFromReader(strings.NewReader(input))
	if err == nil {
		t.Error("Expected error for too many items in NDJSON")
	}
	if !strings.Contains(err.Error(), "too many items") {
		t.Errorf("Expected 'too many items' error, got: %v", err)
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestWriteResults_EncoderError(t *testing.T) {
	results := []Result{
		{ID: "1", Index: 0, Success: true},
	}

	err := WriteResults(&errorWriter{}, results)
	if err == nil {
		t.Error("Expected error when writer fails")
	}
}

func TestReadItemsFromReader_ReadError(t *testing.T) {
	_, err := ReadItemsFromReader(&errorReader{})
	if err == nil {
		t.Error("Expected error when reader fails")
	}
	if !strings.Contains(err.Error(), "failed to read input") {
		t.Errorf("Expected 'failed to read input' error, got: %v", err)
	}
}

package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// failWriter is a writer that always returns an error after writing n bytes.
type failWriter struct {
	written int
	failAt  int
}

func (w *failWriter) Write(p []byte) (n int, err error) {
	if w.written >= w.failAt {
		return 0, errors.New("write failed")
	}
	w.written += len(p)
	if w.written >= w.failAt {
		return len(p), errors.New("write failed")
	}
	return len(p), nil
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		colorMode string
	}{
		{"auto mode", "auto"},
		{"never mode", "never"},
		{"always mode", "always"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := New(&buf, FormatText, tt.colorMode)
			if f == nil {
				t.Fatal("New returned nil")
			}
			if f.w != &buf {
				t.Error("Writer not set correctly")
			}
			if f.format != FormatText {
				t.Error("Format not set correctly")
			}
		})
	}
}

func TestFormatterText(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "auto")

	headers := []string{"ID", "NAME", "STATUS"}
	rows := [][]string{
		{"1", "Order A", "pending"},
		{"2", "Order B", "completed"},
	}

	f.Table(headers, rows)

	output := buf.String()
	if !strings.Contains(output, "ID") {
		t.Error("Missing header in output")
	}
	if !strings.Contains(output, "Order A") {
		t.Error("Missing row data in output")
	}
}

func TestFormatterJSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never")

	data := map[string]string{"id": "123", "name": "Test"}
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("Unexpected id: %s", result["id"])
	}
}

func TestFormatterJSONIsPretty(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never")

	data := map[string]string{"id": "123", "name": "Test"}
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\n  \"id\": \"123\"") {
		t.Fatalf("expected pretty JSON output, got %q", output)
	}
}

func TestFormatterJSONLinesModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{name: "jsonl", mode: "jsonl"},
		{name: "ndjson", mode: "ndjson"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := New(&buf, FormatJSON, "never").WithJSONMode(tt.mode)

			data := []map[string]string{
				{"id": "1"},
				{"id": "2"},
			}
			if err := f.JSON(data); err != nil {
				t.Fatalf("JSON() returned error: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			if len(lines) != 2 {
				t.Fatalf("expected 2 JSON lines, got %d: %q", len(lines), buf.String())
			}

			for i, line := range lines {
				var obj map[string]string
				if err := json.Unmarshal([]byte(line), &obj); err != nil {
					t.Fatalf("line %d is not valid JSON: %v (%q)", i, err, line)
				}
				if obj["id"] != []string{"1", "2"}[i] {
					t.Fatalf("line %d id = %q, want %q", i, obj["id"], []string{"1", "2"}[i])
				}
			}
		})
	}
}

func TestFormatterJSON_EmptyRawMessageOutputsNull(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never")

	// An empty (non-nil) RawMessage marshals to 0 bytes, which is invalid JSON.
	// We normalize it to null.
	if err := f.JSON(json.RawMessage{}); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	var got any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	if got != nil {
		t.Fatalf("expected null, got %v", got)
	}
}

func TestFormatterJSON_EmptyRawMessagePointerOutputsNull(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never")

	rm := json.RawMessage{}
	if err := f.JSON(&rm); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	var got any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	if got != nil {
		t.Fatalf("expected null, got %v", got)
	}
}

func TestFormatterJSONQueryDoesNotPanicOnTypedSlice(t *testing.T) {
	type order struct {
		ID string `json:"id"`
	}

	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never").WithQuery("length")
	if err := f.JSON([]order{{ID: "o1"}, {ID: "o2"}}); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	var got int
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestWithQuery(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never")

	result := f.WithQuery(".name")
	if result != f {
		t.Error("WithQuery should return the same formatter")
	}
	if f.query != ".name" {
		t.Errorf("Query not set correctly: got %s, want .name", f.query)
	}
}

func TestWithItemsOnly(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never")

	result := f.WithItemsOnly(true)
	if result != f {
		t.Error("WithItemsOnly should return the same formatter")
	}
	if !f.itemsOnly {
		t.Error("itemsOnly not set correctly")
	}
}

func TestFormatterJSONItemsOnlyUnwrap(t *testing.T) {
	type resp struct {
		Items      []map[string]interface{} `json:"items"`
		Pagination map[string]interface{}   `json:"pagination"`
	}

	data := &resp{
		Items: []map[string]interface{}{
			{"id": "1"},
			{"id": "2"},
		},
		Pagination: map[string]interface{}{"current_page": 1},
	}

	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never").WithItemsOnly(true)
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	var out []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(out))
	}
	if out[0]["id"] != "1" {
		t.Fatalf("Unexpected first id: %v", out[0]["id"])
	}
}

func TestFormatterJSONLinesItemsOnlyUnwrap(t *testing.T) {
	type resp struct {
		Items      []map[string]interface{} `json:"items"`
		Pagination map[string]interface{}   `json:"pagination"`
	}

	data := &resp{
		Items: []map[string]interface{}{
			{"id": "1"},
			{"id": "2"},
		},
		Pagination: map[string]interface{}{"current_page": 1},
	}

	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never").WithJSONMode("jsonl").WithItemsOnly(true)
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSON lines, got %d: %q", len(lines), buf.String())
	}
}

func TestFormatterJSONLinesWithQuery(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never").
		WithJSONMode("jsonl").
		WithQuery(".[] | {id}")

	data := []map[string]interface{}{
		{"id": "1", "name": "A"},
		{"id": "2", "name": "B"},
	}
	if err := f.JSON(data); err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 JSON lines, got %d: %q", len(lines), buf.String())
	}

	for _, line := range lines {
		if strings.Contains(line, "\n  ") {
			t.Fatalf("expected compact JSONL output, got pretty JSON line %q", line)
		}
	}
}

func TestFilteredJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		query    string
		wantErr  bool
		contains string
	}{
		{
			name:     "valid query",
			data:     map[string]interface{}{"name": "test", "id": 123},
			query:    ".name",
			wantErr:  false,
			contains: "test",
		},
		{
			name:     "array filter",
			data:     []interface{}{map[string]interface{}{"id": 1}, map[string]interface{}{"id": 2}},
			query:    ".[0].id",
			wantErr:  false,
			contains: "1",
		},
		{
			name:    "invalid query syntax",
			data:    map[string]interface{}{"name": "test"},
			query:   ".invalid[",
			wantErr: true,
		},
		{
			name:    "query runtime error",
			data:    map[string]interface{}{"name": "test"},
			query:   ".name | .foo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := New(&buf, FormatJSON, "never").WithQuery(tt.query)

			err := f.JSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.contains != "" {
				if !strings.Contains(buf.String(), tt.contains) {
					t.Errorf("Output %q does not contain %q", buf.String(), tt.contains)
				}
			}
		})
	}
}

func TestOutput(t *testing.T) {
	t.Run("JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		f := New(&buf, FormatJSON, "never")

		data := map[string]interface{}{"id": "123"}
		headers := []string{"ID"}
		rowFunc := func(item interface{}) []string {
			m := item.(map[string]interface{})
			return []string{m["id"].(string)}
		}

		if err := f.Output(data, headers, rowFunc); err != nil {
			t.Fatalf("Output() error: %v", err)
		}

		if !strings.Contains(buf.String(), "123") {
			t.Error("Expected JSON output to contain '123'")
		}
	})

	t.Run("Text format with slice", func(t *testing.T) {
		var buf bytes.Buffer
		f := New(&buf, FormatText, "never")

		data := []interface{}{
			map[string]interface{}{"id": "1", "name": "First"},
			map[string]interface{}{"id": "2", "name": "Second"},
		}
		headers := []string{"ID", "NAME"}
		rowFunc := func(item interface{}) []string {
			m := item.(map[string]interface{})
			return []string{m["id"].(string), m["name"].(string)}
		}

		if err := f.Output(data, headers, rowFunc); err != nil {
			t.Fatalf("Output() error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "First") || !strings.Contains(output, "Second") {
			t.Errorf("Expected table output, got: %s", output)
		}
	})

	t.Run("Text format with single item", func(t *testing.T) {
		var buf bytes.Buffer
		f := New(&buf, FormatText, "never")

		data := map[string]interface{}{"id": "single"}
		headers := []string{"ID"}
		rowFunc := func(item interface{}) []string {
			m := item.(map[string]interface{})
			return []string{m["id"].(string)}
		}

		if err := f.Output(data, headers, rowFunc); err != nil {
			t.Fatalf("Output() error: %v", err)
		}

		if !strings.Contains(buf.String(), "single") {
			t.Error("Expected table output to contain 'single'")
		}
	})
}

func TestSuccess(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "never")

	f.Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success message not found in output: %s", output)
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "never")

	f.Error("Something went wrong")

	output := buf.String()
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("Error message not found in output: %s", output)
	}
}

func TestWarning(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "never")

	f.Warning("Be careful")

	output := buf.String()
	if !strings.Contains(output, "Be careful") {
		t.Errorf("Warning message not found in output: %s", output)
	}
}

func TestNewContext(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "never")

	ctx := context.Background()
	newCtx := NewContext(ctx, f)

	if newCtx == ctx {
		t.Error("NewContext should return a new context")
	}
}

func TestFromContext(t *testing.T) {
	t.Run("formatter in context", func(t *testing.T) {
		var buf bytes.Buffer
		f := New(&buf, FormatText, "never")

		ctx := NewContext(context.Background(), f)
		retrieved := FromContext(ctx)

		if retrieved != f {
			t.Error("FromContext should return the stored formatter")
		}
	})

	t.Run("no formatter in context", func(t *testing.T) {
		ctx := context.Background()
		retrieved := FromContext(ctx)

		if retrieved == nil {
			t.Fatal("FromContext should return a default formatter when none in context")
		}
		if retrieved.format != FormatText {
			t.Error("Default formatter should have FormatText")
		}
	})
}

func TestTableWithSingleColumn(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "never")

	headers := []string{"NAME"}
	rows := [][]string{
		{"Item1"},
		{"Item2"},
	}

	f.Table(headers, rows)

	output := buf.String()
	if !strings.Contains(output, "NAME") {
		t.Error("Missing header")
	}
	if !strings.Contains(output, "Item1") || !strings.Contains(output, "Item2") {
		t.Error("Missing row data")
	}
}

func TestTableWithEmptyRows(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, FormatText, "never")

	headers := []string{"ID", "NAME"}
	rows := [][]string{}

	f.Table(headers, rows)

	output := buf.String()
	if !strings.Contains(output, "ID") {
		t.Error("Missing header even with empty rows")
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatText != "text" {
		t.Errorf("FormatText should be 'text', got %s", FormatText)
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON should be 'json', got %s", FormatJSON)
	}
	if JSONModeJSONL != "jsonl" {
		t.Errorf("JSONModeJSONL should be 'jsonl', got %s", JSONModeJSONL)
	}
	if JSONModeNDJSON != "ndjson" {
		t.Errorf("JSONModeNDJSON should be 'ndjson', got %s", JSONModeNDJSON)
	}
}

func TestFilteredJSONEncodeError(t *testing.T) {
	// Use a writer that fails after accepting some bytes
	// This should trigger the enc.Encode(v) error path
	fw := &failWriter{failAt: 1}
	f := New(fw, FormatJSON, "never").WithQuery(".")

	data := map[string]interface{}{"name": "test"}
	err := f.JSON(data)

	if err == nil {
		t.Error("Expected error from failing writer, got nil")
	}
}

func TestFormatterQueryAliases(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{
				"id":           "abc123",
				"order_number": "20260108054948642",
				"status":       "confirmed",
			},
		},
	}
	var buf bytes.Buffer
	f := New(&buf, FormatJSON, "never").WithQuery(".it[] | {i, on, st}")
	err := f.JSON(data)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["id"] != "abc123" {
		t.Errorf("id = %v, want abc123", result["id"])
	}
	if result["order_number"] != "20260108054948642" {
		t.Errorf("order_number = %v, want 20260108054948642", result["order_number"])
	}
	if result["status"] != "confirmed" {
		t.Errorf("status = %v, want confirmed", result["status"])
	}
}

func TestFilteredJSON_JSONLMode_MatchesWriteJSONLines(t *testing.T) {
	data := []map[string]string{{"a": "1"}, {"b": "2"}}

	var buf1 bytes.Buffer
	f1 := New(&buf1, FormatJSON, "never")
	f1 = f1.WithJSONMode("jsonl")
	if err := f1.writeJSONLines(data); err != nil {
		t.Fatal(err)
	}

	var buf2 bytes.Buffer
	f2 := New(&buf2, FormatJSON, "never")
	f2 = f2.WithJSONMode("jsonl")
	f2 = f2.WithQuery(".[]")
	if err := f2.filteredJSON(data); err != nil {
		t.Fatal(err)
	}

	if buf1.String() != buf2.String() {
		t.Errorf("JSONL output mismatch:\nwriteJSONLines: %q\nfilteredJSON:   %q", buf1.String(), buf2.String())
	}
}

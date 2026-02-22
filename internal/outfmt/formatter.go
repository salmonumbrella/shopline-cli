package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/itchyny/gojq"
	"github.com/muesli/termenv"
	"github.com/salmonumbrella/shopline-cli/internal/queryalias"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type JSONMode string

const (
	JSONModeJSON   JSONMode = "json"
	JSONModeJSONL  JSONMode = "jsonl"
	JSONModeNDJSON JSONMode = "ndjson"
)

type contextKey struct{}

// Formatter handles output formatting.
type Formatter struct {
	w      io.Writer
	format Format
	output *termenv.Output
	query  string
	// jsonMode controls JSON rendering style: pretty "json" or line-delimited "jsonl"/"ndjson".
	jsonMode JSONMode
	// itemsOnly unwraps common list responses (structs with an Items field) to emit only the items array in JSON mode.
	itemsOnly bool
	// idPrefix formats the first column as [prefix:$id] when set.
	idPrefix string
}

// New creates a new formatter.
func New(w io.Writer, format Format, colorMode string) *Formatter {
	profile := termenv.ColorProfile()
	switch colorMode {
	case "never":
		profile = termenv.Ascii
	case "always":
		profile = termenv.TrueColor
	}

	return &Formatter{
		w:        w,
		format:   format,
		output:   termenv.NewOutput(w, termenv.WithProfile(profile)),
		jsonMode: JSONModeJSON,
	}
}

// WithQuery sets a JQ query for filtering.
func (f *Formatter) WithQuery(query string) *Formatter {
	f.query = query
	return f
}

// WithJSONMode sets JSON rendering mode: "json" (pretty), "jsonl", or "ndjson" (line-delimited).
func (f *Formatter) WithJSONMode(mode string) *Formatter {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case string(JSONModeJSONL):
		f.jsonMode = JSONModeJSONL
	case string(JSONModeNDJSON):
		f.jsonMode = JSONModeNDJSON
	default:
		f.jsonMode = JSONModeJSON
	}
	return f
}

// WithItemsOnly unwraps list responses to emit only the items array for JSON output.
func (f *Formatter) WithItemsOnly(itemsOnly bool) *Formatter {
	f.itemsOnly = itemsOnly
	return f
}

// WithIDPrefix formats the first column as [prefix:$id] for tables.
func (f *Formatter) WithIDPrefix(prefix string) *Formatter {
	f.idPrefix = prefix
	return f
}

// Table outputs data as a text table.
func (f *Formatter) Table(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)

	// Print headers
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t") //nolint:errcheck
		}
		fmt.Fprint(tw, f.output.String(h).Bold()) //nolint:errcheck
	}
	fmt.Fprintln(tw) //nolint:errcheck

	// Print rows
	for _, row := range rows {
		if f.idPrefix != "" && len(row) > 0 {
			if _, _, ok := ParseID(row[0]); !ok && row[0] != "" {
				row[0] = FormatID(f.idPrefix, row[0])
			}
		}
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t") //nolint:errcheck
			}
			fmt.Fprint(tw, col) //nolint:errcheck
		}
		fmt.Fprintln(tw) //nolint:errcheck
	}

	tw.Flush() //nolint:errcheck
}

// JSON outputs data as JSON.
func (f *Formatter) JSON(data interface{}) error {
	// If an API returns an empty response body but we still attempt to print it as json.RawMessage,
	// encoding/json would emit 0 bytes (invalid JSON). Normalize empty raw messages to null.
	if rm, ok := data.(json.RawMessage); ok && len(rm) == 0 {
		data = json.RawMessage(nil)
	}
	if prm, ok := data.(*json.RawMessage); ok && prm != nil && len(*prm) == 0 {
		data = json.RawMessage(nil)
	}

	if f.itemsOnly {
		if unwrapped, ok := unwrapItemsField(data); ok {
			data = unwrapped
		}
	}
	if f.query != "" {
		return f.filteredJSON(data)
	}
	if f.jsonMode == JSONModeJSONL || f.jsonMode == JSONModeNDJSON {
		return f.writeJSONLines(data)
	}

	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *Formatter) encodeJSONLValues(enc *json.Encoder, data interface{}) error {
	for _, value := range splitJSONLinesValues(data) {
		if err := enc.Encode(value); err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) writeJSONLines(data interface{}) error {
	enc := json.NewEncoder(f.w)
	return f.encodeJSONLValues(enc, data)
}

func splitJSONLinesValues(data interface{}) []interface{} {
	if data == nil {
		return []interface{}{nil}
	}

	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return []interface{}{nil}
		}
		v = v.Elem()
	}

	if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && v.Type().Elem().Kind() != reflect.Uint8 {
		values := make([]interface{}, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			values = append(values, v.Index(i).Interface())
		}
		return values
	}

	return []interface{}{v.Interface()}
}

func unwrapItemsField(data interface{}) (interface{}, bool) {
	if data == nil {
		return nil, false
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return data, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return data, false
	}

	f := v.FieldByName("Items")
	if !f.IsValid() {
		return data, false
	}
	if f.Kind() != reflect.Slice && f.Kind() != reflect.Array {
		return data, false
	}

	return f.Interface(), true
}

func (f *Formatter) filteredJSON(data interface{}) error {
	// gojq's type system is based on dynamic JSON values (map[string]any, []any, etc).
	// Passing typed Go structs/slices can trigger panics in TypeOf().
	//
	// To keep agent workflows safe (no panics, predictable jq behavior), we normalize
	// the input to a generic JSON value first.
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for query: %w", err)
	}
	var normalized any
	if err := json.Unmarshal(b, &normalized); err != nil {
		return fmt.Errorf("failed to unmarshal data for query: %w", err)
	}

	normalizedQuery := queryalias.Normalize(f.query, queryalias.ContextQuery)
	query, err := gojq.Parse(normalizedQuery)
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}

	iter := query.Run(normalized)
	enc := json.NewEncoder(f.w)
	if f.jsonMode == JSONModeJSON {
		enc.SetIndent("", "  ")
	}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}

		if f.jsonMode == JSONModeJSONL || f.jsonMode == JSONModeNDJSON {
			if err := f.encodeJSONLValues(enc, v); err != nil {
				return err
			}
			continue
		}
		if err := enc.Encode(v); err != nil {
			return err
		}
	}

	return nil
}

// Output outputs data in the configured format.
func (f *Formatter) Output(data interface{}, headers []string, rowFunc func(interface{}) []string) error {
	if f.format == FormatJSON {
		return f.JSON(data)
	}

	// For text format, need to convert to rows
	items, ok := data.([]interface{})
	if !ok {
		// Single item
		items = []interface{}{data}
	}

	var rows [][]string
	for _, item := range items {
		rows = append(rows, rowFunc(item))
	}

	f.Table(headers, rows)
	return nil
}

// Success prints a success message.
func (f *Formatter) Success(msg string) {
	fmt.Fprintln(f.w, f.output.String("✓").Foreground(termenv.ANSIGreen), msg) //nolint:errcheck
}

// Error prints an error message.
func (f *Formatter) Error(msg string) {
	fmt.Fprintln(f.w, f.output.String("✗").Foreground(termenv.ANSIRed), msg) //nolint:errcheck
}

// Warning prints a warning message.
func (f *Formatter) Warning(msg string) {
	fmt.Fprintln(f.w, f.output.String("!").Foreground(termenv.ANSIYellow), msg) //nolint:errcheck
}

// NewContext adds the formatter to a context.
func NewContext(ctx context.Context, f *Formatter) context.Context {
	return context.WithValue(ctx, contextKey{}, f)
}

// FromContext retrieves the formatter from a context.
func FromContext(ctx context.Context) *Formatter {
	if f, ok := ctx.Value(contextKey{}).(*Formatter); ok {
		return f
	}
	return New(os.Stdout, FormatText, "auto")
}

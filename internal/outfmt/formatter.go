package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/itchyny/gojq"
	"github.com/muesli/termenv"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type contextKey struct{}

// Formatter handles output formatting.
type Formatter struct {
	w      io.Writer
	format Format
	output *termenv.Output
	query  string
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
		w:      w,
		format: format,
		output: termenv.NewOutput(w, termenv.WithProfile(profile)),
	}
}

// WithQuery sets a JQ query for filtering.
func (f *Formatter) WithQuery(query string) *Formatter {
	f.query = query
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
	if f.query != "" {
		return f.filteredJSON(data)
	}

	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *Formatter) filteredJSON(data interface{}) error {
	query, err := gojq.Parse(f.query)
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}

	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}

		enc := json.NewEncoder(f.w)
		enc.SetIndent("", "  ")
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

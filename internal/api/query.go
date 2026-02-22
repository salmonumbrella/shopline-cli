package api

import (
	"net/url"
	"strconv"
	"time"
)

// QueryBuilder helps construct URL query parameters with a fluent interface.
// Zero values are automatically skipped (empty strings, 0 ints, nil pointers).
type QueryBuilder struct {
	params url.Values
}

// NewQuery creates a new QueryBuilder.
func NewQuery() *QueryBuilder {
	return &QueryBuilder{params: url.Values{}}
}

// Int adds an integer parameter if val > 0.
func (q *QueryBuilder) Int(key string, val int) *QueryBuilder {
	if val > 0 {
		q.params.Set(key, strconv.Itoa(val))
	}
	return q
}

// String adds a string parameter if val is non-empty.
func (q *QueryBuilder) String(key, val string) *QueryBuilder {
	if val != "" {
		q.params.Set(key, val)
	}
	return q
}

// Strings adds all non-empty values as repeated query parameters.
func (q *QueryBuilder) Strings(key string, vals []string) *QueryBuilder {
	for _, v := range vals {
		if v == "" {
			continue
		}
		q.params.Add(key, v)
	}
	return q
}

// Time adds a time parameter formatted as RFC3339 if t is non-nil.
func (q *QueryBuilder) Time(key string, t *time.Time) *QueryBuilder {
	if t != nil {
		q.params.Set(key, t.Format(time.RFC3339))
	}
	return q
}

// Bool adds a boolean parameter if val is true.
// For tri-state booleans (true/false/unset), use BoolPtr instead.
func (q *QueryBuilder) Bool(key string, val bool) *QueryBuilder {
	if val {
		q.params.Set(key, "true")
	}
	return q
}

// BoolPtr adds a boolean parameter if the pointer is non-nil.
// This supports tri-state booleans where nil means "not set".
func (q *QueryBuilder) BoolPtr(key string, val *bool) *QueryBuilder {
	if val != nil {
		q.params.Set(key, strconv.FormatBool(*val))
	}
	return q
}

// Build returns the encoded query string with leading "?" if params exist.
// Returns empty string if no parameters were added.
func (q *QueryBuilder) Build() string {
	if len(q.params) == 0 {
		return ""
	}
	return "?" + q.params.Encode()
}

// Encode returns just the encoded parameters without the leading "?".
// Useful when you need to append to an existing query string.
func (q *QueryBuilder) Encode() string {
	return q.params.Encode()
}

// Len returns the number of parameters added.
func (q *QueryBuilder) Len() int {
	return len(q.params)
}

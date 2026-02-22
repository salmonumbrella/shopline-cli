package api

import (
	"testing"
	"time"
)

func TestQueryBuilder_Empty(t *testing.T) {
	q := NewQuery()
	if got := q.Build(); got != "" {
		t.Errorf("Build() = %q, want empty string", got)
	}
	if got := q.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
}

func TestQueryBuilder_Int(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		val      int
		wantLen  int
		contains string
	}{
		{"positive value", "page", 5, 1, "page=5"},
		{"zero value skipped", "page", 0, 0, ""},
		{"negative value skipped", "page", -1, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery().Int(tt.key, tt.val)
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("Len() = %d, want %d", got, tt.wantLen)
			}
			if tt.contains != "" && q.Encode() != tt.contains {
				t.Errorf("Encode() = %q, want %q", q.Encode(), tt.contains)
			}
		})
	}
}

func TestQueryBuilder_String(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		val      string
		wantLen  int
		contains string
	}{
		{"non-empty value", "status", "active", 1, "status=active"},
		{"empty value skipped", "status", "", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery().String(tt.key, tt.val)
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("Len() = %d, want %d", got, tt.wantLen)
			}
			if tt.contains != "" && q.Encode() != tt.contains {
				t.Errorf("Encode() = %q, want %q", q.Encode(), tt.contains)
			}
		})
	}
}

func TestQueryBuilder_Strings(t *testing.T) {
	q := NewQuery().Strings("categoryIds", []string{"cat_1", "", "cat_2"})

	if got := q.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
	if got := q.Encode(); got != "categoryIds=cat_1&categoryIds=cat_2" {
		t.Errorf("Encode() = %q, want %q", got, "categoryIds=cat_1&categoryIds=cat_2")
	}
}

func TestQueryBuilder_Time(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		key     string
		val     *time.Time
		wantLen int
	}{
		{"non-nil time", "created_at", &now, 1},
		{"nil time skipped", "created_at", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery().Time(tt.key, tt.val)
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("Len() = %d, want %d", got, tt.wantLen)
			}
		})
	}

	// Verify time formatting
	q := NewQuery().Time("since", &now)
	want := "since=2024-01-15T10%3A30%3A00Z" // URL-encoded RFC3339
	if got := q.Encode(); got != want {
		t.Errorf("Time encoding = %q, want %q", got, want)
	}
}

func TestQueryBuilder_Bool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		val      bool
		wantLen  int
		contains string
	}{
		{"true value", "active", true, 1, "active=true"},
		{"false value skipped", "active", false, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery().Bool(tt.key, tt.val)
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("Len() = %d, want %d", got, tt.wantLen)
			}
			if tt.contains != "" && q.Encode() != tt.contains {
				t.Errorf("Encode() = %q, want %q", q.Encode(), tt.contains)
			}
		})
	}
}

func TestQueryBuilder_BoolPtr(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		key      string
		val      *bool
		wantLen  int
		contains string
	}{
		{"true pointer", "accepts_marketing", &trueVal, 1, "accepts_marketing=true"},
		{"false pointer", "accepts_marketing", &falseVal, 1, "accepts_marketing=false"},
		{"nil pointer skipped", "accepts_marketing", nil, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery().BoolPtr(tt.key, tt.val)
			if got := q.Len(); got != tt.wantLen {
				t.Errorf("Len() = %d, want %d", got, tt.wantLen)
			}
			if tt.contains != "" && q.Encode() != tt.contains {
				t.Errorf("Encode() = %q, want %q", q.Encode(), tt.contains)
			}
		})
	}
}

func TestQueryBuilder_Chaining(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	trueVal := true

	q := NewQuery().
		Int("page", 1).
		Int("page_size", 25).
		String("status", "active").
		Time("since", &now).
		BoolPtr("featured", &trueVal)

	if got := q.Len(); got != 5 {
		t.Errorf("Len() = %d, want 5", got)
	}

	result := q.Build()
	if result[0] != '?' {
		t.Errorf("Build() should start with '?', got %q", result)
	}

	// Verify all params are present (order may vary due to map)
	encoded := q.Encode()
	expected := []string{"page=1", "page_size=25", "status=active", "featured=true"}
	for _, exp := range expected {
		found := false
		for _, part := range splitParams(encoded) {
			if part == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected param %q not found in %q", exp, encoded)
		}
	}
}

func TestQueryBuilder_SkipsZeroValues(t *testing.T) {
	q := NewQuery().
		Int("page", 0).
		Int("limit", -5).
		String("status", "").
		Time("since", nil).
		Bool("active", false).
		BoolPtr("featured", nil)

	if got := q.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0 (all zero values should be skipped)", got)
	}
	if got := q.Build(); got != "" {
		t.Errorf("Build() = %q, want empty string", got)
	}
}

func TestQueryBuilder_Build_Encode_Difference(t *testing.T) {
	q := NewQuery().Int("page", 1)

	build := q.Build()
	encode := q.Encode()

	if build != "?page=1" {
		t.Errorf("Build() = %q, want ?page=1", build)
	}
	if encode != "page=1" {
		t.Errorf("Encode() = %q, want page=1", encode)
	}
}

// splitParams splits URL-encoded params by &
func splitParams(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '&' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	return result
}

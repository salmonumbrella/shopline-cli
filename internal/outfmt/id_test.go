package outfmt

import "testing"

func TestFormatID(t *testing.T) {
	tests := []struct {
		prefix string
		id     string
		want   string
	}{
		{"order", "12345", "[order:$12345]"},
		{"product", "abc-123", "[product:$abc-123]"},
		{"customer", "", "[customer:$]"}, // handles empty gracefully
	}

	for _, tt := range tests {
		got := FormatID(tt.prefix, tt.id)
		if got != tt.want {
			t.Errorf("FormatID(%q, %q) = %q, want %q", tt.prefix, tt.id, got, tt.want)
		}
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		input      string
		wantPrefix string
		wantID     string
		wantOK     bool
	}{
		{"[order:$12345]", "order", "12345", true},
		{"[product:$abc-123]", "product", "abc-123", true},
		{"invalid", "", "", false},
		{"[malformed", "", "", false},
	}

	for _, tt := range tests {
		prefix, id, ok := ParseID(tt.input)
		if ok != tt.wantOK {
			t.Errorf("ParseID(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			continue
		}
		if ok {
			if prefix != tt.wantPrefix || id != tt.wantID {
				t.Errorf("ParseID(%q) = (%q, %q), want (%q, %q)",
					tt.input, prefix, id, tt.wantPrefix, tt.wantID)
			}
		}
	}
}

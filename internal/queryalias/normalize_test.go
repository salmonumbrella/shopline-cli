package queryalias

import "testing"

func TestConfigError(t *testing.T) {
	if err := ConfigError(); err != nil {
		t.Fatalf("expected nil config error, got %v", err)
	}
}

func TestEntriesValidity(t *testing.T) {
	values := Entries()
	if len(values) == 0 {
		t.Fatal("Entries() must not be empty")
	}

	aliasSeen := make(map[string]struct{}, len(values))
	canonicalSeen := make(map[string]struct{}, len(values))
	for _, entry := range values {
		if entry.Alias == "" || entry.Canonical == "" {
			t.Fatalf("empty alias entry: %+v", entry)
		}
		if len(entry.Alias) > 3 {
			t.Fatalf("alias %q exceeds 3 characters", entry.Alias)
		}
		if entry.Alias == entry.Canonical {
			t.Fatalf("alias %q must differ from canonical %q", entry.Alias, entry.Canonical)
		}
		if _, ok := aliasSeen[entry.Alias]; ok {
			t.Fatalf("duplicate alias %q", entry.Alias)
		}
		aliasSeen[entry.Alias] = struct{}{}
		if _, ok := canonicalSeen[entry.Canonical]; ok {
			t.Fatalf("duplicate canonical key %q", entry.Canonical)
		}
		canonicalSeen[entry.Canonical] = struct{}{}
	}
}

func TestCanonical(t *testing.T) {
	tests := []struct {
		alias string
		want  string
		ok    bool
	}{
		{alias: "st", want: "status", ok: true},
		{alias: "on", want: "order_number", ok: true},
		{alias: "ci", want: "customer_id", ok: true},
		{alias: "qty", want: "quantity", ok: true},
		{alias: "tt", want: "title_translations", ok: true},
		{alias: "ts", want: "total_spent", ok: true},
		{alias: "tn", want: "tracking_number", ok: true},
		{alias: "missing", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := Canonical(tt.alias)
		if ok != tt.ok {
			t.Fatalf("Canonical(%q) ok=%v, want %v", tt.alias, ok, tt.ok)
		}
		if got != tt.want {
			t.Fatalf("Canonical(%q)=%q, want %q", tt.alias, got, tt.want)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "single alias", in: "st", want: "status"},
		{name: "nested path", in: "sa.cy", want: "shipping_address.city"},
		{name: "order aliases", in: "on.pst", want: "order_number.payment_status"},
		{name: "product title translations", in: "tt", want: "title_translations"},
		{name: "multiple aliases", in: "ci.fn", want: "customer_id.first_name"},
		{name: "long form unchanged", in: "order_number", want: "order_number"},
		{name: "mixed case unchanged", in: "St", want: "St"},
		{name: "unknown unchanged", in: "unknown_key", want: "unknown_key"},
		{name: "empty input", in: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.in, ContextPath)
			if got != tt.want {
				t.Fatalf("Normalize(path, %q)=%q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "basic dot paths",
			in:   `.it[] | select(.st == "open") | .i`,
			want: `.items[] | select(.status == "open") | .id`,
		},
		{
			name: "nested path aliases",
			in:   `.li[0].pr`,
			want: `.line_items[0].price`,
		},
		{
			name: "order aliases with customer",
			in:   `.it[] | select(.pst == "paid") | .cn`,
			want: `.items[] | select(.payment_status == "paid") | .customer_name`,
		},
		{
			name: "address aliases",
			in:   `.sa.cy`,
			want: `.shipping_address.city`,
		},
		{
			name: "function alias select",
			in:   `.it[] | sl(.st == "active") | .i`,
			want: `.items[] | select(.status == "active") | .id`,
		},
		{
			name: "recursive descent",
			in:   `..it | .ca`,
			want: `..items | .created_at`,
		},
		{
			name: "quoted bracket key preserved",
			in:   `.it[0]["st"]`,
			want: `.items[0]["st"]`,
		},
		{
			name: "mixed case token preserved",
			in:   `.St | .IT | .st`,
			want: `.St | .IT | .status`,
		},
		{
			name: "strings and comments preserved",
			in:   ".st as $x | \"keep .st and #comment\" # .st alias here\n.it",
			want: ".status as $x | \"keep .st and #comment\" # .st alias here\n.items",
		},
		{
			name: "unknown token unchanged",
			in:   `.unknown_key | .st`,
			want: `.unknown_key | .status`,
		},
		{
			name: "quoted keys only",
			in:   `.["st"] | .["it"]`,
			want: `.["st"] | .["it"]`,
		},
		{
			name: "variables are not rewritten as function aliases",
			in:   `.it[] | $sl | .st`,
			want: `.items[] | $sl | .status`,
		},
		{
			name: "del builtin preserved as bare token",
			in:   `.data | del(.temp)`,
			want: `.data | del(.temp)`,
		},
		{
			name: "shorthand single key",
			in:   `{i}`,
			want: `{id}`,
		},
		{
			name: "shorthand multiple keys",
			in:   `{i, n}`,
			want: `{id, name}`,
		},
		{
			name: "shorthand mixed with dot path",
			in:   `{i, s: .st}`,
			want: `{id, s: .status}`,
		},
		{
			name: "shorthand in pipeline",
			in:   `.it[] | {i, st, on}`,
			want: `.items[] | {id, status, order_number}`,
		},
		{
			name: "key-value pair key not rewritten",
			in:   `{i: .st}`,
			want: `{i: .status}`,
		},
		{
			name: "key-value pair string value",
			in:   `{n: "hello"}`,
			want: `{n: "hello"}`,
		},
		{
			name: "shorthand nested braces",
			in:   `{a: {i}}`,
			want: `{a: {id}}`,
		},
		{
			name: "shorthand unknown token unchanged",
			in:   `{foo}`,
			want: `{foo}`,
		},
		{
			name: "product and currency aliases",
			in:   `.it[] | {pi, sk, pr, cu}`,
			want: `.items[] | {product_id, sku, price, currency}`,
		},
		{
			name: "title translations alias with locale key",
			in:   `.it[] | {i, tl: .tt["zh-hant"]}`,
			want: `.items[] | {id, tl: .title_translations["zh-hant"]}`,
		},
		{
			name: "pagination aliases",
			in:   `.mt | {pg, pgs, tc}`,
			want: `.meta | {page, page_size, total_count}`,
		},
		{
			name: "empty input",
			in:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.in, ContextQuery)
			if got != tt.want {
				t.Fatalf("Normalize(query, %q)=%q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeUnknownContext(t *testing.T) {
	in := `.st`
	got := Normalize(in, Context(999))
	if got != in {
		t.Fatalf("Normalize with unknown context rewrote input: got %q want %q", got, in)
	}
}

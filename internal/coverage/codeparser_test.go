package coverage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseEndpointsFromGoFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("path separator assumptions in test fixture")
	}

	dir := t.TempDir()
	src := `package api

import (
	"context"
	"fmt"
)

type Client struct{}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error { return nil }
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error { return nil }

func (c *Client) Foo(ctx context.Context, id string) error {
	var out any
	_ = c.Get(ctx, "/orders", &out)
	_ = c.Get(ctx, fmt.Sprintf("/orders/%s/items/%d", id, 1), &out)

	// Variable path patterns common in this repo.
	path := "/products"
	_ = c.Get(ctx, path, &out)
	searchPath := "/products/search" + "?" + "q=hi"
	_ = c.Get(ctx, searchPath, &out)
	return nil
}
`
	p := filepath.Join(dir, "fixture.go")
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	eps, err := ParseEndpointsFromGoFiles([]string{filepath.Join(dir, "*.go")})
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	for _, ep := range eps {
		seen[ep.Key()] = true
	}

	if !seen["GET /orders"] {
		t.Fatalf("missing GET /orders, got keys=%v", keys(seen))
	}
	if !seen["GET /orders/{}/items/{}"] {
		t.Fatalf("missing GET /orders/{}/items/{}, got keys=%v", keys(seen))
	}
	if !seen["GET /products"] {
		t.Fatalf("missing GET /products, got keys=%v", keys(seen))
	}
	if !seen["GET /products/search"] {
		t.Fatalf("missing GET /products/search, got keys=%v", keys(seen))
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

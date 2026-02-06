package coverage

import "testing"

func TestParseEndpointFromDocMarkdown(t *testing.T) {
	md := `
Powered by ReadMe

# Get Order

Ask AI

get

https://open.shopline.io/v1/orders/{id}

To get detailed information for a specific order with its ID
`
	ep, ok := ParseEndpointFromDocMarkdown(md)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if ep.Method != "GET" {
		t.Fatalf("method: got %q want %q", ep.Method, "GET")
	}
	if ep.Path != "/orders/{}" {
		t.Fatalf("path: got %q want %q", ep.Path, "/orders/{}")
	}
}

func TestParseEndpointFromDocMarkdown_Storefront(t *testing.T) {
	md := `
Add cart items

post

https://dummyHandle.shoplineapp.com/storefront-api/v1/carts/{cart_id}/items
`
	ep, ok := ParseEndpointFromDocMarkdown(md)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if ep.Method != "POST" {
		t.Fatalf("method: got %q want %q", ep.Method, "POST")
	}
	if ep.Path != "/carts/{}/items" {
		t.Fatalf("path: got %q want %q", ep.Path, "/carts/{}/items")
	}
}

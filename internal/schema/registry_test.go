package schema

import (
	"testing"
)

func TestRegistry_Resources(t *testing.T) {
	r := NewRegistry()
	r.Register(Resource{
		Name:        "orders",
		Description: "Manage customer orders",
		Commands:    []string{"list", "get", "cancel"},
	})

	resources := r.Resources()
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Name != "orders" {
		t.Errorf("expected name 'orders', got %q", resources[0].Name)
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	r.Register(Resource{Name: "orders"})

	res, ok := r.Get("orders")
	if !ok {
		t.Fatal("expected to find 'orders'")
	}
	if res.Name != "orders" {
		t.Errorf("expected 'orders', got %q", res.Name)
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("expected not to find 'nonexistent'")
	}
}

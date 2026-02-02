package schema

import "testing"

func TestGlobalRegistry(t *testing.T) {
	// Reset for test isolation
	globalRegistry = NewRegistry()

	Register(Resource{Name: "test-resource"})

	resources := All()
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	res, ok := Get("test-resource")
	if !ok {
		t.Fatal("expected to find 'test-resource'")
	}
	if res.Name != "test-resource" {
		t.Errorf("expected 'test-resource', got %q", res.Name)
	}
}

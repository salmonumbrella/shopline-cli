package schema

var globalRegistry = NewRegistry()

// Register adds a resource to the global registry.
func Register(res Resource) {
	globalRegistry.Register(res)
}

// Get retrieves a resource from the global registry.
func Get(name string) (Resource, bool) {
	return globalRegistry.Get(name)
}

// All returns all registered resources.
func All() []Resource {
	return globalRegistry.Resources()
}

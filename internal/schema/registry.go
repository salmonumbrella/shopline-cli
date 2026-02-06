package schema

// Resource describes a CLI resource (command group).
type Resource struct {
	Name        string   `json:"name"`                // e.g., "orders"
	Description string   `json:"description"`         // e.g., "Manage customer orders"
	Commands    []string `json:"commands,omitempty"`  // e.g., ["list", "get", "create", "cancel"]
	IDField     string   `json:"id_field,omitempty"`  // e.g., "id" - primary identifier field
	IDPrefix    string   `json:"id_prefix,omitempty"` // e.g., "order" - for [order:$id] formatting
}

// Registry holds all registered resources.
type Registry struct {
	resources map[string]Resource
	order     []string // preserve registration order
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		resources: make(map[string]Resource),
	}
}

// Register adds a resource to the registry.
func (r *Registry) Register(res Resource) {
	if _, exists := r.resources[res.Name]; !exists {
		r.order = append(r.order, res.Name)
	}
	r.resources[res.Name] = res
}

// Get retrieves a resource by name.
func (r *Registry) Get(name string) (Resource, bool) {
	res, ok := r.resources[name]
	return res, ok
}

// Resources returns all resources in registration order.
func (r *Registry) Resources() []Resource {
	result := make([]Resource, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.resources[name])
	}
	return result
}

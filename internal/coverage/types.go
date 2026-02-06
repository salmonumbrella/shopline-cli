package coverage

// Endpoint is a normalized representation of an HTTP endpoint.
// Paths are always relative to https://open.shopline.io/v1 (e.g. "/orders/{id}").
type Endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`

	// Optional metadata for debugging / traceability.
	DocURL string `json:"doc_url,omitempty"`
	Source string `json:"source,omitempty"`
	Title  string `json:"title,omitempty"`
}

func (e Endpoint) Key() string {
	return e.Method + " " + e.Path
}

package api

import "encoding/json"

// Pagination represents the pagination metadata from API responses.
type Pagination struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalCount  int `json:"total_count"`
	TotalPages  int `json:"total_pages"`
}

// ListResponse is a generic paginated list response type.
// All list endpoints return this structure with the appropriate item type.
type ListResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
	// Legacy fields for backward compatibility with tests
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalCount int  `json:"total_count"`
	HasMore    bool `json:"has_more"`
}

// UnmarshalJSON implements custom unmarshaling to populate legacy fields from pagination.
func (r *ListResponse[T]) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type Alias ListResponse[T]
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Populate legacy fields from pagination if they're not set
	if r.Pagination.TotalCount > 0 && r.TotalCount == 0 {
		r.TotalCount = r.Pagination.TotalCount
	}
	if r.Pagination.CurrentPage > 0 && r.Page == 0 {
		r.Page = r.Pagination.CurrentPage
	}
	if r.Pagination.PerPage > 0 && r.PageSize == 0 {
		r.PageSize = r.Pagination.PerPage
	}
	if r.Pagination.TotalPages > r.Pagination.CurrentPage {
		r.HasMore = true
	}

	return nil
}

package api

import (
	"context"
	"fmt"
	"strings"
)

// CustomFieldType represents the type of custom field.
type CustomFieldType string

const (
	CustomFieldTypeText    CustomFieldType = "text"
	CustomFieldTypeNumber  CustomFieldType = "number"
	CustomFieldTypeDate    CustomFieldType = "date"
	CustomFieldTypeBoolean CustomFieldType = "boolean"
	CustomFieldTypeSelect  CustomFieldType = "select"
	CustomFieldTypeMulti   CustomFieldType = "multi_select"
	CustomFieldTypeFile    CustomFieldType = "file"
	CustomFieldTypeURL     CustomFieldType = "url"
	CustomFieldTypeEmail   CustomFieldType = "email"
	CustomFieldTypeJSON    CustomFieldType = "json"
)

// CustomFieldOwnerType represents the owner type of custom field.
type CustomFieldOwnerType string

const (
	CustomFieldOwnerProduct  CustomFieldOwnerType = "product"
	CustomFieldOwnerVariant  CustomFieldOwnerType = "variant"
	CustomFieldOwnerCustomer CustomFieldOwnerType = "customer"
	CustomFieldOwnerOrder    CustomFieldOwnerType = "order"
	CustomFieldOwnerShop     CustomFieldOwnerType = "shop"
)

// CustomField represents a custom field definition.
// Note: The actual API response has different field names than expected.
type CustomField struct {
	ID               string                 `json:"field_id"`
	Type             CustomFieldType        `json:"type"`
	NameTranslations map[string]string      `json:"name_translations"`
	HintTranslations map[string]string      `json:"hint_translations"`
	Options          map[string]interface{} `json:"options"`
	MemberInfoReward string                 `json:"member_info_reward"`
}

// CustomFieldsListOptions contains options for listing custom fields.
type CustomFieldsListOptions struct {
	Page      int
	PageSize  int
	OwnerType CustomFieldOwnerType
	Type      CustomFieldType
}

// CustomFieldsListResponse wraps a slice of CustomFields.
// Note: The API returns an array directly, not paginated.
type CustomFieldsListResponse struct {
	Items      []CustomField
	TotalCount int
}

// CustomFieldCreateRequest contains the request body for creating a custom field.
type CustomFieldCreateRequest struct {
	Name         string               `json:"name"`
	Key          string               `json:"key"`
	Description  string               `json:"description,omitempty"`
	Type         CustomFieldType      `json:"type"`
	OwnerType    CustomFieldOwnerType `json:"owner_type"`
	Required     bool                 `json:"required,omitempty"`
	Searchable   bool                 `json:"searchable,omitempty"`
	Visible      bool                 `json:"visible,omitempty"`
	DefaultValue string               `json:"default_value,omitempty"`
	Options      []string             `json:"options,omitempty"`
	Validation   string               `json:"validation,omitempty"`
	Position     int                  `json:"position,omitempty"`
}

// CustomFieldUpdateRequest contains the request body for updating a custom field.
type CustomFieldUpdateRequest struct {
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	Required     *bool    `json:"required,omitempty"`
	Searchable   *bool    `json:"searchable,omitempty"`
	Visible      *bool    `json:"visible,omitempty"`
	DefaultValue string   `json:"default_value,omitempty"`
	Options      []string `json:"options,omitempty"`
	Validation   string   `json:"validation,omitempty"`
	Position     int      `json:"position,omitempty"`
}

// ListCustomFields retrieves a list of custom fields.
// Note: The API returns an array directly, not a paginated response.
func (c *Client) ListCustomFields(ctx context.Context, opts *CustomFieldsListOptions) (*CustomFieldsListResponse, error) {
	path := "/custom_fields"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("owner_type", string(opts.OwnerType)).
			String("type", string(opts.Type)).
			Build()
	}

	var items []CustomField
	if err := c.Get(ctx, path, &items); err != nil {
		return nil, err
	}
	return &CustomFieldsListResponse{
		Items:      items,
		TotalCount: len(items),
	}, nil
}

// GetCustomField retrieves a single custom field by ID.
func (c *Client) GetCustomField(ctx context.Context, id string) (*CustomField, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("custom field id is required")
	}
	var field CustomField
	if err := c.Get(ctx, fmt.Sprintf("/custom_fields/%s", id), &field); err != nil {
		return nil, err
	}
	return &field, nil
}

// CreateCustomField creates a new custom field.
func (c *Client) CreateCustomField(ctx context.Context, req *CustomFieldCreateRequest) (*CustomField, error) {
	var field CustomField
	if err := c.Post(ctx, "/custom_fields", req, &field); err != nil {
		return nil, err
	}
	return &field, nil
}

// UpdateCustomField updates an existing custom field.
func (c *Client) UpdateCustomField(ctx context.Context, id string, req *CustomFieldUpdateRequest) (*CustomField, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("custom field id is required")
	}
	var field CustomField
	if err := c.Put(ctx, fmt.Sprintf("/custom_fields/%s", id), req, &field); err != nil {
		return nil, err
	}
	return &field, nil
}

// DeleteCustomField deletes a custom field.
func (c *Client) DeleteCustomField(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("custom field id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/custom_fields/%s", id))
}

package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MediaType represents the type of media.
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeModel3D  MediaType = "model_3d"
	MediaTypeExternal MediaType = "external_video"
)

// Media represents a Shopline media file.
type Media struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Position    int       `json:"position"`
	MediaType   MediaType `json:"media_type"`
	Alt         string    `json:"alt"`
	Src         string    `json:"src"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	MimeType    string    `json:"mime_type"`
	FileSize    int64     `json:"file_size"`
	Duration    int       `json:"duration"`
	PreviewURL  string    `json:"preview_url"`
	ExternalURL string    `json:"external_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MediasListOptions contains options for listing medias.
type MediasListOptions struct {
	Page      int
	PageSize  int
	ProductID string
	MediaType string
}

// MediasListResponse is the paginated response for medias.
type MediasListResponse = ListResponse[Media]

// MediaCreateRequest contains the data for creating a media.
type MediaCreateRequest struct {
	ProductID   string    `json:"product_id"`
	MediaType   MediaType `json:"media_type"`
	Src         string    `json:"src,omitempty"`
	Alt         string    `json:"alt,omitempty"`
	Position    int       `json:"position,omitempty"`
	ExternalURL string    `json:"external_url,omitempty"`
}

// MediaUpdateRequest contains the data for updating a media.
type MediaUpdateRequest struct {
	Alt      string `json:"alt,omitempty"`
	Position int    `json:"position,omitempty"`
}

// ListMedias retrieves a list of medias.
func (c *Client) ListMedias(ctx context.Context, opts *MediasListOptions) (*MediasListResponse, error) {
	path := "/medias"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("product_id", opts.ProductID).
			String("media_type", opts.MediaType).
			Build()
	}

	var resp MediasListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMedia retrieves a single media by ID.
func (c *Client) GetMedia(ctx context.Context, id string) (*Media, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("media id is required")
	}
	var media Media
	if err := c.Get(ctx, fmt.Sprintf("/medias/%s", id), &media); err != nil {
		return nil, err
	}
	return &media, nil
}

// CreateMedia creates a new media.
func (c *Client) CreateMedia(ctx context.Context, req *MediaCreateRequest) (*Media, error) {
	var media Media
	if err := c.Post(ctx, "/medias", req, &media); err != nil {
		return nil, err
	}
	return &media, nil
}

// UpdateMedia updates an existing media.
func (c *Client) UpdateMedia(ctx context.Context, id string, req *MediaUpdateRequest) (*Media, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("media id is required")
	}
	var media Media
	if err := c.Put(ctx, fmt.Sprintf("/medias/%s", id), req, &media); err != nil {
		return nil, err
	}
	return &media, nil
}

// DeleteMedia deletes a media.
func (c *Client) DeleteMedia(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("media id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/medias/%s", id))
}

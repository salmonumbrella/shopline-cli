package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FileStatus represents the status of a file.
type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusReady      FileStatus = "ready"
	FileStatusFailed     FileStatus = "failed"
	FileStatusProcessing FileStatus = "processing"
)

// File represents a Shopline file.
type File struct {
	ID          string     `json:"id"`
	Filename    string     `json:"filename"`
	MimeType    string     `json:"mime_type"`
	FileSize    int64      `json:"file_size"`
	URL         string     `json:"url"`
	Alt         string     `json:"alt"`
	Status      FileStatus `json:"status"`
	ContentType string     `json:"content_type"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FilesListOptions contains options for listing files.
type FilesListOptions struct {
	Page        int
	PageSize    int
	ContentType string
	Status      string
}

// FilesListResponse is the paginated response for files.
type FilesListResponse = ListResponse[File]

// FileCreateRequest contains the data for creating a file.
type FileCreateRequest struct {
	Filename    string `json:"filename"`
	URL         string `json:"url,omitempty"`
	Alt         string `json:"alt,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// FileUpdateRequest contains the data for updating a file.
type FileUpdateRequest struct {
	Filename string `json:"filename,omitempty"`
	Alt      string `json:"alt,omitempty"`
}

// ListFiles retrieves a list of files.
func (c *Client) ListFiles(ctx context.Context, opts *FilesListOptions) (*FilesListResponse, error) {
	path := "/files"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("content_type", opts.ContentType).
			String("status", opts.Status).
			Build()
	}

	var resp FilesListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFile retrieves a single file by ID.
func (c *Client) GetFile(ctx context.Context, id string) (*File, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("file id is required")
	}
	var file File
	if err := c.Get(ctx, fmt.Sprintf("/files/%s", id), &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// CreateFile creates a new file.
func (c *Client) CreateFile(ctx context.Context, req *FileCreateRequest) (*File, error) {
	var file File
	if err := c.Post(ctx, "/files", req, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// UpdateFile updates an existing file.
func (c *Client) UpdateFile(ctx context.Context, id string, req *FileUpdateRequest) (*File, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("file id is required")
	}
	var file File
	if err := c.Put(ctx, fmt.Sprintf("/files/%s", id), req, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteFile deletes a file.
func (c *Client) DeleteFile(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("file id is required")
	}
	return c.Delete(ctx, fmt.Sprintf("/files/%s", id))
}

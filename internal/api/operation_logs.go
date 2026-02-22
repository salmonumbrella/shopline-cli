package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// OperationLogAction represents the action type of an operation log.
type OperationLogAction string

const (
	OperationLogActionCreate OperationLogAction = "create"
	OperationLogActionUpdate OperationLogAction = "update"
	OperationLogActionDelete OperationLogAction = "delete"
	OperationLogActionLogin  OperationLogAction = "login"
	OperationLogActionLogout OperationLogAction = "logout"
	OperationLogActionExport OperationLogAction = "export"
	OperationLogActionImport OperationLogAction = "import"
)

// OperationLog represents an operation audit log entry.
type OperationLog struct {
	ID           string             `json:"id"`
	Action       OperationLogAction `json:"action"`
	ResourceType string             `json:"resource_type"`
	ResourceID   string             `json:"resource_id"`
	ResourceName string             `json:"resource_name"`
	UserID       string             `json:"user_id"`
	UserEmail    string             `json:"user_email"`
	UserName     string             `json:"user_name"`
	IPAddress    string             `json:"ip_address"`
	UserAgent    string             `json:"user_agent"`
	Changes      map[string]Change  `json:"changes"`
	Metadata     map[string]string  `json:"metadata"`
	CreatedAt    time.Time          `json:"created_at"`
}

// Change represents a field change in an operation log.
type Change struct {
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
}

// OperationLogsListOptions contains options for listing operation logs.
type OperationLogsListOptions struct {
	Page         int
	PageSize     int
	Action       OperationLogAction
	ResourceType string
	ResourceID   string
	UserID       string
	StartDate    *time.Time
	EndDate      *time.Time
}

// OperationLogsListResponse is the paginated response for operation logs.
type OperationLogsListResponse = ListResponse[OperationLog]

// ListOperationLogs retrieves a list of operation logs.
func (c *Client) ListOperationLogs(ctx context.Context, opts *OperationLogsListOptions) (*OperationLogsListResponse, error) {
	path := "/operation_logs"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("action", string(opts.Action)).
			String("resource_type", opts.ResourceType).
			String("resource_id", opts.ResourceID).
			String("user_id", opts.UserID).
			Time("start_date", opts.StartDate).
			Time("end_date", opts.EndDate).
			Build()
	}

	var resp OperationLogsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetOperationLog retrieves a single operation log by ID.
func (c *Client) GetOperationLog(ctx context.Context, id string) (*OperationLog, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("operation log id is required")
	}
	var log OperationLog
	if err := c.Get(ctx, fmt.Sprintf("/operation_logs/%s", id), &log); err != nil {
		return nil, err
	}
	return &log, nil
}

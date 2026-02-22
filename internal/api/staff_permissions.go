package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetStaffPermissions retrieves permission details for a staff account.
//
// Docs: GET /staffs/{id}/permissions
func (c *Client) GetStaffPermissions(ctx context.Context, staffID string) (json.RawMessage, error) {
	if strings.TrimSpace(staffID) == "" {
		return nil, fmt.Errorf("staff id is required")
	}
	var resp json.RawMessage
	if err := c.Get(ctx, fmt.Sprintf("/staffs/%s/permissions", staffID), &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

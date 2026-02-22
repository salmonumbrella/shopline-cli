package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Agent represents a sales agent.
type Agent struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AgentsListOptions contains options for listing agents.
type AgentsListOptions struct {
	Page     int
	PageSize int
	Status   string
}

// AgentsListResponse is the paginated response for agents.
type AgentsListResponse = ListResponse[Agent]

// ListAgents retrieves a list of agents.
func (c *Client) ListAgents(ctx context.Context, opts *AgentsListOptions) (*AgentsListResponse, error) {
	path := "/agents"
	if opts != nil {
		path += NewQuery().
			Int("page", opts.Page).
			Int("page_size", opts.PageSize).
			String("status", opts.Status).
			Build()
	}

	var resp AgentsListResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAgent retrieves a single agent by ID.
func (c *Client) GetAgent(ctx context.Context, id string) (*Agent, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("agent id is required")
	}
	var agent Agent
	if err := c.Get(ctx, fmt.Sprintf("/agents/%s", id), &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

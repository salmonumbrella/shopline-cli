package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAgentsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		now := time.Now()
		resp := AgentsListResponse{
			Items: []Agent{
				{
					ID:        "agent_123",
					Name:      "John Doe",
					Email:     "john@example.com",
					Phone:     "+1234567890",
					Status:    "active",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "agent_456",
					Name:      "Jane Smith",
					Email:     "jane@example.com",
					Phone:     "+0987654321",
					Status:    "inactive",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	agents, err := client.ListAgents(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents.Items) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents.Items))
	}
	if agents.Items[0].ID != "agent_123" {
		t.Errorf("Unexpected agent ID: %s", agents.Items[0].ID)
	}
	if agents.Items[0].Name != "John Doe" {
		t.Errorf("Unexpected agent name: %s", agents.Items[0].Name)
	}
	if agents.Items[0].Email != "john@example.com" {
		t.Errorf("Unexpected agent email: %s", agents.Items[0].Email)
	}
}

func TestAgentsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", query.Get("page_size"))
		}
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}

		resp := AgentsListResponse{
			Items:      []Agent{{ID: "agent_123", Name: "John Doe", Status: "active"}},
			Page:       2,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &AgentsListOptions{
		Page:     2,
		PageSize: 10,
		Status:   "active",
	}
	agents, err := client.ListAgents(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents.Items) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents.Items))
	}
	if agents.Page != 2 {
		t.Errorf("Expected page 2, got %d", agents.Page)
	}
	if agents.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", agents.PageSize)
	}
}

func TestAgentsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/agents/agent_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		now := time.Now()
		agent := Agent{
			ID:        "agent_123",
			Name:      "John Doe",
			Email:     "john@example.com",
			Phone:     "+1234567890",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		_ = json.NewEncoder(w).Encode(agent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	agent, err := client.GetAgent(context.Background(), "agent_123")
	if err != nil {
		t.Fatalf("GetAgent failed: %v", err)
	}

	if agent.ID != "agent_123" {
		t.Errorf("Unexpected agent ID: %s", agent.ID)
	}
	if agent.Name != "John Doe" {
		t.Errorf("Unexpected agent name: %s", agent.Name)
	}
	if agent.Email != "john@example.com" {
		t.Errorf("Unexpected agent email: %s", agent.Email)
	}
	if agent.Phone != "+1234567890" {
		t.Errorf("Unexpected agent phone: %s", agent.Phone)
	}
	if agent.Status != "active" {
		t.Errorf("Unexpected agent status: %s", agent.Status)
	}
}

func TestGetAgentEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetAgent(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "agent id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAgentsListAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
		})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListAgents(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for API failure, got nil")
	}
}

func TestGetAgentAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Agent not found",
		})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.GetAgent(context.Background(), "nonexistent_agent")
	if err == nil {
		t.Error("Expected error for API failure, got nil")
	}
}

func TestAgentsListEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AgentsListResponse{
			Items:      []Agent{},
			Page:       1,
			PageSize:   20,
			TotalCount: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	agents, err := client.ListAgents(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents.Items) != 0 {
		t.Errorf("Expected 0 agents, got %d", len(agents.Items))
	}
	if agents.TotalCount != 0 {
		t.Errorf("Expected total count 0, got %d", agents.TotalCount)
	}
}

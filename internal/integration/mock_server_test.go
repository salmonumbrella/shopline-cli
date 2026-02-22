package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

func TestMockServer(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure a response
	server.On(http.MethodGet, "/products", http.StatusOK, map[string]interface{}{
		"products": []map[string]interface{}{
			{"id": "123", "title": "Test Product"},
		},
	})

	// Create API client pointing to mock server
	client := api.NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Make a request
	var result struct {
		Products []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"products"`
	}
	err := client.Get(context.Background(), "/products", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify response was parsed correctly
	if len(result.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(result.Products))
	}
	if result.Products[0].ID != "123" {
		t.Errorf("expected product ID '123', got '%s'", result.Products[0].ID)
	}
	if result.Products[0].Title != "Test Product" {
		t.Errorf("expected product title 'Test Product', got '%s'", result.Products[0].Title)
	}

	// Verify request was recorded
	requests := server.Requests()
	if len(requests) != 1 {
		t.Fatalf("expected 1 recorded request, got %d", len(requests))
	}
	if requests[0].Method != http.MethodGet {
		t.Errorf("expected GET method, got %s", requests[0].Method)
	}
	if requests[0].Path != "/products" {
		t.Errorf("expected path '/products', got '%s'", requests[0].Path)
	}
}

func TestMockServerNotFound(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Create API client pointing to mock server
	client := api.NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	// Make a request to unconfigured endpoint
	var result interface{}
	err := client.Get(context.Background(), "/unconfigured", &result)

	// Should get an error (404 from server)
	if err == nil {
		t.Fatal("expected error for unconfigured endpoint, got nil")
	}

	// Verify request was still recorded
	requests := server.Requests()
	if len(requests) != 1 {
		t.Fatalf("expected 1 recorded request, got %d", len(requests))
	}
	if requests[0].Path != "/unconfigured" {
		t.Errorf("expected path '/unconfigured', got '%s'", requests[0].Path)
	}
}

func TestMockServerReset(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure a response
	server.On(http.MethodGet, "/test", http.StatusOK, "hello")

	// Make a request using standard http client
	resp, err := http.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	// Verify request was recorded
	if server.RequestCount() != 1 {
		t.Errorf("expected 1 request, got %d", server.RequestCount())
	}

	// Reset the server
	server.Reset()

	// Verify state is cleared
	if server.RequestCount() != 0 {
		t.Errorf("expected 0 requests after reset, got %d", server.RequestCount())
	}

	// Verify responses are cleared (should get 404 now)
	resp, err = http.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after reset, got %d", resp.StatusCode)
	}
}

func TestMockServerWithHeaders(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure a response with custom headers
	server.OnWithHeaders(http.MethodGet, "/with-headers", http.StatusOK, "body", map[string]string{
		"X-Custom-Header": "custom-value",
		"Content-Type":    "text/plain",
	})

	resp, err := http.Get(server.URL + "/with-headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected X-Custom-Header 'custom-value', got '%s'", resp.Header.Get("X-Custom-Header"))
	}
	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got '%s'", resp.Header.Get("Content-Type"))
	}
}

func TestMockServerLastRequest(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// No requests yet
	if server.LastRequest() != nil {
		t.Error("expected nil LastRequest when no requests made")
	}

	server.On(http.MethodGet, "/first", http.StatusOK, nil)
	server.On(http.MethodGet, "/second", http.StatusOK, nil)

	// Make requests
	resp1, _ := http.Get(server.URL + "/first")
	if resp1 != nil {
		_ = resp1.Body.Close()
	}
	resp2, _ := http.Get(server.URL + "/second")
	if resp2 != nil {
		_ = resp2.Body.Close()
	}

	last := server.LastRequest()
	if last == nil {
		t.Fatal("expected non-nil LastRequest")
	}
	if last.Path != "/second" {
		t.Errorf("expected last request path '/second', got '%s'", last.Path)
	}
}

func TestMockServerConcurrency(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.On(http.MethodGet, "/concurrent", http.StatusOK, "ok")

	// Make concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := http.Get(server.URL + "/concurrent")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				_ = resp.Body.Close()
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all requests were recorded
	if server.RequestCount() != 10 {
		t.Errorf("expected 10 requests, got %d", server.RequestCount())
	}
}

func TestMockServerPostWithBody(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.On(http.MethodPost, "/products", http.StatusCreated, map[string]interface{}{
		"product": map[string]interface{}{
			"id":    "456",
			"title": "New Product",
		},
	})

	client := api.NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	body := map[string]interface{}{
		"product": map[string]interface{}{
			"title": "New Product",
		},
	}

	var result struct {
		Product struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"product"`
	}

	err := client.Post(context.Background(), "/products", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Product.ID != "456" {
		t.Errorf("expected product ID '456', got '%s'", result.Product.ID)
	}

	// Verify request body was recorded
	last := server.LastRequest()
	if last == nil {
		t.Fatal("expected recorded request")
	}
	if last.Body == "" {
		t.Error("expected non-empty request body")
	}
}

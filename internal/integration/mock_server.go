// Package integration provides utilities for integration testing the Shopline CLI.
package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockResponse represents a configured response for the mock server.
type MockResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// RecordedRequest represents a request that was received by the mock server.
type RecordedRequest struct {
	Method string
	Path   string
	Body   string
	Query  map[string][]string
}

// MockServer wraps httptest.Server with request recording and response configuration.
type MockServer struct {
	*httptest.Server

	mu        sync.RWMutex
	responses map[string]MockResponse
	requests  []RecordedRequest
}

// NewMockServer creates a new mock API server.
func NewMockServer() *MockServer {
	ms := &MockServer{
		responses: make(map[string]MockResponse),
		requests:  make([]RecordedRequest, 0),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the request
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close() //nolint:errcheck

		ms.mu.Lock()
		ms.requests = append(ms.requests, RecordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   string(body),
			Query:  r.URL.Query(),
		})
		ms.mu.Unlock()

		// Look up the configured response
		key := r.Method + " " + r.URL.Path
		ms.mu.RLock()
		resp, ok := ms.responses[key]
		ms.mu.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Set headers
		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}

		// Set content type if not already set
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		w.WriteHeader(resp.StatusCode)

		// Write body
		if resp.Body != nil {
			switch b := resp.Body.(type) {
			case string:
				_, _ = w.Write([]byte(b))
			case []byte:
				_, _ = w.Write(b)
			default:
				data, err := json.Marshal(resp.Body)
				if err == nil {
					_, _ = w.Write(data)
				}
			}
		}
	}))

	return ms
}

// On configures a response for a given method and path.
func (ms *MockServer) On(method, path string, statusCode int, body interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := method + " " + path
	ms.responses[key] = MockResponse{
		StatusCode: statusCode,
		Body:       body,
	}
}

// OnWithHeaders configures a response with custom headers for a given method and path.
func (ms *MockServer) OnWithHeaders(method, path string, statusCode int, body interface{}, headers map[string]string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := method + " " + path
	ms.responses[key] = MockResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    headers,
	}
}

// Requests returns all recorded requests.
func (ms *MockServer) Requests() []RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Return a copy to prevent data races
	result := make([]RecordedRequest, len(ms.requests))
	copy(result, ms.requests)
	return result
}

// LastRequest returns the most recent request, or nil if none recorded.
func (ms *MockServer) LastRequest() *RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if len(ms.requests) == 0 {
		return nil
	}
	req := ms.requests[len(ms.requests)-1]
	return &req
}

// Reset clears all recorded requests and configured responses.
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.requests = make([]RecordedRequest, 0)
	ms.responses = make(map[string]MockResponse)
}

// RequestCount returns the number of recorded requests.
func (ms *MockServer) RequestCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return len(ms.requests)
}

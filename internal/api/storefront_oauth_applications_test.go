package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStorefrontOAuthApplications(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/storefront/oauth_applications", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resp := StorefrontOAuthApplicationsListResponse{
				Items: []StorefrontOAuthApplication{
					{
						ID:        "app_1",
						Name:      "Test App",
						ClientID:  "cid_1",
						Scopes:    []string{"read"},
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				},
				TotalCount: 1,
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodPost:
			var req StorefrontOAuthApplicationCreateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			_ = json.NewEncoder(w).Encode(StorefrontOAuthApplication{
				ID:           "app_new",
				Name:         req.Name,
				ClientID:     "cid_new",
				RedirectURIs: req.RedirectURIs,
				Scopes:       req.Scopes,
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			})
		default:
			t.Fatalf("method: %s", r.Method)
		}
	})
	mux.HandleFunc("/storefront/oauth_applications/app_1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(StorefrontOAuthApplication{ID: "app_1", Name: "Test App"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("method: %s", r.Method)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	if _, err := client.ListStorefrontOAuthApplications(context.Background(), &StorefrontOAuthApplicationsListOptions{Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("ListStorefrontOAuthApplications: %v", err)
	}
	if _, err := client.GetStorefrontOAuthApplication(context.Background(), "app_1"); err != nil {
		t.Fatalf("GetStorefrontOAuthApplication: %v", err)
	}
	if _, err := client.CreateStorefrontOAuthApplication(context.Background(), &StorefrontOAuthApplicationCreateRequest{Name: "X"}); err != nil {
		t.Fatalf("CreateStorefrontOAuthApplication: %v", err)
	}
	if err := client.DeleteStorefrontOAuthApplication(context.Background(), "app_1"); err != nil {
		t.Fatalf("DeleteStorefrontOAuthApplication: %v", err)
	}
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWishListItemsEndpoints(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wish_list_items" {
			t.Fatalf("path: got %s want %s", r.URL.Path, "/wish_list_items")
		}
		switch r.Method {
		case http.MethodGet, http.MethodPost, http.MethodDelete:
		default:
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	if _, err := client.ListWishListItems(context.Background(), &WishListItemsListOptions{Page: 1, PageSize: 2}); err != nil {
		t.Fatalf("ListWishListItems: %v", err)
	}
	if _, err := client.CreateWishListItem(context.Background(), json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("CreateWishListItem: %v", err)
	}
	if _, err := client.DeleteWishListItems(context.Background(), json.RawMessage(`{"ids":["1"]}`)); err != nil {
		t.Fatalf("DeleteWishListItems: %v", err)
	}
}

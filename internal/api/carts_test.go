package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCartsEndpoints(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/carts/exchange", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: got %s want %s", r.Method, http.MethodPost)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "body": body})
	})
	mux.HandleFunc("/carts/cart_123/prepare", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: got %s want %s", r.Method, http.MethodPost)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"prepared": true})
	})
	mux.HandleFunc("/carts/cart_123/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = json.NewEncoder(w).Encode(map[string]any{"action": "add"})
		case http.MethodPatch:
			_ = json.NewEncoder(w).Encode(map[string]any{"action": "update"})
		case http.MethodDelete:
			_ = json.NewEncoder(w).Encode(map[string]any{"action": "delete"})
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})
	mux.HandleFunc("/carts/cart_123/items/metafields", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method: got %s want %s", r.Method, http.MethodGet)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": "mf_1"}}})
	})
	mux.HandleFunc("/carts/cart_123/items/metafields/bulk", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})
	mux.HandleFunc("/carts/cart_123/items/app_metafields", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method: got %s want %s", r.Method, http.MethodGet)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": "amf_1"}}})
	})
	mux.HandleFunc("/carts/cart_123/items/app_metafields/bulk", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	if _, err := client.ExchangeCart(context.Background(), json.RawMessage(`{"from":"USD","to":"CAD"}`)); err != nil {
		t.Fatalf("ExchangeCart: %v", err)
	}
	if _, err := client.PrepareCart(context.Background(), "cart_123", nil); err != nil {
		t.Fatalf("PrepareCart: %v", err)
	}
	if _, err := client.AddCartItems(context.Background(), "cart_123", json.RawMessage(`{"items":[{"id":"x"}]}`)); err != nil {
		t.Fatalf("AddCartItems: %v", err)
	}
	if _, err := client.UpdateCartItems(context.Background(), "cart_123", json.RawMessage(`{"items":[{"id":"x"}]}`)); err != nil {
		t.Fatalf("UpdateCartItems: %v", err)
	}
	if _, err := client.DeleteCartItems(context.Background(), "cart_123", json.RawMessage(`{"item_ids":["x"]}`)); err != nil {
		t.Fatalf("DeleteCartItems: %v", err)
	}
	if _, err := client.ListCartItemMetafields(context.Background(), "cart_123"); err != nil {
		t.Fatalf("ListCartItemMetafields: %v", err)
	}
	if err := client.BulkCreateCartItemMetafields(context.Background(), "cart_123", json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkCreateCartItemMetafields: %v", err)
	}
	if err := client.BulkUpdateCartItemMetafields(context.Background(), "cart_123", json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkUpdateCartItemMetafields: %v", err)
	}
	if err := client.BulkDeleteCartItemMetafields(context.Background(), "cart_123", json.RawMessage(`{"ids":[]}`)); err != nil {
		t.Fatalf("BulkDeleteCartItemMetafields: %v", err)
	}
	if _, err := client.ListCartItemAppMetafields(context.Background(), "cart_123"); err != nil {
		t.Fatalf("ListCartItemAppMetafields: %v", err)
	}
	if err := client.BulkCreateCartItemAppMetafields(context.Background(), "cart_123", json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkCreateCartItemAppMetafields: %v", err)
	}
	if err := client.BulkUpdateCartItemAppMetafields(context.Background(), "cart_123", json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkUpdateCartItemAppMetafields: %v", err)
	}
	if err := client.BulkDeleteCartItemAppMetafields(context.Background(), "cart_123", json.RawMessage(`{"ids":[]}`)); err != nil {
		t.Fatalf("BulkDeleteCartItemAppMetafields: %v", err)
	}
}

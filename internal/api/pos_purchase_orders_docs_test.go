package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPOSPurchaseOrdersDocs(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/pos/purchase_orders" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if got := r.URL.Query().Get("page"); got != "2" {
				t.Fatalf("expected page=2, got %q", got)
			}
			if got := r.URL.Query().Get("page_size"); got != "50" {
				t.Fatalf("expected page_size=50, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
		}))
		defer server.Close()

		client := NewClient("token")
		client.BaseURL = server.URL
		client.SetUseOpenAPI(false)

		raw, err := client.ListPOSPurchaseOrders(context.Background(), &POSPurchaseOrdersListOptions{Page: 2, PageSize: 50})
		if err != nil {
			t.Fatalf("ListPOSPurchaseOrders failed: %v", err)
		}
		if len(raw) == 0 {
			t.Fatalf("expected response body")
		}
	})

	t.Run("get/update/child", func(t *testing.T) {
		var lastPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lastPath = r.URL.Path
			switch r.Method {
			case http.MethodGet:
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "po_1"})
			case http.MethodPut, http.MethodPost:
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if body["ok"] != true {
					t.Fatalf("expected ok=true, got %v", body)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			default:
				t.Fatalf("unexpected method %s", r.Method)
			}
		}))
		defer server.Close()

		client := NewClient("token")
		client.BaseURL = server.URL
		client.SetUseOpenAPI(false)

		_, err := client.GetPOSPurchaseOrder(context.Background(), "po_1")
		if err != nil {
			t.Fatalf("GetPOSPurchaseOrder failed: %v", err)
		}
		if lastPath != "/pos/purchase_orders/po_1" {
			t.Fatalf("unexpected path: %s", lastPath)
		}

		_, err = client.UpdatePOSPurchaseOrder(context.Background(), "po_1", map[string]any{"ok": true})
		if err != nil {
			t.Fatalf("UpdatePOSPurchaseOrder failed: %v", err)
		}
		if lastPath != "/pos/purchase_orders/po_1" {
			t.Fatalf("unexpected path: %s", lastPath)
		}

		_, err = client.CreatePOSPurchaseOrderChild(context.Background(), "po_1", map[string]any{"ok": true})
		if err != nil {
			t.Fatalf("CreatePOSPurchaseOrderChild failed: %v", err)
		}
		if lastPath != "/pos/purchase_orders/po_1/child" {
			t.Fatalf("unexpected path: %s", lastPath)
		}
	})

	t.Run("create/bulk_delete", func(t *testing.T) {
		var lastPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lastPath = r.URL.Path
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				t.Fatalf("unexpected method %s", r.Method)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}
			if body["ok"] != true {
				t.Fatalf("expected ok=true, got %v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		}))
		defer server.Close()

		client := NewClient("token")
		client.BaseURL = server.URL
		client.SetUseOpenAPI(false)

		_, err := client.CreatePOSPurchaseOrder(context.Background(), map[string]any{"ok": true})
		if err != nil {
			t.Fatalf("CreatePOSPurchaseOrder failed: %v", err)
		}
		if lastPath != "/pos/purchase_orders" {
			t.Fatalf("unexpected path: %s", lastPath)
		}

		_, err = client.BulkDeletePOSPurchaseOrders(context.Background(), map[string]any{"ok": true})
		if err != nil {
			t.Fatalf("BulkDeletePOSPurchaseOrders failed: %v", err)
		}
		if lastPath != "/pos/purchase_orders/bulk_delete" {
			t.Fatalf("unexpected path: %s", lastPath)
		}
	})
}

func TestPOSPurchaseOrderDocsEmptyID(t *testing.T) {
	client := NewClient("token")
	client.SetUseOpenAPI(false)

	if _, err := client.GetPOSPurchaseOrder(context.Background(), " "); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.UpdatePOSPurchaseOrder(context.Background(), " ", map[string]any{}); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := client.CreatePOSPurchaseOrderChild(context.Background(), " ", map[string]any{}); err == nil {
		t.Fatalf("expected error")
	}
}

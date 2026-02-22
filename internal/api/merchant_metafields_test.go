package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMerchantMetafieldsEndpoints(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/merchants/current/metafields", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPost:
		default:
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/merchants/current/metafields/mf_1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPut, http.MethodDelete:
		default:
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "mf_1"})
	})
	mux.HandleFunc("/merchants/current/metafields/bulk", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("method: %s", r.Method)
		}
	})

	mux.HandleFunc("/merchants/current/app_metafields", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPost:
		default:
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/merchants/current/app_metafields/amf_1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPut, http.MethodDelete:
		default:
			t.Fatalf("method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "amf_1"})
	})
	mux.HandleFunc("/merchants/current/app_metafields/bulk", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("method: %s", r.Method)
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := NewClient("token")
	client.BaseURL = srv.URL

	if _, err := client.ListMerchantMetafields(context.Background()); err != nil {
		t.Fatalf("ListMerchantMetafields: %v", err)
	}
	if _, err := client.GetMerchantMetafield(context.Background(), "mf_1"); err != nil {
		t.Fatalf("GetMerchantMetafield: %v", err)
	}
	if _, err := client.CreateMerchantMetafield(context.Background(), json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("CreateMerchantMetafield: %v", err)
	}
	if _, err := client.UpdateMerchantMetafield(context.Background(), "mf_1", json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("UpdateMerchantMetafield: %v", err)
	}
	if err := client.DeleteMerchantMetafield(context.Background(), "mf_1"); err != nil {
		t.Fatalf("DeleteMerchantMetafield: %v", err)
	}
	if err := client.BulkCreateMerchantMetafields(context.Background(), json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkCreateMerchantMetafields: %v", err)
	}
	if err := client.BulkUpdateMerchantMetafields(context.Background(), json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkUpdateMerchantMetafields: %v", err)
	}
	if err := client.BulkDeleteMerchantMetafields(context.Background(), json.RawMessage(`{"ids":[]}`)); err != nil {
		t.Fatalf("BulkDeleteMerchantMetafields: %v", err)
	}

	if _, err := client.ListMerchantAppMetafields(context.Background()); err != nil {
		t.Fatalf("ListMerchantAppMetafields: %v", err)
	}
	if _, err := client.GetMerchantAppMetafield(context.Background(), "amf_1"); err != nil {
		t.Fatalf("GetMerchantAppMetafield: %v", err)
	}
	if _, err := client.CreateMerchantAppMetafield(context.Background(), json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("CreateMerchantAppMetafield: %v", err)
	}
	if _, err := client.UpdateMerchantAppMetafield(context.Background(), "amf_1", json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("UpdateMerchantAppMetafield: %v", err)
	}
	if err := client.DeleteMerchantAppMetafield(context.Background(), "amf_1"); err != nil {
		t.Fatalf("DeleteMerchantAppMetafield: %v", err)
	}
	if err := client.BulkCreateMerchantAppMetafields(context.Background(), json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkCreateMerchantAppMetafields: %v", err)
	}
	if err := client.BulkUpdateMerchantAppMetafields(context.Background(), json.RawMessage(`{"items":[]}`)); err != nil {
		t.Fatalf("BulkUpdateMerchantAppMetafields: %v", err)
	}
	if err := client.BulkDeleteMerchantAppMetafields(context.Background(), json.RawMessage(`{"ids":[]}`)); err != nil {
		t.Fatalf("BulkDeleteMerchantAppMetafields: %v", err)
	}
}

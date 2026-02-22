package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWishListsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/wish_lists" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := WishListsListResponse{
			Items: []WishList{
				{ID: "wl_123", CustomerID: "cust_1", Name: "Birthday Wishes", ItemCount: 5},
				{ID: "wl_456", CustomerID: "cust_1", Name: "Holiday List", ItemCount: 10},
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

	wishLists, err := client.ListWishLists(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListWishLists failed: %v", err)
	}

	if len(wishLists.Items) != 2 {
		t.Errorf("Expected 2 wish lists, got %d", len(wishLists.Items))
	}
	if wishLists.Items[0].ID != "wl_123" {
		t.Errorf("Unexpected wish list ID: %s", wishLists.Items[0].ID)
	}
}

func TestWishListsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("customer_id") != "cust_1" {
			t.Errorf("Expected customer_id=cust_1, got %s", r.URL.Query().Get("customer_id"))
		}
		if r.URL.Query().Get("is_public") != "true" {
			t.Errorf("Expected is_public=true, got %s", r.URL.Query().Get("is_public"))
		}

		resp := WishListsListResponse{
			Items:      []WishList{{ID: "wl_123", CustomerID: "cust_1", IsPublic: true}},
			Page:       1,
			PageSize:   20,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	isPublic := true
	opts := &WishListsListOptions{
		CustomerID: "cust_1",
		IsPublic:   &isPublic,
	}
	wishLists, err := client.ListWishLists(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListWishLists failed: %v", err)
	}

	if len(wishLists.Items) != 1 {
		t.Errorf("Expected 1 wish list, got %d", len(wishLists.Items))
	}
}

func TestWishListsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wish_lists/wl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		wishList := WishList{
			ID:         "wl_123",
			CustomerID: "cust_1",
			Name:       "Birthday Wishes",
			ItemCount:  5,
			IsPublic:   true,
			ShareURL:   "https://store.com/wishlist/abc123",
		}
		_ = json.NewEncoder(w).Encode(wishList)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	wishList, err := client.GetWishList(context.Background(), "wl_123")
	if err != nil {
		t.Fatalf("GetWishList failed: %v", err)
	}

	if wishList.ID != "wl_123" {
		t.Errorf("Unexpected wish list ID: %s", wishList.ID)
	}
	if wishList.Name != "Birthday Wishes" {
		t.Errorf("Unexpected name: %s", wishList.Name)
	}
}

func TestWishListsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/wish_lists" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req WishListCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.CustomerID != "cust_1" {
			t.Errorf("Unexpected customer ID: %s", req.CustomerID)
		}
		if req.Name != "My Wish List" {
			t.Errorf("Unexpected name: %s", req.Name)
		}

		wishList := WishList{
			ID:         "wl_new",
			CustomerID: req.CustomerID,
			Name:       req.Name,
			ItemCount:  0,
		}
		_ = json.NewEncoder(w).Encode(wishList)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WishListCreateRequest{
		CustomerID: "cust_1",
		Name:       "My Wish List",
	}
	wishList, err := client.CreateWishList(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWishList failed: %v", err)
	}

	if wishList.ID != "wl_new" {
		t.Errorf("Unexpected wish list ID: %s", wishList.ID)
	}
}

func TestWishListsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/wish_lists/wl_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteWishList(context.Background(), "wl_123")
	if err != nil {
		t.Fatalf("DeleteWishList failed: %v", err)
	}
}

func TestWishListsAddItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/wish_lists/wl_123/items" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req WishListItemCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		item := WishListItem{
			ID:        "item_new",
			ProductID: req.ProductID,
			Title:     "Test Product",
			Price:     "29.99",
		}
		_ = json.NewEncoder(w).Encode(item)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &WishListItemCreateRequest{
		ProductID: "prod_123",
	}
	item, err := client.AddWishListItem(context.Background(), "wl_123", req)
	if err != nil {
		t.Fatalf("AddWishListItem failed: %v", err)
	}

	if item.ID != "item_new" {
		t.Errorf("Unexpected item ID: %s", item.ID)
	}
}

func TestWishListsRemoveItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/wish_lists/wl_123/items/item_456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.RemoveWishListItem(context.Background(), "wl_123", "item_456")
	if err != nil {
		t.Fatalf("RemoveWishListItem failed: %v", err)
	}
}

func TestGetWishListEmptyID(t *testing.T) {
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
			_, err := client.GetWishList(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "wish list id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestDeleteWishListEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteWishList(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "wish list id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestAddWishListItemEmptyIDs(t *testing.T) {
	client := NewClient("token")

	_, err := client.AddWishListItem(context.Background(), "", &WishListItemCreateRequest{ProductID: "prod_123"})
	if err == nil {
		t.Error("Expected error for empty wish list ID, got nil")
	}
	if err != nil && err.Error() != "wish list id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestRemoveWishListItemEmptyIDs(t *testing.T) {
	client := NewClient("token")

	err := client.RemoveWishListItem(context.Background(), "", "item_123")
	if err == nil {
		t.Error("Expected error for empty wish list ID, got nil")
	}
	if err.Error() != "wish list id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	err = client.RemoveWishListItem(context.Background(), "wl_123", "")
	if err == nil {
		t.Error("Expected error for empty item ID, got nil")
	}
	if err.Error() != "item id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

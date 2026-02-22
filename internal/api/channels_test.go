package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChannelsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/channels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ChannelsListResponse{
			Items: []Channel{
				{ID: "ch_123", Name: "Online Store", Handle: "online-store", Type: "online_store", Active: true},
				{ID: "ch_456", Name: "POS", Handle: "pos", Type: "point_of_sale", Active: true},
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

	channels, err := client.ListChannels(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListChannels failed: %v", err)
	}

	if len(channels.Items) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels.Items))
	}
	if channels.Items[0].ID != "ch_123" {
		t.Errorf("Unexpected channel ID: %s", channels.Items[0].ID)
	}
}

func TestChannelsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/ch_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		channel := Channel{
			ID:           "ch_123",
			Name:         "Online Store",
			Handle:       "online-store",
			Type:         "online_store",
			Active:       true,
			ProductCount: 100,
		}
		_ = json.NewEncoder(w).Encode(channel)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	channel, err := client.GetChannel(context.Background(), "ch_123")
	if err != nil {
		t.Fatalf("GetChannel failed: %v", err)
	}

	if channel.ID != "ch_123" {
		t.Errorf("Unexpected channel ID: %s", channel.ID)
	}
	if channel.ProductCount != 100 {
		t.Errorf("Unexpected product count: %d", channel.ProductCount)
	}
}

func TestGetChannelEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.GetChannel(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "channel id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestChannelsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		channel := Channel{ID: "ch_new", Name: "Mobile App", Handle: "mobile-app", Type: "mobile"}
		_ = json.NewEncoder(w).Encode(channel)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ChannelCreateRequest{
		Name: "Mobile App",
		Type: "mobile",
	}

	channel, err := client.CreateChannel(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateChannel failed: %v", err)
	}

	if channel.ID != "ch_new" {
		t.Errorf("Unexpected channel ID: %s", channel.ID)
	}
}

func TestChannelsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		channel := Channel{ID: "ch_123", Name: "Updated Store", Active: false}
		_ = json.NewEncoder(w).Encode(channel)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	name := "Updated Store"
	active := false
	req := &ChannelUpdateRequest{
		Name:   &name,
		Active: &active,
	}

	channel, err := client.UpdateChannel(context.Background(), "ch_123", req)
	if err != nil {
		t.Fatalf("UpdateChannel failed: %v", err)
	}

	if channel.Name != "Updated Store" {
		t.Errorf("Unexpected channel name: %s", channel.Name)
	}
}

func TestChannelsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteChannel(context.Background(), "ch_123")
	if err != nil {
		t.Fatalf("DeleteChannel failed: %v", err)
	}
}

func TestChannelProductsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/ch_123/products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := ChannelProductsResponse{
			Items: []ChannelProduct{
				{ProductID: "prod_1", Published: true},
				{ProductID: "prod_2", Published: true},
			},
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	products, err := client.ListChannelProducts(context.Background(), "ch_123", 1, 20)
	if err != nil {
		t.Fatalf("ListChannelProducts failed: %v", err)
	}

	if len(products.Items) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products.Items))
	}
}

func TestPublishProductToChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/products" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &ChannelPublishProductRequest{ProductID: "prod_1"}
	err := client.PublishProductToChannel(context.Background(), "ch_123", req)
	if err != nil {
		t.Fatalf("PublishProductToChannel failed: %v", err)
	}
}

func TestUnpublishProductFromChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/channels/ch_123/products/prod_1" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.UnpublishProductFromChannel(context.Background(), "ch_123", "prod_1")
	if err != nil {
		t.Fatalf("UnpublishProductFromChannel failed: %v", err)
	}
}

func TestChannelsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("Expected page_size=50, got %s", query.Get("page_size"))
		}
		if query.Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", query.Get("active"))
		}

		resp := ChannelsListResponse{
			Items:      []Channel{{ID: "ch_123", Name: "Store", Active: true}},
			Page:       2,
			PageSize:   50,
			TotalCount: 100,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	active := true
	opts := &ChannelsListOptions{
		Page:     2,
		PageSize: 50,
		Active:   &active,
	}
	channels, err := client.ListChannels(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListChannels failed: %v", err)
	}

	if len(channels.Items) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels.Items))
	}
}

func TestUpdateChannelEmptyID(t *testing.T) {
	client := NewClient("token")

	name := "Test"
	req := &ChannelUpdateRequest{Name: &name}
	_, err := client.UpdateChannel(context.Background(), "", req)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestDeleteChannelEmptyID(t *testing.T) {
	client := NewClient("token")

	err := client.DeleteChannel(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestListChannelProductsEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.ListChannelProducts(context.Background(), "", 1, 20)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestPublishProductToChannelEmptyID(t *testing.T) {
	client := NewClient("token")

	req := &ChannelPublishProductRequest{ProductID: "prod_1"}
	err := client.PublishProductToChannel(context.Background(), "", req)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestUnpublishProductFromChannelEmptyIDs(t *testing.T) {
	client := NewClient("token")

	err := client.UnpublishProductFromChannel(context.Background(), "", "prod_1")
	if err == nil {
		t.Error("Expected error for empty channel ID, got nil")
	}

	err = client.UnpublishProductFromChannel(context.Background(), "ch_123", "")
	if err == nil {
		t.Error("Expected error for empty product ID, got nil")
	}
}

func TestChannelsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	_, err := client.ListChannels(context.Background(), nil)
	if err == nil {
		t.Error("Expected error from ListChannels")
	}

	_, err = client.GetChannel(context.Background(), "ch_123")
	if err == nil {
		t.Error("Expected error from GetChannel")
	}

	req := &ChannelCreateRequest{Name: "Test", Type: "test"}
	_, err = client.CreateChannel(context.Background(), req)
	if err == nil {
		t.Error("Expected error from CreateChannel")
	}

	name := "Updated"
	updateReq := &ChannelUpdateRequest{Name: &name}
	_, err = client.UpdateChannel(context.Background(), "ch_123", updateReq)
	if err == nil {
		t.Error("Expected error from UpdateChannel")
	}

	err = client.DeleteChannel(context.Background(), "ch_123")
	if err == nil {
		t.Error("Expected error from DeleteChannel")
	}

	_, err = client.ListChannelProducts(context.Background(), "ch_123", 1, 20)
	if err == nil {
		t.Error("Expected error from ListChannelProducts")
	}

	publishReq := &ChannelPublishProductRequest{ProductID: "prod_1"}
	err = client.PublishProductToChannel(context.Background(), "ch_123", publishReq)
	if err == nil {
		t.Error("Expected error from PublishProductToChannel")
	}

	err = client.UnpublishProductFromChannel(context.Background(), "ch_123", "prod_1")
	if err == nil {
		t.Error("Expected error from UnpublishProductFromChannel")
	}
}

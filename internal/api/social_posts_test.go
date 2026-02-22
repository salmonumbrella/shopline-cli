package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminClientGetSocialChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/social-posts/channels"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"channels":[{"id":"ch1","name":"Facebook"}]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetSocialChannels(context.Background())
	if err != nil {
		t.Fatalf("GetSocialChannels() error = %v", err)
	}
	if string(result) != `{"channels":[{"id":"ch1","name":"Facebook"}]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetChannelPosts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/social-posts/channels/posts"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("partyChannelId"); got != "ch123" {
			t.Errorf("partyChannelId = %q, want %q", got, "ch123")
		}
		if got := r.URL.Query().Get("pageSize"); got != "10" {
			t.Errorf("pageSize = %q, want %q", got, "10")
		}
		if got := r.URL.Query().Get("type"); got != "POST" {
			t.Errorf("type = %q, want %q", got, "POST")
		}
		_, _ = w.Write([]byte(`{"posts":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetChannelPosts(context.Background(), &SocialChannelPostsOptions{
		PartyChannelID: "ch123",
		PageSize:       10,
		Type:           "POST",
	})
	if err != nil {
		t.Fatalf("GetChannelPosts() error = %v", err)
	}
	if string(result) != `{"posts":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetSocialCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/social-posts/categories"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"categories":[{"id":"cat1","name":"Shoes"}]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetSocialCategories(context.Background())
	if err != nil {
		t.Fatalf("GetSocialCategories() error = %v", err)
	}
	if string(result) != `{"categories":[{"id":"cat1","name":"Shoes"}]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientSearchSocialProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/social-posts/products/search"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("query"); got != "bag" {
			t.Errorf("query = %q, want %q", got, "bag")
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Errorf("page = %q, want %q", got, "1")
		}
		if got := r.URL.Query().Get("pageSize"); got != "20" {
			t.Errorf("pageSize = %q, want %q", got, "20")
		}
		_, _ = w.Write([]byte(`{"products":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.SearchSocialProducts(context.Background(), &SocialProductSearchOptions{
		Query:    "bag",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("SearchSocialProducts() error = %v", err)
	}
	if string(result) != `{"products":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientListSalesEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("pageNum"); got != "1" {
			t.Errorf("pageNum = %q, want %q", got, "1")
		}
		if got := r.URL.Query().Get("pageSize"); got != "10" {
			t.Errorf("pageSize = %q, want %q", got, "10")
		}
		if got := r.URL.Query().Get("salesType"); got != "POST" {
			t.Errorf("salesType = %q, want %q", got, "POST")
		}
		_, _ = w.Write([]byte(`{"items":[],"total":0}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.ListSalesEvents(context.Background(), &SalesEventListOptions{
		PageNum:   1,
		PageSize:  10,
		SalesType: "POST",
	})
	if err != nil {
		t.Fatalf("ListSalesEvents() error = %v", err)
	}
	if string(result) != `{"items":[],"total":0}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetSalesEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("fieldScopes"); got != "DETAILS,PRODUCT_LIST" {
			t.Errorf("fieldScopes = %q, want %q", got, "DETAILS,PRODUCT_LIST")
		}
		_, _ = w.Write([]byte(`{"id":"evt_123","title":"Summer Sale"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetSalesEvent(context.Background(), "evt_123", "DETAILS,PRODUCT_LIST")
	if err != nil {
		t.Fatalf("GetSalesEvent() error = %v", err)
	}
	if string(result) != `{"id":"evt_123","title":"Summer Sale"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientCreateSalesEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req CreateSalesEventRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.Platform != "facebook" {
			t.Errorf("platform = %q, want %q", req.Platform, "facebook")
		}
		if req.Type != 1 {
			t.Errorf("type = %d, want %d", req.Type, 1)
		}
		if req.Title != "Flash Sale" {
			t.Errorf("title = %q, want %q", req.Title, "Flash Sale")
		}
		_, _ = w.Write([]byte(`{"id":"evt_new"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.CreateSalesEvent(context.Background(), &CreateSalesEventRequest{
		Platform: "facebook",
		Type:     1,
		Title:    "Flash Sale",
	})
	if err != nil {
		t.Fatalf("CreateSalesEvent() error = %v", err)
	}
	if string(result) != `{"id":"evt_new"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientScheduleSalesEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/schedule"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req ScheduleSalesEventRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.StartTime != 1700000000 {
			t.Errorf("start_time = %d, want %d", req.StartTime, 1700000000)
		}
		if req.EndTime != 1700086400 {
			t.Errorf("end_time = %d, want %d", req.EndTime, 1700086400)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	err := client.ScheduleSalesEvent(context.Background(), "evt_123", &ScheduleSalesEventRequest{
		StartTime: 1700000000,
		EndTime:   1700086400,
	})
	if err != nil {
		t.Fatalf("ScheduleSalesEvent() error = %v", err)
	}
}

func TestAdminClientDeleteSalesEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	err := client.DeleteSalesEvent(context.Background(), "evt_123")
	if err != nil {
		t.Fatalf("DeleteSalesEvent() error = %v", err)
	}
}

func TestAdminClientPublishSalesEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/publish"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	err := client.PublishSalesEvent(context.Background(), "evt_123")
	if err != nil {
		t.Fatalf("PublishSalesEvent() error = %v", err)
	}
}

func TestAdminClientAddSalesEventProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/products"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req AddSalesEventProductsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if len(req.SPUList) != 1 {
			t.Fatalf("spuList count = %d, want 1", len(req.SPUList))
		}
		if req.SPUList[0].SPUID != "spu_1" {
			t.Errorf("spuId = %q, want %q", req.SPUList[0].SPUID, "spu_1")
		}
		if len(req.SPUList[0].SKUList) != 1 {
			t.Fatalf("skuList count = %d, want 1", len(req.SPUList[0].SKUList))
		}
		if req.SPUList[0].SKUList[0].SKUID != "sku_1" {
			t.Errorf("skuId = %q, want %q", req.SPUList[0].SKUList[0].SKUID, "sku_1")
		}
		_, _ = w.Write([]byte(`{"added":1}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.AddSalesEventProducts(context.Background(), "evt_123", &AddSalesEventProductsRequest{
		SPUList: []SalesEventSPU{
			{
				SPUID: "spu_1",
				SKUList: []SalesEventSKU{
					{SKUID: "sku_1", KeyList: []string{"K1"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("AddSalesEventProducts() error = %v", err)
	}
	if string(result) != `{"added":1}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientUpdateSalesEventProductKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/products/keys"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req UpdateProductKeysRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if len(req.SPUList) != 1 {
			t.Fatalf("spuList count = %d, want 1", len(req.SPUList))
		}
		if req.SPUList[0].SPUID != "spu_2" {
			t.Errorf("spuId = %q, want %q", req.SPUList[0].SPUID, "spu_2")
		}
		_, _ = w.Write([]byte(`{"updated":1}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.UpdateSalesEventProductKeys(context.Background(), "evt_123", &UpdateProductKeysRequest{
		SPUList: []SalesEventSPU{
			{
				SPUID:      "spu_2",
				DefaultKey: "BUY",
				SKUList: []SalesEventSKU{
					{SKUID: "sku_2", KeyList: []string{"BUY", "+1"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateSalesEventProductKeys() error = %v", err)
	}
	if string(result) != `{"updated":1}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientLinkFacebookPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/posts/facebook"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req LinkFacebookPostRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.PageID != "page_1" {
			t.Errorf("pageId = %q, want %q", req.PageID, "page_1")
		}
		if req.PageName != "My Page" {
			t.Errorf("pageName = %q, want %q", req.PageName, "My Page")
		}
		if len(req.PostList) != 1 {
			t.Fatalf("postList count = %d, want 1", len(req.PostList))
		}
		if req.PostList[0].PostID != "post_fb1" {
			t.Errorf("postId = %q, want %q", req.PostList[0].PostID, "post_fb1")
		}
		_, _ = w.Write([]byte(`{"linked":true}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.LinkFacebookPost(context.Background(), "evt_123", &LinkFacebookPostRequest{
		PageID:   "page_1",
		PageName: "My Page",
		PostList: []SalesEventPost{
			{PostID: "post_fb1", PostTitle: "Summer Deals"},
		},
	})
	if err != nil {
		t.Fatalf("LinkFacebookPost() error = %v", err)
	}
	if string(result) != `{"linked":true}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientLinkInstagramPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/posts/instagram"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req LinkInstagramPostRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.PageID != "ig_page_1" {
			t.Errorf("pageId = %q, want %q", req.PageID, "ig_page_1")
		}
		if req.PageName != "My IG" {
			t.Errorf("pageName = %q, want %q", req.PageName, "My IG")
		}
		if len(req.PostList) != 1 {
			t.Fatalf("postList count = %d, want 1", len(req.PostList))
		}
		if req.PostList[0].PostID != "post_ig1" {
			t.Errorf("postId = %q, want %q", req.PostList[0].PostID, "post_ig1")
		}
		_, _ = w.Write([]byte(`{"linked":true}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.LinkInstagramPost(context.Background(), "evt_123", &LinkInstagramPostRequest{
		PageID:   "ig_page_1",
		PageName: "My IG",
		PostList: []SalesEventPost{
			{PostID: "post_ig1", PostTitle: "IG Sale"},
		},
	})
	if err != nil {
		t.Fatalf("LinkInstagramPost() error = %v", err)
	}
	if string(result) != `{"linked":true}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientLinkFBGroupPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/sales-events/evt_123/posts/fb-group"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req LinkFBGroupPostRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.PageID != "grp_1" {
			t.Errorf("pageId = %q, want %q", req.PageID, "grp_1")
		}
		if req.RelationURL != "https://facebook.com/groups/123/posts/456" {
			t.Errorf("relationUrl = %q, want %q", req.RelationURL, "https://facebook.com/groups/123/posts/456")
		}
		_, _ = w.Write([]byte(`{"linked":true}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.LinkFBGroupPost(context.Background(), "evt_123", &LinkFBGroupPostRequest{
		PageID:      "grp_1",
		RelationURL: "https://facebook.com/groups/123/posts/456",
	})
	if err != nil {
		t.Fatalf("LinkFBGroupPost() error = %v", err)
	}
	if string(result) != `{"linked":true}` {
		t.Errorf("result = %s", string(result))
	}
}

package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminAPIErrorTyped(t *testing.T) {
	err := adminAPIError(404, []byte(`{"error":"not found"}`))
	var adminErr *AdminAPIError
	if !errors.As(err, &adminErr) {
		t.Fatalf("expected *AdminAPIError, got %T", err)
	}
	if adminErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", adminErr.StatusCode)
	}
	if adminErr.Body != "not found" {
		t.Errorf("Body = %q, want %q", adminErr.Body, "not found")
	}
}

func TestAdminAPIErrorTyped_WithHint(t *testing.T) {
	// Errors with hints should still be unwrappable to AdminAPIError.
	err := adminAPIError(400, []byte(`{"error":"登入已超時，請重新登入"}`))
	var adminErr *AdminAPIError
	if !errors.As(err, &adminErr) {
		t.Fatalf("expected *AdminAPIError via errors.As, got %T", err)
	}
	if adminErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", adminErr.StatusCode)
	}
}

func TestIsAdminRouteUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"404", adminAPIError(404, []byte(`not found`)), true},
		{"405", adminAPIError(405, []byte(`method not allowed`)), true},
		{"400", adminAPIError(400, []byte(`bad request`)), false},
		{"generic error", errors.New("some error"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAdminRouteUnavailable(tc.err); got != tc.want {
				t.Errorf("isAdminRouteUnavailable() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAdminClientMerchantPath(t *testing.T) {
	client := newTestAdminClient("https://admin.example.com/api/public")

	got := client.merchantPath("/orders/abc/comments")
	want := "https://admin.example.com/api/public/merchants/test-merchant/orders/abc/comments"
	if got != want {
		t.Errorf("merchantPath() = %q, want %q", got, want)
	}
}

func TestAdminClientDo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("Content-Type = %q, want empty (no body)", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	var result json.RawMessage
	err := client.do(context.Background(), "GET", server.URL+"/test", nil, &result)
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}

	if string(result) != `{"status":"ok"}` {
		t.Errorf("result = %s, want %s", string(result), `{"status":"ok"}`)
	}
}

func TestAdminClientDo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	var result json.RawMessage
	err := client.do(context.Background(), "GET", server.URL+"/test", nil, &result)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if want := `admin API error 400: bad request`; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestAdminAPIErrorHints(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		contains string
	}{
		{
			name:     "session expired hint",
			body:     `{"error":"Failed to fetch payments merchant ID: 登入已超時，請重新登入"}`,
			contains: "Shopline session expired",
		},
		{
			name:     "csrf hint",
			body:     `{"error":"Failed to extract CSRF token from Shoplytics dashboard"}`,
			contains: "could not fetch dashboard CSRF",
		},
		{
			name:     "generic request failed hint",
			body:     `{"error":"Request failed"}`,
			contains: "verify payload IDs exist for this merchant and session is valid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := adminAPIError(http.StatusBadRequest, []byte(tc.body))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := err.Error(); !strings.Contains(got, tc.contains) {
				t.Fatalf("error = %q, want to contain %q", got, tc.contains)
			}
		})
	}
}

func TestAdminClientCommentOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/orders/ord_123/comments"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want %q (request has body)", got, "application/json")
		}

		body, _ := io.ReadAll(r.Body)
		var req AdminCommentRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.Comment != "test comment" {
			t.Errorf("comment = %q, want %q", req.Comment, "test comment")
		}
		if req.IsPrivate != true {
			t.Errorf("isPrivate = %v, want true", req.IsPrivate)
		}

		_, _ = w.Write([]byte(`{"id":"comment_1"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.CommentOrder(context.Background(), "ord_123", &AdminCommentRequest{
		Comment:   "test comment",
		IsPrivate: true,
	})
	if err != nil {
		t.Fatalf("CommentOrder() error = %v", err)
	}
	if string(result) != `{"id":"comment_1"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientHideProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/products/prod_456/hide"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}

		// Verify no body was sent (or empty body)
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("expected no body, got %s", string(body))
		}

		_, _ = w.Write([]byte(`{"status":"hidden"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.HideProduct(context.Background(), "prod_456")
	if err != nil {
		t.Fatalf("HideProduct() error = %v", err)
	}
	if string(result) != `{"status":"hidden"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetShipmentStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/orders/ord_789/shipment/status"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}

		resp := AdminShipmentStatus{
			OrderNumber:      "ORD-001",
			Executed:         true,
			DeliveryPlatform: "sfexpress",
			Shipment:         "ship_1",
			TrackingNumber:   "SF1234567890",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	status, err := client.GetShipmentStatus(context.Background(), "ord_789")
	if err != nil {
		t.Fatalf("GetShipmentStatus() error = %v", err)
	}
	if status.OrderNumber != "ORD-001" {
		t.Errorf("OrderNumber = %q, want %q", status.OrderNumber, "ORD-001")
	}
	if !status.Executed {
		t.Error("Executed = false, want true")
	}
	if status.TrackingNumber != "SF1234567890" {
		t.Errorf("TrackingNumber = %q, want %q", status.TrackingNumber, "SF1234567890")
	}
}

func TestAdminClientListLivestreams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		// Verify path starts with the merchant prefix
		wantPrefix := "/merchants/test-merchant/livestreams"
		if r.URL.Path != wantPrefix {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPrefix)
		}

		// Verify query params
		if got := r.URL.Query().Get("pageNum"); got != "1" {
			t.Errorf("pageNum = %q, want %q", got, "1")
		}
		if got := r.URL.Query().Get("pageSize"); got != "20" {
			t.Errorf("pageSize = %q, want %q", got, "20")
		}
		if got := r.URL.Query().Get("salesType"); got != "live" {
			t.Errorf("salesType = %q, want %q", got, "live")
		}

		_, _ = w.Write([]byte(`{"items":[],"total":0}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.ListLivestreams(context.Background(), &AdminListStreamsOptions{
		PageNum:   1,
		PageSize:  20,
		SalesType: "live",
	})
	if err != nil {
		t.Fatalf("ListLivestreams() error = %v", err)
	}
	if string(result) != `{"items":[],"total":0}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientDeleteLivestream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/livestreams/stream_123"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	err := client.DeleteLivestream(context.Background(), "stream_123")
	if err != nil {
		t.Fatalf("DeleteLivestream() error = %v", err)
	}
}

func TestAdminClientRemoveStreamProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/livestreams/stream_123/products"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}

		// Verify body is present for DELETE
		body, _ := io.ReadAll(r.Body)
		var req AdminRemoveStreamProductsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if len(req.ProductIDs) != 2 {
			t.Errorf("productIDs count = %d, want 2", len(req.ProductIDs))
		}
		if req.ProductIDs[0] != "prod_1" || req.ProductIDs[1] != "prod_2" {
			t.Errorf("productIDs = %v, want [prod_1 prod_2]", req.ProductIDs)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	err := client.RemoveStreamProducts(context.Background(), "stream_123", &AdminRemoveStreamProductsRequest{
		ProductIDs: []string{"prod_1", "prod_2"},
	})
	if err != nil {
		t.Fatalf("RemoveStreamProducts() error = %v", err)
	}
}

func TestAdminClientListConversations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/conversations"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("platform"); got != "order_messages" {
			t.Errorf("platform = %q, want %q", got, "order_messages")
		}
		if got := r.URL.Query().Get("page_num"); got != "2" {
			t.Errorf("page_num = %q, want %q", got, "2")
		}
		if got := r.URL.Query().Get("page_size"); got != "24" {
			t.Errorf("page_size = %q, want %q", got, "24")
		}
		if got := r.URL.Query().Get("state_filter"); got != "unread" {
			t.Errorf("state_filter = %q, want %q", got, "unread")
		}
		if got := r.URL.Query().Get("is_archived"); got != "true" {
			t.Errorf("is_archived = %q, want %q", got, "true")
		}
		if got := r.URL.Query().Get("search_type"); got != "message" {
			t.Errorf("search_type = %q, want %q", got, "message")
		}
		if got := r.URL.Query().Get("query"); got != "shipping" {
			t.Errorf("query = %q, want %q", got, "shipping")
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	archived := true
	client := newTestAdminClient(server.URL)
	result, err := client.ListConversations(context.Background(), &AdminListConversationsOptions{
		Platform:    "order_messages",
		PageNum:     2,
		PageSize:    24,
		StateFilter: "unread",
		IsArchived:  &archived,
		SearchType:  "message",
		Query:       "shipping",
	})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if string(result) != `{"success":true}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientListConversations_FallbackToLegacyPath(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		switch requestCount {
		case 1:
			if r.URL.Path != "/merchants/test-merchant/conversations" {
				t.Errorf("first path = %q, want %q", r.URL.Path, "/merchants/test-merchant/conversations")
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		case 2:
			if r.URL.Path != "/merchants/test-merchant/message-center/shop-messages" {
				t.Errorf("second path = %q, want %q", r.URL.Path, "/merchants/test-merchant/message-center/shop-messages")
			}
			if got := r.URL.Query().Get("page_num"); got != "1" {
				t.Errorf("page_num = %q, want %q", got, "1")
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.ListConversations(context.Background(), &AdminListConversationsOptions{
		PageNum: 1,
	})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
	if string(result) != `{"success":true}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/conversations/conv_123/messages"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req AdminSendMessageRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.Type != "message" {
			t.Errorf("type = %q, want %q", req.Type, "message")
		}
		if req.Platform != "order_messages" {
			t.Errorf("platform = %q, want %q", req.Platform, "order_messages")
		}
		if req.Content != "ok" {
			t.Errorf("content = %q, want %q", req.Content, "ok")
		}
		if req.ConversationID != "conv_123" {
			t.Errorf("conversation_id = %q, want %q", req.ConversationID, "conv_123")
		}
		_, _ = w.Write([]byte(`{"id":"msg_1"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.SendMessage(context.Background(), "conv_123", &AdminSendMessageRequest{
		Type:     "message",
		Platform: "order_messages",
		Content:  "ok",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if string(result) != `{"id":"msg_1"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientSendMessage_FallbackToLegacyPath(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		switch requestCount {
		case 1:
			if r.URL.Path != "/merchants/test-merchant/conversations/conv_123/messages" {
				t.Errorf("first path = %q, want %q", r.URL.Path, "/merchants/test-merchant/conversations/conv_123/messages")
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		case 2:
			if r.URL.Path != "/merchants/test-merchant/message-center/conversations/conv_123/messages" {
				t.Errorf("second path = %q, want %q", r.URL.Path, "/merchants/test-merchant/message-center/conversations/conv_123/messages")
			}
			_, _ = w.Write([]byte(`{"id":"msg_legacy"}`))
		default:
			t.Fatalf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.SendMessage(context.Background(), "conv_123", &AdminSendMessageRequest{
		Type:     "message",
		Platform: "order_messages",
		Content:  "ok",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
	if string(result) != `{"id":"msg_legacy"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetStreamActiveVideos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/livestreams/stream_123/active-videos"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("platform"); got != "FACEBOOK" {
			t.Errorf("platform = %q, want %q", got, "FACEBOOK")
		}
		_, _ = w.Write([]byte(`{"list":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetStreamActiveVideos(context.Background(), "stream_123", "FACEBOOK")
	if err != nil {
		t.Fatalf("GetStreamActiveVideos() error = %v", err)
	}
	if string(result) != `{"list":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientToggleStreamProductDisplay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/livestreams/stream_123/products/prod_123/toggle-display"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req AdminToggleStreamProductRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.Status != "DISPLAYING" {
			t.Errorf("status = %q, want %q", req.Status, "DISPLAYING")
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.ToggleStreamProductDisplay(context.Background(), "stream_123", "prod_123", &AdminToggleStreamProductRequest{
		Status: "DISPLAYING",
	})
	if err != nil {
		t.Fatalf("ToggleStreamProductDisplay() error = %v", err)
	}
	if string(result) != `{"success":true}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientCreateExpressLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/express-links"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req AdminCreateExpressLinkRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.UserID != "user_1" {
			t.Errorf("user_id = %q, want %q", req.UserID, "user_1")
		}
		if req.Campaign.ID != "camp_1" {
			t.Errorf("campaign._id = %q, want %q", req.Campaign.ID, "camp_1")
		}
		_, _ = w.Write([]byte(`{"url":"https://example.test/express"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.CreateExpressLink(context.Background(), &AdminCreateExpressLinkRequest{
		Products: []AdminExpressLinkProduct{{ID: "prod_1", VariationID: "var_1"}},
		UserID:   "user_1",
		Campaign: AdminExpressLinkCampaign{ID: "camp_1"},
	})
	if err != nil {
		t.Fatalf("CreateExpressLink() error = %v", err)
	}
	if string(result) != `{"url":"https://example.test/express"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetPaymentsPayouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/payments/payouts"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("from"); got != "1704067200000" {
			t.Errorf("from = %q, want %q", got, "1704067200000")
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetPaymentsPayouts(context.Background(), &AdminPaymentsPayoutsOptions{From: 1704067200000})
	if err != nil {
		t.Fatalf("GetPaymentsPayouts() error = %v", err)
	}
	if string(result) != `{"items":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetPaymentsAccountSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/payments/account-summary"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"balance":1000}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetPaymentsAccountSummary(context.Background())
	if err != nil {
		t.Fatalf("GetPaymentsAccountSummary() error = %v", err)
	}
	if string(result) != `{"balance":1000}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetShoplyticsCustomersNewAndReturning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/shoplytics/customers/new-and-returning"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("startDate"); got != "2026-01-01" {
			t.Errorf("startDate = %q, want %q", got, "2026-01-01")
		}
		if got := r.URL.Query().Get("endDate"); got != "2026-01-31" {
			t.Errorf("endDate = %q, want %q", got, "2026-01-31")
		}
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetShoplyticsCustomersNewAndReturning(context.Background(), &AdminShoplyticsNewReturningOptions{
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
	})
	if err != nil {
		t.Fatalf("GetShoplyticsCustomersNewAndReturning() error = %v", err)
	}
	if string(result) != `{"data":{}}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetShoplyticsCustomersFirstOrderChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/shoplytics/customers/first-order-channels"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"channels":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetShoplyticsCustomersFirstOrderChannels(context.Background())
	if err != nil {
		t.Fatalf("GetShoplyticsCustomersFirstOrderChannels() error = %v", err)
	}
	if string(result) != `{"channels":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetShoplyticsPaymentsMethodsGrid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/shoplytics/payments/methods-grid"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"methods":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetShoplyticsPaymentsMethodsGrid(context.Background())
	if err != nil {
		t.Fatalf("GetShoplyticsPaymentsMethodsGrid() error = %v", err)
	}
	if string(result) != `{"methods":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientListOrderComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/orders/ord_123/comments"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}

		// Verify no body was sent
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("expected no body, got %s", string(body))
		}

		resp := []AdminOrderComment{
			{
				ID:        "comment_1",
				Comment:   "Order cancelled per customer request",
				IsPrivate: true,
				Author:    "admin@example.com",
				CreatedAt: "2026-02-15T10:00:00Z",
			},
			{
				ID:        "comment_2",
				Comment:   "Refund processed",
				IsPrivate: false,
				Author:    "support@example.com",
				CreatedAt: "2026-02-15T11:00:00Z",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	comments, err := client.ListOrderComments(context.Background(), "ord_123")
	if err != nil {
		t.Fatalf("ListOrderComments() error = %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
	if comments[0].ID != "comment_1" {
		t.Errorf("comments[0].ID = %q, want %q", comments[0].ID, "comment_1")
	}
	if comments[0].Comment != "Order cancelled per customer request" {
		t.Errorf("comments[0].Comment = %q", comments[0].Comment)
	}
	if !comments[0].IsPrivate {
		t.Error("comments[0].IsPrivate = false, want true")
	}
	if comments[1].Author != "support@example.com" {
		t.Errorf("comments[1].Author = %q", comments[1].Author)
	}
}

func TestAdminClientListOrderComments_NullResponse(t *testing.T) {
	// When the API returns null, the client should return a non-nil empty slice.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`null`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	comments, err := client.ListOrderComments(context.Background(), "ord_empty")
	if err != nil {
		t.Fatalf("ListOrderComments() error = %v", err)
	}
	if comments == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestAdminClientListOrderComments_NotFound(t *testing.T) {
	// When the API returns 404, the client should return an appropriate error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"order not found"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	comments, err := client.ListOrderComments(context.Background(), "ord_nonexistent")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if comments != nil {
		t.Errorf("expected nil comments on error, got %v", comments)
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code 404, got: %v", err)
	}
	if !strings.Contains(err.Error(), "order not found") {
		t.Errorf("error should contain API message, got: %v", err)
	}
}

func TestAdminClientListInstantMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/message-center/instant-messages"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Errorf("page = %q, want %q", got, "1")
		}
		if got := r.URL.Query().Get("page_size"); got != "10" {
			t.Errorf("page_size = %q, want %q", got, "10")
		}
		_, _ = w.Write([]byte(`{"conversations":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.ListInstantMessages(context.Background(), &AdminListInstantMessagesOptions{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListInstantMessages() error = %v", err)
	}
	if string(result) != `{"conversations":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetInstantMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/message-center/instant-messages/conv_456/messages"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("search_type"); got != "up" {
			t.Errorf("search_type = %q, want %q", got, "up")
		}
		_, _ = w.Write([]byte(`{"messages":[]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetInstantMessages(context.Background(), "conv_456", &AdminInstantMessagesQuery{
		SearchType:   "up",
		UseMessageID: "msg_100",
		CreateTime:   "2026-02-15T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("GetInstantMessages() error = %v", err)
	}
	if string(result) != `{"messages":[]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientSendInstantMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/message-center/instant-messages/send"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		body, _ := io.ReadAll(r.Body)
		var req AdminSendInstantMessageRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.ConversationID != "conv_789" {
			t.Errorf("conversation_id = %q, want %q", req.ConversationID, "conv_789")
		}
		if req.Content != "hello" {
			t.Errorf("content = %q, want %q", req.Content, "hello")
		}
		_, _ = w.Write([]byte(`{"id":"im_1"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.SendInstantMessage(context.Background(), &AdminSendInstantMessageRequest{
		ConversationID: "conv_789",
		Content:        "hello",
	})
	if err != nil {
		t.Fatalf("SendInstantMessage() error = %v", err)
	}
	if string(result) != `{"id":"im_1"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetMessageCenterChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/message-center/channels"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"channels":["line","facebook"]}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetMessageCenterChannels(context.Background())
	if err != nil {
		t.Fatalf("GetMessageCenterChannels() error = %v", err)
	}
	if string(result) != `{"channels":["line","facebook"]}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetMessageCenterStaffInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/message-center/staff-info"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"staff_id":"staff_1"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetMessageCenterStaffInfo(context.Background())
	if err != nil {
		t.Fatalf("GetMessageCenterStaffInfo() error = %v", err)
	}
	if string(result) != `{"staff_id":"staff_1"}` {
		t.Errorf("result = %s", string(result))
	}
}

func TestAdminClientGetMessageCenterProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		wantPath := "/merchants/test-merchant/message-center/profiles/scope_123"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		_, _ = w.Write([]byte(`{"name":"Customer A"}`))
	}))
	defer server.Close()

	client := newTestAdminClient(server.URL)
	result, err := client.GetMessageCenterProfile(context.Background(), "scope_123")
	if err != nil {
		t.Fatalf("GetMessageCenterProfile() error = %v", err)
	}
	if string(result) != `{"name":"Customer A"}` {
		t.Errorf("result = %s", string(result))
	}
}

// TestAdminClientInterfaceCompliance verifies that AdminClient implements AdminAPIClient.
var _ AdminAPIClient = (*AdminClient)(nil)

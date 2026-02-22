package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const maxAdminResponseSize = 10 << 20 // 10 MB

func adminBaseURL() string {
	return os.Getenv("SHOPLINE_ADMIN_BASE_URL")
}

// AdminClient is the HTTP client for undocumented Shopline admin endpoints.
// It proxies requests through a separate API service with bearer token auth.
type AdminClient struct {
	baseURL    string
	token      string
	merchantID string
	httpClient *http.Client
}

func newAdminHTTPClient() *http.Client {
	return &http.Client{
		Timeout: httpTimeout,
	}
}

// NewAdminClient creates a new Admin API client.
func NewAdminClient(token, merchantID string) *AdminClient {
	return &AdminClient{
		baseURL:    adminBaseURL(),
		token:      token,
		merchantID: merchantID,
		httpClient: newAdminHTTPClient(),
	}
}

// newTestAdminClient creates an AdminClient for testing with a custom base URL.
func newTestAdminClient(baseURL string) *AdminClient {
	return &AdminClient{
		baseURL:    baseURL,
		token:      "test-token",
		merchantID: "test-merchant",
		httpClient: newAdminHTTPClient(),
	}
}

func (w *AdminClient) merchantPath(path string) string {
	return fmt.Sprintf("%s/merchants/%s%s", w.baseURL, w.merchantID, path)
}

func (w *AdminClient) do(ctx context.Context, method, url string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAdminResponseSize))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return adminAPIError(resp.StatusCode, respBody)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

// AdminAPIError is a typed error for admin API responses with HTTP status >= 400.
type AdminAPIError struct {
	StatusCode int
	Body       string
}

func (e *AdminAPIError) Error() string {
	return fmt.Sprintf("admin API error %d: %s", e.StatusCode, e.Body)
}

func adminAPIError(statusCode int, body []byte) error {
	trimmedBody := strings.TrimSpace(string(body))
	message := trimmedBody

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.Error) != "" {
		message = strings.TrimSpace(payload.Error)
	}

	base := &AdminAPIError{StatusCode: statusCode, Body: message}

	// Known backend failure modes that require user action outside this CLI.
	switch {
	case strings.Contains(message, "登入已超時") || strings.Contains(strings.ToLower(message), "login expired"):
		return fmt.Errorf(
			"%w (Shopline session expired; refresh the login session and retry)",
			base,
		)
	case strings.Contains(message, "Failed to extract CSRF token"):
		return fmt.Errorf(
			"%w (could not fetch dashboard CSRF; verify the Shopline session is valid for this user/account and retry)",
			base,
		)
	case strings.EqualFold(message, "Request failed"):
		return fmt.Errorf(
			"%w (upstream request failed; verify payload IDs exist for this merchant and session is valid)",
			base,
		)
	}

	return base
}

func isAdminRouteUnavailable(err error) bool {
	var adminErr *AdminAPIError
	if errors.As(err, &adminErr) {
		return adminErr.StatusCode == 404 || adminErr.StatusCode == 405
	}
	return false
}

// --- Orders ---

func (w *AdminClient) CommentOrder(ctx context.Context, orderID string, req *AdminCommentRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/comments"), req, &result)
	return result, err
}

func (w *AdminClient) ListOrderComments(ctx context.Context, orderID string) ([]AdminOrderComment, error) {
	var result []AdminOrderComment
	err := w.do(ctx, "GET", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/comments"), nil, &result)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = []AdminOrderComment{}
	}
	return result, nil
}

func (w *AdminClient) AdminRefundOrder(ctx context.Context, orderID string, req *AdminRefundRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/refund"), req, &result)
	return result, err
}

func (w *AdminClient) ReissueReceipt(ctx context.Context, orderID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/receipts/reissue"), nil, &result)
	return result, err
}

// --- Products ---

func (w *AdminClient) HideProduct(ctx context.Context, productID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "PUT", w.merchantPath("/products/"+url.PathEscape(productID)+"/hide"), nil, &result)
	return result, err
}

func (w *AdminClient) PublishProduct(ctx context.Context, productID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "PUT", w.merchantPath("/products/"+url.PathEscape(productID)+"/publish"), nil, &result)
	return result, err
}

func (w *AdminClient) UnpublishProduct(ctx context.Context, productID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "PUT", w.merchantPath("/products/"+url.PathEscape(productID)+"/unpublish"), nil, &result)
	return result, err
}

// --- Shipping ---

func (w *AdminClient) GetShipmentStatus(ctx context.Context, orderID string) (*AdminShipmentStatus, error) {
	var result AdminShipmentStatus
	err := w.do(ctx, "GET", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/shipment/status"), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (w *AdminClient) GetTrackingNumber(ctx context.Context, orderID string) (*AdminTrackingResponse, error) {
	var result AdminTrackingResponse
	err := w.do(ctx, "GET", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/tracking"), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (w *AdminClient) ExecuteShipment(ctx context.Context, orderID string, req *AdminExecuteShipmentRequest) (*AdminTrackingResponse, error) {
	var result AdminTrackingResponse
	err := w.do(ctx, "POST", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/shipment/execute"), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (w *AdminClient) PrintPackingLabel(ctx context.Context, orderID string, req *AdminPrintLabelRequest) (*AdminPackingLabel, error) {
	var result AdminPackingLabel
	err := w.do(ctx, "POST", w.merchantPath("/orders/"+url.PathEscape(orderID)+"/labels/print"), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Livestreams ---

func (w *AdminClient) ListLivestreams(ctx context.Context, opts *AdminListStreamsOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.Int("pageNum", opts.PageNum)
		q.Int("pageSize", opts.PageSize)
		q.String("salesType", opts.SalesType)
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/livestreams"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) GetLivestream(ctx context.Context, streamID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/livestreams/"+url.PathEscape(streamID)), nil, &result)
	return result, err
}

func (w *AdminClient) CreateLivestream(ctx context.Context, req *AdminCreateStreamRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/livestreams"), req, &result)
	return result, err
}

func (w *AdminClient) UpdateLivestream(ctx context.Context, streamID string, req *AdminUpdateStreamRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "PUT", w.merchantPath("/livestreams/"+url.PathEscape(streamID)), req, &result)
	return result, err
}

func (w *AdminClient) DeleteLivestream(ctx context.Context, streamID string) error {
	return w.do(ctx, "DELETE", w.merchantPath("/livestreams/"+url.PathEscape(streamID)), nil, nil)
}

func (w *AdminClient) AddStreamProducts(ctx context.Context, streamID string, req *AdminAddStreamProductsRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/products"), req, &result)
	return result, err
}

func (w *AdminClient) RemoveStreamProducts(ctx context.Context, streamID string, req *AdminRemoveStreamProductsRequest) error {
	return w.do(ctx, "DELETE", w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/products"), req, nil)
}

func (w *AdminClient) StartLivestream(ctx context.Context, streamID string, req *AdminStartStreamRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/start"), req, &result)
	return result, err
}

func (w *AdminClient) EndLivestream(ctx context.Context, streamID string) error {
	return w.do(ctx, "POST", w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/end"), nil, nil)
}

func (w *AdminClient) GetStreamComments(ctx context.Context, streamID string, pageNo int) (json.RawMessage, error) {
	q := NewQuery()
	q.Int("pageNo", pageNo)
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/comments"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) GetStreamActiveVideos(ctx context.Context, streamID, platform string) (json.RawMessage, error) {
	q := NewQuery().String("platform", platform)
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/active-videos"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) ToggleStreamProductDisplay(ctx context.Context, streamID, productID string, req *AdminToggleStreamProductRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(
		ctx,
		"POST",
		w.merchantPath("/livestreams/"+url.PathEscape(streamID)+"/products/"+url.PathEscape(productID)+"/toggle-display"),
		req,
		&result,
	)
	return result, err
}

// --- Message Center ---

func (w *AdminClient) ListConversations(ctx context.Context, opts *AdminListConversationsOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.String("platform", opts.Platform)
		q.Int("page_num", opts.PageNum)
		q.Int("page_size", opts.PageSize)
		q.String("state_filter", opts.StateFilter)
		q.BoolPtr("is_archived", opts.IsArchived)
		q.String("search_type", opts.SearchType)
		q.String("query", opts.Query)
	}
	var result json.RawMessage
	// Public API reference exposes /conversations; keep a legacy fallback for older deployments.
	err := w.do(ctx, "GET", w.merchantPath("/conversations"+q.Build()), nil, &result)
	if err != nil && isAdminRouteUnavailable(err) {
		err = w.do(ctx, "GET", w.merchantPath("/message-center/shop-messages"+q.Build()), nil, &result)
	}
	return result, err
}

func (w *AdminClient) SendMessage(ctx context.Context, conversationID string, req *AdminSendMessageRequest) (json.RawMessage, error) {
	if req == nil {
		return nil, fmt.Errorf("request body is required")
	}

	// API contract requires conversation_id in both path and body.
	body := *req
	body.ConversationID = conversationID

	var result json.RawMessage
	// Public API reference exposes /conversations/{conversationId}/messages; keep legacy fallback.
	err := w.do(ctx, "POST", w.merchantPath("/conversations/"+url.PathEscape(conversationID)+"/messages"), &body, &result)
	if err != nil && isAdminRouteUnavailable(err) {
		err = w.do(
			ctx,
			"POST",
			w.merchantPath("/message-center/conversations/"+url.PathEscape(conversationID)+"/messages"),
			&body,
			&result,
		)
	}
	return result, err
}

func (w *AdminClient) ListInstantMessages(ctx context.Context, opts *AdminListInstantMessagesOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.Int("page", opts.Page)
		q.String("search_type", opts.SearchType)
		q.String("route", opts.Route)
		q.String("unread_type", opts.UnreadType)
		q.Int("page_size", opts.PageSize)
		q.Strings("party_channel_id_list", opts.PartyChannelIDList)
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/message-center/instant-messages"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) GetInstantMessages(ctx context.Context, conversationID string, qv *AdminInstantMessagesQuery) (json.RawMessage, error) {
	q := NewQuery()
	if qv != nil {
		q.String("search_type", qv.SearchType)
		q.String("use_message_id", qv.UseMessageID)
		q.String("create_time", qv.CreateTime)
	}
	var result json.RawMessage
	err := w.do(
		ctx,
		"GET",
		w.merchantPath("/message-center/instant-messages/"+url.PathEscape(conversationID)+"/messages"+q.Build()),
		nil,
		&result,
	)
	return result, err
}

func (w *AdminClient) SendInstantMessage(ctx context.Context, req *AdminSendInstantMessageRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/message-center/instant-messages/send"), req, &result)
	return result, err
}

func (w *AdminClient) GetMessageCenterChannels(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/message-center/channels"), nil, &result)
	return result, err
}

func (w *AdminClient) GetMessageCenterStaffInfo(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/message-center/staff-info"), nil, &result)
	return result, err
}

func (w *AdminClient) GetMessageCenterProfile(ctx context.Context, scopeID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/message-center/profiles/"+url.PathEscape(scopeID)), nil, &result)
	return result, err
}

// --- Express Links ---

func (w *AdminClient) CreateExpressLink(ctx context.Context, req *AdminCreateExpressLinkRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/express-links"), req, &result)
	return result, err
}

// --- Payments ---

func (w *AdminClient) GetPaymentsPayouts(ctx context.Context, opts *AdminPaymentsPayoutsOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.String("from", strconv.FormatInt(opts.From, 10))
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/payments/payouts"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) GetPaymentsAccountSummary(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/payments/account-summary"), nil, &result)
	return result, err
}

// --- Shoplytics ---

func (w *AdminClient) GetShoplyticsCustomersNewAndReturning(ctx context.Context, opts *AdminShoplyticsNewReturningOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.String("startDate", opts.StartDate)
		q.String("endDate", opts.EndDate)
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/shoplytics/customers/new-and-returning"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) GetShoplyticsCustomersFirstOrderChannels(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/shoplytics/customers/first-order-channels"), nil, &result)
	return result, err
}

func (w *AdminClient) GetShoplyticsPaymentsMethodsGrid(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/shoplytics/payments/methods-grid"), nil, &result)
	return result, err
}

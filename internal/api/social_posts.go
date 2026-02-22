package api

import (
	"context"
	"encoding/json"
	"net/url"
)

// --- Social Posts: Channels ---

func (w *AdminClient) GetSocialChannels(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/social-posts/channels"), nil, &result)
	return result, err
}

func (w *AdminClient) GetChannelPosts(ctx context.Context, opts *SocialChannelPostsOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.String("partyChannelId", opts.PartyChannelID)
		q.Int("pageSize", opts.PageSize)
		q.String("type", opts.Type)
		q.String("since", opts.Since)
		q.String("before", opts.Before)
		q.String("after", opts.After)
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/social-posts/channels/posts"+q.Build()), nil, &result)
	return result, err
}

// --- Social Posts: Categories & Products ---

func (w *AdminClient) GetSocialCategories(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/social-posts/categories"), nil, &result)
	return result, err
}

func (w *AdminClient) SearchSocialProducts(ctx context.Context, opts *SocialProductSearchOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.String("query", opts.Query)
		q.Int("page", opts.Page)
		q.Int("pageSize", opts.PageSize)
		q.String("searchType", opts.SearchType)
		q.Strings("categoryIds", opts.CategoryIDs)
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/social-posts/products/search"+q.Build()), nil, &result)
	return result, err
}

// --- Sales Events ---

func (w *AdminClient) ListSalesEvents(ctx context.Context, opts *SalesEventListOptions) (json.RawMessage, error) {
	q := NewQuery()
	if opts != nil {
		q.Int("pageNum", opts.PageNum)
		q.Int("pageSize", opts.PageSize)
		q.String("salesType", opts.SalesType)
	}
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/sales-events"+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) GetSalesEvent(ctx context.Context, eventID string, fieldScopes string) (json.RawMessage, error) {
	q := NewQuery()
	q.String("fieldScopes", fieldScopes)
	var result json.RawMessage
	err := w.do(ctx, "GET", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+q.Build()), nil, &result)
	return result, err
}

func (w *AdminClient) CreateSalesEvent(ctx context.Context, req *CreateSalesEventRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/sales-events"), req, &result)
	return result, err
}

func (w *AdminClient) ScheduleSalesEvent(ctx context.Context, eventID string, req *ScheduleSalesEventRequest) error {
	return w.do(ctx, "PUT", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/schedule"), req, nil)
}

func (w *AdminClient) DeleteSalesEvent(ctx context.Context, eventID string) error {
	return w.do(ctx, "DELETE", w.merchantPath("/sales-events/"+url.PathEscape(eventID)), nil, nil)
}

func (w *AdminClient) PublishSalesEvent(ctx context.Context, eventID string) error {
	return w.do(ctx, "PUT", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/publish"), nil, nil)
}

func (w *AdminClient) AddSalesEventProducts(ctx context.Context, eventID string, req *AddSalesEventProductsRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/products"), req, &result)
	return result, err
}

func (w *AdminClient) UpdateSalesEventProductKeys(ctx context.Context, eventID string, req *UpdateProductKeysRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/products/keys"), req, &result)
	return result, err
}

func (w *AdminClient) LinkFacebookPost(ctx context.Context, eventID string, req *LinkFacebookPostRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/posts/facebook"), req, &result)
	return result, err
}

func (w *AdminClient) LinkInstagramPost(ctx context.Context, eventID string, req *LinkInstagramPostRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/posts/instagram"), req, &result)
	return result, err
}

func (w *AdminClient) LinkFBGroupPost(ctx context.Context, eventID string, req *LinkFBGroupPostRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := w.do(ctx, "POST", w.merchantPath("/sales-events/"+url.PathEscape(eventID)+"/posts/fb-group"), req, &result)
	return result, err
}

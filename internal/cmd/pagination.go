package cmd

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

// PageResult holds items from a single page fetch.
type PageResult[T any] struct {
	Items      []T
	TotalCount int
	HasMore    bool
	Pagination api.Pagination
}

// PageFetcher fetches a single page of results.
type PageFetcher[T any] func(ctx context.Context, page, pageSize int) (PageResult[T], error)

// FetchAllPages accumulates items across multiple pages up to the given limit.
// startPage is the first page number to fetch (typically 1).
// If limit <= 0, it fetches only the first page via the caller's normal path.
func FetchAllPages[T any](ctx context.Context, startPage, pageSize, limit int, fetch PageFetcher[T]) ([]T, int, bool, api.Pagination, error) {
	curPage := startPage
	perPage := pageSize
	if perPage <= 0 || perPage > limit {
		perPage = limit
	}

	items := make([]T, 0, limit)
	totalCount := 0
	hasMore := false
	var pagination api.Pagination

	for len(items) < limit {
		pr, err := fetch(ctx, curPage, perPage)
		if err != nil {
			return nil, 0, false, api.Pagination{}, err
		}
		if totalCount == 0 {
			totalCount = pr.TotalCount
			pagination = pr.Pagination
		}
		items = append(items, pr.Items...)
		hasMore = pr.HasMore

		if !pr.HasMore || len(pr.Items) == 0 {
			break
		}
		curPage++
	}

	if len(items) > limit {
		items = items[:limit]
		hasMore = true
	}

	return items, totalCount, hasMore, pagination, nil
}

// fetchAllIntoResponse is a convenience wrapper that calls FetchAllPages and
// assembles the result into a ListResponse[T].
func fetchAllIntoResponse[T any](ctx context.Context, startPage, pageSize, limit int, fetch PageFetcher[T]) (*api.ListResponse[T], error) {
	items, totalCount, hasMore, pagination, err := FetchAllPages(ctx, startPage, pageSize, limit, fetch)
	if err != nil {
		return nil, err
	}

	perPage := pageSize
	if perPage <= 0 || perPage > limit {
		perPage = limit
	}

	return &api.ListResponse[T]{
		Items:      items,
		Page:       startPage,
		PageSize:   perPage,
		TotalCount: totalCount,
		HasMore:    hasMore,
		Pagination: api.Pagination{
			CurrentPage: startPage,
			PerPage:     perPage,
			TotalCount:  pagination.TotalCount,
			TotalPages:  pagination.TotalPages,
		},
	}, nil
}

// fetchList handles the common pattern of either fetching a single page or
// multiple pages (when --limit is set). The singleFetch function fetches one
// page with the caller's native options. The errMsg is used in error wrapping
// (e.g. "failed to list orders").
func fetchList[T any](
	ctx context.Context,
	limit int,
	startPage int,
	pageSize int,
	singleFetch func() (*api.ListResponse[T], error),
	pagedFetch func(page, size int) (*api.ListResponse[T], error),
	errMsg string,
) (*api.ListResponse[T], error) {
	if limit > 0 {
		return fetchAllIntoResponse(ctx, startPage, pageSize, limit, func(_ context.Context, page, size int) (PageResult[T], error) {
			r, err := pagedFetch(page, size)
			if err != nil {
				return PageResult[T]{}, fmt.Errorf("%s: %w", errMsg, err)
			}
			return PageResult[T]{
				Items:      r.Items,
				TotalCount: r.TotalCount,
				HasMore:    r.HasMore,
				Pagination: r.Pagination,
			}, nil
		})
	}

	r, err := singleFetch()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}
	return r, nil
}

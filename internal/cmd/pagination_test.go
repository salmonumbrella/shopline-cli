package cmd

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

func TestFetchAllPagesRespectsLimit(t *testing.T) {
	calls := 0
	items, totalCount, hasMore, pagination, err := FetchAllPages[int](
		context.Background(),
		1,
		2,
		5,
		func(_ context.Context, page, pageSize int) (PageResult[int], error) {
			calls++
			if pageSize != 2 {
				t.Fatalf("expected pageSize=2, got %d", pageSize)
			}
			switch page {
			case 1:
				return PageResult[int]{
					Items:      []int{1, 2},
					TotalCount: 7,
					HasMore:    true,
					Pagination: api.Pagination{CurrentPage: 1, PerPage: 2, TotalCount: 7, TotalPages: 4},
				}, nil
			case 2:
				return PageResult[int]{
					Items:      []int{3, 4},
					TotalCount: 7,
					HasMore:    true,
					Pagination: api.Pagination{CurrentPage: 2, PerPage: 2, TotalCount: 7, TotalPages: 4},
				}, nil
			case 3:
				return PageResult[int]{
					Items:      []int{5, 6},
					TotalCount: 7,
					HasMore:    true,
					Pagination: api.Pagination{CurrentPage: 3, PerPage: 2, TotalCount: 7, TotalPages: 4},
				}, nil
			default:
				t.Fatalf("unexpected page %d", page)
				return PageResult[int]{}, nil
			}
		},
	)
	if err != nil {
		t.Fatalf("FetchAllPages returned error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 page fetches, got %d", calls)
	}
	if !reflect.DeepEqual(items, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("unexpected items: %#v", items)
	}
	if totalCount != 7 {
		t.Fatalf("totalCount=%d, want 7", totalCount)
	}
	if !hasMore {
		t.Fatalf("expected hasMore=true when items are trimmed to limit")
	}
	if pagination.TotalCount != 7 || pagination.TotalPages != 4 {
		t.Fatalf("unexpected pagination: %#v", pagination)
	}
}

func TestFetchAllPagesStopsWhenNoMore(t *testing.T) {
	calls := 0
	items, totalCount, hasMore, _, err := FetchAllPages[int](
		context.Background(),
		1,
		3,
		10,
		func(_ context.Context, page, _ int) (PageResult[int], error) {
			calls++
			switch page {
			case 1:
				return PageResult[int]{Items: []int{1, 2}, TotalCount: 3, HasMore: true}, nil
			case 2:
				return PageResult[int]{Items: []int{3}, TotalCount: 3, HasMore: false}, nil
			default:
				t.Fatalf("unexpected page %d", page)
				return PageResult[int]{}, nil
			}
		},
	)
	if err != nil {
		t.Fatalf("FetchAllPages returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 page fetches, got %d", calls)
	}
	if !reflect.DeepEqual(items, []int{1, 2, 3}) {
		t.Fatalf("unexpected items: %#v", items)
	}
	if totalCount != 3 {
		t.Fatalf("totalCount=%d, want 3", totalCount)
	}
	if hasMore {
		t.Fatalf("expected hasMore=false on terminal page")
	}
}

func TestFetchAllPagesReturnsFetchError(t *testing.T) {
	wantErr := errors.New("boom")
	_, _, _, _, err := FetchAllPages[int](
		context.Background(),
		1,
		20,
		5,
		func(_ context.Context, _, _ int) (PageResult[int], error) {
			return PageResult[int]{}, wantErr
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestFetchListSingleFetchWhenLimitZero(t *testing.T) {
	singleCalls := 0
	pagedCalls := 0
	resp, err := fetchList[int](
		context.Background(),
		0,
		1,
		20,
		func() (*api.ListResponse[int], error) {
			singleCalls++
			return &api.ListResponse[int]{Items: []int{1, 2}, TotalCount: 2}, nil
		},
		func(_, _ int) (*api.ListResponse[int], error) {
			pagedCalls++
			return nil, errors.New("should not call paged fetch")
		},
		"failed to list",
	)
	if err != nil {
		t.Fatalf("fetchList returned error: %v", err)
	}
	if singleCalls != 1 || pagedCalls != 0 {
		t.Fatalf("unexpected calls: single=%d paged=%d", singleCalls, pagedCalls)
	}
	if !reflect.DeepEqual(resp.Items, []int{1, 2}) {
		t.Fatalf("unexpected items: %#v", resp.Items)
	}
}

func TestFetchListPagedWhenLimitSet(t *testing.T) {
	singleCalls := 0
	pagedCalls := 0
	resp, err := fetchList[int](
		context.Background(),
		3,
		1,
		2,
		func() (*api.ListResponse[int], error) {
			singleCalls++
			return nil, errors.New("should not call single fetch")
		},
		func(page, _ int) (*api.ListResponse[int], error) {
			pagedCalls++
			switch page {
			case 1:
				return &api.ListResponse[int]{
					Items:      []int{1, 2},
					TotalCount: 3,
					HasMore:    true,
					Pagination: api.Pagination{CurrentPage: 1, PerPage: 2, TotalCount: 3, TotalPages: 2},
				}, nil
			case 2:
				return &api.ListResponse[int]{
					Items:      []int{3},
					TotalCount: 3,
					HasMore:    false,
					Pagination: api.Pagination{CurrentPage: 2, PerPage: 2, TotalCount: 3, TotalPages: 2},
				}, nil
			default:
				t.Fatalf("unexpected page %d", page)
				return nil, nil
			}
		},
		"failed to list",
	)
	if err != nil {
		t.Fatalf("fetchList returned error: %v", err)
	}
	if singleCalls != 0 || pagedCalls != 2 {
		t.Fatalf("unexpected calls: single=%d paged=%d", singleCalls, pagedCalls)
	}
	if !reflect.DeepEqual(resp.Items, []int{1, 2, 3}) {
		t.Fatalf("unexpected items: %#v", resp.Items)
	}
}

func TestFetchListWrapsErrors(t *testing.T) {
	_, err := fetchList[int](
		context.Background(),
		2,
		1,
		20,
		func() (*api.ListResponse[int], error) {
			return nil, errors.New("single fail")
		},
		func(_, _ int) (*api.ListResponse[int], error) {
			return nil, errors.New("paged fail")
		},
		"failed to list widgets",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to list widgets") {
		t.Fatalf("expected wrapped error message, got %v", err)
	}
}

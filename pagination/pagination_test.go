package pagination

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestParsePagination(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedPage   int
		expectedSize   int
		expectedOffset int
		expectedLimit  int
	}{
		{
			name:           "default values",
			queryParams:    map[string]string{},
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
			expectedLimit:  20,
		},
		{
			name:           "custom page and size",
			queryParams:    map[string]string{"page": "2", "page_size": "50"},
			expectedPage:   2,
			expectedSize:   50,
			expectedOffset: 50,
			expectedLimit:  50,
		},
		{
			name:           "page less than 1",
			queryParams:    map[string]string{"page": "0"},
			expectedPage:   1, // Should default to 1
			expectedSize:   20,
			expectedOffset: 0,
			expectedLimit:  20,
		},
		{
			name:           "page_size exceeds max",
			queryParams:    map[string]string{"page_size": "200"},
			expectedPage:   1,
			expectedSize:   100, // Should cap at max
			expectedOffset: 0,
			expectedLimit:  100,
		},
		{
			name:           "page_size less than min",
			queryParams:    map[string]string{"page_size": "0"},
			expectedPage:   1,
			expectedSize:   20, // Should use default
			expectedOffset: 0,
			expectedLimit:  20,
		},
		{
			name:           "with sort and order",
			queryParams:    map[string]string{"sort": "created_at", "order": "desc"},
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
			expectedLimit:  20,
		},
		{
			name:           "with filters",
			queryParams:    map[string]string{"filter_status": "active", "filter_role": "admin"},
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
			expectedLimit:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequest(tt.queryParams)
			params := ParsePagination(req, cfg)

			if params.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, params.Page)
			}
			if params.PageSize != tt.expectedSize {
				t.Errorf("expected page_size %d, got %d", tt.expectedSize, params.PageSize)
			}
			if params.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, params.Offset)
			}
			if params.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, params.Limit)
			}
		})
	}
}

func TestPaginateSlice(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	users := make([]User, 100)
	for i := 0; i < 100; i++ {
		users[i] = User{ID: i + 1, Name: "User"}
	}

	tests := []struct {
		name           string
		items          []User
		params         PaginationParams
		expectedCount  int
		expectedTotal  int64
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{
			name:           "first page",
			items:          users,
			params:         PaginationParams{Page: 1, PageSize: 20, Offset: 0, Limit: 20},
			expectedCount:  20,
			expectedTotal:  100,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:           "middle page",
			items:          users,
			params:         PaginationParams{Page: 3, PageSize: 20, Offset: 40, Limit: 20},
			expectedCount:  20,
			expectedTotal:  100,
			expectedHasNext: true,
			expectedHasPrev: true,
		},
		{
			name:           "last page",
			items:          users,
			params:         PaginationParams{Page: 5, PageSize: 20, Offset: 80, Limit: 20},
			expectedCount:  20,
			expectedTotal:  100,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:           "empty slice",
			items:          []User{},
			params:         PaginationParams{Page: 1, PageSize: 20, Offset: 0, Limit: 20},
			expectedCount:  0,
			expectedTotal:  0,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:           "out of range page",
			items:          users,
			params:         PaginationParams{Page: 10, PageSize: 20, Offset: 180, Limit: 20},
			expectedCount:  0,
			expectedTotal:  100,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PaginateSlice(tt.items, tt.params)

			if len(result.Data) != tt.expectedCount {
				t.Errorf("expected %d items, got %d", tt.expectedCount, len(result.Data))
			}
			if result.Pagination.TotalRecords != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, result.Pagination.TotalRecords)
			}
			if result.Pagination.HasNext != tt.expectedHasNext {
				t.Errorf("expected has_next %v, got %v", tt.expectedHasNext, result.Pagination.HasNext)
			}
			if result.Pagination.HasPrev != tt.expectedHasPrev {
				t.Errorf("expected has_prev %v, got %v", tt.expectedHasPrev, result.Pagination.HasPrev)
			}
		})
	}
}

func TestPaginateQuery(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	tests := []struct {
		name          string
		params        PaginationParams
		totalRecords  int64
		fetchCount    int
		expectedError bool
	}{
		{
			name:          "successful pagination",
			params:        PaginationParams{Page: 1, PageSize: 20, Offset: 0, Limit: 20},
			totalRecords:  100,
			fetchCount:    20,
			expectedError: false,
		},
		{
			name:          "empty result set",
			params:        PaginationParams{Page: 1, PageSize: 20, Offset: 0, Limit: 20},
			totalRecords:  0,
			fetchCount:    0,
			expectedError: false,
		},
		{
			name:          "out of range page",
			params:        PaginationParams{Page: 10, PageSize: 20, Offset: 180, Limit: 20},
			totalRecords:  100,
			fetchCount:    0,
			expectedError: false,
		},
		{
			name:          "count function error",
			params:        PaginationParams{Page: 1, PageSize: 20, Offset: 0, Limit: 20},
			totalRecords:  0,
			fetchCount:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			countFn := func(ctx context.Context) (int64, error) {
				if tt.expectedError && tt.name == "count function error" {
					return 0, ErrCountFailed
				}
				return tt.totalRecords, nil
			}

			fetchFn := func(ctx context.Context, limit, offset int) ([]User, error) {
				users := make([]User, tt.fetchCount)
				for i := 0; i < tt.fetchCount; i++ {
					users[i] = User{ID: offset + i + 1, Name: "User"}
				}
				return users, nil
			}

			result, err := PaginateQuery(context.Background(), tt.params, countFn, fetchFn)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result.Data) != tt.fetchCount {
				t.Errorf("expected %d items, got %d", tt.fetchCount, len(result.Data))
			}

			if result.Pagination.TotalRecords != tt.totalRecords {
				t.Errorf("expected total %d, got %d", tt.totalRecords, result.Pagination.TotalRecords)
			}
		})
	}
}

func TestNewPaginationMeta(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		pageSize      int
		totalRecords  int64
		expectedPages int
		expectedNext  *int
		expectedPrev  *int
	}{
		{
			name:          "first page",
			page:          1,
			pageSize:      20,
			totalRecords:  100,
			expectedPages: 5,
			expectedNext:  intPtr(2),
			expectedPrev:  nil,
		},
		{
			name:          "middle page",
			page:          3,
			pageSize:      20,
			totalRecords:  100,
			expectedPages: 5,
			expectedNext:  intPtr(4),
			expectedPrev:  intPtr(2),
		},
		{
			name:          "last page",
			page:          5,
			pageSize:      20,
			totalRecords:  100,
			expectedPages: 5,
			expectedNext:  nil,
			expectedPrev:  intPtr(4),
		},
		{
			name:          "empty result",
			page:          1,
			pageSize:      20,
			totalRecords:  0,
			expectedPages: 1,
			expectedNext:  nil,
			expectedPrev:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewPaginationMeta(tt.page, tt.pageSize, tt.totalRecords)

			if meta.TotalPages != tt.expectedPages {
				t.Errorf("expected %d pages, got %d", tt.expectedPages, meta.TotalPages)
			}

			if !intPtrEqual(meta.NextPage, tt.expectedNext) {
				t.Errorf("expected next_page %v, got %v", tt.expectedNext, meta.NextPage)
			}

			if !intPtrEqual(meta.PrevPage, tt.expectedPrev) {
				t.Errorf("expected prev_page %v, got %v", tt.expectedPrev, meta.PrevPage)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultPageSize != 20 {
		t.Errorf("expected default page size 20, got %d", cfg.DefaultPageSize)
	}

	if cfg.MaxPageSize != 100 {
		t.Errorf("expected max page size 100, got %d", cfg.MaxPageSize)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}

	// Test custom config
	customCfg := DefaultConfig().WithMaxPageSize(200)
	if customCfg.MaxPageSize != 200 {
		t.Errorf("expected max page size 200, got %d", customCfg.MaxPageSize)
	}
}

// Helper functions

func createRequest(params map[string]string) *http.Request {
	u := url.URL{}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	return req
}

func intPtr(i int) *int {
	return &i
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}


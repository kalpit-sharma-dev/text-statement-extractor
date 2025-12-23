package pagination

import (
	"net/http"
	"strconv"
	"strings"
)

// PaginationParams holds validated pagination parameters extracted from HTTP requests.
type PaginationParams struct {
	// Page is the current page number (1-indexed).
	Page int

	// PageSize is the number of items per page.
	PageSize int

	// Offset is the calculated offset for database queries (0-indexed).
	Offset int

	// Limit is the calculated limit for database queries.
	Limit int

	// Sort is the field name to sort by (optional).
	// For multi-field sorting, use SortFields instead.
	Sort string

	// Order is the sort order: "asc" or "desc" (default: "asc").
	// For multi-field sorting, use SortFields instead.
	Order string

	// SortFields is a list of sort fields for multi-field sorting (optional).
	// Takes precedence over Sort/Order if provided.
	// Example: []SortField{{Field: "name", Order: "asc"}, {Field: "created_at", Order: "desc"}}
	SortFields []SortField

	// Filters is a map of filter criteria (optional).
	Filters map[string]interface{}

	// Search is a search query string for text search (optional).
	Search string

	// Fields is a list of field names to include in the response (optional).
	// If empty, all fields are returned. Used for field selection/projection.
	Fields []string
}

// ParsePagination extracts and validates pagination parameters from an HTTP request.
// It works with any framework that provides *http.Request (net/http, Gin, Fiber, Echo).
//
// Query parameters:
//   - page: Page number (default: from config)
//   - page_size: Items per page (default: from config)
//   - sort: Field to sort by (optional, single field)
//   - order: Sort order - "asc" or "desc" (optional, default: "asc")
//   - sort_fields: Comma-separated fields for multi-field sorting (optional)
//   - sort_orders: Comma-separated orders for multi-field sorting (optional)
//   - search: Search query string for text search (optional)
//   - fields: Comma-separated field names to include in response (optional)
//   - filter_*: Filter criteria (optional, e.g., filter_status=active)
//
// Example:
//   GET /users?page=2&page_size=50&sort=created_at&order=desc&search=john
//   GET /users?page=1&page_size=20&sort_fields=name,created_at&sort_orders=asc,desc
//   GET /users?page=1&page_size=20&fields=id,name,email
func ParsePagination(r *http.Request, cfg Config) PaginationParams {
	params := PaginationParams{
		Page:     cfg.DefaultPage,
		PageSize: cfg.DefaultPageSize,
		Order:    "asc",
		Filters:  make(map[string]interface{}),
	}

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
			params.Page = page
		}
	}

	// Parse page_size
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			// Validate and normalize page_size
			if pageSize < cfg.MinPageSize {
				pageSize = cfg.DefaultPageSize
			} else if pageSize > cfg.MaxPageSize {
				pageSize = cfg.MaxPageSize
			}
			params.PageSize = pageSize
		}
	}

	// Parse sort (single field - for backward compatibility)
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = strings.TrimSpace(sort)
	}

	// Parse order (single field - for backward compatibility)
	if order := r.URL.Query().Get("order"); order != "" {
		order = strings.ToLower(strings.TrimSpace(order))
		if order == "desc" || order == "asc" {
			params.Order = order
		}
	}

	// Parse multi-field sorting
	sortStr := r.URL.Query().Get("sort_fields")
	orderStr := r.URL.Query().Get("sort_orders")
	if sortStr != "" {
		params.SortFields = ParseSortFields(sortStr, orderStr)
		// If multi-field sorting is used, clear single field sort
		if len(params.SortFields) > 0 {
			params.Sort = ""
			params.Order = ""
		}
	} else if params.Sort != "" {
		// Convert single field sort to SortFields for consistency
		params.SortFields = []SortField{{Field: params.Sort, Order: params.Order}}
	}

	// Parse field selection
	if fieldsStr := r.URL.Query().Get("fields"); fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		params.Fields = make([]string, 0, len(fields))
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field != "" {
				params.Fields = append(params.Fields, field)
			}
		}
	}

	// Parse search
	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = strings.TrimSpace(search)
	}

	// Parse filters (filter_* query parameters)
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filter_") {
			filterKey := strings.TrimPrefix(key, "filter_")
			// Take the first value if multiple are provided
			if len(values) > 0 {
				params.Filters[filterKey] = values[0]
			}
		}
	}

	// Calculate offset and limit
	params.Offset = (params.Page - 1) * params.PageSize
	params.Limit = params.PageSize

	// Ensure offset is never negative (safety check)
	if params.Offset < 0 {
		params.Offset = 0
	}

	return params
}

// Validate checks if the pagination parameters are valid.
func (p PaginationParams) Validate() error {
	if p.Page < 1 {
		return ErrInvalidPage
	}
	if p.PageSize < 1 {
		return ErrInvalidPageSize
	}
	if p.Offset < 0 {
		return ErrInvalidPage
	}
	return nil
}


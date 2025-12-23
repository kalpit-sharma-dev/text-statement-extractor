package pagination

// PaginationResult represents a paginated response with data and metadata.
// It uses Go generics to provide type-safe pagination for any data type.
type PaginationResult[T any] struct {
	// Data contains the paginated items.
	Data []T `json:"data"`

	// Pagination contains pagination metadata.
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata following the standard API response format.
type PaginationMeta struct {
	// Page is the current page number (1-indexed).
	Page int `json:"page"`

	// PageSize is the number of items per page.
	PageSize int `json:"page_size"`

	// TotalRecords is the total number of records across all pages.
	TotalRecords int64 `json:"total_records"`

	// TotalPages is the total number of pages.
	TotalPages int `json:"total_pages"`

	// HasNext indicates if there is a next page.
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there is a previous page.
	HasPrev bool `json:"has_prev"`

	// NextPage is the next page number, or nil if there is no next page.
	NextPage *int `json:"next_page"`

	// PrevPage is the previous page number, or nil if there is no previous page.
	PrevPage *int `json:"prev_page"`

	// CursorNext is the cursor for the next page (for cursor-based pagination, optional).
	CursorNext string `json:"cursor_next,omitempty"`

	// CursorPrev is the cursor for the previous page (for cursor-based pagination, optional).
	CursorPrev string `json:"cursor_prev,omitempty"`
}

// NewPaginationMeta creates pagination metadata from parameters and total record count.
func NewPaginationMeta(page, pageSize int, totalRecords int64) PaginationMeta {
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	var nextPage *int
	if hasNext {
		np := page + 1
		nextPage = &np
	}

	var prevPage *int
	if hasPrev {
		pp := page - 1
		prevPage = &pp
	}

	return PaginationMeta{
		Page:         page,
		PageSize:     pageSize,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
		NextPage:     nextPage,
		PrevPage:     prevPage,
		// CursorNext and CursorPrev are set separately when using cursor pagination
	}
}

// WithCursors adds cursor fields to pagination metadata.
// This is useful when combining offset and cursor pagination.
func (m PaginationMeta) WithCursors(nextCursor, prevCursor string) PaginationMeta {
	m.CursorNext = nextCursor
	m.CursorPrev = prevCursor
	return m
}


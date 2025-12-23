package pagination

import (
	"context"
	"fmt"
)

// PaginateSlice paginates an in-memory slice of items.
// This is useful for paginating data that's already loaded into memory.
//
// Example:
//   items := []User{...}
//   params := ParsePagination(r, DefaultConfig())
//   result := PaginateSlice(items, params)
func PaginateSlice[T any](items []T, params PaginationParams) PaginationResult[T] {
	totalRecords := int64(len(items))

	// Handle empty slice
	if totalRecords == 0 {
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(params.Page, params.PageSize, 0),
		}
	}

	// Calculate pagination bounds
	start := params.Offset
	end := params.Offset + params.PageSize

	// Handle out-of-range pages
	if start >= len(items) {
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
		}
	}

	// Cap end to slice length
	if end > len(items) {
		end = len(items)
	}

	// Extract paginated items
	paginatedItems := items[start:end]

	return PaginationResult[T]{
		Data:       paginatedItems,
		Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
	}
}

// PaginateQuery paginates data from a database or external API using callback functions.
// This is the primary function for paginating database queries.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - params: Validated pagination parameters
//   - countFn: Function that returns the total record count
//   - fetchFn: Function that fetches the paginated data (limit, offset)
//
// Example:
//   result, err := PaginateQuery(
//       ctx,
//       params,
//       func(ctx context.Context) (int64, error) {
//           return repo.CountUsers(ctx)
//       },
//       func(ctx context.Context, limit, offset int) ([]User, error) {
//           return repo.FetchUsers(ctx, limit, offset)
//       },
//   )
func PaginateQuery[T any](
	ctx context.Context,
	params PaginationParams,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
) (PaginationResult[T], error) {
	// Validate parameters
	if err := params.Validate(); err != nil {
		return PaginationResult[T]{}, fmt.Errorf("invalid pagination params: %w", err)
	}

	// Get total count
	totalRecords, err := countFn(ctx)
	if err != nil {
		return PaginationResult[T]{}, fmt.Errorf("%w: %v", ErrCountFailed, err)
	}

	// Handle empty result set
	if totalRecords == 0 {
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(params.Page, params.PageSize, 0),
		}, nil
	}

	// Check if page is out of range
	totalPages := int((totalRecords + int64(params.PageSize) - 1) / int64(params.PageSize))
	if params.Page > totalPages {
		// Return empty result for out-of-range pages
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
		}, nil
	}

	// Fetch paginated data
	data, err := fetchFn(ctx, params.Limit, params.Offset)
	if err != nil {
		return PaginationResult[T]{}, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	return PaginationResult[T]{
		Data:       data,
		Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
	}, nil
}


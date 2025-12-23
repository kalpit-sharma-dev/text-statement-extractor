package pagination

import (
	"context"
	"fmt"
)

// KeysetParams holds keyset pagination parameters.
// Keyset pagination uses indexed columns for positioning, which is faster than offset for large datasets.
type KeysetParams struct {
	// Limit is the maximum number of items to return.
	Limit int

	// KeysetField is the field name used for keyset positioning (e.g., "id", "created_at").
	KeysetField string

	// KeysetValue is the value of the keyset field from the last item of the previous page.
	// For the first page, this should be nil or empty.
	KeysetValue interface{}

	// Direction is "next" or "prev" (default: "next").
	Direction string
}

// KeysetResult represents a keyset pagination response.
type KeysetResult[T any] struct {
	// Data contains the paginated items.
	Data []T `json:"data"`

	// NextKeyset is the keyset value for the next page (from the last item).
	NextKeyset interface{} `json:"next_keyset,omitempty"`

	// PrevKeyset is the keyset value for the previous page (from the first item).
	PrevKeyset interface{} `json:"prev_keyset,omitempty"`

	// HasNext indicates if there is a next page.
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there is a previous page.
	HasPrev bool `json:"has_prev"`
}

// KeysetExtractor is a function that extracts the keyset value from an item.
type KeysetExtractor[T any] func(T) interface{}

// PaginateKeyset paginates data using keyset (seek) pagination.
// This is more efficient than offset pagination for large datasets.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - params: Keyset pagination parameters
//   - fetchFn: Function that fetches data using keyset positioning
//   - extractor: Function that extracts keyset value from an item
//
// Example:
//   result, err := PaginateKeyset(
//       ctx,
//       KeysetParams{Limit: 20, KeysetField: "id", KeysetValue: nil},
//       func(ctx context.Context, field string, value interface{}, limit int, direction string) ([]User, error) {
//           // SELECT * FROM users WHERE id > ? ORDER BY id LIMIT ?
//           return repo.FetchUsersByKeyset(ctx, field, value, limit, direction)
//       },
//       func(u User) interface{} { return u.ID },
//   )
func PaginateKeyset[T any](
	ctx context.Context,
	params KeysetParams,
	fetchFn func(context.Context, string, interface{}, int, string) ([]T, error),
	extractor KeysetExtractor[T],
) (KeysetResult[T], error) {
	// Validate limit
	if params.Limit < 1 {
		params.Limit = 20 // Default limit
	}

	// Set default direction
	if params.Direction == "" {
		params.Direction = "next"
	}

	// Fetch data
	data, err := fetchFn(ctx, params.KeysetField, params.KeysetValue, params.Limit+1, params.Direction)
	if err != nil {
		return KeysetResult[T]{}, fmt.Errorf("failed to fetch keyset data: %w", err)
	}

	// Determine if there are more pages
	hasNext := len(data) > params.Limit
	if hasNext {
		data = data[:params.Limit] // Remove the extra item used for detection
	}

	// Extract keyset values
	var nextKeyset, prevKeyset interface{}
	hasPrev := params.KeysetValue != nil

	if len(data) > 0 {
		// Next keyset from last item
		if hasNext {
			nextKeyset = extractor(data[len(data)-1])
		}

		// Previous keyset from first item
		if hasPrev {
			prevKeyset = extractor(data[0])
		}
	}

	return KeysetResult[T]{
		Data:       data,
		NextKeyset: nextKeyset,
		PrevKeyset: prevKeyset,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// BuildKeysetQuery builds a SQL WHERE clause for keyset pagination.
// This is a helper function for SQL-based repositories.
//
// Example:
//   whereClause, args := BuildKeysetQuery("id", 123, "next")
//   // Returns: "id > ?", []interface{}{123}
func BuildKeysetQuery(field string, value interface{}, direction string) (string, []interface{}) {
	if value == nil {
		return "", []interface{}{}
	}

	var operator string
	if direction == "prev" {
		operator = "<"
	} else {
		operator = ">"
	}

	return fmt.Sprintf("%s %s ?", field, operator), []interface{}{value}
}

// BuildKeysetQueryWithOrder builds a SQL ORDER BY clause for keyset pagination.
// For "prev" direction, the order is reversed.
//
// Example:
//   orderClause := BuildKeysetQueryWithOrder("id", "next")
//   // Returns: "ORDER BY id ASC"
func BuildKeysetQueryWithOrder(field string, direction string) string {
	order := "ASC"
	if direction == "prev" {
		order = "DESC"
	}
	return fmt.Sprintf("ORDER BY %s %s", field, order)
}


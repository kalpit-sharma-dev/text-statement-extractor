package pagination

// GraphQLEdge represents a GraphQL edge in a connection.
type GraphQLEdge[T any] struct {
	// Node is the item in the edge.
	Node T `json:"node"`

	// Cursor is the cursor for this edge.
	Cursor string `json:"cursor"`
}

// GraphQLPageInfo represents GraphQL page info following Relay spec.
type GraphQLPageInfo struct {
	// HasNextPage indicates if there is a next page.
	HasNextPage bool `json:"hasNextPage"`

	// HasPreviousPage indicates if there is a previous page.
	HasPreviousPage bool `json:"hasPreviousPage"`

	// StartCursor is the cursor of the first edge.
	StartCursor *string `json:"startCursor,omitempty"`

	// EndCursor is the cursor of the last edge.
	EndCursor *string `json:"endCursor,omitempty"`
}

// GraphQLConnection represents a GraphQL connection following Relay spec.
type GraphQLConnection[T any] struct {
	// Edges contains the edges of the connection.
	Edges []GraphQLEdge[T] `json:"edges"`

	// PageInfo contains pagination information.
	PageInfo GraphQLPageInfo `json:"pageInfo"`

	// TotalCount is the total number of items (optional, not in Relay spec but commonly used).
	TotalCount *int64 `json:"totalCount,omitempty"`
}

// ToGraphQLConnection converts a PaginationResult to a GraphQL connection.
//
// Example:
//   connection := ToGraphQLConnection(
//       result,
//       func(u User) string {
//           return EncodeCursor(map[string]interface{}{"id": u.ID}, "next")
//       },
//   )
func ToGraphQLConnection[T any](
	result PaginationResult[T],
	cursorExtractor func(T) string,
) GraphQLConnection[T] {
	edges := make([]GraphQLEdge[T], len(result.Data))

	for i, item := range result.Data {
		cursor := cursorExtractor(item)
		edges[i] = GraphQLEdge[T]{
			Node:   item,
			Cursor: cursor,
		}
	}

	var startCursor, endCursor *string
	if len(edges) > 0 {
		sc := edges[0].Cursor
		ec := edges[len(edges)-1].Cursor
		startCursor = &sc
		endCursor = &ec
	}

	pageInfo := GraphQLPageInfo{
		HasNextPage:     result.Pagination.HasNext,
		HasPreviousPage: result.Pagination.HasPrev,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	connection := GraphQLConnection[T]{
		Edges:    edges,
		PageInfo: pageInfo,
	}

	// Optionally include total count
	totalCount := result.Pagination.TotalRecords
	if totalCount > 0 {
		connection.TotalCount = &totalCount
	}

	return connection
}

// FromGraphQLConnectionArgs converts GraphQL connection arguments to PaginationParams.
// GraphQL connections typically use: first, after, last, before
func FromGraphQLConnectionArgs(first, last *int, after, before *string) PaginationParams {
	params := PaginationParams{
		Page:     1,
		PageSize: 20,
		Order:    "asc",
		Filters:  make(map[string]interface{}),
	}

	// Use first or last to determine page size
	if first != nil && *first > 0 {
		params.PageSize = *first
	} else if last != nil && *last > 0 {
		params.PageSize = *last
	}

	// Parse cursor if provided
	if after != nil && *after != "" {
		// Decode cursor to get position
		// This is a simplified version - adjust based on your cursor format
		params.Filters["after_cursor"] = *after
	}

	if before != nil && *before != "" {
		params.Filters["before_cursor"] = *before
	}

	// Calculate offset (simplified - adjust based on cursor position)
	params.Offset = 0
	params.Limit = params.PageSize

	return params
}


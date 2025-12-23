package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"your-module/pagination"
)

// ============================================================================
// HATEOAS Links Example
// ============================================================================

func ExampleWithLinks(w http.ResponseWriter, r *http.Request) {
	cfg := pagination.DefaultConfig()
	params := pagination.ParsePagination(r, cfg)

	// Get paginated result (from service/repository)
	result := pagination.PaginationResult[User]{
		Data: []User{{ID: 1, Name: "User 1"}},
		Pagination: pagination.NewPaginationMeta(1, 20, 100),
	}

	// Add HATEOAS links
	baseURL := "https://api.example.com/users"
	queryParams := r.URL.Query()
	resultWithLinks := pagination.WithLinks(result, baseURL, queryParams)

	// Response includes links:
	// {
	//   "data": [...],
	//   "pagination": {...},
	//   "links": {
	//     "first": "https://api.example.com/users?page=1&page_size=20",
	//     "last": "https://api.example.com/users?page=5&page_size=20",
	//     "next": "https://api.example.com/users?page=2&page_size=20",
	//     "prev": null,
	//     "self": "https://api.example.com/users?page=1&page_size=20"
	//   }
	// }
	_ = resultWithLinks
}

// ============================================================================
// Keyset Pagination Example
// ============================================================================

type UserRepositoryKeyset struct {
	// db *sql.DB
}

func (r *UserRepositoryKeyset) ListUsersKeyset(
	ctx context.Context,
	params pagination.KeysetParams,
) (pagination.KeysetResult[User], error) {
	return pagination.PaginateKeyset(
		ctx,
		params,
		func(ctx context.Context, field string, value interface{}, limit int, direction string) ([]User, error) {
			// Build SQL query using keyset
			whereClause, args := pagination.BuildKeysetQuery(field, value, direction)
			orderClause := pagination.BuildKeysetQueryWithOrder(field, direction)

			query := fmt.Sprintf(
				"SELECT id, name, email FROM users WHERE %s %s LIMIT %d",
				whereClause, orderClause, limit,
			)

			// Execute query
			// rows, err := r.db.QueryContext(ctx, query, args...)
			// ... process rows

			return []User{}, nil // Mock
		},
		func(u User) interface{} {
			return u.ID // Extract keyset value
		},
	)
}

// Example SQL query for keyset pagination:
// SELECT * FROM users WHERE id > 100 ORDER BY id ASC LIMIT 20
// This is faster than: SELECT * FROM users ORDER BY id LIMIT 20 OFFSET 100

// ============================================================================
// OpenAPI Schema Example
// ============================================================================

func ExampleOpenAPISchema() {
	cfg := pagination.DefaultConfig()

	// Define schema for User item
	userSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":    "integer",
				"format":  "int64",
				"example": 1,
			},
			"name": map[string]interface{}{
				"type":    "string",
				"example": "John Doe",
			},
			"email": map[string]interface{}{
				"type":    "string",
				"format":  "email",
				"example": "john@example.com",
			},
		},
		"required": []string{"id", "name", "email"},
	}

	// Generate OpenAPI schema
	schema := pagination.GenerateOpenAPISchema(cfg, userSchema)

	// Use schema in OpenAPI/Swagger documentation
	// This can be integrated with swaggo/swag, go-swagger, etc.
	fmt.Printf("Query Parameters: %+v\n", schema.QueryParameters)
	fmt.Printf("Response Schema: %+v\n", schema.ResponseSchema)
}

// ============================================================================
// Combined Example: Offset + Keyset + Links
// ============================================================================

func ExampleCombined(w http.ResponseWriter, r *http.Request) {
	cfg := pagination.DefaultConfig()

	// Check if keyset pagination is requested
	if r.URL.Query().Get("keyset") != "" {
		// Use keyset pagination for large datasets
		keysetParams := pagination.KeysetParams{
			Limit:       20,
			KeysetField: "id",
			KeysetValue: nil, // First page
			Direction:   "next",
		}

		// Fetch with keyset
		// result := repo.ListUsersKeyset(ctx, keysetParams)
		_ = keysetParams
	} else {
		// Use offset pagination for smaller datasets
		params := pagination.ParsePagination(r, cfg)
		result, _ := pagination.PaginateQuery(
			r.Context(),
			params,
			func(ctx context.Context) (int64, error) { return 100, nil },
			func(ctx context.Context, limit, offset int) ([]User, error) {
				return []User{}, nil
			},
		)

		// Add HATEOAS links
		baseURL := "https://api.example.com/users"
		resultWithLinks := pagination.WithLinks(result, baseURL, r.URL.Query())

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resultWithLinks)
	}
}


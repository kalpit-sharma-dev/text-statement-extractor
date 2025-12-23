# Pagination Utility for Golang REST APIs

A **generic, reusable, production-ready pagination utility** for Golang REST APIs. Framework-agnostic, database-agnostic, and designed with clean architecture principles.

## Features

- ✅ **Framework-agnostic**: Works with net/http, Gin, Fiber, Echo, and any framework using `*http.Request`
- ✅ **Database-agnostic**: Supports SQL (PostgreSQL, MySQL, YugabyteDB), NoSQL (MongoDB, Cassandra, Aerospike, Elasticsearch), and in-memory data
- ✅ **Type-safe**: Uses Go generics (Go 1.20+) for compile-time type safety
- ✅ **Zero reflection**: No runtime type checking overhead
- ✅ **Zero global state**: All configuration is explicit
- ✅ **Context support**: Full context.Context integration for cancellation and timeouts
- ✅ **Multiple pagination strategies**: Offset-based, cursor-based, and keyset pagination
- ✅ **HATEOAS support**: Built-in pagination links generation
- ✅ **OpenAPI schema**: Generate OpenAPI 3.0 schemas for API documentation

## Installation

```bash
go get your-module/pagination
```

## Quick Start

### Basic Usage

```go
package main

import (
    "net/http"
    "your-module/pagination"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
    // Parse pagination parameters
    cfg := pagination.DefaultConfig()
    params := pagination.ParsePagination(r, cfg)

    // Get paginated results
    result, err := pagination.PaginateQuery(
        r.Context(),
        params,
        func(ctx context.Context) (int64, error) {
            return repo.CountUsers(ctx)
        },
        func(ctx context.Context, limit, offset int) ([]User, error) {
            return repo.FetchUsers(ctx, limit, offset)
        },
    )

    // Send JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### Response Format

```json
{
  "data": [
    {"id": 1, "name": "User 1"},
    {"id": 2, "name": "User 2"}
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_records": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false,
    "next_page": 2,
    "prev_page": null
  }
}
```

## API Reference

### Core Functions

#### `ParsePagination(r *http.Request, cfg Config) PaginationParams`

Parses pagination parameters from HTTP request query parameters.

**Query Parameters:**
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `sort`: Field to sort by (optional)
- `order`: Sort order - "asc" or "desc" (optional, default: "asc")
- `filter_*`: Filter criteria (optional, e.g., `filter_status=active`)

#### `PaginateSlice[T any](items []T, params PaginationParams) PaginationResult[T]`

Paginates an in-memory slice of items.

#### `PaginateQuery[T any](ctx context.Context, params PaginationParams, countFn, fetchFn) (PaginationResult[T], error)`

Paginates data from a database or external API using callback functions.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `params`: Validated pagination parameters
- `countFn`: Function that returns total record count
- `fetchFn`: Function that fetches paginated data (limit, offset)

### Cursor Pagination

#### `PaginateCursor[T any](items []T, cursor CursorParams, extractor func(T) map[string]interface{}) CursorResult[T]`

Paginates using cursor-based pagination (useful for large datasets).

### Keyset Pagination

#### `PaginateKeyset[T any](ctx context.Context, params KeysetParams, fetchFn, extractor) (KeysetResult[T], error)`

Paginates using keyset (seek) pagination (faster than offset for large datasets).

### Advanced Features

#### `WithLinks[T any](result PaginationResult[T], baseURL string, queryParams url.Values) PaginationResultWithLinks[T]`

Adds HATEOAS pagination links to the response.

#### `GenerateOpenAPISchema(cfg Config, itemSchema map[string]interface{}) OpenAPISchema`

Generates OpenAPI 3.0 schema for pagination query parameters and response.

## Examples

### net/http Example

See [examples/nethttp_example.go](examples/nethttp_example.go)

### Gin Framework Example

See [examples/gin_example.go](examples/gin_example.go)

### SQL Repository Example

See [examples/sql_repository_example.go](examples/sql_repository_example.go)

### NoSQL Examples

See [examples/nosql_examples.go](examples/nosql_examples.go) for:
- MongoDB
- Cassandra/YugabyteDB Cassandra
- Aerospike
- Elasticsearch

### In-Memory Example

See [examples/in_memory_example.go](examples/in_memory_example.go)

### Advanced Features

See [examples/advanced_features_example.go](examples/advanced_features_example.go) for:
- HATEOAS links
- Keyset pagination
- OpenAPI schema generation

## Configuration

```go
// Default configuration
cfg := pagination.DefaultConfig()
// DefaultPageSize: 20
// MaxPageSize: 100
// MinPageSize: 1
// DefaultPage: 1

// Custom configuration
cfg := pagination.DefaultConfig().
    WithDefaultPageSize(50).
    WithMaxPageSize(200)
```

## Pagination Strategies

### 1. Offset-Based Pagination

**Use when:**
- Small to medium datasets (< 10K records)
- Total count is needed
- Simple implementation is preferred

**Example:**
```go
params := pagination.ParsePagination(r, cfg)
result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
```

### 2. Cursor-Based Pagination

**Use when:**
- Large datasets (> 10K records)
- Real-time data where total count is expensive
- Avoiding duplicate/missing records during concurrent updates

**Example:**
```go
cursorParams := pagination.CursorParams{
    Cursor:    r.URL.Query().Get("cursor"),
    Limit:     20,
    Direction: "next",
}
result := pagination.PaginateCursor(items, cursorParams, extractor)
```

### 3. Keyset Pagination

**Use when:**
- Very large datasets (> 100K records)
- Deep pagination (page 100+)
- Performance is critical

**Example:**
```go
keysetParams := pagination.KeysetParams{
    Limit:       20,
    KeysetField: "id",
    KeysetValue: nil, // First page
    Direction:   "next",
}
result, err := pagination.PaginateKeyset(ctx, keysetParams, fetchFn, extractor)
```

## Testing

Run tests:

```bash
go test ./pagination/...
```

Run tests with coverage:

```bash
go test -cover ./pagination/...
```

## Performance Considerations

1. **Offset Pagination**: Becomes slow for large offsets (e.g., page 1000+)
2. **Count Queries**: Can be expensive on large tables - consider caching or skipping
3. **Indexing**: Index columns used in `ORDER BY` and `WHERE` clauses
4. **Cursor/Keyset**: Use for datasets > 10K records or deep pagination

## Best Practices

1. **Always use context** for cancellation and timeouts
2. **Set appropriate max_page_size** based on your use case
3. **Use cursor/keyset pagination** for datasets > 10K records
4. **Index columns** used in sorting and filtering
5. **Monitor pagination performance** in production
6. **Implement proper error handling** and logging
7. **Validate and sanitize** all query parameters
8. **Use parameterized queries** to prevent SQL injection

## Security

- ✅ Input validation and sanitization
- ✅ Parameterized queries support
- ✅ Resource limits (max_page_size)
- ✅ Context-based timeouts
- ✅ No SQL injection vulnerabilities

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]

## Support

[Your Support Information Here]


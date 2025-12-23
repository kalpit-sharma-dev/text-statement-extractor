# Golang Pagination Library - Complete Documentation

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Installation & Setup](#installation--setup)
4. [Core Features](#core-features)
5. [Pagination Strategies](#pagination-strategies)
6. [Advanced Features](#advanced-features)
7. [Framework Integrations](#framework-integrations)
8. [Database Integrations](#database-integrations)
9. [API Reference](#api-reference)
10. [Usage Examples](#usage-examples)
11. [Best Practices](#best-practices)
12. [Performance Optimization](#performance-optimization)
13. [Security Considerations](#security-considerations)
14. [Troubleshooting](#troubleshooting)
15. [Migration Guide](#migration-guide)
16. [FAQ](#faq)

---

## Overview

### What is This Library?

The Golang Pagination Library is a **generic, reusable, production-ready pagination utility** for Golang REST APIs. It provides a unified interface for implementing pagination across any framework, database, or data source.

### Key Features

| Feature | Description |
|---------|-------------|
| **Framework-Agnostic** | Works with net/http, Gin, Fiber, Echo, Chi, and any framework using `*http.Request` |
| **Database-Agnostic** | Supports SQL (PostgreSQL, MySQL, YugabyteDB), NoSQL (MongoDB, Cassandra, Aerospike, Elasticsearch), and in-memory data |
| **Type-Safe** | Uses Go generics (Go 1.20+) for compile-time type safety |
| **Zero Reflection** | No runtime type checking overhead |
| **Zero Global State** | All configuration is explicit and passed as parameters |
| **Context Support** | Full context.Context integration for cancellation and timeouts |
| **Multiple Strategies** | Offset-based, cursor-based, and keyset pagination |
| **Enterprise Features** | Caching, retry logic, metrics, compression, and more |

### Use Cases

- ✅ REST API pagination endpoints
- ✅ Large dataset pagination (100K+ records)
- ✅ Real-time data streaming
- ✅ Time-series data (logs, events)
- ✅ GraphQL APIs
- ✅ Batch processing and ETL
- ✅ Data export functionality

---

## Architecture

### Design Principles

1. **Separation of Concerns**: Each component has a single responsibility
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Interface Segregation**: Small, focused interfaces
4. **Open/Closed Principle**: Open for extension, closed for modification
5. **Zero Reflection**: Type safety through generics
6. **Zero Global State**: All configuration is explicit

### Package Structure

```
pagination/
├── Core Components
│   ├── config.go          # Configuration management
│   ├── errors.go          # Custom error types
│   ├── params.go          # Request parameter parsing
│   ├── response.go        # Response structures
│   ├── pagination.go      # Core pagination logic
│   └── helpers.go         # Helper functions
│
├── Pagination Strategies
│   ├── cursor.go          # Cursor-based pagination
│   ├── keyset.go          # Keyset pagination
│   └── time_based.go      # Time-based pagination
│
├── Advanced Features
│   ├── cache.go           # Caching layer
│   ├── metrics.go         # Metrics collection
│   ├── retry.go           # Retry logic
│   ├── etag.go            # ETag support
│   ├── compression.go     # Compression
│   ├── transformers.go    # Response transformers
│   ├── batch.go           # Batch processing
│   ├── streaming.go       # Streaming pagination
│   ├── partition.go       # Partition-aware pagination
│   ├── links.go           # HATEOAS links
│   ├── openapi.go         # OpenAPI schema
│   ├── graphql.go         # GraphQL support
│   └── sorting.go         # Multi-field sorting
│
├── Framework Integrations
│   └── middleware/
│       ├── gin.go         # Gin middleware
│       ├── fiber.go       # Fiber middleware
│       ├── echo.go        # Echo middleware
│       └── chi.go         # Chi middleware
│
├── Database Integrations
│   └── query_builders/
│       ├── gorm.go        # GORM integration
│       └── squirrel.go    # Squirrel integration
│
└── Export
    └── export/
        ├── csv.go         # CSV export
        └── json.go        # JSON export
```

### Component Flow

```
HTTP Request
    ↓
ParsePagination() → PaginationParams
    ↓
PaginateQuery() / PaginateSlice() / PaginateCursor()
    ↓
PaginationResult[T]
    ↓
JSON Response
```

---

## Installation & Setup

### Prerequisites

- Go 1.20 or higher
- Go modules enabled

### Installation

```bash
go get your-module/pagination
```

### Basic Setup

```go
package main

import (
    "net/http"
    "your-module/pagination"
)

func main() {
    // Configure pagination
    cfg := pagination.DefaultConfig()
    
    // Use in your handlers
    http.HandleFunc("/users", ListUsers)
    http.ListenAndServe(":8080", nil)
}
```

### Configuration

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

---

## Core Features

### 1. Request Parameter Parsing

The library automatically parses pagination parameters from HTTP requests.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Page number (1-indexed) |
| `page_size` | int | 20 | Items per page (max: 100) |
| `sort` | string | - | Field to sort by (single field) |
| `sort_fields` | string | - | Comma-separated fields for multi-field sorting |
| `sort_orders` | string | - | Comma-separated orders (asc/desc) |
| `order` | string | asc | Sort order: "asc" or "desc" |
| `search` | string | - | Search query string |
| `fields` | string | - | Comma-separated fields to include |
| `filter_*` | any | - | Filter criteria (e.g., `filter_status=active`) |

#### Example Request

```
GET /users?page=2&page_size=50&sort=created_at&order=desc&search=john&filter_status=active
```

#### Code Example

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    cfg := pagination.DefaultConfig()
    params := pagination.ParsePagination(r, cfg)
    
    // params.Page = 2
    // params.PageSize = 50
    // params.Sort = "created_at"
    // params.Order = "desc"
    // params.Search = "john"
    // params.Filters = map[string]interface{}{"status": "active"}
}
```

### 2. Standard Response Format

All pagination responses follow a consistent format:

```json
{
  "data": [
    {"id": 1, "name": "User 1", "email": "user1@example.com"},
    {"id": 2, "name": "User 2", "email": "user2@example.com"}
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_records": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false,
    "next_page": 2,
    "prev_page": null,
    "cursor_next": "",
    "cursor_prev": ""
  }
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `data` | array | Paginated items |
| `pagination.page` | int | Current page number |
| `pagination.page_size` | int | Items per page |
| `pagination.total_records` | int64 | Total number of records |
| `pagination.total_pages` | int | Total number of pages |
| `pagination.has_next` | bool | Whether there is a next page |
| `pagination.has_prev` | bool | Whether there is a previous page |
| `pagination.next_page` | int\|null | Next page number |
| `pagination.prev_page` | int\|null | Previous page number |
| `pagination.cursor_next` | string | Cursor for next page (optional) |
| `pagination.cursor_prev` | string | Cursor for previous page (optional) |

### 3. In-Memory Pagination

Paginate data that's already loaded into memory:

```go
func ListProducts(w http.ResponseWriter, r *http.Request) {
    // Load all products (from cache, etc.)
    products := []Product{
        {ID: 1, Name: "Product 1"},
        {ID: 2, Name: "Product 2"},
        // ... 1000 products
    }
    
    // Parse pagination params
    cfg := pagination.DefaultConfig()
    params := pagination.ParsePagination(r, cfg)
    
    // Paginate in-memory slice
    result := pagination.PaginateSlice(products, params)
    
    // Return JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### 4. Database Pagination

Paginate data from databases using callback functions:

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    cfg := pagination.DefaultConfig()
    params := pagination.ParsePagination(r, cfg)
    
    result, err := pagination.PaginateQuery(
        r.Context(),
        params,
        // Count function
        func(ctx context.Context) (int64, error) {
            var count int64
            err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
            return count, err
        },
        // Fetch function
        func(ctx context.Context, limit, offset int) ([]User, error) {
            rows, err := db.QueryContext(ctx,
                "SELECT id, name, email FROM users ORDER BY id LIMIT ? OFFSET ?",
                limit, offset,
            )
            if err != nil {
                return nil, err
            }
            defer rows.Close()
            
            var users []User
            for rows.Next() {
                var user User
                if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
                    return nil, err
                }
                users = append(users, user)
            }
            return users, rows.Err()
        },
    )
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

---

## Pagination Strategies

### 1. Offset-Based Pagination

**Best for**: Small to medium datasets (< 10K records), when total count is needed

**How it works**: Uses `LIMIT` and `OFFSET` SQL clauses

**Pros**:
- Simple to implement
- Provides total count
- Easy to navigate to specific pages

**Cons**:
- Slow for large offsets (page 1000+)
- Can skip or duplicate records during concurrent updates

**Example**:

```go
params := pagination.ParsePagination(r, cfg)
result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
```

**SQL Generated**:
```sql
SELECT * FROM users ORDER BY id LIMIT 20 OFFSET 0;
SELECT COUNT(*) FROM users;
```

### 2. Cursor-Based Pagination

**Best for**: Large datasets (> 10K records), real-time data, avoiding duplicates

**How it works**: Uses a cursor (encoded position) to fetch next/previous items

**Pros**:
- Fast regardless of position
- No duplicate/missing records
- Works well with real-time data

**Cons**:
- No total count (or expensive to calculate)
- Can't jump to specific pages
- More complex implementation

**Example**:

```go
cursorParams := pagination.CursorParams{
    Cursor:    r.URL.Query().Get("cursor"),
    Limit:     20,
    Direction: "next",
}

cursorExtractor := func(u User) map[string]interface{} {
    return map[string]interface{}{
        "id":        u.ID,
        "created_at": u.CreatedAt,
    }
}

result := pagination.PaginateCursor(users, cursorParams, cursorExtractor)
```

**Response**:
```json
{
  "data": [...],
  "next_cursor": "eyJpZCI6MjAsImNyZWF0ZWRfYXQiOiIyMDI0LTAxLTE1VDEwOjAwOjAwIn0=",
  "prev_cursor": "",
  "has_next": true,
  "has_prev": false
}
```

### 3. Keyset Pagination

**Best for**: Very large datasets (> 100K records), deep pagination

**How it works**: Uses indexed columns (keyset) for positioning instead of offset

**Pros**:
- Fast for deep pagination
- Constant time complexity
- No offset calculation needed

**Cons**:
- Requires indexed columns
- More complex queries
- No total count

**Example**:

```go
keysetParams := pagination.KeysetParams{
    Limit:       20,
    KeysetField: "id",
    KeysetValue: nil, // First page
    Direction:   "next",
}

result, err := pagination.PaginateKeyset(
    ctx,
    keysetParams,
    func(ctx context.Context, field string, value interface{}, limit int, direction string) ([]User, error) {
        whereClause, args := pagination.BuildKeysetQuery(field, value, direction)
        orderClause := pagination.BuildKeysetQueryWithOrder(field, direction)
        
        query := fmt.Sprintf("SELECT * FROM users WHERE %s %s LIMIT %d",
            whereClause, orderClause, limit)
        
        // Execute query...
        return users, nil
    },
    func(u User) interface{} {
        return u.ID
    },
)
```

**SQL Generated**:
```sql
SELECT * FROM users WHERE id > 100 ORDER BY id ASC LIMIT 20;
```

### 4. Time-Based Pagination

**Best for**: Time-series data (logs, events, transactions)

**How it works**: Paginates using time ranges instead of offsets

**Example**:

```go
params := pagination.ParseTimePaginationParams(r)
// params.StartTime = time.Now().AddDate(0, 0, -7)
// params.EndTime = time.Now()
// params.Limit = 50
// params.Direction = "forward"

result, err := pagination.PaginateByTime(
    ctx,
    params,
    func(ctx context.Context, start, end time.Time, limit int, direction string) ([]LogEntry, error) {
        return repo.FetchLogsByTimeRange(ctx, start, end, limit)
    },
    func(entry LogEntry) time.Time {
        return entry.Timestamp
    },
)
```

---

## Advanced Features

### 1. Multi-Field Sorting

Sort by multiple fields simultaneously:

**Request**:
```
GET /users?sort_fields=name,created_at&sort_orders=asc,desc
```

**Code**:
```go
params := pagination.ParsePagination(r, cfg)
// params.SortFields = [
//     {Field: "name", Order: "asc"},
//     {Field: "created_at", Order: "desc"}
// ]

// Generate SQL ORDER BY clause
orderBy := params.SortFields.ToSQLOrderBy()
// Returns: "ORDER BY name ASC, created_at DESC"
```

**Alternative Format**:
```
GET /users?sort_fields=name:asc,created_at:desc
```

### 2. Field Selection / Projection

Reduce response payload by selecting specific fields:

**Request**:
```
GET /users?fields=id,name,email
```

**Code**:
```go
params := pagination.ParsePagination(r, cfg)
// params.Fields = ["id", "name", "email"]

// Apply field selection in your fetch function
func fetchUsers(ctx context.Context, limit, offset int) ([]User, error) {
    fields := strings.Join(params.Fields, ", ")
    query := fmt.Sprintf("SELECT %s FROM users LIMIT ? OFFSET ?", fields)
    // ...
}
```

### 3. Caching Layer

Cache total counts and pages to reduce database load:

**Cache Interface**:
```go
type Cache interface {
    GetCount(ctx context.Context, key string) (int64, bool)
    SetCount(ctx context.Context, key string, count int64, ttl time.Duration) error
    GetPage(ctx context.Context, key string) ([]byte, bool)
    SetPage(ctx context.Context, key string, data []byte, ttl time.Duration) error
}
```

**Usage**:
```go
// Configure cache
cacheConfig := pagination.DefaultCacheConfig()
cacheConfig.CountTTL = 5 * time.Minute
cacheConfig.PageTTL = 1 * time.Minute
cacheConfig.EnableCountCache = true
cacheConfig.EnablePageCache = false // Usually disabled for dynamic data

// Use with cache
result, err := pagination.PaginateQueryWithCache(
    ctx,
    params,
    cache, // Your cache implementation (Redis, Memcached, etc.)
    cacheConfig,
    countFn,
    fetchFn,
)
```

**Cache Implementation Example (Redis)**:
```go
type RedisCache struct {
    client *redis.Client
}

func (r *RedisCache) GetCount(ctx context.Context, key string) (int64, bool) {
    val, err := r.client.Get(ctx, key).Int64()
    return val, err == nil
}

func (r *RedisCache) SetCount(ctx context.Context, key string, count int64, ttl time.Duration) error {
    return r.client.Set(ctx, key, count, ttl).Err()
}
```

### 4. Framework Middleware

#### Gin Framework

```go
import (
    "github.com/gin-gonic/gin"
    "your-module/pagination"
    "your-module/pagination/middleware"
)

func main() {
    router := gin.Default()
    
    // Add pagination middleware
    router.Use(middleware.GinPaginationMiddleware(pagination.DefaultConfig()))
    
    router.GET("/users", func(c *gin.Context) {
        // Get pagination params from context
        params := middleware.GetPaginationParams(c)
        
        // Use params...
        result, err := pagination.PaginateQuery(
            c.Request.Context(),
            params,
            countFn,
            fetchFn,
        )
        
        c.JSON(http.StatusOK, result)
    })
    
    router.Run(":8080")
}
```

#### Fiber Framework

```go
import (
    "github.com/gofiber/fiber/v2"
    "your-module/pagination"
    "your-module/pagination/middleware"
)

func main() {
    app := fiber.New()
    
    app.Use(middleware.FiberPaginationMiddleware(pagination.DefaultConfig()))
    
    app.Get("/users", func(c *fiber.Ctx) error {
        params := middleware.GetPaginationParams(c)
        // Use params...
        return c.JSON(result)
    })
    
    app.Listen(":8080")
}
```

#### Echo Framework

```go
import (
    "github.com/labstack/echo/v4"
    "your-module/pagination"
    "your-module/pagination/middleware"
)

func main() {
    e := echo.New()
    
    e.Use(middleware.EchoPaginationMiddleware(pagination.DefaultConfig()))
    
    e.GET("/users", func(c echo.Context) error {
        params := middleware.GetPaginationParams(c)
        // Use params...
        return c.JSON(http.StatusOK, result)
    })
    
    e.Start(":8080")
}
```

#### Chi Router

```go
import (
    "github.com/go-chi/chi/v5"
    "your-module/pagination"
    "your-module/pagination/middleware"
)

func main() {
    r := chi.NewRouter()
    
    r.Use(middleware.ChiPaginationMiddleware(pagination.DefaultConfig()))
    
    r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
        params := middleware.GetPaginationParams(r)
        // Use params...
        json.NewEncoder(w).Encode(result)
    })
    
    http.ListenAndServe(":8080", r)
}
```

### 5. Query Builder Integration

#### GORM Integration

```go
import (
    "gorm.io/gorm"
    "your-module/pagination"
    "your-module/pagination/query_builders"
)

func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    // Apply pagination to GORM query
    query := r.db.Model(&User{})
    query = query_builders.ApplyPaginationToGORM(query, params)
    
    var users []User
    if err := query.Find(&users).Error; err != nil {
        return pagination.PaginationResult[User]{}, err
    }
    
    // Get count
    var count int64
    countQuery := r.db.Model(&User{})
    countQuery = query_builders.ApplyPaginationToGORM(countQuery, params)
    if err := countQuery.Count(&count).Error; err != nil {
        return pagination.PaginationResult[User]{}, err
    }
    
    return pagination.PaginationResult[User]{
        Data:       users,
        Pagination: pagination.NewPaginationMeta(params.Page, params.PageSize, count),
    }, nil
}
```

#### Squirrel Integration

```go
import (
    "github.com/Masterminds/squirrel"
    "your-module/pagination"
    "your-module/pagination/query_builders"
)

func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    // Build query
    builder := squirrel.Select("*").From("users")
    builder = query_builders.ApplyPaginationToSquirrel(builder, params)
    
    sql, args, err := builder.ToSql()
    if err != nil {
        return pagination.PaginationResult[User]{}, err
    }
    
    // Execute query
    rows, err := r.db.QueryContext(ctx, sql, args...)
    // ... process rows
    
    // Get count
    countBuilder := squirrel.Select("COUNT(*)").From("users")
    countBuilder = query_builders.BuildCountQuery(countBuilder, params)
    // ... execute count query
    
    return result, nil
}
```

### 6. Batch Processing

Process paginated data in batches (useful for ETL, exports):

```go
type UserProcessor struct{}

func (p *UserProcessor) ProcessBatch(ctx context.Context, items []User, batchNum int) error {
    // Process batch
    for _, user := range items {
        // Do something with user
        fmt.Printf("Processing user %d in batch %d\n", user.ID, batchNum)
    }
    return nil
}

func ExportUsers(ctx context.Context) error {
    params := pagination.PaginationParams{
        Page:     1,
        PageSize: 100,
    }
    
    config := pagination.DefaultBatchConfig()
    config.BatchSize = 100
    config.ContinueOnError = false
    config.ProgressCallback = func(processed int, total int64, batchNum int) {
        fmt.Printf("Processed %d/%d items (batch %d)\n", processed, total, batchNum)
    }
    
    processor := &UserProcessor{}
    
    return pagination.ProcessPaginatedData(
        ctx,
        params,
        countFn,
        fetchFn,
        processor,
        config,
    )
}
```

### 7. Streaming Pagination

Stream results instead of loading into memory:

```go
func StreamUsers(w http.ResponseWriter, r *http.Request) {
    params := pagination.ParsePagination(r, cfg)
    
    config := pagination.DefaultStreamConfig()
    config.Format = "jsonl" // JSON Lines format
    config.IncludeMetadata = false
    
    w.Header().Set("Content-Type", "application/x-ndjson")
    
    err := pagination.StreamPaginatedData(
        r.Context(),
        params,
        countFn,
        fetchFn,
        w,
        config,
    )
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

**Response Format (JSONL)**:
```
{"id":1,"name":"User 1"}
{"id":2,"name":"User 2"}
{"id":3,"name":"User 3"}
```

### 8. Export Functionality

#### CSV Export

```go
import (
    "encoding/csv"
    "os"
    "your-module/pagination/export"
)

func ExportUsersToCSV(w http.ResponseWriter, r *http.Request) {
    result, _ := getPaginatedUsers(r)
    
    writer := csv.NewWriter(w)
    defer writer.Flush()
    
    headers := []string{"ID", "Name", "Email", "Created At"}
    
    fieldExtractor := func(u User) []string {
        return []string{
            strconv.Itoa(u.ID),
            u.Name,
            u.Email,
            u.CreatedAt.Format(time.RFC3339),
        }
    }
    
    err := export.ExportToCSV(result, writer, headers, fieldExtractor)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=users.csv")
}
```

#### JSON Export

```go
import "your-module/pagination/export"

func ExportUsersToJSON(w http.ResponseWriter, r *http.Request) {
    result, _ := getPaginatedUsers(r)
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=users.json")
    
    err := export.ExportToJSON(result, w, true) // Include metadata
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

### 9. GraphQL Support

Convert pagination results to GraphQL connections:

```go
import "your-module/pagination"

func GraphQLUsersResolver(ctx context.Context, first *int, after *string) (pagination.GraphQLConnection[User], error) {
    // Convert GraphQL args to pagination params
    params := pagination.FromGraphQLConnectionArgs(first, nil, after, nil)
    
    // Get paginated result
    result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
    if err != nil {
        return pagination.GraphQLConnection[User]{}, err
    }
    
    // Convert to GraphQL connection
    cursorExtractor := func(u User) string {
        cursor, _ := pagination.EncodeCursor(map[string]interface{}{"id": u.ID}, "next")
        return cursor
    }
    
    return pagination.ToGraphQLConnection(result, cursorExtractor), nil
}
```

**GraphQL Response**:
```json
{
  "edges": [
    {
      "node": {"id": 1, "name": "User 1"},
      "cursor": "eyJpZCI6MX0="
    }
  ],
  "pageInfo": {
    "hasNextPage": true,
    "hasPreviousPage": false,
    "startCursor": "eyJpZCI6MX0=",
    "endCursor": "eyJpZCI6MjB9"
  },
  "totalCount": 100
}
```

### 10. Retry Logic

Automatic retry on transient failures:

```go
retryConfig := pagination.DefaultRetryConfig()
retryConfig.MaxAttempts = 3
retryConfig.InitialBackoff = 100 * time.Millisecond
retryConfig.MaxBackoff = 5 * time.Second
retryConfig.RetryableErrors = func(err error) bool {
    // Only retry on network errors
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection")
}

result, err := pagination.PaginateQueryWithRetry(
    ctx,
    params,
    retryConfig,
    countFn,
    fetchFn,
)
```

### 11. ETag Support

HTTP cache validation:

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    result, err := getPaginatedUsers(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Generate ETag
    etag, err := pagination.GenerateETag(result)
    if err == nil {
        w.Header().Set("ETag", etag)
        
        // Check if client has cached version
        if match := r.Header.Get("If-None-Match"); match == etag {
            w.WriteHeader(http.StatusNotModified)
            return
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### 12. Compression

Compress large responses:

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    result, _ := getPaginatedUsers(r)
    
    // Compress result
    compressed, originalSize, err := pagination.CompressResult(result)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    ratio := pagination.CompressionRatio(originalSize, len(compressed))
    fmt.Printf("Compression ratio: %.2f%%\n", ratio*100)
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Encoding", "gzip")
    w.Write(compressed)
}
```

### 13. Partition-Aware Pagination

Pagination within sharded/partitioned databases:

```go
params := pagination.PartitionPaginationParams{
    PartitionKey:   "user_id",
    PartitionValue: 12345,
    PaginationParams: pagination.PaginationParams{
        Page:     1,
        PageSize: 20,
    },
}

result, err := pagination.PaginatePartition(
    ctx,
    params,
    func(ctx context.Context, key string, value interface{}) (int64, error) {
        // Count records in partition
        return repo.CountByPartition(ctx, key, value)
    },
    func(ctx context.Context, key string, value interface{}, limit, offset int) ([]User, error) {
        // Fetch records from partition
        return repo.FetchByPartition(ctx, key, value, limit, offset)
    },
)
```

### 14. Response Transformers

Transform data before returning (masking, formatting):

```go
// Mask sensitive data
maskedResult := pagination.WithTransformer(result, func(u User) User {
    u.Email = maskEmail(u.Email)
    u.Phone = maskPhone(u.Phone)
    return u
})

// Format fields
formattedResult := pagination.WithTransformer(result, func(u User) User {
    u.CreatedAt = u.CreatedAt.Format("2006-01-02")
    return u
})

// Chain multiple transformers
finalResult := pagination.WithTransformers(
    result,
    maskTransformer,
    formatTransformer,
    customTransformer,
)
```

### 15. HATEOAS Links

Add pagination links to responses:

```go
import "your-module/pagination"

func ListUsers(w http.ResponseWriter, r *http.Request) {
    result, _ := getPaginatedUsers(r)
    
    baseURL := "https://api.example.com/users"
    resultWithLinks := pagination.WithLinks(result, baseURL, r.URL.Query())
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resultWithLinks)
}
```

**Response with Links**:
```json
{
  "data": [...],
  "pagination": {...},
  "links": {
    "first": "https://api.example.com/users?page=1&page_size=20",
    "last": "https://api.example.com/users?page=5&page_size=20",
    "next": "https://api.example.com/users?page=2&page_size=20",
    "prev": null,
    "self": "https://api.example.com/users?page=1&page_size=20"
  }
}
```

### 16. Metrics Collection

Track pagination performance and usage:

```go
type MyMetricsCollector struct{}

func (m *MyMetricsCollector) RecordPaginationRequest(ctx context.Context, params pagination.PaginationParams) {
    // Track request
    metrics.IncrementCounter("pagination.requests", map[string]string{
        "page_size": strconv.Itoa(params.PageSize),
    })
}

func (m *MyMetricsCollector) RecordPaginationDuration(ctx context.Context, duration time.Duration, params pagination.PaginationParams) {
    // Track duration
    metrics.RecordHistogram("pagination.duration", duration.Seconds())
}

func (m *MyMetricsCollector) RecordPaginationError(ctx context.Context, err error, params pagination.PaginationParams) {
    // Track errors
    metrics.IncrementCounter("pagination.errors", map[string]string{
        "error_type": err.Error(),
    })
}

// Use metrics collector
collector := &MyMetricsCollector{}
result, err := pagination.PaginateQueryWithCustomMetrics(
    ctx,
    params,
    countFn,
    fetchFn,
    collector,
)
```

---

## Database Integrations

### PostgreSQL / YugabyteDB

```go
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    return pagination.PaginateQuery(
        ctx,
        params,
        func(ctx context.Context) (int64, error) {
            var count int64
            err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
            return count, err
        },
        func(ctx context.Context, limit, offset int) ([]User, error) {
            query := `
                SELECT id, name, email, created_at 
                FROM users 
                ORDER BY id 
                LIMIT $1 OFFSET $2
            `
            rows, err := r.db.QueryContext(ctx, query, limit, offset)
            // ... process rows
            return users, nil
        },
    )
}
```

### MongoDB

```go
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    return pagination.PaginateQuery(
        ctx,
        params,
        func(ctx context.Context) (int64, error) {
            filter := bson.M{}
            if len(params.Filters) > 0 {
                filter = bson.M(params.Filters)
            }
            return r.collection.CountDocuments(ctx, filter)
        },
        func(ctx context.Context, limit, offset int) ([]User, error) {
            opts := options.Find().
                SetLimit(int64(limit)).
                SetSkip(int64(offset))
            
            if params.Sort != "" {
                sortDir := 1
                if params.Order == "desc" {
                    sortDir = -1
                }
                opts.SetSort(bson.D{{Key: params.Sort, Value: sortDir}})
            }
            
            cursor, err := r.collection.Find(ctx, bson.M{}, opts)
            // ... decode results
            return users, nil
        },
    )
}
```

### Elasticsearch

```go
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    return pagination.PaginateQuery(
        ctx,
        params,
        func(ctx context.Context) (int64, error) {
            req := esapi.CountRequest{Index: []string{"users"}}
            res, err := req.Do(ctx, r.client)
            // ... parse response
            return total, nil
        },
        func(ctx context.Context, limit, offset int) ([]User, error) {
            req := esapi.SearchRequest{
                Index: []string{"users"},
                From:  &offset,
                Size:  &limit,
            }
            res, err := req.Do(ctx, r.client)
            // ... parse hits
            return users, nil
        },
    )
}
```

### Cassandra / YugabyteDB Cassandra

```go
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    // Note: Cassandra doesn't support COUNT(*) efficiently
    // Use cursor-based pagination instead
    
    cursorParams := pagination.CursorParams{
        Cursor:    r.URL.Query().Get("cursor"),
        Limit:     params.PageSize,
        Direction: "next",
    }
    
    // Fetch items
    items := []User{} // Fetch from Cassandra
    
    cursorExtractor := func(u User) map[string]interface{} {
        return map[string]interface{}{
            "id":        u.ID,
            "timestamp": u.Timestamp,
        }
    }
    
    result := pagination.PaginateCursor(items, cursorParams, cursorExtractor)
    return pagination.PaginationResult[User]{
        Data: result.Data,
        Pagination: pagination.PaginationMeta{
            // Note: TotalRecords may be -1 for Cassandra
        },
    }, nil
}
```

---

## API Reference

### Core Functions

#### `ParsePagination(r *http.Request, cfg Config) PaginationParams`

Parses and validates pagination parameters from HTTP request.

**Parameters**:
- `r`: HTTP request
- `cfg`: Pagination configuration

**Returns**: Validated `PaginationParams`

#### `PaginateSlice[T any](items []T, params PaginationParams) PaginationResult[T]`

Paginates an in-memory slice.

**Parameters**:
- `items`: Slice to paginate
- `params`: Pagination parameters

**Returns**: Paginated result

#### `PaginateQuery[T any](ctx context.Context, params PaginationParams, countFn, fetchFn) (PaginationResult[T], error)`

Paginates data from database/API.

**Parameters**:
- `ctx`: Context
- `params`: Pagination parameters
- `countFn`: Function returning total count
- `fetchFn`: Function fetching paginated data

**Returns**: Paginated result or error

### Helper Functions

#### `CalculateOffset(page, pageSize int) int`

Calculates offset from page and page size.

#### `CalculateTotalPages(totalRecords int64, pageSize int) int`

Calculates total pages from total records and page size.

#### `HasNextPage(currentPage, totalPages int) bool`

Checks if there is a next page.

#### `HasPrevPage(currentPage int) bool`

Checks if there is a previous page.

### Cursor Functions

#### `EncodeCursor(position map[string]interface{}, direction string) (string, error)`

Encodes cursor values to base64 string.

#### `DecodeCursor(cursorStr string) (map[string]interface{}, string, error)`

Decodes cursor string to values.

#### `PaginateCursor[T any](items []T, cursor CursorParams, extractor func(T) map[string]interface{}) CursorResult[T]`

Paginates using cursor-based pagination.

### Keyset Functions

#### `PaginateKeyset[T any](ctx context.Context, params KeysetParams, fetchFn, extractor) (KeysetResult[T], error)`

Paginates using keyset pagination.

#### `BuildKeysetQuery(field string, value interface{}, direction string) (string, []interface{})`

Builds SQL WHERE clause for keyset pagination.

### Advanced Functions

#### `PaginateQueryWithCache[T any](ctx, params, cache, cacheConfig, countFn, fetchFn) (PaginationResult[T], error)`

Paginates with caching support.

#### `PaginateQueryWithRetry[T any](ctx, params, retryConfig, countFn, fetchFn) (PaginationResult[T], error)`

Paginates with automatic retry.

#### `PaginateByTime[T any](ctx, params TimePaginationParams, fetchFn, timeExtractor) (TimePaginationResult[T], error)`

Paginates time-series data.

#### `ProcessPaginatedData[T any](ctx, params, countFn, fetchFn, processor, config) error`

Processes paginated data in batches.

#### `StreamPaginatedData[T any](ctx, params, countFn, fetchFn, writer, config) error`

Streams paginated results.

#### `WithLinks[T any](result PaginationResult[T], baseURL string, queryParams url.Values) PaginationResultWithLinks[T]`

Adds HATEOAS links to response.

#### `WithTransformer[T any](result PaginationResult[T], transformer Transformer[T]) PaginationResult[T]`

Applies transformer to result items.

#### `GenerateETag[T any](result PaginationResult[T]) (string, error)`

Generates ETag for HTTP caching.

#### `CompressResult[T any](result PaginationResult[T]) ([]byte, int, error)`

Compresses pagination result.

---

## Best Practices

### 1. Always Use Context

```go
// ✅ Good
result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)

// ❌ Bad
result, err := pagination.PaginateQuery(context.Background(), params, countFn, fetchFn)
```

### 2. Set Appropriate Limits

```go
// ✅ Good - Reasonable max page size
cfg := pagination.DefaultConfig().WithMaxPageSize(100)

// ❌ Bad - Too large, can cause performance issues
cfg := pagination.DefaultConfig().WithMaxPageSize(10000)
```

### 3. Choose the Right Strategy

| Dataset Size | Recommended Strategy |
|--------------|---------------------|
| < 1K records | Offset-based |
| 1K - 10K records | Offset-based (with caching) |
| 10K - 100K records | Cursor-based or Keyset |
| > 100K records | Keyset or Time-based |

### 4. Index Your Database

```sql
-- ✅ Good - Indexed columns for sorting
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_name ON users(name);

-- ❌ Bad - No indexes, slow queries
SELECT * FROM users ORDER BY created_at LIMIT 20 OFFSET 1000;
```

### 5. Cache Count Queries

```go
// ✅ Good - Cache expensive count queries
cacheConfig := pagination.DefaultCacheConfig()
cacheConfig.EnableCountCache = true
cacheConfig.CountTTL = 5 * time.Minute

// ❌ Bad - No caching, every request hits database
```

### 6. Use Parameterized Queries

```go
// ✅ Good - Parameterized, safe from SQL injection
query := "SELECT * FROM users WHERE id = $1 LIMIT $2 OFFSET $3"
rows, err := db.QueryContext(ctx, query, userID, limit, offset)

// ❌ Bad - String concatenation, vulnerable to SQL injection
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d LIMIT %d OFFSET %d", userID, limit, offset)
```

### 7. Handle Errors Gracefully

```go
// ✅ Good - Proper error handling
result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
if err != nil {
    if errors.Is(err, pagination.ErrCountFailed) {
        // Handle count error
        return nil, fmt.Errorf("failed to count records: %w", err)
    }
    return nil, err
}

// ❌ Bad - Ignoring errors
result, _ := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
```

### 8. Validate Input

```go
// ✅ Good - Validate before use
params := pagination.ParsePagination(r, cfg)
if err := params.Validate(); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

// ❌ Bad - No validation
params := pagination.ParsePagination(r, cfg)
// Use params without validation
```

---

## Performance Optimization

### 1. Use Caching

**Impact**: Reduces database load by 80-90% for read-heavy endpoints

```go
cacheConfig := pagination.DefaultCacheConfig()
cacheConfig.EnableCountCache = true
cacheConfig.CountTTL = 5 * time.Minute
```

### 2. Choose Efficient Pagination Strategy

**Offset Pagination**:
- Fast for first pages
- Slow for deep pagination (page 1000+)
- Use for: Small datasets, when total count is needed

**Cursor Pagination**:
- Constant time regardless of position
- No total count needed
- Use for: Large datasets, real-time data

**Keyset Pagination**:
- Fastest for deep pagination
- Requires indexed columns
- Use for: Very large datasets, deep pagination

### 3. Optimize Database Queries

```sql
-- ✅ Good - Use indexes
CREATE INDEX idx_users_created_at ON users(created_at);
SELECT * FROM users ORDER BY created_at LIMIT 20 OFFSET 0;

-- ❌ Bad - No index, full table scan
SELECT * FROM users ORDER BY created_at LIMIT 20 OFFSET 0;
```

### 4. Limit Page Size

```go
// ✅ Good - Reasonable limit
cfg := pagination.DefaultConfig().WithMaxPageSize(100)

// ❌ Bad - Too large, memory issues
cfg := pagination.DefaultConfig().WithMaxPageSize(10000)
```

### 5. Use Streaming for Large Exports

```go
// ✅ Good - Stream large datasets
config := pagination.DefaultStreamConfig()
pagination.StreamPaginatedData(ctx, params, countFn, fetchFn, writer, config)

// ❌ Bad - Load all into memory
result, _ := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
json.NewEncoder(writer).Encode(result) // May OOM for large datasets
```

### 6. Monitor Performance

```go
collector := &MetricsCollector{}
result, err := pagination.PaginateQueryWithCustomMetrics(
    ctx, params, countFn, fetchFn, collector,
)
// Track: request count, duration, errors
```

---

## Security Considerations

### 1. Input Validation

The library automatically validates:
- Page numbers (must be ≥ 1)
- Page sizes (must be within min/max bounds)
- Sort orders (must be "asc" or "desc")

### 2. SQL Injection Prevention

**Always use parameterized queries**:

```go
// ✅ Good
query := "SELECT * FROM users WHERE id = $1 LIMIT $2 OFFSET $3"
rows, err := db.QueryContext(ctx, query, userID, limit, offset)

// ❌ Bad
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d LIMIT %d OFFSET %d", userID, limit, offset)
```

### 3. Field Name Sanitization

For multi-field sorting, the library sanitizes field names:

```go
// ✅ Good - Field names are sanitized
params.SortFields.ToSQLOrderBy() // Only allows alphanumeric, underscore, dot

// ❌ Bad - Direct string interpolation
query := fmt.Sprintf("ORDER BY %s", userInput) // Vulnerable!
```

### 4. Resource Limits

```go
// ✅ Good - Enforce max page size
cfg := pagination.DefaultConfig().WithMaxPageSize(100)

// ❌ Bad - No limits, DoS vulnerability
cfg := pagination.DefaultConfig().WithMaxPageSize(1000000)
```

### 5. Data Masking

Use transformers to mask sensitive data:

```go
maskedResult := pagination.WithTransformer(result, func(u User) User {
    u.Email = maskEmail(u.Email)
    u.SSN = maskSSN(u.SSN)
    return u
})
```

### 6. Context Timeouts

```go
// ✅ Good - Set timeout
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()
result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)

// ❌ Bad - No timeout, can hang indefinitely
result, err := pagination.PaginateQuery(r.Context(), params, countFn, fetchFn)
```

---

## Troubleshooting

### Problem: Slow Pagination Queries

**Symptoms**: High latency on pagination endpoints

**Solutions**:
1. Add indexes on sort columns
2. Use cursor/keyset pagination for large datasets
3. Enable caching for count queries
4. Reduce max page size

### Problem: Out of Memory Errors

**Symptoms**: Application crashes on large pagination requests

**Solutions**:
1. Reduce max page size
2. Use streaming pagination
3. Use cursor-based pagination (no total count)
4. Process data in batches

### Problem: Duplicate/Missing Records

**Symptoms**: Records appear multiple times or are skipped

**Solutions**:
1. Use cursor-based pagination instead of offset
2. Ensure stable sort order (include unique field)
3. Use keyset pagination with unique keys

### Problem: Count Query is Slow

**Symptoms**: COUNT(*) queries take too long

**Solutions**:
1. Enable count caching
2. Use approximate counts for very large tables
3. Skip count for cursor pagination
4. Maintain counter tables

### Problem: Invalid Cursor Errors

**Symptoms**: "invalid cursor" errors

**Solutions**:
1. Ensure cursor is properly encoded
2. Don't modify cursor format
3. Validate cursor before use
4. Handle cursor expiration

---

## Migration Guide

### From Custom Pagination

**Step 1**: Replace custom pagination logic

```go
// Before
page, _ := strconv.Atoi(r.URL.Query().Get("page"))
limit := 20
offset := (page - 1) * limit

// After
params := pagination.ParsePagination(r, pagination.DefaultConfig())
```

**Step 2**: Update repository methods

```go
// Before
func (r *Repo) ListUsers(page, limit int) ([]User, int, error) {
    // Custom pagination logic
}

// After
func (r *Repo) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    return pagination.PaginateQuery(ctx, params, r.countUsers, r.fetchUsers)
}
```

**Step 3**: Update response format

```go
// Before
response := map[string]interface{}{
    "users": users,
    "page": page,
    "total": total,
}

// After
response := pagination.PaginationResult[User]{
    Data: users,
    Pagination: pagination.NewPaginationMeta(params.Page, params.PageSize, total),
}
```

### From GORM Pagination

**Step 1**: Extract count and fetch logic

```go
// Before
var users []User
var total int64
db.Model(&User{}).Count(&total)
db.Offset(offset).Limit(limit).Find(&users)

// After
countFn := func(ctx context.Context) (int64, error) {
    var count int64
    err := db.Model(&User{}).Count(&count).Error
    return count, err
}

fetchFn := func(ctx context.Context, limit, offset int) ([]User, error) {
    var users []User
    err := db.Offset(offset).Limit(limit).Find(&users).Error
    return users, err
}
```

**Step 2**: Use PaginateQuery

```go
result, err := pagination.PaginateQuery(ctx, params, countFn, fetchFn)
```

### From MongoDB Pagination

**Step 1**: Update count and find operations

```go
// Before
count, _ := collection.CountDocuments(ctx, filter)
cursor, _ := collection.Find(ctx, filter, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)))

// After
countFn := func(ctx context.Context) (int64, error) {
    return collection.CountDocuments(ctx, filter)
}

fetchFn := func(ctx context.Context, limit, offset int) ([]User, error) {
    opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit))
    cursor, err := collection.Find(ctx, filter, opts)
    // ... decode
    return users, nil
}
```

---

## FAQ

### Q: Which pagination strategy should I use?

**A**: 
- **Offset**: Small datasets (< 10K), need total count
- **Cursor**: Large datasets (> 10K), real-time data
- **Keyset**: Very large datasets (> 100K), deep pagination
- **Time-based**: Logs, events, time-series data

### Q: How do I handle very large datasets?

**A**: 
1. Use cursor or keyset pagination
2. Enable caching for count queries
3. Use streaming pagination
4. Process in batches

### Q: Can I use this with GraphQL?

**A**: Yes! Use `ToGraphQLConnection()` to convert pagination results to GraphQL Relay connections.

### Q: How do I cache pagination results?

**A**: Implement the `Cache` interface and use `PaginateQueryWithCache()`.

### Q: Can I sort by multiple fields?

**A**: Yes! Use `sort_fields=name,created_at&sort_orders=asc,desc` or the `SortFields` parameter.

### Q: How do I export paginated data?

**A**: Use the export functions: `ExportToCSV()` or `ExportToJSON()`.

### Q: Is this library thread-safe?

**A**: Yes, all functions are stateless and thread-safe.

### Q: Can I use this with my ORM?

**A**: Yes! We provide integrations for GORM and Squirrel. You can also use the callback-based approach with any ORM.

### Q: How do I handle errors?

**A**: The library returns standard Go errors. Check for specific error types:
- `pagination.ErrInvalidPage`
- `pagination.ErrInvalidPageSize`
- `pagination.ErrCountFailed`
- `pagination.ErrFetchFailed`

### Q: Can I customize the response format?

**A**: Yes! Use `WithTransformer()` to modify the response, or build your own response structure.

---

## Conclusion

The Golang Pagination Library provides a comprehensive, production-ready solution for implementing pagination in REST APIs. With support for multiple strategies, frameworks, databases, and advanced features, it's suitable for any use case from simple CRUD APIs to enterprise-scale applications.

### Key Takeaways

1. ✅ **Framework-agnostic**: Works with any Go web framework
2. ✅ **Database-agnostic**: Supports SQL, NoSQL, and in-memory data
3. ✅ **Type-safe**: Uses Go generics for compile-time safety
4. ✅ **Production-ready**: Includes caching, retry, metrics, and more
5. ✅ **Well-tested**: Comprehensive test coverage
6. ✅ **Well-documented**: Complete API reference and examples

### Getting Help

- Check the examples in `pagination/examples/`
- Review the API reference above
- See the architecture document for design details
- Check the troubleshooting section for common issues

---

**Last Updated**: 2024
**Version**: 1.0.0
**License**: [Your License]


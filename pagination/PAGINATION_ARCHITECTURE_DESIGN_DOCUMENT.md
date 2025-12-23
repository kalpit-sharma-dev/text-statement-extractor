# Pagination Utility - Architecture Design Document

## 1. Executive Summary

This document describes the architecture, design decisions, and implementation of a **generic, reusable, production-ready pagination utility** for Golang REST APIs. The utility is designed to be framework-agnostic, database-agnostic, and follows clean architecture principles with zero dependencies on specific ORMs or frameworks.

### 1.1 Objectives
- Provide a unified pagination interface for any REST API
- Support multiple data sources (SQL, NoSQL, in-memory, external APIs)
- Maintain framework independence (net/http, Gin, Fiber, Echo)
- Follow Go best practices with generics, context support, and zero reflection
- Enable easy integration into controller → service → repository layers

---

## 2. Architecture Overview

### 2.1 Design Principles

1. **Separation of Concerns**: Each component has a single responsibility
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Interface Segregation**: Small, focused interfaces
4. **Open/Closed Principle**: Open for extension, closed for modification
5. **Zero Reflection**: Type safety through generics
6. **Zero Global State**: All configuration is explicit and passed as parameters

### 2.2 Package Structure

```
pagination/
├── pagination.go        // Core pagination logic & public API
├── params.go            // Request parameter parsing & validation
├── response.go          // Standard response structures
├── cursor.go            // Cursor-based pagination implementation
├── config.go            // Configuration & defaults
└── errors.go            // Custom error types
```

### 2.3 Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Request                          │
│              (net/http, Gin, Fiber, Echo)               │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   ParsePagination()   │
         │   (params.go)          │
         └───────────┬────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │  PaginationParams     │
         │  (validated & normalized)│
         └───────────┬────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
        ▼                         ▼
┌───────────────┐        ┌───────────────┐
│ PaginateSlice │        │ PaginateQuery │
│ (in-memory)   │        │ (DB/API)      │
└───────┬───────┘        └───────┬───────┘
        │                        │
        └────────────┬───────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │  PaginationResult[T]  │
         │  (response.go)        │
         └───────────┬────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   JSON Response       │
         │   (standard format)   │
         └───────────────────────┘
```

---

## 3. Core Components

### 3.1 Configuration (config.go)

**Purpose**: Centralized configuration management with sensible defaults.

**Key Design Decisions**:
- No global state - configuration is passed explicitly
- Immutable configuration struct
- Builder pattern for configuration creation
- Environment-aware defaults

**Configuration Fields**:
- `DefaultPageSize`: Default items per page (default: 20)
- `MaxPageSize`: Maximum allowed page size (default: 100)
- `MinPageSize`: Minimum allowed page size (default: 1)
- `DefaultPage`: Default page number (default: 1)

### 3.2 Parameters (params.go)

**Purpose**: Parse and validate pagination parameters from HTTP requests.

**Key Design Decisions**:
- Framework-agnostic parsing (works with `*http.Request`)
- Automatic validation and normalization
- Support for query parameters: `page`, `page_size`, `sort`, `order`, `filters`
- Safe defaults for invalid inputs

**Validation Rules**:
- `page < 1` → auto-correct to 1
- `page_size < 1` → use default
- `page_size > max_page_size` → cap to max
- `offset` calculated as `(page - 1) * page_size`, never negative

### 3.3 Response (response.go)

**Purpose**: Standardized pagination response structure.

**Key Design Decisions**:
- Generic type `PaginationResult[T]` for type safety
- Complete metadata (page, total_records, has_next, etc.)
- JSON tags for API compatibility
- Nullable fields for optional values (prev_page, next_page)

**Response Structure**:
```go
type PaginationResult[T any] struct {
    Data       []T              `json:"data"`
    Pagination PaginationMeta    `json:"pagination"`
}

type PaginationMeta struct {
    Page        int    `json:"page"`
    PageSize    int    `json:"page_size"`
    TotalRecords int64 `json:"total_records"`
    TotalPages  int    `json:"total_pages"`
    HasNext     bool   `json:"has_next"`
    HasPrev     bool   `json:"has_prev"`
    NextPage    *int   `json:"next_page"`
    PrevPage    *int   `json:"prev_page"`
}
```

### 3.4 Core Pagination (pagination.go)

**Purpose**: Core pagination logic for different data sources.

**Key Functions**:

1. **PaginateSlice[T]**: In-memory pagination
   - Works with any slice type
   - Zero-copy where possible
   - Handles edge cases (empty slice, out-of-range pages)

2. **PaginateQuery**: Database/API pagination
   - Accepts count and fetch functions
   - Supports context for cancellation
   - Handles errors gracefully

**Key Design Decisions**:
- Generic functions for type safety
- Context support for cancellation and timeouts
- Function-based approach (no interface requirements)
- Separation of count and fetch operations

### 3.5 Cursor Pagination (cursor.go)

**Purpose**: Cursor-based pagination for large datasets and real-time data.

**Key Design Decisions**:
- Encoded cursor strings (base64 JSON)
- Forward and backward pagination
- Cursor validation
- Support for composite keys

**Use Cases**:
- Large datasets where offset becomes slow
- Real-time data where total count is expensive
- Avoiding duplicate/missing records during concurrent updates

---

## 4. Data Flow

### 4.1 Offset-Based Pagination Flow

```
1. HTTP Request → ParsePagination()
   ├─ Extract: page, page_size, sort, filters
   ├─ Validate: bounds checking, normalization
   └─ Return: PaginationParams

2. PaginationParams → Repository Layer
   ├─ Calculate: offset = (page - 1) * page_size
   ├─ Execute: COUNT query (total_records)
   └─ Execute: SELECT query (LIMIT/OFFSET)

3. Repository → PaginateQuery()
   ├─ Call: countFn(ctx) → total_records
   ├─ Call: fetchFn(ctx, limit, offset) → data
   ├─ Calculate: total_pages, has_next, has_prev
   └─ Return: PaginationResult[T]

4. PaginationResult[T] → JSON Response
   └─ Serialize: standard pagination format
```

### 4.2 Cursor-Based Pagination Flow

```
1. HTTP Request → ParseCursor()
   ├─ Extract: cursor, limit, direction
   └─ Return: CursorParams

2. CursorParams → Repository Layer
   ├─ Decode: cursor → position values
   ├─ Execute: SELECT query (WHERE id > cursor, LIMIT)
   └─ Return: data + next_cursor

3. Repository → PaginateCursor()
   ├─ Process: items with cursor extraction
   ├─ Generate: next_cursor, prev_cursor
   └─ Return: CursorResult[T]
```

---

## 5. Integration Patterns

### 5.1 Controller Layer Integration

**Pattern**: Parse pagination params in controller, pass to service.

```go
func (c *Controller) ListUsers(w http.ResponseWriter, r *http.Request) {
    params := pagination.ParsePagination(r, pagination.DefaultConfig())
    result, err := c.service.ListUsers(r.Context(), params)
    // ... handle response
}
```

### 5.2 Service Layer Integration

**Pattern**: Service orchestrates business logic, delegates to repository.

```go
func (s *Service) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    return s.repo.ListUsers(ctx, params)
}
```

### 5.3 Repository Layer Integration

**Pattern**: Repository implements data access with pagination.

```go
func (r *Repository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
    return pagination.PaginateQuery(
        ctx,
        params,
        r.countUsers,
        r.fetchUsers,
    )
}
```

---

## 6. Database-Specific Considerations

### 6.1 SQL Databases (PostgreSQL, YugabyteDB)

**Implementation**:
- Use `COUNT(*)` for total records
- Use `LIMIT` and `OFFSET` for data fetching
- Support `ORDER BY` for sorting
- Parameterized queries for security

**Performance**:
- Index on sort columns
- Consider cursor pagination for large datasets
- Use `EXPLAIN ANALYZE` to optimize queries

### 6.2 NoSQL Databases

**Aerospike**:
- Use `Query` with `RecordSet` and pagination
- Implement custom count aggregation
- Use secondary indexes for filtering

**Cassandra/YugabyteDB Cassandra**:
- Use `SELECT` with `LIMIT` and token-based pagination
- Avoid `COUNT(*)` (expensive)
- Prefer cursor-based pagination

**MongoDB**:
- Use `CountDocuments()` and `Find()` with `Skip()` and `Limit()`
- Use indexes for sorting
- Consider aggregation pipeline for complex filters

**Elasticsearch**:
- Use `_search` API with `from` and `size`
- Use `track_total_hits` for accurate counts
- Support scroll API for large result sets

---

## 7. Error Handling

### 7.1 Error Types

- `ErrInvalidPage`: Page number is invalid
- `ErrInvalidPageSize`: Page size is invalid
- `ErrInvalidCursor`: Cursor is malformed
- `ErrCountFailed`: Count operation failed
- `ErrFetchFailed`: Fetch operation failed

### 7.2 Error Handling Strategy

- Validation errors: Return 400 Bad Request
- Database errors: Return 500 Internal Server Error
- Context cancellation: Return 408 Request Timeout
- Log errors with context for debugging

---

## 8. Performance Considerations

### 8.1 Offset Pagination Limitations

- **Problem**: `OFFSET` becomes slow for large offsets (e.g., page 1000)
- **Solution**: Use cursor-based pagination for large datasets
- **Threshold**: Consider cursor pagination when offset > 10,000

### 8.2 Count Query Optimization

- **Problem**: `COUNT(*)` can be expensive on large tables
- **Solutions**:
  - Cache count results for static data
  - Use approximate counts for very large datasets
  - Skip count for cursor pagination

### 8.3 Indexing Strategy

- Index columns used in `ORDER BY`
- Index columns used in `WHERE` filters
- Composite indexes for multi-column sorting/filtering

### 8.4 Memory Management

- Stream large result sets (don't load all into memory)
- Use `LIMIT` to cap result sizes
- Implement result set caching for frequently accessed pages

---

## 9. Testing Strategy

### 9.1 Unit Tests

- Parameter parsing and validation
- Edge cases (empty results, out-of-range pages)
- Configuration validation
- Cursor encoding/decoding

### 9.2 Integration Tests

- Full pagination flow with mock repositories
- Framework-specific parsing (Gin, Fiber, Echo)
- Database-specific implementations

### 9.3 Performance Tests

- Benchmark offset vs cursor pagination
- Measure count query performance
- Test with large datasets (100K+ records)

---

## 10. Advanced Features

### 10.1 Keyset Pagination

- Alternative to offset for large datasets
- Uses indexed columns for positioning
- Faster than offset for deep pagination

### 10.2 HATEOAS Links

- Generate pagination links in response
- Support for first, last, next, prev links
- Framework-aware URL generation

### 10.3 OpenAPI Schema Generation

- Generate OpenAPI 3.0 schemas for pagination
- Include query parameters and response structure
- Support for Swagger/ReDoc documentation

### 10.4 Rate Limit Safe Pagination

- Respect rate limits in external API calls
- Implement backoff strategies
- Cache results when appropriate

---

## 11. Security Considerations

### 11.1 Input Validation

- Sanitize all query parameters
- Validate page and page_size ranges
- Prevent SQL injection (parameterized queries)
- Prevent NoSQL injection (typed filters)

### 11.2 Resource Limits

- Enforce max_page_size to prevent DoS
- Implement query timeout via context
- Rate limit pagination endpoints

### 11.3 Data Exposure

- Don't expose internal IDs in cursors (use encrypted tokens)
- Sanitize error messages (don't leak DB structure)
- Implement proper access control

---

## 12. Future Enhancements

1. **GraphQL Support**: Pagination for GraphQL resolvers
2. **Streaming Pagination**: Server-sent events for real-time data
3. **Multi-database Transactions**: Pagination across distributed databases
4. **Intelligent Caching**: Cache-aware pagination
5. **Analytics Integration**: Track pagination usage patterns

---

## 13. Conclusion

This pagination utility provides a robust, flexible, and production-ready solution for implementing pagination in Golang REST APIs. It follows clean architecture principles, supports multiple data sources, and is designed for easy integration into existing codebases.

The design prioritizes:
- **Type Safety**: Through Go generics
- **Performance**: Through efficient algorithms and database optimization
- **Flexibility**: Through framework and database agnosticism
- **Maintainability**: Through clean code and comprehensive documentation

---

## Appendix A: API Reference

See implementation files for complete API documentation.

## Appendix B: Migration Guide

For migrating from existing pagination implementations:
1. Replace custom pagination logic with `ParsePagination()`
2. Update repository methods to use `PaginateQuery()`
3. Update response serialization to use `PaginationResult[T]`
4. Test thoroughly with existing data

## Appendix C: Best Practices

1. Always use context for cancellation
2. Set appropriate max_page_size based on your use case
3. Use cursor pagination for datasets > 10K records
4. Index columns used in sorting and filtering
5. Monitor pagination performance in production
6. Implement proper error handling and logging


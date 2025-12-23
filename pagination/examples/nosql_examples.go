package main

import (
	"context"
	"fmt"

	"your-module/pagination"
)

// Transaction represents a sample transaction model.
type Transaction struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
	AccountID string  `json:"account_id"`
}

// ============================================================================
// Aerospike Example
// ============================================================================

type TransactionRepositoryAerospike struct {
	// client *as.Client
}

func (r *TransactionRepositoryAerospike) ListTransactions(
	ctx context.Context,
	params pagination.PaginationParams,
) (pagination.PaginationResult[Transaction], error) {
	return pagination.PaginateQuery(
		ctx,
		params,
		func(ctx context.Context) (int64, error) {
			// Aerospike count operation
			// policy := as.NewQueryPolicy()
			// stmt := as.NewStatement("namespace", "set")
			// recordSet, err := r.client.Query(policy, stmt)
			// count := 0
			// for record := range recordSet.Results() {
			//     count++
			// }
			// return int64(count), nil
			return 1000, nil // Mock
		},
		func(ctx context.Context, limit, offset int) ([]Transaction, error) {
			// Aerospike query with pagination
			// policy := as.NewQueryPolicy()
			// policy.MaxRecords = uint64(limit)
			// stmt := as.NewStatement("namespace", "set")
			// recordSet, err := r.client.Query(policy, stmt)
			// ... process records
			return []Transaction{}, nil // Mock
		},
	)
}

// ============================================================================
// MongoDB Example
// ============================================================================

type TransactionRepositoryMongo struct {
	// collection *mongo.Collection
}

func (r *TransactionRepositoryMongo) ListTransactions(
	ctx context.Context,
	params pagination.PaginationParams,
) (pagination.PaginationResult[Transaction], error) {
	return pagination.PaginateQuery(
		ctx,
		params,
		func(ctx context.Context) (int64, error) {
			// MongoDB count
			// filter := bson.M{}
			// if len(params.Filters) > 0 {
			//     filter = bson.M(params.Filters)
			// }
			// count, err := r.collection.CountDocuments(ctx, filter)
			// return count, err
			return 1000, nil // Mock
		},
		func(ctx context.Context, limit, offset int) ([]Transaction, error) {
			// MongoDB find with pagination
			// filter := bson.M{}
			// if len(params.Filters) > 0 {
			//     filter = bson.M(params.Filters)
			// }
			// opts := options.Find().
			//     SetLimit(int64(limit)).
			//     SetSkip(int64(offset))
			// if params.Sort != "" {
			//     sortDir := 1
			//     if params.Order == "desc" {
			//         sortDir = -1
			//     }
			//     opts.SetSort(bson.D{{Key: params.Sort, Value: sortDir}})
			// }
			// cursor, err := r.collection.Find(ctx, filter, opts)
			// ... decode results
			return []Transaction{}, nil // Mock
		},
	)
}

// ============================================================================
// Cassandra/YugabyteDB Cassandra Example
// ============================================================================

type TransactionRepositoryCassandra struct {
	// session *gocql.Session
}

func (r *TransactionRepositoryCassandra) ListTransactions(
	ctx context.Context,
	params pagination.PaginationParams,
) (pagination.PaginationResult[Transaction], error) {
	// Note: Cassandra doesn't support COUNT(*) efficiently.
	// For cursor-based pagination, use token-based approach.
	return pagination.PaginateQuery(
		ctx,
		params,
		func(ctx context.Context) (int64, error) {
			// Avoid COUNT(*) in Cassandra - it's expensive
			// Option 1: Use approximate count from metadata
			// Option 2: Maintain a counter table
			// Option 3: Return -1 and skip total_records calculation
			return -1, nil // Indicates count not available
		},
		func(ctx context.Context, limit, offset int) ([]Transaction, error) {
			// Cassandra query with LIMIT
			// Note: Cassandra doesn't support OFFSET efficiently
			// Use token-based pagination for better performance
			// query := r.session.Query(
			//     "SELECT id, amount, timestamp, account_id FROM transactions LIMIT ?",
			//     limit,
			// ).WithContext(ctx)
			// ... execute and scan
			return []Transaction{}, nil // Mock
		},
	)
}

// ============================================================================
// Elasticsearch Example
// ============================================================================

type TransactionRepositoryElasticsearch struct {
	// client *elasticsearch.Client
}

func (r *TransactionRepositoryElasticsearch) ListTransactions(
	ctx context.Context,
	params pagination.PaginationParams,
) (pagination.PaginationResult[Transaction], error) {
	return pagination.PaginateQuery(
		ctx,
		params,
		func(ctx context.Context) (int64, error) {
			// Elasticsearch count API
			// req := esapi.CountRequest{
			//     Index: []string{"transactions"},
			// }
			// res, err := req.Do(ctx, r.client)
			// ... parse response for total
			return 1000, nil // Mock
		},
		func(ctx context.Context, limit, offset int) ([]Transaction, error) {
			// Elasticsearch search API with from/size
			// req := esapi.SearchRequest{
			//     Index: []string{"transactions"},
			//     From:  &offset,
			//     Size:  &limit,
			// }
			// res, err := req.Do(ctx, r.client)
			// ... parse response and extract hits
			return []Transaction{}, nil // Mock
		},
	)
}

// ============================================================================
// Cursor-Based Pagination Example (Recommended for NoSQL)
// ============================================================================

func (r *TransactionRepositoryMongo) ListTransactionsCursor(
	ctx context.Context,
	cursorParams pagination.CursorParams,
) (pagination.CursorResult[Transaction], error) {
	// Fetch all matching items (or use a more efficient approach)
	items := []Transaction{} // Fetch from database

	// Extract cursor from transaction (using ID and timestamp)
	cursorExtractor := func(t Transaction) map[string]interface{} {
		return map[string]interface{}{
			"id":        t.ID,
			"timestamp": t.Timestamp,
		}
	}

	return pagination.PaginateCursor(items, cursorParams, cursorExtractor), nil
}

// Example usage in a handler
func ExampleCursorHandler() {
	// Parse cursor from query parameter
	// cursor := r.URL.Query().Get("cursor")
	// limit := 20
	// if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
	//     limit, _ = strconv.Atoi(limitStr)
	// }

	cursorParams := pagination.CursorParams{
		Cursor:    "", // From query parameter
		Limit:     20,
		Direction: "next",
	}

	// Use cursor pagination
	// result := repo.ListTransactionsCursor(ctx, cursorParams)
	_ = cursorParams
}


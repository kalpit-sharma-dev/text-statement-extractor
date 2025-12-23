package pagination

import (
	"context"
	"time"
)

// MetricsCollector is an interface for collecting pagination metrics.
// Implement this interface to track pagination performance and usage.
type MetricsCollector interface {
	// RecordPaginationRequest records a pagination request with its parameters.
	RecordPaginationRequest(ctx context.Context, params PaginationParams)

	// RecordPaginationDuration records the time taken for a pagination operation.
	RecordPaginationDuration(ctx context.Context, duration time.Duration, params PaginationParams)

	// RecordPaginationError records a pagination error.
	RecordPaginationError(ctx context.Context, err error, params PaginationParams)

	// RecordTotalRecordsQuery records the time taken for a total records count query.
	RecordTotalRecordsQuery(ctx context.Context, duration time.Duration)

	// RecordDataFetchQuery records the time taken for a data fetch query.
	RecordDataFetchQuery(ctx context.Context, duration time.Duration, limit, offset int)
}

// NoOpMetricsCollector is a no-op implementation of MetricsCollector.
// Use this when metrics collection is not needed.
type NoOpMetricsCollector struct{}

func (n NoOpMetricsCollector) RecordPaginationRequest(ctx context.Context, params PaginationParams) {}
func (n NoOpMetricsCollector) RecordPaginationDuration(ctx context.Context, duration time.Duration, params PaginationParams) {}
func (n NoOpMetricsCollector) RecordPaginationError(ctx context.Context, err error, params PaginationParams) {}
func (n NoOpMetricsCollector) RecordTotalRecordsQuery(ctx context.Context, duration time.Duration) {}
func (n NoOpMetricsCollector) RecordDataFetchQuery(ctx context.Context, duration time.Duration, limit, offset int) {}

// DefaultMetricsCollector is the default metrics collector (no-op).
var DefaultMetricsCollector MetricsCollector = NoOpMetricsCollector{}

// SetMetricsCollector sets the global metrics collector.
// This is optional and can be set to nil to disable metrics.
func SetMetricsCollector(collector MetricsCollector) {
	if collector == nil {
		DefaultMetricsCollector = NoOpMetricsCollector{}
		return
	}
	DefaultMetricsCollector = collector
}

// PaginateQueryWithMetrics is like PaginateQuery but includes metrics collection.
func PaginateQueryWithMetrics[T any](
	ctx context.Context,
	params PaginationParams,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
) (PaginationResult[T], error) {
	start := time.Now()

	// Record request
	DefaultMetricsCollector.RecordPaginationRequest(ctx, params)

	// Execute pagination
	result, err := PaginateQuery(ctx, params, countFn, fetchFn)

	// Record duration
	duration := time.Since(start)
	DefaultMetricsCollector.RecordPaginationDuration(ctx, duration, params)

	// Record error if any
	if err != nil {
		DefaultMetricsCollector.RecordPaginationError(ctx, err, params)
	}

	return result, err
}

// PaginateQueryWithCustomMetrics is like PaginateQuery but uses a custom metrics collector.
func PaginateQueryWithCustomMetrics[T any](
	ctx context.Context,
	params PaginationParams,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
	collector MetricsCollector,
) (PaginationResult[T], error) {
	if collector == nil {
		return PaginateQuery(ctx, params, countFn, fetchFn)
	}

	start := time.Now()
	collector.RecordPaginationRequest(ctx, params)

	// Wrap count function with metrics
	countFnWithMetrics := func(ctx context.Context) (int64, error) {
		countStart := time.Now()
		count, err := countFn(ctx)
		collector.RecordTotalRecordsQuery(ctx, time.Since(countStart))
		return count, err
	}

	// Wrap fetch function with metrics
	fetchFnWithMetrics := func(ctx context.Context, limit, offset int) ([]T, error) {
		fetchStart := time.Now()
		data, err := fetchFn(ctx, limit, offset)
		collector.RecordDataFetchQuery(ctx, time.Since(fetchStart), limit, offset)
		return data, err
	}

	result, err := PaginateQuery(ctx, params, countFnWithMetrics, fetchFnWithMetrics)

	duration := time.Since(start)
	collector.RecordPaginationDuration(ctx, duration, params)

	if err != nil {
		collector.RecordPaginationError(ctx, err, params)
	}

	return result, err
}


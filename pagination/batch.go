package pagination

import (
	"context"
	"fmt"
)

// BatchProcessor is an interface for processing paginated data in batches.
type BatchProcessor[T any] interface {
	// ProcessBatch processes a batch of items.
	// Return an error to stop processing.
	ProcessBatch(ctx context.Context, items []T, batchNum int) error
}

// BatchConfig holds configuration for batch processing.
type BatchConfig struct {
	// BatchSize is the number of items to process per batch (default: 100).
	BatchSize int

	// ContinueOnError determines if processing should continue on error (default: false).
	ContinueOnError bool

	// ProgressCallback is called after each batch is processed.
	ProgressCallback func(processed int, total int64, batchNum int)
}

// DefaultBatchConfig returns a batch configuration with sensible defaults.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		BatchSize:       100,
		ContinueOnError: false,
		ProgressCallback: nil,
	}
}

// ProcessPaginatedData processes all paginated data in batches.
// This is useful for ETL, exports, and bulk operations.
//
// Example:
//   processor := &MyBatchProcessor{}
//   err := ProcessPaginatedData(
//       ctx,
//       params,
//       countFn,
//       fetchFn,
//       processor,
//       DefaultBatchConfig(),
//   )
func ProcessPaginatedData[T any](
	ctx context.Context,
	params PaginationParams,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
	processor BatchProcessor[T],
	config BatchConfig,
) error {
	if processor == nil {
		return fmt.Errorf("batch processor cannot be nil")
	}

	// Get total count
	totalRecords, err := countFn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	if totalRecords == 0 {
		return nil
	}

	// Set default batch size
	if config.BatchSize < 1 {
		config.BatchSize = 100
	}

	// Process in batches
	processed := 0
	batchNum := 0
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = config.BatchSize
	}

	for offset := 0; offset < int(totalRecords); offset += pageSize {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Fetch batch
		limit := pageSize
		if offset+limit > int(totalRecords) {
			limit = int(totalRecords) - offset
		}

		items, err := fetchFn(ctx, limit, offset)
		if err != nil {
			if config.ContinueOnError {
				// Log error and continue
				continue
			}
			return fmt.Errorf("failed to fetch batch at offset %d: %w", offset, err)
		}

		// Process batch
		batchNum++
		if err := processor.ProcessBatch(ctx, items, batchNum); err != nil {
			if config.ContinueOnError {
				// Log error and continue
				continue
			}
			return fmt.Errorf("batch processor failed at batch %d: %w", batchNum, err)
		}

		processed += len(items)

		// Call progress callback
		if config.ProgressCallback != nil {
			config.ProgressCallback(processed, totalRecords, batchNum)
		}

		// Break if no more items
		if len(items) < limit {
			break
		}
	}

	return nil
}

// ProcessAllPages processes all pages of paginated data.
// Similar to ProcessPaginatedData but uses page-based iteration.
func ProcessAllPages[T any](
	ctx context.Context,
	params PaginationParams,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
	processor BatchProcessor[T],
	config BatchConfig,
) error {
	// Get total count
	totalRecords, err := countFn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	if totalRecords == 0 {
		return nil
	}

	// Calculate total pages
	totalPages := CalculateTotalPages(totalRecords, params.PageSize)
	batchNum := 0

	// Process each page
	for page := 1; page <= totalPages; page++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Update params for current page
		currentParams := params
		currentParams.Page = page
		currentParams.Offset = (page - 1) * params.PageSize

		// Fetch page
		items, err := fetchFn(ctx, currentParams.Limit, currentParams.Offset)
		if err != nil {
			if config.ContinueOnError {
				continue
			}
			return fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		// Process batch
		batchNum++
		if err := processor.ProcessBatch(ctx, items, batchNum); err != nil {
			if config.ContinueOnError {
				continue
			}
			return fmt.Errorf("batch processor failed at page %d: %w", page, err)
		}

		// Call progress callback
		if config.ProgressCallback != nil {
			processed := page * params.PageSize
			if processed > int(totalRecords) {
				processed = int(totalRecords)
			}
			config.ProgressCallback(processed, totalRecords, batchNum)
		}
	}

	return nil
}


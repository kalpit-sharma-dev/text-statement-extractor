package pagination

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// StreamConfig holds configuration for streaming pagination.
type StreamConfig struct {
	// BatchSize is the number of items to stream per batch (default: 100).
	BatchSize int

	// Format is the output format: "json" or "jsonl" (default: "jsonl").
	Format string

	// IncludeMetadata determines if pagination metadata should be included (default: false).
	IncludeMetadata bool
}

// DefaultStreamConfig returns a stream configuration with sensible defaults.
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		BatchSize:       100,
		Format:          "jsonl", // JSON Lines format (one JSON object per line)
		IncludeMetadata: false,
	}
}

// StreamPaginatedData streams paginated results to a writer.
// This is memory-efficient for large datasets.
//
// Example:
//   err := StreamPaginatedData(
//       ctx,
//       params,
//       countFn,
//       fetchFn,
//       writer,
//       DefaultStreamConfig(),
//   )
func StreamPaginatedData[T any](
	ctx context.Context,
	params PaginationParams,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
	writer io.Writer,
	config StreamConfig,
) error {
	if writer == nil {
		return fmt.Errorf("writer cannot be nil")
	}

	// Set default batch size
	if config.BatchSize < 1 {
		config.BatchSize = 100
	}

	// Get total count
	totalRecords, err := countFn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	// Write metadata if requested
	if config.IncludeMetadata {
		meta := map[string]interface{}{
			"total_records": totalRecords,
			"page_size":     params.PageSize,
			"format":        config.Format,
		}
		metaJSON, _ := json.Marshal(meta)
		if _, err := writer.Write(metaJSON); err != nil {
			return fmt.Errorf("failed to write metadata: %w", err)
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	// Stream data in batches
	offset := 0
	encoder := json.NewEncoder(writer)

	for offset < int(totalRecords) {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Calculate batch size
		batchSize := config.BatchSize
		if offset+batchSize > int(totalRecords) {
			batchSize = int(totalRecords) - offset
		}

		// Fetch batch
		items, err := fetchFn(ctx, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch batch at offset %d: %w", offset, err)
		}

		// Stream items
		for _, item := range items {
			if config.Format == "jsonl" {
				// JSON Lines format: one JSON object per line
				if err := encoder.Encode(item); err != nil {
					return fmt.Errorf("failed to encode item: %w", err)
				}
			} else {
				// Regular JSON array format
				itemJSON, err := json.Marshal(item)
				if err != nil {
					return fmt.Errorf("failed to marshal item: %w", err)
				}
				if _, err := writer.Write(itemJSON); err != nil {
					return fmt.Errorf("failed to write item: %w", err)
				}
				if _, err := writer.Write([]byte(",")); err != nil {
					return fmt.Errorf("failed to write comma: %w", err)
				}
			}
		}

		offset += len(items)

		// Break if no more items
		if len(items) < batchSize {
			break
		}
	}

	return nil
}


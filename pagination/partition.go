package pagination

import (
	"context"
	"fmt"
)

// PartitionPaginationParams holds parameters for partition-aware pagination.
type PartitionPaginationParams struct {
	// PartitionKey is the partition/shard key.
	PartitionKey string

	// PartitionValue is the value of the partition key.
	PartitionValue interface{}

	// Standard pagination parameters
	PaginationParams
}

// PartitionPaginationResult represents a partition-aware pagination response.
type PartitionPaginationResult[T any] struct {
	// Standard pagination result
	PaginationResult[T]

	// PartitionInfo contains partition information.
	PartitionInfo PartitionInfo `json:"partition_info"`
}

// PartitionInfo contains information about the partition used.
type PartitionInfo struct {
	// PartitionKey is the partition key used.
	PartitionKey string `json:"partition_key"`

	// PartitionValue is the partition value used.
	PartitionValue interface{} `json:"partition_value"`

	// TotalPartitions is the total number of partitions (optional).
	TotalPartitions *int `json:"total_partitions,omitempty"`
}

// PaginatePartition paginates data within a specific partition.
// This is useful for sharded/partitioned databases.
//
// Example:
//   params := PartitionPaginationParams{
//       PartitionKey:   "user_id",
//       PartitionValue: 12345,
//       PaginationParams: pagination.PaginationParams{
//           Page:     1,
//           PageSize: 20,
//       },
//   }
//   result, err := PaginatePartition(ctx, params, countFn, fetchFn)
func PaginatePartition[T any](
	ctx context.Context,
	params PartitionPaginationParams,
	countFn func(context.Context, string, interface{}) (int64, error),
	fetchFn func(context.Context, string, interface{}, int, int) ([]T, error),
) (PartitionPaginationResult[T], error) {
	if params.PartitionKey == "" {
		return PartitionPaginationResult[T]{}, fmt.Errorf("partition key cannot be empty")
	}

	// Get count for this partition
	totalRecords, err := countFn(ctx, params.PartitionKey, params.PartitionValue)
	if err != nil {
		return PartitionPaginationResult[T]{}, fmt.Errorf("failed to count partition records: %w", err)
	}

	// Handle empty result set
	if totalRecords == 0 {
		return PartitionPaginationResult[T]{
			PaginationResult: PaginationResult[T]{
				Data:       []T{},
				Pagination: NewPaginationMeta(params.Page, params.PageSize, 0),
			},
			PartitionInfo: PartitionInfo{
				PartitionKey:   params.PartitionKey,
				PartitionValue: params.PartitionValue,
			},
		}, nil
	}

	// Check if page is out of range
	totalPages := int((totalRecords + int64(params.PageSize) - 1) / int64(params.PageSize))
	if params.Page > totalPages {
		return PartitionPaginationResult[T]{
			PaginationResult: PaginationResult[T]{
				Data:       []T{},
				Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
			},
			PartitionInfo: PartitionInfo{
				PartitionKey:   params.PartitionKey,
				PartitionValue: params.PartitionValue,
			},
		}, nil
	}

	// Fetch data from partition
	data, err := fetchFn(ctx, params.PartitionKey, params.PartitionValue, params.Limit, params.Offset)
	if err != nil {
		return PartitionPaginationResult[T]{}, fmt.Errorf("failed to fetch partition data: %w", err)
	}

	return PartitionPaginationResult[T]{
		PaginationResult: PaginationResult[T]{
			Data:       data,
			Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
		},
		PartitionInfo: PartitionInfo{
			PartitionKey:   params.PartitionKey,
			PartitionValue: params.PartitionValue,
		},
	}, nil
}

// PaginateAcrossPartitions paginates data across multiple partitions.
// This aggregates results from multiple partitions.
func PaginateAcrossPartitions[T any](
	ctx context.Context,
	partitions []PartitionPaginationParams,
	countFn func(context.Context, string, interface{}) (int64, error),
	fetchFn func(context.Context, string, interface{}, int, int) ([]T, error),
	aggregateFn func([][]T) []T,
) (PaginationResult[T], error) {
	if len(partitions) == 0 {
		return PaginationResult[T]{}, fmt.Errorf("at least one partition is required")
	}

	// Aggregate counts from all partitions
	var totalRecords int64
	for _, partition := range partitions {
		count, err := countFn(ctx, partition.PartitionKey, partition.PartitionValue)
		if err != nil {
			return PaginationResult[T]{}, fmt.Errorf("failed to count partition %v: %w", partition.PartitionValue, err)
		}
		totalRecords += count
	}

	if totalRecords == 0 {
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(1, 20, 0),
		}, nil
	}

	// Fetch from each partition and aggregate
	var allData [][]T
	for _, partition := range partitions {
		data, err := fetchFn(ctx, partition.PartitionKey, partition.PartitionValue, partition.Limit, partition.Offset)
		if err != nil {
			return PaginationResult[T]{}, fmt.Errorf("failed to fetch from partition %v: %w", partition.PartitionValue, err)
		}
		allData = append(allData, data)
	}

	// Aggregate results
	aggregatedData := aggregateFn(allData)

	// Use first partition's pagination params for metadata
	firstPartition := partitions[0]
	return PaginationResult[T]{
		Data:       aggregatedData,
		Pagination: NewPaginationMeta(firstPartition.Page, firstPartition.PageSize, totalRecords),
	}, nil
}


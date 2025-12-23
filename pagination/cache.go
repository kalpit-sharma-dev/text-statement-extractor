package pagination

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Cache is an interface for caching pagination results and counts.
type Cache interface {
	// GetCount retrieves a cached total count.
	GetCount(ctx context.Context, key string) (int64, bool)

	// SetCount stores a total count with TTL.
	SetCount(ctx context.Context, key string, count int64, ttl time.Duration) error

	// GetPage retrieves a cached page of data.
	GetPage(ctx context.Context, key string) ([]byte, bool)

	// SetPage stores a page of data with TTL.
	SetPage(ctx context.Context, key string, data []byte, ttl time.Duration) error

	// InvalidateCount invalidates a cached count.
	InvalidateCount(ctx context.Context, key string) error

	// InvalidatePage invalidates a cached page.
	InvalidatePage(ctx context.Context, key string) error
}

// CacheConfig holds configuration for pagination caching.
type CacheConfig struct {
	// CountTTL is the TTL for cached counts (default: 5 minutes).
	CountTTL time.Duration

	// PageTTL is the TTL for cached pages (default: 1 minute).
	PageTTL time.Duration

	// EnableCountCache enables/disables count caching (default: true).
	EnableCountCache bool

	// EnablePageCache enables/disables page caching (default: false).
	EnablePageCache bool

	// KeyPrefix is the prefix for cache keys (default: "pagination:").
	KeyPrefix string
}

// DefaultCacheConfig returns a cache configuration with sensible defaults.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		CountTTL:         5 * time.Minute,
		PageTTL:          1 * time.Minute,
		EnableCountCache: true,
		EnablePageCache:  false, // Page caching disabled by default (can be memory intensive)
		KeyPrefix:        "pagination:",
	}
}

// generateCacheKey generates a cache key from pagination parameters.
func generateCacheKey(prefix string, params PaginationParams, keyType string) string {
	// Create a hash of the parameters
	keyData := fmt.Sprintf("%d:%d:%s:%s:%v:%s",
		params.Page,
		params.PageSize,
		params.Sort,
		params.Order,
		params.Filters,
		params.Search,
	)

	hash := sha256.Sum256([]byte(keyData))
	hashStr := hex.EncodeToString(hash[:])[:16] // Use first 16 chars

	return fmt.Sprintf("%s%s:%s", prefix, keyType, hashStr)
}

// PaginateQueryWithCache paginates data with caching support.
func PaginateQueryWithCache[T any](
	ctx context.Context,
	params PaginationParams,
	cache Cache,
	cacheConfig CacheConfig,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
) (PaginationResult[T], error) {
	if cache == nil {
		// Fall back to regular pagination if cache is nil
		return PaginateQuery(ctx, params, countFn, fetchFn)
	}

	// Generate cache keys
	countKey := generateCacheKey(cacheConfig.KeyPrefix, params, "count")
	pageKey := generateCacheKey(cacheConfig.KeyPrefix, params, "page")

	// Try to get cached count
	var totalRecords int64
	var countCached bool
	if cacheConfig.EnableCountCache {
		if cached, found := cache.GetCount(ctx, countKey); found {
			totalRecords = cached
			countCached = true
		}
	}

	// Try to get cached page
	var cachedData []byte
	var pageCached bool
	if cacheConfig.EnablePageCache {
		if data, found := cache.GetPage(ctx, pageKey); found {
			cachedData = data
			pageCached = true
		}
	}

	// If both are cached, reconstruct result
	if countCached && pageCached {
		var items []T
		if err := json.Unmarshal(cachedData, &items); err == nil {
			return PaginationResult[T]{
				Data:       items,
				Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
			}, nil
		}
		// If unmarshal fails, fall through to fetch fresh data
	}

	// Fetch count if not cached
	if !countCached {
		count, err := countFn(ctx)
		if err != nil {
			return PaginationResult[T]{}, fmt.Errorf("%w: %v", ErrCountFailed, err)
		}
		totalRecords = count

		// Cache the count
		if cacheConfig.EnableCountCache {
			_ = cache.SetCount(ctx, countKey, totalRecords, cacheConfig.CountTTL)
		}
	}

	// Handle empty result set
	if totalRecords == 0 {
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(params.Page, params.PageSize, 0),
		}, nil
	}

	// Check if page is out of range
	totalPages := int((totalRecords + int64(params.PageSize) - 1) / int64(params.PageSize))
	if params.Page > totalPages {
		return PaginationResult[T]{
			Data:       []T{},
			Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
		}, nil
	}

	// Fetch data if not cached
	var data []T
	if !pageCached {
		fetched, err := fetchFn(ctx, params.Limit, params.Offset)
		if err != nil {
			return PaginationResult[T]{}, fmt.Errorf("%w: %v", ErrFetchFailed, err)
		}
		data = fetched

		// Cache the page
		if cacheConfig.EnablePageCache {
			if jsonData, err := json.Marshal(data); err == nil {
				_ = cache.SetPage(ctx, pageKey, jsonData, cacheConfig.PageTTL)
			}
		}
	} else {
		// Use cached data
		if err := json.Unmarshal(cachedData, &data); err != nil {
			// If unmarshal fails, fetch fresh
			fetched, err := fetchFn(ctx, params.Limit, params.Offset)
			if err != nil {
				return PaginationResult[T]{}, fmt.Errorf("%w: %v", ErrFetchFailed, err)
			}
			data = fetched
		}
	}

	return PaginationResult[T]{
		Data:       data,
		Pagination: NewPaginationMeta(params.Page, params.PageSize, totalRecords),
	}, nil
}

// InvalidateCacheForParams invalidates cache entries for given pagination parameters.
func InvalidateCacheForParams(ctx context.Context, cache Cache, cacheConfig CacheConfig, params PaginationParams) error {
	if cache == nil {
		return nil
	}

	countKey := generateCacheKey(cacheConfig.KeyPrefix, params, "count")
	pageKey := generateCacheKey(cacheConfig.KeyPrefix, params, "page")

	_ = cache.InvalidateCount(ctx, countKey)
	_ = cache.InvalidatePage(ctx, pageKey)

	return nil
}


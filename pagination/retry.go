package pagination

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig holds configuration for retry logic.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts (default: 3).
	MaxAttempts int

	// InitialBackoff is the initial backoff duration (default: 100ms).
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration (default: 5s).
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff (default: 2.0).
	BackoffMultiplier float64

	// RetryableErrors is a function that determines if an error is retryable.
	// If nil, all errors are considered retryable.
	RetryableErrors func(error) bool
}

// DefaultRetryConfig returns a retry configuration with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:      3,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      5 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableErrors: nil, // Retry all errors by default
	}
}

// PaginateQueryWithRetry paginates data with automatic retry on failure.
//
// Example:
//   result, err := PaginateQueryWithRetry(
//       ctx,
//       params,
//       DefaultRetryConfig(),
//       countFn,
//       fetchFn,
//   )
func PaginateQueryWithRetry[T any](
	ctx context.Context,
	params PaginationParams,
	retryConfig RetryConfig,
	countFn func(context.Context) (int64, error),
	fetchFn func(context.Context, int, int) ([]T, error),
) (PaginationResult[T], error) {
	// Set defaults
	if retryConfig.MaxAttempts < 1 {
		retryConfig.MaxAttempts = 3
	}
	if retryConfig.InitialBackoff < 0 {
		retryConfig.InitialBackoff = 100 * time.Millisecond
	}
	if retryConfig.MaxBackoff < retryConfig.InitialBackoff {
		retryConfig.MaxBackoff = 5 * time.Second
	}
	if retryConfig.BackoffMultiplier < 1.0 {
		retryConfig.BackoffMultiplier = 2.0
	}

	var lastErr error
	backoff := retryConfig.InitialBackoff

	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		// Try to paginate
		result, err := PaginateQuery(ctx, params, countFn, fetchFn)
		if err == nil {
			return result, nil
		}

		// Check if error is retryable
		if retryConfig.RetryableErrors != nil && !retryConfig.RetryableErrors(err) {
			return PaginationResult[T]{}, err
		}

		lastErr = err

		// Don't sleep on last attempt
		if attempt < retryConfig.MaxAttempts-1 {
			// Calculate backoff
			sleepDuration := backoff
			if sleepDuration > retryConfig.MaxBackoff {
				sleepDuration = retryConfig.MaxBackoff
			}

			// Sleep with context cancellation support
			select {
			case <-ctx.Done():
				return PaginationResult[T]{}, ctx.Err()
			case <-time.After(sleepDuration):
			}

			// Increase backoff for next attempt
			backoff = time.Duration(float64(backoff) * retryConfig.BackoffMultiplier)
		}
	}

	return PaginationResult[T]{}, fmt.Errorf("max retry attempts (%d) exceeded: %w", retryConfig.MaxAttempts, lastErr)
}


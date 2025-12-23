package pagination

import "errors"

var (
	// ErrInvalidPage is returned when the page number is invalid.
	ErrInvalidPage = errors.New("invalid page number")

	// ErrInvalidPageSize is returned when the page size is invalid.
	ErrInvalidPageSize = errors.New("invalid page size")

	// ErrInvalidCursor is returned when the cursor is malformed or invalid.
	ErrInvalidCursor = errors.New("invalid cursor")

	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid pagination configuration")

	// ErrCountFailed is returned when the count operation fails.
	ErrCountFailed = errors.New("failed to count records")

	// ErrFetchFailed is returned when the fetch operation fails.
	ErrFetchFailed = errors.New("failed to fetch records")

	// ErrEmptyResult is returned when attempting to paginate an empty dataset.
	ErrEmptyResult = errors.New("empty result set")
)


package export

import (
	"encoding/json"
	"fmt"
	"io"
	"your-module/pagination"
)

// ExportToJSON exports paginated data to JSON format.
//
// Example:
//   err := ExportToJSON(result, os.Stdout)
func ExportToJSON[T any](
	result pagination.PaginationResult[T],
	writer io.Writer,
	includeMetadata bool,
) error {
	if writer == nil {
		return fmt.Errorf("writer cannot be nil")
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if includeMetadata {
		// Export with metadata
		return encoder.Encode(result)
	}

	// Export only data array
	return encoder.Encode(result.Data)
}

// ExportToJSONL exports paginated data to JSON Lines format (one JSON object per line).
//
// Example:
//   err := ExportToJSONL(result, os.Stdout)
func ExportToJSONL[T any](
	result pagination.PaginationResult[T],
	writer io.Writer,
) error {
	if writer == nil {
		return fmt.Errorf("writer cannot be nil")
	}

	encoder := json.NewEncoder(writer)

	for _, item := range result.Data {
		if err := encoder.Encode(item); err != nil {
			return fmt.Errorf("failed to encode item: %w", err)
		}
	}

	return nil
}


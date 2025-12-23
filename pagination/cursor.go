package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CursorParams holds cursor-based pagination parameters.
type CursorParams struct {
	// Cursor is the encoded cursor string (base64 JSON).
	Cursor string

	// Limit is the maximum number of items to return.
	Limit int

	// Direction indicates pagination direction: "next" or "prev".
	Direction string
}

// CursorResult represents a cursor-based pagination response.
type CursorResult[T any] struct {
	// Data contains the paginated items.
	Data []T `json:"data"`

	// NextCursor is the cursor for the next page, or empty if no next page.
	NextCursor string `json:"next_cursor,omitempty"`

	// PrevCursor is the cursor for the previous page, or empty if no previous page.
	PrevCursor string `json:"prev_cursor,omitempty"`

	// HasNext indicates if there is a next page.
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there is a previous page.
	HasPrev bool `json:"has_prev"`
}

// CursorValue holds the decoded cursor values.
// This is used internally for cursor encoding/decoding.
type CursorValue struct {
	// Position holds the position values (e.g., ID, timestamp, composite key).
	Position map[string]interface{} `json:"position"`

	// Direction indicates the direction of pagination.
	Direction string `json:"direction"`
}

// EncodeCursor encodes cursor values into a base64 JSON string.
// This creates a cursor that can be safely passed in URLs.
func EncodeCursor(position map[string]interface{}, direction string) (string, error) {
	cursor := CursorValue{
		Position:  position,
		Direction: direction,
	}

	jsonData, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	return base64.URLEncoding.EncodeToString(jsonData), nil
}

// DecodeCursor decodes a base64 JSON cursor string into cursor values.
func DecodeCursor(cursorStr string) (map[string]interface{}, string, error) {
	if cursorStr == "" {
		return nil, "", ErrInvalidCursor
	}

	jsonData, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid base64 encoding: %v", ErrInvalidCursor, err)
	}

	var cursor CursorValue
	if err := json.Unmarshal(jsonData, &cursor); err != nil {
		return nil, "", fmt.Errorf("%w: invalid JSON: %v", ErrInvalidCursor, err)
	}

	return cursor.Position, cursor.Direction, nil
}

// PaginateCursor paginates a slice using cursor-based pagination.
// This is useful for large datasets where offset pagination becomes slow.
//
// The cursorExtractor function should extract the cursor value from each item.
// Typically, this would be the item's ID or a composite key.
//
// Example:
//   items := []User{...}
//   cursorParams := CursorParams{Cursor: "...", Limit: 20, Direction: "next"}
//   result := PaginateCursor(items, cursorParams, func(u User) map[string]interface{} {
//       return map[string]interface{}{"id": u.ID}
//   })
func PaginateCursor[T any](
	items []T,
	cursor CursorParams,
	cursorExtractor func(T) map[string]interface{},
) CursorResult[T] {
	if len(items) == 0 {
		return CursorResult[T]{
			Data:    []T{},
			HasNext: false,
			HasPrev: false,
		}
	}

	// Validate limit
	if cursor.Limit < 1 {
		cursor.Limit = 20 // Default limit
	}

	var result []T
	var startIdx int
	var hasPrev bool

	// If cursor is provided, find the starting position
	if cursor.Cursor != "" {
		position, direction, err := DecodeCursor(cursor.Cursor)
		if err != nil {
			// Invalid cursor, return empty result
			return CursorResult[T]{
				Data:    []T{},
				HasNext: false,
				HasPrev: false,
			}
		}

		// Find the item matching the cursor position
		for i, item := range items {
			itemCursor := cursorExtractor(item)
			if matchesCursor(itemCursor, position) {
				if direction == "next" {
					startIdx = i + 1
				} else {
					startIdx = i - 1
					if startIdx < 0 {
						startIdx = 0
					}
				}
				hasPrev = startIdx > 0
				break
			}
		}
	}

	// Extract items up to the limit
	endIdx := startIdx + cursor.Limit
	if endIdx > len(items) {
		endIdx = len(items)
	}

	if startIdx < len(items) {
		result = items[startIdx:endIdx]
	}

	// Determine pagination metadata
	hasNext := endIdx < len(items)
	hasPrev = hasPrev || startIdx > 0

	// Generate cursors
	var nextCursor, prevCursor string
	if len(result) > 0 {
		if hasNext {
			lastItem := result[len(result)-1]
			nextPos := cursorExtractor(lastItem)
			nextCursor, _ = EncodeCursor(nextPos, "next")
		}

		if hasPrev && startIdx > 0 {
			firstItem := result[0]
			prevPos := cursorExtractor(firstItem)
			prevCursor, _ = EncodeCursor(prevPos, "prev")
		}
	}

	return CursorResult[T]{
		Data:       result,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// matchesCursor checks if an item's cursor matches the target cursor position.
func matchesCursor(itemCursor, targetCursor map[string]interface{}) bool {
	if len(itemCursor) != len(targetCursor) {
		return false
	}

	for key, targetValue := range targetCursor {
		itemValue, exists := itemCursor[key]
		if !exists {
			return false
		}

		// Simple equality check (for production, you might want more sophisticated comparison)
		if fmt.Sprintf("%v", itemValue) != fmt.Sprintf("%v", targetValue) {
			return false
		}
	}

	return true
}


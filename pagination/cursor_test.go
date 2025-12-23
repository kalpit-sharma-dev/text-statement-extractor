package pagination

import (
	"testing"
)

func TestEncodeDecodeCursor(t *testing.T) {
	tests := []struct {
		name     string
		position map[string]interface{}
		direction string
	}{
		{
			name:      "simple cursor",
			position:  map[string]interface{}{"id": 123},
			direction: "next",
		},
		{
			name:      "composite cursor",
			position:  map[string]interface{}{"id": 123, "timestamp": 1609459200},
			direction: "next",
		},
		{
			name:      "prev direction",
			position:  map[string]interface{}{"id": 456},
			direction: "prev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeCursor(tt.position, tt.direction)
			if err != nil {
				t.Fatalf("failed to encode cursor: %v", err)
			}

			decodedPos, decodedDir, err := DecodeCursor(encoded)
			if err != nil {
				t.Fatalf("failed to decode cursor: %v", err)
			}

			if decodedDir != tt.direction {
				t.Errorf("expected direction %s, got %s", tt.direction, decodedDir)
			}

			if len(decodedPos) != len(tt.position) {
				t.Errorf("expected position length %d, got %d", len(tt.position), len(decodedPos))
			}

			for k, v := range tt.position {
				if decodedPos[k] != v {
					t.Errorf("expected position[%s] = %v, got %v", k, v, decodedPos[k])
				}
			}
		})
	}
}

func TestDecodeInvalidCursor(t *testing.T) {
	tests := []struct {
		name  string
		cursor string
	}{
		{
			name:   "empty cursor",
			cursor: "",
		},
		{
			name:   "invalid base64",
			cursor: "not-base64!!!",
		},
		{
			name:   "invalid JSON",
			cursor: "aW52YWxpZA==", // base64("invalid")
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := DecodeCursor(tt.cursor)
			if err == nil {
				t.Errorf("expected error for invalid cursor, got nil")
			}
		})
	}
}

func TestPaginateCursor(t *testing.T) {
	type Item struct {
		ID        int
		Timestamp int64
	}

	items := make([]Item, 100)
	for i := 0; i < 100; i++ {
		items[i] = Item{ID: i + 1, Timestamp: int64(1609459200 + i)}
	}

	cursorExtractor := func(item Item) map[string]interface{} {
		return map[string]interface{}{
			"id":        item.ID,
			"timestamp": item.Timestamp,
		}
	}

	tests := []struct {
		name          string
		cursor        CursorParams
		expectedCount int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{
			name:          "first page without cursor",
			cursor:        CursorParams{Cursor: "", Limit: 20, Direction: "next"},
			expectedCount: 20,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:          "with valid cursor",
			cursor:        CursorParams{Cursor: "", Limit: 20, Direction: "next"},
			expectedCount: 20,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:          "last page",
			cursor:        CursorParams{Cursor: "", Limit: 20, Direction: "next"},
			expectedCount: 20,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:          "empty items",
			cursor:        CursorParams{Cursor: "", Limit: 20, Direction: "next"},
			expectedCount: 0,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create cursor for middle page test
			if tt.name == "with valid cursor" {
				// Encode cursor for item at index 20
				pos, _ := EncodeCursor(cursorExtractor(items[20]), "next")
				tt.cursor.Cursor = pos
			}

			// Use appropriate items slice
			testItems := items
			if tt.name == "empty items" {
				testItems = []Item{}
			} else if tt.name == "last page" {
				// Use last 20 items
				testItems = items[80:]
			}

			result := PaginateCursor(testItems, tt.cursor, cursorExtractor)

			if len(result.Data) != tt.expectedCount {
				t.Errorf("expected %d items, got %d", tt.expectedCount, len(result.Data))
			}

			if result.HasNext != tt.expectedHasNext {
				t.Errorf("expected has_next %v, got %v", tt.expectedHasNext, result.HasNext)
			}

			if result.HasPrev != tt.expectedHasPrev {
				t.Errorf("expected has_prev %v, got %v", tt.expectedHasPrev, result.HasPrev)
			}
		})
	}
}


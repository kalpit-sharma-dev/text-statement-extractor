package pagination

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// TimePaginationParams holds parameters for time-based pagination.
type TimePaginationParams struct {
	// StartTime is the start of the time range (inclusive).
	StartTime time.Time

	// EndTime is the end of the time range (inclusive).
	EndTime time.Time

	// Limit is the maximum number of items to return.
	Limit int

	// Direction is the pagination direction: "forward" or "backward".
	Direction string

	// TimeField is the field name used for time-based filtering (default: "created_at").
	TimeField string

	// CursorTime is the cursor timestamp for pagination (optional).
	CursorTime *time.Time
}

// TimePaginationResult represents a time-based pagination response.
type TimePaginationResult[T any] struct {
	// Data contains the paginated items.
	Data []T `json:"data"`

	// NextCursorTime is the cursor timestamp for the next page.
	NextCursorTime *time.Time `json:"next_cursor_time,omitempty"`

	// PrevCursorTime is the cursor timestamp for the previous page.
	PrevCursorTime *time.Time `json:"prev_cursor_time,omitempty"`

	// HasNext indicates if there is a next page.
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there is a previous page.
	HasPrev bool `json:"has_prev"`

	// StartTime is the actual start time used for the query.
	StartTime time.Time `json:"start_time"`

	// EndTime is the actual end time used for the query.
	EndTime time.Time `json:"end_time"`
}

// DefaultTimePaginationParams returns default time pagination parameters.
func DefaultTimePaginationParams() TimePaginationParams {
	now := time.Now()
	return TimePaginationParams{
		StartTime: now.AddDate(0, 0, -30), // Last 30 days
		EndTime:   now,
		Limit:     20,
		Direction: "forward",
		TimeField: "created_at",
	}
}

// PaginateByTime paginates time-series data using time ranges.
// This is optimized for logs, events, and time-ordered data.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - params: Time pagination parameters
//   - fetchFn: Function that fetches data within a time range
//
// Example:
//   params := TimePaginationParams{
//       StartTime: time.Now().AddDate(0, 0, -7),
//       EndTime:   time.Now(),
//       Limit:     50,
//       Direction: "forward",
//   }
//   result, err := PaginateByTime(ctx, params, func(ctx context.Context, start, end time.Time, limit int) ([]LogEntry, error) {
//       return repo.FetchLogsByTimeRange(ctx, start, end, limit)
//   })
func PaginateByTime[T any](
	ctx context.Context,
	params TimePaginationParams,
	fetchFn func(context.Context, time.Time, time.Time, int, string) ([]T, error),
	timeExtractor func(T) time.Time,
) (TimePaginationResult[T], error) {
	// Validate limit
	if params.Limit < 1 {
		params.Limit = 20
	}

	// Set default time field
	if params.TimeField == "" {
		params.TimeField = "created_at"
	}

	// Set default direction
	if params.Direction != "backward" {
		params.Direction = "forward"
	}

	// Determine actual time range
	startTime := params.StartTime
	endTime := params.EndTime

	// If cursor is provided, use it for pagination
	if params.CursorTime != nil {
		if params.Direction == "forward" {
			startTime = *params.CursorTime
		} else {
			endTime = *params.CursorTime
		}
	}

	// Validate time range
	if startTime.After(endTime) {
		return TimePaginationResult[T]{}, fmt.Errorf("start_time must be before end_time")
	}

	// Fetch one extra item to determine if there's a next page
	fetchLimit := params.Limit + 1
	data, err := fetchFn(ctx, startTime, endTime, fetchLimit, params.Direction)
	if err != nil {
		return TimePaginationResult[T]{}, fmt.Errorf("failed to fetch time-based data: %w", err)
	}

	// Determine if there are more pages
	hasNext := len(data) > params.Limit
	if hasNext {
		data = data[:params.Limit] // Remove the extra item
	}

	// Extract cursor times
	var nextCursorTime, prevCursorTime *time.Time
	hasPrev := params.CursorTime != nil

	if len(data) > 0 {
		// Next cursor from last item
		if hasNext && timeExtractor != nil {
			lastTime := timeExtractor(data[len(data)-1])
			nextCursorTime = &lastTime
		}

		// Previous cursor from first item
		if hasPrev && timeExtractor != nil {
			firstTime := timeExtractor(data[0])
			prevCursorTime = &firstTime
		}
	}

	return TimePaginationResult[T]{
		Data:          data,
		NextCursorTime: nextCursorTime,
		PrevCursorTime: prevCursorTime,
		HasNext:       hasNext,
		HasPrev:       hasPrev,
		StartTime:     startTime,
		EndTime:       endTime,
	}, nil
}

// ParseTimePaginationParams parses time pagination parameters from HTTP request.
func ParseTimePaginationParams(r *http.Request) TimePaginationParams {
	params := DefaultTimePaginationParams()

	// Parse start_time
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
			params.StartTime = startTime
		}
	}

	// Parse end_time
	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
			params.EndTime = endTime
		}
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	}

	// Parse direction
	if direction := r.URL.Query().Get("direction"); direction != "" {
		if direction == "backward" {
			params.Direction = "backward"
		}
	}

	// Parse cursor_time
	if cursorStr := r.URL.Query().Get("cursor_time"); cursorStr != "" {
		if cursorTime, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			params.CursorTime = &cursorTime
		}
	}

	return params
}


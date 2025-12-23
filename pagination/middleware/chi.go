package middleware

import (
	"context"
	"net/http"

	"your-module/pagination"
)

const PaginationKey = "pagination_params"

// ChiPaginationMiddleware creates a Chi middleware that parses pagination parameters
// and stores them in the request context for use in handlers.
//
// Usage:
//   r.Use(middleware.ChiPaginationMiddleware(pagination.DefaultConfig()))
//   r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
//       params := r.Context().Value(PaginationKey).(pagination.PaginationParams)
//       // Use params...
//   })
func ChiPaginationMiddleware(cfg pagination.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := pagination.ParsePagination(r, cfg)
			ctx := r.Context()
			ctx = context.WithValue(ctx, PaginationKey, params)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetPaginationParams retrieves pagination parameters from Chi request context.
// Returns default params if not found.
func GetPaginationParams(r *http.Request) pagination.PaginationParams {
	if val := r.Context().Value(PaginationKey); val != nil {
		if params, ok := val.(pagination.PaginationParams); ok {
			return params
		}
	}
	return pagination.PaginationParams{
		Page:     pagination.DefaultConfig().DefaultPage,
		PageSize: pagination.DefaultConfig().DefaultPageSize,
		Order:    "asc",
		Filters:  make(map[string]interface{}),
	}
}


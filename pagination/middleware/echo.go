package middleware

import (
	"github.com/labstack/echo/v4"
	"your-module/pagination"
)

const PaginationKey = "pagination_params"

// EchoPaginationMiddleware creates an Echo middleware that parses pagination parameters
// and stores them in the context for use in handlers.
//
// Usage:
//   e.Use(middleware.EchoPaginationMiddleware(pagination.DefaultConfig()))
//   e.GET("/users", func(c echo.Context) error {
//       params := c.Get(PaginationKey).(pagination.PaginationParams)
//       // Use params...
//   })
func EchoPaginationMiddleware(cfg pagination.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			params := pagination.ParsePagination(c.Request(), cfg)
			c.Set(PaginationKey, params)
			return next(c)
		}
	}
}

// GetPaginationParams retrieves pagination parameters from Echo context.
// Returns default params if not found.
func GetPaginationParams(c echo.Context) pagination.PaginationParams {
	if val := c.Get(PaginationKey); val != nil {
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


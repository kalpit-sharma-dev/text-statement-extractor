package middleware

import (
	"github.com/gofiber/fiber/v2"
	"your-module/pagination"
)

const PaginationKey = "pagination_params"

// FiberPaginationMiddleware creates a Fiber middleware that parses pagination parameters
// and stores them in the context for use in handlers.
//
// Usage:
//   app.Use(middleware.FiberPaginationMiddleware(pagination.DefaultConfig()))
//   app.Get("/users", func(c *fiber.Ctx) error {
//       params := c.Locals(PaginationKey).(pagination.PaginationParams)
//       // Use params...
//   })
func FiberPaginationMiddleware(cfg pagination.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := pagination.ParsePagination(c.Request(), cfg)
		c.Locals(PaginationKey, params)
		return c.Next()
	}
}

// GetPaginationParams retrieves pagination parameters from Fiber context.
// Returns default params if not found.
func GetPaginationParams(c *fiber.Ctx) pagination.PaginationParams {
	if val := c.Locals(PaginationKey); val != nil {
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


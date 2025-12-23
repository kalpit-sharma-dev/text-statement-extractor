package middleware

import (
	"github.com/gin-gonic/gin"
	"your-module/pagination"
)

const PaginationKey = "pagination_params"

// GinPaginationMiddleware creates a Gin middleware that parses pagination parameters
// and stores them in the context for use in handlers.
//
// Usage:
//   router.Use(middleware.GinPaginationMiddleware(pagination.DefaultConfig()))
//   router.GET("/users", func(c *gin.Context) {
//       params := c.MustGet(PaginationKey).(pagination.PaginationParams)
//       // Use params...
//   })
func GinPaginationMiddleware(cfg pagination.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := pagination.ParsePagination(c.Request, cfg)
		c.Set(PaginationKey, params)
		c.Next()
	}
}

// GetPaginationParams retrieves pagination parameters from Gin context.
// Returns default params if not found.
func GetPaginationParams(c *gin.Context) pagination.PaginationParams {
	if val, exists := c.Get(PaginationKey); exists {
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


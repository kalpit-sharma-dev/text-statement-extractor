package query_builders

import (
	"github.com/Masterminds/squirrel"
	"your-module/pagination"
)

// ApplyPaginationToSquirrel applies pagination parameters to a Squirrel SelectBuilder.
//
// Example:
//   query := squirrel.Select("*").From("users")
//   query = ApplyPaginationToSquirrel(query, params)
//   sql, args, err := query.ToSql()
func ApplyPaginationToSquirrel(builder squirrel.SelectBuilder, params pagination.PaginationParams) squirrel.SelectBuilder {
	// Apply offset and limit
	if params.Offset > 0 {
		builder = builder.Offset(uint64(params.Offset))
	}
	if params.Limit > 0 {
		builder = builder.Limit(uint64(params.Limit))
	}

	// Apply sorting
	if len(params.SortFields) > 0 {
		// Multi-field sorting
		for _, sortField := range params.SortFields {
			order := sortField.Order
			if order != "desc" {
				order = "asc"
			}
			builder = builder.OrderBy(sortField.Field + " " + order)
		}
	} else if params.Sort != "" {
		// Single field sorting
		order := params.Order
		if order != "desc" {
			order = "asc"
		}
		builder = builder.OrderBy(params.Sort + " " + order)
	}

	// Apply filters
	for key, value := range params.Filters {
		builder = builder.Where(squirrel.Eq{key: value})
	}

	// Apply search
	if params.Search != "" {
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"name": "%" + params.Search + "%"},
			squirrel.Like{"email": "%" + params.Search + "%"},
		})
	}

	return builder
}

// BuildCountQuery builds a COUNT query with filters applied.
func BuildCountQuery(baseQuery squirrel.SelectBuilder, params pagination.PaginationParams) squirrel.SelectBuilder {
	// Convert SELECT to SELECT COUNT(*)
	countQuery := baseQuery.Columns("COUNT(*)")

	// Apply filters
	for key, value := range params.Filters {
		countQuery = countQuery.Where(squirrel.Eq{key: value})
	}

	// Apply search
	if params.Search != "" {
		countQuery = countQuery.Where(squirrel.Or{
			squirrel.Like{"name": "%" + params.Search + "%"},
			squirrel.Like{"email": "%" + params.Search + "%"},
		})
	}

	return countQuery
}


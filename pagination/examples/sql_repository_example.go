package main

import (
	"context"
	"database/sql"
	"fmt"

	"your-module/pagination"
)

// User represents a sample user model.
type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// UserRepository handles user data access with SQL database.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ListUsers returns paginated users from the database.
// This example works with PostgreSQL, MySQL, YugabyteDB, etc.
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
	return pagination.PaginateQuery(
		ctx,
		params,
		// Count function with filters
		func(ctx context.Context) (int64, error) {
			return r.countUsers(ctx, params.Filters)
		},
		// Fetch function with sorting and filters
		func(ctx context.Context, limit, offset int) ([]User, error) {
			return r.fetchUsers(ctx, limit, offset, params.Sort, params.Order, params.Filters)
		},
	)
}

// countUsers returns the total count of users with optional filters.
func (r *UserRepository) countUsers(ctx context.Context, filters map[string]interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM users"
	args := []interface{}{}
	argPos := 1

	// Apply filters
	if len(filters) > 0 {
		query += " WHERE "
		conditions := []string{}
		for key, value := range filters {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argPos))
			args = append(args, value)
			argPos++
		}
		query += fmt.Sprintf(" %s", fmt.Sprintf("%s", conditions[0]))
		if len(conditions) > 1 {
			for i := 1; i < len(conditions); i++ {
				query += fmt.Sprintf(" AND %s", conditions[i])
			}
		}
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// fetchUsers fetches paginated users with sorting and filters.
func (r *UserRepository) fetchUsers(
	ctx context.Context,
	limit, offset int,
	sort, order string,
	filters map[string]interface{},
) ([]User, error) {
	// Build query
	query := "SELECT id, name, email, created_at FROM users"
	args := []interface{}{}
	argPos := 1

	// Apply filters
	if len(filters) > 0 {
		query += " WHERE "
		conditions := []string{}
		for key, value := range filters {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argPos))
			args = append(args, value)
			argPos++
		}
		query += fmt.Sprintf(" %s", conditions[0])
		if len(conditions) > 1 {
			for i := 1; i < len(conditions); i++ {
				query += fmt.Sprintf(" AND %s", conditions[i])
			}
		}
	}

	// Apply sorting
	if sort != "" {
		// Validate sort field to prevent SQL injection
		allowedSorts := map[string]bool{
			"id": true, "name": true, "email": true, "created_at": true,
		}
		if allowedSorts[sort] {
			if order != "desc" {
				order = "asc"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", sort, order)
		}
	} else {
		query += " ORDER BY id ASC"
	}

	// Apply pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// Example usage with pgx (PostgreSQL driver)
/*
import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryPgx struct {
	pool *pgxpool.Pool
}

func (r *UserRepositoryPgx) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
	return pagination.PaginateQuery(
		ctx,
		params,
		func(ctx context.Context) (int64, error) {
			var count int64
			err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
			return count, err
		},
		func(ctx context.Context, limit, offset int) ([]User, error) {
			rows, err := r.pool.Query(ctx,
				"SELECT id, name, email, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2",
				limit, offset,
			)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var users []User
			for rows.Next() {
				var user User
				if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
					return nil, err
				}
				users = append(users, user)
			}
			return users, rows.Err()
		},
	)
}
*/


package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"your-module/pagination"
)

// User represents a sample user model.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserController handles user-related HTTP requests.
type UserController struct {
	service *UserService
}

// NewUserController creates a new user controller.
func NewUserController(service *UserService) *UserController {
	return &UserController{service: service}
}

// ListUsers handles GET /users with pagination support.
func (c *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters from request
	cfg := pagination.DefaultConfig()
	params := pagination.ParsePagination(r, cfg)

	// Get paginated results from service
	result, err := c.service.ListUsers(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// UserService handles user business logic.
type UserService struct {
	repo *UserRepository
}

// NewUserService creates a new user service.
func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ListUsers returns paginated users.
func (s *UserService) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
	return s.repo.ListUsers(ctx, params)
}

// UserRepository handles user data access.
type UserRepository struct {
	// In a real implementation, this would contain database connection
	// For example: db *sql.DB, or mongoClient *mongo.Client, etc.
}

// NewUserRepository creates a new user repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// ListUsers returns paginated users from the database.
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
	// Use PaginateQuery with count and fetch functions
	return pagination.PaginateQuery(
		ctx,
		params,
		// Count function
		r.countUsers,
		// Fetch function
		r.fetchUsers,
	)
}

// countUsers returns the total count of users.
func (r *UserRepository) countUsers(ctx context.Context) (int64, error) {
	// Example: SELECT COUNT(*) FROM users
	// In real implementation, execute actual database query
	return 100, nil // Mock value
}

// fetchUsers fetches paginated users from the database.
func (r *UserRepository) fetchUsers(ctx context.Context, limit, offset int) ([]User, error) {
	// Example: SELECT * FROM users ORDER BY id LIMIT ? OFFSET ?
	// In real implementation, execute actual database query
	users := []User{
		{ID: offset + 1, Name: "User 1", Email: "user1@example.com"},
		{ID: offset + 2, Name: "User 2", Email: "user2@example.com"},
	}
	return users, nil
}

// Example HTTP server setup
func main() {
	repo := NewUserRepository()
	service := NewUserService(repo)
	controller := NewUserController(service)

	http.HandleFunc("/users", controller.ListUsers)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"your-module/pagination"
)

// User represents a sample user model.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserController handles user-related HTTP requests using Gin.
type UserController struct {
	service *UserService
}

// NewUserController creates a new user controller.
func NewUserController(service *UserService) *UserController {
	return &UserController{service: service}
}

// ListUsers handles GET /users with pagination support.
// Gin's *gin.Context.Request is compatible with *http.Request.
func (c *UserController) ListUsers(ctx *gin.Context) {
	// Parse pagination parameters from request
	cfg := pagination.DefaultConfig()
	params := pagination.ParsePagination(ctx.Request, cfg)

	// Get paginated results from service
	result, err := c.service.ListUsers(ctx.Request.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return paginated response
	ctx.JSON(http.StatusOK, result)
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
}

// NewUserRepository creates a new user repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// ListUsers returns paginated users from the database.
func (r *UserRepository) ListUsers(ctx context.Context, params pagination.PaginationParams) (pagination.PaginationResult[User], error) {
	return pagination.PaginateQuery(
		ctx,
		params,
		r.countUsers,
		r.fetchUsers,
	)
}

// countUsers returns the total count of users.
func (r *UserRepository) countUsers(ctx context.Context) (int64, error) {
	// Example: SELECT COUNT(*) FROM users WHERE ...
	// Apply filters from params.Filters if needed
	return 100, nil
}

// fetchUsers fetches paginated users from the database.
func (r *UserRepository) fetchUsers(ctx context.Context, limit, offset int) ([]User, error) {
	// Example: SELECT * FROM users ORDER BY ? LIMIT ? OFFSET ?
	// Apply sorting from params.Sort and params.Order
	users := []User{
		{ID: offset + 1, Name: "User 1", Email: "user1@example.com"},
		{ID: offset + 2, Name: "User 2", Email: "user2@example.com"},
	}
	return users, nil
}

// Example Gin server setup
func main() {
	router := gin.Default()

	repo := NewUserRepository()
	service := NewUserService(repo)
	controller := NewUserController(service)

	router.GET("/users", controller.ListUsers)

	router.Run(":8080")
}


package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"your-module/pagination"
)

// Product represents a sample product model.
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// ProductService handles product business logic with in-memory data.
type ProductService struct {
	products []Product
}

// NewProductService creates a new product service with sample data.
func NewProductService() *ProductService {
	// Sample data
	products := make([]Product, 0, 100)
	for i := 1; i <= 100; i++ {
		products = append(products, Product{
			ID:    i,
			Name:  fmt.Sprintf("Product %d", i),
			Price: float64(i * 10),
		})
	}

	return &ProductService{products: products}
}

// ListProducts returns paginated products from in-memory slice.
func (s *ProductService) ListProducts(r *http.Request) pagination.PaginationResult[Product] {
	// Parse pagination parameters
	cfg := pagination.DefaultConfig()
	params := pagination.ParsePagination(r, cfg)

	// Paginate the in-memory slice
	return pagination.PaginateSlice(s.products, params)
}

// Example HTTP handler
func ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	service := NewProductService()
	result := service.ListProducts(r)

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Example with filtering and sorting
func (s *ProductService) ListProductsFiltered(r *http.Request) pagination.PaginationResult[Product] {
	cfg := pagination.DefaultConfig()
	params := pagination.ParsePagination(r, cfg)

	// Apply filters to in-memory data
	filtered := s.products
	if minPrice, ok := params.Filters["min_price"].(string); ok {
		// Filter by minimum price
		// ... implementation
	}

	// Apply sorting
	if params.Sort == "price" {
		// Sort by price
		// ... implementation
	}

	// Paginate filtered and sorted results
	return pagination.PaginateSlice(filtered, params)
}


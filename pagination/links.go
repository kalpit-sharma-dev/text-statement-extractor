package pagination

import (
	"fmt"
	"net/url"
)

// PaginationLinks contains HATEOAS-style pagination links.
type PaginationLinks struct {
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Self  string `json:"self,omitempty"`
}

// PaginationResultWithLinks extends PaginationResult with HATEOAS links.
type PaginationResultWithLinks[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Links      PaginationLinks `json:"links,omitempty"`
}

// WithLinks adds HATEOAS pagination links to a PaginationResult.
// baseURL should be the base URL of the API endpoint (e.g., "https://api.example.com/users").
// queryParams should contain the original query parameters from the request.
func WithLinks[T any](
	result PaginationResult[T],
	baseURL string,
	queryParams url.Values,
) PaginationResultWithLinks[T] {
	links := PaginationLinks{}

	// Build base URL with existing query params (excluding pagination params)
	base := buildBaseURL(baseURL, queryParams)

	// Self link (current page)
	links.Self = buildPageURL(base, result.Pagination.Page, result.Pagination.PageSize)

	// First page
	if result.Pagination.TotalPages > 0 {
		links.First = buildPageURL(base, 1, result.Pagination.PageSize)
	}

	// Last page
	if result.Pagination.TotalPages > 0 {
		links.Last = buildPageURL(base, result.Pagination.TotalPages, result.Pagination.PageSize)
	}

	// Next page
	if result.Pagination.HasNext && result.Pagination.NextPage != nil {
		links.Next = buildPageURL(base, *result.Pagination.NextPage, result.Pagination.PageSize)
	}

	// Previous page
	if result.Pagination.HasPrev && result.Pagination.PrevPage != nil {
		links.Prev = buildPageURL(base, *result.Pagination.PrevPage, result.Pagination.PageSize)
	}

	return PaginationResultWithLinks[T]{
		Data:       result.Data,
		Pagination: result.Pagination,
		Links:      links,
	}
}

// buildBaseURL constructs the base URL with non-pagination query parameters.
func buildBaseURL(baseURL string, queryParams url.Values) string {
	if len(queryParams) == 0 {
		return baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	// Copy non-pagination query parameters
	q := url.Values{}
	for key, values := range queryParams {
		if key != "page" && key != "page_size" {
			for _, v := range values {
				q.Add(key, v)
			}
		}
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// buildPageURL constructs a URL with pagination parameters.
func buildPageURL(baseURL string, page, pageSize int) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	u.RawQuery = q.Encode()

	return u.String()
}


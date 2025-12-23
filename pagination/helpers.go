package pagination

// CalculateOffset calculates the offset for pagination based on page number and page size.
// Formula: offset = (page - 1) * pageSize
//
// Example:
//   offset := CalculateOffset(2, 20) // Returns 20
func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	offset := (page - 1) * pageSize
	if offset < 0 {
		return 0
	}
	return offset
}

// CalculateTotalPages calculates the total number of pages based on total records and page size.
// Formula: totalPages = ceil(totalRecords / pageSize)
//
// Example:
//   totalPages := CalculateTotalPages(100, 20) // Returns 5
//   totalPages := CalculateTotalPages(101, 20) // Returns 6
func CalculateTotalPages(totalRecords int64, pageSize int) int {
	if pageSize < 1 {
		return 1
	}
	if totalRecords == 0 {
		return 1
	}
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		return 1
	}
	return totalPages
}

// HasNextPage determines if there is a next page.
//
// Example:
//   hasNext := HasNextPage(2, 5, 20) // Returns true (page 2 of 5)
func HasNextPage(currentPage, totalPages int) bool {
	return currentPage < totalPages
}

// HasPrevPage determines if there is a previous page.
//
// Example:
//   hasPrev := HasPrevPage(2, 5) // Returns true
func HasPrevPage(currentPage int) bool {
	return currentPage > 1
}

// GetNextPage returns the next page number, or 0 if there is no next page.
//
// Example:
//   nextPage := GetNextPage(2, 5) // Returns 3
//   nextPage := GetNextPage(5, 5) // Returns 0
func GetNextPage(currentPage, totalPages int) int {
	if currentPage < totalPages {
		return currentPage + 1
	}
	return 0
}

// GetPrevPage returns the previous page number, or 0 if there is no previous page.
//
// Example:
//   prevPage := GetPrevPage(2, 5) // Returns 1
//   prevPage := GetPrevPage(1, 5) // Returns 0
func GetPrevPage(currentPage int) int {
	if currentPage > 1 {
		return currentPage - 1
	}
	return 0
}

// ValidatePage validates and normalizes a page number.
// Returns 1 if page is less than 1.
//
// Example:
//   page := ValidatePage(0)  // Returns 1
//   page := ValidatePage(-5) // Returns 1
//   page := ValidatePage(5)  // Returns 5
func ValidatePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

// ValidatePageSize validates and normalizes a page size.
// Returns defaultSize if pageSize is less than minSize.
// Returns maxSize if pageSize is greater than maxSize.
//
// Example:
//   size := ValidatePageSize(0, 1, 20, 100)   // Returns 20
//   size := ValidatePageSize(200, 1, 20, 100) // Returns 100
//   size := ValidatePageSize(50, 1, 20, 100)  // Returns 50
func ValidatePageSize(pageSize, minSize, defaultSize, maxSize int) int {
	if pageSize < minSize {
		return defaultSize
	}
	if pageSize > maxSize {
		return maxSize
	}
	return pageSize
}


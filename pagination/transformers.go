package pagination

// Transformer is a function that transforms an item.
type Transformer[T any] func(T) T

// WithTransformer applies a transformer function to all items in a pagination result.
// This is useful for data masking, formatting, or custom transformations.
//
// Example:
//   maskedResult := WithTransformer(result, func(u User) User {
//       u.Email = maskEmail(u.Email)
//       return u
//   })
func WithTransformer[T any](
	result PaginationResult[T],
	transformer Transformer[T],
) PaginationResult[T] {
	if transformer == nil {
		return result
	}

	transformedData := make([]T, len(result.Data))
	for i, item := range result.Data {
		transformedData[i] = transformer(item)
	}

	return PaginationResult[T]{
		Data:       transformedData,
		Pagination: result.Pagination,
	}
}

// WithTransformers applies multiple transformers in sequence.
func WithTransformers[T any](
	result PaginationResult[T],
	transformers ...Transformer[T],
) PaginationResult[T] {
	if len(transformers) == 0 {
		return result
	}

	current := result
	for _, transformer := range transformers {
		current = WithTransformer(current, transformer)
	}

	return current
}

// MaskFields masks specified fields in the result.
// This is a convenience function for data privacy.
func MaskFields[T any](
	result PaginationResult[T],
	fieldMasker func(T) T,
) PaginationResult[T] {
	return WithTransformer(result, fieldMasker)
}

// FormatFields formats specified fields in the result.
// This is a convenience function for data formatting.
func FormatFields[T any](
	result PaginationResult[T],
	formatter func(T) T,
) PaginationResult[T] {
	return WithTransformer(result, formatter)
}

// FilterFields filters out items that don't match a predicate.
// Note: This changes the total count, so use with caution.
func FilterFields[T any](
	result PaginationResult[T],
	predicate func(T) bool,
) PaginationResult[T] {
	if predicate == nil {
		return result
	}

	filteredData := make([]T, 0, len(result.Data))
	for _, item := range result.Data {
		if predicate(item) {
			filteredData = append(filteredData, item)
		}
	}

	// Update total records (approximate - actual count may differ)
	newTotal := int64(len(filteredData))
	if len(result.Data) > 0 {
		// Estimate based on filtered ratio
		filterRatio := float64(len(filteredData)) / float64(len(result.Data))
		newTotal = int64(float64(result.Pagination.TotalRecords) * filterRatio)
	}

	return PaginationResult[T]{
		Data: filteredData,
		Pagination: PaginationMeta{
			Page:         result.Pagination.Page,
			PageSize:     result.Pagination.PageSize,
			TotalRecords: newTotal,
			TotalPages:   CalculateTotalPages(newTotal, result.Pagination.PageSize),
			HasNext:      result.Pagination.HasNext,
			HasPrev:      result.Pagination.HasPrev,
			NextPage:     result.Pagination.NextPage,
			PrevPage:     result.Pagination.PrevPage,
		},
	}
}


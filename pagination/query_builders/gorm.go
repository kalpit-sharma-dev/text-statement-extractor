package query_builders

import (
	"gorm.io/gorm"
	"your-module/pagination"
)

// ApplyPaginationToGORM applies pagination parameters to a GORM query.
//
// Example:
//   var users []User
//   query := db.Model(&User{})
//   query = ApplyPaginationToGORM(query, params)
//   query.Find(&users)
func ApplyPaginationToGORM(db *gorm.DB, params pagination.PaginationParams) *gorm.DB {
	// Apply offset and limit
	if params.Offset > 0 {
		db = db.Offset(params.Offset)
	}
	if params.Limit > 0 {
		db = db.Limit(params.Limit)
	}

	// Apply sorting
	if len(params.SortFields) > 0 {
		// Multi-field sorting
		for _, sortField := range params.SortFields {
			order := sortField.Order
			if order != "desc" {
				order = "asc"
			}
			db = db.Order(sortField.Field + " " + order)
		}
	} else if params.Sort != "" {
		// Single field sorting
		order := params.Order
		if order != "desc" {
			order = "asc"
		}
		db = db.Order(params.Sort + " " + order)
	}

	// Apply filters
	for key, value := range params.Filters {
		db = db.Where(key+" = ?", value)
	}

	// Apply search (if supported by your model)
	if params.Search != "" {
		// This is a simple example - adjust based on your search requirements
		db = db.Where("name LIKE ? OR email LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	return db
}

// CountWithGORM counts records using GORM with filters applied.
//
// Example:
//   var count int64
//   query := db.Model(&User{})
//   query = ApplyFiltersToGORM(query, params)
//   query.Count(&count)
func CountWithGORM(db *gorm.DB, params pagination.PaginationParams, model interface{}) (int64, error) {
	var count int64
	query := db.Model(model)

	// Apply filters
	for key, value := range params.Filters {
		query = query.Where(key+" = ?", value)
	}

	// Apply search
	if params.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}


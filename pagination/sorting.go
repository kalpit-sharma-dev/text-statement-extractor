package pagination

import (
	"strings"
)

// SortField represents a single sort field with its order.
type SortField struct {
	// Field is the field name to sort by.
	Field string

	// Order is the sort order: "asc" or "desc".
	Order string
}

// ParseSortFields parses multiple sort fields from a comma-separated string.
// Format: "field1,field2,field3" with optional order: "field1:asc,field2:desc"
// Or use separate sort and order params: "field1,field2" and "asc,desc"
func ParseSortFields(sortStr, orderStr string) []SortField {
	if sortStr == "" {
		return nil
	}

	fields := strings.Split(sortStr, ",")
	orders := strings.Split(orderStr, ",")

	sortFields := make([]SortField, 0, len(fields))
	defaultOrder := "asc"

	for i, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		// Check if order is embedded in field (e.g., "name:asc")
		var order string
		if strings.Contains(field, ":") {
			parts := strings.Split(field, ":")
			if len(parts) == 2 {
				field = strings.TrimSpace(parts[0])
				order = strings.ToLower(strings.TrimSpace(parts[1]))
				if order != "asc" && order != "desc" {
					order = defaultOrder
				}
			} else {
				order = defaultOrder
			}
		} else {
			// Use order from orderStr if available
			if i < len(orders) {
				order = strings.ToLower(strings.TrimSpace(orders[i]))
				if order != "asc" && order != "desc" {
					order = defaultOrder
				}
			} else {
				order = defaultOrder
			}
		}

		sortFields = append(sortFields, SortField{
			Field: field,
			Order: order,
		})
	}

	return sortFields
}

// ToSQLOrderBy converts sort fields to SQL ORDER BY clause.
// Example: []SortField{{Field: "name", Order: "asc"}, {Field: "created_at", Order: "desc"}}
// Returns: "ORDER BY name ASC, created_at DESC"
func (sf []SortField) ToSQLOrderBy() string {
	if len(sf) == 0 {
		return ""
	}

	clauses := make([]string, len(sf))
	for i, field := range sf {
		// Basic SQL injection prevention - only allow alphanumeric, underscore, dot
		fieldName := sanitizeFieldName(field.Field)
		order := strings.ToUpper(field.Order)
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		clauses[i] = fieldName + " " + order
	}

	return "ORDER BY " + strings.Join(clauses, ", ")
}

// sanitizeFieldName removes potentially dangerous characters from field names.
// In production, you should use a whitelist of allowed field names.
func sanitizeFieldName(field string) string {
	// Remove any characters that aren't alphanumeric, underscore, or dot
	var result strings.Builder
	for _, r := range field {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}
	return result.String()
}


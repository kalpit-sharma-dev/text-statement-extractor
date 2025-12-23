package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"your-module/pagination"
)

// ExportToCSV exports paginated data to CSV format.
//
// Example:
//   writer := csv.NewWriter(os.Stdout)
//   err := ExportToCSV(result, writer, []string{"ID", "Name", "Email"})
func ExportToCSV[T any](
	result pagination.PaginationResult[T],
	writer *csv.Writer,
	headers []string,
	fieldExtractor func(T) []string,
) error {
	if writer == nil {
		return fmt.Errorf("csv writer cannot be nil")
	}

	// Write headers
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	// Write data rows
	for _, item := range result.Data {
		var row []string
		if fieldExtractor != nil {
			row = fieldExtractor(item)
		} else {
			// Default: use reflection to extract fields
			row = extractFieldsByReflection(item)
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}

// extractFieldsByReflection extracts field values from a struct using reflection.
func extractFieldsByReflection(item interface{}) []string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return []string{fmt.Sprintf("%v", item)}
	}

	var values []string
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Skip fields with json:"-" tag
		if tag := fieldType.Tag.Get("json"); tag == "-" {
			continue
		}

		var value string
		switch field.Kind() {
		case reflect.String:
			value = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value = strconv.FormatUint(field.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			value = strconv.FormatFloat(field.Float(), 'f', -1, 64)
		case reflect.Bool:
			value = strconv.FormatBool(field.Bool())
		default:
			value = fmt.Sprintf("%v", field.Interface())
		}

		values = append(values, value)
	}

	return values
}


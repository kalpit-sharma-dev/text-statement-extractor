package utils

import (
	"time"
)

// ParseDate parses date in DD/MM/YY format
func ParseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/06",   // DD/MM/YY
		"02/01/2006", // DD/MM/YYYY
		"2006-01-02", // YYYY-MM-DD
		"02-01-2006", // DD-MM-YYYY
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

// FormatDate formats date to string
func FormatDate(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}

	switch format {
	case "YYYY-MM-DD":
		return t.Format("2006-01-02")
	case "DD/MM/YYYY":
		return t.Format("02/01/2006")
	case "Month":
		return t.Format("Jan")
	case "MonthFull":
		return t.Format("January")
	default:
		return t.Format("02/01/2006")
	}
}

// GetMonthName returns month name from date string
func GetMonthName(dateStr string) string {
	t, err := ParseDate(dateStr)
	if err != nil {
		return ""
	}
	return t.Format("Jan")
}

// GetYear returns year from date string
func GetYear(dateStr string) string {
	t, err := ParseDate(dateStr)
	if err != nil {
		return ""
	}
	return t.Format("2006")
}

// IsSameMonth checks if two dates are in the same month
func IsSameMonth(date1, date2 string) bool {
	t1, err1 := ParseDate(date1)
	t2, err2 := ParseDate(date2)
	if err1 != nil || err2 != nil {
		return false
	}
	return t1.Year() == t2.Year() && t1.Month() == t2.Month()
}

// GetDaysInMonth returns number of days in a month
func GetDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}


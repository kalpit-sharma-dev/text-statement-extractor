package utils

import (
	"sort"
)

// SortByAmountDesc sorts transactions by amount descending
func SortByAmountDesc(transactions []interface{}, getAmount func(interface{}) float64) {
	sort.Slice(transactions, func(i, j int) bool {
		return getAmount(transactions[i]) > getAmount(transactions[j])
	})
}

// SortByDateDesc sorts transactions by date descending
func SortByDateDesc(transactions []interface{}, getDate func(interface{}) string) {
	sort.Slice(transactions, func(i, j int) bool {
		dateI, _ := ParseDate(getDate(transactions[i]))
		dateJ, _ := ParseDate(getDate(transactions[j]))
		return dateI.After(dateJ)
	})
}

// TopN returns top N items from a sorted slice
func TopN(items []interface{}, n int) []interface{} {
	if n > len(items) {
		n = len(items)
	}
	return items[:n]
}


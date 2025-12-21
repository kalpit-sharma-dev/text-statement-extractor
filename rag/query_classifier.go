package rag

import (
	"strings"
)

// QueryType represents the type of query
type QueryType string

const (
	QueryTypeCalculation QueryType = "calculation" // "What is my total expense?"
	QueryTypeComparison  QueryType = "comparison"   // "Which month had more expenses?"
	QueryTypeListing     QueryType = "listing"     // "What are my top expenses?"
	QueryTypeSpecific    QueryType = "specific"     // "What was my expense on X?"
	QueryTypeTrend       QueryType = "trend"        // "How did my expenses change?"
)

// QueryClassifier classifies queries for optimized retrieval
type QueryClassifier struct{}

// NewQueryClassifier creates a new query classifier
func NewQueryClassifier() *QueryClassifier {
	return &QueryClassifier{}
}

// ClassifyQuery determines the type of query
func (qc *QueryClassifier) ClassifyQuery(query string) QueryType {
	queryLower := strings.ToLower(query)

	// Calculation indicators
	calcKeywords := []string{"what is", "calculate", "total", "sum", "how much", "net", "what was", "what were"}
	for _, kw := range calcKeywords {
		if strings.Contains(queryLower, kw) {
			return QueryTypeCalculation
		}
	}

	// Comparison indicators
	compKeywords := []string{"which", "compare", "more", "less", "higher", "lower", "better", "worse", "greater", "smaller"}
	for _, kw := range compKeywords {
		if strings.Contains(queryLower, kw) {
			return QueryTypeComparison
		}
	}

	// Listing indicators
	listKeywords := []string{"list", "what are", "show me", "all", "top", "bottom", "name", "tell me"}
	for _, kw := range listKeywords {
		if strings.Contains(queryLower, kw) {
			return QueryTypeListing
		}
	}

	// Trend indicators
	trendKeywords := []string{"trend", "change", "increase", "decrease", "over time", "pattern", "fluctuation", "variation"}
	for _, kw := range trendKeywords {
		if strings.Contains(queryLower, kw) {
			return QueryTypeTrend
		}
	}

	return QueryTypeSpecific // Default
}

// GetOptimalChunkTypes returns preferred chunk types for query type
func (qc *QueryClassifier) GetOptimalChunkTypes(queryType QueryType) []string {
	switch queryType {
	case QueryTypeCalculation:
		return []string{"account_summary", "transaction_breakdown", "monthly_summary"}
	case QueryTypeComparison:
		return []string{"monthly_summary", "category_summary", "transactions"}
	case QueryTypeListing:
		return []string{"top_expenses", "top_beneficiaries", "category_summary"}
	case QueryTypeTrend:
		return []string{"monthly_summary", "transactions", "category_summary"}
	default:
		return []string{"account_summary", "transactions"}
	}
}

// GetOptimalTopK returns optimal TopK for query type
func (qc *QueryClassifier) GetOptimalTopK(queryType QueryType, defaultTopK int) int {
	switch queryType {
	case QueryTypeCalculation, QueryTypeSpecific:
		return 3 // Fewer chunks, more focused
	case QueryTypeComparison, QueryTypeTrend:
		return 7 // More chunks for comparison
	case QueryTypeListing:
		return 5 // Default
	default:
		return defaultTopK
	}
}


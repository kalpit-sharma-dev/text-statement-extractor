package rag

import (
	"strings"
)

// QueryEnhancer enhances queries for better retrieval
type QueryEnhancer struct {
	config *Config
}

// NewQueryEnhancer creates a new query enhancer
func NewQueryEnhancer(config *Config) *QueryEnhancer {
	return &QueryEnhancer{
		config: config,
	}
}

// EnhanceQuery expands and improves the query
func (qe *QueryEnhancer) EnhanceQuery(query string) string {
	// Convert to lowercase for processing
	lowerQuery := strings.ToLower(query)

	// Financial domain-specific expansions
	expansions := map[string][]string{
		"total expense": {"total expense", "total spending", "total expenditure", "sum of expenses", "expense total"},
		"income":       {"income", "salary", "credit", "deposit", "inflow", "earnings"},
		"transaction":  {"transaction", "payment", "transfer", "debit", "credit", "payment"},
		"category":     {"category", "type", "classification", "group", "spending category"},
		"merchant":     {"merchant", "vendor", "shop", "store", "beneficiary", "payee"},
		"month":        {"month", "monthly", "period", "monthly summary"},
		"highest":      {"highest", "maximum", "max", "largest", "biggest", "top"},
		"lowest":       {"lowest", "minimum", "min", "smallest", "bottom"},
		"balance":      {"balance", "account balance", "closing balance", "opening balance"},
		"expense":      {"expense", "spending", "expenditure", "outflow", "debit"},
		"investment":   {"investment", "savings", "deposit", "fixed deposit"},
		"upi":          {"upi", "unified payments interface", "mobile payment"},
		"imps":         {"imps", "immediate payment service", "bank transfer"},
		"neft":         {"neft", "national electronic funds transfer"},
	}

	// Find matching expansions
	enhancedTerms := []string{query} // Always include original
	for key, synonyms := range expansions {
		if strings.Contains(lowerQuery, key) {
			enhancedTerms = append(enhancedTerms, synonyms...)
		}
	}

	// Build enhanced query (limit to 5 additional terms to avoid dilution)
	enhancedQuery := query
	if len(enhancedTerms) > 1 {
		maxTerms := 6 // Original + 5 synonyms
		if len(enhancedTerms) > maxTerms {
			enhancedTerms = enhancedTerms[:maxTerms]
		}
		enhancedQuery = strings.Join(enhancedTerms, " ")
	}

	return enhancedQuery
}

// EnhanceQueryWithContext adds context-aware enhancements
func (qe *QueryEnhancer) EnhanceQueryWithContext(query string, conversationHistory []string) string {
	enhanced := qe.EnhanceQuery(query)

	// Add context from conversation history
	if len(conversationHistory) > 0 {
		// Extract key terms from recent conversation (last 2 messages)
		startIdx := 0
		if len(conversationHistory) > 2 {
			startIdx = len(conversationHistory) - 2
		}
		recentContext := strings.Join(conversationHistory[startIdx:], " ")
		
		// Extract important terms from context (numbers, amounts, categories)
		contextTerms := qe.extractImportantTerms(recentContext)
		if len(contextTerms) > 0 {
			enhanced = enhanced + " " + strings.Join(contextTerms[:min(3, len(contextTerms))], " ")
		}
	}

	return enhanced
}

// extractImportantTerms extracts important terms (numbers, amounts, categories)
func (qe *QueryEnhancer) extractImportantTerms(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	important := []string{}

	// Look for currency amounts, numbers, and category names
	for _, word := range words {
		// Skip common words
		if isCommonWord(word) {
			continue
		}
		
		// Include if it looks like a number or amount
		if strings.Contains(word, "₹") || strings.Contains(word, ",") {
			important = append(important, word)
		}
		
		// Include category-like words (longer words are more likely to be categories)
		if len(word) > 5 {
			important = append(important, word)
		}
	}

	return important
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "this": true,
		"that": true, "these": true, "those": true, "what": true, "which": true,
		"who": true, "whom": true, "where": true, "when": true, "why": true,
		"how": true, "my": true, "your": true, "his": true, "her": true,
		"its": true, "our": true, "their": true, "and": true, "or": true,
		"but": true, "if": true, "then": true, "else": true, "for": true,
		"with": true, "from": true, "to": true, "of": true, "in": true,
		"on": true, "at": true, "by": true, "about": true, "into": true,
	}
	return commonWords[word]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


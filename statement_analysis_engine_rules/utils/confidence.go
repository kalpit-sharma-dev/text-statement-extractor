package utils

import "strings"

// RuleVersion is the current version of classification rules
const RuleVersion = "v1.3.0"

// CalculateConfidence calculates confidence score for classification
// Returns value between 0.0 and 1.0
func CalculateConfidence(
	matchedKeywords []string,
	hasGateway bool,
	hasMerchant bool,
	hasAmountPattern bool,
	hasRecurrencePattern bool,
) float64 {
	confidence := 0.0
	
	// Base confidence from keywords (each keyword adds 0.15, max 0.6)
	keywordScore := float64(len(matchedKeywords)) * 0.15
	if keywordScore > 0.6 {
		keywordScore = 0.6
	}
	confidence += keywordScore
	
	// Gateway detection adds confidence (0.2)
	if hasGateway {
		confidence += 0.2
	}
	
	// Merchant detection adds confidence (0.15)
	if hasMerchant {
		confidence += 0.15
	}
	
	// Amount pattern match adds confidence (0.05)
	if hasAmountPattern {
		confidence += 0.05
	}
	
	// Recurrence pattern adds confidence (0.1)
	if hasRecurrencePattern {
		confidence += 0.1
	}
	
	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	// Minimum confidence is 0.3 if we have at least one keyword
	if len(matchedKeywords) > 0 && confidence < 0.3 {
		confidence = 0.3
	}
	
	return confidence
}

// DetectAmountPattern detects if amount matches common patterns
func DetectAmountPattern(amount float64) (pattern string, matched bool) {
	if amount <= 0 {
		return "", false
	}
	
	// Charges (₹10-₹50)
	if amount >= 10 && amount <= 50 {
		return "CHARGE", true
	}
	
	// Food delivery (₹200-₹600)
	if amount >= 200 && amount <= 600 {
		return "FOOD", true
	}
	
	// Utilities (₹800-₹2,500)
	if amount >= 800 && amount <= 2500 {
		return "UTILITY", true
	}
	
	// Round amounts (likely investments or large purchases)
	if amount >= 1000 && int(amount)%1000 == 0 {
		return "ROUND", true
	}
	
	return "", false
}

// ExtractMatchedKeywords extracts keywords that matched from narration
func ExtractMatchedKeywords(narration string, patterns []string) []string {
	upper := strings.ToUpper(narration)
	matched := make([]string, 0)
	
	for _, pattern := range patterns {
		if strings.Contains(upper, pattern) {
			matched = append(matched, pattern)
		}
	}
	
	return matched
}


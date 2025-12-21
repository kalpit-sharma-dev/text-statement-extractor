package rag

import (
	"fmt"
	"regexp"
	"strings"
)

// AnswerValidator validates LLM responses against context
type AnswerValidator struct{}

// NewAnswerValidator creates a new answer validator
func NewAnswerValidator() *AnswerValidator {
	return &AnswerValidator{}
}

// ValidateAnswer checks if answer is supported by context
func (av *AnswerValidator) ValidateAnswer(answer string, context string, query string) (bool, string) {
	// Extract numbers from answer
	answerNumbers := av.extractNumbers(answer)

	// Check if numbers exist in context
	for _, num := range answerNumbers {
		if !strings.Contains(context, num) {
			return false, fmt.Sprintf("Answer contains number %s not found in context", num)
		}
	}

	// Check for "I don't know" patterns (these are valid)
	dontKnowPatterns := []string{
		"don't have", "not in the", "not available", "cannot find",
		"don't have that information", "not found", "unavailable",
	}
	for _, pattern := range dontKnowPatterns {
		if strings.Contains(strings.ToLower(answer), pattern) {
			return true, "Valid 'not found' response"
		}
	}

	// Check if answer is too generic
	if len(answer) < 20 {
		return false, "Answer is too short/generic"
	}

	// Check if answer contains currency symbols (good sign for financial answers)
	if strings.Contains(answer, "₹") {
		// Extract currency amounts and verify they're in context
		currencyAmounts := av.extractCurrencyAmounts(answer)
		for _, amount := range currencyAmounts {
			if !strings.Contains(context, amount) {
				return false, fmt.Sprintf("Answer contains amount %s not found in context", amount)
			}
		}
	}

	return true, "Answer validated"
}

// extractNumbers extracts currency amounts and numbers
func (av *AnswerValidator) extractNumbers(text string) []string {
	// Pattern for currency: ₹X,XXX.XX or ₹X,XXX
	currencyPattern := regexp.MustCompile(`₹[\d,]+\.?\d*`)
	matches := currencyPattern.FindAllString(text, -1)

	// Pattern for standalone numbers (4+ digits, likely amounts)
	numberPattern := regexp.MustCompile(`\b\d{4,}[,.]?\d*\b`)
	numberMatches := numberPattern.FindAllString(text, -1)

	allNumbers := append(matches, numberMatches...)

	// Deduplicate
	seen := make(map[string]bool)
	unique := []string{}
	for _, num := range allNumbers {
		if !seen[num] {
			seen[num] = true
			unique = append(unique, num)
		}
	}

	return unique
}

// extractCurrencyAmounts extracts only currency amounts
func (av *AnswerValidator) extractCurrencyAmounts(text string) []string {
	currencyPattern := regexp.MustCompile(`₹[\d,]+\.?\d*`)
	matches := currencyPattern.FindAllString(text, -1)

	// Deduplicate
	seen := make(map[string]bool)
	unique := []string{}
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			unique = append(unique, match)
		}
	}

	return unique
}


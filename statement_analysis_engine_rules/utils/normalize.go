package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// NormalizeNarration normalizes narration text for better matching
func NormalizeNarration(narration string) string {
	// Remove extra spaces
	narration = strings.TrimSpace(narration)
	narration = regexp.MustCompile(`\s+`).ReplaceAllString(narration, " ")

	// Remove special characters that don't add meaning
	narration = regexp.MustCompile(`[^\w\s-]`).ReplaceAllString(narration, " ")

	return strings.TrimSpace(narration)
}

// NormalizeMerchantName normalizes merchant name
func NormalizeMerchantName(name string) string {
	name = strings.TrimSpace(name)
	
	// Remove common suffixes
	suffixes := []string{" PVT LTD", " PRIVATE LIMITED", " LIMITED", " LTD", " INC"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(strings.ToUpper(name), suffix)
	}

	return strings.TrimSpace(name)
}

// MaskAccountNumber masks account number for display
func MaskAccountNumber(accountNo string) string {
	if len(accountNo) < 4 {
		return "XXXX"
	}
	return "XXXXXX" + accountNo[len(accountNo)-4:]
}

// FormatAmount formats amount with commas
func FormatAmount(amount float64) string {
	// Simple formatting - can be enhanced
	formatted := fmt.Sprintf("%.2f", amount)
	return strings.ReplaceAll(formatted, ".", ",")
}


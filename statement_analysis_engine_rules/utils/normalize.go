package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// NormalizeNarration normalizes narration text for better matching
// CRITICAL: Also removes account statement footer text that contaminates classification
func NormalizeNarration(narration string) string {
	// Remove account statement footer/metadata that contaminates classification
	// Footer patterns to remove:
	// - "Account Branch :", "Address :", "City :", "State :", etc.
	// - "SHOP NO.", "FLORENCE BUILDING", etc. (account address)
	// - "Phone no. :", "Email :", "OD Limit :", etc.
	// - "Account No :", "Account Status :", etc.
	// - "RTGS/NEFT IFSC :", "MICR :", "Branch Code :", etc.
	// - "Statement From :", "To :", etc.
	
	// Remove footer patterns (everything after common footer markers)
	footerMarkers := []string{
		"Account Branch :",
		"Address        :",
		"City           :",
		"State          :",
		"Phone no.      :",
		"Email          :",
		"OD Limit       :",
		"Account No     :",
		"Account Status :",
		"Statement From :",
		"RTGS/NEFT IFSC :",
		"MICR :",
		"Branch Code    :",
		"Account Type   :",
		"SHOP NO.",
		"FLORENCE BUILDING",
		"VIMAN NAGAR",
		"PUNE411014",
		"GHAZIABAD",
		"MAHARASHTRA",
		"UTTAR PRADESH",
		"JOINT HOLDERS :",
		"PRIME POTENTIAL",
		"Open Date  :",
		"Nomination :",
	}
	
	for _, marker := range footerMarkers {
		if idx := strings.Index(narration, marker); idx > 0 {
			// Keep only the part before the footer marker
			// But preserve some context (first 200 chars should be enough for transaction narration)
			if idx < 200 {
				// If marker appears early, it might be part of narration, so check if it's followed by account details
				afterMarker := narration[idx:]
				// If followed by account-like patterns, it's footer
				if strings.Contains(afterMarker, "Account") || strings.Contains(afterMarker, "Branch") ||
					strings.Contains(afterMarker, "Address") || strings.Contains(afterMarker, "Email") {
					narration = narration[:idx]
					break
				}
			} else {
				// Marker appears late, definitely footer - truncate
				narration = narration[:idx]
				break
			}
		}
	}
	
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


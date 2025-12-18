package utils

import (
	"strings"
)

// PaymentRail represents the payment rail/channel used
type PaymentRail string

const (
	RailUPI        PaymentRail = "UPI"
	RailIMPS       PaymentRail = "IMPS"
	RailNEFT       PaymentRail = "NEFT"
	RailRTGS       PaymentRail = "RTGS"
	RailACH        PaymentRail = "ACH"
	RailNACH       PaymentRail = "NACH"
	RailECS        PaymentRail = "ECS"
	RailPOS        PaymentRail = "POS"
	RailNetBanking PaymentRail = "NetBanking"
	RailCheque     PaymentRail = "Cheque"
	RailUnknown    PaymentRail = "Unknown"
)

// DetectPaymentRail detects the payment rail from narration
// Layer 2: Payment Rail Detection (Context)
func DetectPaymentRail(narration string) PaymentRail {
	upper := strings.ToUpper(narration)

	// Check in order of specificity
	if strings.Contains(upper, "POS ") || strings.Contains(upper, "POS-") || strings.Contains(upper, "POS/") {
		return RailPOS
	}

	if strings.Contains(upper, "ACH D-") || strings.Contains(upper, "ACH C-") || strings.Contains(upper, "ACH DR") || strings.Contains(upper, "ACH CR") {
		return RailACH
	}

	if strings.Contains(upper, "NACH") {
		return RailNACH
	}

	if strings.Contains(upper, "ECS") {
		return RailECS
	}

	if strings.Contains(upper, "RTGS") {
		return RailRTGS
	}

	if strings.Contains(upper, "NEFT") {
		return RailNEFT
	}

	if strings.Contains(upper, "IMPS") {
		return RailIMPS
	}

	if strings.Contains(upper, "UPI-") || strings.Contains(upper, "UPI ") || strings.Contains(upper, "UPI/") ||
		strings.Contains(upper, "UPI@") || strings.Contains(upper, "@YBL") || strings.Contains(upper, "@PAYTM") ||
		strings.Contains(upper, "@OK") || strings.Contains(upper, "@AXL") || strings.Contains(upper, "@IBL") ||
		strings.Contains(upper, "@PTYES") || strings.Contains(upper, "PAYTM") || strings.Contains(upper, "PHONEPE") ||
		strings.Contains(upper, "GOOGLEPAY") || strings.Contains(upper, "GPAY") || strings.Contains(upper, "BHIM") {
		return RailUPI
	}

	if strings.Contains(upper, "IB ") || strings.Contains(upper, "IB-") || strings.Contains(upper, "IB/") ||
		strings.Contains(upper, "NET BANKING") || strings.Contains(upper, "FUNDS TRANSFER") {
		return RailNetBanking
	}

	if strings.Contains(upper, "CHQ") || strings.Contains(upper, "CHEQUE") {
		return RailCheque
	}

	return RailUnknown
}

// IsPersonToPersonTransfer checks if transaction is likely a person-to-person transfer
// Layer 7: Negative Classification (exclude transfers)
func IsPersonToPersonTransfer(narration string, merchant string, amount float64) bool {
	upper := strings.ToUpper(narration)
	merchantUpper := strings.ToUpper(merchant)

	// Check for explicit transfer indicators
	if strings.Contains(upper, "FUND TRANSFER") || strings.Contains(upper, "FUNDS TRANSFER") ||
		strings.Contains(upper, "SELF TRANSFER") || strings.Contains(upper, "TO SELF") {
		return true
	}

	// Check if merchant is a person name (common patterns)
	// Person names typically don't contain business keywords
	businessKeywords := []string{
		"PVT", "LTD", "LIMITED", "LLP", "INC", "CORP", "COMPANY", "COMP",
		"STORE", "SHOP", "MARKET", "TRADERS", "TRADING", "ENTERPRISE",
		"SERVICES", "SERVICE", "SOLUTIONS", "SOL", "TECHNOLOGIES", "TECH",
		"HOTEL", "RESTAURANT", "CAFE", "BAKERY", "PHARMACY", "MEDICAL",
		"HOSPITAL", "CLINIC", "BANK", "FINANCE", "FINSERV",
	}

	hasBusinessKeyword := false
	for _, keyword := range businessKeywords {
		if strings.Contains(merchantUpper, keyword) {
			hasBusinessKeyword = true
			break
		}
	}

	// If no business keywords and merchant looks like a person name
	if !hasBusinessKeyword && merchantUpper != "" && merchantUpper != "UNKNOWN" {
		// Person names typically have spaces, are 2-4 words, and don't have numbers
		words := strings.Fields(merchantUpper)
		if len(words) >= 2 && len(words) <= 4 {
			// Check if it contains numbers (businesses often have numbers)
			hasNumbers := false
			for _, word := range words {
				for _, char := range word {
					if char >= '0' && char <= '9' {
						hasNumbers = true
						break
					}
				}
				if hasNumbers {
					break
				}
			}
			if !hasNumbers {
				// Likely a person name
				return true
			}
		}
	}

	// Large IMPS/UPI transfers to individuals (likely person-to-person)
	if amount >= 10000 {
		if strings.Contains(upper, "IMPS") || strings.Contains(upper, "UPI") {
			// If merchant is a person name (no business keywords)
			if !hasBusinessKeyword && merchantUpper != "" && merchantUpper != "UNKNOWN" {
				words := strings.Fields(merchantUpper)
				if len(words) >= 2 && len(words) <= 4 {
					return true
				}
			}
		}
	}

	return false
}


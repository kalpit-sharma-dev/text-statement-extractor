package rules

import (
	"regexp"
	"strings"
)

// ExtractBeneficiary extracts beneficiary name from narration
func ExtractBeneficiary(narration string, method string) string {
	narration = strings.TrimSpace(narration)

	// For IMPS/NEFT/RTGS transactions
	if method == "IMPS" || method == "NEFT" || method == "RTGS" {
		// Pattern: IMPS-REF-MR/MRS/MS NAME-BANK-ACCOUNT-REASON
		re := regexp.MustCompile(`(?:IMPS|NEFT|RTGS)[- ]+(?:[^-]+-)?(?:MR|MRS|MS|MR\.|MRS\.|MS\.)[\s]+([A-Z\s]+?)(?:-|@|$|BANK|SBIN|HDFC|ICICI|AXIS|PNB|SBI)`)
		matches := re.FindStringSubmatch(narration)
		if len(matches) > 1 {
			beneficiary := strings.TrimSpace(matches[1])
			// Clean up
			beneficiary = strings.TrimSuffix(beneficiary, " -")
			beneficiary = strings.TrimSuffix(beneficiary, "-")
			if len(beneficiary) > 0 && len(beneficiary) < 50 {
				return beneficiary
			}
		}

		// Alternative pattern: Extract name between method and bank code
		re = regexp.MustCompile(`(?:IMPS|NEFT|RTGS)[- ]+[^-]+-([A-Z\s]+?)-[A-Z]{4}`)
		matches = re.FindStringSubmatch(narration)
		if len(matches) > 1 {
			beneficiary := strings.TrimSpace(matches[1])
			if len(beneficiary) > 0 && len(beneficiary) < 50 {
				return beneficiary
			}
		}
	}

	// For UPI transactions, extract payee name
	if method == "UPI" {
		merchant, payee := ExtractUPIDetails(narration)
		if merchant != "" {
			return merchant
		}
		if payee != "" {
			return payee
		}
	}

	// For salary transactions
	if strings.Contains(strings.ToUpper(narration), "SALARY") {
		re := regexp.MustCompile(`(?:SALARY|SAL)[\s]+(?:FOR|FROM)[\s]+([A-Z\s]+)`)
		matches := re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// For EMI/Loan transactions
	if method == "EMI" {
		re := regexp.MustCompile(`(?:EMI|LOAN|INSTALLMENT)[\s]+(?:FOR|OF)[\s]+([A-Z\s]+)`)
		matches := re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}

		// Extract loan provider name
		loanProviders := []string{
			"HDFC", "SBI", "ICICI", "AXIS", "PNB", "BOI",
			"HOME LOAN", "PERSONAL LOAN", "CAR LOAN",
		}
		for _, provider := range loanProviders {
			if strings.Contains(strings.ToUpper(narration), provider) {
				return provider
			}
		}
	}

	// For ACH transactions (recurring payments)
	if method == "ACH" {
		re := regexp.MustCompile(`ACH[-\s]+[CD][-\s]+([A-Z\s]+?)(?:-|LIMITED|LTD|PVT|PRIVATE)`)
		matches := re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// IsRecurringPayment checks if transaction appears to be recurring
func IsRecurringPayment(narration string, amount float64, date string, previousTransactions []string) bool {
	narration = strings.ToUpper(narration)

	// Check for recurring keywords
	recurringKeywords := []string{
		"INSTALLMENT", "EMI", "SIP", "RECURRING",
		"MONTHLY", "PREMIUM", "SUBSCRIPTION",
	}

	for _, keyword := range recurringKeywords {
		if strings.Contains(narration, keyword) {
			return true
		}
	}

	// Check if same amount and similar narration appeared before
	// This is a simple check - can be enhanced with date-based logic
	for _, prevNarration := range previousTransactions {
		if strings.Contains(prevNarration, narration) || strings.Contains(narration, prevNarration) {
			// Similar narration found - likely recurring
			return true
		}
	}

	return false
}

// ExtractRecurringDetails extracts details for recurring payments
func ExtractRecurringDetails(narration string) (name string, pattern string) {
	narration = strings.ToUpper(narration)

	// Extract name from common patterns
	re := regexp.MustCompile(`(?:FOR|TO|FROM)[\s]+([A-Z\s]+?)(?:-|LIMITED|LTD|MONTHLY|INSTALLMENT)`)
	matches := re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		name = strings.TrimSpace(matches[1])
	}

	// Determine pattern
	if strings.Contains(narration, "MONTHLY") {
		pattern = "Monthly"
	} else if strings.Contains(narration, "WEEKLY") {
		pattern = "Weekly"
	} else if strings.Contains(narration, "YEARLY") || strings.Contains(narration, "ANNUAL") {
		pattern = "Yearly"
	} else {
		pattern = "Monthly" // Default
	}

	return name, pattern
}


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
		// HDFC IMPS format: IMPS-REF-NAME-BANK-ACCOUNT-PURPOSE
		// Example: "IMPS-409612502129-MR  SUDHIR  KUMAR-SBIN-XXXXXXXXXXXX8121-REQPAY"
		// Pattern 1: Extract name after MR/MRS/MS prefix
		re := regexp.MustCompile(`(?:IMPS|NEFT|RTGS)[\s]*(?:CR|DR)?[- ]+(?:[^-]+-)?(?:MR|MRS|MS|MR\.|MRS\.|MS\.)[\s]+([A-Z\s]+?)(?:-|@|$|BANK|SBIN|HDFC|ICICI|AXIS|PNB|SBI|PUNB)`)
		matches := re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			beneficiary := strings.TrimSpace(matches[1])
			// Clean up
			beneficiary = strings.TrimSuffix(beneficiary, " -")
			beneficiary = strings.TrimSuffix(beneficiary, "-")
			if len(beneficiary) > 0 && len(beneficiary) < 50 {
				return beneficiary
			}
		}

		// HDFC RTGS format: RTGS CR/DR-IFSC-NAME-NAME-REF
		// Example: "RTGS CR-PUNB0041010-RAKESH CHANDRA SATI-KALPIT SHARMA-PUNBR52024042317337926"
		// Pattern 2: Extract first name after IFSC (for RTGS)
		if method == "RTGS" {
			re = regexp.MustCompile(`RTGS[\s]*(?:CR|DR)[- ]+[A-Z0-9]+-([A-Z\s]+?)-[A-Z\s]+`)
			matches = re.FindStringSubmatch(strings.ToUpper(narration))
			if len(matches) > 1 {
				beneficiary := strings.TrimSpace(matches[1])
				if len(beneficiary) > 0 && len(beneficiary) < 50 {
					return beneficiary
				}
			}
		}

		// Pattern 3: Extract name between method and bank code (generic fallback)
		re = regexp.MustCompile(`(?:IMPS|NEFT|RTGS)[\s]*(?:CR|DR)?[- ]+[^-]+-([A-Z\s]+?)-[A-Z]{4}`)
		matches = re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			beneficiary := strings.TrimSpace(matches[1])
			// Remove MR/MRS/MS prefix if present
			beneficiary = strings.TrimPrefix(beneficiary, "MR ")
			beneficiary = strings.TrimPrefix(beneficiary, "MRS ")
			beneficiary = strings.TrimPrefix(beneficiary, "MS ")
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
	// HDFC format: P:REF BANK SALARY FOR MONTH YEAR (e.g., "P:K16675 HDFC BANK SALARY FOR APR 2024")
	if strings.Contains(strings.ToUpper(narration), "SALARY") {
		// Extract bank name from salary narration
		re := regexp.MustCompile(`(?:P:[A-Z0-9]+\s+)?([A-Z\s]+?)\s+BANK\s+SALARY`)
		matches := re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			bankName := strings.TrimSpace(matches[1])
			if bankName != "" {
				return bankName + " BANK"
			}
		}
		// Fallback: Extract employer name
		re = regexp.MustCompile(`(?:SALARY|SAL)[\s]+(?:FOR|FROM)[\s]+([A-Z\s]+)`)
		matches = re.FindStringSubmatch(strings.ToUpper(narration))
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
	// HDFC format: ACH C/D- MERCHANT-REF
	// Examples: "ACH D- TP ACH MAXLIFEINSURA-1424041803", "ACH C- ICICI SECURITIES LIM-3614387"
	if method == "ACH" {
		// Pattern 1: ACH C/D- MERCHANT-REF
		re := regexp.MustCompile(`ACH[\s]*(?:C|D|CR|DR)[-\s]+([A-Z\s]+?)(?:-|LIMITED|LTD|PVT|PRIVATE|INSURA|SECURITIES)`)
		matches := re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			merchant := strings.TrimSpace(matches[1])
			// Clean up common prefixes
			merchant = strings.TrimPrefix(merchant, "TP ACH ")
			if len(merchant) > 0 && len(merchant) < 50 {
				return merchant
			}
		}
		// Pattern 2: Fallback - extract after ACH C/D
		re = regexp.MustCompile(`ACH[\s]*(?:C|D|CR|DR)[-\s]+([A-Z\s]+?)(?:-|$)`)
		matches = re.FindStringSubmatch(strings.ToUpper(narration))
		if len(matches) > 1 {
			merchant := strings.TrimSpace(matches[1])
			if len(merchant) > 0 && len(merchant) < 50 {
				return merchant
			}
		}
	}

	return ""
}

// IsRecurringPayment checks if transaction explicitly indicates recurring payment
func IsRecurringPayment(narration string, amount float64, date string, previousTransactions []string) bool {
	narration = strings.ToUpper(narration)

	// Only flag as recurring if EXPLICIT recurring keywords are present in narration
	// These are keywords that explicitly indicate a recurring/subscription payment
	explicitRecurringKeywords := []string{
		"INSTALLMENT", "INSTALMENT", "EMI", "SIP", "RECURRING",
		"SUBSCRIPTION", "AUTO DEBIT", "AUTODEBIT", "STANDING INSTRUCTION",
		"SI DEBIT", "NACH", "ECS",
		// Specific recurring patterns
		"RD INSTALLMENT", "FD INSTALLMENT", "LOAN INSTALLMENT",
	}

	for _, keyword := range explicitRecurringKeywords {
		if strings.Contains(narration, keyword) {
			return true
		}
	}

	// Check for insurance premium (explicit recurring)
	if strings.Contains(narration, "PREMIUM") && 
	   (strings.Contains(narration, "INSURANCE") || strings.Contains(narration, "POLICY")) {
		return true
	}

	// Note: Removed the generic "MONTHLY" keyword and similarity checks
	// These caused false positives for regular frequent payments
	// Pattern-based recurring detection is now handled by CalculateRecurringPayments()

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


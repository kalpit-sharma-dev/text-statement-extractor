package suppression

import (
	"classify/statement_analysis_engine_rules/models"
	"strings"
)

// SuppressionRule determines if a transaction should be excluded from anomaly detection
type SuppressionRule struct {
	// Skip anomaly detection entirely for these
	SkipAnomalyDetection bool
	
	// Cap severity at a maximum level
	MaxSeverity string // "INFO", "LOW", "MEDIUM", "HIGH", "CRITICAL"
}

// Suppressor applies bank-grade suppression rules
type Suppressor struct {
	// Trusted merchants (known, recurring)
	trustedMerchants map[string]bool
	
	// Utility merchants (always excluded)
	utilityMerchants map[string]bool
	
	// Minimum amount threshold (ignore small transactions)
	minAmountThreshold float64
	
	// Recurring payment patterns
	recurringCategories map[string]bool
}

// NewSuppressor creates a new suppressor with default rules
func NewSuppressor() *Suppressor {
	return &Suppressor{
		trustedMerchants: buildTrustedMerchants(),
		utilityMerchants: buildUtilityMerchants(),
		minAmountThreshold: 1000.0, // Ignore transactions < ₹1K
		recurringCategories: buildRecurringCategories(),
	}
}

// ShouldSuppress determines if transaction should be suppressed
func (s *Suppressor) ShouldSuppress(txn models.ClassifiedTransaction) SuppressionRule {
	rule := SuppressionRule{
		SkipAnomalyDetection: false,
		MaxSeverity:          "CRITICAL", // No cap by default
	}
	
	merchant := strings.ToUpper(strings.TrimSpace(txn.Merchant))
	category := txn.Category
	amount := txn.WithdrawalAmt
	
	// Rule 1: Skip credits/refunds entirely (beneficial transactions)
	if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
		if category == "Refund" || category == "Income" {
			rule.SkipAnomalyDetection = true
			return rule
		}
	}
	
	// Rule 2: Skip utility bills (always predictable, recurring)
	if s.utilityMerchants[merchant] || isUtilityCategory(category) {
		rule.SkipAnomalyDetection = true
		return rule
	}
	
	// Rule 3: Skip small transactions (noise)
	if amount > 0 && amount < s.minAmountThreshold {
		rule.SkipAnomalyDetection = true
		return rule
	}
	
	// Rule 4: Severity caps for specific categories
	
	// Salary - max INFO (notification, not alert)
	if category == "Income" || txn.Method == "Salary" {
		rule.MaxSeverity = "INFO"
		return rule
	}
	
	// Self transfers - max LOW (awareness, not risk)
	if category == "Investment" && isSelfTransfer(txn) {
		rule.MaxSeverity = "LOW"
		return rule
	}
	
	// Credit card payments (CRED, etc.) - max MEDIUM
	if s.trustedMerchants[merchant] && isCreditCardPayment(category, merchant) {
		rule.MaxSeverity = "MEDIUM"
		return rule
	}
	
	// Recurring categories - cap at MEDIUM
	if s.recurringCategories[category] {
		rule.MaxSeverity = "MEDIUM"
		return rule
	}
	
	return rule
}

// buildTrustedMerchants returns map of trusted merchants
func buildTrustedMerchants() map[string]bool {
	trusted := []string{
		"CRED", "CRED CLUB", "CRED PAY",
		"PAYTM", "PHONEPE", "GPAY", "GOOGLE PAY",
		"AMAZON PAY", "AMAZONPAY",
		"ZERODHA", "GROWW", "UPSTOX", "ANGEL ONE",
		"HDFC BANK", "ICICI BANK", "SBI BANK", "AXIS BANK", "KOTAK BANK",
		"PAYTM MONEY", "PAYTM PAYMENTS",
	}
	
	m := make(map[string]bool)
	for _, t := range trusted {
		m[t] = true
	}
	return m
}

// buildUtilityMerchants returns map of utility merchants
func buildUtilityMerchants() map[string]bool {
	utilities := []string{
		"AIRTEL", "VI", "VODAFONE IDEA", "JIO", "RELIANCE JIO",
		"BSES", "TATA POWER", "ADANI ELECTRICITY", "MAHARASHTRA STATE ELECTRICITY",
		"BEST", "MUMBAI ELECTRICITY",
		"INDANE", "HP GAS", "BHARAT GAS",
		"WATER BOARD", "MUNICIPAL CORPORATION",
		"BROADBAND", "FIBER", "INTERNET",
	}
	
	m := make(map[string]bool)
	for _, u := range utilities {
		m[u] = true
	}
	return m
}

// buildRecurringCategories returns categories that are typically recurring
func buildRecurringCategories() map[string]bool {
	recurring := []string{
		"Bills_Utilities",
		"Loan",
		"EMI",
		"RD", // Recurring Deposit
		"SIP", // Systematic Investment Plan
	}
	
	m := make(map[string]bool)
	for _, r := range recurring {
		m[r] = true
	}
	return m
}

// isUtilityCategory checks if category is utility-related
func isUtilityCategory(category string) bool {
	utilityCategories := []string{
		"Bills_Utilities",
		"Utilities",
		"Bills",
	}
	
	categoryUpper := strings.ToUpper(category)
	for _, uc := range utilityCategories {
		if strings.Contains(categoryUpper, strings.ToUpper(uc)) {
			return true
		}
	}
	return false
}

// isSelfTransfer checks if transaction is a self-transfer
func isSelfTransfer(txn models.ClassifiedTransaction) bool {
	// Check method
	selfMethods := []string{"Self_Transfer", "IMPS", "NEFT", "RTGS"}
	methodUpper := strings.ToUpper(txn.Method)
	for _, sm := range selfMethods {
		if strings.Contains(methodUpper, strings.ToUpper(sm)) {
			// Additional check: beneficiary name matches account holder patterns
			beneficiary := strings.ToUpper(txn.Beneficiary)
			if beneficiary != "" {
				// Common self-transfer patterns
				selfPatterns := []string{"SELF", "OWN", "SAME", "ACCOUNT"}
				for _, pattern := range selfPatterns {
					if strings.Contains(beneficiary, pattern) {
						return true
					}
				}
			}
			return true
		}
	}
	return false
}

// isCreditCardPayment checks if transaction is a credit card bill payment
func isCreditCardPayment(category, merchant string) bool {
	merchantUpper := strings.ToUpper(merchant)
	
	// CRED is a credit card payment platform
	if strings.Contains(merchantUpper, "CRED") {
		return true
	}
	
	// Credit card bill payment category
	if category == "Loan" || category == "Bills_Utilities" {
		ccKeywords := []string{"CREDIT", "CARD", "BILL", "PAYMENT"}
		for _, keyword := range ccKeywords {
			if strings.Contains(merchantUpper, keyword) {
				return true
			}
		}
	}
	
	return false
}


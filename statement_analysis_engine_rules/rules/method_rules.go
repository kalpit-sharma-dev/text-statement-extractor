package rules

import (
	"regexp"
	"strings"
)

// ClassifyMethod classifies the transaction method based on narration
func ClassifyMethod(narration string) string {
	narration = strings.ToUpper(narration)

	// UPI patterns
	upiPatterns := []string{
		"UPI-", "UPI ", "UPI/", "UPI@", "UPIINTENT", "UPI TRANSACTION",
		"PAYTM", "PHONEPE", "GOOGLEPAY", "BHIM", "AMAZONPAY",
		"@YBL", "@PAYTM", "@OK", "@AXL", "@IBL", "@PTYES",
	}

	// IMPS patterns
	impsPatterns := []string{
		"IMPS-", "IMPS ", "IMPS/", "INSTANT PAYMENT",
	}

	// NEFT patterns
	neftPatterns := []string{
		"NEFT-", "NEFT ", "NEFT/", "NATIONAL ELECTRONIC FUND TRANSFER",
	}

	// RTGS patterns
	rtgsPatterns := []string{
		"RTGS", "REAL TIME GROSS SETTLEMENT",
	}

	// Investment/Savings patterns (exclude these from EMI)
	investmentPatterns := []string{
		"RD", "FD", "SIP", "RECURRING DEPOSIT", "FIXED DEPOSIT",
		"MUTUAL FUND", "INVESTMENT", "PPF", "ELSS", "RD INSTALLMENT",
		"INDIAN CLEARING CORPORATION", "INDIAN CLEARING CORPORATION LIMITED",
		"INDIAN C LEARING CORPORATION", "INDIAN C LEARING CORPORATION LIMITED", // Handle typo with space
		"NSDL", "CDSL", "CLEARING CORPORATION",
	}

	// EMI patterns (loans/repayments - exclude investments)
	emiPatterns := []string{
		"EMI", "LOAN", "REPAYMENT",
		"HOME LOAN", "PERSONAL LOAN", "CAR LOAN", "EDUCATION LOAN",
		"LOAN INSTALLMENT", "LOAN EMI", "LOAN REPAYMENT",
	}

	// ACH patterns
	achPatterns := []string{
		"ACH", "AUTOMATED CLEARING HOUSE",
		"ACH C-", "ACH D-", "ACH CR", "ACH DR",
	}

	// ATM Withdrawal patterns (check before Debit Card)
	atmWithdrawalPatterns := []string{
		"EAW", "ATW", "NWD", "ATM WITHDRAWAL", "ATM CASH WITHDRAWAL",
		"ELECTRONIC ATM WITHDRAWAL", "ATM CASH",
	}
	
	// Debit Card patterns
	debitCardPatterns := []string{
		"DC", "POS", "DEBIT CARD", "ATM", "CASH WITHDRAWAL",
		"SWIPE", "CARD TRANSACTION", "VISA", "MASTERCARD",
	}

	// Net Banking patterns
	netBankingPatterns := []string{
		"NET BANKING", "ONLINE BANKING", "INTERNET BANKING",
		"IB ", "IB-", "IB/", "ONLINE TRANSFER",
	}

	// Salary patterns
	salaryPatterns := []string{
		"SALARY", "SAL FOR", "PAYROLL", "WAGES", "BONUS",
	}

	// Interest patterns
	interestPatterns := []string{
		"INTEREST", "INTEREST PAID", "INTEREST CREDIT",
	}

	// Dividend patterns
	dividendPatterns := []string{
		"DIV", "DIVIDEND", "DIVIDEND CREDIT", "DIV CR",
	}

	// Insurance premium patterns (check before other methods)
	insurancePatterns := []string{
		"HLIC", "HLIC_INST", "HLIC INST", "HDFC LIFE", "LIC", "INSURANCE",
		"PREMIUM", "MAXLIFE", "SBI LIFE", "ICICI PRUDENTIAL", "BAJAJ ALLIANZ",
	}

	// Check patterns
	checkPatterns := []string{
		"CHQ", "CHEQUE", "CHEQUE NO",
	}

	// Check for insurance premium FIRST (before RD to catch HLIC_INST)
	// Insurance premiums should be classified as "Insurance" method
	for _, pattern := range insurancePatterns {
		if strings.Contains(narration, pattern) {
			// Additional check: if it contains "INST" or "INSTALLMENT", it's likely insurance premium
			if strings.Contains(narration, "INST") || strings.Contains(narration, "INSTALLMENT") ||
				strings.Contains(narration, "PREMIUM") {
				return "Insurance"
			}
			// Also return Insurance if it's clearly an insurance company
			if strings.Contains(narration, "HLIC") || strings.Contains(narration, "HDFC LIFE") ||
				strings.Contains(narration, "LIC") || strings.Contains(narration, "MAXLIFE") ||
				strings.Contains(narration, "SBI LIFE") {
				return "Insurance"
			}
		}
	}

	// Check for dividends (income) - should be classified as "Dividend"
	for _, pattern := range dividendPatterns {
		if strings.Contains(narration, pattern) {
			return "Dividend"
		}
	}

	// Check for Indian Clearing Corporation (investment-related)
	// This handles stock market investments/clearing transactions
	// If debited (withdrawal), it's an investment purchase
	// If credited (deposit), it's investment returns/income
	// Handle both correct spelling and typo variations (with/without space between C and LEARING/CLEARING)
	if strings.Contains(narration, "INDIAN CLEARING CORPORATION") ||
		strings.Contains(narration, "INDIAN CLEARING CORPORATION LIMITED") ||
		strings.Contains(narration, "INDIAN C LEARING CORPORATION") ||
		strings.Contains(narration, "INDIAN C LEARING CORPORATION LIMITED") ||
		strings.Contains(narration, "NSDL") ||
		strings.Contains(narration, "CDSL") {
		// Classify as Investment method
		// Category will be determined by whether it's a deposit (income/returns) or withdrawal (investment)
		return "Investment"
	}

	// Check for investments/savings FIRST (RD, FD, SIP) - exclude from EMI and other methods
	// These should NOT be classified as EMI even if they contain "INSTALLMENT"
	// Also classify them as their own method type for proper tracking
	// Check RD with word boundaries to avoid false matches (e.g., "PAYTMQRD" in UPI)
	// Match: " RD ", " RD-", "RD INSTALLMENT", "RD-", or starts with "RD "
	if (strings.Contains(narration, " RD ") ||
		strings.Contains(narration, " RD-") ||
		strings.Contains(narration, "RD INSTALLMENT") ||
		strings.Contains(narration, "RECURRING DEPOSIT") ||
		strings.HasPrefix(narration, "RD ") ||
		strings.HasPrefix(narration, "RD-")) &&
		!strings.Contains(narration, "PAYTMQRD") && // Exclude UPI transactions with "QRD"
		!strings.Contains(narration, "UPI") { // Exclude UPI transactions
		return "RD" // Recurring Deposit
	}
	if (strings.Contains(narration, " FD ") ||
		strings.Contains(narration, " FD-") ||
		strings.Contains(narration, "FIXED DEPOSIT") ||
		strings.HasPrefix(narration, "FD ") ||
		strings.HasPrefix(narration, "FD-")) &&
		!strings.Contains(narration, "UPI") {
		return "FD" // Fixed Deposit
	}
	if (strings.Contains(narration, " SIP ") ||
		strings.Contains(narration, "SIP ") ||
		strings.Contains(narration, "SIP-") ||
		strings.HasPrefix(narration, "SIP ")) &&
		!strings.Contains(narration, "UPI") {
		return "SIP" // Systematic Investment Plan
	}

	// Check UPI
	for _, pattern := range upiPatterns {
		if strings.Contains(narration, pattern) {
			return "UPI"
		}
	}

	// Check IMPS
	for _, pattern := range impsPatterns {
		if strings.Contains(narration, pattern) {
			return "IMPS"
		}
	}

	// Check NEFT
	for _, pattern := range neftPatterns {
		if strings.Contains(narration, pattern) {
			return "NEFT"
		}
	}

	// Check RTGS
	for _, pattern := range rtgsPatterns {
		if strings.Contains(narration, pattern) {
			return "RTGS"
		}
	}

	// Check for other investment patterns (but don't set method, let category handle it)
	isInvestment := false
	for _, pattern := range investmentPatterns {
		if strings.Contains(narration, pattern) {
			isInvestment = true
			break
		}
	}

	// Check EMI (only if not an investment)
	// EMI requires explicit loan-related keywords, not just "INSTALLMENT"
	if !isInvestment {
		// Check for explicit EMI/LOAN keywords
		hasLoanKeyword := strings.Contains(narration, "LOAN") ||
			strings.Contains(narration, "EMI") ||
			strings.Contains(narration, "REPAYMENT")

		// Only classify as EMI if it has loan-related keywords
		// Don't match standalone "INSTALLMENT" (could be RD, FD, etc.)
		if hasLoanKeyword {
			for _, pattern := range emiPatterns {
				if strings.Contains(narration, pattern) {
					return "EMI"
				}
			}
		}
	}

	// Check ACH
	for _, pattern := range achPatterns {
		if strings.Contains(narration, pattern) {
			return "ACH"
		}
	}

	// Check ATM Withdrawal (before Debit Card to catch EAW/ATW)
	for _, pattern := range atmWithdrawalPatterns {
		if strings.Contains(narration, pattern) {
			return "ATMWithdrawal"
		}
	}

	// Check Debit Card
	for _, pattern := range debitCardPatterns {
		if strings.Contains(narration, pattern) {
			return "DebitCard"
		}
	}

	// Check Net Banking
	for _, pattern := range netBankingPatterns {
		if strings.Contains(narration, pattern) {
			return "NetBanking"
		}
	}

	// Check Salary
	for _, pattern := range salaryPatterns {
		if strings.Contains(narration, pattern) {
			return "Salary"
		}
	}

	// Check Interest
	for _, pattern := range interestPatterns {
		if strings.Contains(narration, pattern) {
			return "Interest"
		}
	}

	// Check Dividend (if not already matched above)
	for _, pattern := range dividendPatterns {
		if strings.Contains(narration, pattern) {
			return "Dividend"
		}
	}

	// Check Cheque
	for _, pattern := range checkPatterns {
		if strings.Contains(narration, pattern) {
			return "Cheque"
		}
	}

	// Default to Other
	return "Other"
}

// IsBillPayment checks if transaction is a bill payment
func IsBillPayment(narration string) bool {
	narration = strings.ToUpper(narration)
	billPatterns := []string{
		"BILL", "RECHARGE", "PREPAID", "POSTPAID",
		"ELECTRICITY", "WATER", "GAS", "PHONE", "INTERNET",
		"INSURANCE", "PREMIUM", "LIC", "HDFC LIFE", "HLIC", "HLIC_INST", "HLIC INST",
		"MAXLIFE", "SBI LIFE", "ICICI PRUDENTIAL", "BAJAJ ALLIANZ",
		"PVVNL", "IGL", "AIRTEL", "JIO", "VODAFONE",
	}

	for _, pattern := range billPatterns {
		if strings.Contains(narration, pattern) {
			return true
		}
	}
	return false
}

// ExtractUPIDetails extracts merchant/payee name from UPI narration
func ExtractUPIDetails(narration string) (merchant string, payee string) {
	narration = strings.TrimSpace(narration)

	// UPI format: UPI-MERCHANT NAME-UPIID@BANK-REF-UPI
	// Extract merchant name (between UPI- and first - or @)
	re := regexp.MustCompile(`UPI-([^-@]+)`)
	matches := re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		merchant = strings.TrimSpace(matches[1])
	}

	// Extract payee from UPI ID (before @)
	re = regexp.MustCompile(`([^@\s]+)@`)
	matches = re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		payee = strings.TrimSpace(matches[1])
	}

	return merchant, payee
}

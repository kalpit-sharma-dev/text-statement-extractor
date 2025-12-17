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

	// EMI patterns
	emiPatterns := []string{
		"EMI", "LOAN", "INSTALLMENT", "REPAYMENT",
		"HOME LOAN", "PERSONAL LOAN", "CAR LOAN", "EDUCATION LOAN",
	}

	// ACH patterns
	achPatterns := []string{
		"ACH", "AUTOMATED CLEARING HOUSE",
		"ACH C-", "ACH D-", "ACH CR", "ACH DR",
	}

	// Debit Card patterns
	debitCardPatterns := []string{
		"POS", "DEBIT CARD", "ATM", "CASH WITHDRAWAL",
		"SWIPE", "CARD TRANSACTION", "VISA", "MASTERCARD",
	}

	// Net Banking patterns
	netBankingPatterns := []string{
		"NET BANKING", "ONLINE BANKING", "INTERNET BANKING",
		"IB ", "IB-", "IB/", "ONLINE TRANSFER",
	}

	// Salary patterns
	salaryPatterns := []string{
		"SALARY", "SAL FOR", "PAYROLL", "WAGES",
	}

	// Interest patterns
	interestPatterns := []string{
		"INTEREST", "INTEREST PAID", "INTEREST CREDIT",
	}

	// Check patterns
	checkPatterns := []string{
		"CHQ", "CHEQUE", "CHEQUE NO",
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

	// Check EMI
	for _, pattern := range emiPatterns {
		if strings.Contains(narration, pattern) {
			return "EMI"
		}
	}

	// Check ACH
	for _, pattern := range achPatterns {
		if strings.Contains(narration, pattern) {
			return "ACH"
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
		"INSURANCE", "PREMIUM", "LIC", "HDFC LIFE", "MAXLIFE",
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


package rules

import (
	"regexp"
	"strings"
)

// ClassifyMethod classifies the transaction method based on narration
func ClassifyMethod(narration string) string {
	narration = strings.ToUpper(narration)

	// UPI Reversal patterns (check before regular UPI)
	// Pattern: REV-UPI-08821130001725-KALPIT.COOL2006@OKHDFCBANK-209945000965-UPI
	upiReversalPatterns := []string{
		"REV-UPI", "REV UPI", "REV-UPI-", "REV UPI-",
		"UPI REVERSAL", "UPI REV", "UPI REFUND",
	}

	// UPI patterns
	upiPatterns := []string{
		"UPI-", "UPI ", "UPI/", "UPI@", "UPIINTENT", "UPI TRANSACTION",
		"PAYTM", "PHONEPE", "GOOGLEPAY", "BHIM", "AMAZONPAY",
		"@YBL", "@PAYTM", "@OK", "@OKAXIS", "@OKHDFC", "@OKICICI", "@AXL", "@IBL", "@PTYES",
		"@IDFCFIRST", "@OKAXIS", "@OKAXIS", "@HDFCBANK", "@AXISBANK",
	}

	// IMPS Reversal patterns (check before regular IMPS)
	// Pattern: REV-IMPS-112900179557-KALPIT KUMAR SHARMA-PYTM-XXXXXXXXXXXX8734-WAZIRX
	impsReversalPatterns := []string{
		"REV-IMPS", "REV IMPS", "REV-IMPS-", "REV IMPS-",
		"IMPS REVERSAL", "IMPS REV", "IMPS REFUND",
	}

	// IMPS patterns (including ICICI codes)
	impsPatterns := []string{
		"IMPS-", "IMPS ", "IMPS/", "INSTANT PAYMENT",
		"MMT", "MMT-", "MMT ", // Mobile Money Transfer (Insta FT - IMPS)
	}

	// NEFT patterns (including ICICI codes)
	// Pattern: NEFT CR-YESB0000001-ZERODHA BROKING LIMITED NSE CLIENT-KALPIT KUMAR SHARMA
	neftPatterns := []string{
		"NEFT-", "NEFT ", "NEFT/", "NATIONAL ELECTRONIC FUND TRANSFER",
		"NEFT CR", "NEFT CR-", "NEFT CR ", // NEFT Credit
		"NEFT DR", "NEFT DR-", "NEFT DR ", // NEFT Debit
		"N CHG", "N-CHG", // NEFT Charges
	}

	// RTGS patterns (HDFC format: RTGS CR/DR-IFSC-NAME-NAME-REF)
	rtgsPatterns := []string{
		"RTGS", "REAL TIME GROSS SETTLEMENT",
		"RTGS CR", "RTGS DR", "RTGS CR-", "RTGS DR-",
	}
	
	// Self-Transfer patterns (ICICI internal transfers)
	selfTransferPatterns := []string{
		"INF-", "INF ", "INF/", "INTERNET FUND TRANSFER IN LINKED ACCOUNTS",
		"INFT-", "INFT ", "INFT/", "INTERNAL FUND TRANSFER",
	}

	// Investment/Savings patterns (exclude these from EMI)
	investmentPatterns := []string{
		"RD", "FD", "SIP", "RECURRING DEPOSIT", "FIXED DEPOSIT",
		"MUTUAL FUND", "INVESTMENT", "PPF", "ELSS", "RD INSTALLMENT",
		"INDIAN CLEARING CORPORATION", "INDIAN CLEARING CORPORATION LIMITED",
		"INDIAN C LEARING CORPORATION", "INDIAN C LEARING CORPORATION LIMITED", // Handle typo with space
		"NSDL", "CDSL", "CLEARING CORPORATION",
		"ZERODHA", "ZERODHA BROKING", "ZERODHA BROKING LTD", "ZERODHABROKING",
		"ZERODHAMF", "ZERODHA COIN", "ICCL ZERODHA", // Mutual fund patterns
		"KITE", "KITE DEPOSIT", // Zerodha Kite trading platform
		"BROKING", "BROKING LTD", "HSL SEC", "HSL", "SEC", // Stock broking companies
		"ANGEL BROKING", "ICICI SECURITIES", "HDFC SECURITIES", "KOTAK SECURITIES",
		"SHAREKHAN", "MOTILAL OSWAL", "IIFL", "5PAISA",
		"EBA", "EBA-", "EBA ", // ICICI Direct transactions (ICICI-specific - used as fallback)
		"SGB", "SGB-", "SGB ", "SOVEREIGN GOLD BOND", // Sovereign Gold Bond
	}

	// EMI patterns (loans/repayments - exclude investments)
	// Pattern: EMI 4452581 CHQ S44525810472 04214452581
	emiPatterns := []string{
		"EMI", "LOAN", "REPAYMENT",
		"HOME LOAN", "PERSONAL LOAN", "CAR LOAN", "EDUCATION LOAN",
		"LOAN INSTALLMENT", "LOAN EMI", "LOAN REPAYMENT",
		"EMI CHQ", "EMI CHEQUE", // EMI with cheque reference
		"LNPY", "LNPY-", "LNPY ", "LINKED LOAN PAYMENT", // ICICI loan payment code (ICICI-specific - used as fallback)
	}

	// ACH patterns (HDFC format: ACH C/D- MERCHANT-REF)
	// Examples: "ACH D- TP ACH MAXLIFEINSURA-1424041803", "ACH C- ICICI SECURITIES LIM-3614387"
	achPatterns := []string{
		"ACH", "AUTOMATED CLEARING HOUSE",
		"ACH C-", "ACH D-", "ACH C ", "ACH D ", "ACH CR", "ACH DR",
		"ACH C", "ACH D", // Handle cases without dash/space
	}

	// ATM Withdrawal patterns (check before Debit Card)
	atmWithdrawalPatterns := []string{
		"EAW", "ATW", "NWD", "ATM WITHDRAWAL", "ATM CASH WITHDRAWAL",
		"ELECTRONIC ATM WITHDRAWAL", "ATM CASH",
		"VAT", "MAT", "NFS", "CCWD", // ICICI ATM codes (ICICI-specific - used as fallback)
	}

	// POS Reversal/Refund patterns (check before regular POS)
	posReversalPatterns := []string{
		"CRV POS", "CRV POS ", "CRV POS-", "POS REVERSAL", "POS REFUND",
		"REVERSAL POS", "REFUND POS", "CARD REVERSAL", "CARD REFUND",
	}

	// International Card Markup/Charges patterns
	intlMarkupPatterns := []string{
		"INTL POS", "INTL POS TXN MARKUP", "DC INTL POS", ".DC INTL POS",
		"INTERNATIONAL POS", "FOREIGN TRANSACTION", "FX MARKUP", "FOREIGN EXCHANGE",
	}

	// Debit Card patterns
	debitCardPatterns := []string{
		"DC", "POS", "DEBIT CARD", "ATM", "CASH WITHDRAWAL",
		"SWIPE", "CARD TRANSACTION", "VISA", "MASTERCARD",
		"VPS", "IPS", "VPS-", "IPS-", // ICICI debit card transaction codes (ICICI-specific - used as fallback)
	}

	// Net Banking patterns
	netBankingPatterns := []string{
		"NET BANKING", "ONLINE BANKING", "INTERNET BANKING",
		"IB ", "IB-", "IB/", "ONLINE TRANSFER",
	}

	// Salary patterns (HDFC format: P:REF BANK SALARY FOR MONTH YEAR)
	salaryPatterns := []string{
		"SALARY", "SAL FOR", "PAYROLL", "WAGES", "BONUS",
		"P:", // HDFC salary prefix (P:K16675 HDFC BANK SALARY FOR APR 2024)
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
		"LCCBRN CMS", "UCCBRN CMS", // ICICI cheque collection codes (ICICI-specific - used as fallback)
	}
	
	// Bill payment patterns (ICICI-specific - used as fallback after generic patterns)
	// NOTE: These are ICICI bank-specific transaction codes. They are checked AFTER generic patterns
	// (UPI, IMPS, NEFT, etc.) to ensure generic patterns take priority for all banks.
	billPaymentPatterns := []string{
		"BBPS", "BBPS-", "BBPS ", "BHARAT BILL PAYMENT",
		"BPAY", "BPAY-", "BPAY ", "BILL PAYMENT",
		"RCHG", "RCHG-", "RCHG ", "RECHARGE",
		"TOP", "TOP-", "TOP ", "MOBILE RECHARGE",
		"BIL", "BIL-", "BIL ", "INTERNET BILL PAYMENT",
		"PAVC", "PAVC-", "PAVC ", "PAY ANY VISA CREDIT CARD",
	}
	
	// Online shopping patterns (ICICI-specific - used as fallback after generic patterns)
	// NOTE: ICICI-specific code. Checked after generic UPI patterns.
	onlineShoppingPatterns := []string{
		"ONL", "ONL-", "ONL ", "ONLINE SHOPPING",
	}
	
	// Tax payment patterns (ICICI-specific - used as fallback after generic patterns)
	// NOTE: ICICI-specific codes. Checked after generic patterns.
	taxPatterns := []string{
		"DTAX", "DTAX-", "DTAX ", "DIRECT TAX",
		"IDTX", "IDTX-", "IDTX ", "INDIRECT TAX",
	}

	// Check for self-transfer patterns FIRST (ICICI internal transfers)
	// INF/INFT are internal fund transfers within ICICI Bank (linked accounts)
	for _, pattern := range selfTransferPatterns {
		if strings.Contains(narration, pattern) {
			return "Self_Transfer"
		}
	}
	
	// Check for insurance premium (before RD to catch HLIC_INST)
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

	// Check for stock broking companies (Zerodha, etc.) and ICICI investment codes
	// These are investment-related transactions
	// EBA = ICICI Direct transactions (stock trading)
	// SGB = Sovereign Gold Bond
	// ZERODHAMF = Zerodha Mutual Fund
	// ICCL ZERODHA COIN = Zerodha Coin (mutual fund platform)
	// KITE = Zerodha Kite trading platform
	if strings.Contains(narration, "ZERODHA") ||
		strings.Contains(narration, "ZERODHA BROKING") ||
		strings.Contains(narration, "ZERODHA BROKING LTD") ||
		strings.Contains(narration, "ZERODHABROKING") ||
		strings.Contains(narration, "ZERODHAMF") ||
		strings.Contains(narration, "ZERODHA COIN") ||
		strings.Contains(narration, "ICCL ZERODHA") ||
		strings.Contains(narration, "KITE") ||
		strings.Contains(narration, "KITE DEPOSIT") ||
		strings.Contains(narration, "HSL SEC") ||
		(strings.Contains(narration, "HSL") && strings.Contains(narration, "SEC")) ||
		strings.Contains(narration, "EBA-") || strings.Contains(narration, "EBA ") || strings.HasPrefix(narration, "EBA") ||
		strings.Contains(narration, "SGB-") || strings.Contains(narration, "SGB ") || strings.HasPrefix(narration, "SGB") ||
		strings.Contains(narration, "SOVEREIGN GOLD BOND") {
		// Classify as Investment method
		return "Investment"
	}

	// Check for investments/savings FIRST (RD, FD, SIP) - exclude from EMI and other methods
	// These should NOT be classified as EMI even if they contain "INSTALLMENT"
	// Also classify them as their own method type for proper tracking
	// HDFC RD format: ACCOUNT-RD INSTALLMENT-MONTH YEAR (e.g., "50400334918713- RD INSTALLMENT-APR 2024")
	// Check RD with word boundaries to avoid false matches (e.g., "PAYTMQRD" in UPI)
	// Match: " RD ", " RD-", "RD INSTALLMENT", "RD-", or starts with "RD "
	if (strings.Contains(narration, " RD ") ||
		strings.Contains(narration, " RD-") ||
		strings.Contains(narration, "- RD ") ||
		strings.Contains(narration, "-RD ") ||
		strings.Contains(narration, "RD INSTALLMENT") ||
		strings.Contains(narration, "RECURRING DEPOSIT") ||
		strings.HasPrefix(narration, "RD ") ||
		strings.HasPrefix(narration, "RD-")) &&
		!strings.Contains(narration, "PAYTMQRD") && // Exclude UPI transactions with "QRD"
		!strings.Contains(narration, "UPI") { // Exclude UPI transactions
		return "RD" // Recurring Deposit
	}
	// Check FD patterns (Fixed Deposit)
	// Pattern: FD THROUGH NET-50300618314680:KALPIT KUMAR SHARMA
	// Pattern: IB FD PREMAT PRINCIPAL-50300618314680
	// Pattern: IB FD PREMAT INT PAID-50300618314680
	if strings.Contains(narration, "FD THROUGH NET") ||
		strings.Contains(narration, "FD THROUGH") ||
		strings.Contains(narration, "FD PREMAT") ||
		strings.Contains(narration, "FD PREMATURE") ||
		(strings.Contains(narration, " FD ") ||
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

	// Check UPI Reversal (before regular UPI)
	for _, pattern := range upiReversalPatterns {
		if strings.Contains(narration, pattern) {
			return "UPIReversal"
		}
	}

	// Check UPI
	for _, pattern := range upiPatterns {
		if strings.Contains(narration, pattern) {
			return "UPI"
		}
	}

	// Check IMPS Reversal (before regular IMPS)
	for _, pattern := range impsReversalPatterns {
		if strings.Contains(narration, pattern) {
			return "IMPSReversal"
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

	// Check RTGS (HDFC format: RTGS CR/DR-IFSC-NAME-NAME-REF)
	// Check for RTGS CR/DR first (more specific)
	if strings.HasPrefix(narration, "RTGS CR") || strings.HasPrefix(narration, "RTGS DR") {
		return "RTGS"
	}
	// Check other RTGS patterns
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

	// Check POS Reversal/Refund (before regular POS)
	for _, pattern := range posReversalPatterns {
		if strings.Contains(narration, pattern) {
			return "CardReversal"
		}
	}

	// Check International Card Markup/Charges (before regular POS)
	for _, pattern := range intlMarkupPatterns {
		if strings.Contains(narration, pattern) {
			return "CardCharges"
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

	// Check Salary (HDFC format: P:REF BANK SALARY FOR MONTH YEAR)
	// Check for P: prefix first (HDFC-specific but common pattern)
	if strings.HasPrefix(narration, "P:") && strings.Contains(narration, "SALARY") {
		return "Salary"
	}
	// Check other salary patterns
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
	
	// Check Bill Payment (ICICI-specific codes)
	// BBPS, BPAY, RCHG, TOP, BIL, PAVC
	for _, pattern := range billPaymentPatterns {
		if strings.Contains(narration, pattern) {
			return "BillPaid"
		}
	}
	
	// Check Online Shopping (ICICI-specific code)
	// ONL = Online shopping transaction (payment on third party website)
	for _, pattern := range onlineShoppingPatterns {
		if strings.Contains(narration, pattern) {
			return "OnlineShopping"
		}
	}
	
	// Check Tax Payment (ICICI-specific codes)
	// DTAX = Direct Tax, IDTX = Indirect Tax
	for _, pattern := range taxPatterns {
		if strings.Contains(narration, pattern) {
			return "TaxPayment"
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
		// ICICI-specific bill payment codes
		"BBPS", "BPAY", "RCHG", "TOP", "BIL-", "PAVC",
	}

	for _, pattern := range billPatterns {
		if strings.Contains(narration, pattern) {
			return true
		}
	}
	return false
}

// ExtractUPIDetails extracts merchant/payee name from UPI narration
// Handles multiple UPI formats:
// - UPI-MERCHANT-VPA@BANK-REF-UPI (standard format)
// - UPI-PERSON NAME-VPA@BANK-REF-UPI (P2P format)
// - UPI-MERCHANT-PAYTMQR...@PAYTM-BANK-REF-UPI (QR code format)
func ExtractUPIDetails(narration string) (merchant string, payee string) {
	narration = strings.TrimSpace(narration)
	
	// UPI format: UPI-MERCHANT/PERSON NAME-VPA@BANK-REF-UPI
	// Extract merchant/person name (between UPI- and first - or @)
	// Handle cases where name might contain spaces or hyphens
	re := regexp.MustCompile(`UPI-([^-@]+?)(?:-|@|$)`)
	matches := re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		merchant = strings.TrimSpace(matches[1])
		// Clean up common suffixes that might be part of merchant name
		merchant = strings.TrimSuffix(merchant, " -")
		merchant = strings.TrimSuffix(merchant, "-")
	}

	// Extract payee from UPI ID (VPA before @)
	// Format: VPA@BANK or PAYTMQR...@PAYTM
	re = regexp.MustCompile(`([^@\s]+)@`)
	matches = re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		payee = strings.TrimSpace(matches[1])
		// For QR codes, extract merchant name from QR data if possible
		if strings.HasPrefix(payee, "PAYTMQR") {
			// QR code format - merchant name is in the narration before QR
			if merchant == "" {
				// Try to extract from QR pattern
				re2 := regexp.MustCompile(`UPI-([^-]+)-PAYTMQR`)
				matches2 := re2.FindStringSubmatch(narration)
				if len(matches2) > 1 {
					merchant = strings.TrimSpace(matches2[1])
				}
			}
		}
	}

	return merchant, payee
}

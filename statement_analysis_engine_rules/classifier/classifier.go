package classifier

import (
	"classify/statement_analysis_engine_rules/analytics"
	"classify/statement_analysis_engine_rules/models"
	"classify/statement_analysis_engine_rules/rules"
	"classify/statement_analysis_engine_rules/utils"
	"strings"
)

// ClassifyTransaction classifies a single transaction
// Implements: Narration → Signals → Facts → Category
// customerName is optional - if provided, used for self-transfer detection
func ClassifyTransaction(txn models.ClassifiedTransaction, customerName string) models.ClassifiedTransaction {
	// Step 1: Clean narration first (critical - improves accuracy by 20-30%)
	normalizedNarration := utils.NormalizeNarration(txn.Narration)

	// Step 2: Extract signals (Channel, Gateway, Merchant, Intent)
	// Separate concepts: Channel, Gateway, Merchant, Intent
	// Channel detection (payment method)
	txn.Method = rules.ClassifyMethod(normalizedNarration)

	// Gateway detection (separate from channel)
	gateway := utils.ExtractGateway(normalizedNarration)

	// Merchant extraction and canonicalization (separate from category)
	rawMerchant := rules.ExtractMerchantName(normalizedNarration)
	if rawMerchant == "Unknown" {
		rawMerchant = ""
	}
	// Canonicalize merchant (normalize aliases - critical for long-term maintenance)
	canonicalMerchant, _ := utils.CanonicalizeMerchant(rawMerchant)
	if canonicalMerchant != "" {
		txn.Merchant = canonicalMerchant
	} else {
		txn.Merchant = rawMerchant
	}

	// Step 3: Classify category (Intent) with amount for charge detection
	// Determine the transaction amount - prioritize withdrawal for expense detection
	// Edge case: If both amounts exist (shouldn't happen in normal statements), use the larger one
	var amount float64
	if txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
		// Pure debit transaction
		amount = txn.WithdrawalAmt
	} else if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
		// Pure credit transaction
		amount = txn.DepositAmt
	} else if txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 {
		// Edge case: Both amounts present (unusual but handle it)
		// Use the larger amount for classification
		if txn.WithdrawalAmt > txn.DepositAmt {
			amount = txn.WithdrawalAmt
		} else {
			amount = txn.DepositAmt
		}
	} else {
		// No amount (shouldn't happen, but default to 0)
		amount = 0.0
	}

	// Get category with metadata (matched keywords, confidence, etc.)
	categoryResult := rules.ClassifyCategoryWithMetadata(normalizedNarration, txn.Merchant, amount)

	// Step 3.5: CRITICAL FIX - Credit transactions CANNOT be expenses
	// Expenses are only for debit transactions (money spent)
	// Credit transactions (deposits) should be Income, Refund, Investment (returns), or Other
	isCreditTransaction := txn.DepositAmt > 0 && txn.WithdrawalAmt == 0

	// List of expense categories that should ONLY apply to debit transactions
	expenseCategories := map[string]bool{
		"Shopping":        true,
		"Dining":          true,
		"Travel":          true,
		"Fuel":            true,
		"Groceries":       true,
		"Food_Delivery":   true,
		"Bills_Utilities": true,
		"Loan":            true,
		"Loan_EMI":        true,
		"LOAN_EMI":        true,
		"Healthcare":      true,
		"Education":       true,
		"Entertainment":   true,
		"Other":           true, // Other can be expense or income, but if credit, prefer Income/Refund
	}

	// If this is a credit transaction and was classified as an expense category, override it
	if isCreditTransaction && expenseCategories[categoryResult.Category] {
		// Credit transactions should be Income, Refund, Investment (returns), or Other (not expense categories)
		// Check for specific income/refund patterns first
		normalizedUpper := strings.ToUpper(normalizedNarration)

		// Check if it's a refund from shopping merchants
		refundMerchants := []string{
			"AMAZON", "FLIPKART", "MYNTRA", "AJIO", "MEESHO", "NYKAA",
			"ZARA", "HNM", "SHOPPERS STOP", "LIFESTYLE", "PANTALOONS",
			"CROMA", "RELIANCE DIGITAL", "VIJAY SALES", "SIMPL",
		}
		merchantUpper := strings.ToUpper(txn.Merchant)
		isRefundMerchant := false
		for _, refundMerchant := range refundMerchants {
			if strings.Contains(normalizedUpper, refundMerchant) || strings.Contains(merchantUpper, refundMerchant) {
				isRefundMerchant = true
				break
			}
		}

		if isRefundMerchant {
			categoryResult.Category = "Refund"
			categoryResult.Confidence = 0.90
			categoryResult.Reason = "Refund detected - credit from shopping merchant (expense category overridden)"
		} else {
			// Default credit transactions to Income or Other (not expense categories)
			// Check if it looks like investment returns
			investmentReturnKeywords := []string{
				"INVESTMENT", "BROKERAGE", "DEMAT", "TRADING", "MUTUAL FUND", "SIP",
				"STOCK", "SHARE", "SECURITIES", "BROKING", "BROKER",
				"FD", "FIXED DEPOSIT", "RD", "RECURRING DEPOSIT",
				"PPF", "PUBLIC PROVIDENT FUND", "NPS", "NATIONAL PENSION SYSTEM",
				"ZERODHA", "UPSTOX", "GROWW", "COIN",
			}
			hasInvestmentKeyword := false
			for _, keyword := range investmentReturnKeywords {
				if strings.Contains(normalizedUpper, keyword) {
					hasInvestmentKeyword = true
					break
				}
			}

			if hasInvestmentKeyword {
				categoryResult.Category = "Investment"
				categoryResult.Confidence = 0.85
				categoryResult.Reason = "Credit transaction classified as Investment (expense category overridden)"
			} else {
				// Default to Income for credit transactions that were misclassified as expenses
				categoryResult.Category = "Income"
				categoryResult.Confidence = 0.80
				categoryResult.Reason = "Credit transaction cannot be expense - classified as Income (expense category overridden)"
			}
		}
	}

	// Step 3.6: Handle Refunds and Reimbursements
	// Also handle POS reversals (CRV POS) and IMPS reversals (REV-IMPS) - these are refunds
	if txn.Method == "CardReversal" {
		// POS reversal is always a refund
		categoryResult.Category = "Refund"
		categoryResult.Confidence = 0.95
		categoryResult.Reason = "Card reversal/refund detected (CRV POS)"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "REFUND", "CRV_POS")
	} else if txn.Method == "IMPSReversal" {
		// IMPS reversal is always a refund
		// Pattern: REV-IMPS-112900179557-KALPIT KUMAR SHARMA-PYTM-XXXXXXXXXXXX8734-WAZIRX
		categoryResult.Category = "Refund"
		categoryResult.Confidence = 0.95
		categoryResult.Reason = "IMPS reversal/refund detected (REV-IMPS)"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "REFUND", "REV_IMPS")
	} else if txn.Method == "UPIReversal" {
		// UPI reversal is always a refund
		// Pattern: REV-UPI-08821130001725-KALPIT.COOL2006@OKHDFCBANK-209945000965-UPI
		categoryResult.Category = "Refund"
		categoryResult.Confidence = 0.95
		categoryResult.Reason = "UPI reversal/refund detected (REV-UPI)"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "REFUND", "REV_UPI")
	}

	// Step 3.7: Handle Loan EMI Reimbursements
	// Credit entries for loan EMI reimbursements should be categorized as "Reimbursement", not "Income"
	// Pattern: "P: A54152 STAFF LOAN EMI REC ..."
	if isCreditTransaction {
		normalizedUpper := strings.ToUpper(normalizedNarration)
		// Check for loan EMI reimbursement patterns
		loanEmiReimbursementPatterns := []string{
			"STAFF LOAN EMI REC",
			"LOAN EMI REC",
			"LOAN EMI REIMBURSEMENT",
			"EMI REC",
			"EMI REIMBURSEMENT",
			"LOAN REC",
		}
		for _, pattern := range loanEmiReimbursementPatterns {
			if strings.Contains(normalizedUpper, pattern) {
				categoryResult.Category = "Reimbursement"
				categoryResult.Confidence = 0.95
				categoryResult.Reason = "Loan EMI reimbursement detected - credit entry for loan EMI payment"
				categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "REIMBURSEMENT", "LOAN_EMI_REC")
				break
			}
		}
	}
	txn.Category = categoryResult.Category

	// Step 4: Extract beneficiary
	txn.Beneficiary = rules.ExtractBeneficiary(normalizedNarration, txn.Method)

	// Step 4.5: Handle FD premature closure and interest
	// Pattern: IB FD PREMAT PRINCIPAL-50300618314680 (withdrawal - principal returned)
	// Pattern: IB FD PREMAT INT PAID-50300618314680 (credit - interest paid)
	normalizedUpper := strings.ToUpper(normalizedNarration)
	if strings.Contains(normalizedUpper, "FD PREMAT") {
		if strings.Contains(normalizedUpper, "INT PAID") || strings.Contains(normalizedUpper, "INTEREST PAID") {
			// FD interest credit - this is income
			categoryResult.Category = "Income"
			categoryResult.Confidence = 0.95
			categoryResult.Reason = "FD premature closure interest paid (income)"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "FD", "INTEREST", "INCOME")
		} else if strings.Contains(normalizedUpper, "PRINCIPAL") {
			// FD principal withdrawal - this is investment withdrawal (not expense)
			categoryResult.Category = "Investment"
			categoryResult.Confidence = 0.95
			categoryResult.Reason = "FD premature closure principal returned (investment withdrawal)"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "FD", "PRINCIPAL", "INVESTMENT")
		}
	}

	// Step 5: Determine if income or expense
	// Dividends and Salary are always income (even if they come as deposits)
	if txn.Method == "Dividend" || txn.Method == "Salary" || txn.Method == "Interest" {
		txn.IsIncome = true
		// Also ensure category is Income for these payment methods
		if txn.Method == "Salary" {
			txn.Category = "Income"
			categoryResult.Category = "Income"
			categoryResult.Confidence = 0.98
			categoryResult.Reason = "Salary income detected"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SALARY")
		} else if txn.Method == "Dividend" {
			txn.Category = "Income"
			categoryResult.Category = "Income"
			categoryResult.Confidence = 0.95
			categoryResult.Reason = "Dividend income detected"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "DIVIDEND")
		} else if txn.Method == "Interest" {
			txn.Category = "Income"
			categoryResult.Category = "Income"
			categoryResult.Confidence = 0.95
			categoryResult.Reason = "Interest income detected"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "INTEREST")
		}
	} else {
		txn.IsIncome = txn.DepositAmt > 0
	}

	// Step 6: Priority overrides (high confidence rules)
	// Detect self-transfers (IMPS/NEFT/RTGS to same account holder)
	// Generic approach: Compare beneficiary name with account holder name (from metadata)
	// If beneficiary name matches account holder name, it's a self-transfer
	// normalizedUpper is already defined above, so we reuse it here
	
	// Pattern 0: Check for "OWN" in narration - indicates self-transfer
	// Examples: "NEFT DR-ICIC0004289-428901501818-NETBANK-...-OWN"
	//           "IMPS-500312393031-AKSHAY IDFC-IDFB-...-OWN"
	if strings.Contains(normalizedUpper, "-OWN") || strings.Contains(normalizedUpper, " OWN") || 
		strings.Contains(normalizedUpper, "OWN ") || strings.Contains(normalizedUpper, "-OWN-") {
		// "OWN" in narration indicates transfer to own account
		txn.Category = "Investment"
		categoryResult.Category = "Investment"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SELF_TRANSFER", "OWN")
		categoryResult.Confidence = 0.95
		categoryResult.Reason = "Self-transfer detected - 'OWN' indicator in narration"
	}

	if (txn.Method == "IMPS" || txn.Method == "NEFT" || txn.Method == "RTGS") && txn.Beneficiary != "" {
		beneficiaryUpper := strings.ToUpper(txn.Beneficiary)

		// Self-transfer indicators (generic patterns):
		// 1. First name of customer appears in narration - HIGHEST PRIORITY (people often use first name in beneficiary)
		// 2. Beneficiary name matches account holder name (from metadata) - HIGH CONFIDENCE
		// 3. Beneficiary name appears in narration + large round amounts (fallback)

		// Check for explicit self-transfer patterns
		isSelfTransfer := false

		// Pattern 1: Check if customer's first name appears in narration
		// People usually keep first name in beneficiary account name
		if customerName != "" {
			// Extract first name from customer name
			firstName := extractFirstName(customerName)
			if firstName != "" && len(firstName) >= 3 { // Only check if first name is at least 3 characters (avoid false positives)
				firstNameUpper := strings.ToUpper(firstName)
				// Check if first name appears in narration
				// Check for word boundaries to avoid partial matches (e.g., "KALPIT" shouldn't match "KALPITKUMAR")
				// Patterns: space/hyphen before and/or after, or at start/end of string
				if strings.Contains(normalizedUpper, " "+firstNameUpper+" ") ||
					strings.Contains(normalizedUpper, "-"+firstNameUpper+"-") ||
					strings.Contains(normalizedUpper, "-"+firstNameUpper+" ") ||
					strings.Contains(normalizedUpper, " "+firstNameUpper+"-") ||
					strings.HasPrefix(normalizedUpper, firstNameUpper+" ") ||
					strings.HasPrefix(normalizedUpper, firstNameUpper+"-") ||
					strings.HasSuffix(normalizedUpper, " "+firstNameUpper) ||
					strings.HasSuffix(normalizedUpper, "-"+firstNameUpper) ||
					normalizedUpper == firstNameUpper {
					// First name found in narration - high confidence self-transfer
					isSelfTransfer = true
				}
			}
		}

		// Pattern 2: Check if full customer name appears in narration
		// If customer account holder name is occurring in narration, it should be considered as self-transfer
		if !isSelfTransfer && customerName != "" {
			customerNameUpper := strings.ToUpper(customerName)
			// Remove common prefixes for matching
			customerNameUpper = strings.TrimPrefix(customerNameUpper, "MR. ")
			customerNameUpper = strings.TrimPrefix(customerNameUpper, "MR ")
			customerNameUpper = strings.TrimPrefix(customerNameUpper, "MRS. ")
			customerNameUpper = strings.TrimPrefix(customerNameUpper, "MRS ")
			customerNameUpper = strings.TrimPrefix(customerNameUpper, "MS. ")
			customerNameUpper = strings.TrimPrefix(customerNameUpper, "MS ")
			customerNameUpper = strings.TrimSpace(customerNameUpper)
			
			// Check if customer name appears in narration
			if customerNameUpper != "" && strings.Contains(normalizedUpper, customerNameUpper) {
				// Customer name found in narration - high confidence self-transfer
				isSelfTransfer = true
			}
		}

		// Pattern 3: Compare beneficiary name with account holder name (from metadata)
		// This is the most accurate method - compares all words in both names
		// Handles: full name vs partial name, different order, missing middle name
		if !isSelfTransfer && customerName != "" {
			if utils.MatchNames(txn.Beneficiary, customerName) {
				// High confidence: beneficiary name matches account holder name
				isSelfTransfer = true
			}
		}

		// Pattern 2: Fallback - Check if narration contains the beneficiary name + large round amounts
		// This is a weaker indicator but still useful when customerName is not available
		if !isSelfTransfer && strings.Contains(normalizedUpper, beneficiaryUpper) {
			// Additional check: Large round amounts are common for self-transfers
			// Typical self-transfers: ₹10K, ₹20K, ₹50K, ₹1L, etc.
			if amount >= 10000 {
				// Check if amount is a round number (common for self-transfers)
				isRoundAmount := (int(amount)%10000 == 0) || (int(amount)%50000 == 0) || (int(amount)%100000 == 0)
				if isRoundAmount {
					isSelfTransfer = true
				} else {
					// Even if not round, if beneficiary name is in narration and amount is large,
					// it's likely a self-transfer (medium confidence)
					if amount >= 50000 {
						isSelfTransfer = true
					}
				}
			}
		}

		if isSelfTransfer {
			// Self-transfers are treated as Investments (moving money to savings/investment accounts)
			txn.Category = "Investment"
			categoryResult.Category = "Investment"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SELF_TRANSFER", txn.Method)
			if customerName != "" {
				// Check if first name was used for detection
				firstName := extractFirstName(customerName)
				if firstName != "" && strings.Contains(normalizedUpper, strings.ToUpper(firstName)) {
					categoryResult.Confidence = 0.97 // Very high confidence when first name matches in narration
					categoryResult.Reason = "Self-transfer detected - customer's first name found in narration"
				} else {
					categoryResult.Confidence = 0.98 // Very high confidence when full customerName matches
					categoryResult.Reason = "Self-transfer detected - customer name matches beneficiary"
				}
			} else {
				categoryResult.Confidence = 0.95 // High confidence for fallback pattern
				categoryResult.Reason = "Self-transfer detected - classified as Investment (savings movement)"
			}
		}
	}

	// If Method is Self_Transfer, ensure Category is Investment
	// Self-transfers to own accounts are considered investments/savings movements
	if txn.Method == "Self_Transfer" {
		txn.Category = "Investment"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SELF_TRANSFER", "INF", "INFT")
		categoryResult.Confidence = 0.98 // Very high confidence for internal transfers
		categoryResult.Reason = "Internal fund transfer detected (INF/INFT) - classified as Investment (savings movement)"
	}

	// If Method is OnlineShopping (ICICI ONL code), ensure Category is Shopping
	// Only apply to debit transactions - expenses cannot be credits
	if txn.Method == "OnlineShopping" && txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
		txn.Category = "Shopping"
		categoryResult.Category = "Shopping"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "ONLINE_SHOPPING", "ONL")
		categoryResult.Confidence = 0.90 // High confidence for ONL code
		categoryResult.Reason = "Online shopping transaction detected (ONL) - classified as Shopping"
	}

	// If Method is TaxPayment (ICICI DTAX/IDTX codes), ensure Category is Bills_Utilities
	// Only apply to debit transactions - expenses cannot be credits
	if txn.Method == "TaxPayment" && txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
		txn.Category = "Bills_Utilities"
		categoryResult.Category = "Bills_Utilities"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "TAX_PAYMENT", "DTAX", "IDTX")
		categoryResult.Confidence = 0.95 // Very high confidence for tax payments
		categoryResult.Reason = "Tax payment detected (DTAX/IDTX) - classified as Bills_Utilities"
	}

	// If Method is RD, ensure Category is Investment
	// RD (Recurring Deposit) is a savings/investment product
	if txn.Method == "RD" {
		txn.Category = "Investment"
		categoryResult.Category = "Investment"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "RD", "RECURRING_DEPOSIT")
		categoryResult.Confidence = 0.98 // Very high confidence for RD
		categoryResult.Reason = "RD (Recurring Deposit) detected - classified as Investment"
	}

	// If Method is FD, ensure Category is Investment
	// FD (Fixed Deposit) is a savings/investment product
	// Exception: FD interest credits are already handled above as Income
	if txn.Method == "FD" && !strings.Contains(normalizedUpper, "INT PAID") && !strings.Contains(normalizedUpper, "INTEREST PAID") {
		txn.Category = "Investment"
		categoryResult.Category = "Investment"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "FD", "FIXED_DEPOSIT")
		categoryResult.Confidence = 0.98 // Very high confidence for FD
		categoryResult.Reason = "FD (Fixed Deposit) detected - classified as Investment"
	}

	// If Method is SIP, ensure Category is Investment
	// SIP (Systematic Investment Plan) is an investment product
	if txn.Method == "SIP" {
		txn.Category = "Investment"
		categoryResult.Category = "Investment"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SIP", "SYSTEMATIC_INVESTMENT")
		categoryResult.Confidence = 0.98 // Very high confidence for SIP
		categoryResult.Reason = "SIP (Systematic Investment Plan) detected - classified as Investment"
	}

	// If Method is EMI, ensure Category is Loan
	// Only apply to debit transactions - expenses cannot be credits
	if txn.Method == "EMI" && txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
		txn.Category = "Loan"
		categoryResult.Category = "Loan"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "EMI")
		categoryResult.Confidence = 0.95 // High confidence for EMI
		categoryResult.Reason = "EMI method detected - classified as Loan expense"
	}

	// If Method is Insurance, check if it's investment-type insurance
	// Investment-type insurance (ULIP, Endowment) should be Investment, not Bills_Utilities
	if txn.Method == "Insurance" {
		normalizedUpper := strings.ToUpper(normalizedNarration)
		investmentInsuranceKeywords := []string{
			"ULIP", "ENDOWMENT", "WHOLE LIFE", "MONEY BACK",
			"RETIREMENT", "PENSION PLAN", "SAVINGS PLAN",
		}
		for _, keyword := range investmentInsuranceKeywords {
			if strings.Contains(normalizedUpper, keyword) {
				// Override category to Investment for investment-type insurance
				txn.Category = "Investment"
				categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "INSURANCE", keyword)
				categoryResult.Confidence = 0.90
				categoryResult.Reason = "Investment-type insurance premium detected - classified as Investment expense"
				break
			}
		}
		// If not investment-type, keep as Bills_Utilities (default classification)
	}

	// Check if bill payment
	// Only apply to debit transactions - expenses cannot be credits
	if rules.IsBillPayment(normalizedNarration) {
		if txn.Category == "Other" && txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
			txn.Category = "Bills_Utilities"
			categoryResult.Category = "Bills_Utilities"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "BILL_PAYMENT")
			categoryResult.Reason = "Bill payment gateway detected"
		}
	}

	// Step 6.4: Improve ACH D categorization
	// ACH D - HDFC BANK LTD patterns are typically credit card payments or loan payments
	// Pattern: "ACH D- HDFC BANK LTD-408491108"
	if txn.Method == "ACH" && txn.WithdrawalAmt > 0 && txn.DepositAmt == 0 {
		normalizedUpper := strings.ToUpper(normalizedNarration)
		// Check for bank names in ACH D patterns - these are typically credit card or loan payments
		bankNames := []string{
			"HDFC BANK", "ICICI BANK", "SBI BANK", "AXIS BANK", "KOTAK BANK",
			"YES BANK", "IDFC BANK", "IDFC FIRST BANK", "PNB BANK",
		}
		hasBankName := false
		for _, bankName := range bankNames {
			if strings.Contains(normalizedUpper, "ACH D") && strings.Contains(normalizedUpper, bankName) {
				hasBankName = true
				break
			}
		}
		
		if hasBankName && txn.Category == "Other" {
			// Check if it looks like a credit card payment (recurring, similar amounts)
			// Or if it contains credit card related keywords
			creditCardKeywords := []string{"CREDIT CARD", "CC", "CARD BILL", "CARD PAYMENT"}
			isCreditCardPayment := false
			for _, keyword := range creditCardKeywords {
				if strings.Contains(normalizedUpper, keyword) {
					isCreditCardPayment = true
					break
				}
			}
			
			if isCreditCardPayment {
				txn.Category = "Bills_Utilities"
				categoryResult.Category = "Bills_Utilities"
				categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "CREDIT_CARD", "ACH_D")
				categoryResult.Confidence = 0.90
				categoryResult.Reason = "ACH D credit card payment detected"
			} else {
				// Default to Loan payment for ACH D to bank (could be loan EMI or credit card)
				// User can correct if needed, but this is better than "Other"
				txn.Category = "Loan"
				categoryResult.Category = "Loan"
				categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "ACH_D", "BANK")
				categoryResult.Confidence = 0.85
				categoryResult.Reason = "ACH D bank payment detected - likely loan or credit card payment"
			}
		}
	}

	// Step 6.5: FINAL SAFEGUARD - Ensure credit transactions are NEVER classified as expenses
	// This is a critical check to catch any edge cases that might have slipped through
	// Check if this is a credit transaction (pure credit or net credit)
	isCreditTxn := (txn.DepositAmt > 0 && txn.WithdrawalAmt == 0) ||
		(txn.DepositAmt > 0 && txn.WithdrawalAmt > 0 && txn.DepositAmt > txn.WithdrawalAmt)

	expenseCats := map[string]bool{
		"Shopping":        true,
		"Dining":          true,
		"Travel":          true,
		"Fuel":            true,
		"Groceries":       true,
		"Food_Delivery":   true,
		"Bills_Utilities": true,
		"Loan":            true,
		"Loan_EMI":        true,
		"LOAN_EMI":        true,
		"Healthcare":      true,
		"Education":       true,
		"Entertainment":   true,
	}

	if isCreditTxn && expenseCats[txn.Category] {
		// Credit transaction (or net credit) was classified as expense - override to Income
		// This should not happen if earlier checks worked, but this is a final safeguard
		txn.Category = "Income"
		categoryResult.Category = "Income"
		categoryResult.Confidence = 0.80
		categoryResult.Reason = "FINAL SAFEGUARD: Credit transaction cannot be expense - classified as Income"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "CREDIT_SAFEGUARD")
	}

	// Step 7: Build classification metadata (for explainability)
	// Detect amount pattern (secondary signal)
	amountPattern, hasAmountPattern := utils.DetectAmountPattern(amount)
	if hasAmountPattern {
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, amountPattern)
	}

	// Set gateway and channel (separate concepts)
	categoryResult.Gateway = gateway
	categoryResult.Channel = txn.Method
	categoryResult.RuleVersion = utils.RuleVersion

	// If no reason set, generate one
	if categoryResult.Reason == "" {
		categoryResult.Reason = generateReason(txn.Category, categoryResult.MatchedKeywords, gateway, txn.Method)
	}

	// Store metadata (for explainability - critical for debugging, audits, user trust)
	txn.ClassificationMetadata = models.ClassificationMetadata{
		Confidence:      categoryResult.Confidence,
		MatchedKeywords: categoryResult.MatchedKeywords,
		Gateway:         categoryResult.Gateway,
		Channel:         categoryResult.Channel,
		RuleVersion:     categoryResult.RuleVersion,
		Reason:          categoryResult.Reason,
	}

	return txn
}

// extractFirstName extracts the first name from a full name
// Handles names with prefixes like MR., MRS., MS., DR., etc.
// Returns the first word after removing prefixes
func extractFirstName(fullName string) string {
	if fullName == "" {
		return ""
	}
	
	// Normalize: remove common prefixes and trim
	name := strings.TrimSpace(fullName)
	name = strings.ToUpper(name)
	
	// Remove common prefixes
	prefixes := []string{"MR.", "MR ", "MRS.", "MRS ", "MS.", "MS ", "DR.", "DR ", "PROF.", "PROF "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			name = strings.TrimSpace(name)
			break
		}
	}
	
	// Split into words and return the first word
	words := strings.Fields(name)
	if len(words) > 0 {
		return words[0]
	}
	
	return ""
}

// generateReason creates a human-readable explanation for classification
func generateReason(category string, keywords []string, gateway string, channel string) string {
	reason := "Classified as " + category
	if len(keywords) > 0 {
		reason += " based on keywords: " + keywords[0]
		if len(keywords) > 1 {
			reason += ", " + keywords[1]
		}
	}
	if gateway != "" {
		reason += " via " + gateway
	}
	if channel != "" {
		reason += " (" + channel + ")"
	}
	return reason
}

// ClassifyTransactions classifies a list of transactions
// customerName is optional - if provided, used for self-transfer detection
func ClassifyTransactions(transactions []models.ClassifiedTransaction, customerName string) []models.ClassifiedTransaction {
	classified := make([]models.ClassifiedTransaction, len(transactions))

	// First pass: classify all transactions
	for i, txn := range transactions {
		classified[i] = ClassifyTransaction(txn, customerName)
	}

	// Second pass: detect recurring payments using comprehensive detection
	// PERFORMANCE FIX: Detect all recurring payments ONCE, then build lookup map
	// This avoids O(N²) complexity of calling DetectRecurringPayments() for each transaction
	detector := analytics.NewRecurringPaymentDetector(classified)
	recurringPayments := detector.DetectRecurringPayments()

	// Build lookup map for fast matching: signature -> RecurringPayment
	recurringMap := make(map[string]models.RecurringPayment)
	for _, rp := range recurringPayments {
		// Store by signature for fast lookup
		recurringMap[rp.Name] = rp
	}

	// Third pass: match each transaction to recurring payments using lookup map
	for i := range classified {
		recurringMetadata := analytics.MatchTransactionToRecurring(classified[i], detector, recurringMap)
		classified[i].IsRecurring = recurringMetadata.IsRecurring
		classified[i].RecurringMetadata = recurringMetadata
	}

	return classified
}

// ConvertFromTxtTransaction converts from extracted statement transaction to classified transaction
func ConvertFromTxtTransaction(date, narration, chequeRefNo, valueDate string, withdrawalAmt, depositAmt, closingBalance float64) models.ClassifiedTransaction {
	return models.ClassifiedTransaction{
		Date:           date,
		Narration:      narration,
		ChequeRefNo:    chequeRefNo,
		ValueDate:      valueDate,
		WithdrawalAmt:  withdrawalAmt,
		DepositAmt:     depositAmt,
		ClosingBalance: closingBalance,
		Method:         "",
		Category:       "",
		Merchant:       "",
		Beneficiary:    "",
		IsIncome:       depositAmt > 0,
		IsRecurring:    false,
	}
}

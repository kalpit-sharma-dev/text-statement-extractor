package classifier

import (
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
	amount := txn.DepositAmt
	if txn.WithdrawalAmt > 0 {
		amount = txn.WithdrawalAmt
	}
	
	// Get category with metadata (matched keywords, confidence, etc.)
	categoryResult := rules.ClassifyCategoryWithMetadata(normalizedNarration, txn.Merchant, amount)
	
	// Step 3.5: Handle Refunds - Credits from shopping merchants should be classified as Refund
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
	} else if txn.DepositAmt > 0 && txn.WithdrawalAmt == 0 {
		// Credits from shopping merchants are likely refunds
		refundMerchants := []string{
			"AMAZON", "FLIPKART", "MYNTRA", "AJIO", "MEESHO", "NYKAA",
			"ZARA", "HNM", "SHOPPERS STOP", "LIFESTYLE", "PANTALOONS",
			"CROMA", "RELIANCE DIGITAL", "VIJAY SALES", "SIMPL",
		}
		normalizedUpper := strings.ToUpper(normalizedNarration)
		merchantUpper := strings.ToUpper(txn.Merchant)
		
		for _, refundMerchant := range refundMerchants {
			if strings.Contains(normalizedUpper, refundMerchant) || strings.Contains(merchantUpper, refundMerchant) {
				categoryResult.Category = "Refund"
				categoryResult.Confidence = 0.90
				categoryResult.Reason = "Refund detected - credit from shopping merchant"
				categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "REFUND", refundMerchant)
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
	if (txn.Method == "IMPS" || txn.Method == "NEFT" || txn.Method == "RTGS") && txn.Beneficiary != "" {
		normalizedUpper := strings.ToUpper(normalizedNarration)
		beneficiaryUpper := strings.ToUpper(txn.Beneficiary)
		
		// Self-transfer indicators (generic patterns):
		// 1. Beneficiary name matches account holder name (from metadata) - HIGHEST CONFIDENCE
		// 2. Beneficiary name appears in narration + large round amounts (fallback)
		
		// Check for explicit self-transfer patterns
		isSelfTransfer := false
		
		// Pattern 1: Compare beneficiary name with account holder name (from metadata)
		// This is the most accurate method - compares all words in both names
		// Handles: full name vs partial name, different order, missing middle name
		if customerName != "" {
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
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "SELF_TRANSFER", txn.Method)
			if customerName != "" {
				categoryResult.Confidence = 0.98 // Very high confidence when customerName matches
			} else {
				categoryResult.Confidence = 0.95 // High confidence for fallback pattern
			}
			categoryResult.Reason = "Self-transfer detected - classified as Investment (savings movement)"
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
	if txn.Method == "OnlineShopping" {
		txn.Category = "Shopping"
		categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "ONLINE_SHOPPING", "ONL")
		categoryResult.Confidence = 0.90 // High confidence for ONL code
		categoryResult.Reason = "Online shopping transaction detected (ONL) - classified as Shopping"
	}
	
	// If Method is TaxPayment (ICICI DTAX/IDTX codes), ensure Category is Bills_Utilities
	if txn.Method == "TaxPayment" {
		txn.Category = "Bills_Utilities"
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
	if txn.Method == "EMI" {
		txn.Category = "Loan"
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
	if rules.IsBillPayment(normalizedNarration) {
		if txn.Category == "Other" {
			txn.Category = "Bills_Utilities"
			categoryResult.MatchedKeywords = append(categoryResult.MatchedKeywords, "BILL_PAYMENT")
			categoryResult.Reason = "Bill payment gateway detected"
		}
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
	previousNarrations := make([]string, 0)

	for i, txn := range transactions {
		// Classify transaction
		classified[i] = ClassifyTransaction(txn, customerName)

		// Check for recurring payments
		previousNarrations = append(previousNarrations, txn.Narration)
		if len(previousNarrations) > 10 {
			previousNarrations = previousNarrations[1:] // Keep last 10
		}

		classified[i].IsRecurring = rules.IsRecurringPayment(
			txn.Narration,
			txn.WithdrawalAmt+txn.DepositAmt,
			txn.Date,
			previousNarrations,
		)
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
